package fixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

// TelecomBillingPipeline builds an enterprise telecom CDR mediation and invoicing DAG.
func TelecomBillingPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_telecom_billing"),
		TenantID:    "telco-global",
		Name:        "Telecom Billing & CDR Mediation Pipeline",
		Version:     1,
		Description: "End-to-end Call Detail Record ingestion, rating, taxation, invoicing, and general ledger posting",
		Timeout:     60 * time.Second,
		Steps: []core.StepDefinition{
			{
				ID:       "cdr-ingest",
				TaskType: "cdr_ingest",
			},
			{
				ID:        "cdr-dedup",
				TaskType:  "dedup",
				DependsOn: []string{"cdr-ingest"},
			},
			{
				ID:        "tariff-rating",
				TaskType:  "rating",
				DependsOn: []string{"cdr-dedup"},
			},
			{
				ID:        "discount-engine",
				TaskType:  "discount",
				DependsOn: []string{"tariff-rating"},
			},
			{
				ID:        "tax-calculator",
				TaskType:  "taxation",
				DependsOn: []string{"discount-engine"},
			},
			{
				ID:        "invoice-generator",
				TaskType:  "invoicing",
				DependsOn: []string{"tax-calculator"},
			},
			{
				ID:        "gl-posting",
				TaskType:  "gl_post",
				DependsOn: []string{"invoice-generator"},
			},
		},
	}
}

// CDRIngestHandler processes raw ASN.1 / binary network switch records.
type CDRIngestHandler struct{}

func (h *CDRIngestHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	payload := fmt.Sprintf(`{"switch_id":"MSS-ORD-01","cdrs_read":250000,"format":"asn1_ber","ingested_at":"%s"}`, time.Now().UTC().Format(time.RFC3339))
	return &worker.StepResult{Output: json.RawMessage(payload)}, nil
}

// CDRDedupHandler removes duplicate transmission logs from overlapping cell towers.
type CDRDedupHandler struct{}

func (h *CDRDedupHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"total_cdrs":250000,"unique_cdrs":249820,"duplicates_dropped":180}`),
	}, nil
}

// TariffRatingHandler looks up calling zones and applies per-minute/megabyte rates.
type TariffRatingHandler struct{}

func (h *TariffRatingHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"rated_cdrs":249820,"voice_rated_usd":48291.40,"data_rated_usd":112930.12,"sms_rated_usd":3490.10}`),
	}, nil
}

// DiscountEngineHandler applies loyalty, volume, and family-plan bundle discounts.
type DiscountEngineHandler struct{}

func (h *DiscountEngineHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"subtotal_usd":164711.62,"discounts_applied_usd":24706.74,"net_rated_usd":140004.88}`),
	}, nil
}

// TaxCalculatorHandler computes federal, state, and local regulatory taxes (USF, E911).
type TaxCalculatorHandler struct{}

func (h *TaxCalculatorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"usf_tax_usd":9800.34,"state_sales_tax_usd":11200.39,"e911_fee_usd":1250.00,"total_tax_usd":22250.73}`),
	}, nil
}

// InvoiceGeneratorHandler compiles bill statements for individual and enterprise subscriber accounts.
type InvoiceGeneratorHandler struct{}

func (h *InvoiceGeneratorHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	return &worker.StepResult{
		Output: json.RawMessage(`{"invoices_created":14200,"total_billed_usd":162255.61,"pdf_storage":"s3://telco-invoices/2026-03/"}`),
	}, nil
}

// GLPostingHandler writes journal debit/credit lines into SAP/ERP financial ledgers.
type GLPostingHandler struct{}

func (h *GLPostingHandler) Execute(ctx context.Context, sctx worker.StepContext) (*worker.StepResult, error) {
	batchID := fmt.Sprintf("GL-TELCO-%d", time.Now().UnixNano())
	payload := fmt.Sprintf(`{"journal_batch_id":"%s","ar_debit_usd":162255.61,"revenue_credit_usd":140004.88,"tax_credit_usd":22250.73,"balanced":true}`, batchID)
	return &worker.StepResult{Output: json.RawMessage(payload)}, nil
}
