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
