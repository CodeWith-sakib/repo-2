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

	payloadBytes, err := json.Marshal(e.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling event payload: %w", err)
	}

	return &CloudEvent{
		SpecVersion:     "1.0",
		ID:              string(e.ID),
		Source:          source,
		Type:            fmt.Sprintf("io.kestrelflow.event.%s", e.Type),
		Subject:         string(e.RunID),
		Time:            e.Timestamp,
		DataContentType: "application/json",
		Data:            payloadBytes,
		Extensions: map[string]interface{}{
			"tenant": string(e.TenantID),
		},
	}, nil
}

func FromCloudEvent(ce *CloudEvent) (*core.Event, error) {
	if ce == nil || ce.SpecVersion != "1.0" {
		return nil, fmt.Errorf("unsupported or nil CloudEvent specification version")
	}

	var payload map[string]interface{}
	if len(ce.Data) > 0 {
		if err := json.Unmarshal(ce.Data, &payload); err != nil {
			return nil, fmt.Errorf("failed unmarshaling cloudevent data: %w", err)
		}
	}

	tenant, _ := ce.Extensions["tenant"].(string)

	return &core.Event{
		ID:        core.ID(ce.ID),
		RunID:     core.ID(ce.Subject),
		TenantID:  tenant,
		Timestamp: ce.Time,
		Payload:   payload,
	}, nil
}
