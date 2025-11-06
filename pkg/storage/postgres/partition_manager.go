package postgres

import (
	"fmt"
	"strings"
	"time"
)

// PartitionGranularity defines partition time boundaries.
type PartitionGranularity string

const (
	GranularityDaily   PartitionGranularity = "DAILY"
	GranularityMonthly PartitionGranularity = "MONTHLY"
)

// RangePartitionDefinition describes a single table partition.
type RangePartitionDefinition struct {
	ParentTable   string
	PartitionName string
	FromTimestamp time.Time
	ToTimestamp   time.Time
}

// TablePartitionManager generates declarative partitioning DDL and retention policies.
type TablePartitionManager struct {
	parentTable  string
	partitionCol string
	granularity  PartitionGranularity
}

// NewTablePartitionManager creates a partition manager.
func NewTablePartitionManager(parentTable, partitionCol string, gran PartitionGranularity) (*TablePartitionManager, error) {
	if parentTable == "" || partitionCol == "" {
		return nil, fmt.Errorf("parent table and partition column cannot be empty")
	}
	if gran != GranularityDaily && gran != GranularityMonthly {
		gran = GranularityMonthly
	}
	return &TablePartitionManager{
		parentTable:  parentTable,
		partitionCol: partitionCol,
		granularity:  gran,
	}, nil
}

// ParentTableDDL returns DDL to create the base partitioned table.
func (m *TablePartitionManager) ParentTableDDL(columnsSQL string) string {
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n%s\n) PARTITION BY RANGE (%s);",
		m.parentTable, columnsSQL, m.partitionCol)
}

// PlanPartitionsAhead generates upcoming partition definitions for the next N intervals.
func (m *TablePartitionManager) PlanPartitionsAhead(startFrom time.Time, count int) []RangePartitionDefinition {
	if count <= 0 {
		count = 1
	}

	var partitions []RangePartitionDefinition

	if m.granularity == GranularityMonthly {
		curr := time.Date(startFrom.Year(), startFrom.Month(), 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < count; i++ {
			next := curr.AddDate(0, 1, 0)
			name := fmt.Sprintf("%s_y%04dm%02d", m.parentTable, curr.Year(), curr.Month())
			partitions = append(partitions, RangePartitionDefinition{
				ParentTable:   m.parentTable,
				PartitionName: name,
				FromTimestamp: curr,
				ToTimestamp:   next,
			})
			curr = next
		}
	} else {
		// Daily
		curr := time.Date(startFrom.Year(), startFrom.Month(), startFrom.Day(), 0, 0, 0, 0, time.UTC)
		for i := 0; i < count; i++ {
			next := curr.AddDate(0, 0, 1)
			name := fmt.Sprintf("%s_y%04dm%02dd%02d", m.parentTable, curr.Year(), curr.Month(), curr.Day())
			partitions = append(partitions, RangePartitionDefinition{
				ParentTable:   m.parentTable,
				PartitionName: name,
				FromTimestamp: curr,
				ToTimestamp:   next,
			})
			curr = next
		}
	}

	return partitions
}

// CreatePartitionDDL generates the SQL statement to create a partition table.
func (m *TablePartitionManager) CreatePartitionDDL(p RangePartitionDefinition) string {
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s');",
		p.PartitionName, p.ParentTable,
		p.FromTimestamp.Format("2006-01-02 15:04:05-07"),
		p.ToTimestamp.Format("2006-01-02 15:04:05-07"))
}

// DetachPartitionDDL generates SQL to concurrently detach a partition before archiving or dropping.
func (m *TablePartitionManager) DetachPartitionDDL(partitionName string, concurrently bool) string {
	concurrentClause := ""
	if concurrently {
		concurrentClause = " CONCURRENTLY"
	}
	return fmt.Sprintf("ALTER TABLE %s DETACH PARTITION %s%s;",
		m.parentTable, partitionName, concurrentClause)
}

// DropPartitionDDL generates SQL to drop an old detached partition.
func (m *TablePartitionManager) DropPartitionDDL(partitionName string) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s;", partitionName)
}

// GenerateMaintenanceScript creates a combined migration script ensuring partitions exist for the next N intervals.
func (m *TablePartitionManager) GenerateMaintenanceScript(startFrom time.Time, count int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("-- Partition creation script for %s\n", m.parentTable))

	plans := m.PlanPartitionsAhead(startFrom, count)
	for _, p := range plans {
		sb.WriteString(m.CreatePartitionDDL(p))
		sb.WriteString("\n")
	}

	return sb.String()
}
