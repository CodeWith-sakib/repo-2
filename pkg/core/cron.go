package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type CronSchedule struct {
	Minute     []int
	Hour       []int
	DayOfMonth []int
	Month      []int
	DayOfWeek  []int
}

func ParseCron(spec string) (*CronSchedule, error) {
	fields := strings.Fields(spec)
	if len(fields) != 5 {
		return nil, fmt.Errorf("%w: cron expression must contain 5 fields", ErrValidationFailed)
	}

	minute, err := parseCronField(fields[0], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("minute field error: %w", err)
	}

	hour, err := parseCronField(fields[1], 0, 23)
	if err != nil {
		return nil, fmt.Errorf("hour field error: %w", err)
	}

	dom, err := parseCronField(fields[2], 1, 31)
	if err != nil {
		return nil, fmt.Errorf("day of month field error: %w", err)
	}

	month, err := parseCronField(fields[3], 1, 12)
	if err != nil {
		return nil, fmt.Errorf("month field error: %w", err)
	}

	dow, err := parseCronField(fields[4], 0, 6)
	if err != nil {
		return nil, fmt.Errorf("day of week field error: %w", err)
	}

	return &CronSchedule{
		Minute:     minute,
		Hour:       hour,
		DayOfMonth: dom,
		Month:      month,
		DayOfWeek:  dow,
	}, nil
}

func parseCronField(field string, min, max int) ([]int, error) {
	if field == "*" {
		res := make([]int, max-min+1)
		for i := range res {
			res[i] = min + i
		}
		return res, nil
	}

	parts := strings.Split(field, ",")
	valMap := make(map[int]bool)

	for _, part := range parts {
		if strings.Contains(part, "/") {
			subParts := strings.Split(part, "/")
			if len(subParts) != 2 {
				return nil, fmt.Errorf("invalid step notation %s", part)
			}
			step, err := strconv.Atoi(subParts[1])
			if err != nil || step <= 0 {
				return nil, fmt.Errorf("invalid step value %s", subParts[1])
			}

			start := min
			end := max
			if subParts[0] != "*" {
				start, err = strconv.Atoi(subParts[0])
				if err != nil {
					return nil, err
				}
			}

			for i := start; i <= end; i += step {
				if i >= min && i <= max {
					valMap[i] = true
				}
			}
		} else if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range notation %s", part)
			}
			start, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, err
			}
			end, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return nil, err
			}
			if start > end || start < min || end > max {
				return nil, fmt.Errorf("range out of bounds %d-%d", start, end)
			}
			for i := start; i <= end; i++ {
				valMap[i] = true
			}
		} else {
			val, err := strconv.Atoi(part)
			if err != nil {
				return nil, err
			}
			if val < min || val > max {
				return nil, fmt.Errorf("value %d out of bounds [%d, %d]", val, min, max)
			}
			valMap[val] = true
		}
	}

	result := make([]int, 0, len(valMap))
	for k := range valMap {
		result = append(result, k)
	}
	return result, nil
}

func (s *CronSchedule) Matches(t time.Time) bool {
	minute := t.Minute()
	hour := t.Hour()
	dom := t.Day()
	month := int(t.Month())
	dow := int(t.Weekday())

	return containsInt(s.Minute, minute) &&
		containsInt(s.Hour, hour) &&
		containsInt(s.DayOfMonth, dom) &&
		containsInt(s.Month, month) &&
		containsInt(s.DayOfWeek, dow)
}

func (s *CronSchedule) Next(from time.Time) time.Time {
	t := from.Truncate(time.Minute).Add(time.Minute)
	// Lookahead max 5 years
	deadline := from.Add(5 * 365 * 24 * time.Hour)

	for t.Before(deadline) {
		if s.Matches(t) {
			return t
		}
		t = t.Add(time.Minute)
	}
	return time.Time{}
}

func containsInt(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
