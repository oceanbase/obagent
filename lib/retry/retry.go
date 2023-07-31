package retry

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	defaultRetryTimeout = time.Second * 3

	statusRetrying  = 1
	statusRetryExit = 0
)

type RetryConfig struct {
	Name          string
	RetryEvent    func(ctx context.Context) error
	RetryTimeout  time.Duration
	MaxRetryTimes int32
	RetryDuration RetryDurationConfig
}

func NewRetry() *Retry {
	return &Retry{
		exit:   make(chan struct{}, 1),
		status: statusRetryExit,
	}
}

type Retry struct {
	conf       RetryConfig
	retryTimes int32
	ctx        context.Context
	cancelFunc context.CancelFunc
	exit       chan struct{}
	mutex      sync.Mutex
	status     int32
}

func (r *Retry) plusRetryTimes() {
	atomic.SwapInt32(&r.retryTimes, r.retryTimes+1)
}

func (r *Retry) getRetryTimes() int32 {
	return atomic.LoadInt32(&r.retryTimes)
}

func (r *Retry) Do(ctx context.Context, conf RetryConfig) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.cancel()

	if conf.RetryTimeout == 0 {
		conf.RetryTimeout = defaultRetryTimeout
	}
	if conf.RetryDuration == nil {
		conf.RetryDuration = &DefaultNextDuration{
			BaseDuration: time.Second,
			MaxDuration:  10 * time.Minute,
		}
	}
	r.conf = conf
	cancelCtx, cancel := context.WithCancel(ctx)
	r.cancelFunc = cancel
	atomic.StoreInt32(&r.status, statusRetrying)
	go r.do(cancelCtx)
}

func (r *Retry) Cancel() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.cancel()
}

func (r *Retry) cancel() {
	if atomic.LoadInt32(&r.status) == statusRetrying {
		// If a task is being executed, cancel it and wait for it to exit before initiating a retry task.
		r.cancelFunc()
		<-r.exit
		log.Debug("cancelled the pre-task, and pre-task exited.")
	} else if atomic.LoadInt32(&r.status) == statusRetryExit {
		select {
		case <-r.exit:
			log.Debug("cancelled with the pre-task exited.")
		default:
			log.Debug("cancelled without pre-task.")
		}
	}
	atomic.SwapInt32(&r.retryTimes, 0)
	if r.conf.RetryDuration != nil {
		r.conf.RetryDuration.Reset()
	}
}

func (r *Retry) do(ctx context.Context) {
	defer func() {
		log.WithContext(ctx).Debugf("retry task %s exit", r.conf.Name)
		atomic.StoreInt32(&r.status, statusRetryExit)
		r.exit <- struct{}{}
	}()

	duration := r.conf.RetryDuration.NextDuration()
	for {
		if r.conf.MaxRetryTimes > 0 && r.getRetryTimes() > r.conf.MaxRetryTimes {
			log.WithContext(ctx).Infof("retry task %s has reached max retry times %d, exit.", r.conf.Name, r.conf.MaxRetryTimes)
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(duration):
			if !r.retryOnce(ctx) {
				// return if retry success
				return
			}
			duration = r.conf.RetryDuration.NextDuration()
		}
	}
}

func (r *Retry) retryOnce(ctx context.Context) (toBeContinue bool) {
	defer func() { r.plusRetryTimes() }()
	log.WithContext(ctx).Debugf("execute %s retry %d times", r.conf.Name, r.getRetryTimes())

	timeout := time.After(r.conf.RetryTimeout)
	ch := make(chan error, 1)
	go func() {
		err := r.conf.RetryEvent(ctx)
		if err != nil {
			log.WithContext(ctx).Infof("execute %s %d times, result err: %s", r.conf.Name, r.getRetryTimes(), err)
		} else if r.getRetryTimes() > 0 {
			log.WithContext(ctx).Infof("execute %s successfully, total retry %d times.", r.conf.Name, r.getRetryTimes())
		}
		ch <- err
	}()
	select {
	case <-timeout:
		log.WithContext(ctx).Warnf("execute %s timeout.", r.conf.Name)
		return true
	case err := <-ch:
		return err != nil
	case <-ctx.Done():
		log.WithContext(ctx).Infof("execute retry task %s is cancelled.", r.conf.Name)
		return false
	}
}

type RetryDurationConfig interface {
	NextDuration() time.Duration
	Reset()
}

type DefaultNextDuration struct {
	BaseDuration time.Duration // 1s
	MaxDuration  time.Duration
	nextDuration time.Duration
}

func (def *DefaultNextDuration) Reset() {
	def.nextDuration = 0
}

func (def *DefaultNextDuration) NextDuration() time.Duration {
	if def.nextDuration >= def.MaxDuration {
		return def.MaxDuration
	}
	if def.nextDuration == 0 {
		def.nextDuration = def.BaseDuration
		return 0
	}
	currentDuration := def.nextDuration
	def.nextDuration = currentDuration * 2
	return currentDuration
}
