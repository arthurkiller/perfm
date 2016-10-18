package perfM

//Config define the Config about perfM
type Config struct {
	Frequency  int //set for the sampling frequency
	BufferSize int //set for the global time channel buffer size
	BinsNumber int //set the histogram bins number
}

//new gen the config
func NewConfig(options ...Options) Config {
	c := Config{}
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

//WithBuffersize set the buffer size of config
func WithBufferSize(i int) Options {
	return func(o *Config) {
		o.BufferSize = i
	}
}

//Withbinsnumber set the bins number of config
func WithBinsNumber(i int) Options {
	return func(o *Config) {
		o.BinsNumber = i
	}
}
