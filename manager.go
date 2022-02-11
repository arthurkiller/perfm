package perfm

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Monitor implemeneted perfMonitor
type Monitor struct {
	c    Config        //configration for perfm
	done chan struct{} //stop the perfm

	wg      sync.WaitGroup //wait group to block the stop and sync the work thread
	starter chan struct{}

	collector *Collector //get the request cost from every done()
	total     int64      //total request by count
	errCount  int64      //error counter count error request

	//job implement benchmark job
	//error occoured in job.Do will be collected
	job Job
}

// NewMonitor generate perfm
func NewMonitor(c Config) PerfMonitor {
	return &Monitor{
		c:         c,
		done:      make(chan struct{}),
		wg:        sync.WaitGroup{},
		starter:   make(chan struct{}),
		collector: NewCollector(&c),
		total:     0,
		errCount:  0,
	}
}

// Reset regist a job into Monitor fro benchmark
// This operation will reset the PerfMonitor
func (p *Monitor) Reset(job Job) {
	p.job = job

	// reset monitor
	p.done = make(chan struct{})
	p.wg = sync.WaitGroup{}
	p.errCount = 0
}

// TODO merge into one
// totalworker
// durationworker
func (p *Monitor) totalWorker() {
	// copy local job
	job, err := p.job.Copy()
	if err != nil {
		fmt.Println("error in do copy", err)
		return
	}
	// defer clean job
	defer job.After()
	var start time.Time
	// done local wg
	p.wg.Done()

	var l int64
	// wait for start
	<-p.starter
	for { // main work loop
		select {
		case <-p.done: // on close
			p.wg.Done()
			return
		default:
			// check if the request reach the goal
			if l = atomic.AddInt64(&p.total, 1); l > p.c.Number {
				if l == p.c.Number+1 { // double check, the last worker
					// only one should do close
					close(p.done)
				}
				// other goroutine exit now
				// TODO XXX continue?
				atomic.AddInt64(&p.total, -1)
				p.wg.Done()
				return
			}

			if err = job.Pre(); err != nil {
				fmt.Println("error in do pre job", err)
				p.wg.Done()
				return
			}
			start = time.Now()
			err = job.Do()
			p.collector.Collect(time.Since(start))
			if err != nil {
				atomic.AddInt64(&p.errCount, 1)
			}
		}
	}
}

func (p *Monitor) durationWorker() {
	// copy local job
	job, err := p.job.Copy()
	if err != nil {
		fmt.Println("error in do copy", err)
		return
	}
	// defer clean job
	defer job.After()
	var start time.Time
	// done local wg
	p.wg.Done()

	// wait for start
	<-p.starter
	for { // main work loop
		select {
		case <-p.done: // on close
			p.wg.Done()
			return
		default:
			atomic.AddInt64(&p.total, 1)
			if err = job.Pre(); err != nil {
				fmt.Println("error in do pre job", err)
				p.wg.Done()
				return
			}
			start = time.Now()
			err = job.Do()
			p.collector.Collect(time.Since(start))
			if err != nil {
				atomic.AddInt64(&p.errCount, 1)
			}
		}
	}
}

// Start the benchmark with given arguments on regisit
func (p *Monitor) Start(j Job) {
	if j != nil {
		p.job = j
	}
	if p.job == nil {
		panic("error job does not registered yet")
	}

	// If job implement descripetion as Stringer
	if _, ok := p.job.(fmt.Stringer); ok {
		fmt.Println(p.job)
	}

	fmt.Println("==================JOB STARTED====================")
	// 1. start monitor
	// 2. start all job, wait on wg
	// 3. run all jobs, start benchmark
	p.collector.Start()
	p.wg.Add(p.c.Parallel)
	// in test duration module
	// start all the worker and do job till cancelled
	for i := 0; i < p.c.Parallel; i++ {
		if p.c.Number != 0 {
			go p.totalWorker()
		} else {
			go p.durationWorker()
		}
	}

	// wait all job goroutine created
	p.wg.Wait()

	// now all goroutine created successfully, add wg again before start
	p.wg.Add(p.c.Parallel)

	// send signal, start all job
	close(p.starter)
	if p.c.Number == 0 { // for duration mode, sleep then stop
		time.Sleep(time.Second * time.Duration(p.c.Duration))
		close(p.done)
	}
	p.wg.Wait() // wait worker for exit

	// send close to collector, wait for collection done
	p.collector.WaitStop()

	fmt.Println("===================JOB DONE=======================")
	// print error info
	if p.errCount != 0 {
		fmt.Printf("Total errors: %v\t Error percentage: %.3f%%\n", p.errCount,
			float64(p.errCount*100)/float64(p.total))
	}

	// then do print
	p.collector.PrintResult(os.Stdout)

	// wait job done and do summarize
	p.wg.Wait()
}
