package events

import (
	"encoding/json"
	"fmt"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func SerializeEvent(e *core.Event) ([]byte, error) {
	if e == nil {
		return nil, fmt.Errorf("cannot serialize nil event")
	}
	return json.Marshal(e)
}

func DeserializeEvent(data []byte) (*core.Event, error) {
	var e core.Event
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, fmt.Errorf("failed to deserialize event: %w", err)
	}
	return &e, nil
}
