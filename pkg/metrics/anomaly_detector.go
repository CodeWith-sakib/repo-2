package metrics

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// AnomalySeverity indicates anomaly impact.
type AnomalySeverity string

const (
	SeverityNone     AnomalySeverity = "NONE"
	SeverityWarning  AnomalySeverity = "WARNING"
	SeverityCritical AnomalySeverity = "CRITICAL"
)

// AnomalyAssessment reports whether an observation is anomalous.
type AnomalyAssessment struct {
	Value       float64
	Mean        float64
	StdDev      float64
	ZScore      float64
	IsAnomaly   bool
	Severity    AnomalySeverity
	Description string
}

// AnomalyDetectorConfig sets anomaly detection thresholds.
type AnomalyDetectorConfig struct {
	MinSamples         int     // minimum warm-up observations before flagging
	WarningZThreshold  float64 // default 2.0 (95% confidence)
	CriticalZThreshold float64 // default 3.0 (99.7% confidence)
}

// DefaultAnomalyDetectorConfig returns standard 2-sigma / 3-sigma thresholds.
func DefaultAnomalyDetectorConfig() AnomalyDetectorConfig {
	return AnomalyDetectorConfig{
		MinSamples:         20,
		WarningZThreshold:  2.0,
		CriticalZThreshold: 3.0,
	}
}

// StreamingAnomalyDetector computes online mean and variance using Welford's algorithm.
type StreamingAnomalyDetector struct {
	mu           sync.RWMutex
	cfg          AnomalyDetectorConfig
	count        int64
	mean         float64
	m2           float64 // sum of squared differences from mean
	lastUpdated  time.Time
	anomalyCount int64
}

// NewStreamingAnomalyDetector initializes an online anomaly detector.
func NewStreamingAnomalyDetector(cfg AnomalyDetectorConfig) *StreamingAnomalyDetector {
	if cfg.MinSamples <= 0 {
		cfg.MinSamples = 10
	}
	if cfg.WarningZThreshold <= 0 {
		cfg.WarningZThreshold = 2.0
	}
	if cfg.CriticalZThreshold <= 0 {
		cfg.CriticalZThreshold = 3.0
	}
	return &StreamingAnomalyDetector{
		cfg: cfg,
	}
}

// Observe incorporates a new data point and evaluates if it is anomalous.
func (d *StreamingAnomalyDetector) Observe(val float64) AnomalyAssessment {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.count++
	d.lastUpdated = time.Now()

	// Prior statistics before this observation
	priorMean := d.mean
	priorStdDev := 0.0
	if d.count > 1 {
		priorStdDev = math.Sqrt(d.m2 / float64(d.count-1))
	}

	// Update Welford accumulator
	delta := val - d.mean
	d.mean += delta / float64(d.count)
	delta2 := val - d.mean
	d.m2 += delta * delta2

	// Evaluate anomaly against prior baseline if warmed up
	if d.count < int64(d.cfg.MinSamples) || priorStdDev == 0 {
		return AnomalyAssessment{
			Value:       val,
			Mean:        d.mean,
			StdDev:      priorStdDev,
			ZScore:      0,
			IsAnomaly:   false,
			Severity:    SeverityNone,
			Description: fmt.Sprintf("Warming up (%d/%d samples)", d.count, d.cfg.MinSamples),
		}
	}

	zscore := (val - priorMean) / priorStdDev
	absZ := math.Abs(zscore)

	assessment := AnomalyAssessment{
		Value:  val,
		Mean:   priorMean,
		StdDev: priorStdDev,
		ZScore: zscore,
	}

	if absZ >= d.cfg.CriticalZThreshold {
		assessment.IsAnomaly = true
		assessment.Severity = SeverityCritical
		assessment.Description = fmt.Sprintf("Critical anomaly detected: value %.2f is %.2f stddevs away from mean %.2f", val, zscore, priorMean)
		d.anomalyCount++
	} else if absZ >= d.cfg.WarningZThreshold {
		assessment.IsAnomaly = true
		assessment.Severity = SeverityWarning
		assessment.Description = fmt.Sprintf("Warning anomaly detected: value %.2f is %.2f stddevs away from mean %.2f", val, zscore, priorMean)
		d.anomalyCount++
	} else {
		assessment.Severity = SeverityNone
		assessment.Description = "Normal observation"
	}

	return assessment
}

// Stats returns running count, mean, and standard deviation.
func (d *StreamingAnomalyDetector) Stats() (count int64, mean, stdDev float64, anomalies int64) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	sd := 0.0
	if d.count > 1 {
		sd = math.Sqrt(d.m2 / float64(d.count-1))
	}
	return d.count, d.mean, sd, d.anomalyCount
}
