package main

import (
	"net/http"

	"github.com/arthurkiller/perfm"
)

type job struct {
	// job data here
	url string
}

// Copy will called in parallel
func (j *job) Copy() perfm.Job {
	jc := *j
	return &jc
}

func (j *job) Pre() {
	// do pre job
}
func (j *job) Do() error {
	// do benchmark job
	_, err := http.Get(j.url)
	return err
}
func (j *job) After() {
	// do clean job
}

func main() {
	perfm := perfm.New(perfm.WithBinsNumber(15), perfm.WithParallel(5), perfm.WithDuration(10))

	j := &job{}
	j.url = "http://www.baidu.com"

	perfm.Regist(j)

	perfm.Start()
	perfm.Wait()
}
