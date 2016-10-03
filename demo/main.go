package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"strconv"
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
	perfm := perfM.New(perfM.Config{})
	go perfm.Start()

	for {
		s, err := br.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Println(err)
				return
			}
			break
		}
		i, _ := strconv.Atoi(s)
		i %= 10
		wg.Add(1)
		go func() {
			t := perfm.Do()
			time.Sleep(time.Millisecond * time.Duration(i))
			t.Done(perfm)
			defer wg.Done()
		}()
	}

	wg.Wait()
	perfm.Stop()
}
