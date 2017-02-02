package perfm

//Config define the Config about perfm
type Config struct {
	Frequency  int //set for the sampling frequency
	BufferSize int //set for the global time channel buffer size
	BinsNumber int //set the histogram bins number

	// GrowthFactor is the growth factor of the buckets.
	// A value of 0.1 indicates that bucket N+1 will be 10% larger than bucket N.
	GrowthFactor float64
	// BaseBucketSize is the size of the first bucket.
	BaseBucketSize float64
	// MinValue is the lower bound of the first bucket.
	MinValue int64

	NoPrint bool
}

//NewConfig gen the config
func NewConfig(options ...Options) Config {
	c := Config{1, 655359, 20, 1.4, 1000, 100, false}
	for _, o := range options {
		o(&c)
	}
	return c
}

//Options define the options of congif
type Options func(*Config)

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

func WithGrowthFactor(i float64) Options {
	return func(o *Config) {
		o.GrowthFactor = i
	}
}
func WithBaseBucketSize(i float64) Options {
	return func(o *Config) {
		o.BaseBucketSize = i
	}
}
func WithMinValue(i int64) Options {
	return func(o *Config) {
		o.MinValue = i
	}
}
func WithNoPrint() Options {
	return func(o *Config) {
		o.NoPrint = true
	}
}
