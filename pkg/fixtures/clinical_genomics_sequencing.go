package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// ClinicalGenomicsPipeline builds a clinical NGS whole genome sequencing variant calling DAG.
func ClinicalGenomicsPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_genomics_ngs"),
		TenantID:    "precision-medicine-lab",
		Name:        "Clinical Whole Genome NGS Variant Calling Pipeline",
		Version:     1,
		Description: "Illumina BCL demultiplexing, BWA-MEM GRCh38 alignment, SAMtools coordinate indexing, Picard duplicate marking, GATK HaplotypeCaller variant calling, Ensembl VEP annotation, and ACMG pathogenicity classification",
		Timeout:     180 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "bcl-fastq-demultiplex",
				TaskType: "demux",
			},
			{
				ID:        "bwa-mem-alignment",
				TaskType:  "bwa_align",
				DependsOn: []string{"bcl-fastq-demultiplex"},
			},
			{
				ID:        "coordinate-sort-and-index",
				TaskType:  "bam_sort",
				DependsOn: []string{"bwa-mem-alignment"},
			},
			{
				ID:        "picard-mark-duplicates",
				TaskType:  "mark_dups",
				DependsOn: []string{"coordinate-sort-and-index"},
			},
			{
				ID:        "gatk-haplotype-caller",
				TaskType:  "gatk_call",
				DependsOn: []string{"picard-mark-duplicates"},
			},
			{
				ID:        "vep-variant-annotation",
				TaskType:  "vep_annotate",
				DependsOn: []string{"gatk-haplotype-caller"},
			},
			{
				ID:        "acmg-clinical-classification",
				TaskType:  "acmg_classify",
				DependsOn: []string{"vep-variant-annotation"},
			},
		},
	}
}

// DemuxHandler parses raw binary Illumina BCL flowcell files into paired-end FASTQ reads.
type DemuxHandler struct{}

func (h *DemuxHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"flowcell_id":"FC-NOVASEQ-771","samples_demultiplexed":48,"total_reads":820000000,"q30_bases_pct":92.4}`),
	}, nil
}

// BWAAlignHandler aligns paired FASTQ reads against the human GRCh38 reference genome.
type BWAAlignHandler struct{}

func (h *BWAAlignHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"reference":"GRCh38_no_alt_analysis_set","reads_mapped_pct":99.2,"properly_paired_pct":98.1,"mean_mapq":58.4}`),
	}, nil
}

// BAMSortHandler sorts alignments into genomic coordinate order and outputs indexed BAM files.
type BAMSortHandler struct{}

func (h *BAMSortHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"sorted_bam_uri":"s3://genomics-lake/bams/patient-991.sorted.bam","indexed":true,"file_size_gb":68.4}`),
	}, nil
}

// MarkDupsHandler flags optical and PCR duplicate fragments to avoid PCR bias in variant frequency.
type MarkDupsHandler struct{}

func (h *MarkDupsHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"duplicate_reads_pct":8.4,"optical_duplicates":14200,"mean_target_coverage":42.5}`),
	}, nil
}

// GATKCallHandler runs GATK HaplotypeCaller in GVCF mode with local de novo reassembly of active regions.
type GATKCallHandler struct{}

func (h *GATKCallHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"raw_snvs":4250000,"raw_indels":810000,"ti_tv_ratio":2.12,"gvcf_uri":"s3://genomics-lake/vcfs/patient-991.g.vcf.gz"}`),
	}, nil
}

// VEPAnnotateHandler annotates genomic coordinates with HGVS nomenclature, gnomAD allele frequencies, and ClinVar records.
type VEPAnnotateHandler struct{}

func (h *VEPAnnotateHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"annotated_variants":5060000,"rare_variants_under_1pct":18420,"missense_count":12400,"nonsense_stop_gain":85}`),
	}, nil
}

// ACMGClassifyHandler classifies high-impact clinical variants following ACMG/AMP five-tier guidelines.
type ACMGClassifyHandler struct{}

func (h *ACMGClassifyHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	reportID := fmt.Sprintf("DX-GENOMICS-%d", time.Now().UnixNano())
	payload := fmt.Sprintf(`{"clinical_report_id":"%s","pathogenic_variants":1,"likely_pathogenic":2,"vus_variants":14,"findings":[{"gene":"BRCA1","hgvs":"c.68_69delAG","classification":"PATHOGENIC","inheritance":"Autosomal_Dominant"}]}`, reportID)
	return &worker.StepResult{Output: json.RawMessage(payload)}, nil
}
