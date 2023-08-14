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

package file

import (
	"os"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func (f FileImpl) CreateSymbolicLink(sourcePath string, targetPath string) error {
	linker, ok := Fs.(afero.Symlinker)
	if !ok {
		err := errors.New("symlink not supported by current file system")
		log.WithError(err).Error("create symbolic link failed")
		return err
	}
	err := linker.SymlinkIfPossible(sourcePath, targetPath)
	if err != nil {
		log.WithFields(log.Fields{
			"source": sourcePath,
			"target": targetPath,
		}).WithError(err).Info("create symbolic link failed")
		return errors.Errorf("failed to create symbolic link: %s", err)
	}
	log.WithFields(log.Fields{
		"source": sourcePath,
		"target": targetPath,
	}).Info("create symbolic link done")
	return nil
}

func (f FileImpl) SymbolicLinkExists(linkPath string) (bool, error) {
	linker, ok := Fs.(afero.Symlinker)
	if !ok {
		err := errors.New("symlink not supported by current file system")
		log.WithError(err).Error("check symbolic link exists failed")
		return false, err
	}

	fileInfo, lstatCalled, err := linker.LstatIfPossible(linkPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.WithError(err).Warn("symbolicLinkExists.lstatIfPossible failed")
			return false, nil
		}
		log.WithError(err).Error("check symbolic link exists failed")
		return false, errors.Errorf("failed to check symbolic link, lstat failed: %s", err)
	}
	if !lstatCalled {
		log.Infof("lstat not called, file %v is not symbolic link", linkPath)
		return false, nil
	}
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		log.Infof("file %v is symbolic link", linkPath)
		return true, nil
	} else {
		log.Infof("file %v is not symbolic link", linkPath)
		return false, nil
	}
}
