package events

import (
	"strings"
	"sync"
)

// TrieNode represents an element in the hierarchical topic routing tree.
type TrieNode struct {
	children    map[string]*TrieNode
	subscribers map[string]bool
	isWildcard  bool
}

// TopicRoutingTrie routes messages across hierarchical dot-delimited topics with wildcard support.
type TopicRoutingTrie struct {
	mu   sync.RWMutex
	root *TrieNode
}

// NewTopicRoutingTrie constructs a new prefix routing trie.
func NewTopicRoutingTrie() *TopicRoutingTrie {
	return &TopicRoutingTrie{
		root: &TrieNode{
			children:    make(map[string]*TrieNode),
			subscribers: make(map[string]bool),
		},
	}
}

// Subscribe registers a subscriber ID for a topic pattern (supports '*' segment wildcard and '#' multi-level wildcard).
func (t *TopicRoutingTrie) Subscribe(topicPattern, subscriberID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	parts := strings.Split(topicPattern, ".")
	curr := t.root

	for _, part := range parts {
		if _, ok := curr.children[part]; !ok {
			curr.children[part] = &TrieNode{
				children:    make(map[string]*TrieNode),
				subscribers: make(map[string]bool),
				isWildcard:  part == "*" || part == "#",
			}
		}
		curr = curr.children[part]
	}
	curr.subscribers[subscriberID] = true
}

// MatchTopic finds all subscribers matching a published topic name.
func (t *TopicRoutingTrie) MatchTopic(topic string) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	parts := strings.Split(topic, ".")
	results := make(map[string]bool)

	var search func(node *TrieNode, idx int)
	search = func(node *TrieNode, idx int) {
		if node == nil {
			return
		}

		// Check multi-level wildcard "#"
		if hashChild, ok := node.children["#"]; ok {
			for sub := range hashChild.subscribers {
				results[sub] = true
			}
		}

		if idx == len(parts) {
			for sub := range node.subscribers {
				results[sub] = true
			}
			return
		}

		part := parts[idx]

		// Exact match
		if exactChild, ok := node.children[part]; ok {
			search(exactChild, idx+1)
		}

		// Single segment wildcard "*"
		if starChild, ok := node.children["*"]; ok {
			search(starChild, idx+1)
		}
	}

	search(t.root, 0)

	matched := make([]string, 0, len(results))
	for sub := range results {
		matched = append(matched, sub)
	}
	return matched
}
