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

package file

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/executor/agent"
	"github.com/oceanbase/obagent/lib/file"
	"github.com/oceanbase/obagent/lib/http"
	"github.com/oceanbase/obagent/lib/shell"
)

var libFile file.File = file.FileImpl{}
var libShell shell.Shell = shell.ShellImpl{}
var libHttp http.Http = http.HttpImpl{}

type BasicFileInfo struct {
	Path string `json:"path"`
}

type DownloadFileParam struct {
	agent.TaskToken
	Source           DownloadFileSource `json:"source"`           // download file source
	Target           string             `json:"target"`           // target file relative path
	ValidateChecksum bool               `json:"validateChecksum"` // whether validate checksum
	Checksum         string             `json:"checksum"`         // expected SHA-1 checksum of file
}

type DownloadFileSource struct {
	Type DownloadFileSourceType `json:"type"` // source type, OCP_URL or LOCAL_FILE
	Url  string                 `json:"url"`  // download http url, only valid when type is OCP_URL
	Path string                 `json:"path"` // local file path, only valid when type is LOCAL_FILE
}

type DownloadFileSourceType string

const (
	DownloadFileFromLocalFile DownloadFileSourceType = "LOCAL_FILE"
)

type GetFileExistsParam struct {
	FilePath string `json:"filePath"` // path of file
}

type ExistsResult struct {
	Exists bool `json:"exists"` // if path exists
}

type RemoveFilesParam struct {
	PathList []string `json:"pathList"`
}

type GetRealStaticPathParam struct {
	SymbolicLink string `json:"symbolicLink"` // symbolic link path
}

type RealStaticPath struct {
	Path string `json:"path"` // real absolute path
}

type GetDirectoryUsedParam struct {
	agent.TaskToken
	Path          string `json:"path"`          // directory path
	TimeoutMillis int64  `json:"timeoutMillis"` // shell command timeout
}

type DirectoryUsed struct {
	DirectoryUsedBytes int64 `json:"directoryUsedBytes"` // directory used bytes
}

func DownloadFile(ctx context.Context, param DownloadFileParam) (*BasicFileInfo, *errors.OcpAgentError) {
	targetPath := NewPathFromRelPath(param.Target)
	ctxLog := log.WithContext(ctx).WithFields(log.Fields{
		"source":           fmt.Sprintf("%#v", param.Source),
		"target":           targetPath,
		"validateChecksum": param.ValidateChecksum,
		"expectedChecksum": param.Checksum,
	})
	ctxLog.Info("download file")
	fileInfo := &BasicFileInfo{
		Path: param.Target,
	}
	if exists, err := libFile.FileExists(targetPath.FullPath()); err == nil && exists {
		if sha1sum, err := libFile.Sha1Checksum(targetPath.FullPath()); err == nil && sha1sum == param.Checksum {
			ctxLog.Info("download file skipped, file with same checksum already exists")
			return fileInfo, nil
		}
	}

	err := libFile.CreateDirectoryForUser(filepath.Dir(targetPath.FullPath()), file.AdminUser, file.AdminGroup)
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	if param.Source.Type == DownloadFileFromLocalFile {
		ctxLog.Info("download file from local file")
		err = libFile.CopyFile(param.Source.Path, targetPath.FullPath(), 0777)
		if err != nil {
			ctxLog.WithError(err).Errorf("download file from local file failed")
			return nil, errors.Occur(errors.ErrDownloadFile, param.Source.Path, err)
		}
	} else {
		ctxLog.Info("download file from ocp url")
		err = libHttp.DownloadFile(targetPath.FullPath(), param.Source.Url)
		if err != nil {
			ctxLog.WithError(err).Error("download file from ocp url failed")
			return nil, errors.Occur(errors.ErrDownloadFile, param.Source.Url, err)
		}
	}

	if param.ValidateChecksum {
		checksum, err := libFile.Sha1Checksum(targetPath.FullPath())
		if err != nil {
			return nil, errors.Occur(errors.ErrUnexpected, err)
		}
		if checksum != param.Checksum {
			ctxLog.WithFields(log.Fields{
				"actualChecksum": checksum,
			}).Error("download file validate checksum failed, invalid checksum")
			return nil, errors.Occur(errors.ErrInvalidChecksum)
		}
	} else {
		ctxLog.Info("download file validate checksum skipped")
	}
	ctxLog.Info("download file done")
	return fileInfo, nil
}

func IsFileExists(ctx context.Context, param GetFileExistsParam) (*ExistsResult, *errors.OcpAgentError) {
	if !checkFilePath(param.FilePath) {
		return nil, errors.Occur(errors.ErrIllegalArgument)
	}

	targetPath := NewPathFromRelPath(param.FilePath)
	ctxLog := log.WithContext(ctx).WithFields(log.Fields{
		"FilePath": param.FilePath,
	})

	ctxLog.Info("test path exists")
	exists, err := libFile.FileExists(targetPath.FullPath())
	if err != nil {
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return &ExistsResult{
		Exists: exists,
	}, nil
}

func RemoveFiles(ctx context.Context, param RemoveFilesParam) *errors.OcpAgentError {
	ctxLog := log.WithContext(ctx).WithFields(log.Fields{
		"PathList": param.PathList,
	})
	ctxLog.Info("remove files")
	for _, path := range param.PathList {
		targetPath := NewPathFromRelPath(path).FullPath()
		var err error = nil
		if libFile.IsDir(targetPath) {
			err = libFile.RemoveDirectory(targetPath)
		} else if libFile.IsFile(targetPath) {
			err = libFile.RemoveFileIfExists(targetPath)
		}

		if err != nil {
			ctxLog.WithError(err).Errorf("remove file:'%s' failed.", path)
			return errors.Occur(errors.ErrUnexpected, err)
		}
	}
	ctxLog.Info("remove files done")
	return nil
}

func GetRealStaticPath(ctx context.Context, param GetRealStaticPathParam) (*RealStaticPath, *errors.OcpAgentError) {
	if !checkFilePath(param.SymbolicLink) {
		return nil, errors.Occur(errors.ErrIllegalArgument)
	}
	ctxLog := log.WithContext(ctx).WithFields(log.Fields{
		"SymbolicLink": param.SymbolicLink,
	})
	ctxLog.Info("get real path by symbolic link")
	realPath, err := libFile.GetRealPathBySymbolicLink(param.SymbolicLink)
	if err != nil {
		ctxLog.WithError(err).Errorf("get real path:'%s' failed.", param.SymbolicLink)
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return &RealStaticPath{
		Path: realPath,
	}, nil
}

type PermissionType string

const (
	accessRead    PermissionType = "ACCESS_READ"
	accessWrite   PermissionType = "ACCESS_WRITE"
	accessExecute PermissionType = "ACCESS_EXECUTE"
)

var filePermissionShellValue = map[PermissionType]string{
	accessRead:    "r",
	accessWrite:   "w",
	accessExecute: "x",
}

type CheckDirectoryPermissionParm struct {
	Directory  string         `json:"directory"`  // directory to check permission
	User       string         `json:"user"`       // host user to check directory permissions, e.g. admin
	Permission PermissionType `json:"permission"` // expected permission of storage directory
}

type CheckDirectoryPermissionResult string

const (
	checkFailed        CheckDirectoryPermissionResult = "CHECK_FAILED"
	directoryNotExists CheckDirectoryPermissionResult = "DIRECTORY_NOT_EXISTS"
	hasPermission      CheckDirectoryPermissionResult = "HAS_PERMISSION"
	noPermission       CheckDirectoryPermissionResult = "NO_PERMISSION"
)

const checkDirectoryPermissionCommand = "if [ -\"%s\" \"%s\" ]; then echo 0; else echo 1; fi"
const hasPermissionOutput = "0"

func CheckDirectoryPermission(ctx context.Context, param CheckDirectoryPermissionParm) (CheckDirectoryPermissionResult, *errors.OcpAgentError) {
	ctxlog := log.WithContext(ctx).WithFields(log.Fields{
		"directory":  param.Directory,
		"user":       param.User,
		"permission": param.Permission,
	})

	exists, err := libFile.FileExists(param.Directory)
	if err != nil {
		ctxlog.WithError(err).Info("check directory permission failed, cannot check directory exists")
		return checkFailed, nil
	}
	if !exists {
		ctxlog.Info("check directory permission done, directory not exists")
		return directoryNotExists, nil
	}
	if !libFile.IsDir(param.Directory) {
		ctxlog.Info("check directory permission done, path is not directory")
		return directoryNotExists, nil
	}

	cmd := fmt.Sprintf(checkDirectoryPermissionCommand, filePermissionShellValue[param.Permission], param.Directory)
	executeResult, err := libShell.NewCommand(cmd).WithContext(ctx).WithUser(param.User).WithOutputType(shell.StdOutput).Execute()
	if err != nil {
		ctxlog.WithError(err).Info("check directory permission failed, cannot check directory permission")
		return checkFailed, nil
	}
	if strings.TrimSpace(executeResult.Output) == hasPermissionOutput {
		ctxlog.Info("check directory permission done, directory has permission")
		return hasPermission, nil
	} else {
		ctxlog.Info("check directory permission done, directory has no permission")
		return noPermission, nil
	}
}

func GetDirectoryUsed(ctx context.Context, param GetDirectoryUsedParam) (*DirectoryUsed, *errors.OcpAgentError) {
	ctxLog := log.WithContext(ctx).WithField("Path", param.Path)
	ctxLog.Info("get directory used bytes")
	result, err := libFile.GetDirectoryUsedBytes(param.Path, time.Duration(param.TimeoutMillis)*time.Millisecond)
	if err != nil {
		ctxLog.WithError(err).Errorf("get directory '%s' used failed", param.Path)
		return nil, errors.Occur(errors.ErrUnexpected, err)
	}
	return &DirectoryUsed{
		DirectoryUsedBytes: result,
	}, nil
}

func checkFilePath(filePath string) bool {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}
	if absPath != filePath {
		return false
	}
	return true
}
