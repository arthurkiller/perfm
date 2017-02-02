package perfm

import (
	"log"
	"time"

	hist "github.com/shafreeck/fperf/stats"
)

//Job defined the timer
type Job interface {
	Done() //count the cost about this job and add to the perfmonitor count channel
}

type job struct {
	start time.Time    //set for every single request start time
	p     *perfmonitor //store the perf monitor use for timer
}

func (j *job) Done() {
	cost := time.Since(j.start)
	j.p.collect(cost)
}

//PerfMonitor define the atcion about perfmonitor
type PerfMonitor interface {
	Start()                //start the perf monitor
	Stop()                 //stop the perf montior
	collect(time.Duration) //send the cost to the channel
	Do() Job               //set a timer to count the single request's cost
}

type perfmonitor struct {
	done           chan int           //stor the perfm
	counter        int                //count the sum of the request
	startTime      time.Time          //keep the start time
	timer          <-chan time.Time   //the frequency sampling timer
	Collector      chan time.Duration //get the request cost from every done()
	localCount     int                //count for the number in the sampling times
	localTimeCount time.Duration      //count for the sampling time total costs
	histogram      *hist.Histogram    //used to print the histogram
	Buffer         chan int64         //buffer the test time to decrease the influence when add to the historgam
	NoPrint        bool
}

//New gengrate the perfm monitor
func New(conf Config) PerfMonitor {
	histopt := hist.HistogramOptions{
		NumBuckets:     conf.BinsNumber,
		GrowthFactor:   conf.GrowthFactor,
		BaseBucketSize: conf.BaseBucketSize,
		MinValue:       conf.MinValue,
	}

	return &perfmonitor{
		done:      make(chan int, 0),
		counter:   0,
		startTime: time.Now(),
		timer:     time.Tick(time.Second * time.Duration(conf.Frequency)),
		Collector: make(chan time.Duration, conf.BufferSize),
		histogram: hist.NewHistogram(histopt),
		Buffer:    make(chan int64, 100000000),
		NoPrint:   conf.NoPrint,
	}
}

func (p *perfmonitor) Start() {
	p.startTime = time.Now()
	if p.NoPrint {
		for {
			select {
			case cost := <-p.Collector:
				p.counter++
				p.localCount++
				p.localTimeCount += cost
				p.Buffer <- int64(cost)
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
		return
	}
	for {
		select {
		case cost := <-p.Collector:
			p.counter++
			p.localCount++
			p.localTimeCount += cost
			p.Buffer <- int64(cost)
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

func (p *perfmonitor) Stop() {
	for {
		select {
		case d := <-p.Buffer:
			p.histogram.Add(d)
		default:
			close(p.done)
			// here show the histogram
			log.Println(p.histogram.String())
			return
		}
	}

}

func (p *perfmonitor) Do() Job {
	presentJob := new(job)
	presentJob.start = time.Now()
	presentJob.p = p
	return presentJob
}

func (p *perfmonitor) collect(t time.Duration) {
	p.Collector <- t
}
