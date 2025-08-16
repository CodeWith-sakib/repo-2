package worker

import (
	"fmt"
	"sync"
	"time"
)

// HandoffState represents the phase of a cooperative task handoff.
type HandoffState string

const (
	HandoffStateInitiated    HandoffState = "initiated"
	HandoffStateAcknowledged HandoffState = "acknowledged"
	HandoffStateCommitted    HandoffState = "committed"
	HandoffStateAborted      HandoffState = "aborted"
)

// HandoffMessage conveys state from source worker to target worker during graceful drain/migration.
type HandoffMessage struct {
	TaskID      string
	SourceNode  string
	TargetNode  string
	StepPayload map[string]interface{}
	FenceToken  uint64
	InitiatedAt time.Time
	State       HandoffState
}

// CooperativeHandoffCoordinator coordinates zero-loss migration of executing tasks between worker instances.
type CooperativeHandoffCoordinator struct {
	mu       sync.Mutex
	handoffs map[string]*HandoffMessage
	nodeID   string
	timeout  time.Duration
}

// NewCooperativeHandoffCoordinator creates a handoff manager for a node.
func NewCooperativeHandoffCoordinator(nodeID string, timeout time.Duration) *CooperativeHandoffCoordinator {
	return &CooperativeHandoffCoordinator{
		handoffs: make(map[string]*HandoffMessage),
		nodeID:   nodeID,
		timeout:  timeout,
	}
}

// InitiateHandoff begins task migration from this node to targetNode.
func (c *CooperativeHandoffCoordinator) InitiateHandoff(taskID string, targetNode string, payload map[string]interface{}, fenceToken uint64) (*HandoffMessage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.handoffs[taskID]; exists {
		return nil, fmt.Errorf("handoff already active for task %s", taskID)
	}

	msg := &HandoffMessage{
		TaskID:      taskID,
		SourceNode:  c.nodeID,
		TargetNode:  targetNode,
		StepPayload: payload,
		FenceToken:  fenceToken,
		InitiatedAt: time.Now(),
		State:       HandoffStateInitiated,
	}
	c.handoffs[taskID] = msg
	return msg, nil
}

// AcknowledgeHandoff is called by the target node to accept responsibility.
func (c *CooperativeHandoffCoordinator) AcknowledgeHandoff(taskID string, targetNode string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	msg, exists := c.handoffs[taskID]
	if !exists {
		return fmt.Errorf("handoff not found for task %s", taskID)
	}
	if msg.TargetNode != targetNode {
		return fmt.Errorf("target node mismatch: expected %s, got %s", msg.TargetNode, targetNode)
	}
	if msg.State != HandoffStateInitiated {
		return fmt.Errorf("cannot acknowledge handoff in state %s", msg.State)
	}

	msg.State = HandoffStateAcknowledged
	return nil
}

// CommitHandoff marks the migration complete, allowing source worker to release local buffers.
func (c *CooperativeHandoffCoordinator) CommitHandoff(taskID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	msg, exists := c.handoffs[taskID]
	if !exists {
		return fmt.Errorf("handoff not found for task %s", taskID)
	}
	if msg.State != HandoffStateAcknowledged {
		return fmt.Errorf("cannot commit handoff from state %s", msg.State)
	}

	msg.State = HandoffStateCommitted
	return nil
}

// GetHandoff retrieves current handoff record.
func (c *CooperativeHandoffCoordinator) GetHandoff(taskID string) (*HandoffMessage, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	msg, exists := c.handoffs[taskID]
	return msg, exists
}
