package core

import (
	"reflect"
	"testing"
)

func TestFunctionLibrary(t *testing.T) {
	fl := Fns

	// String fns
	if fl.Upper("kestrel") != "KESTREL" {
		t.Errorf("Upper mismatch")
	}
	if fl.Substring("kestrelflow", 0, 7) != "kestrel" {
		t.Errorf("Substring mismatch: %s", fl.Substring("kestrelflow", 0, 7))
	}
	if !fl.Contains("pipeline-status", "status") {
		t.Errorf("Contains mismatch")
	}

	// Math fns
	if fl.Min(5, 2, 8, 1, 9) != 1 {
		t.Errorf("Min mismatch")
	}
	if fl.Max(5, 2, 8, 1, 9) != 9 {
		t.Errorf("Max mismatch")
	}
	if fl.Clamp(15, 0, 10) != 10 {
		t.Errorf("Clamp mismatch")
	}

	// Collection fns
	list := []string{"a", "b", "a", "c", "b"}
	uniq := fl.Unique(list)
	if !reflect.DeepEqual(uniq, []string{"a", "b", "c"}) {
		t.Errorf("Unique mismatch: %v", uniq)
	}

	// Type conversion
	valInt, err := fl.ToInt("42")
	if err != nil || valInt != 42 {
		t.Errorf("ToInt failed: %v", err)
	}
}
