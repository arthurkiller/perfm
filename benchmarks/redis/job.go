package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/arthurkiller/perfm"
	"github.com/gomodule/redigo/redis"

	_ "net/http/pprof"
)

type job struct {
	// job private data
	cli             redis.Conn
	rr              *rand.Rand
	splitedCommands []string
	needGen         []int
	needGen2        []int
	command         string

	host       string
	port       int
	auth       string
	randRange  int
	randRange2 int
	batching   int

	argv redis.Args
}

func (j *job) String() string {
	// print out the job
	// you can leave this blank
	return fmt.Sprintf("runing test on: %s", j.command)
}

// Copy will called in parallel
func (j *job) Copy() (perfm.Job, error) {
	jc := *j
	var err error
	var opts []redis.DialOption
	if jc.auth != "" {
		opts = append(opts, redis.DialPassword(jc.auth))
	}
	jc.cli, err = redis.Dial("tcp", fmt.Sprintf("%s:%d", jc.host, jc.port), opts...)

	// prepare destkey
	// without command, as the first command
	jc.command = j.splitedCommands[0]
	jc.argv = make(redis.Args, len(j.splitedCommands)-1)
	for i := 1; i < len(j.splitedCommands); i++ {
		jc.argv[i-1] = j.splitedCommands[i]
	}
	jc.rr = rand.New(rand.NewSource(time.Now().UnixNano()))

	return &jc, err
}

func (j *job) Pre() error {
	if j.batching > 0 {
		return j.preparePipline()
	}
	return j.prepareUnary()
}

func (j *job) prepareUnary() error {
	// prepare key
	if len(j.needGen) > 0 {
		// add fields
		for _, i := range j.needGen {
			j.argv[i-1] = j.splitedCommands[i] + fmt.Sprintf("%d", j.rr.Uint32()%uint32(j.randRange))
		}
	}

	if len(j.needGen2) > 0 {
		// add fields
		for _, i := range j.needGen2 {
			j.argv[i-1] = j.splitedCommands[i] + fmt.Sprintf("%d", j.rr.Uint32()%uint32(j.randRange2))
		}
	}
	return nil
}

func (j *job) preparePipline() (err error) {
	if err = j.cli.Send("MULTI"); err != nil {
		return
	}
	for i := 0; i < j.batching; i++ {
		j.prepareUnary()
		if err = j.cli.Send(j.command, j.argv...); err != nil {
			return
		}
	}
	return
}

func (j *job) Do() (err error) {
	if j.batching > 0 { //pipeline mode
		_, err = j.cli.Do("EXEC")
	} else { // unary mode
		_, err = j.cli.Do(j.command, j.argv...)
	}
	return
}

func (j *job) After() {
	j.cli.Close()
}

func (j *job) processCommand() {
	re := regexp.MustCompile(`__RAND__`)
	re2 := regexp.MustCompile(`__RAND2__`)
	rawArgs := strings.Split(j.command, " ")

	for i := 0; i < len(rawArgs); i++ {
		if rawArgs[i] == "" {
			rawArgs = append(rawArgs[:i], rawArgs[i+1:]...)
			i--
		}
	}

	splitedArgs := make([]string, len(rawArgs))
	var needGen []int
	var needGen2 []int
	splitedArgs[0] = rawArgs[0]

	for i := 1; i < len(rawArgs); i++ {
		if re.Match([]byte(rawArgs[i])) {
			splitedArgs[i] = re.ReplaceAllString(rawArgs[i], "")
			needGen = append(needGen, i)
		} else if re2.Match([]byte(rawArgs[i])) {
			splitedArgs[i] = re2.ReplaceAllString(rawArgs[i], "")
			needGen2 = append(needGen2, i)
		} else {
			splitedArgs[i] = rawArgs[i]
		}
	}
	j.splitedCommands, j.needGen, j.needGen2 = splitedArgs, needGen, needGen2
}

func main() {
	parallel := flag.Int("p", 4, "number of parallel")
	count := flag.Int64("c", 0, "number of total tests runned, this will dissable duration option")
	duration := flag.Int("d", 30, "testing duration in second")

	host := flag.String("h", "127.0.0.1", "cluster host")
	port := flag.Int("port", 6379, "cluster port")
	auth := flag.String("a", "", "cluster auth")
	command := flag.String("command", "tr.getbit foo __RAND__", "testing command, you can add __RAND__ or __RAND2__ as random field to the command, but for each field, only the last __RAND__ will be replaced")
	randRange := flag.Int("r", 100000000, "random range for __RAND__")
	randRange2 := flag.Int("r2", 100000000, "random range for __RAND2__")
	batching := flag.Int("batching", 0, "pipeline mode, control batching size inside MULTI")

	flag.Parse()
	if !flag.Parsed() {
		flag.Usage()
		os.Exit(0)
	}

	j := job{
		host:       *host,
		port:       *port,
		auth:       *auth,
		command:    *command,
		randRange:  *randRange,
		randRange2: *randRange2,
		batching:   *batching,
	}
	j.processCommand()

	var opts = []perfm.Options{perfm.WithBinsNumber(10), perfm.WithParallel(*parallel)}
	if *count != 0 {
		opts = append(opts, perfm.WithNumber(*count))
	} else {
		opts = append(opts, perfm.WithDuration(*duration))
	}

	perfm.New(opts...).Start(&j)
}
