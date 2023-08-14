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

package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestRetryAlwaysFail(t *testing.T) {
	ctx := context.Background()
	retry := NewRetry()
	retryConf := RetryConfig{
		Name: "always-fail-retry-task",
		RetryEvent: func(ctx context.Context) error {
			return errors.New("failed")
		},
		RetryTimeout:  time.Second,
		MaxRetryTimes: 2,
		RetryDuration: &DefaultNextDuration{
			BaseDuration: time.Second,
			MaxDuration:  time.Second * 2,
		},
	}
	log.Infof("start:%+v", time.Now())
	retry.Do(ctx, retryConf)

	time.AfterFunc(time.Second, func() {
		retry.Do(ctx, retryConf)
	})

	time.Sleep(time.Second * 6)
}

func TestRetry_FirstTimesWithCancel(t *testing.T) {
	retry := NewRetry()
	retry.Cancel()
}
