package core

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestDynamicFanOutCoordinator_Success(t *testing.T) {
	cfg := FanOutConfig{
		MaxParallelism: 4,
		Policy:         PolicyCollectAll,
	}
	coord := NewDynamicFanOutCoordinator(cfg)

	items := []interface{}{"alpha", "beta", "gamma"}
	processor := func(index int, item interface{}) (json.RawMessage, error) {
		str := item.(string)
		return json.RawMessage(fmt.Sprintf(`{"processed":%q}`, str)), nil
	}

	res, err := coord.ScatterGather(items, processor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalItems != 3 || res.Successful != 3 || res.Failed != 0 {
		t.Errorf("unexpected counts: total=%d succ=%d fail=%d", res.TotalItems, res.Successful, res.Failed)
	}
}

func TestDynamicFanOutCoordinator_FailFast(t *testing.T) {
	cfg := FanOutConfig{
		MaxParallelism: 1, // serial for deterministic stop
		Policy:         PolicyFailFast,
	}
	coord := NewDynamicFanOutCoordinator(cfg)

	items := []interface{}{1, 2, 3, 4}
	processor := func(index int, item interface{}) (json.RawMessage, error) {
		val := item.(int)
		if val == 2 {
			return nil, fmt.Errorf("item 2 rejected")
		}
		return json.RawMessage(`{}`), nil
	}

	_, err := coord.ScatterGather(items, processor)
	if err == nil {
		t.Fatal("expected error with PolicyFailFast")
	}
}

func TestDynamicFanOutCoordinator_IgnoreFailures(t *testing.T) {
	cfg := FanOutConfig{
		MaxParallelism: 2,
		Policy:         PolicyIgnoreFailures,
	}
	coord := NewDynamicFanOutCoordinator(cfg)

	items := []interface{}{"ok", "bad", "ok"}
	processor := func(index int, item interface{}) (json.RawMessage, error) {
		if item.(string) == "bad" {
			return nil, fmt.Errorf("corrupt")
		}
		return json.RawMessage(`{}`), nil
	}

	res, err := coord.ScatterGather(items, processor)
	if err != nil {
		t.Fatalf("expected nil error under PolicyIgnoreFailures, got %v", err)
	}
	if res.Successful != 2 || res.Failed != 1 {
		t.Errorf("expected 2 succ, 1 fail: got %d, %d", res.Successful, res.Failed)
	}
}
