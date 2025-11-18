package postgres

import (
	"encoding/binary"
	"fmt"
	"time"
)

// CDCAction classifies CDC event operations.
type CDCAction string

const (
	ActionInsert CDCAction = "INSERT"
	ActionUpdate CDCAction = "UPDATE"
	ActionDelete CDCAction = "DELETE"
)

// CDCColumn represents a column attribute in relation metadata.
type CDCColumn struct {
	Name    string
	TypeOID uint32
	KeyFlag bool
}

// CDCRelation holds table schema metadata for logical decoding.
type CDCRelation struct {
	RelationID uint32
	Namespace  string
	TableName  string
	Columns    []CDCColumn
}

// CDCChangeEvent represents a decoded database mutation event.
type CDCChangeEvent struct {
	Action      CDCAction
	Table       string
	LSN         uint64
	Timestamp   time.Time
	OldValues   map[string]interface{}
	NewValues   map[string]interface{}
	Transaction uint32
}

// LogicalReplicationDecoder parses PostgreSQL pgoutput logical replication message streams.
type LogicalReplicationDecoder struct {
	relations     map[uint32]*CDCRelation
	currentXID    uint32
	currentLSN    uint64
	currentTxTime time.Time
}

// NewLogicalReplicationDecoder creates a logical decoder.
func NewLogicalReplicationDecoder() *LogicalReplicationDecoder {
	return &LogicalReplicationDecoder{
		relations: make(map[uint32]*CDCRelation),
	}
}

// RegisterRelation stores relation metadata received in 'R' messages.
func (d *LogicalReplicationDecoder) RegisterRelation(rel *CDCRelation) {
	d.relations[rel.RelationID] = rel
}

// DecodeMessage parses a raw replication message payload. Returns a CDCChangeEvent if the message represents a data mutation.
func (d *LogicalReplicationDecoder) DecodeMessage(data []byte) (*CDCChangeEvent, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty replication message")
	}

	msgType := data[0]

	switch msgType {
	case 'B': // Begin
		if len(data) < 21 {
			return nil, fmt.Errorf("malformed Begin message: length %d", len(data))
		}
		d.currentLSN = binary.BigEndian.Uint64(data[1:9])
		// Timestamp is microsec since 2000-01-01
		tsMicro := binary.BigEndian.Uint64(data[9:17])
		epoch2000 := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		d.currentTxTime = epoch2000.Add(time.Duration(tsMicro) * time.Microsecond)
		d.currentXID = binary.BigEndian.Uint32(data[17:21])
		return nil, nil

	case 'C': // Commit
		return nil, nil

	case 'I': // Insert: 'I' (1 byte) + relID (4 bytes) + 'N' (1 byte) + tuple data
		if len(data) < 6 {
			return nil, fmt.Errorf("malformed Insert message: length %d", len(data))
		}
		relID := binary.BigEndian.Uint32(data[1:5])
		rel, ok := d.relations[relID]
		if !ok {
			return nil, fmt.Errorf("unknown relation ID %d", relID)
		}

		newVals := d.parseTupleValues(data[5:], rel)

		return &CDCChangeEvent{
			Action:      ActionInsert,
			Table:       fmt.Sprintf("%s.%s", rel.Namespace, rel.TableName),
			LSN:         d.currentLSN,
			Timestamp:   d.currentTxTime,
			NewValues:   newVals,
			Transaction: d.currentXID,
		}, nil

	case 'D': // Delete: 'D' (1 byte) + relID (4 bytes) + 'K'/'O' (1 byte) + tuple data
		if len(data) < 6 {
			return nil, fmt.Errorf("malformed Delete message: length %d", len(data))
		}
		relID := binary.BigEndian.Uint32(data[1:5])
		rel, ok := d.relations[relID]
		if !ok {
			return nil, fmt.Errorf("unknown relation ID %d", relID)
		}

		oldVals := d.parseTupleValues(data[5:], rel)

		return &CDCChangeEvent{
			Action:      ActionDelete,
			Table:       fmt.Sprintf("%s.%s", rel.Namespace, rel.TableName),
			LSN:         d.currentLSN,
			Timestamp:   d.currentTxTime,
			OldValues:   oldVals,
			Transaction: d.currentXID,
		}, nil

	default:
		return nil, nil
	}
}

func (d *LogicalReplicationDecoder) parseTupleValues(data []byte, rel *CDCRelation) map[string]interface{} {
	values := make(map[string]interface{})
	if len(data) < 3 {
		return values
	}

	// tuple type is data[0] ('N', 'K', 'O')
	numCols := int(binary.BigEndian.Uint16(data[1:3]))
	offset := 3

	for i := 0; i < numCols && i < len(rel.Columns); i++ {
		if offset >= len(data) {
			break
		}
		colType := data[offset]
		offset++

		colName := rel.Columns[i].Name

		switch colType {
		case 'n': // null
			values[colName] = nil
		case 't': // text formatted data: 4 bytes length + bytes
			if offset+4 <= len(data) {
				length := int(binary.BigEndian.Uint32(data[offset : offset+4]))
				offset += 4
				if offset+length <= len(data) {
					values[colName] = string(data[offset : offset+length])
					offset += length
				}
			}
		}
	}

	return values
}
