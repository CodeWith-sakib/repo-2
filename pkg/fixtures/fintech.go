package fixtures

import (
	"encoding/json"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func NewFintechPaymentClearingPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_fintech_clearing"),
		Name:        "Automated Clearing House (ACH) & Wire Settlement Pipeline",
		Description: "Multi-party payment settlement workflow with AML sanctions screening, ledger double-entry reconciliation, and SWIFT message dispatch.",
		Version:     1,
		Steps: []core.StepDefinition{
			{
				ID:       "parse_iso20022_message",
				TaskType: "transform",
				Config:   json.RawMessage(`{"format":"pacs.008.001.09","strict":true}`),
			},
			{
				ID:        "aml_ofac_screening",
				TaskType:  "http",
				DependsOn: []string{"parse_iso20022_message"},
				Config:    json.RawMessage(`{"url":"https://compliance.fintech.internal/v2/screen","method":"POST"}`),
			},
			{
				ID:        "account_balance_hold",
				TaskType:  "sql",
				DependsOn: []string{"parse_iso20022_message"},
				Config:    json.RawMessage(`{"query":"UPDATE accounts SET hold_balance = hold_balance + :amount WHERE account_id = :src_acc"}`),
			},
			{
				ID:        "fraud_risk_score",
				TaskType:  "http",
				DependsOn: []string{"parse_iso20022_message"},
				Config:    json.RawMessage(`{"url":"https://fraud.fintech.internal/score","method":"POST"}`),
			},
			{
				ID:        "double_entry_ledger_journal",
				TaskType:  "sql",
				DependsOn: []string{"aml_ofac_screening", "account_balance_hold", "fraud_risk_score"},
				Config:    json.RawMessage(`{"query":"INSERT INTO journal_entries (tx_id, dr_account, cr_account, amount) VALUES (:tx, :dr, :cr, :amt)"}`),
			},
			{
				ID:        "generate_swift_mt103",
				TaskType:  "transform",
				DependsOn: []string{"double_entry_ledger_journal"},
				Config:    json.RawMessage(`{"template":"MT103_SINGLE_CUSTOMER_CREDIT"}`),
			},
			{
				ID:        "transmit_fedwire",
				TaskType:  "http",
				DependsOn: []string{"generate_swift_mt103"},
				Config:    json.RawMessage(`{"url":"https://gateway.fedwire.gov/v1/messages","method":"POST"}`),
			},
			{
				ID:        "notify_customer_sms_email",
				TaskType:  "http",
				DependsOn: []string{"transmit_fedwire"},
				Config:    json.RawMessage(`{"url":"https://notifications.fintech.internal/push","method":"POST"}`),
			},
		},
	}
}
