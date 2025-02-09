package core

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type ValueKind int

const (
	KindUnknown ValueKind = iota
	KindNil
	KindBool
	KindInt
	KindFloat
	KindString
	KindList
	KindMap
	KindTime
)

type TypedValue struct {
	Kind  ValueKind
	Value interface{}
}

func InferKind(v interface{}) ValueKind {
	if v == nil {
		return KindNil
	}
	switch v.(type) {
	case bool:
		return KindBool
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return KindInt
	case float32, float64:
		return KindFloat
	case string:
		return KindString
	case []interface{}, []string, []int:
		return KindList
	case map[string]interface{}:
		return KindMap
	case time.Time:
		return KindTime
	default:
		return KindUnknown
	}
}

func CoerceToFloat(v interface{}) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	case bool:
		if val {
			return 1.0, true
		}
		return 0.0, true
	default:
		return 0, false
	}
}

func CoerceToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int, int64, int32:
		return fmt.Sprintf("%d", val)
	case float64:
		if math.IsNaN(val) {
			return "NaN"
		}
		if math.IsInf(val, 0) {
			return "Inf"
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func CoerceToBool(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		s := strings.ToLower(strings.TrimSpace(val))
		return s == "true" || s == "1" || s == "yes" || s == "on"
	case int, int64:
		return val != 0
	case float64:
		return val != 0.0 && !math.IsNaN(val)
	default:
		return false
	}
}
