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
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const defaultDirectoryPerm = 0755

func (f FileImpl) CreateDirectory(path string) error {
	err := os.MkdirAll(path, defaultDirectoryPerm)
	if err != nil {
		return errors.Errorf("failed to create directory %s: %s", path, err)
	}
	return nil
}

func (f FileImpl) CreateDirectoryForUser(path string, userName string, groupName string) error {
	err := f.CreateDirectory(path)
	if err != nil {
		return errors.Wrap(err, "create directory for user")
	}
	err = f.ChownDirectory(path, userName, groupName, true)
	if err != nil {
		return errors.Wrap(err, "create directory for user")
	}
	return nil
}

func (f FileImpl) RemoveDirectory(path string) error {
	err := Fs.RemoveAll(path)
	if err != nil {
		return errors.Errorf("failed to remove directory %s: %s", path, err)
	}
	return nil
}

func (f FileImpl) ChownDirectory(path string, userName string, groupName string, recursive bool) error {
	var err error
	uid, err := lookForUid(userName)
	if err != nil {
		return errors.Errorf("failed to chown directory, invalid user name %s: %s", userName, err)
	}
	gid, err := lookForGid(groupName)
	if err != nil {
		return errors.Errorf("failed to chown directory, invalid group name %s: %s", groupName, err)
	}
	if recursive {
		err = afero.Walk(Fs, path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			return os.Lchown(path, uid, gid)
		})
	} else {
		err = os.Lchown(path, uid, gid)
	}
	if err != nil {
		log.WithFields(log.Fields{
			"path":      path,
			"userName":  userName,
			"groupName": groupName,
			"recursive": recursive,
		}).WithError(err).Error("chown directory failed")
		return errors.Errorf("failed to chown directory %s: %s", path, err)
	}
	log.WithFields(log.Fields{
		"path":      path,
		"userName":  userName,
		"groupName": groupName,
		"recursive": recursive,
	}).Info("chown directory done")
	return nil
}

func lookForUid(userName string) (int, error) {
	u, err := user.Lookup(userName)
	if err != nil {
		return 0, err
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return 0, err
	}
	return uid, nil
}

func lookForGid(groupName string) (int, error) {
	g, err := user.LookupGroup(groupName)
	if err != nil {
		return 0, err
	}
	gid, err := strconv.Atoi(g.Gid)
	if err != nil {
		return 0, err
	}
	return gid, nil
}

func (f FileImpl) ListFiles(basePath string, flag int) ([]string, error) {
	var result []string
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if flag&DirType != 0 {
				result = append(result, path)
			}
		} else {
			if info.Mode()&fs.ModeSymlink != 0 {
				if flag&LinkType != 0 {
					result = append(result, path)
				}
			} else {
				if flag&FileType != 0 {
					result = append(result, path)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.Errorf("failed to list files under %s: %s", basePath, err)
	}
	return result, nil
}
