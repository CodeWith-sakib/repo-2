package core

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type FunctionLibrary struct{}

var Fns = &FunctionLibrary{}

// String operations
func (fl *FunctionLibrary) Upper(s string) string { return strings.ToUpper(s) }
func (fl *FunctionLibrary) Lower(s string) string { return strings.ToLower(s) }
func (fl *FunctionLibrary) Trim(s string) string  { return strings.TrimSpace(s) }

func (fl *FunctionLibrary) Split(s, sep string) []string {
	return strings.Split(s, sep)
}

func (fl *FunctionLibrary) Join(elems []string, sep string) string {
	return strings.Join(elems, sep)
}

func (fl *FunctionLibrary) Substring(s string, start, length int) string {
	runes := []rune(s)
	if start < 0 {
		start = 0
	}
	if start >= len(runes) {
		return ""
	}
	end := start + length
	if end > len(runes) || length < 0 {
		end = len(runes)
	}
	return string(runes[start:end])
}

func (fl *FunctionLibrary) Replace(s, oldStr, newStr string) string {
	return strings.ReplaceAll(s, oldStr, newStr)
}

func (fl *FunctionLibrary) Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func (fl *FunctionLibrary) HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func (fl *FunctionLibrary) HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// Math operations
func (fl *FunctionLibrary) Abs(n float64) float64   { return math.Abs(n) }
func (fl *FunctionLibrary) Ceil(n float64) float64  { return math.Ceil(n) }
func (fl *FunctionLibrary) Floor(n float64) float64 { return math.Floor(n) }
func (fl *FunctionLibrary) Round(n float64) float64 { return math.Round(n) }

func (fl *FunctionLibrary) Min(nums ...float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	m := nums[0]
	for _, n := range nums[1:] {
		if n < m {
			m = n
		}
	}
	return m
}

func (fl *FunctionLibrary) Max(nums ...float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	m := nums[0]
	for _, n := range nums[1:] {
		if n > m {
			m = n
		}
	}
	return m
}

func (fl *FunctionLibrary) Clamp(val, minVal, maxVal float64) float64 {
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

// Array collection operations
func (fl *FunctionLibrary) Filter(items []interface{}, predicate func(interface{}) bool) []interface{} {
	out := make([]interface{}, 0)
	for _, it := range items {
		if predicate(it) {
			out = append(out, it)
		}
	}
	return out
}

func (fl *FunctionLibrary) Map(items []interface{}, mapper func(interface{}) interface{}) []interface{} {
	out := make([]interface{}, len(items))
	for i, it := range items {
		out[i] = mapper(it)
	}
	return out
}

func (fl *FunctionLibrary) Reduce(items []interface{}, initial interface{}, reducer func(acc, it interface{}) interface{}) interface{} {
	acc := initial
	for _, it := range items {
		acc = reducer(acc, it)
	}
	return acc
}

func (fl *FunctionLibrary) Reverse(items []interface{}) []interface{} {
	out := make([]interface{}, len(items))
	for i, it := range items {
		out[len(items)-1-i] = it
	}
	return out
}

func (fl *FunctionLibrary) Unique(items []string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(items))
	for _, it := range items {
		if !seen[it] {
			seen[it] = true
			out = append(out, it)
		}
	}
	return out
}

func (fl *FunctionLibrary) Chunk(items []interface{}, size int) [][]interface{} {
	if size <= 0 {
		return [][]interface{}{items}
	}
	var chunks [][]interface{}
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}

// Type coercion
func (fl *FunctionLibrary) ToInt(val interface{}) (int64, error) {
	switch v := val.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", val)
	}
}

func (fl *FunctionLibrary) ToFloat(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float", val)
	}
}

func (fl *FunctionLibrary) ToString(val interface{}) string {
	return fmt.Sprintf("%v", val)
}

func (fl *FunctionLibrary) ParseTimestamp(val, layout string) (time.Time, error) {
	if layout == "" {
		layout = time.RFC3339
	}
	return time.Parse(layout, val)
}
