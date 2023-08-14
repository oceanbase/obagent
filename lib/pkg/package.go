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

package pkg

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/file"
	"github.com/oceanbase/obagent/lib/shell"
	"github.com/oceanbase/obagent/lib/shellf"
	"github.com/oceanbase/obagent/lib/system"
)

const (
	commandGroupPackageQuery            = "package.query"
	commandGroupPackageInstall          = "package.install"
	commandGroupPackageInstallDowngrade = "package.install.downgrade"
	commandGroupPackageInstallRelocate  = "package.install.relocate"
	commandGroupPackageUninstall        = "package.uninstall"
)

var packageNameRegex = regexp.MustCompile(`([a-zA-Z0-9\-]+)-(\d+.\d+.\d+(.\d+)?)-([\d.]+)\.([a-zA-Z0-9]+)\.([a-zA-Z0-9_]+)`)

type PackageInfo struct {
	FullPackageName string `json:"fullPackageName"` // full package name, e.g. `oceanbase-2.2.77-20210522122736.el7.x86_64`
	Name            string `json:"name"`            // package name, e.g. `oceanbase`
	Version         string `json:"version"`         // package version, e.g. `2.2.77`
	BuildNumber     string `json:"buildNumber"`     // package build number, e.g. `20210522122736`
	Os              string `json:"os"`              // package os, e.g. `el7`
	Architecture    string `json:"architecture"`    // package architecture, e.g. `x86_64`
}

type Package interface {
	// PackageInstalled Check whether software package is installed by name.
	// packageName can be `oceanbase` or `oceanbase-2.2.77.20210522122736.el7.x86_64`
	PackageInstalled(packageName string) (bool, error)

	// GetPackageInfo Get information of installed software package by name.
	// There should exists only one version of software package matching packageName,
	// otherwise an err will be returned.
	// packageName can be `oceanbase` or `oceanbase-2.2.77.20210522122736.el7.x86_64`
	GetPackageInfo(packageName string) (*PackageInfo, error)

	// FindPackageInfo Find information of installed software package by name.
	// If multiple installed package matches this name, multiple results will be returned.
	// packageName can be `oceanbase` or `oceanbase-2.2.77.20210522122736.el7.x86_64`
	FindPackageInfo(packageName string) ([]*PackageInfo, error)

	// InstallPackage Install software package from rpm file.
	InstallPackage(file string) error

	// InstallPackageToCustomPath Install software package from rpm file, with custom install path.
	InstallPackageToCustomPath(file string, installPath string) error

	// DowngradePackage Downgrade software package. Install an older version from rpm file.
	DowngradePackage(file string) error

	// UninstallPackage Uninstall software package by name.
	// packageName can be `oceanbase` or `oceanbase-2.2.77.20210522122736.el7.x86_64`
	UninstallPackage(packageName string) error

	// ExtractPackageAllFiles Extract all files from rpm file.
	ExtractPackageAllFiles(extractBasePath string, rpmPath string) error

	// ExtractPackageSingleFile Extract a single file from rpm file.
	ExtractPackageSingleFile(extractBasePath string, rpmPath string, fileInRpmPath string) error
}

type PackageImpl struct {
}

var libFile file.File = file.FileImpl{}
var libSystem system.System = system.SystemImpl{}
var libShell shell.Shell = shell.ShellImpl{}
var libShellf = shellf.Shelf()

func (p PackageImpl) PackageInstalled(packageName string) (bool, error) {
	packageNames, err := getInstalledPackageName(packageName)
	if err != nil {
		return false, errors.Wrap(err, "check package installed")
	}
	return len(packageNames) > 0, nil
}

func (p PackageImpl) GetPackageInfo(packageName string) (*PackageInfo, error) {
	packageNames, err := getInstalledPackageName(packageName)
	if err != nil {
		return nil, errors.Wrap(err, "get package info")
	}

	if len(packageNames) == 0 {
		return nil, errors.Errorf("cannot get package info of %v, no package installed", packageName)
	} else if len(packageNames) > 1 {
		return nil, errors.Errorf("cannot get package info of %v, multiple packages installed: %v", packageName, strings.Join(packageNames, ","))
	}
	return parsePackageInfo(packageNames[0])
}

func (p PackageImpl) FindPackageInfo(packageName string) ([]*PackageInfo, error) {
	packageNames, err := getInstalledPackageName(packageName)
	if err != nil {
		return nil, errors.Wrap(err, "find package info")
	}

	result := make([]*PackageInfo, 0)
	for _, packageName := range packageNames {
		if len(strings.TrimSpace(packageName)) == 0 {
			continue
		}
		info, parseErr := parsePackageInfo(packageName)
		if parseErr == nil {
			result = append(result, info)
		}
	}
	return result, nil
}

func parsePackageInfo(packageName string) (*PackageInfo, error) {
	info := &PackageInfo{
		FullPackageName: packageName,
	}
	groups := packageNameRegex.FindStringSubmatch(packageName)
	if len(groups) == 7 {
		info.Name = groups[1]
		info.Version = groups[2]
		info.BuildNumber = groups[4]
		info.Os = groups[5]
		info.Architecture = groups[6]
	}
	if info.Name == "" || info.Version == "" || info.Os == "" || info.Architecture == "" {
		return nil, errors.Errorf("invalid package name: %s", packageName)
	}
	return info, nil
}

func getInstalledPackageName(packageName string) ([]string, error) {
	command, err := getQueryPackageCommand(packageName)
	if err != nil {
		return nil, errors.Wrap(err, "get installed package name")
	}
	executeResult, err := command.ExecuteAllowFailure()
	if err != nil {
		return nil, errors.Wrap(err, "get installed package name")
	}
	if !executeResult.IsSuccessful() {
		return []string{}, nil
	}
	return executeResult.Lines(), nil
}

func (p PackageImpl) InstallPackage(file string) error {
	if hostInfo, err := libSystem.GetHostInfo(); err == nil && hostInfo.OsPlatformFamily == "debian" {
		log.Info("on debian platform, extract install package before install package")
		if err := p.extractInstallPackage(file); err != nil {
			return errors.Wrapf(err, "install package %s", file)
		}
	}
	command, err := getInstallPackageCommand(file)
	if err != nil {
		return errors.Wrapf(err, "install package %s", file)
	}
	_, err = command.Execute()
	if err != nil {
		return errors.Wrapf(err, "install package %s", file)
	}
	return nil
}

func (p PackageImpl) InstallPackageToCustomPath(file string, installPath string) error {
	if hostInfo, err := libSystem.GetHostInfo(); err == nil && hostInfo.OsPlatformFamily == "debian" {
		log.Info("on debian platform, extract install package before install package")
		if err := p.extractInstallPackage(file); err != nil {
			return errors.Wrapf(err, "install package %s to path %s", file, installPath)
		}
	}
	command, err := getInstallPackageRelocatePathCommand(file, installPath)
	if err != nil {
		return errors.Wrapf(err, "install package %s to path %s", file, installPath)
	}
	_, err = command.Execute()
	if err != nil {
		return errors.Wrapf(err, "install package %s to path %s", file, installPath)
	}
	return nil
}

func (p PackageImpl) DowngradePackage(file string) error {
	if hostInfo, err := libSystem.GetHostInfo(); err == nil && hostInfo.OsPlatformFamily == "debian" {
		log.Info("on debian platform, extract install package before install package")
		if err := p.extractInstallPackage(file); err != nil {
			return errors.Wrapf(err, "downgrade package %s", file)
		}
	}
	command, err := getDowngradePackageCommand(file)
	if err != nil {
		return errors.Wrapf(err, "downgrade package %s", file)
	}
	_, err = command.Execute()
	if err != nil {
		return errors.Wrapf(err, "downgrade package %s", file)
	}
	return nil
}

func (p PackageImpl) extractInstallPackage(rpmFile string) error {
	extractDir := filepath.Join("/tmp/rpms/extract/", filepath.Base(rpmFile))

	_ = libFile.RemoveDirectory(extractDir)
	err := libFile.CreateDirectoryForUser(extractDir, file.AdminUser, file.AdminGroup)
	if err != nil {
		return errors.Wrap(err, "extract install package")
	}

	err = p.ExtractPackageAllFiles(extractDir, rpmFile)
	if err != nil {
		return errors.Wrap(err, "extract install package")
	}

	files, err := libFile.ListFiles(extractDir, file.FileType)
	if err != nil {
		return errors.Wrap(err, "extract install package")
	}
	for _, path := range files {
		pathInRpm := strings.TrimPrefix(path, extractDir)
		quotedPath := fmt.Sprintf("%s/\"%s\"", filepath.Dir(pathInRpm), filepath.Base(pathInRpm))

		cmd := fmt.Sprintf("install -D %s%s %s", extractDir, quotedPath, quotedPath)
		_, err := libShell.NewCommand(cmd).Execute()
		if err != nil {
			return errors.Wrap(err, "extract install package")
		}
	}

	return nil
}

func (p PackageImpl) UninstallPackage(packageName string) error {
	command, err := getUninstallPackageCommand(packageName)
	if err != nil {
		return errors.Wrapf(err, "uninstall package %s", packageName)
	}
	_, err = command.Execute()
	if err != nil {
		return errors.Wrapf(err, "uninstall package %s", packageName)
	}
	return nil
}

func (p PackageImpl) ExtractPackageAllFiles(extractBasePath string, rpmPath string) error {
	cmd := getExtractPackageAllFileCommand(extractBasePath, rpmPath)
	return p.extractPackage(cmd)
}

func (p PackageImpl) ExtractPackageSingleFile(extractBasePath string, rpmPath string, fileInRpmPath string) error {
	cmd := getExtractPackageSingleFileCommand(extractBasePath, rpmPath, fileInRpmPath)
	return p.extractPackage(cmd)
}

func (p PackageImpl) extractPackage(cmd string) error {
	_, err := libShell.NewCommand(cmd).WithTimeout(3 * time.Minute).Execute()
	if err != nil {
		return errors.Wrap(err, "extract package")
	}
	return nil
}

func getQueryPackageCommand(packageName string) (shell.Command, error) {
	args := map[string]string{
		"PACKAGE_NAME": packageName,
	}
	return libShellf.GetCommandForCurrentPlatform(commandGroupPackageQuery, args)
}

func getInstallPackageCommand(file string) (shell.Command, error) {
	args := map[string]string{
		"PACKAGE_FILE": file,
	}
	return libShellf.GetCommandForCurrentPlatform(commandGroupPackageInstall, args)
}

func getDowngradePackageCommand(file string) (shell.Command, error) {
	args := map[string]string{
		"PACKAGE_FILE": file,
	}
	return libShellf.GetCommandForCurrentPlatform(commandGroupPackageInstallDowngrade, args)
}

func getInstallPackageRelocatePathCommand(file string, installPath string) (shell.Command, error) {
	args := map[string]string{
		"PACKAGE_FILE": file,
		"INSTALL_PATH": installPath,
	}
	return libShellf.GetCommandForCurrentPlatform(commandGroupPackageInstallRelocate, args)
}

func getUninstallPackageCommand(packageName string) (shell.Command, error) {
	args := map[string]string{
		"PACKAGE_NAME": packageName,
	}
	return libShellf.GetCommandForCurrentPlatform(commandGroupPackageUninstall, args)
}

// Extract all files from rpm
func getExtractPackageAllFileCommand(extractBasePath string, rpmPath string) string {
	return fmt.Sprintf("cd %v && rpm2cpio %v | cpio -id", extractBasePath, rpmPath)
}

// Extract one single file from rpm
// fileInRpmPath should start with `./`, e.g. `./home/admin/oceanbase/etc/oceanbase_upgrade_dep.yml`
func getExtractPackageSingleFileCommand(extractBasePath string, rpmPath string, fileInRpmPath string) string {
	fileInRpmPath = "./" + strings.TrimPrefix(fileInRpmPath, "/")
	return fmt.Sprintf("cd %v && rpm2cpio %v | cpio -id %v", extractBasePath, rpmPath, fileInRpmPath)
}
