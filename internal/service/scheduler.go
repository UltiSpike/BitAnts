package service

import (
	"fmt"
	"github.com/panjf2000/ants/v2"
	"sync"
	"time"
)

type Scheduler struct {
	waitGroup *sync.WaitGroup
	pool      *ants.Pool
	ticker    *time.Ticker
	count     int
}

func NewScheduler(workerNums, maxWorkersPerSeconds int) (*Scheduler, error) {
	var err error
	wg := &sync.WaitGroup{}
	pool, err := ants.NewPool(workerNums)
	var ticker *time.Ticker
	if maxWorkersPerSeconds > 0 {
		ticker = time.NewTicker(
			time.Second / time.Duration(maxWorkersPerSeconds))
	}
	scheduler := &Scheduler{
		waitGroup: wg,
		pool:      pool,
		ticker:    ticker,
		count:     0,
	}

	return scheduler, err
}

func (s *Scheduler) Submit(task func()) error {
	if s.ticker != nil {
		<-s.ticker.C
	}
	s.waitGroup.Add(1)
	s.count++
	return s.pool.Submit(func() {
		task()
		fmt.Println("task", s.count, "finished")
		s.waitGroup.Done()
	})
}

func (s *Scheduler) WaitUntilFinish() {
	s.waitGroup.Wait()
}

func (s *Scheduler) Release() {
	s.pool.Release()
	if s.ticker != nil {
		s.ticker.Stop()
	}
}
