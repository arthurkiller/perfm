package main

import (
	"net/http"

	"github.com/arthurkiller/perfm"
)

func main() {
	perfm := perfm.New(perfm.WithBinsNumber(15), perfm.WithParallel(1), perfm.WithDuration(3))
	perfm.Registe(func() (err error) {
		_, err = http.Get("https://www.baudu.com")
		return err
	})
	perfm.Start()
	perfm.Wait()
}
