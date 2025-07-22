package fixtures

import (
	"encoding/json"

	"github.com/kestrelflow/kestrelflow/pkg/core"
)

func NewHealthcareHL7IngestionPipeline() *core.WorkflowDefinition {
	return &core.WorkflowDefinition{
		ID:          core.NewID("wf_healthcare_hl7"),
		Name:        "HL7 v2 & FHIR Electronic Health Record Ingestion",
		Description: "HIPAA-compliant EHR clinical telemetry ingestion, patient identifier de-identification, and FHIR resource generation.",
		Version:     1,
		Steps: []core.StepDefinition{
			{
				ID:       "receive_mllp_packet",
				TaskType: "transform",
				Config:   json.RawMessage(`{"protocol":"mllp","version":"2.5.1"}`),
			},
			{
				ID:        "deidentify_phi",
				TaskType:  "transform",
				DependsOn: []string{"receive_mllp_packet"},
				Config:    json.RawMessage(`{"anonymization":"safe_harbor","hash_salt":"vault:phi-salt"}`),
			},
			{
				ID:        "map_to_fhir_resources",
				TaskType:  "transform",
				DependsOn: []string{"deidentify_phi"},
				Config:    json.RawMessage(`{"resources":["Patient","Observation","Encounter"]}`),
			},
			{
				ID:        "persist_clinical_lakehouse",
				TaskType:  "sql",
				DependsOn: []string{"map_to_fhir_resources"},
				Config:    json.RawMessage(`{"query":"INSERT INTO fhir_observations SELECT * FROM json_populate_recordset(null::fhir_observation, :data)"}`),
			},
			{
				ID:        "trigger_critical_lab_alert",
				TaskType:  "http",
				DependsOn: []string{"persist_clinical_lakehouse"},
				Condition: "payload.critical_value == true",
				Config:    json.RawMessage(`{"url":"https://pagerduty.hospital.internal/alerts","method":"POST"}`),
			},
		},
	}
}
