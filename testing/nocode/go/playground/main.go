package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"text/template"

	"github.com/go-yaml/yaml"
)

type ShellVar struct {
	sh string `yaml:"sh"`
}

type Config struct {
	Playground struct {
		Vars     map[string]interface{} `yaml:"vars"`
		Commands []struct {
			Name        string   `yaml:"name"`
			Description string   `yaml:"description"`
			Command     string   `yaml:"command"`
			Group       []string `yaml:"group"`
		} `yaml:"commands"`
		Projects []struct {
			Name  string                 `yaml:"name"`
			Vars  map[string]interface{} `yaml:"vars"`
			Group []string               `yaml:"group"`
		} `yaml:"projects"`
	} `yaml:"playground"`
}

func joinPath(parts ...interface{}) string {
	var pathParts []string
	for _, part := range parts {
		pathParts = append(pathParts, fmt.Sprintf("%v", part))
	}
	return strings.Join(pathParts, "/")
}

func resolveVar(value interface{}) (interface{}, error) {
	valueType := reflect.TypeOf(value)

	if valueType == reflect.TypeOf("") {
		// If it's a string, return it as is
		return value.(string), nil
	} else if valueType == reflect.TypeOf(map[interface{}]interface{}{}) {
		// If it's a map, iterate through key-value pairs
		valMap := value.(map[interface{}]interface{})
		if cmd, ok := valMap["sh"].(string); ok {
			cmdOutput, err := exec.Command("sh", []string{"-c", cmd}...).CombinedOutput()
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(string(cmdOutput)), nil
		}
	}

	return value, nil
}

func main() {
	// Read the YAML file
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		fmt.Printf("Error reading YAML file: %v\n", err)
		os.Exit(1)
	}

	// Parse the YAML data into a Config struct
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	// Define a custom template function map
	funcMap := template.FuncMap{
		"joinPath": joinPath,
	}

	// Create a new template with custom functions
	tpl, err := template.New("commandTemplate").Funcs(funcMap).Parse(config.Playground.Commands[0].Command)
	if err != nil {
		fmt.Printf("Error parsing template: %v\n", err)
		os.Exit(1)
	}

	// Resolve variables from 'vars' field
	dataMap := make(map[string]interface{})
	for key, value := range config.Playground.Vars {
		resolvedValue, err := resolveVar(value)
		if err != nil {
			fmt.Printf("Error resolving variable %s: %v\n", key, err)
			os.Exit(1)
		}

		// if the resolveValue is a string, we can parse it as a template
		valueType := reflect.TypeOf(resolvedValue)
		if valueType == reflect.TypeOf("") {
			// execute the template with the resolved value and our latest dataMap
			tpl, err := tpl.Parse(resolvedValue.(string))
			if err != nil {
				fmt.Printf("Error parsing template: %v\n", err)
				os.Exit(1)
			}

			// Execute the template
			var buf bytes.Buffer
			err = tpl.Execute(&buf, dataMap)
			if err != nil {
				fmt.Printf("Error executing template: %v\n", err)
				os.Exit(1)
			}

			resolvedValue = buf.String()
		}

		var vars map[string]interface{}
		varMap, ok := dataMap["vars"]
		if !ok {
			vars = make(map[string]interface{})
		} else {
			vars = varMap.(map[string]interface{})
		}

		vars[key] = resolvedValue
		dataMap["vars"] = vars
	}

	for _, project := range config.Playground.Projects {
		projectVars := make(map[string]interface{})
		projectVars["name"] = project.Name
		projectVars["group"] = project.Group

		for key, value := range project.Vars {
			resolvedValue, err := resolveVar(value)
			if err != nil {
				fmt.Printf("Error resolving variable %s: %v\n", key, err)
				os.Exit(1)
			}

			// if the resolveValue is a string, we can parse it as a template
			valueType := reflect.TypeOf(resolvedValue)
			if valueType == reflect.TypeOf("") {
				// execute the template with the resolved value and our latest dataMap
				tpl, err := tpl.Parse(resolvedValue.(string))
				if err != nil {
					fmt.Printf("Error parsing template: %v\n", err)
					os.Exit(1)
				}

				// Make a copy of dataMap
				dataMapCopy := make(map[string]interface{})
				for k, v := range dataMap {
					dataMapCopy[k] = v
				}

				// Add our self to the dataMap
				dataMapCopy["self"] = projectVars

				// fmt.Println("Executing template on", resolvedValue, "with", dataMapCopy)

				// Execute the template
				var buf bytes.Buffer
				err = tpl.Execute(&buf, dataMapCopy)
				if err != nil {
					fmt.Printf("Error executing template: %v\n", err)
					os.Exit(1)
				}

				// fmt.Println("templated", resolvedValue, "to", buf.String())

				resolvedValue = buf.String()
			}

			var vars map[string]interface{}
			varMap, ok := projectVars["vars"]
			if !ok {
				vars = make(map[string]interface{})
			} else {
				vars = varMap.(map[string]interface{})
			}

			vars[key] = resolvedValue
			projectVars["vars"] = vars
		}
		dataMap[project.Name] = projectVars
	}

	fmt.Println("Final dataMap")
	fmt.Println(dataMap)

}
