package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

type Variable struct {
	Name   string   `yaml:"name"`
	Bool   *bool    `yaml:"bool,omitempty"`
	Int    *int     `yaml:"int,omitempty"`
	Float  *float64 `yaml:"float,omitempty"`
	String *string  `yaml:"string,omitempty"`
	Shell  *string  `yaml:"shell,omitempty"`
	Env    *string  `yaml:"env,omitempty"`
	File   *string  `yaml:"file,omitempty"`
	Json   *string  `yaml:"json,omitempty"`
	Yaml   *string  `yaml:"yaml,omitempty"`
}

type Config struct {
	Vars []Variable `yaml:"vars"`
}

func main() {
	// Read the YAML file
	yamlFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		fmt.Println("Error reading YAML file:", err)
		return
	}

	// Parse the YAML content into a Config struct
	var config Config
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		fmt.Println("Error parsing YAML:", err)
		return
	}

	// Initialize a map to store variable values
	varMap := make(map[string]interface{})

	// Process the variables
	for _, variable := range config.Vars {
		var value interface{}

		if variable.Bool != nil {
			value = *variable.Bool
		} else if variable.String != nil {
			text := *variable.String

			// Use text/template to replace placeholders
			tmpl, err := template.New("var").Parse(text)
			if err == nil {
				varBuf := &strings.Builder{}
				if err := tmpl.Execute(varBuf, varMap); err == nil {
					text = varBuf.String()
				}
			}

			value = text
		} else if variable.Int != nil {
			value = *variable.Int
		} else if variable.Float != nil {
			value = *variable.Float
		} else if variable.File != nil {
			text := *variable.File

			// Use text/template to replace placeholders
			tmpl, err := template.New("var").Parse(text)
			if err == nil {
				varBuf := &strings.Builder{}
				if err := tmpl.Execute(varBuf, varMap); err == nil {
					text = varBuf.String()
				}
			}

			// Read the file content
			content, err := ioutil.ReadFile(text)
			if err != nil {
				fmt.Printf("Error reading file for variable %s: %v\n", variable.Name, err)
				continue
			}
			value = string(content)
		} else if variable.Shell != nil {
			text := *variable.Shell

			// Use text/template to replace placeholders
			tmpl, err := template.New("var").Parse(text)
			if err == nil {
				varBuf := &strings.Builder{}
				if err := tmpl.Execute(varBuf, varMap); err == nil {
					text = varBuf.String()
				}
			}

			// Execute the shell command and capture the output
			cmd := exec.Command("sh", "-c", text)
			cmd.Stderr = os.Stderr
			out, err := cmd.Output()
			if err != nil {
				fmt.Printf("Error executing shell command for variable %s: %v\n", variable.Name, err)
				continue
			}
			value = string(out)
		} else if variable.Env != nil {
			text := *variable.Env

			// Use text/template to replace placeholders
			tmpl, err := template.New("var").Parse(text)
			if err == nil {
				varBuf := &strings.Builder{}
				if err := tmpl.Execute(varBuf, varMap); err == nil {
					text = varBuf.String()
				}
			}

			// Read the environmental variable
			value = os.Getenv(text)
		} else if variable.Json != nil {
			text := *variable.Json

			// Use text/template to replace placeholders
			tmpl, err := template.New("var").Parse(text)
			if err == nil {
				varBuf := &strings.Builder{}
				if err := tmpl.Execute(varBuf, varMap); err == nil {
					text = varBuf.String()
				}
			}

			// Parse the JSON content into a map
			var jsonMap map[string]interface{}
			if err := json.Unmarshal([]byte(text), &jsonMap); err != nil {
				fmt.Printf("Error parsing JSON for variable %s: %v\n", variable.Name, err)
				continue
			}
			value = jsonMap
		} else if variable.Yaml != nil {
			text := *variable.Yaml

			// Use text/template to replace placeholders
			tmpl, err := template.New("var").Parse(text)
			if err == nil {
				varBuf := &strings.Builder{}
				if err := tmpl.Execute(varBuf, varMap); err == nil {
					text = varBuf.String()
				}
			}

			// Parse the YAML content into a map
			var yamlMap map[string]interface{}
			if err := yaml.Unmarshal([]byte(text), &yamlMap); err != nil {
				fmt.Printf("Error parsing YAML for variable %s: %v\n", variable.Name, err)
				continue
			}
			value = yamlMap
		}

		// Store the variable value in the map
		varMap[variable.Name] = value

		// You can now use the 'value' for further processing
		fmt.Printf("Variable %s: %v\n", variable.Name, value)
	}
}
