package perfM

import (
	"log"
	"time"

	hist "github.com/VividCortex/gohistogram"
)

type Job interface {
	Done(PerfMonitor) //count the cost about this job and add to the perfmonitor count channel
}

type job struct {
	start time.Time //set for every single request start time
}

func (j *job) Done(p PerfMonitor) {
	cost := time.Since(j.start)
	p.collect(cost)
}

type PerfMonitor interface {
	Start()                //start the perf monitor
	Stop()                 //stop the perf montior
	collect(time.Duration) //send the cost to the channel
	Do() Job               //set a timer to count the single request's cost
}

type perfMonitor struct {
	done           chan int               //stor the perfM
	counter        int                    //count the sum of the request
	startTime      time.Time              //keep the start time
	timer          <-chan time.Time       //the frequency sampling timer
	Collector      chan time.Duration     //get the request cost from every done()
	localCount     int                    //count for the number in the sampling times
	localTimeCount time.Duration          //count for the sampling time total costs
	histogram      *hist.NumericHistogram //used to print the histogram
	Buffer         chan float64           //buffer the test time to decrease the influence when add to the historgam
}

func New(conf Config) PerfMonitor {
	if conf.BinsNumber == 0 {
		conf.BinsNumber = 10
		conf.BufferSize = 1000000
		conf.Frequency = 1
	}
	return &perfMonitor{
		done:      make(chan int, 0),
		counter:   0,
		startTime: time.Now(),
		timer:     time.Tick(time.Duration(int64(1000000000 * conf.Frequency))),
		Collector: make(chan time.Duration, conf.BufferSize),
		histogram: hist.NewHistogram(conf.BinsNumber),
		Buffer:    make(chan float64, 100000000),
	}
}

func (p *perfMonitor) Start() {
	p.startTime = time.Now()
	for {
		select {
		case cost := <-p.Collector:
			p.counter++
			p.localCount++
			p.localTimeCount += cost
			p.Buffer <- float64(cost)
		case <-p.timer:
			if p.localCount == 0 {
				continue
			}
			log.Println("Qps: ", p.localCount, "Avg Latency: ", p.localTimeCount.Nanoseconds()/int64(p.localCount)/1000000)
			p.localCount = 0
			p.localTimeCount = 0
		case <-p.done:
			return
		}
	}
}

func (p *perfMonitor) Stop() {
	for {
		select {
		case d := <-p.Buffer:
			p.histogram.Add(d)
		default:
			p.done <- 1
			// here show the histogram
			log.Println(p.histogram.String())
			return
		}
	}

}

func (p *perfMonitor) Do() Job {
	presentJob := new(job)
	presentJob.start = time.Now()
	return presentJob
}

func (p *perfMonitor) collect(t time.Duration) {
	p.Collector <- t
}
