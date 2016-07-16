package perfM

import "time"

type Job interface {
	Done(PerfMonitor) error //count the cost about this job and add to the perfmonitor count channel
}

type job struct {
	start time.Time //set for every single request start time
}

func (j *job) Done(p *perfMonitor) error {
	cost := time.Since(j.start)
	p.GlobalChannel <- cost
	return nil
}

type PerfMonitor interface {
	Start() error     //start the perf monitor
	Stop() error      //stop the perf montior
	Do() (Job, error) //set a timer to count the single request's cost
}

type perfMonitor struct {
	counter        int                //count the sum of the request
	startTime      time.Time          //keep the start time
	timer          time.Timer         //the frequency sampling timer
	GlobalChannel  chan time.Duration //get the request cost from every done()
	localCount     int                //count for the number in the sampling times
	localTimeCount time.Duration      //count for the sampling time total costs
}

func New(conf Config) {
	return &perfMonitor{
		counter:       0,
		startTime:     time.Time,
		timer:         time.NewTimer(time.Second * int64(conf.Frequency)),
		GlobalChannel: make(chan time.Duration, conf.BufferSize),
	}
}

func (p *perfMonitor) Start() error {
	p.startTime = time.Now()
	for {
		select {
		case cost := <-p.GlobalChannel:
			p.counter++
			p.localCount++
			p.localTimeCount += cost
		case <-p.timer:
			//TODO:show the courently performence info
			p.localCount = 0
			p.localTimeCount = 0
		}
	}
	return nil
}

func (p *perfMonitor) Stop() error {
	//TODO:show the info of the performence test
}

func (p *perfMonitor) Do() (*Job, error) {
	presentJob := new(job)
	presentJob.start = time.Now()
	return presentJob, nil
}
