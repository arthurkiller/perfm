package main

import (
	"net/http"

	"github.com/arthurkiller/perfm"
)

type job struct {
	// job private data
	url string
}

// Copy will called in parallel
func (j *job) Copy() (perfm.Job, error) {
	jc := *j
	return &jc, nil
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
