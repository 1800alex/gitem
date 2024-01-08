package cfgvars

import (
	"os"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	// Test case 1: Empty variables
	variables := []Variable{}
	expected := map[string]interface{}{}
	result, err := Parse(variables, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Unexpected result. Expected: %v, Got: %v", expected, result)
	}

	// Test case 2: Single variable
	variables = []Variable{
		{
			Name:   "name",
			String: stringPtr("John Doe"),
		},
	}
	expected = map[string]interface{}{
		"name": "John Doe",
	}
	result, err = Parse(variables, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Unexpected result. Expected: %v, Got: %v", expected, result)
	}

	// Test case 3: Multiple variables
	variables = []Variable{
		{
			Name:   "name",
			String: stringPtr("John Doe"),
		},
		{
			Name: "age",
			Int:  stringPtr("30"),
		},
		{
			Name: "isMarried",
			Bool: stringPtr("true"),
		},
	}
	expected = map[string]interface{}{
		"name":      "John Doe",
		"age":       30,
		"isMarried": true,
	}
	result, err = Parse(variables, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Unexpected result. Expected: %v, Got: %v", expected, result)
	}

	// Test case 4: Multiple variables with placeholders
	variables = []Variable{
		{
			Name: "boolvar",
			Bool: stringPtr("true"),
		},
		{
			Name: "intvar",
			Int:  stringPtr("42"),
		},
		{
			Name:  "floatvar",
			Float: stringPtr("3.14"),
		},
		{
			Name:   "stringvar",
			String: stringPtr("some text"),
		},
		{
			Name: "home",
			Env:  stringPtr("HOME"),
		},
		{
			Name:   "testdata",
			String: stringPtr("testdata"),
		},
		{
			Name: "filevar",
			File: stringPtr("{{.testdata}}/file.txt"),
		},
		{
			Name: "jsonvar",
			Json: stringPtr(`{"foo": "bar"}`),
		},
		{
			Name: "yamlvar",
			Yaml: stringPtr("foo: bar\nbaz: qux\nval: 42"),
		},
		{
			Name: "yamlvar2",
			Yaml: stringPtr("foo: bar"),
		},
		{
			Name:  "upper",
			Shell: stringPtr("cat {{.testdata}}/file.txt | tr '[:lower:]' '[:upper:]' | tr -d '\n'"),
		},
		{
			Name: "var1",
			Int:  stringPtr("10"),
		},
		{
			Name: "var2",
			Eval: stringPtr("23 + {{.var1}}"),
		},
		{
			Name: "bool1",
			Eval: stringPtr("{{.var1}} > 30"),
		},
		{
			Name: "bool2",
			Eval: stringPtr("{{.bool1}} && {{.var1}} < 50"),
		},
		{
			Name: "bool3",
			Eval: stringPtr("{{.var1}} > 0"),
		},
		{
			Name: "var3",
			Eval: stringPtr("{{.yamlvar.val}} + 1"),
		},
	}

	expected = map[string]interface{}{
		"boolvar":   true,
		"intvar":    42,
		"floatvar":  3.14,
		"stringvar": "some text",
		"home":      os.Getenv("HOME"),
		"testdata":  "testdata",
		"filevar":   "hello world",
		"jsonvar": map[string]interface{}{
			"foo": "bar",
		},
		"yamlvar": map[string]interface{}{
			"foo": "bar",
			"baz": "qux",
			"val": 42,
		},
		"yamlvar2": map[string]interface{}{
			"foo": "bar",
		},
		"upper": "HELLO WORLD",
		"var1":  10,
		"var2":  float64(33),
		"bool1": false,
		"bool2": false,
		"bool3": true,
		"var3":  float64(43),
	}
	result, err = Parse(variables, false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Unexpected result. Expected: %v, Got: %v", expected, result)
	}
}

func stringPtr(s string) *string {
	return &s
}
