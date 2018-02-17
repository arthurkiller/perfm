package perfm

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	hist "github.com/VividCortex/gohistogram"
)

// Job give out a job for parallel call
type Job interface {
	// Copy will copy a job for parallel call
	Copy() Job
	// Pre will called before do
	Pre()
	// Do contains the core job here
	Do() error
	// After contains the clean job after Do called
	After()
}

//PerfMonitor define the atcion about perfmonitor
type PerfMonitor interface {
	Regist(job Job) //regist the job to monitor
	Start()         //start the perf monitor
	Wait()          //wait for the benchmark done
}

type perfmonitor struct {
	Sum      float64 //Sum of the per request cost
	Stdev    float64 //Standard Deviation
	Mean     float64 //Mean about distribution
	Total    int64   //total request by count
	duration int     //test total time
	number   int     //total test request
	workers  int     //benchmark parallel worker number
	noPrint  bool    //disable the stdout print

	done           chan int           //stop the perfm
	startTime      time.Time          //keep the start time
	timer          <-chan time.Time   //the frequency sampling timer
	collector      chan time.Duration //get the request cost from every done()
	errCount       int64              //error counter count error request
	localCount     int                //count for the number in the sampling times
	localTimeCount time.Duration      //count for the sampling time total costs
	buffer         chan int64         //buffer the test time for latter add to the historgam
	histogram      hist.Histogram     //used to print the histogram
	wg             sync.WaitGroup     //wait group to block the stop and sync the work thread

	//job implement benchmark job
	//error will be collected occoured in job.Do
	job Job
}

//New gengrate the perfm monitor
func New(options ...Options) PerfMonitor {
	conf := newConfig(options...)

	var p *perfmonitor
	p = &perfmonitor{
		done:      make(chan int, 0),
		workers:   conf.Parallel,
		duration:  conf.Duration,
		number:    conf.Number,
		startTime: time.Now(),
		timer:     time.Tick(time.Second * time.Duration(conf.Frequency)),
		collector: make(chan time.Duration, conf.BufferSize),
		histogram: hist.NewHistogram(conf.BinsNumber),
		buffer:    make(chan int64, 100000000),
		noPrint:   conf.NoPrint,
		wg:        sync.WaitGroup{},
	}
	return p
}

// Regist a job into perfmonitor fro benchmark
func (p *perfmonitor) Regist(job Job) {
	p.job = job
}

// Start the benchmark with given arguments on regisit
func (p *perfmonitor) Start() {
	if p.job == nil {
		panic("error job does not regist correctly")
	}

	var localwg sync.WaitGroup

	p.wg.Add(1)
	go func() {
		p.startTime = time.Now()
		var cost time.Duration
		for {
			select {
			case cost = <-p.collector:
				p.localCount++
				p.localTimeCount += cost
				p.buffer <- int64(cost)
			case <-p.timer:
				if p.localCount == 0 {
					continue
				}
				fmt.Printf("Qps: %d \t  Avg Latency: %.3fms\n", p.localCount, float64(p.localTimeCount.Nanoseconds()/int64(p.localCount))/1000000)
				p.localCount = 0
				p.localTimeCount = 0
			case <-p.done:
				localwg.Wait()
				close(p.collector)
				for cost := range p.collector {
					p.localCount++
					p.localTimeCount += cost
					p.buffer <- int64(cost)
				}
				fmt.Printf("Qps: %d \t  Avg Latency: %.3fms\n", p.localCount, float64(p.localTimeCount.Nanoseconds()/int64(p.localCount))/1000000)

				p.wg.Done()
				return
			}
		}
	}()

	if p.number > 0 {
		// in total request module
		sum := int64(p.number)
		for i := 0; i < p.workers; i++ {
			localwg.Add(1)
			go func() {
				defer localwg.Done()
				job := p.job.Copy()
				var err error
				var start time.Time
				for {
					select {
					case <-p.done:
						return
					default:
						job.Pre()
						start = time.Now()
						err = job.Do()
						p.collector <- time.Since(start)
						if err != nil {
							atomic.AddInt64(&p.errCount, 1)
						}
						job.After()
						if atomic.AddInt64(&p.Total, 1) == sum {
							// check if the request reach the goal
							close(p.done)
							return
						}
					}
				}
			}()
		}
	} else {
		// in test duration module
		// start all the worker and do job till cancelled
		for i := 0; i < p.workers; i++ {
			localwg.Add(1)
			go func() {
				defer localwg.Done()
				job := p.job.Copy()
				var err error
				var start time.Time
				for {
					select {
					case <-p.done:
						return
					default:
						job.Pre()
						start = time.Now()
						err = job.Do()
						p.collector <- time.Since(start)
						atomic.AddInt64(&p.Total, 1)
						job.After()
						if err != nil {
							atomic.AddInt64(&p.errCount, 1)
						}
					}
				}
			}()
		}

		p.wg.Add(1)
		go func() {
			// stoper to cancell all the workers
			p.wg.Done()
			time.Sleep(time.Second * time.Duration(p.duration))
			close(p.done)
			return
		}()
	}
}

// Wait for the benchmark task done and caculate the result
func (p *perfmonitor) Wait() {
	p.wg.Wait()
	var sum2, i, d, max, min int64
	min = 0x7fffffffffffffff
	for i = 0; i < p.Total; i++ {
		d = <-p.buffer
		p.histogram.Add(float64(d))
		p.Sum += float64(d)
		sum2 += d * d
		if d > max {
			max = d
		}
		if d < min {
			min = d
		}
	}

	p.Mean = p.histogram.(*hist.NumericHistogram).Mean()
	p.Stdev = math.Sqrt((float64(sum2) - 2*float64(p.Mean*p.Sum) + float64(float64(p.Total)*p.Mean*p.Mean)) / float64(p.Total))

	// here show the histogram
	if !p.noPrint {
		fmt.Println("\n===============================================")
		if p.errCount != 0 {
			fmt.Printf("Total errors: %v\t Error percentage: %.3f%%\n", p.errCount, float64(p.errCount/p.Total*100))
		}
		fmt.Printf("MAX: %.3vms MIN: %.3vms MEAN: %.3vms STDEV: %.3f CV: %.3f%% ", max/1000000, min/1000000, p.Mean/1000000, p.Stdev/1000000, p.Stdev/float64(p.Mean)*100)
		fmt.Println(p.histogram)
	}
}
