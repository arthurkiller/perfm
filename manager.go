package perfm

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

// Monitor implemeneted perfMonitor
type Monitor struct {
	c     Config        //configration for perfm
	done  chan struct{} //stop the perfm, used in duratoin worker
	total int64         //total request, used in total worker

	wg      sync.WaitGroup //wait group to block the stop and sync the work thread
	starter chan struct{}

	collector *Collector //get the request cost from every done()

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
	}
}

// Reset regist a job into Monitor fro benchmark
// This operation will reset the PerfMonitor
func (p *Monitor) Reset(job Job) {
	p.job = job

	// reset monitor
	p.done = make(chan struct{})
	p.wg = sync.WaitGroup{}
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
				p.collector.ReportError(err)
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
				p.collector.ReportError(err)
			}
		}
	}
}

func (p *Monitor) processSiginter() {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	close(p.done)
	p.wg.Wait() // wait worker for exit

	// send close to collector, wait for collection done
	p.collector.WaitStop()

	fmt.Println("===============SIGINT RECEIVED====================")

	// then do print
	p.collector.PrintResult(os.Stdout)

	// wait job done and do summarize
	p.wg.Wait()

	os.Exit(0)
}

// Start the benchmark with given arguments on regisit
func (p *Monitor) Start(j Job) {
	if j != nil {
		p.job = j
	}
	if p.job == nil {
		panic("error job does not registered yet")
	}
	go p.processSiginter()

	// If job implement descripetion as Stringer
	if _, ok := p.job.(fmt.Stringer); ok {
		fmt.Println(p.job)
	}

	fmt.Println("==================JOB STARTED====================")
	// Steps:
	// 1. start all job, wait on wg
	// 2. run all jobs, start benchmark

	// wait all job goroutine created
	p.wg.Add(p.c.Parallel)
	for i := 0; i < p.c.Parallel; i++ {
		if p.c.Number != 0 {
			go p.totalWorker()
		} else {
			go p.durationWorker()
		}
	}
	p.wg.Wait()

	// now all goroutine created successfully, add wg again before start
	// this is used for exit waitting
	p.wg.Add(p.c.Parallel)

	p.collector.Start()  // mark the start time
	close(p.starter)     // send signal, start all jobs
	if p.c.Number == 0 { // duration mode, sleep then stop
		time.Sleep(time.Second * time.Duration(p.c.Duration))
		close(p.done)
	}
	p.wg.Wait() // wait workers exit

	// send close to collector, wait until collection done
	p.collector.WaitStop()

	fmt.Println("===================JOB DONE=======================")
	p.collector.PrintResult(os.Stdout)

	// wait job done and do summarize
	p.wg.Wait()
}
