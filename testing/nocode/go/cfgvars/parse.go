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

func Parse(variables []Variable, supressErrors bool) (map[string]interface{}, error) {
	// Initialize a map to store variable values
	varMap := make(map[string]interface{})

	// Process the variables
	for _, variable := range variables {
		if variable.Name == "" {
			// fmt.Println("Error: Variable name is empty")
			continue
		}

		var value interface{}

		templatedText := ""

		if variable.String != nil {
			templatedText = *variable.String
		} else if variable.File != nil {
			templatedText = *variable.File
		} else if variable.Shell != nil {
			templatedText = *variable.Shell
		} else if variable.Env != nil {
			templatedText = *variable.Env
		} else if variable.Json != nil {
			templatedText = *variable.Json
		} else if variable.Yaml != nil {
			templatedText = *variable.Yaml
		} else if variable.Eval != nil {
			templatedText = *variable.Eval
		} else if variable.Bool != nil {
			templatedText = *variable.Bool
		} else if variable.Int != nil {
			templatedText = *variable.Int
		} else if variable.Float != nil {
			templatedText = *variable.Float
		}

		if templatedText != "" {
			// Use text/template to replace placeholders
			tmpl, err := template.New("var").Parse(templatedText)
			if err == nil {
				varBuf := &strings.Builder{}
				if err := tmpl.Execute(varBuf, varMap); err == nil {
					templatedText = varBuf.String()
				}
			}
		}

		if variable.Bool != nil {
			// Parse the boolean value
			lowerText := strings.ToLower(templatedText)

			if len(lowerText) > 0 && (lowerText[0] == 't' || lowerText[0] == 'y' || lowerText == "on" || lowerText == "1") {
				value = true
			} else if len(lowerText) > 0 && (lowerText[0] == 'f' || lowerText[0] == 'n' || lowerText == "off" || lowerText == "0") {
				value = false
			} else {
				if !supressErrors {
					return nil, fmt.Errorf("failed to parse boolean for variable %s: %v", variable.Name, templatedText)
				}
				continue
			}
		} else if variable.String != nil {
			value = templatedText
		} else if variable.Int != nil {
			// Parse the integer value
			var err error
			value, err = strconv.Atoi(templatedText)
			if err != nil {
				if !supressErrors {
					return nil, fmt.Errorf("failed to parse integer for variable %s: %v", variable.Name, err)
				}
				continue
			}
		} else if variable.Float != nil {
			// Parse the float value
			var err error
			value, err = strconv.ParseFloat(templatedText, 64)
			if err != nil {
				if !supressErrors {
					return nil, fmt.Errorf("failed to parse float for variable %s: %v", variable.Name, err)
				}
				continue
			}
		} else if variable.File != nil {
			// Read the file content
			content, err := ioutil.ReadFile(templatedText)
			if err != nil {
				if !supressErrors {
					return nil, fmt.Errorf("failed to read file for variable %s: %v", variable.Name, err)
				}
				continue
			}
			value = string(content)
		} else if variable.Shell != nil {
			// Execute the shell command and capture the output
			cmd := exec.Command("sh", "-c", templatedText)
			cmd.Stderr = os.Stderr
			out, err := cmd.Output()
			if err != nil {
				if !supressErrors {
					return nil, fmt.Errorf("failed to execute shell command for variable %s: %v", variable.Name, err)
				}
				continue
			}
			value = string(out)
		} else if variable.Env != nil {
			// Read the environmental variable
			value = os.Getenv(templatedText)
		} else if variable.Json != nil {
			// Parse the JSON content into a map
			var jsonMap map[string]interface{}
			if err := json.Unmarshal([]byte(templatedText), &jsonMap); err != nil {
				if !supressErrors {
					return nil, fmt.Errorf("failed to parse JSON for variable %s: %v", variable.Name, err)
				}
				continue
			}
			value = jsonMap
		} else if variable.Yaml != nil {
			// Parse the YAML content into a map
			var yamlMap map[string]interface{}
			if err := yaml.Unmarshal([]byte(templatedText), &yamlMap); err != nil {
				if !supressErrors {
					return nil, fmt.Errorf("failed to parse YAML for variable %s: %v", variable.Name, err)
				}
				continue
			}
			value = yamlMap
		} else if variable.Eval != nil {
			expression, err := govaluate.NewEvaluableExpression(templatedText)
			if err != nil {
				if !supressErrors {
					return nil, fmt.Errorf("failed to parse expression for variable %s: %v", variable.Name, err)
					continue
				}
			}

			result, err := expression.Evaluate(varMap)
			if err != nil {
				if !supressErrors {
					return nil, fmt.Errorf("failed to evaluate expression for variable %s: %v", variable.Name, err)
				}
				continue
			}

			value = result
		}

		// Store the variable value in the map
		varMap[variable.Name] = value
	}

	return varMap, nil
}

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
