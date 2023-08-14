/*
 * Copyright (c) 2023 OceanBase
 * OCP Express is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package pkg

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/executor/agent"
	"github.com/oceanbase/obagent/executor/file"
	libfile "github.com/oceanbase/obagent/lib/file"
	"github.com/oceanbase/obagent/lib/pkg"
	"github.com/oceanbase/obagent/lib/shell"
)

var libShell shell.Shell = shell.ShellImpl{}
var libFile libfile.File = libfile.FileImpl{}
var libPackage pkg.Package = pkg.PackageImpl{}

type GetPackageInfoParam struct {
	Name string `json:"name" binding:"required"` // package name, e.g. `oceanbase` or `oceanbase-2.2.77-20210522122736.el7.x86_64`
}

type GetPackageInfoResult struct {
	PackageInfoList []*pkg.PackageInfo `json:"packageInfoList"`
}

type InstallPackageParam struct {
	agent.TaskToken
	Name        string  `json:"name" binding:"required"` // package name, e.g. `oceanbase` or `oceanbase-2.2.77-20210522122736.el7.x86_64`
	File        string  `json:"file" binding:"required"` // rpm file relative path, e.g. `rpms/oceanbase-2.2.77-20210522122736.el7.x86_64.rpm`
	InstallPath *string `json:"installPath"`             // custom install path, e.g. custom `/home/admin/oceanbase` to `/home/a/oceanbase`
}

type UninstallPackageParam struct {
	Name string `json:"name" binding:"required"` // package name, e.g. `oceanbase` or `oceanbase-2.2.77-20210522122736.el7.x86_64`
}

type InstallPackageResult struct {
	Changed     bool             `json:"changed"` // whether the install action is performed
	PackageInfo *pkg.PackageInfo `json:"packageInfo"`
}

type UninstallPackageResult struct {
	Changed bool `json:"changed"` // whether the uninstall action is performed
}

type InstalledPackageNamesResult struct {
	Names []string `json:"names"` // every installed package that share the same name
}

type ExtractPackageParam struct {
	agent.TaskToken
	PackageFile   string `json:"packageFile" binding:"required"`                       // rpm file relative path, e.g. `rpms/oceanbase-2.2.77-20210522122736.el7.x86_64.rpm`
	TargetPath    string `json:"targetPath" binding:"required"`                        // target path to store extracted files, e.g. `rpms/extract`
	ExtractAll    bool   `json:"extractAll"`                                           // whether to extract all files in rpm, true for all files, false for single file
	FileInPackage string `json:"fileInPackage" binding:"required_if=ExtractAll false"` // extract file path in rpm, only valid when ExtractAll is false, e.g. `/home/admin/oceanbase/etc/upgrade_pre.py`
}

type ExtractPackageResult struct {
	ExtractAll bool    `json:"extractAll"`         // whether extracted all files in rpm, true for all files, false for single file
	BasePath   string  `json:"basePath"`           // base path of all extracted files, e.g. `rpms/extract/oceanbase-xxx`
	FilePath   *string `json:"filePath,omitempty"` // path of the extracted single file, only valid when ExtractAll is false, e.g. `rpms/extract/oceanbase-xxx/home/admin/oceanbase/etc/upgrade_pre.py`
}

func GetPackageInfo(ctx context.Context, param GetPackageInfoParam) (*GetPackageInfoResult, *errors.OcpAgentError) {
	packageName := param.Name
	packageInfos, err := libPackage.FindPackageInfo(packageName)
	if err != nil {
		return nil, errors.Occur(errors.ErrQueryPackage, err)
	}
	log.WithContext(ctx).WithFields(log.Fields{
		"packageName":  packageName,
		"packageInfos": fmt.Sprintf("%#v", packageInfos),
	}).Info("get package info done")
	return &GetPackageInfoResult{PackageInfoList: packageInfos}, nil
}

func InstallPackage(ctx context.Context, param InstallPackageParam) (*InstallPackageResult, *errors.OcpAgentError) {
	packageName := param.Name
	rpmPath := file.NewPathFromRelPath(param.File)
	ctxlog := log.WithContext(ctx).WithFields(log.Fields{
		"packageName": packageName,
		"rpmFile":     rpmPath,
		"installPath": param.InstallPath,
	})

	if exists, err := libPackage.PackageInstalled(packageName); err == nil && exists {
		result := &InstallPackageResult{
			Changed: false,
		}
		if packageInfo, err := libPackage.GetPackageInfo(packageName); err == nil {
			result.PackageInfo = packageInfo
		}
		ctxlog.WithFields(log.Fields{
			"changed":     result.Changed,
			"packageInfo": result.PackageInfo,
		}).Info("install package skipped, package already installed")
		return result, nil
	}

	var err error
	if param.InstallPath != nil {
		err = libPackage.InstallPackageToCustomPath(rpmPath.FullPath(), *param.InstallPath)
	} else {
		err = libPackage.InstallPackage(rpmPath.FullPath())
	}
	if err != nil {
		return nil, errors.Occur(errors.ErrInstallPackage, err)
	}

	result := &InstallPackageResult{
		Changed: true,
	}
	if packageInfo, err := libPackage.GetPackageInfo(packageName); err == nil {
		result.PackageInfo = packageInfo
	}
	ctxlog.WithFields(log.Fields{
		"changed":     result.Changed,
		"packageInfo": result.PackageInfo,
	}).Info("install package done")
	return result, nil
}

func UninstallPackage(ctx context.Context, param UninstallPackageParam) (*UninstallPackageResult, *errors.OcpAgentError) {
	packageName := param.Name
	ctxlog := log.WithContext(ctx).WithField("packageName", packageName)
	if exists, err := libPackage.PackageInstalled(packageName); err == nil && !exists {
		ctxlog.Info("uninstall package skipped, package not exists")
		return &UninstallPackageResult{
			Changed: false,
		}, nil
	}
	err := libPackage.UninstallPackage(packageName)
	if err != nil {
		return nil, errors.Occur(errors.ErrUninstallPackage, err)
	}
	ctxlog.Info("uninstall package done")
	return &UninstallPackageResult{
		Changed: true,
	}, nil
}

func ExtractPackage(ctx context.Context, param ExtractPackageParam) (*ExtractPackageResult, *errors.OcpAgentError) {
	extractAll := param.ExtractAll
	rpmPath := file.NewPathFromRelPath(param.PackageFile)
	targetPath := file.NewPathFromRelPath(param.TargetPath)
	extractBasePath := targetPath.Join(rpmPath.FileName())
	fileInRpmPath := param.FileInPackage
	ctxlog := log.WithContext(ctx).WithFields(log.Fields{
		"extractBasePath": extractBasePath,
		"extractAll":      extractAll,
		"rpmPath":         rpmPath,
		"fileInRpmPath":   fileInRpmPath,
	})
	ctxlog.Info("extract package start")

	_ = libFile.RemoveDirectory(extractBasePath.FullPath())
	err := libFile.CreateDirectoryForUser(extractBasePath.FullPath(), libfile.AdminUser, libfile.AdminGroup)
	if err != nil {
		ctxlog.WithError(err).Error("extract package failed, cannot create directory")
		return nil, errors.Occur(errors.ErrCreateDirectory, extractBasePath, err)
	}

	if extractAll {
		err = libPackage.ExtractPackageAllFiles(extractBasePath.FullPath(), rpmPath.FullPath())
	} else {
		err = libPackage.ExtractPackageSingleFile(extractBasePath.FullPath(), rpmPath.FullPath(), fileInRpmPath)
	}

	if err != nil {
		return nil, errors.Occur(errors.ErrExtractPackage, err)
	}

	ctxlog.Info("extract package done")
	var result *ExtractPackageResult
	if extractAll {
		result = &ExtractPackageResult{
			ExtractAll: extractAll,
			BasePath:   extractBasePath.RelPath,
		}
	} else {
		extractFilePath := extractBasePath.Join(fileInRpmPath).RelPath
		result = &ExtractPackageResult{
			ExtractAll: extractAll,
			BasePath:   extractBasePath.RelPath,
			FilePath:   &extractFilePath,
		}
	}
	return result, nil
}
