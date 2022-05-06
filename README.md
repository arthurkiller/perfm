# perfm [![Build Status](https://travis-ci.org/arthurkiller/perfm.svg?branch=master)](https://travis-ci.org/arthurkiller/perfm) [![Go Report Card](https://goreportcard.com/badge/github.com/arthurkiller/perfm)](https://goreportcard.com/report/github.com/arthurkiller/perfm)
a golang performence testing platform

## What's new
* v3.0 is comming out
* reconstruct the project, make design clean and clear
* upgrade histogram implementation, now it can treate ___streaming data___ and caculate STDEV and CV

## What's in it

```
┌─────────────────────────────────────────────┐
│ ┌─Manager─────────────────┐      ┌───────┐  │
│ │                         │     ┌┴──────┐│  │
│ │                         │    ┌┴Workers││  │
│ │                         ├────►       │││  │
│ └──────────────────┬──────┘    │ Job   │││  │
│                    │           │       │├┘  │
│ ┌─Collector────────▼──────┐    │       ││   │
│ │                         ◄────┤       ├┘   │
│ │       Histogram         │    └───────┘    │
│ └─────────────────────────┘                 │
└─────────────────────────────────────────────┘
```

## Understand perfm Workflow

* a perfm job work like this...
```golang
for {
    job.Pre()

    start := time.Now()
    job.Do()
    count = time.Since(start)
}
job.After()
```


* perfm jobs work on different goroutine

```bash
    for parallels {
        job.Copy()
    }
```

```
 +---------+ +---------+ +---------+
 |   job   | |   job   | |   job   |
 |         | |         | |         |
 |for{     | |for{     | |for{     |
 | pre()   | | pre()   | | pre()   |
 | do()    | | do()    | | do()    | ... ...
 |}        | |}        | |}        |
 | after() | | after() | | after() |
 +---------+ +---------+ +---------+
```


## Short Example

___2 steps to start your benchmark!___

1. implement you own `perfm.Job`
2. call `perfm.Start(Job)`

You can start with `Wizard.sh` creating your job templates.

basic http benchmark job by example
```go
type job struct {
	// job private data
	url string
}

// Copy will called in parallel
func (j *job) Copy() (perfm.Job,error) {
	jc := *j
	return &jc, nil
}

func (j *job) Pre() error {
	// do pre job, prepare the data
    return nil
}
func (j *job) Do() error {
	// do benchmark job, the cost will be count
	_, err := http.Get(j.url)
	return err
}
func (j *job) After() {
	// do clean job, only called in the end of the job
}

// start perfm
perfm := perfm.New(perfm.WithBinsNumber(15), perfm.WithParallel(5), perfm.WithDuration(10))
j := &job{}
j.url = "http://www.baidu.com"
perfm.Start(j)
```

<img width="1358" alt="image" src="https://user-images.githubusercontent.com/11133870/167170304-6bb8a62f-4075-4409-8357-ada8a5974344.png">

## Milestone
* version 0.1
    support the qps and average cost counting
* version 1.0
    change the perfm into a testing interface, just rejuest and start, the test will be automaticly done
* version 2.0
    ~~add the excel/numbers .cvs file export. make it easy to draw graphic with other data processor.~~
* version 3.0
    * reconstruct on API and better support for streaming data.
	* add color print later
