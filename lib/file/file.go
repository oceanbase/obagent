package file

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/moby/sys/mountinfo"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/lib/shell"
)

const (
	AdminUser  = "admin"
	AdminGroup = "admin"
)

const (
	FileType = 0b1
	DirType  = 0b10
	LinkType = 0b100
)

const computeDirectoryUsedCommandTpl = "du -sk %v |awk '{print $1}'"

var libShell shell.Shell = shell.ShellImpl{}

type File interface {
	FileExists(path string) (bool, error)

	FileExistsEqualsSha1sum(path string, sha1sum string) (bool, error)

	Sha1Checksum(path string) (string, error)

	ReadFile(path string) (string, error)

	SaveFile(path string, content string, mode os.FileMode) error

	CopyFile(sourcePath string, targetPath string, mode os.FileMode) error

	RemoveFileIfExists(path string) error

	CreateDirectory(path string) error

	CreateDirectoryForUser(path string, userName string, groupName string) error

	RemoveDirectory(path string) error

	ChownDirectory(path string, userName string, groupName string, recursive bool) error

	ListFiles(basePath string, flag int) ([]string, error)

	CreateSymbolicLink(sourcePath string, targetPath string) error

	SymbolicLinkExists(linkPath string) (bool, error)

	GetRealPathBySymbolicLink(symbolicLink string) (string, error)

	IsDir(path string) bool

	IsDirEmpty(path string) (bool, error)

	IsFile(path string) bool

	GetDirectoryUsedBytes(path string, timeout time.Duration) (int64, error)

	FindFilesByRegexAndTimeSpan(ctx context.Context, findFilesParam FindFilesParam) ([]os.FileInfo, error)

	FindFilesByRegexAndMTime(ctx context.Context, dir string, fileRegex string, matchFunc MatchRegexFunc,
		mTime time.Duration, now time.Time, getFileTime GetFileTimeFunc) ([]os.FileInfo, error)

	GetFileStatInfo(ctx context.Context, f *os.File) (*FileInfoEx, error)

	Mount(ctx context.Context, source, target, mType, options string) error

	Unmount(ctx context.Context, path string) error

	IsMountPoint(ctx context.Context, fileName string) (bool, error)

	GetMountInfos(ctx context.Context, filter mountinfo.FilterFunc) ([]*mountinfo.Info, error)

	GetAllParentDirectories(ctx context.Context, path string) []string
}

type FileImpl struct {
}

func (f FileImpl) FileExists(path string) (bool, error) {
	return fileExists(path)
}

func fileExists(path string) (bool, error) {
	_, err := Fs.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.Errorf("failed to check file exists: %s", err)
	}
}

func (f FileImpl) Sha1Checksum(path string) (string, error) {
	file, err := Fs.Open(path)
	if err != nil {
		return "", errors.Errorf("failed to compute sha1 for %s, cannot open file: %s", path, err)
	}
	defer file.Close()

	hasher := sha1.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", errors.Errorf("failed to compute sha1 for %s, cannot read file: %s", path, err)
	}

	value := hex.EncodeToString(hasher.Sum(nil))
	log.WithFields(log.Fields{
		"file":     path,
		"checksum": value,
	}).Info("compute sha1 done")
	return value, nil
}

func (f FileImpl) ReadFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.Errorf("failed to read file %s: %s", path, err)
	}
	return string(data), nil
}

func (f FileImpl) SaveFile(path string, content string, mode os.FileMode) error {
	data := []byte(content)
	err := ioutil.WriteFile(path, data, mode)
	if err != nil {
		return errors.Errorf("failed to save file %s: %s", path, err)
	}
	return nil
}

func (f FileImpl) CopyFile(sourcePath string, targetPath string, mode os.FileMode) error {
	data, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return errors.Errorf("failed to copy file, cannot open file %s: %s", sourcePath, err)
	}

	err = ioutil.WriteFile(targetPath, data, mode)
	if err != nil {
		return errors.Wrapf(err, "failed to copy file, cannot write to file %s: %s", targetPath, err)
	}
	return nil
}

func (f FileImpl) RemoveFileIfExists(path string) error {
	if exists, err := f.FileExists(path); err == nil && !exists {
		log.WithField("file", path).Info("file not exists, skip remove")
		return nil
	}
	err := Fs.Remove(path)
	if err != nil {
		return errors.Errorf("failed to remove file %s: %s", path, err)
	}
	return nil
}

func (f FileImpl) FileExistsEqualsSha1sum(path string, expectSha1sum string) (bool, error) {
	exists, err := f.FileExists(path)
	if err != nil {
		return false, errors.Wrap(err, "check file exists with sha1")
	}
	if !exists {
		return false, nil
	}
	sha1sum, err := f.Sha1Checksum(path)
	if err != nil {
		return false, errors.Wrap(err, "check file exists with sha1")
	}
	return sha1sum == expectSha1sum, nil
}

func (f FileImpl) IsDir(path string) bool {
	return isDir(path)
}

func isDir(path string) bool {
	file, err := Fs.Stat(path)
	return err == nil && file.IsDir()
}

func (f FileImpl) IsDirEmpty(path string) (bool, error) {
	if !f.IsDir(path) {
		return false, errors.Errorf("specific path is not a dir: %s", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()
	_, err = file.Readdir(1)
	if err != nil && err == io.EOF {
		return true, nil
	}
	return false, err
}

func (f FileImpl) IsFile(path string) bool {
	file, err := Fs.Stat(path)
	return err == nil && !file.IsDir()
}

func (f FileImpl) GetRealPathBySymbolicLink(symbolicLink string) (string, error) {
	realPath, err := filepath.EvalSymlinks(symbolicLink)
	if err != nil {
		return "", errors.Errorf("failed to get real path of symlink %s: %s", symbolicLink, err)
	}
	return realPath, nil
}

func (f FileImpl) GetDirectoryUsedBytes(path string, timeout time.Duration) (int64, error) {
	command := fmt.Sprintf(computeDirectoryUsedCommandTpl, path)
	executeResult, err := libShell.NewCommand(command).WithOutputType(shell.StdOutput).WithTimeout(timeout).Execute()
	if err != nil {
		return 0, errors.Wrap(err, "get directory used bytes")
	}
	output := strings.TrimSpace(executeResult.Output)
	used, err := strconv.ParseInt(output, 10, 64)
	if err != nil {
		return 0, errors.Errorf("failed to get directory used bytes, invalid output %s: %s", output, err)
	}
	// KB -> B
	return used * 1024, nil
}

type MatchRegexFunc func(reg string, content string) (matched bool, err error)

type GetFileTimeFunc func(info os.FileInfo) (fileTime time.Time, err error)

func GetFileModTime(info os.FileInfo) (fileTime time.Time, err error) {
	return info.ModTime(), nil
}

// FindFilesByRegexAndMTime 按照 regex 以及 mTime 查询（mTime 为 0 那么忽略该条件），行为模仿了 linux find 命令
// 目前实现如下逻辑
// -mtime +duration : 列出在 now - duration 之前（不含本身）被更改过内容的文件名
// -mtime -duration : 列出在 now ~ now - duration 之内（不含本身）被更改过内容的文件名
func (f FileImpl) FindFilesByRegexAndMTime(
	ctx context.Context,
	dir string,
	fileRegex string,
	matchFunc MatchRegexFunc,
	mTime time.Duration,
	now time.Time,
	getFileTime GetFileTimeFunc,
) ([]os.FileInfo, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil, errors.Errorf("dir is empty")
	}
	if fileRegex == "" {
		fileRegex = ".*"
	}

	ctxLog := log.WithContext(ctx)
	skipMTime := mTime == 0
	isMTimeUnSigned := mTime > 0
	if mTime > 0 {
		mTime = -1 * mTime
	}
	checkPoint := now.Add(mTime)
	matchedFiles := make([]os.FileInfo, 0)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			ctxLog.WithError(err).Errorf("prevent panic by handling failure accessing a path %q", path)
			return err
		}
		isMatchedRegex, err := matchFunc(fileRegex, info.Name())
		if err != nil {
			return err
		}
		infoTime, err := getFileTime(info)
		if err != nil {
			return err
		}

		isMatchedMTime := skipMTime ||
			(!skipMTime && ((isMTimeUnSigned && infoTime.Before(checkPoint)) ||
				(!isMTimeUnSigned && (infoTime.After(checkPoint) || infoTime.Equal(checkPoint)))))

		if info.Mode().IsRegular() && isMatchedRegex && isMatchedMTime {
			matchedFiles = append(matchedFiles, info)
		}

		ctxLog.WithField("path", path).Debugf("visited file or dirs")
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matchedFiles, nil
}

type MatchMTimeFunc func(mTime, startTime, endTime time.Time) (matched bool, err error)
type MatchRegexpsFunc func(regs []string, content string) (matched bool, err error)

type FindFilesParam struct {
	Dir         string
	FileRegexps []string
	MatchRegex  MatchRegexpsFunc
	StartTime   time.Time
	EndTime     time.Time
	GetFileTime GetFileTimeFunc
	MatchMTime  MatchMTimeFunc
}

func (f FileImpl) FindFilesByRegexAndTimeSpan(
	ctx context.Context,
	param FindFilesParam,
) ([]os.FileInfo, error) {
	dir := strings.TrimSpace(param.Dir)
	if dir == "" || len(param.FileRegexps) == 0 {
		return nil, nil
	}

	ctxLog := log.WithContext(ctx)

	matchedFiles := make([]os.FileInfo, 0)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			ctxLog.WithError(err).Errorf("prevent panic by handling failure accessing a path %q", path)
			return err
		}
		isMatchedRegex, err := param.MatchRegex(param.FileRegexps, info.Name())
		if err != nil {
			return err
		}
		infoTime, err := param.GetFileTime(info)
		if err != nil {
			return err
		}

		isMatchedMTime, err := param.MatchMTime(infoTime, param.StartTime, param.EndTime)
		if err != nil {
			return err
		}

		if info.Mode().IsRegular() && isMatchedRegex && isMatchedMTime {
			matchedFiles = append(matchedFiles, info)
		}

		ctxLog.WithField("path", path).Debugf("visited file or dirs")
		return nil
	})
	if err != nil {
		return nil, err
	}
	return matchedFiles, nil
}

// GetFileStatInfo 获取文件 stat 信息，包含 fileId，devId 等
func (f FileImpl) GetFileStatInfo(ctx context.Context, file *os.File) (*FileInfoEx, error) {
	ctxLog := log.WithContext(ctx)
	ret, err := GetFileInfo(file)
	if err != nil {
		ctxLog.WithError(err).Error("failed get file info")
		return nil, err
	}
	return ret, nil
}

type PermissionType string

const (
	AccessRead    PermissionType = "ACCESS_READ"
	AccessWrite   PermissionType = "ACCESS_WRITE"
	AccessExecute PermissionType = "ACCESS_EXECUTE"
)

type CheckDirPermissionParam struct {
	Directory  string         `json:"directory"`  // directory to check permission
	User       string         `json:"user"`       // host user to check directory permissions, e.g. admin
	Permission PermissionType `json:"permission"` // expected permission of storage directory
}

type CheckDirectoryPermissionResult string

const (
	CheckFailed        CheckDirectoryPermissionResult = "CHECK_FAILED"
	DirectoryNotExists CheckDirectoryPermissionResult = "DIRECTORY_NOT_EXISTS"
	HasPermission      CheckDirectoryPermissionResult = "HAS_PERMISSION"
	NoPermission       CheckDirectoryPermissionResult = "NO_PERMISSION"
)

const checkDirectoryPermissionCommand = "if [ -\"%s\" \"%s\" ]; then echo 0; else echo 1; fi"
const hasPermissionOutput = "0"

var filePermissionShellValue = map[PermissionType]string{
	AccessRead:    "r",
	AccessWrite:   "w",
	AccessExecute: "x",
}

func (f FileImpl) CheckDirectoryPermission(ctx context.Context, dir string, user string, permissionType PermissionType) (CheckDirectoryPermissionResult, error) {
	ctxlog := log.WithContext(ctx).WithFields(log.Fields{
		"directory":      dir,
		"user":           user,
		"permissionType": permissionType,
	})

	exists, err := fileExists(dir)
	if err != nil {
		ctxlog.WithError(err).Info("check directory permission failed, cannot check directory exists")
		return CheckFailed, nil
	}
	if !exists {
		ctxlog.Info("check directory permission done, directory not exists")
		return DirectoryNotExists, nil
	}
	if !isDir(dir) {
		ctxlog.Info("check directory permission done, path is not directory")
		return DirectoryNotExists, nil
	}

	cmd := fmt.Sprintf(checkDirectoryPermissionCommand, filePermissionShellValue[permissionType], dir)
	executeResult, err := libShell.NewCommand(cmd).WithContext(ctx).WithUser(user).WithOutputType(shell.StdOutput).Execute()
	if err != nil {
		ctxlog.WithError(err).Info("check directory permission failed, cannot check directory permission")
		return CheckFailed, nil
	}
	if strings.TrimSpace(executeResult.Output) == hasPermissionOutput {
		ctxlog.Info("check directory permission done, directory has permission")
		return HasPermission, nil
	} else {
		ctxlog.Info("check directory permission done, directory has no permission")
		return NoPermission, nil
	}
}

func (f FileImpl) GetAllParentDirectories(ctx context.Context, path string) []string {
	curDir := path
	allParents := make([]string, 0)
	for curDir != "/" {
		curDir = filepath.Dir(curDir)
		allParents = append(allParents, curDir)
	}
	return allParents
}
