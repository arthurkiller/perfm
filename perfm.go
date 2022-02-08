package perfm

import (
	"fmt"
	"io"
	"sync"
	"time"

	hist "github.com/arthurkiller/perfm/histogram"
)

// Job give out a job for parallel call
// 1. start workers
// 		1. workers call job.Copy()
// 		2. for-loop do
// 			* job.Pre()
// 			* job.Do()
// 		3. after for-loop call job.After()
// 2. caculate the summary
type Job interface {
	// Copy will copy a job for parallel call
	Copy() (Job, error)
	// Pre will called before do
	Pre() error
	// Do contains the core job here
	Do() error
	// After contains the clean job after job done
	After()
}

//PerfMonitor define the atcion about perfmonitor
type PerfMonitor interface {
	Reset(Job) //regist the job to perfm
	Start()    //start the perf monitor
}

// New perfmonitor
func New(options ...Options) PerfMonitor {
	config := NewConfig(options...)
	return NewMonitor(config)
}

// BUFFERLEN set for duration channel length
const BUFFERLEN = 0x7FFFF

// Collector collect all perfm config and do the statistic
type Collector struct {
	Sum   float64 //Sum of the per request cost
	Stdev float64 //Standard Deviation
	Mean  float64 //Mean about distribution
	Total int64   //total request by count

	conf          *Config
	wg            sync.WaitGroup
	durationCache chan int64             // duration cache buffer, wait for operation
	histogram     *hist.NumericHistogram // used to print the histogram
	done          chan struct{}          // close channel

	localtimer     <-chan time.Time // print timer
	localCount     int64            // count for the number in the sampling times
	localTimeCount int64            // count for the sampling time total costs
}

//Config define the Config about perfm
type Config struct {
	// manager config
	Duration  int   `json:"duration"`  // benchmark duration in second
	Number    int64 `json:"number"`    // total requests
	Parallel  int   `json:"parallel"`  // parallel worker numbers
	NoPrint   bool  `json:"no_print"`  // disable statistic print
	Frequency int   `json:"frequency"` // sampling frequency, control the precision

	// collector config
	BinsNumber int `json:"bins_number"` // set the histogram bins number

	// XXX
	//GrowthFactor   float64 `json:"growth_factor"`    // GrowthFactor is the growth factor of the buckets.
	//MinValue       int64   `json:"min_value"`        // MinValue is the lower bound of the first bucket.
	//BaseBucketSize float64 `json:"base_bucket_size"` // BaseBucketSize is the size of the first bucket.
	//BufferSize     int     `json:"buffer_size"`      // set for the global time channel buffer size
	// A value of 0.1 indicates that bucket N+1 will be 10% larger than bucket N.
}

//NewConfig gen the config
func NewConfig(options ...Options) Config {
	c := Config{
		Duration:   10,
		Number:     0,
		Parallel:   4,
		NoPrint:    false,
		Frequency:  1,
		BinsNumber: 15,
		//GrowthFactor:   1.4,
		//MinValue:       1000000,
		//BaseBucketSize: 1000000,
	}
	for _, o := range options {
		o(&c)
	}
	return c
}

//Options define the options of congif
type Options func(*Config)

//WithParallel set the workers
func WithParallel(i int) Options {
	return func(o *Config) {
		o.Parallel = i
	}
}

//WithDuration set the test running duration
func WithDuration(i int) Options {
	return func(o *Config) {
		o.Duration = i
	}
}

//WithNumber set the total benchmark request
func WithNumber(i int64) Options {
	return func(o *Config) {
		o.Number = i
	}
}

//WithFrequency set the frequency
func WithFrequency(i int) Options {
	return func(o *Config) {
		o.Frequency = i
	}
}

//WithBinsNumber set the bins number of config
func WithBinsNumber(i int) Options {
	return func(o *Config) {
		o.BinsNumber = i
	}
}

//WithNoPrint will disable output during benchmarking
func WithNoPrint() Options {
	return func(o *Config) {
		o.NoPrint = true
	}
}

// NewCollector create collector
// 1. create collector
// 2. run the goroutine monitor for duration
// 3. do the collection
func NewCollector(c *Config) *Collector {
	cc := &Collector{
		wg:            sync.WaitGroup{},
		durationCache: make(chan int64, BUFFERLEN),
		localtimer:    time.NewTicker(time.Second * time.Duration(c.Frequency)).C,
		histogram:     hist.NewHistogram(c.BinsNumber),
		done:          make(chan struct{}),
	}
	return cc
}

// Start the collector
func (c *Collector) Start() {
	c.wg.Add(1) // add wg, and wait for goroutine start
	go c.run()
	c.wg.Wait()
	c.wg.Add(1) // add wg and wait goroutine successfully stopped
}

func (c *Collector) run() {
	var cost int64
	c.wg.Done() // generate new collector goroutine, makesure it has started
	for {
		select {
		case cost = <-c.durationCache: // on collection, main operation
			c.Total++
			c.localCount++
			c.localTimeCount += cost
			c.histogram.Add(cost)

		case <-c.localtimer: // print timer per second
			if c.localCount == 0 {
				continue
			}
			if !c.conf.NoPrint {
				fmt.Println(c)
			}
			c.localCount = 0
			c.localTimeCount = 0

		case <-c.done: // close notify channel
			for cost := range c.durationCache {
				c.Total++
				c.localCount++
				c.localTimeCount += cost
				c.histogram.Add(cost)
			}
			if !c.conf.NoPrint {
				fmt.Println(c)
			}
			c.wg.Done() // signal wg done on exiting
			return
		}
	}
}

func (c *Collector) String() string {
	return fmt.Sprintf("%s \t  Qps: %d \t  Avg Latency: %.3fms", time.Now().Format("15:04:05.000"),
		c.localCount, float64(c.localTimeCount/c.localCount)/1000000)
}

// WaitStop will consume all
func (c *Collector) WaitStop() {
	close(c.done)
	c.wg.Wait()
}

func (c *Collector) PrintResult(io.Writer) {
	fmt.Println("\n==================SUMMARIZE=======================")
	// here show the histogram
	fmt.Printf("MAX: %.3dms MIN: %.3dms MEAN: %.3dms STDEV: %.3f CV: %.3f%% ",
		c.histogram.Max()/1000000, c.histogram.Min()/1000000, c.histogram.Mean()/1000000,
		c.histogram.STDEV()/1000000, c.histogram.CV())

	// print histogram chart
	fmt.Println(c.histogram)

	fmt.Println("===============================================")
	fmt.Printf("Summary:\n70%% in:\t%.3dms\n80%% in:\t%.3dms\n90%% in:\t%.3dms\n95%% in:\t%.3dms\n99%% in:\t%.3dms",
		c.histogram.Quantile(0.7)/1000000, c.histogram.Quantile(0.8)/1000000, c.histogram.Quantile(0.9)/1000000,
		c.histogram.Quantile(0.95)/1000000, c.histogram.Quantile(0.99)/1000000)
}

// Collect a time duration and add to histogram
func (c *Collector) Collect(d time.Duration) {
	c.durationCache <- int64(d)
}
