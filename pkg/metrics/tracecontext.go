package metrics

import (
	"fmt"
	"strings"
)

type Traceparent struct {
	Version  string
	TraceID  string
	ParentID string
	Flags    string
}

func ParseTraceparent(header string) (*Traceparent, error) {
	parts := strings.Split(header, "-")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid traceparent format: %s", header)
	}

	if len(parts[0]) != 2 || len(parts[1]) != 32 || len(parts[2]) != 16 || len(parts[3]) != 2 {
		return nil, fmt.Errorf("invalid segment length in traceparent: %s", header)
	}

	return &Traceparent{
		Version:  parts[0],
		TraceID:  parts[1],
		ParentID: parts[2],
		Flags:    parts[3],
	}, nil
}

func (t *Traceparent) String() string {
	return fmt.Sprintf("%s-%s-%s-%s", t.Version, t.TraceID, t.ParentID, t.Flags)
}
