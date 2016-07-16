package profM

import "time"

type Job interface {
	Done(ProfMonitor) error //count the cost about this job and add to the profmonitor count channel
}

type job struct {
	start time.Time //set for every single request start time
}

func (j *job) Done(p *pronfMonitor) error {
	cost := time.Since(j.start)
	p.GlobalChannel <- cost
	return nil
}

type ProfMonitor interface {
	Start() error     //start the prof monitor
	Stop() error      //stop the prof montior
	Do() (Job, error) //set a timer to count the single request's cost
}

type profMonitor struct {
	counter        int                //count the sum of the request
	startTime      time.Time          //keep the start time
	timer          time.Timer         //the frequency sampling timer
	GlobalChannel  chan time.Duration //get the request cost from every done()
	localCount     int                //count for the number in the sampling times
	localTimeCount time.Duration      //count for the sampling time total costs
}

func New(conf Config) {
	return &pronfMonitor{
		counter:       0,
		startTime:     time.Time,
		timer:         time.NewTimer(time.Second * int64(conf.Frequency)),
		GlobalChannel: make(chan time.Duration, conf.BufferSize),
	}
}

func (p *profMonitor) Start() error {
	p.startTime = time.Now()
	for {
		select {
		case cost := <-p.GlobalChannel:
			p.counter++
			p.localCount++
			p.localTimeCount += cost
		case <-p.timer:
			//TODO:show the courently prof info
			p.localCount = 0
			p.localTimeCount = 0
		}
	}
	return nil
}

func (p *profMonitor) Stop() error {
	//TODO:show the info of the prof test
}

func (p *profMonitor) Do() (*Job, error) {
	presentJob := new(job)
	presentJob.start = time.Now()
	return presentJob, nil
}
