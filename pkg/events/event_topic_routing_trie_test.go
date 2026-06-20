package events

import (
	"testing"
)

func TestTopicRoutingTrie(t *testing.T) {
	trie := NewTopicRoutingTrie()

	trie.Subscribe("telemetry.orders.created", "sub-exact")
	trie.Subscribe("telemetry.orders.*", "sub-wild-star")
	trie.Subscribe("telemetry.#", "sub-wild-hash")
	trie.Subscribe("alerts.critical", "sub-alerts")

	// Match exact + star + hash
	res1 := trie.MatchTopic("telemetry.orders.created")
	expected1 := map[string]bool{"sub-exact": true, "sub-wild-star": true, "sub-wild-hash": true}

	if len(res1) != len(expected1) {
		t.Fatalf("expected 3 subscribers, got %d (%v)", len(expected1), res1)
	}
	for _, id := range res1 {
		if !expected1[id] {
			t.Errorf("unexpected subscriber: %s", id)
		}
	}

	// Match star + hash but not exact
	res2 := trie.MatchTopic("telemetry.orders.updated")
	expected2 := map[string]bool{"sub-wild-star": true, "sub-wild-hash": true}
	if len(res2) != len(expected2) {
		t.Fatalf("expected 2 subscribers, got %d (%v)", len(expected2), res2)
	}

	// Alerts match
	res3 := trie.MatchTopic("alerts.critical")
	if len(res3) != 1 || res3[0] != "sub-alerts" {
		t.Errorf("expected only sub-alerts, got %v", res3)
	}
}
