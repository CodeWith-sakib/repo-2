package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type ID string

func NewID(prefix string) ID {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	if prefix == "" {
		return ID(hex.EncodeToString(b))
	}
	return ID(fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b)))
}

func (id ID) String() string {
	return string(id)
}

func (id ID) IsEmpty() bool {
	return len(id) == 0
}

type Metadata map[string]string

func (m Metadata) Clone() Metadata {
	if m == nil {
		return nil
	}
	c := make(Metadata, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

type JSONPayload []byte

func (p JSONPayload) Clone() JSONPayload {
	if p == nil {
		return nil
	}
	c := make(JSONPayload, len(p))
	copy(c, p)
	return c
}

type Priority int

const (
	PriorityLow      Priority = 10
	PriorityNormal   Priority = 50
	PriorityHigh     Priority = 80
	PriorityCritical Priority = 100
)

type TimeRange struct {
	Start time.Time
	End   time.Time
}

func (tr TimeRange) Contains(t time.Time) bool {
	if !tr.Start.IsZero() && t.Before(tr.Start) {
		return false
	}
	if !tr.End.IsZero() && t.After(tr.End) {
		return false
	}
	return true
}
