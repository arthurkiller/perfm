package main

import (
	"net/http"

	"github.com/arthurkiller/perfm"
)

func main() {
	perfm := perfm.New(perfm.WithBinsNumber(15), perfm.WithParallel(5), perfm.WithDuration(10))

	perfm.Regist(func() (err error) {
		_, err = http.Get("http://www.baudu.com")
		return err
	})

	perfm.Start()
	perfm.Wait()
}
