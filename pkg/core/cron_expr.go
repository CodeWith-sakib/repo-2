package core

import (
	"fmt"
	"strings"
	"time"
)

type CronScheduleDescriptor struct {
	Expression string `json:"expression"`
	TimeZone   string `json:"timezone"`
}

func DescribeCron(expr string) (string, error) {
	parts := strings.Fields(expr)
	if len(parts) != 5 {
		return "", fmt.Errorf("cron expression must have 5 fields, got %d", len(parts))
	}

	minute, hour, dom, month, dow := parts[0], parts[1], parts[2], parts[3], parts[4]
	var desc strings.Builder
	desc.WriteString("Runs ")

	if minute == "*" && hour == "*" {
		desc.WriteString("every minute")
	} else if minute == "0" && hour == "*" {
		desc.WriteString("every hour, at the start of the hour")
	} else if minute == "0" && hour == "0" {
		desc.WriteString("every day at midnight")
	} else {
		desc.WriteString(fmt.Sprintf("at %s:%s", hour, minute))
	}

	if dom != "*" {
		desc.WriteString(fmt.Sprintf(" on day of month %s", dom))
	}
	if month != "*" {
		desc.WriteString(fmt.Sprintf(" in month %s", month))
	}
	if dow != "*" {
		desc.WriteString(fmt.Sprintf(" on day of week %s", dow))
	}

	return desc.String(), nil
}

func NextNRuns(sched *CronSchedule, from time.Time, count int) []time.Time {
	var runs []time.Time
	curr := from
	for i := 0; i < count; i++ {
		nextRun := sched.Next(curr)
		if nextRun.IsZero() {
			break
		}
		runs = append(runs, nextRun)
		curr = nextRun
	}
	return runs
}
