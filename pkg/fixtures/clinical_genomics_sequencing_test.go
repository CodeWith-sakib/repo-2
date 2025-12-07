package fixtures

import (
	"context"
	"testing"

	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

func TestClinicalGenomicsPipeline_Validate(t *testing.T) {
	pipeline := ClinicalGenomicsPipeline()
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("pipeline validation failed: %v", err)
	}
}

func TestClinicalGenomicsPipeline_AllHandlers(t *testing.T) {
	ctx := context.Background()
	sctx := worker.StepContext{RunID: "run-dna-01", StepID: "bcl-fastq-demultiplex"}

	handlers := []struct {
		name    string
		handler worker.TaskExecutor
	}{
		{"demux", &DemuxHandler{}},
		{"align", &BWAAlignHandler{}},
		{"sort", &BAMSortHandler{}},
		{"mark_dups", &MarkDupsHandler{}},
		{"gatk", &GATKCallHandler{}},
		{"vep", &VEPAnnotateHandler{}},
		{"acmg", &ACMGClassifyHandler{}},
	}

	for _, h := range handlers {
		res, err := h.handler.Execute(ctx, sctx)
		if err != nil {
			t.Fatalf("handler %s failed: %v", h.name, err)
		}
		if len(res.Output) == 0 {
			t.Errorf("handler %s returned empty output", h.name)
		}
	}
}
