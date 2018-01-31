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
func (j *job) Copy() *job {
	jc := *j
	return &jc
}

func (j *job) Pre() error {
	// do pre job
	return nil
}
func (j *job) Do() error {
	// do benchmark job
	_, err := http.Get(j.url)
	return err
}
func (j *job) After() error {
	// do clean job
	return nil
}

func main() {
	perfm := perfm.New(perfm.WithBinsNumber(15), perfm.WithParallel(1), perfm.WithNumber(20))

	j := &job{}
	j.url = "http://www.baidu.com"

	perfm.Regist(j)

	perfm.Start()
	perfm.Wait()
}
