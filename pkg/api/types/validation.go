package types

import (
	"fmt"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	identRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]{1,128}$`)
)

type FieldValidationError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

type StructValidator struct {
	mu         sync.RWMutex
	customTags map[string]func(val reflect.Value, param string) bool
}

func NewStructValidator() *StructValidator {
	v := &StructValidator{
		customTags: make(map[string]func(val reflect.Value, param string) bool),
	}
	v.registerDefaults()
	return v
}

func (v *StructValidator) registerDefaults() {
	v.customTags["required"] = func(val reflect.Value, _ string) bool {
		if !val.IsValid() {
			return false
		}
		switch val.Kind() {
		case reflect.String:
			return strings.TrimSpace(val.String()) != ""
		case reflect.Slice, reflect.Map, reflect.Array:
			return val.Len() > 0
		case reflect.Ptr, reflect.Interface:
			return !val.IsNil()
		default:
			return true
		}
	}

	v.customTags["min"] = func(val reflect.Value, param string) bool {
		limit, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return false
		}
		switch val.Kind() {
		case reflect.String:
			return int64(len(val.String())) >= limit
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return val.Int() >= limit
		case reflect.Slice, reflect.Map:
			return int64(val.Len()) >= limit
		default:
			return true
		}
	}

	v.customTags["max"] = func(val reflect.Value, param string) bool {
		limit, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return false
		}
		switch val.Kind() {
		case reflect.String:
			return int64(len(val.String())) <= limit
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return val.Int() <= limit
		case reflect.Slice, reflect.Map:
			return int64(val.Len()) <= limit
		default:
			return true
		}
	}

	v.customTags["identifier"] = func(val reflect.Value, _ string) bool {
		if val.Kind() != reflect.String {
			return false
		}
		return identRegex.MatchString(val.String())
	}

	v.customTags["email"] = func(val reflect.Value, _ string) bool {
		if val.Kind() != reflect.String {
			return false
		}
		_, err := mail.ParseAddress(val.String())
		return err == nil
	}

	v.customTags["url"] = func(val reflect.Value, _ string) bool {
		if val.Kind() != reflect.String {
			return false
		}
		u, err := url.ParseRequestURI(val.String())
		return err == nil && u.Scheme != "" && u.Host != ""
	}
}

func (v *StructValidator) Validate(s interface{}) []FieldValidationError {
	var errs []FieldValidationError
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return errs
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)
		fieldType := typ.Field(i)
		tag := fieldType.Tag.Get("validate")
		if tag == "" || tag == "-" {
			continue
		}

		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			var ruleName, ruleParam string
			if idx := strings.Index(rule, "="); idx != -1 {
				ruleName = rule[:idx]
				ruleParam = rule[idx+1:]
			} else {
				ruleName = rule
			}

			fn, ok := v.customTags[ruleName]
			if ok {
				if !fn(fieldVal, ruleParam) {
					errs = append(errs, FieldValidationError{
						Field:   fieldType.Name,
						Rule:    ruleName,
						Message: fmt.Sprintf("field %s violates rule %s", fieldType.Name, rule),
					})
				}
			}
		}
	}

	return errs
}
