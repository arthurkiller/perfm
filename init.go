package perfm

// for basic useage
var p PerfMonitor

// Reset the job to perfm
func Reset(j Job) {
	p.Reset(j)
}

// Start the perf monitor
func Start(j Job) {
	p.Start(j)
}

func init() {
	p = New(WithBinsNumber(15), WithParallel(5), WithDuration(5))
}
