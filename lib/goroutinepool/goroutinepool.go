package goroutinepool

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

type GoroutinePool struct {
	jobQueneSize, runPoolSize int
	queue                     chan job
	close                     chan bool
	closewg                   sync.WaitGroup
	closed                    uint32
}

type job struct {
	name string
	fn   func() error
}

func (gr *GoroutinePool) Put(jobName string, fn func() error) error {
	if atomic.LoadUint32(&gr.closed) == 1 {
		return errors.New("goruntinuepool closed")
	}
	gr.queue <- job{
		name: jobName,
		fn:   fn,
	}
	return nil
}

func (gr *GoroutinePool) PutWithTimeout(jobName string, fn func() error, duration time.Duration) error {
	if atomic.LoadUint32(&gr.closed) == 1 {
		return errors.New("goruntinuepool closed")
	}
	timeoutWrapper := func() error {
		finished := make(chan bool, 1)
		var err error
		go func() {
			err = fn()
			finished <- true
		}()

		after := time.After(duration)
		select {
		case <-finished:
			return err
		case <-after:
			log.Errorf("job %s timeout", jobName)
			return errors.New(fmt.Sprintf("job %s timeout", jobName))
		}
	}
	gr.queue <- job{
		fn:   timeoutWrapper,
		name: jobName,
	}
	return nil
}

func (gr *GoroutinePool) Close() {
	atomic.StoreUint32(&gr.closed, 1)
	close(gr.close)
	gr.closewg.Wait()
}

func NewGoroutinePool(poolName string, runPoolSize, jobQueneSize int) (*GoroutinePool, error) {
	if runPoolSize > jobQueneSize {
		return nil, errors.New("goroutinepool worker size should be less than job quene size")
	}
	if runPoolSize <= 0 || jobQueneSize <= 0 {
		return nil, errors.New("run pool size or job queue size should be greater than 0")
	}
	pool := &GoroutinePool{
		jobQueneSize: jobQueneSize,
		runPoolSize:  runPoolSize,
		queue:        make(chan job, jobQueneSize),
		close:        make(chan bool),
	}

	pool.closewg.Add(runPoolSize)
	for i := 0; i < runPoolSize; i++ {
		go func(i int) {
			defer pool.closewg.Done()
			for {
				select {
				case <-pool.close:
					log.Infof("#gorountine-pool-%s-%d, closed", poolName, i)
					return
				case job := <-pool.queue:
					log.Debugf("#gorountine-pool-%s-%d, got a job %s", poolName, i, job.name)
					if err := job.fn(); err != nil {
						log.Errorf("#gorountine-pool-%s-%d, job %s, got err: %+v", poolName, i, job.name, err)
					}
					log.Debugf("#gorountine-pool-%s-%d, %s job done", poolName, i, job.name)
				}
			}
		}(i)
	}
	return pool, nil
}
