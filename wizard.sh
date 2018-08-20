#!/bin/bash 
filename="job.go"

if [ -e $filename ]; then
    mv $filename $filename.bak
fi

cat > $filename << CODE
package main

import (
    "flag"
	"os"
    
    "github.com/arthurkiller/perfm"
)

type job struct {
    // job private data
}

func (j *job) String() string {
    // print out the job
    // you can leave this blank
    return ""
}

// Copy will called in parallel
func (j *job) Copy() (perfm.Job, error) {
    jc := *j
    return &jc, nil
}

func (j *job) Pre() error {
    // do pre job
    // you can leave this blank
    return nil
}
func (j *job) Do() (err error) {
    // do benchmark job

    return
}
func (j *job) After() {
    // do clean job
    // you can leave this blank
}

func main() {
    parallel := flag.Int("p", 5, "number of parallel")
    count := flag.Int("c", 0, "number of total tests times")
    duration := flag.Int("d", 10, "number of total tests times")
    bin := flag.Int("bin", 15, "number of histogram bins")
	flag.Parse()
	if !flag.Parsed() {
		flag.Usage()
		os.Exit(0)
	}
    
    var p perfm.PerfMonitor
    if *count == 0 {
    	p = perfm.New(perfm.WithBinsNumber(*bin), perfm.WithParallel(*parallel), perfm.WithDuration(*duration))
    } else {
    	p = perfm.New(perfm.WithBinsNumber(*bin), perfm.WithParallel(*parallel), perfm.WithNumber(*count))
    }
    j := job{}
    
    p.Regist(&j)
    p.Start()
    p.Wait()
}
CODE
