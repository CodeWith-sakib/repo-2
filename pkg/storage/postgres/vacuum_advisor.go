package postgres

import (
	"fmt"
	"strings"
	"time"
)

// VacuumAction describes the severity and type of recommended vacuuming.
type VacuumAction string

const (
	ActionNone          VacuumAction = "NONE"
	ActionVacuum        VacuumAction = "VACUUM"
	ActionVacuumAnalyze VacuumAction = "VACUUM ANALYZE"
	ActionVacuumFreeze  VacuumAction = "VACUUM FREEZE"
	ActionVacuumFull    VacuumAction = "VACUUM FULL"
)

// TableBloatStats contains statistics gathered from pg_stat_user_tables and pg_class.
type TableBloatStats struct {
	SchemaName         string
	TableName          string
	LiveTuples         int64
	DeadTuples         int64
	TableSizeBytes     int64
	LastVacuum         time.Time
	LastAutovacuum     time.Time
	XIDAge             int64
	AutovacuumXIDLimit int64
}

// DeadTupleRatio returns the proportion of dead tuples to total tuples.
func (s *TableBloatStats) DeadTupleRatio() float64 {
	total := s.LiveTuples + s.DeadTuples
	if total == 0 {
		return 0
	}
	return float64(s.DeadTuples) / float64(total)
}

// VacuumRecommendation is an advice recommendation for a table.
type VacuumRecommendation struct {
	Table        string
	Action       VacuumAction
	Urgency      string // "LOW", "MEDIUM", "HIGH", "CRITICAL"
	Reason       string
	SuggestedSQL string
}

// VacuumAdvisorConfig sets thresholds for issuing maintenance recommendations.
type VacuumAdvisorConfig struct {
	DeadRatioThreshold float64 // default 0.20 (20% dead tuples)
	HighBloatThreshold float64 // default 0.50 (50% dead tuples)
	MinDeadTuples      int64   // default 10,000 dead tuples to avoid churning tiny tables
	XIDFreezeRatio     float64 // default 0.80 (80% towards wraparound limit)
}

// DefaultVacuumAdvisorConfig returns production default thresholds.
func DefaultVacuumAdvisorConfig() VacuumAdvisorConfig {
	return VacuumAdvisorConfig{
		DeadRatioThreshold: 0.20,
		HighBloatThreshold: 0.50,
		MinDeadTuples:      10000,
		XIDFreezeRatio:     0.80,
	}
}

// VacuumAdvisor analyzes table statistics and produces automated maintenance recommendations.
type VacuumAdvisor struct {
	cfg VacuumAdvisorConfig
}

// NewVacuumAdvisor creates an advisor with the given thresholds.
func NewVacuumAdvisor(cfg VacuumAdvisorConfig) *VacuumAdvisor {
	return &VacuumAdvisor{cfg: cfg}
}

// EvaluateTable inspects a single table's statistics and recommends action if required.
func (a *VacuumAdvisor) EvaluateTable(stats TableBloatStats) VacuumRecommendation {
	fullName := fmt.Sprintf("%s.%s", stats.SchemaName, stats.TableName)
	if stats.SchemaName == "" {
		fullName = stats.TableName
	}

	// 1. Check for transaction wraparound risk (highest urgency)
	if stats.AutovacuumXIDLimit > 0 && stats.XIDAge > 0 {
		xidRatio := float64(stats.XIDAge) / float64(stats.AutovacuumXIDLimit)
		if xidRatio >= a.cfg.XIDFreezeRatio {
			return VacuumRecommendation{
				Table:        fullName,
				Action:       ActionVacuumFreeze,
				Urgency:      "CRITICAL",
				Reason:       fmt.Sprintf("Transaction ID age (%d) has reached %.1f%% of wraparound limit", stats.XIDAge, xidRatio*100),
				SuggestedSQL: fmt.Sprintf("VACUUM (FREEZE, VERBOSE) %s;", fullName),
			}
		}
	}

	ratio := stats.DeadTupleRatio()

	// Only alert if above minimum dead tuple count to avoid noisy small tables
	if stats.DeadTuples < a.cfg.MinDeadTuples {
		return VacuumRecommendation{
			Table:   fullName,
			Action:  ActionNone,
			Urgency: "LOW",
			Reason:  "Dead tuple count below analysis threshold",
		}
	}

	// Severe table bloat -> recommend VACUUM FULL
	if ratio >= a.cfg.HighBloatThreshold {
		return VacuumRecommendation{
			Table:        fullName,
			Action:       ActionVacuumFull,
			Urgency:      "HIGH",
			Reason:       fmt.Sprintf("Dead tuple ratio %.1f%% exceeds critical bloat threshold %.1f%%", ratio*100, a.cfg.HighBloatThreshold*100),
			SuggestedSQL: fmt.Sprintf("VACUUM FULL %s; -- Or run pg_repack to avoid exclusive table locks", fullName),
		}
	}

	// Moderate bloat -> recommend VACUUM ANALYZE
	if ratio >= a.cfg.DeadRatioThreshold {
		return VacuumRecommendation{
			Table:        fullName,
			Action:       ActionVacuumAnalyze,
			Urgency:      "MEDIUM",
			Reason:       fmt.Sprintf("Dead tuple ratio %.1f%% exceeds vacuum threshold %.1f%%", ratio*100, a.cfg.DeadRatioThreshold*100),
			SuggestedSQL: fmt.Sprintf("VACUUM (ANALYZE, VERBOSE) %s;", fullName),
		}
	}

	return VacuumRecommendation{
		Table:   fullName,
		Action:  ActionNone,
		Urgency: "LOW",
		Reason:  "Table bloat within normal bounds",
	}
}

// GenerateMaintenanceScript compiles a list of SQL statements for all actionable recommendations.
func (a *VacuumAdvisor) GenerateMaintenanceScript(tables []TableBloatStats) string {
	var sb strings.Builder
	sb.WriteString("-- PostgreSQL Automated Vacuum Maintenance Script\n")
	sb.WriteString(fmt.Sprintf("-- Generated: %s\n\n", time.Now().UTC().Format(time.RFC3339)))

	count := 0
	for _, t := range tables {
		rec := a.EvaluateTable(t)
		if rec.Action != ActionNone {
			sb.WriteString(fmt.Sprintf("-- [%s] Table: %s (Reason: %s)\n", rec.Urgency, rec.Table, rec.Reason))
			sb.WriteString(rec.SuggestedSQL)
			sb.WriteString("\n\n")
			count++
		}
	}

	if count == 0 {
		sb.WriteString("-- All tables are healthy. No maintenance required.\n")
	}

	return sb.String()
}
