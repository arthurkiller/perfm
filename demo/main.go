package main

import (
	"net/http"

	"github.com/arthurkiller/perfm"
)

func main() {
	perfm := perfm.New(perfm.WithBinsNumber(15), perfm.WithParallel(1), perfm.WithNumber(20))

	perfm.Regist(nil, func() (err error) {
		_, err = http.Get("http://oa.meitu.com")
		return err
	}, nil)

	perfm.Start()
	perfm.Wait()
}
