# perfm [![Build Status](https://travis-ci.org/arthurkiller/perfm.svg?branch=master)](https://travis-ci.org/arthurkiller/perfm) [![Go Report Card](https://goreportcard.com/badge/github.com/arthurkiller/perfm)](https://goreportcard.com/report/github.com/arthurkiller/perfm)
a golang performence testing platform

## what's new
* v1.0 is comming out
* reconstruct the project for easy use
* remove the divided operation for bench and caculate
* only 4 line of code can create a benchmarker, wow

## demo client
```go
	perfm := perfm.New(perfm.NewConfig())
	perfm.Registe(func() error {
		_, err := http.Get("http://www.baidu.com")
		return err
	})
	perfm.Start()
	perfm.Wait()

```
![test demo](./demo/screen.png)

## Milestone
* version 0.1 
    support the qps and average cost counting
* version 1.0
    change the perfm into a testing interface, just rejuest and start, the test will be automaticly done
