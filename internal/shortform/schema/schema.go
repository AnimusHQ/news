// Package schema is a small, dependency-free JSON Schema (Draft 2020-12 subset)
// validator. It exists so short-form artifact schemas can be shipped as real,
// committed JSON Schema documents and validated against offline with no third
// party dependency (see docs/adr/0003).
//
// Supported keywords: type, required, properties, additionalProperties (bool),
// enum, const, items (single subschema), minItems, maxItems, minimum, maximum,
// minLength, pattern. Annotations $id, $schema, title, description, examples are
// ignored. Any other keyword is rejected at Compile time (fail-closed) so a
// schema can never silently accept instances it does not actually constrain.
package schema

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

// Schema is a compiled schema node.
type Schema struct {
	types     []string
	required  []string
	props     map[string]*Schema
	addlProps *bool
	enum      []any
	hasConst  bool
	constVal  any
	items     *Schema
	minItems  *int
	maxItems  *int
	minimum   *float64
	maximum   *float64
	minLength *int
	pattern   *regexp.Regexp
}

var supportedKeywords = map[string]bool{
	"type": true, "required": true, "properties": true, "additionalProperties": true,
	"enum": true, "const": true, "items": true, "minItems": true, "maxItems": true,
	"minimum": true, "maximum": true, "minLength": true, "pattern": true,
	// ignored annotations
	"$id": true, "$schema": true, "title": true, "description": true, "examples": true,
}

// Compile parses and compiles a JSON Schema document, failing closed on any
// keyword outside the supported subset.
func Compile(raw []byte) (*Schema, error) {
	var node any
	if err := json.Unmarshal(raw, &node); err != nil {
		return nil, fmt.Errorf("schema is not valid JSON: %w", err)
	}
	return compileNode(node, "$")
}

func compileNode(node any, path string) (*Schema, error) {
	obj, ok := node.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: schema node must be an object", path)
	}
	s := &Schema{}
	for key, val := range obj {
		if !supportedKeywords[key] {
			return nil, fmt.Errorf("%s: unsupported schema keyword %q", path, key)
		}
		var err error
		switch key {
		case "type":
			s.types, err = compileStringList(val, path+".type")
		case "required":
			s.required, err = compileStringList(val, path+".required")
		case "properties":
			s.props, err = compileProps(val, path+".properties")
		case "additionalProperties":
			b, ok := val.(bool)
			if !ok {
				err = fmt.Errorf("%s.additionalProperties: only boolean is supported", path)
			} else {
				s.addlProps = &b
			}
		case "enum":
			arr, ok := val.([]any)
			if !ok || len(arr) == 0 {
				err = fmt.Errorf("%s.enum: must be a non-empty array", path)
			} else {
				s.enum = arr
			}
		case "const":
			s.hasConst = true
			s.constVal = val
		case "items":
			s.items, err = compileNode(val, path+".items")
		case "minItems":
			s.minItems, err = compileInt(val, path+".minItems")
		case "maxItems":
			s.maxItems, err = compileInt(val, path+".maxItems")
		case "minLength":
			s.minLength, err = compileInt(val, path+".minLength")
		case "minimum":
			s.minimum, err = compileFloat(val, path+".minimum")
		case "maximum":
			s.maximum, err = compileFloat(val, path+".maximum")
		case "pattern":
			str, ok := val.(string)
			if !ok {
				err = fmt.Errorf("%s.pattern: must be a string", path)
			} else {
				s.pattern, err = regexp.Compile(str)
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

func compileStringList(val any, path string) ([]string, error) {
	switch t := val.(type) {
	case string:
		return []string{t}, nil
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			str, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("%s: entries must be strings", path)
			}
			out = append(out, str)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("%s: must be a string or array of strings", path)
	}
}

func compileProps(val any, path string) (map[string]*Schema, error) {
	obj, ok := val.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s: must be an object", path)
	}
	out := make(map[string]*Schema, len(obj))
	for name, sub := range obj {
		compiled, err := compileNode(sub, path+"."+name)
		if err != nil {
			return nil, err
		}
		out[name] = compiled
	}
	return out, nil
}

func compileInt(val any, path string) (*int, error) {
	f, ok := val.(float64)
	if !ok || f != math.Trunc(f) || f < 0 {
		return nil, fmt.Errorf("%s: must be a non-negative integer", path)
	}
	i := int(f)
	return &i, nil
}

func compileFloat(val any, path string) (*float64, error) {
	f, ok := val.(float64)
	if !ok {
		return nil, fmt.Errorf("%s: must be a number", path)
	}
	return &f, nil
}

// ValidateValue marshals v to JSON, decodes it into a generic instance, and
// validates it. It returns a sorted list of human-readable error strings; an
// empty slice means the instance is valid.
func ValidateValue(s *Schema, v any) []string {
	data, err := json.Marshal(v)
	if err != nil {
		return []string{fmt.Sprintf("$: cannot marshal instance: %v", err)}
	}
	return ValidateBytes(s, data)
}

// ValidateBytes validates raw JSON bytes against the schema.
func ValidateBytes(s *Schema, data []byte) []string {
	var instance any
	if err := json.Unmarshal(data, &instance); err != nil {
		return []string{fmt.Sprintf("$: instance is not valid JSON: %v", err)}
	}
	var errs []string
	s.validate(instance, "$", &errs)
	sort.Strings(errs)
	return errs
}

func (s *Schema) validate(instance any, path string, errs *[]string) {
	if len(s.types) > 0 && !matchesType(s.types, instance) {
		*errs = append(*errs, fmt.Sprintf("%s: expected type %s, got %s", path, strings.Join(s.types, "|"), jsonType(instance)))
		return
	}
	if s.hasConst && !jsonEqual(instance, s.constVal) {
		*errs = append(*errs, fmt.Sprintf("%s: must equal const %v", path, s.constVal))
	}
	if len(s.enum) > 0 && !containsValue(s.enum, instance) {
		*errs = append(*errs, fmt.Sprintf("%s: must be one of %v", path, s.enum))
	}

	switch v := instance.(type) {
	case map[string]any:
		s.validateObject(v, path, errs)
	case []any:
		s.validateArray(v, path, errs)
	case string:
		s.validateString(v, path, errs)
	case float64:
		s.validateNumber(v, path, errs)
	}
}

func (s *Schema) validateObject(obj map[string]any, path string, errs *[]string) {
	for _, req := range s.required {
		if _, ok := obj[req]; !ok {
			*errs = append(*errs, fmt.Sprintf("%s.%s: required property is missing", path, req))
		}
	}
	if s.addlProps != nil && !*s.addlProps {
		for key := range obj {
			if _, ok := s.props[key]; !ok {
				*errs = append(*errs, fmt.Sprintf("%s.%s: additional property is not allowed", path, key))
			}
		}
	}
	for name, sub := range s.props {
		if val, ok := obj[name]; ok {
			sub.validate(val, path+"."+name, errs)
		}
	}
}

func (s *Schema) validateArray(arr []any, path string, errs *[]string) {
	if s.minItems != nil && len(arr) < *s.minItems {
		*errs = append(*errs, fmt.Sprintf("%s: must have at least %d items", path, *s.minItems))
	}
	if s.maxItems != nil && len(arr) > *s.maxItems {
		*errs = append(*errs, fmt.Sprintf("%s: must have at most %d items", path, *s.maxItems))
	}
	if s.items != nil {
		for i, item := range arr {
			s.items.validate(item, fmt.Sprintf("%s[%d]", path, i), errs)
		}
	}
}

func (s *Schema) validateString(str, path string, errs *[]string) {
	if s.minLength != nil && len([]rune(str)) < *s.minLength {
		*errs = append(*errs, fmt.Sprintf("%s: must be at least %d characters", path, *s.minLength))
	}
	if s.pattern != nil && !s.pattern.MatchString(str) {
		*errs = append(*errs, fmt.Sprintf("%s: must match pattern %s", path, s.pattern.String()))
	}
}

func (s *Schema) validateNumber(f float64, path string, errs *[]string) {
	if s.minimum != nil && f < *s.minimum {
		*errs = append(*errs, fmt.Sprintf("%s: must be >= %v", path, *s.minimum))
	}
	if s.maximum != nil && f > *s.maximum {
		*errs = append(*errs, fmt.Sprintf("%s: must be <= %v", path, *s.maximum))
	}
}

func matchesType(allowed []string, v any) bool {
	actual := jsonType(v)
	for _, t := range allowed {
		if t == actual {
			return true
		}
		switch t {
		case "integer":
			if f, ok := v.(float64); ok && f == math.Trunc(f) {
				return true
			}
		case "number":
			if _, ok := v.(float64); ok {
				return true
			}
		}
	}
	return false
}

func jsonType(v any) string {
	switch v.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case float64:
		return "number"
	case string:
		return "string"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return fmt.Sprintf("%T", v)
	}
}

func containsValue(set []any, v any) bool {
	for _, item := range set {
		if jsonEqual(item, v) {
			return true
		}
	}
	return false
}

func jsonEqual(a, b any) bool {
	return reflect.DeepEqual(a, b)
}
