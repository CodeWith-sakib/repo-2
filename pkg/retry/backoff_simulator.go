package retry

import (
	"math"
	"time"
)

type SimulationResult struct {
	Attempts       int
	TotalDuration  time.Duration
	MeanInterval   time.Duration
	MaxInterval    time.Duration
	VarianceNanos  float64
}

func SimulateBackoff(strategy BackoffStrategy, attempts int) SimulationResult {
	var total time.Duration
	var maxInterval time.Duration
	intervals := make([]time.Duration, attempts)

	for i := 0; i < attempts; i++ {
		d := strategy.NextInterval(i + 1)
		intervals[i] = d
		total += d
		if d > maxInterval {
			maxInterval = d
		}
	}

	meanNanos := float64(total.Nanoseconds()) / float64(attempts)
	var sumSquares float64
	for _, iv := range intervals {
		diff := float64(iv.Nanoseconds()) - meanNanos
		sumSquares += diff * diff
	}

	return SimulationResult{
		Attempts:      attempts,
		TotalDuration: total,
		MeanInterval:  time.Duration(meanNanos),
		MaxInterval:   maxInterval,
		VarianceNanos: sumSquares / float64(attempts),
	}
}

func StandardDeviation(res SimulationResult) time.Duration {
	return time.Duration(math.Sqrt(res.VarianceNanos))
}
