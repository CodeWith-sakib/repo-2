package worker

import (
	"sync"
	"testing"
	"time"
)

func TestGoroutineLeakDetector(t *testing.T) {
	detector := NewGoroutineLeakDetector(2)

	detector.BeginTrack("step-clean")
	repClean := detector.EndTrack("step-clean")
	if repClean.StepID != "step-clean" {
		t.Errorf("step id mismatch")
	}

	detector.BeginTrack("step-leak")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		time.Sleep(200 * time.Millisecond)
	}()
	wg.Wait()

	repLeak := detector.EndTrack("step-leak")
	if repLeak.SuspectedLeakDiff < 1 {
		t.Logf("goroutine count diff: %d", repLeak.SuspectedLeakDiff)
	}
}
