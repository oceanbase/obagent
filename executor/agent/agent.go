package agent

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ReadPid(pidPath string) (int, error) {
	content, err := ioutil.ReadFile(pidPath)
	if err != nil {
		return 0, ReadPidFailedErr.NewError(pidPath).WithCause(err)
	}
	ret, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		return 0, ReadPidFailedErr.NewError(pidPath).WithCause(err)
	}
	return ret, nil
}

func BackupPid(runDir, program string) (int, error) {
	pidPath := PidPath(runDir, program)
	if !fileExists(pidPath) {
		return 0, nil
	}

	pid, err := ReadPid(pidPath)
	if err != nil {
		_ = os.Remove(pidPath)
		return 0, nil
		//return 0, BackupPidFailedErr.NewError(pidPath).WithCause(err)
	}
	backupPidPath := BackupPidPath(runDir, program, pid)
	err = os.Rename(pidPath, backupPidPath)
	if err != nil {
		return 0, BackupPidFailedErr.NewError(pidPath).WithCause(err)
	}
	return pid, nil
}

func RestorePid(runDir, program string, pid int) error {
	if pid <= 0 {
		return nil
	}
	backupPidPath := BackupPidPath(runDir, program, pid)
	pidPath := PidPath(runDir, program)
	err := os.Rename(backupPidPath, pidPath)
	if err != nil {
		return RestorePidFailedErr.NewError(pidPath).WithCause(err)
	}
	return nil
}

func PidPath(runDir, program string) string {
	return filepath.Join(runDir, program+".pid")
}

func BackupPidPath(runDir, program string, pid int) string {
	return filepath.Join(runDir, fmt.Sprintf("%s.%d.pid", program, pid))
}

func SocketPath(runDir, program string, pid int) string {
	return filepath.Join(runDir, fmt.Sprintf("%s.%d.sock", program, pid))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}
