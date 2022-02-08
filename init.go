package perfm

// for basic useage
var p PerfMonitor

// Reset the job to perfm
func Reset(j Job) {
	p.Reset(j)
}

// Start the perf monitor
func Start() {
	p.Start()
}

func init() {
	p = New(WithBinsNumber(15), WithParallel(5), WithDuration(5))
}
