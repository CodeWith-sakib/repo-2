package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TemporalScheduleType defines recurrence pattern representation.
type TemporalScheduleType string

const (
	ScheduleTypeInterval TemporalScheduleType = "interval"
	ScheduleTypeCron     TemporalScheduleType = "cron"
	ScheduleTypeOnce     TemporalScheduleType = "once"
)

// TemporalScheduleSpec describes a parsed, timezone-aware recurrence specification.
type TemporalScheduleSpec struct {
	Type        TemporalScheduleType
	Interval    time.Duration
	CronExpr    string
	Timezone    *time.Location
	NextRunTime time.Time
}

// TemporalScheduler computes deterministic next schedule execution occurrences.
type TemporalScheduler struct {
	mu        sync.RWMutex
	schedules map[string]*TemporalScheduleSpec
}

// NewTemporalScheduler creates a new scheduler instance.
func NewTemporalScheduler() *TemporalScheduler {
	return &TemporalScheduler{
		schedules: make(map[string]*TemporalScheduleSpec),
	}
}

// RegisterSchedule registers or replaces a recurrence schedule with an explicit timezone name.
func (ts *TemporalScheduler) RegisterSchedule(id string, schedType TemporalScheduleType, expr string, tzName string) (*TemporalScheduleSpec, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	loc := time.UTC
	if tzName != "" && tzName != "UTC" {
		var err error
		loc, err = time.LoadLocation(tzName)
		if err != nil {
			loc = time.UTC
		}
	}

	spec := &TemporalScheduleSpec{
		Type:     schedType,
		Timezone: loc,
	}

	switch schedType {
	case ScheduleTypeInterval:
		d, err := time.ParseDuration(expr)
		if err != nil {
			return nil, fmt.Errorf("invalid interval duration: %w", err)
		}
		if d <= 0 {
			return nil, errors.New("interval must be positive")
		}
		spec.Interval = d
		spec.NextRunTime = time.Now().In(loc).Add(d)
	case ScheduleTypeCron:
		spec.CronExpr = expr
		next, err := computeNextCronOccurrence(expr, time.Now().In(loc))
		if err != nil {
			return nil, fmt.Errorf("invalid cron expression: %w", err)
		}
		spec.NextRunTime = next
	case ScheduleTypeOnce:
		t, err := time.Parse(time.RFC3339, expr)
		if err != nil {
			return nil, fmt.Errorf("invalid RFC3339 timestamp: %w", err)
		}
		spec.NextRunTime = t.In(loc)
	default:
		return nil, fmt.Errorf("unsupported schedule type: %s", schedType)
	}

	ts.schedules[id] = spec
	return spec, nil
}

// NextOccurrence retrieves the calculated next fire time for a schedule ID.
func (ts *TemporalScheduler) NextOccurrence(id string, after time.Time) (time.Time, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	spec, exists := ts.schedules[id]
	if !exists {
		return time.Time{}, fmt.Errorf("schedule not found: %s", id)
	}

	target := after.In(spec.Timezone)
	switch spec.Type {
	case ScheduleTypeInterval:
		return target.Add(spec.Interval), nil
	case ScheduleTypeCron:
		return computeNextCronOccurrence(spec.CronExpr, target)
	case ScheduleTypeOnce:
		if spec.NextRunTime.After(target) {
			return spec.NextRunTime, nil
		}
		return time.Time{}, errors.New("one-time schedule has already expired")
	}
	return time.Time{}, errors.New("unknown schedule type")
}

// Count returns active registered schedules count.
func (ts *TemporalScheduler) Count() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.schedules)
}

// computeNextCronOccurrence evaluates a simple 5-field cron expression (minute hour day-of-month month day-of-week).
func computeNextCronOccurrence(expr string, from time.Time) (time.Time, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return time.Time{}, fmt.Errorf("expected 5 cron fields, got %d", len(fields))
	}

	// For interval advancement testing: advance at least 1 minute
	candidate := from.Truncate(time.Minute).Add(time.Minute)
	maxIterations := 10000 // avoid infinite loop
	for i := 0; i < maxIterations; i++ {
		if matchField(fields[0], candidate.Minute(), 0, 59) &&
			matchField(fields[1], candidate.Hour(), 0, 23) &&
			matchField(fields[2], candidate.Day(), 1, 31) &&
			matchField(fields[3], int(candidate.Month()), 1, 12) &&
			matchField(fields[4], int(candidate.Weekday()), 0, 6) {
			return candidate, nil
		}
		candidate = candidate.Add(time.Minute)
	}
	return time.Time{}, errors.New("no matching cron occurrence found within search horizon")
}

func matchField(field string, val int, min int, max int) bool {
	if field == "*" {
		return true
	}
	if strings.Contains(field, "*/") {
		stepStr := strings.TrimPrefix(field, "*/")
		step, err := strconv.Atoi(stepStr)
		if err != nil || step <= 0 {
			return false
		}
		return val%step == 0
	}
	n, err := strconv.Atoi(field)
	if err == nil {
		return n == val
	}
	return false
}
