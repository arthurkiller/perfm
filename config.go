package perfm

//Config define the Config about perfm
type Config struct {
	Duration       int     `json:"duration"`         // set for benchmark time in second
	Parallel       int     `json:"parallel"`         // test parallel worker numbers
	Number         int     `json:"number"`           // test total request
	NoPrint        bool    `json:"no_print"`         // disable print
	Frequency      int     `json:"frequency"`        // set for the sampling frequency
	BufferSize     int     `json:"buffer_size"`      // set for the global time channel buffer size
	BinsNumber     int     `json:"bins_number"`      // set the histogram bins number
	GrowthFactor   float64 `json:"growth_factor"`    // GrowthFactor is the growth factor of the buckets.
	MinValue       int64   `json:"min_value"`        // MinValue is the lower bound of the first bucket.
	BaseBucketSize float64 `json:"base_bucket_size"` // BaseBucketSize is the size of the first bucket.
	// A value of 0.1 indicates that bucket N+1 will be 10% larger than bucket N.
}

//NewConfig gen the config
func newConfig(options ...Options) Config {
	c := Config{10, 4, 0, false, 1, 6553500, 15, 1.4, 1000000, 1000000}
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
func WithNumber(i int) Options {
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

//WithBufferSize set the buffer size of config
func WithBufferSize(i int) Options {
	return func(o *Config) {
		o.BufferSize = i
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
