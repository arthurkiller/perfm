package perfM

//Config define the Config about perfM
type Config struct {
	Frequency  int //set for the sampling frequency
	BufferSize int //set for the global time channel buffer size
	BinsNumber int //set the histogram bins number
}
