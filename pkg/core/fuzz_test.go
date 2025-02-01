package core

import (
	"encoding/json"
	"testing"
)

func FuzzParseCron(f *testing.F) {
	seeds := []string{
		"* * * * *",
		"0 0 * * *",
		"*/5 */2 1,15 * 1-5",
		"0 12 * * MON-FRI",
		"invalid cron",
		"60 24 32 13 8",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, expr string) {
		// Target must not panic on any malformed input
		_, _ = ParseCron(expr)
	})
}

func FuzzExtractJSONPath(f *testing.F) {
	seeds := []string{
		"$",
		"$.user.name",
		"items[0].id",
		"data.results[2].score",
		"a[b[c]]",
		"....",
		"[9999999999]",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	sampleData := map[string]interface{}{
		"user": map[string]interface{}{"name": "Alice"},
		"items": []interface{}{
			map[string]interface{}{"id": 1},
		},
	}

	f.Fuzz(func(t *testing.T, path string) {
		_, _ = ExtractJSONPath(sampleData, path)
	})
}

func FuzzExpressionLexer(f *testing.F) {
	seeds := []string{
		"status == 'success'",
		"code >= 200 && code < 300",
		"!(active == false)",
		"'unterminated string",
		"123.456.789",
		"&&&& |||| ==== !!!!",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		lexer := NewLexer(input)
		for {
			tok, err := lexer.NextToken()
			if err != nil || tok.Type == TokenEOF {
				break
			}
		}
	})
}

func FuzzJSONSchemaValidation(f *testing.F) {
	schema := &JSONSchema{
		Type: SchemaTypeObject,
		Properties: map[string]*JSONSchema{
			"id":   {Type: SchemaTypeInteger},
			"name": {Type: SchemaTypeString},
		},
		Required: []string{"id"},
	}

	seeds := [][]byte{
		[]byte(`{"id": 1, "name": "foo"}`),
		[]byte(`{"id": "not an int"}`),
		[]byte(`{}`),
		[]byte(`[1, 2, 3]`),
		[]byte(`not json`),
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, payload []byte) {
		var val interface{}
		if err := json.Unmarshal(payload, &val); err == nil {
			_ = schema.Validate(val)
		}
	})
}
