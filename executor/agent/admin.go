/*
 * Copyright (c) 2023 OceanBase
 * OBAgent is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	process2 "github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/agentd/api"
	"github.com/oceanbase/obagent/lib/command"
	"github.com/oceanbase/obagent/lib/file"
	"github.com/oceanbase/obagent/lib/http"
	"github.com/oceanbase/obagent/lib/path"
	"github.com/oceanbase/obagent/lib/pkg"
	"github.com/oceanbase/obagent/lib/process"
)

type Admin struct {
	conf                AdminConf
	taskStore           command.StatusStore
	downloadPkgFunc     func(opCtx *OpCtx, param DownloadParam) error
	installPkgFunc      func(opCtx *OpCtx, pkgPath string) error
	checkCurrentPkgFunc func(opCtx *OpCtx) error
}

type AdminConf struct {
	RunDir           string
	ConfDir          string
	LogDir           string
	BackupDir        string
	TempDir          string
	TaskStoreDir     string
	AgentPkgName     string
	PkgExt           string
	PkgStoreDir      string
	StartWaitSeconds int
	StopWaitSeconds  int

	AgentdPath string
}

func DefaultAdminConf() AdminConf {
	return AdminConf{
		ConfDir:      path.ConfDir(),
		RunDir:       path.RunDir(),
		LogDir:       path.LogDir(),
		BackupDir:    path.BackupDir(),
		TempDir:      path.TempDir(),
		TaskStoreDir: path.TaskStoreDir(),
		AgentPkgName: filepath.Join(path.AgentDir(), "obagent"),
		PkgExt:       "rpm",
		PkgStoreDir:  path.PkgStoreDir(),

		StartWaitSeconds: 10,
		StopWaitSeconds:  10,
		AgentdPath:       path.AgentdPath(),
	}
}

type OpCtx struct {
	agentdPid   int
	mgrAgentPid int
	monAgentPid int

	curPkgPath string
	newPkgPath string
	tmpPkgPath string
	//tmpPkgInfo    *pkg.PackageInfo
	confBackupDir string

	rollbackErr error

	taskToken    command.ExecutionToken
	storedStatus *command.StoredStatus
}

//todo: update result and progress with async task id

type AdminLock struct {
	Pid int    `json:"pid"`
	Exe string `json:"exe"`
	//TaskToken string `json:"task_token"`
	//Operation string `json:"operation"`
}

func NewAdmin(conf AdminConf) *Admin {
	log.Infof("NewAdmin with config %+v", conf)
	ret := &Admin{
		conf:      conf,
		taskStore: command.NewFileTaskStore(conf.TaskStoreDir),
	}
	ret.checkCurrentPkgFunc = ret.checkCurrentPkg
	ret.downloadPkgFunc = ret.downloadPkg
	ret.installPkgFunc = ret.installPkg
	return ret
}

func (a *Admin) cleanDanglingLock() error {
	log.Infof("process %d try cleaning dangling admin lock", os.Getpid())
	lockPath := filepath.Join(a.conf.RunDir, "admin.lock")
	bytes, err := ioutil.ReadFile(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return CleanDanglingAdminLockFailedErr.NewError(lockPath).WithCause(err)
	}
	lock := AdminLock{}
	err = json.Unmarshal(bytes, &lock)
	if err != nil {
		log.WithError(err).Warnf("invalid lock file %s, will be removed", lockPath)
		err = a.Unlock()
		if err != nil {
			return CleanDanglingAdminLockFailedErr.NewError(lockPath).WithCause(err)
		}
	}

	proc, err := process2.NewProcess(int32(lock.Pid))
	if err == nil {
		exe, _ := proc.Exe()
		if exe == lock.Exe {
			log.Infof("lock file %s owner pid:%d, exe:%s still exists", lockPath, lock.Pid, lock.Exe)
			return nil
		}
	}
	err = a.Unlock()
	if err != nil {
		return CleanDanglingAdminLockFailedErr.NewError(lockPath).WithCause(err)
	}
	return nil
}

func (a *Admin) Lock() error {
	log.Infof("process %d fetching admin lock", os.Getpid())
	lockPath := filepath.Join(a.conf.RunDir, "admin.lock")
	err := a.cleanDanglingLock()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(lockPath, os.O_EXCL|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Warnf("fetch admin lock '%s' failed: %v", lockPath, err)
		return FetchAdminLockFailedErr.NewError(lockPath).WithCause(err)
	}
	defer f.Close()

	exe, _ := os.Executable()
	content, err := json.Marshal(AdminLock{
		Pid: os.Getpid(),
		Exe: exe,
	})
	if err != nil {
		return FetchAdminLockFailedErr.NewError(lockPath).WithCause(err)
	}
	_, err = f.Write(content)
	if err != nil {
		log.Warnf("write admin lock info failed: %v", err)
		return FetchAdminLockFailedErr.NewError(lockPath).WithCause(err)
	}
	log.Infof("process %d got admin lock", os.Getpid())
	return nil
}

func (a *Admin) Unlock() error {
	lockPath := filepath.Join(a.conf.RunDir, "admin.lock")
	err := os.Remove(lockPath)
	if err == nil {
		log.Infof("process %d release admin lock", os.Getpid())
		return nil
	}
	log.Errorf("process %d release admin lock '%s', got error: %v", os.Getpid(), lockPath, err)
	return ReleaseAdminLockFailedErr.NewError(lockPath).WithCause(err)
}

func (a *Admin) StartAgent() error {
	err := a.Lock()
	if err != nil {
		return err
	}
	defer a.Unlock()

	opCtx := OpCtx{}
	log.Info("starting agent")
	err = a.startAgent(&opCtx)
	if err == nil {
		log.Info("start agent succeed")
	} else {
		log.Errorf("start agent failed: %v", err)
	}
	return err
}

func (a *Admin) startAgent(opCtx *OpCtx) (err error) {
	a.progressStart(opCtx, "startAgent")
	defer a.progressEnd(opCtx, "startAgent", err)
	if a.pidFileRunning(path.Agentd) {
		log.Warn("agentd already running")
		return nil // AgentdAlreadyRunningErr.NewError()
	}

	agentdPath := a.conf.AgentdPath
	confPath := filepath.Join(a.conf.ConfDir, "agentd.yaml")
	agentdProc := process.NewProc(process.ProcessConfig{
		Program: agentdPath,
		Args:    []string{"-c", confPath},
		Stdout:  filepath.Join(a.conf.LogDir, "agentd.out.log"),
		Stderr:  filepath.Join(a.conf.LogDir, "agentd.err.log"),
	})
	log.Infof("starting agentd with config file '%s'", confPath)
	err = agentdProc.Start(nil)
	//agentdProc.Pid()
	if err != nil {
		log.Errorf("start agentd got error: %v", err)
		return err
	}
	var status api.Status
	for i := 0; i < a.conf.StartWaitSeconds; i++ {
		time.Sleep(time.Second)
		procState := agentdProc.State()
		if !procState.Running {
			log.Error("start agentd failed, agentd exited")
			err = AgentdExitedQuicklyErr.NewError()
			return
		}
		status, err = a.AgentStatus()
		if err == nil {
			if status.Ready {
				log.Info("agentd is ready")
				return nil
			}
		}
	}
	log.Error("wait for agent ready timeout")
	err = WaitForReadyTimeoutErr.NewError()
	return
}

func (a *Admin) signalPid(pid int, signal syscall.Signal) error {
	if pid == 0 {
		return nil
	}
	p, err := process2.NewProcess(int32(pid))
	if err != nil {
		log.Warnf("pid %d not exists", pid)
		return err
	}
	log.Infof("sending signal %v to pid %d", signal, pid)
	err = p.SendSignal(signal)
	if err != nil {
		log.Errorf("send signal %v to pid %d got error: %v", signal, pid, err)
	}
	return err
}

func (a *Admin) pidRunning(pid int, program string) bool {
	p, err := process2.NewProcess(int32(pid))
	if err != nil {
		return false
	}
	name, err := p.Name()
	if err != nil {
		return false
	}
	if name != program {
		return false
	}
	return true
}

func (a *Admin) pidFileRunning(program string) bool {
	pid, err := a.readPid(program)
	if err != nil {
		return false
	}
	return a.pidRunning(pid, program)
}

func (a *Admin) StopAgent(taskToken TaskToken) (err error) {
	log.Infof("StopAgent: %+v", taskToken)
	err = a.Lock()
	if err != nil {
		return
	}

	defer a.Unlock()
	opCtx := a.taskOpCtx(taskToken.TaskToken)
	err = a.stopAgent(&opCtx)
	return
}

func (a *Admin) stopAgent(opCtx *OpCtx) error {
	pidPath := PidPath(a.conf.RunDir, path.Agentd)
	if !fileExists(pidPath) {
		log.Warnf("pid file '%s' not exists. agentd may not running", pidPath)
		return nil
	}
	pid, err := ReadPid(pidPath)
	if err != nil {
		log.Errorf("can not find agentd via pid file '%s'", pidPath)
		return err
	}
	return a.stopAgentPid(opCtx, pid)
}

func (a *Admin) stopAgentPid(opCtx *OpCtx, pid int) (err error) {
	a.progressStart(opCtx, "stopAgentPid", pid)
	defer a.progressEnd(opCtx, "stopAgentPid", err)

	if !a.pidRunning(pid, path.Agentd) {
		log.Warnf("agentd not running")
		return nil
		//return AgentdNotRunningErr.NewError()
	}
	err = a.signalPid(pid, syscall.SIGTERM)
	if err != nil {
		return err
	}
	for i := 0; i < a.conf.StopWaitSeconds; i++ {
		time.Sleep(time.Second)
		if !a.pidRunning(pid, path.Agentd) {
			log.Infof("agentd with pid '%d' stopped", pid)
			return nil
		}
	}
	log.Errorf("wait for agentd with pid %d exit timeout", pid)
	err = WaitForExitTimeoutErr.NewError()
	return
}

func (a *Admin) AgentStatus() (api.Status, error) {
	log.Info("AgentStatus")
	status := api.Status{}
	cl, err := a.NewClient(path.Agentd)
	if err != nil {
		log.Errorf("failed create client of '%s': %v", path.Agentd, err)
		return status, err
	}
	err = cl.Call("/api/v1/status", nil, &status)
	if err != nil {
		log.Errorf("failed to get agentd status via api: %v", err)
		return status, err
	}
	log.Infof("check agentd status got: %+v", status)
	return status, nil
}

func (a *Admin) StartService(param StartStopServiceParam) error {
	log.Infof("StartService %+v", param)
	cl, err := a.NewClient(path.Agentd)
	if err != nil {
		log.Errorf("failed create client of '%s': %v", path.Agentd, err)
		return err
	}
	return cl.Call("/api/v1/startService", param, nil)
}

func (a *Admin) StopService(param StartStopServiceParam) error {
	log.Infof("StopService %+v", param)
	cl, err := a.NewClient(path.Agentd)
	if err != nil {
		log.Errorf("failed create client of '%s': %v", path.Agentd, err)
		return err
	}
	return cl.Call("/api/v1/stopService", param, nil)
}

func (a *Admin) ReinstallAgent(param ReinstallParam) (err error) {
	err = a.Lock()
	if err != nil {
		return err
	}
	defer a.Unlock()
	log.Infof("reinstall agent. param %+v", param)
	opCtx := a.taskOpCtx(param.TaskToken.TaskToken)
	defer func() {
		a.saveResult(&opCtx, err)
	}()

	err = a.checkCurrentPkgFunc(&opCtx)
	if err != nil {
		return err
	}
	if strings.HasPrefix(param.Source, "http:") || strings.HasPrefix(param.Source, "https:") {
		err = a.downloadPkgFunc(&opCtx, param.DownloadParam)
		if err != nil {
			log.Errorf("download agent package source: '%s' failed: %v", opCtx.tmpPkgPath, err)
			return err
		}
		defer os.Remove(opCtx.tmpPkgPath)
	} else {
		// todo check checksum, version
		log.Infof("use local package file: '%s'", param.Source)
		opCtx.tmpPkgPath = param.Source
	}
	err = a.backupConfig(&opCtx)
	if err != nil {
		return err
	}
	//defer func() {
	//	os.RemoveAll(opCtx.confBackupDir)
	//}()

	err = a.installPkgFunc(&opCtx, opCtx.tmpPkgPath)
	if err != nil {
		log.Errorf("install agent package '%s' failed: %v", opCtx.tmpPkgPath, err)
		return err
	}
	defer func() {
		if err != nil {
			err2 := a.restoreConfig(&opCtx)
			if err2 != nil {
				opCtx.rollbackErr = err2
				log.Errorf("restore config failed %v", err2)
			}
		}
	}()

	defer func() {
		if err != nil {
			// log
			err2 := a.installPkgFunc(&opCtx, opCtx.curPkgPath)
			if err2 != nil {
				opCtx.rollbackErr = err2
				log.Errorf("install prev package '%s' failed: %v", opCtx.curPkgPath, err2)
			}
		}
	}()
	err = a.updateConfig(&opCtx)

	if err != nil {
		return err
	}

	err = a.restartAgent(&opCtx)
	if err != nil {
		a.saveStartRollbackProgress(&opCtx)
		return err
	}
	err = a.saveInstallPackage(&opCtx)
	if err != nil {
		log.Errorf("save installed temp failed, %v", err)
	}
	log.Infof("reinstall agent succeed")
	return nil
}

func (a *Admin) saveInstallPackage(opCtx *OpCtx) (err error) {
	a.progressStart(opCtx, "saveInstallPackage")
	defer a.progressEnd(opCtx, "saveInstallPackage", err)

	var info *pkg.PackageInfo
	info, err = a.currentVersion(opCtx)
	if err != nil {
		return err
	}
	newPkgPath := filepath.Join(a.conf.PkgStoreDir, a.pkgFileName(info))
	err = os.Rename(opCtx.tmpPkgPath, newPkgPath)
	if err != nil {
		log.Errorf("move installed package from '%s' to '%s' failed, %v", opCtx.tmpPkgPath, newPkgPath, err)
		return err
	}
	return nil
}

func (a *Admin) backupPid(opCtx *OpCtx) (err error) {
	a.progressStart(opCtx, "backupPid")
	defer a.progressEnd(opCtx, "backupPid", err)

	agentdPid, err := BackupPid(a.conf.RunDir, path.Agentd)
	if err != nil {
		log.Errorf("backup pid for %s failed %v", path.Agentd, err)
		return err
	}
	opCtx.agentdPid = agentdPid

	mgrAgentPid, err := BackupPid(a.conf.RunDir, path.MgrAgent)
	if err != nil {
		log.Errorf("backup pid for %s failed %v", path.MgrAgent, err)
		return err
	}
	opCtx.mgrAgentPid = mgrAgentPid

	monAgentPid, err := BackupPid(a.conf.RunDir, path.MonAgent)
	if err != nil {
		log.Errorf("backup pid for %s failed %v", path.MonAgent, err)
		return err
	}
	opCtx.monAgentPid = monAgentPid
	return nil
}

func (a *Admin) restorePid(opCtx *OpCtx) {
	a.progressStart(opCtx, "restorePid")
	defer a.progressEnd(opCtx, "restorePid", nil)

	_ = RestorePid(a.conf.RunDir, path.Agentd, opCtx.agentdPid)
	_ = RestorePid(a.conf.RunDir, path.MgrAgent, opCtx.mgrAgentPid)
	_ = RestorePid(a.conf.RunDir, path.MonAgent, opCtx.monAgentPid)
}

func (a *Admin) confirmCanSignal(opCtx *OpCtx) (err error) {
	a.progressStart(opCtx, "confirmCanSignal")
	defer a.progressEnd(opCtx, "confirmCanSignal", err)

	err = a.signalPid(opCtx.agentdPid, 0)
	if err != nil {
		return err
	}
	err = a.signalPid(opCtx.mgrAgentPid, 0)
	if err != nil {
		return err
	}
	err = a.signalPid(opCtx.monAgentPid, 0)
	if err != nil {
		return err
	}
	return nil
}

func (a *Admin) taskOpCtx(token string) OpCtx {
	execToken := command.ExecutionTokenFromString(token)
	storedStatus, err := a.taskStore.Load(execToken)
	if err != nil {
		log.WithField("task_token", token).WithError(err).Warn("load task status failed")
		return OpCtx{}
	}
	return OpCtx{
		taskToken:    execToken,
		storedStatus: &storedStatus,
	}
}

func (a *Admin) RestartAgent(taskToken TaskToken) error {
	log.Infof("RestartAgent: %+v", taskToken)
	err := a.Lock()
	if err != nil {
		return err
	}
	defer a.Unlock()
	opCtx := a.taskOpCtx(taskToken.TaskToken)
	err = a.restartAgent(&opCtx)
	a.saveResult(&opCtx, err)
	return err
}

type ProgressEntry struct {
	Timestamp int64  `json:"timestamp"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
}

func (a *Admin) updateProgress(opCtx *OpCtx, entry ProgressEntry) {
	if opCtx.storedStatus == nil {
		log.WithField("task_token", opCtx.taskToken).Warn("updateProgress: missing storedStatus")
		return
	}
	if opCtx.storedStatus.Progress == nil {
		opCtx.storedStatus.Progress = []ProgressEntry{}
	}
	progress, ok := opCtx.storedStatus.Progress.([]ProgressEntry)
	if !ok {
		log.WithField("task_token", opCtx.taskToken).Error("invalid storedStatus progress")
	}
	progress = append(progress, entry)
	opCtx.storedStatus.Progress = progress
	err := a.taskStore.StoreStatus(opCtx.taskToken, *opCtx.storedStatus)
	if err != nil {
		log.WithField("task_token", opCtx.taskToken).WithError(err).Error("store progress failed")
	}
}

func (a *Admin) progressStart(opCtx *OpCtx, name string, args ...interface{}) {
	log.Infof("%s start: %+v", name, args)
	a.updateProgress(opCtx, ProgressEntry{
		Name:      name,
		Timestamp: time.Now().UnixNano(),
		Status:    "start",
	})
}

func (a *Admin) progressEnd(opCtx *OpCtx, name string, err error) {
	entry := ProgressEntry{
		Name:      name,
		Timestamp: time.Now().UnixNano(),
		Status:    "ok",
	}
	if err != nil {
		entry.Status = "fail"
		entry.Message = err.Error()
		log.WithError(err).Warnf("%s end with err", name)
	} else {
		log.Infof("%s end", name)
	}
	a.updateProgress(opCtx, entry)
}

func (a *Admin) saveResult(opCtx *OpCtx, err error) {
	if opCtx.storedStatus == nil {
		log.WithField("task_token", opCtx.taskToken).Warn("saveResult: missing storedStatus")
		return
	}
	if err != nil {
		a.progressEnd(opCtx, "rollback", opCtx.rollbackErr)
	}
	opCtx.storedStatus.Finished = true
	opCtx.storedStatus.Ok = err == nil
	opCtx.storedStatus.Result = true
	if err != nil {
		opCtx.storedStatus.Err = err.Error()
	}
	err = a.taskStore.StoreStatus(opCtx.taskToken, *opCtx.storedStatus)
	if err != nil {
		log.WithField("task_token", opCtx.taskToken).WithError(err).Errorf("store task result failed")
	}
}

func (a *Admin) saveStartRollbackProgress(opCtx *OpCtx) {
	a.progressStart(opCtx, "rollback")
}

func (a *Admin) restartAgent(opCtx *OpCtx) (err error) {
	a.progressStart(opCtx, "restartAgent")
	defer a.progressEnd(opCtx, "restartAgent", err)

	err = a.backupPid(opCtx)
	if err != nil {
		return err
	}
	err = a.confirmCanSignal(opCtx)

	err = a.startAgent(opCtx)
	if err != nil {
		err2 := a.stopAgent(opCtx)
		if err2 != nil {
			log.WithError(err2).Errorf("stop failed new agentd got error")
		}
		a.restorePid(opCtx)
		return err
	}
	err = a.stopAgentPid(opCtx, opCtx.agentdPid)
	if err != nil {
		// todo: force cleanup???
		return err
	}
	return nil
}

func (a *Admin) checkCurrentPkg(opCtx *OpCtx) (err error) {
	a.progressStart(opCtx, "checkCurrentPkg")
	defer a.progressEnd(opCtx, "checkCurrentPkg", err)

	curVer, err := a.currentVersion(opCtx)
	if err != nil {
		return err
	}
	storePath := filepath.Join(a.conf.PkgStoreDir, a.pkgFileName(curVer))
	_, err = os.Stat(storePath)
	if err != nil {
		log.Errorf("current agent package file '%s' not exists", storePath)
	} else {
		opCtx.curPkgPath = storePath
		log.Infof("found current agent package file '%s'", storePath)
	}
	return err
}

func (a *Admin) pkgFileName(info *pkg.PackageInfo) string {
	// obagent-3.2.0-.alios7.x86_64.rpm
	return fmt.Sprintf("%s.%s", info.FullPackageName, a.conf.PkgExt)
}

type DownloadParam struct {
	Source   string `json:"source"`
	Checksum string `json:"checksum"`
	Version  string `json:"version"`
}

type ReinstallParam struct {
	TaskToken
	DownloadParam
}

type StartStopServiceParam struct {
	TaskToken
	api.StartStopAgentParam
}

type TaskTokenParam interface {
	SetTaskToken(token string)
	GetTaskToken() string
}

type TaskToken struct {
	TaskToken string `json:"taskToken"`
}

func (t *TaskToken) SetTaskToken(token string) {
	t.TaskToken = token
}

func (t *TaskToken) GetTaskToken() string {
	return t.TaskToken
}

func (a *Admin) downloadPkg(opCtx *OpCtx, param DownloadParam) (err error) {
	a.progressStart(opCtx, "downloadPkg")
	defer a.progressEnd(opCtx, "downloadPkg", err)

	pkgFileName := fmt.Sprintf("%s-%s.%s", a.conf.AgentPkgName, param.Version, a.conf.PkgExt)
	tmpPath := filepath.Join(a.conf.TempDir, pkgFileName)
	err = http.HttpImpl{}.DownloadFile(tmpPath, param.Source)
	if err != nil {
		log.Errorf("download package from source '%s', failed: %v", param.Source, err)
		return
	}
	realChecksum, err := file.FileImpl{}.Sha256Checksum(tmpPath)
	if err != nil {
		log.Errorf("calculate checksum of temp package file '%s' failed: %v", tmpPath, err)
		return
	}
	if param.Checksum != realChecksum {
		log.Errorf("checksum not match. expected: '%s', real: '%s'", param.Checksum, realChecksum)
		return ChecksumNotMatchErr.NewError()
	}
	log.Infof("package from source '%s', checksum='%s', version='%s' downloaded. tmp path: %s", param.Source, param.Checksum, param.Version, tmpPath)
	//opCtx.tmpPkgInfo = info
	opCtx.tmpPkgPath = tmpPath
	return
}

func (a *Admin) installPkg(opCtx *OpCtx, pkgPath string) (err error) {
	a.progressStart(opCtx, "installPkg", pkgPath)
	defer a.progressEnd(opCtx, "installPkg", err)
	pkgImpl := pkg.PackageImpl{}
	err = pkgImpl.DowngradePackage(pkgPath) // force install package
	if err != nil {
		log.Errorf("install package '%s' failed: %v", pkgPath, err)
		return
	}
	return
}

func (a *Admin) currentVersion(opCtx *OpCtx) (info *pkg.PackageInfo, err error) {
	a.progressStart(opCtx, "currentVersion")
	defer a.progressEnd(opCtx, "currentVersion", err)

	pkgImpl := pkg.PackageImpl{}

	info, err = pkgImpl.GetPackageInfo(a.conf.AgentPkgName)
	if err != nil {
		log.Warnf("query current installed agent package '%s' version failed: %v", a.conf.AgentPkgName, err)
		return
	}
	log.Infof("current installed agent package '%s' version: %s", a.conf.AgentPkgName, info.Version)
	return
}

func (a *Admin) backupConfig(opCtx *OpCtx) (err error) {
	a.progressStart(opCtx, "backupConfig")
	defer a.progressEnd(opCtx, "backupConfig", err)

	confBackupDir := filepath.Join(a.conf.BackupDir, fmt.Sprintf("conf_%d", time.Now().UnixNano()))
	err = copyFiles(a.conf.ConfDir, confBackupDir)
	//err = os.Rename(a.conf.ConfDir, confBackupDir)
	if err == nil {
		opCtx.confBackupDir = confBackupDir
	}
	return
}

func (a *Admin) restoreConfig(opCtx *OpCtx) (err error) {
	a.progressStart(opCtx, "restoreConfig")
	defer a.progressEnd(opCtx, "restoreConfig", err)

	if opCtx.confBackupDir == "" {
		log.Warn("missing confBackupDir in OpCtx")
		return
	}
	err = os.RemoveAll(a.conf.ConfDir)
	if err != nil {
		return
	}
	err = copyFiles(opCtx.confBackupDir, a.conf.ConfDir)
	//err = os.Rename(opCtx.confBackupDir, a.conf.ConfDir)
	return
}

func (a *Admin) updateConfig(opCtx *OpCtx) (err error) {
	log.Info("updating config")
	a.progressStart(opCtx, "updateConfig")
	defer a.progressEnd(opCtx, "updateConfig", err)

	propertiesDir := "config_properties"
	backup := filepath.Join(opCtx.confBackupDir, propertiesDir)
	target := filepath.Join(a.conf.ConfDir, propertiesDir)
	err = copyFiles(backup, target)
	return
}

func (a *Admin) readPid(program string) (int, error) {
	pidPath := PidPath(a.conf.RunDir, program)
	return ReadPid(pidPath)
}

func (a *Admin) NewClient(program string) (*http.ApiClient, error) {
	pid, err := a.readPid(program)
	if err != nil {
		return nil, err
	}
	socketPath := SocketPath(a.conf.RunDir, program, pid)
	cl := http.NewSocketApiClient(socketPath, time.Second*10)
	return cl, nil
}

func copyFiles(src, dest string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return copyFile(src, dest)
	}
	err = os.MkdirAll(dest, info.Mode())
	if err != nil {
		return err
	}
	entries, err := in.Readdir(0)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		subSrc := filepath.Join(src, entry.Name())
		subDest := filepath.Join(dest, entry.Name())
		if entry.IsDir() {
			err = copyFiles(subSrc, subDest)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(subSrc, subDest)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dest string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	err = out.Sync()
	return
}

func (a *Admin) Repair() {
}
