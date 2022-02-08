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
	c         Config         //configration for perfm
	wg        sync.WaitGroup //wait group to block the stop and sync the work thread
	done      chan int       //stop the perfm
	isStopped int64          //is stopped for double check
	startTime time.Time      //keep the start time

	localwg sync.WaitGroup
	starter *sync.Cond

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
		starter: sync.NewCond(&sync.Mutex{}),
	}
}

// Reset regist a job into Monitor fro benchmark
// This operation will reset the PerfMonitor
func (p *Monitor) Reset(job Job) {
	p.job = job

	// reset monitor
	p.done = make(chan int, 0)
	p.wg = sync.WaitGroup{}
	p.isStopped = 0
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
	p.localwg.Done()

	var l int64
	// wait for start
	p.starter.Wait()
	for { // main work loop
		select {
		case <-p.done: // on close
			p.localwg.Done()
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
				p.localwg.Done()
				return
			}

			if err = job.Pre(); err != nil {
				fmt.Println("error in do pre job", err)
				p.localwg.Done()
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
	p.localwg.Done()

	// wait for start
	p.starter.Wait()
	for { // main work loop
		select {
		case <-p.done: // on close
			p.localwg.Done()
			return
		default:
			atomic.AddInt64(&p.total, 1)
			if err = job.Pre(); err != nil {
				fmt.Println("error in do pre job", err)
				p.localwg.Done()
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
func (p *Monitor) Start() {
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
	p.localwg.Add(p.c.Parallel)
	// in test duration module
	// start all the worker and do job till cancelled
	for i := 0; i < p.c.Parallel; i++ {
		if p.c.Number != 0 {
			go p.totalWorker()
		} else {
			go p.durationWorker()
		}
	}

	// wait all job started
	p.wg.Wait()

	// now all goroutine started
	// add wg again before start
	p.localwg.Add(p.c.Parallel)

	// start all job
	p.starter.Broadcast()

	// for duration mode, sleep and send signal
	if p.c.Number == 0 {
		time.Sleep(time.Second * time.Duration(p.c.Duration))
		close(p.done)
	}
	p.localwg.Wait()
	// stop at here, wait for exit

	// print error info
	if p.errCount != 0 {
		fmt.Printf("Total errors: %v\t Error percentage: %.3f%%\n", p.errCount,
			float64(p.errCount*100)/float64(p.total))
	}

	// stop here
	p.collector.WaitStop()
	p.collector.PrintResult(os.Stdout)

	// wait job done and do summarize
	p.wg.Wait()
}
