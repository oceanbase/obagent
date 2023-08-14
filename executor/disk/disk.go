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

package disk

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/errors"
	"github.com/oceanbase/obagent/lib/system"
)

var libDisk system.Disk = system.DiskImpl{}

type GetDiskUsageParam struct {
	Path string `json:"path" binding:"required"`
}

type GetFsTypeParam struct {
	Path string `json:"path" binding:"required"`
}

func GetDiskUsage(ctx context.Context, param GetDiskUsageParam) (*system.DiskUsage, *errors.OcpAgentError) {
	path := param.Path
	ctxlog := log.WithContext(ctx).WithField("path", path)

	usage, err := libDisk.GetDiskUsage(path)
	if err != nil {
		ctxlog.WithError(err).Error("get disk usage failed")
		return nil, errors.Occur(errors.ErrGetDiskUsage, path, err)
	}

	ctxlog.Infof("get disk usage done, usage=%#v\n", usage)
	return usage, nil
}

func BatchGetDiskInfos(ctx context.Context) ([]*system.DiskInfo, *errors.OcpAgentError) {
	diskInfos, err := libDisk.BatchGetDiskInfos()
	if err != nil {
		log.WithError(err).Error("get disk infos failed")
		return nil, errors.Occur(errors.ErrBatchGetDiskInfos, err)
	}

	log.Infof("get file system type done, diskInfos=%#v\n", diskInfos)
	return diskInfos, nil
}
