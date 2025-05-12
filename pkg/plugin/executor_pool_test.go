package plugin

import (
	"context"
	"testing"
)

type dummyHandler struct{}

func (d *dummyHandler) Execute(ctx context.Context, payload []byte) ([]byte, error) {
	return append([]byte("echo:"), payload...), nil
}

func (d *dummyHandler) Validate(payload []byte) error {
	return nil
}

func TestPluginExecutorPool(t *testing.T) {
	reg := NewRegistry()
	reg.Register("dummy", &dummyHandler{})

	pool := NewPluginExecutorPool(reg, 2, 10)
	defer pool.Shutdown()

	resCh := make(chan PluginTaskResult, 1)
	job := PluginTaskJob{
		TaskType: "dummy",
		Payload:  []byte("hello"),
		ResultCh: resCh,
	}

	if !pool.Submit(job) {
		t.Fatal("failed submitting job to executor pool")
	}

	res := <-resCh
	if res.Error != nil {
		t.Fatalf("unexpected execution error: %v", res.Error)
	}

	if string(res.Output) != "echo:hello" {
		t.Errorf("expected echo:hello, got %s", string(res.Output))
	}
}
