package main

import (
	"net/http"

	"github.com/arthurkiller/perfm"
)

func main() {

	perfm := perfm.New(perfm.WithBinsNumber(15), perfm.WithMinValue(0), perfm.WithGrowthFactor(0.4), perfm.WithBaseBucketSize(20), perfm.WithParallel(5))
	//perfm := perfm.New(perfm.WithBinsNumber(15), perfm.WithParallel(5), perfm.WithNumber(100))

	perfm.Registe(func() error {
		_, err := http.Get("http://www.baidu.com")
		return err
	})
	perfm.Start()
	perfm.Wait()
}
