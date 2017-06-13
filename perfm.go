package perfm

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	hist "github.com/VividCortex/gohistogram"
	//hist "github.com/shafreeck/fperf/stats"
)

//PerfMonitor define the atcion about perfmonitor
type PerfMonitor interface {
	Registe(func() error)
	Start() //start the perf monitor
	Wait()  //wait for the benchmark done
}

type perfmonitor struct {
	Sum      int64   //sum of the per request cost
	Stdev    float64 //Standard Deviation
	Avg      int64   //Average
	Total    int64   //total request by count
	duration int     //test total time
	number   int     //total test request
	workers  int     //benchmark parallel worker number
	noPrint  bool    //disable the stdout print

	done           chan int           //stop the perfm
	startTime      time.Time          //keep the start time
	timer          <-chan time.Time   //the frequency sampling timer
	collector      chan time.Duration //get the request cost from every done()
	errCollector   chan error         //error collector collect error request
	localCount     int                //count for the number in the sampling times
	localTimeCount time.Duration      //count for the sampling time total costs
	buffer         chan int64         //buffer the test time for latter add to the historgam
	histogram      hist.Histogram     //used to print the histogram
	wg             sync.WaitGroup     //wait group to block the stop and sync the work thread
	job            func() error       //job for benchmark, requect will be collected if an error occoured, job must be parallel safe
}

//New gengrate the perfm monitor
func New(options ...Options) PerfMonitor {
	conf := newConfig(options...)

	//histopt := hist.HistogramOptions{
	//	NumBuckets:     conf.BinsNumber,
	//	GrowthFactor:   conf.GrowthFactor,
	//	BaseBucketSize: conf.BaseBucketSize,
	//	MinValue:       conf.MinValue,
	//}

	var p *perfmonitor = &perfmonitor{
		done:         make(chan int, 0),
		workers:      conf.Parallel,
		duration:     conf.Duration,
		number:       conf.Number,
		startTime:    time.Now(),
		timer:        time.Tick(time.Second * time.Duration(conf.Frequency)),
		collector:    make(chan time.Duration, conf.BufferSize),
		errCollector: make(chan error, conf.BufferSize),
		//histogram:    hist.NewHistogram(histopt),
		histogram: hist.NewHistogram(conf.BinsNumber),
		buffer:    make(chan int64, 100000000),
		noPrint:   conf.NoPrint,
		wg:        sync.WaitGroup{},
	}
	return p
}

func (p *perfmonitor) Registe(job func() error) {
	p.job = job
}

func (p *perfmonitor) Start() {
	p.wg.Add(2)

	if p.job == nil {
		panic("error job does not regist yet")
	}

	var localwg sync.WaitGroup
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
				fmt.Println("DBG in print")
				if p.localCount == 0 {
					continue
				}
				fmt.Printf("Qps: %d  Avg Latency: %.3fms\n", p.localCount, float64(p.localTimeCount.Nanoseconds()/int64(p.localCount))/1000000)
				p.localCount = 0
				p.localTimeCount = 0
			case <-p.done:
				localwg.Wait()
				fmt.Println("DBG wati for workers")
				for {
					cost = <-p.collector
					p.buffer <- int64(cost)
					if len(p.collector) == 0 {
						p.wg.Done()
						return
					}
				}
			}
		}
	}()

	// start all the worker and do job till cancelled
	for i := 0; i < p.workers; i++ {
		localwg.Add(1)
		go func() {
			defer localwg.Done()
			job := p.job
			var err error
			var start time.Time
			for {
				select {
				case <-p.done:
					return
				default:
					start = time.Now()
					err = job()
					p.collector <- time.Since(start)
					atomic.AddInt64(&p.Total, 1)
					p.errCollector <- err
				}
			}
		}()
	}

	// stoper to cancell all the workers
	if p.number > 0 {
		// in total request module
		go func() {
			p.wg.Done()
			sum := int64(p.number)
			for {
				if atomic.LoadInt64(&p.Total) == sum {
					close(p.done)
					return
				}
				// 500us sleep time will cost 7% cpu time
				// 1ms sleep time cost only 4% cpu time that could be a better choice
				time.Sleep(time.Millisecond)
			}
		}()
	} else {
		// in test duration module
		go func() {
			fmt.Println("DBG in duration stoper")
			p.wg.Done()
			time.Sleep(time.Second * time.Duration(p.duration))
			close(p.done)
			return
		}()
	}
}

func (p *perfmonitor) Wait() {
	p.wg.Wait()
	var sum2, i, d int64
	fmt.Println("DBG BEFORE DRAW", len(p.buffer), p.Total)
	for i = 0; i < p.Total; i++ {
		d = <-p.buffer
		//fmt.Println("DBG add", d)
		p.histogram.Add(float64(d))
		p.Sum += d
		sum2 += d * d
	}
	fmt.Println("DBG after caculate")

	p.Avg = p.Sum / p.Total
	p.Stdev = math.Sqrt(float64(sum2) - 2*float64(p.Avg*p.Sum) + float64(p.Total*p.Avg*p.Avg)/float64(p.Total))

	// here show the histogram
	if !p.noPrint {
		fmt.Printf("\nSTDEV: %f CV: %f % \n%s\n", p.Stdev, p.Stdev/float64(p.Avg)*100, p.histogram.String())
	}
}
