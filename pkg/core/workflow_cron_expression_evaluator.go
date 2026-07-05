package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidCronPattern = errors.New("invalid cron expression specification")
)

// CronScheduleRules holds parsed minute, hour, day-of-month, month, day-of-week rules.
type CronScheduleRules struct {
	Minutes []int
	Hours   []int
	Days    []int
	Months  []int
	DOW     []int
}

// CronExpressionEvaluator parses standard 5-part cron patterns and calculates subsequent trigger instants.
type CronExpressionEvaluator struct {
	mu sync.RWMutex
}

// NewCronExpressionEvaluator creates an evaluator.
func NewCronExpressionEvaluator() *CronExpressionEvaluator {
	return &CronExpressionEvaluator{}
}

// Parse converts a cron line "M H Dom Mon Dow" into schedule rules.
func (e *CronExpressionEvaluator) Parse(expr string) (*CronScheduleRules, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return nil, fmt.Errorf("%w: expected 5 fields, got %d", ErrInvalidCronPattern, len(parts))
	}

	mins, err := parseCronSubField(parts[0], 0, 59)
	if err != nil {
		return nil, err
	}
	hrs, err := parseCronSubField(parts[1], 0, 23)
	if err != nil {
		return nil, err
	}
	dom, err := parseCronSubField(parts[2], 1, 31)
	if err != nil {
		return nil, err
	}
	mon, err := parseCronSubField(parts[3], 1, 12)
	if err != nil {
		return nil, err
	}
	dow, err := parseCronSubField(parts[4], 0, 6)
	if err != nil {
		return nil, err
	}

	return &CronScheduleRules{
		Minutes: mins,
		Hours:   hrs,
		Days:    dom,
		Months:  mon,
		DOW:     dow,
	}, nil
}

// NextTrigger calculates the earliest matching trigger time strictly after reference 'from'.
func (e *CronExpressionEvaluator) NextTrigger(spec *CronScheduleRules, from time.Time) time.Time {
	// Advance by 1 minute and clear seconds/nanoseconds
	t := from.Truncate(time.Minute).Add(time.Minute)

	for i := 0; i < 365*24*60; i++ { // search up to 1 year forward
		if matchIntInSlice(spec.Months, int(t.Month())) &&
			matchIntInSlice(spec.Days, t.Day()) &&
			matchIntInSlice(spec.DOW, int(t.Weekday())) &&
			matchIntInSlice(spec.Hours, t.Hour()) &&
			matchIntInSlice(spec.Minutes, t.Minute()) {
			return t
		}
		t = t.Add(time.Minute)
	}
	return time.Time{}
}

func parseCronSubField(field string, min, max int) ([]int, error) {
	if field == "*" {
		res := make([]int, max-min+1)
		for i := range res {
			res[i] = min + i
		}
		return res, nil
	}

	if strings.HasPrefix(field, "*/") {
		step, err := strconv.Atoi(field[2:])
		if err != nil || step <= 0 {
			return nil, ErrInvalidCronPattern
		}
		var res []int
		for i := min; i <= max; i += step {
			res = append(res, i)
		}
		return res, nil
	}

	val, err := strconv.Atoi(field)
	if err != nil || val < min || val > max {
		return nil, fmt.Errorf("%w: value %s out of range [%d, %d]", ErrInvalidCronPattern, field, min, max)
	}
	return []int{val}, nil
}

func matchIntInSlice(allowed []int, val int) bool {
	for _, a := range allowed {
		if a == val {
			return true
		}
	}
	return false
}
