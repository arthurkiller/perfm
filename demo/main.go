package main

import (
	"net/http"

	"github.com/arthurkiller/perfm"
)

func main() {
	conf := perfm.NewConfig(perfm.WithBinsNumber(15), perfm.WithMinValue(0),
		perfm.WithGrowthFactor(0.4), perfm.WithBaseBucketSize(20), perfm.WithParallel(5))

	perfm := perfm.New(conf)

	perfm.Registe(func() error {
		_, err := http.Get("http://www.baidu.com")
		return err
	})
	perfm.Start()
	perfm.Wait()
}
