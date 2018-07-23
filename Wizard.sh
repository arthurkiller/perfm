#!/bin/bash

filename="job.go"

if [ -e "job.go" ]; then{
    filename="job_d.go"
}

echo 'package main

import "github.com/arthurkiller/perfm"

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

    # This Must be implemented !

    return
}
func (j *job) After() {
    // do clean job
    // you can leave this blank
}

func main() {
p := perfm.New(perfm.WithBinsNumber(15), perfm.WithParallel(5), perfm.WithDuration(10))
j := job{}

p.Regist(&j)
p.Start()
p.Wait()
}' >> $(filename)
