package postgres

import (
	"encoding/binary"
	"testing"
	"time"
)

func TestLogicalReplicationDecoder_BeginAndInsert(t *testing.T) {
	decoder := NewLogicalReplicationDecoder()

	// Register relation 101: public.orders (id, total)
	rel := &CDCRelation{
		RelationID: 101,
		Namespace:  "public",
		TableName:  "orders",
		Columns: []CDCColumn{
			{Name: "id", TypeOID: 23},
			{Name: "total", TypeOID: 1700},
		},
	}
	decoder.RegisterRelation(rel)

	// Construct Begin message 'B' (21 bytes)
	beginMsg := make([]byte, 21)
	beginMsg[0] = 'B'
	binary.BigEndian.PutUint64(beginMsg[1:9], 1000500) // LSN
	binary.BigEndian.PutUint64(beginMsg[9:17], uint64(time.Now().UnixMicro()))
	binary.BigEndian.PutUint32(beginMsg[17:21], 42) // XID

	evt, err := decoder.DecodeMessage(beginMsg)
	if err != nil || evt != nil {
		t.Fatalf("expected nil event for Begin, got: %v, err=%v", evt, err)
	}

	// Construct Insert message 'I':
	// 'I' (1) + relID 101 (4) + 'N' (1) + numCols 2 (2) + 't' (1) + len 2 (4) + "42" + 't' (1) + len 5 (4) + "99.50"
	insertMsg := []byte{'I'}
	relBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(relBytes, 101)
	insertMsg = append(insertMsg, relBytes...)
	insertMsg = append(insertMsg, 'N') // New tuple marker

	numCols := make([]byte, 2)
	binary.BigEndian.PutUint16(numCols, 2)
	insertMsg = append(insertMsg, numCols...)

	// Col 1: "42"
	insertMsg = append(insertMsg, 't')
	len1 := make([]byte, 4)
	binary.BigEndian.PutUint32(len1, 2)
	insertMsg = append(insertMsg, len1...)
	insertMsg = append(insertMsg, []byte("42")...)

	// Col 2: "99.50"
	insertMsg = append(insertMsg, 't')
	len2 := make([]byte, 4)
	binary.BigEndian.PutUint32(len2, 5)
	insertMsg = append(insertMsg, len2...)
	insertMsg = append(insertMsg, []byte("99.50")...)

	evt, err = decoder.DecodeMessage(insertMsg)
	if err != nil {
		t.Fatalf("decode insert failed: %v", err)
	}

	if evt.Action != ActionInsert {
		t.Errorf("expected ActionInsert, got %s", evt.Action)
	}
	if evt.Table != "public.orders" {
		t.Errorf("expected table public.orders, got %s", evt.Table)
	}
	if evt.Transaction != 42 {
		t.Errorf("expected tx 42, got %d", evt.Transaction)
	}
	if evt.NewValues["id"] != "42" || evt.NewValues["total"] != "99.50" {
		t.Errorf("unexpected values: %+v", evt.NewValues)
	}
}
