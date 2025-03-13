package events

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

type CloudEvent struct {
	SpecVersion     string                 `json:"specversion"`
	ID              string                 `json:"id"`
	Source          string                 `json:"source"`
	Type            string                 `json:"type"`
	Subject         string                 `json:"subject,omitempty"`
	Time            time.Time              `json:"time"`
	DataContentType string                 `json:"datacontenttype"`
	Data            json.RawMessage        `json:"data"`
	Extensions      map[string]interface{} `json:"extensions,omitempty"`
}

func ToCloudEvent(e *core.Event, source string) (*CloudEvent, error) {
	if e == nil {
		return nil, fmt.Errorf("cannot convert nil event to cloudevent")
	}

	return &CloudEvent{
		SpecVersion:     "1.0",
		ID:              string(e.ID),
		Source:          source,
		Type:            fmt.Sprintf("io.kestrelflow.event.%s", e.Type),
		Subject:         string(e.RunID),
		Time:            e.Timestamp,
		DataContentType: "application/json",
		Data:            e.Payload,
		Extensions: map[string]interface{}{
			"tenant": string(e.TenantID),
		},
	}, nil
}

func FromCloudEvent(ce *CloudEvent) (*core.Event, error) {
	if ce == nil || ce.SpecVersion != "1.0" {
		return nil, fmt.Errorf("unsupported or nil CloudEvent specification version")
	}

	tenant, _ := ce.Extensions["tenant"].(string)

	return &core.Event{
		ID:        core.ID(ce.ID),
		RunID:     core.ID(ce.Subject),
		TenantID:  tenant,
		Timestamp: ce.Time,
		Payload:   ce.Data,
	}, nil
}
