package perfm

var p PerfMonitor

func init() {
	p = New(WithBinsNumber(15), WithParallel(5), WithDuration(5))
}
