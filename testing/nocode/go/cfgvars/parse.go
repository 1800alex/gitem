package cfgvars

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"

	"github.com/Knetic/govaluate"
	"gopkg.in/yaml.v2"
)

type Variable struct {
	Name   string  `yaml:"name",json:"name"`
	Bool   *string `yaml:"bool,omitempty",json:"bool,omitempty"`
	Int    *string `yaml:"int,omitempty",json:"int,omitempty"`
	Float  *string `yaml:"float,omitempty",json:"float,omitempty"`
	String *string `yaml:"string,omitempty",json:"string,omitempty"`
	Shell  *string `yaml:"shell,omitempty",json:"shell,omitempty"`
	Env    *string `yaml:"env,omitempty",json:"env,omitempty"`
	File   *string `yaml:"file,omitempty",json:"file,omitempty"`
	Json   *string `yaml:"json,omitempty",json:"json,omitempty"`
	Yaml   *string `yaml:"yaml,omitempty",json:"yaml,omitempty"`
	Eval   *string `yaml:"eval,omitempty",json:"eval,omitempty"`
}

type ParsedVariable struct {
	Name  string      `yaml:"name",json:"name"`
	Value interface{} `yaml:"value",json:"value"`
}

// TODO Make a Parser struct
// TODO Make parser struct hold it's own data map and supressErrors flag
// TODO Make parser struct allow custom template engines

// templateText replaces placeholders in the given text with values from the given data map.
func templateText(text string, data map[string]interface{}) string {
	if text != "" {
		// Use text/template to replace placeholders
		tmpl, err := template.New("var").Parse(text)
		if err == nil {
			varBuf := &strings.Builder{}
			if err := tmpl.Execute(varBuf, data); err == nil {
				return varBuf.String()
			}
		}
	}

	return text
}

// ParseBool parses the given text as a boolean value after replacing placeholders with values from the given data map.
func ParseBool(text string, data map[string]interface{}) (bool, error) {
	// Parse the boolean value
	lowerText := strings.ToLower(templateText(text, data))

	if len(lowerText) > 0 && (lowerText[0] == 't' || lowerText[0] == 'y' || lowerText == "on" || lowerText == "1") {
		return true, nil
	} else if len(lowerText) > 0 && (lowerText[0] == 'f' || lowerText[0] == 'n' || lowerText == "off" || lowerText == "0") {
		return false, nil
	}

	return false, fmt.Errorf("failed to parse boolean: %v", text)
}

// ParseInt parses the given text as an integer value after replacing placeholders with values from the given data map.
func ParseInt(text string, data map[string]interface{}) (int, error) {
	// Parse the integer value
	var err error
	text = templateText(text, data)
	value, err := strconv.Atoi(text)
	if err != nil {
		return 0, fmt.Errorf("failed to parse integer: %v", err)
	}
	return value, nil
}

// ParseFloat parses the given text as a float value after replacing placeholders with values from the given data map.
func ParseFloat(text string, data map[string]interface{}) (float64, error) {
	// Parse the float value
	var err error
	text = templateText(text, data)
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse float: %v", err)
	}
	return value, nil
}

// ParseString parses the given text after replacing placeholders with values from the given data map.
func ParseString(text string, data map[string]interface{}) (string, error) {
	return templateText(text, data), nil
}

// ParseFile reads the given file after replacing placeholders with values from the given data map.
func ParseFile(text string, data map[string]interface{}) (string, error) {
	// Read the file content
	text = templateText(text, data)
	content, err := ioutil.ReadFile(text)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}
	return string(content), nil
}

// ParseShell executes the given shell command after replacing placeholders with values from the given data map.
func ParseShell(text string, data map[string]interface{}) (string, error) {
	// Execute the shell command and capture the output
	text = templateText(text, data)
	cmd := exec.Command("sh", "-c", text)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute shell command: %v", err)
	}
	return string(out), nil
}

// ParseEnv reads the given environmental variable after replacing placeholders with values from the given data map.
func ParseEnv(text string, data map[string]interface{}) (string, error) {
	// Read the environmental variable
	text = templateText(text, data)
	return os.Getenv(text), nil
}

// ParseJson parses the given text as JSON after replacing placeholders with values from the given data map.
func ParseJson(text string, data map[string]interface{}) (map[string]interface{}, error) {
	// Parse the JSON content into a map
	text = templateText(text, data)
	var jsonMap map[string]interface{}
	if err := json.Unmarshal([]byte(text), &jsonMap); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}
	return jsonMap, nil
}

// ParseYaml parses the given text as YAML after replacing placeholders with values from the given data map.
func ParseYaml(text string, data map[string]interface{}) (map[string]interface{}, error) {
	// Parse the YAML content into a map
	text = templateText(text, data)
	var yamlMap map[string]interface{}
	if err := yaml.Unmarshal([]byte(text), &yamlMap); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}
	return yamlMap, nil
}

// ParseEval parses the given text as an expression after replacing placeholders with values from the given data map.
func ParseEval(text string, data map[string]interface{}) (interface{}, error) {
	text = templateText(text, data)
	expression, err := govaluate.NewEvaluableExpression(text)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %v", err)
	}

	result, err := expression.Evaluate(data)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %v", err)
	}

	return result, nil
}

// ParseVariable parses the given variable after replacing placeholders with values from the given data map.
func ParseVariable(variable Variable, data map[string]interface{}) (ParsedVariable, error) {
	value := ParsedVariable{
		Name: variable.Name,
	}

	var err error

	if variable.Bool != nil {
		value.Value, err = ParseBool(*variable.Bool, data)
	} else if variable.String != nil {
		value.Value, err = ParseString(*variable.String, data)
	} else if variable.Int != nil {
		value.Value, err = ParseInt(*variable.Int, data)
	} else if variable.Float != nil {
		value.Value, err = ParseFloat(*variable.Float, data)
	} else if variable.File != nil {
		value.Value, err = ParseFile(*variable.File, data)
	} else if variable.Shell != nil {
		value.Value, err = ParseShell(*variable.Shell, data)
	} else if variable.Env != nil {
		value.Value, err = ParseEnv(*variable.Env, data)
	} else if variable.Json != nil {
		value.Value, err = ParseJson(*variable.Json, data)
	} else if variable.Yaml != nil {
		value.Value, err = ParseYaml(*variable.Yaml, data)
	} else if variable.Eval != nil {
		value.Value, err = ParseEval(*variable.Eval, data)
	} else {
		err = fmt.Errorf("no known value specified for variable: %v", variable.Name)
	}

	return value, err
}

// Parse parses the given variables to a map replacing placeholders with values from the given data map.
func Parse(variables []Variable, supressErrors bool) (map[string]interface{}, error) {
	// Initialize a map to store variable values
	varMap := make(map[string]interface{})

	// Process the variables
	for _, variable := range variables {
		if variable.Name == "" {
			if supressErrors {
				continue
			}
			return nil, fmt.Errorf("variable name is empty")
		}

		// Parse the variable
		value, err := ParseVariable(variable, varMap)
		if err != nil {
			if supressErrors {
				continue
			}
			return nil, err
		}

		// Store the variable value in the map
		varMap[variable.Name] = value.Value
	}

	return varMap, nil
}

// ParseToSlice parses the given variables to a slice preserving the order of variables after replacing placeholders with values from the given data map.
func ParseToSlice(variables []Variable, supressErrors bool) ([]ParsedVariable, error) {
	varMap, err := Parse(variables, supressErrors)
	if err != nil {
		return nil, err
	}

	parsedVariables := make([]ParsedVariable, len(variables), len(variables))

	// Preserve the order of variables
	for i, variable := range variables {
		val, ok := varMap[variable.Name]
		if !ok {
			continue
		}
		parsedVariables[i] = ParsedVariable{
			Name:  variable.Name,
			Value: val,
		}
	}

	return parsedVariables, nil
}
