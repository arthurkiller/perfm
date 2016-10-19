package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/arthurkiller/perfM"
)

func main() {
	f, err := os.Open("./data")
	defer f.Close()
	if err != nil {
		log.Println(err)
		return
	}

	br := bufio.NewReader(f)
	wg := new(sync.WaitGroup)

	conf := perfM.NewConfig(perfM.WithBinsNumber(10))
	perfm := perfM.New(conf)
	go perfm.Start()

	for {
		s, err := br.ReadString('\n')
		s = strings.Trim(s, "\n")
		if err != nil {
			if err != io.EOF {
				log.Println(err)
				return
			}
			break
		}

		i, err := strconv.Atoi(s)
		if err != nil {
			fmt.Println(err)
		}
		i %= 10
		wg.Add(1)
		go func() {
			t := perfm.Do()
			time.Sleep(200 * time.Millisecond * time.Duration(i))
			t.Done(perfm)
			defer wg.Done()
		}()
	}

	wg.Wait()
	perfm.Stop()
}
