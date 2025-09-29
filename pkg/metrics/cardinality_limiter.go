package metrics

import (
	"sync"
	"sync/atomic"
)

// CardinalityConfig specifies thresholds for tag value explosion protection.
type CardinalityConfig struct {
	MaxDistinctPerTag int
	OverflowValue     string
}

// DefaultCardinalityConfig returns standard protection limits.
func DefaultCardinalityConfig() CardinalityConfig {
	return CardinalityConfig{
		MaxDistinctPerTag: 100,
		OverflowValue:     "__overflow__",
	}
}

// tagTracker keeps track of seen values for a specific tag key.
type tagTracker struct {
	mu       sync.RWMutex
	values   map[string]struct{}
	maxVal   int
	overflow string
}

func newTagTracker(maxVal int, overflow string) *tagTracker {
	return &tagTracker{
		values:   make(map[string]struct{}),
		maxVal:   maxVal,
		overflow: overflow,
	}
}

func (t *tagTracker) sanitize(val string) (string, bool) {
	t.mu.RLock()
	if _, ok := t.values[val]; ok {
		t.mu.RUnlock()
		return val, false
	}
	isFull := len(t.values) >= t.maxVal
	t.mu.RUnlock()

	if isFull {
		return t.overflow, true
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Double check after lock
	if _, ok := t.values[val]; ok {
		return val, false
	}
	if len(t.values) >= t.maxVal {
		return t.overflow, true
	}

	t.values[val] = struct{}{}
	return val, false
}

func (t *tagTracker) count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.values)
}

func (t *tagTracker) reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.values = make(map[string]struct{})
}

// CardinalityLimiter protects against metric cardinality explosions.
type CardinalityLimiter struct {
	mu           sync.RWMutex
	trackers     map[string]*tagTracker
	cfg          CardinalityConfig
	overflowHits atomic.Int64
}

// NewCardinalityLimiter initializes a limiter with the specified config.
func NewCardinalityLimiter(cfg CardinalityConfig) *CardinalityLimiter {
	if cfg.MaxDistinctPerTag <= 0 {
		cfg.MaxDistinctPerTag = 100
	}
	if cfg.OverflowValue == "" {
		cfg.OverflowValue = "__overflow__"
	}
	return &CardinalityLimiter{
		trackers: make(map[string]*tagTracker),
		cfg:      cfg,
	}
}

// SanitizeLabel normalizes a label value, returning the original or overflow string.
func (l *CardinalityLimiter) SanitizeLabel(key, val string) string {
	l.mu.RLock()
	tracker, ok := l.trackers[key]
	l.mu.RUnlock()

	if !ok {
		l.mu.Lock()
		if t, exists := l.trackers[key]; exists {
			tracker = t
		} else {
			tracker = newTagTracker(l.cfg.MaxDistinctPerTag, l.cfg.OverflowValue)
			l.trackers[key] = tracker
		}
		l.mu.Unlock()
	}

	sanitized, isOverflow := tracker.sanitize(val)
	if isOverflow {
		l.overflowHits.Add(1)
	}
	return sanitized
}

// SanitizeMap sanitizes all key/value pairs in a label map.
func (l *CardinalityLimiter) SanitizeMap(labels map[string]string) map[string]string {
	res := make(map[string]string, len(labels))
	for k, v := range labels {
		res[k] = l.SanitizeLabel(k, v)
	}
	return res
}

// DistinctCount returns how many distinct values have been registered for a tag key.
func (l *CardinalityLimiter) DistinctCount(key string) int {
	l.mu.RLock()
	tracker, ok := l.trackers[key]
	l.mu.RUnlock()
	if !ok {
		return 0
	}
	return tracker.count()
}

// OverflowHits returns total occurrences where values were mapped to overflow.
func (l *CardinalityLimiter) OverflowHits() int64 {
	return l.overflowHits.Load()
}

// Reset clears the tracker states.
func (l *CardinalityLimiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, t := range l.trackers {
		t.reset()
	}
	l.overflowHits.Store(0)
}
