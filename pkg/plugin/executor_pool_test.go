package plugin

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

type dummyHandler struct{}

func (d *dummyHandler) Type() string { return "dummy" }

func (d *dummyHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: sctx.Input,
	}, nil
}

func (d *dummyHandler) ValidateConfig(config json.RawMessage) error {
	return nil
}

func TestPluginExecutorPool(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(&dummyHandler{})

	pool := NewPluginExecutorPool(reg, 2, 10)
	defer pool.Shutdown()

	resCh := make(chan PluginTaskResult, 1)
	job := PluginTaskJob{
		TaskType: "dummy",
		Context:  worker.StepContext{Input: json.RawMessage(`"hello"`)},
		ResultCh: resCh,
	}

	if !pool.Submit(job) {
		t.Fatal("failed submitting job to executor pool")
	}

	res := <-resCh
	if res.Error != nil {
		t.Fatalf("unexpected execution error: %v", res.Error)
	}

	if string(res.Result.Output) != `"hello"` {
		t.Errorf("expected "hello", got %s", string(res.Result.Output))
	}
}
