package service

//
//type Scheduler struct {
//	waitGroup *sync.WaitGroup
//	pool      *ants.Pool
//	count     int
//	ticker    *time.Ticker
//}
//
//func NewScheduler(workerNums, maxWorkersPerSeconds int) (*Scheduler, error) {
//	var err error
//	wg := &sync.WaitGroup{}
//	pool, err := ants.NewPool(workerNums)
//	var ticker *time.Ticker
//	if maxWorkersPerSeconds > 0 {
//		ticker = time.NewTicker(
//			time.Second / time.Duration(maxWorkersPerSeconds))
//	}
//	service := &Scheduler{
//		waitGroup: wg,
//		pool:      pool,
//		ticker:    ticker,
//		count:     0,
//	}
//
//	return service, err
//}
//
//func (s *Scheduler) Submit(task func()) error {
//	if s.ticker != nil {
//		<-s.ticker.C
//	}
//	s.waitGroup.Add(1)
//	s.count++
//	return s.pool.Submit(func() {
//		defer s.waitGroup.Done()
//		task()
//		fmt.Println("task", s.count, "finished")
//	})
//}
//
//func (s *Scheduler) WaitUntilFinish() {
//	s.waitGroup.Wait()
//}
//
//func (s *Scheduler) Release() {
//	s.pool.Release()
//	if s.ticker != nil {
//		s.ticker.Stop()
//	}
//}
