package schema

import (
	"strings"
	"testing"
)

func mustCompile(t *testing.T, raw string) *Schema {
	t.Helper()
	s, err := Compile([]byte(raw))
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	return s
}

func TestCompileRejectsUnknownKeyword(t *testing.T) {
	_, err := Compile([]byte(`{"type":"object","uniqueItems":true}`))
	if err == nil || !strings.Contains(err.Error(), "unsupported schema keyword") {
		t.Fatalf("expected unsupported keyword error, got %v", err)
	}
}

func TestTypeChecking(t *testing.T) {
	s := mustCompile(t, `{"type":"object"}`)
	if errs := ValidateBytes(s, []byte(`[]`)); len(errs) == 0 {
		t.Fatal("array should not satisfy object type")
	}
	if errs := ValidateBytes(s, []byte(`{}`)); len(errs) != 0 {
		t.Fatalf("object should satisfy object type, got %v", errs)
	}
}

func TestIntegerType(t *testing.T) {
	s := mustCompile(t, `{"type":"integer"}`)
	if errs := ValidateBytes(s, []byte(`3`)); len(errs) != 0 {
		t.Fatalf("3 is an integer, got %v", errs)
	}
	if errs := ValidateBytes(s, []byte(`3.5`)); len(errs) == 0 {
		t.Fatal("3.5 must not satisfy integer")
	}
}

func TestRequiredAndAdditionalProperties(t *testing.T) {
	s := mustCompile(t, `{
      "type":"object",
      "required":["a"],
      "properties":{"a":{"type":"string"}},
      "additionalProperties":false
    }`)
	if errs := ValidateBytes(s, []byte(`{"a":"x"}`)); len(errs) != 0 {
		t.Fatalf("valid object rejected: %v", errs)
	}
	if errs := ValidateBytes(s, []byte(`{}`)); len(errs) == 0 {
		t.Fatal("missing required property must fail")
	}
	if errs := ValidateBytes(s, []byte(`{"a":"x","b":1}`)); len(errs) == 0 {
		t.Fatal("additional property must fail")
	}
}

func TestEnumAndConst(t *testing.T) {
	s := mustCompile(t, `{"enum":["tiktok","youtube"]}`)
	if errs := ValidateBytes(s, []byte(`"tiktok"`)); len(errs) != 0 {
		t.Fatalf("enum member rejected: %v", errs)
	}
	if errs := ValidateBytes(s, []byte(`"vimeo"`)); len(errs) == 0 {
		t.Fatal("non-member must fail enum")
	}
	c := mustCompile(t, `{"const":"9:16"}`)
	if errs := ValidateBytes(c, []byte(`"9:16"`)); len(errs) != 0 {
		t.Fatalf("const match rejected: %v", errs)
	}
	if errs := ValidateBytes(c, []byte(`"16:9"`)); len(errs) == 0 {
		t.Fatal("const mismatch must fail")
	}
}

func TestItemsAndMinItems(t *testing.T) {
	s := mustCompile(t, `{"type":"array","minItems":1,"items":{"type":"string"}}`)
	if errs := ValidateBytes(s, []byte(`["a","b"]`)); len(errs) != 0 {
		t.Fatalf("valid array rejected: %v", errs)
	}
	if errs := ValidateBytes(s, []byte(`[]`)); len(errs) == 0 {
		t.Fatal("empty array must fail minItems")
	}
	if errs := ValidateBytes(s, []byte(`[1]`)); len(errs) == 0 {
		t.Fatal("wrong item type must fail")
	}
}

func TestNumericBounds(t *testing.T) {
	s := mustCompile(t, `{"type":"number","minimum":1,"maximum":10}`)
	if errs := ValidateBytes(s, []byte(`5`)); len(errs) != 0 {
		t.Fatalf("in-range rejected: %v", errs)
	}
	if errs := ValidateBytes(s, []byte(`0`)); len(errs) == 0 {
		t.Fatal("below minimum must fail")
	}
	if errs := ValidateBytes(s, []byte(`11`)); len(errs) == 0 {
		t.Fatal("above maximum must fail")
	}
}

func TestPatternAndMinLength(t *testing.T) {
	s := mustCompile(t, `{"type":"string","pattern":"^sha256:[a-f0-9]{64}$"}`)
	good := `"sha256:` + strings.Repeat("a", 64) + `"`
	if errs := ValidateBytes(s, []byte(good)); len(errs) != 0 {
		t.Fatalf("valid hash rejected: %v", errs)
	}
	if errs := ValidateBytes(s, []byte(`"sha256:xyz"`)); len(errs) == 0 {
		t.Fatal("bad hash must fail pattern")
	}
	ml := mustCompile(t, `{"type":"string","minLength":2}`)
	if errs := ValidateBytes(ml, []byte(`"a"`)); len(errs) == 0 {
		t.Fatal("short string must fail minLength")
	}
}

func TestNestedObjectValidation(t *testing.T) {
	s := mustCompile(t, `{
      "type":"object",
      "required":["target"],
      "properties":{
        "target":{
          "type":"object",
          "required":["fps"],
          "properties":{"fps":{"const":30}}
        }
      }
    }`)
	if errs := ValidateBytes(s, []byte(`{"target":{"fps":30}}`)); len(errs) != 0 {
		t.Fatalf("valid nested rejected: %v", errs)
	}
	if errs := ValidateBytes(s, []byte(`{"target":{"fps":24}}`)); len(errs) == 0 {
		t.Fatal("wrong nested const must fail")
	}
}
