package events

import (
	"testing"
)

func TestConsumerLagMonitor(t *testing.T) {
	mon := NewConsumerLagMonitor()

	mon.RecordLag(0, 1000, 950) // lag = 50
	mon.RecordLag(1, 2500, 2400) // lag = 100

	if mon.TotalLag() != 150 {
		t.Errorf("expected 150 total lag, got %d", mon.TotalLag())
	}
}
