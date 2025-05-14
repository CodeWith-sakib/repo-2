package fixtures

import (
	"encoding/json"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func NewOrderFulfillmentPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_order_fulfillment"),
		Name:        "E-Commerce Order Fulfillment & Payment Clearing",
		Description: "Multi-stage enterprise order processing with inventory reservation, payment authorization, fraud check, and dispatch notification.",
		Version:     1,
		Steps: []core.StepDefinition{
			{
				ID:       "validate_order",
				TaskType: "transform",
				Config:   json.RawMessage(`{"rules":["check_stock","verify_billing_address"]}`),
			},
			{
				ID:        "reserve_inventory",
				TaskType:  "sql",
				DependsOn: []string{"validate_order"},
				Config:    json.RawMessage(`{"query":"UPDATE inventory SET reserved = reserved + 1 WHERE sku = 'PROD-99'"}`),
			},
			{
				ID:        "fraud_detection_probe",
				TaskType:  "http",
				DependsOn: []string{"validate_order"},
				Config:    json.RawMessage(`{"method":"POST","url":"https://fraud.internal/score"}`),
			},
			{
				ID:        "charge_payment",
				TaskType:  "http",
				DependsOn: []string{"reserve_inventory", "fraud_detection_probe"},
				Config:    json.RawMessage(`{"method":"POST","url":"https://payment.gateway/v1/charge"}`),
			},
			{
				ID:        "generate_invoice",
				TaskType:  "shell",
				DependsOn: []string{"charge_payment"},
				Config:    json.RawMessage(`{"command":"generate-pdf --invoice-id INV-100"}`),
			},
			{
				ID:        "dispatch_order",
				TaskType:  "sql",
				DependsOn: []string{"generate_invoice"},
				Config:    json.RawMessage(`{"query":"INSERT INTO shipments (order_id, status) VALUES ('ORD-1', 'QUEUED')"}`),
			},
		},
	}
}
