package gohistogram

// Copyright (c) 2013 VividCortex, Inc. All rights reserved.
// Please see the LICENSE file for applicable license terms.

import (
	"fmt"
	"math"
)

type NumericHistogram struct {
	bins    []bin
	maxbins int

	total int64
	max   int64
	min   int64
	sum   int64
	sum2  int64 // sum up value pow 2
}

// NewHistogram returns a new NumericHistogram with a maximum of n bins.
//
// There is no "optimal" bin count, but somewhere between 20 and 80 bins
// should be sufficient.
func NewHistogram(n int) *NumericHistogram {
	return &NumericHistogram{
		bins:    make([]bin, 0),
		maxbins: n,
		total:   0,
		max:     -math.MaxInt64,
		min:     math.MaxInt64,
		sum:     0,
		sum2:    0,
	}
}

func (h *NumericHistogram) Add(n int64) {
	defer h.trim()
	h.total++
	for i := range h.bins {
		if h.bins[i].value == n {
			h.bins[i].count++
			return
		}

		if h.bins[i].value > n {

			newbin := bin{value: n, count: 1}
			head := append(make([]bin, 0), h.bins[0:i]...)

			head = append(head, newbin)
			tail := h.bins[i:]
			h.bins = append(head, tail...)
			return
		}
	}

	h.bins = append(h.bins, bin{count: 1, value: n})
	h.sum += n
	h.sum2 += n * n
	if n > h.max {
		h.max = n
	} else if n < h.min {
		h.min = n
	}
}

func (h *NumericHistogram) Quantile(q float64) int64 {
	count := q * float64(h.total)
	for i := range h.bins {
		count -= float64(h.bins[i].count)

		if count <= 0 {
			return h.bins[i].value
		}
	}

	return -1
}

// CDF returns the value of the cumulative distribution function at x
func (h *NumericHistogram) CDF(x int64) int64 {
	var count int64 = 0
	for i := range h.bins {
		if h.bins[i].value <= x {
			count += int64(h.bins[i].count)
		}
	}

	return count / int64(h.total)
}

// Mean returns the sample mean of the distribution
func (h *NumericHistogram) Mean() int64 {
	return h.sum / int64(h.total)
}

func (h *NumericHistogram) Max() int64 {
	return h.max
}

func (h *NumericHistogram) Min() int64 {
	return h.min
}

// STDEV for standard deviation
func (h *NumericHistogram) STDEV() float64 {
	return math.Sqrt(float64(h.sum2-2*h.Mean()*h.sum+h.total*h.Mean()*h.Mean()) / float64(h.total))
}

// CV for Coefficient of Variation
func (h *NumericHistogram) CV() float64 {
	return h.STDEV() / float64(h.Mean()) * 100
}

// Variance returns the variance of the distribution
func (h *NumericHistogram) Variance() int64 {
	if h.total == 0 {
		return 0
	}

	var sum int64
	mean := h.Mean()
	for i := range h.bins {
		sum += (h.bins[i].count * (h.bins[i].value - mean) * (h.bins[i].value - mean))
	}
	return sum / int64(h.total)
}

func (h *NumericHistogram) Count() int64 {
	return h.total
}

// trim merges adjacent bins to decrease the bin count to the maximum value
func (h *NumericHistogram) trim() {
	for len(h.bins) > h.maxbins {
		// Find closest bins in terms of value
		var minDelta int64 = math.MaxInt64
		minDeltaIndex := 0
		for i := range h.bins {
			if i == 0 {
				continue
			}

			if delta := h.bins[i].value - h.bins[i-1].value; delta < minDelta {
				minDelta = delta
				minDeltaIndex = i
			}
		}

		// We need to merge bins minDeltaIndex-1 and minDeltaIndex
		totalCount := h.bins[minDeltaIndex-1].count + h.bins[minDeltaIndex].count
		mergedbin := bin{
			value: (h.bins[minDeltaIndex-1].value*
				h.bins[minDeltaIndex-1].count +
				h.bins[minDeltaIndex].value*
					h.bins[minDeltaIndex].count) /
				totalCount, // weighted average
			count: totalCount, // summed heights
		}
		head := append(make([]bin, 0), h.bins[0:minDeltaIndex-1]...)
		tail := append([]bin{mergedbin}, h.bins[minDeltaIndex+1:]...)
		h.bins = append(head, tail...)
	}
}

// String returns a string reprentation of the histogram,
// which is useful for printing to a terminal.
func (h *NumericHistogram) String() string {
	var str string
	str += fmt.Sprintln("Total:", h.total)
	var cum int64
	for i := range h.bins {
		cum += h.bins[i].count
		var bar string
		for j := 0; j < int(int64(h.bins[i].count)/int64(h.total)*100); j++ {
			bar += "█"
		}
		str += fmt.Sprintf("%.3fms\t Count:%.0f[%v %.1f%%]\t %-v\n",
			h.bins[i].value/1000000, h.bins[i].count, cum, (int64(cum)/int64(h.total))*100, bar)
	}
	return str
}
