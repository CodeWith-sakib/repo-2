package postgres

import (
	"testing"
	"time"
)

func TestStatementTimeoutAdapter(t *testing.T) {
	adapter := NewStatementTimeoutAdapter(AdaptiveTimeoutConfig{
		InteractiveTimeout: 500 * time.Millisecond,
		BatchTimeout:       10 * time.Second,
		AnalyticalTimeout:  60 * time.Second,
	})

	sql1 := adapter.BuildSetStatementSQL("INTERACTIVE")
	if sql1 != "SET statement_timeout = 500;" {
		t.Errorf("expected 500ms timeout SQL, got %s", sql1)
	}

	sql2 := adapter.BuildSetStatementSQL("BATCH")
	if sql2 != "SET statement_timeout = 10000;" {
		t.Errorf("expected 10000ms timeout SQL, got %s", sql2)
	}

	sql3 := adapter.BuildSetStatementSQL("UNKNOWN")
	if sql3 != "SET statement_timeout = 10000;" {
		t.Errorf("expected default batch timeout SQL, got %s", sql3)
	}
}
