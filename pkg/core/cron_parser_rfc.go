package core

import (
	"fmt"
	"strings"
	"time"
)

type ParsedCronSchedule struct {
	Spec     string
	Minutes  []int
	Hours    []int
	Days     []int
	Months   []int
	Weekdays []int
}

func (s *ParsedCronSchedule) Matches(t time.Time) bool {
	t = t.UTC()
	if len(s.Minutes) > 0 && !containsInt(s.Minutes, t.Minute()) {
		return false
	}
	if len(s.Hours) > 0 && !containsInt(s.Hours, t.Hour()) {
		return false
	}
	if len(s.Days) > 0 && !containsInt(s.Days, t.Day()) {
		return false
	}
	if len(s.Months) > 0 && !containsInt(s.Months, int(t.Month())) {
		return false
	}
	weekday := int(t.Weekday())
	if len(s.Weekdays) > 0 && !containsInt(s.Weekdays, weekday) {
		return false
	}
	return true
}

func ParseCronSpec(spec string) (*ParsedCronSchedule, error) {
	fields := strings.Fields(spec)
	if len(fields) != 5 {
		return nil, fmt.Errorf("cron spec must have 5 fields, got %d", len(fields))
	}
	return &ParsedCronSchedule{
		Spec: spec,
	}, nil
}
