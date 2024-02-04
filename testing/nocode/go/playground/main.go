package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"text/template"

	"github.com/go-yaml/yaml"
)

type ShellVar struct {
	sh string `yaml:"sh"`
}

type Command struct {
	Description string   `yaml:"description"`
	Command     string   `yaml:"command"`
	Group       []string `yaml:"group"`
}

// InGroup checks if a command is in a group
func (c *Command) InGroup(group string) bool {
	for _, g := range c.Group {
		if g == group {
			return true
		}
	}
	return false
}

type Project struct {
	Vars  map[string]interface{} `yaml:"vars"`
	Group []string               `yaml:"group"`
}

// InGroup checks if a project is in a group
func (p *Project) InGroup(group string) bool {
	for _, g := range p.Group {
		if g == group {
			return true
		}
	}
	return false
}

type Config struct {
	Vars     map[string]interface{} `yaml:"vars"`
	Commands map[string]Command     `yaml:"commands"`
	Projects map[string]Project     `yaml:"projects"`
}

func joinPath(parts ...interface{}) string {
	var pathParts []string
	for _, part := range parts {
		pathParts = append(pathParts, fmt.Sprintf("%v", part))
	}
	return path.Join(pathParts...)
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

type LoadedCommand struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Command     string   `yaml:"command"`
	Group       []string `yaml:"group"`
}

type LoadedProject struct {
	Name  string                 `yaml:"name"`
	Vars  map[string]interface{} `yaml:"vars"`
	Group []string               `yaml:"group"`
}

type LoadedConfig struct {
	Vars     map[string]interface{}   `yaml:"vars"`
	Commands map[string]LoadedCommand `yaml:"commands"`
	Projects map[string]LoadedProject `yaml:"projects"`
}

func LoadedConfigToMap(config *LoadedConfig) map[string]interface{} {
	dataMap := make(map[string]interface{})
	dataMap["vars"] = config.Vars
	dataMap["commands"] = make(map[string]interface{})
	dataMap["projects"] = make(map[string]interface{})

	for key, command := range config.Commands {
		commandMap := make(map[string]interface{})
		commandMap["name"] = command.Name
		commandMap["description"] = command.Description
		commandMap["command"] = command.Command
		commandMap["group"] = command.Group

		dataMap["commands"].(map[string]interface{})[key] = commandMap
	}

	for key, project := range config.Projects {
		projectMap := make(map[string]interface{})
		projectMap["name"] = project.Name
		projectMap["vars"] = project.Vars
		projectMap["group"] = project.Group

		dataMap["projects"].(map[string]interface{})[key] = projectMap
	}

	return dataMap
}

func LoadConfig(configFile string) (*LoadedConfig, error) {
	// Read the YAML file
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("Error reading YAML file: %v", err)
	}

	// Parse the YAML data into a Config struct
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("Error parsing YAML: %v", err)
	}

	// Define a custom template function map
	funcMap := template.FuncMap{
		"joinPath": joinPath,
	}

	loadedConfig := LoadedConfig{
		Vars:     make(map[string]interface{}),
		Commands: make(map[string]LoadedCommand),
		Projects: make(map[string]LoadedProject),
	}

	// Resolve variables from 'vars' field
	for key, value := range config.Vars {
		resolvedValue, err := resolveVar(value)
		if err != nil {
			return nil, fmt.Errorf("Error resolving variable %s: %v", key, err)
		}

		// if the resolveValue is a string, we can parse it as a template
		valueType := reflect.TypeOf(resolvedValue)
		if valueType == reflect.TypeOf("") {
			// execute the template with the resolved value and our latest dataMap
			tpl, err := template.New("commandTemplate").Funcs(funcMap).Parse(resolvedValue.(string))
			if err != nil {
				return nil, fmt.Errorf("Error parsing template: %v", err)
			}

			dataMap := LoadedConfigToMap(&loadedConfig)

			// Execute the template
			var buf bytes.Buffer
			err = tpl.Execute(&buf, dataMap)
			if err != nil {
				return nil, fmt.Errorf("Error executing template: %v", err)
			}

			resolvedValue = buf.String()
		}

		loadedConfig.Vars[key] = resolvedValue
	}

	for projectName, project := range config.Projects {
		// Make a copy of dataMap
		dataMap := LoadedConfigToMap(&loadedConfig)

		projectMap := make(map[string]interface{})
		projectMap["name"] = projectName
		projectMap["vars"] = make(map[string]interface{})
		projectMap["group"] = project.Group

		for key, value := range project.Vars {
			projectMap["vars"].(map[string]interface{})[key] = value
		}

		for key, value := range project.Vars {
			resolvedValue, err := resolveVar(value)
			if err != nil {
				return nil, fmt.Errorf("Error resolving variable %s: %v", key, err)
			}

			// if the resolveValue is a string, we can parse it as a template
			valueType := reflect.TypeOf(resolvedValue)
			if valueType == reflect.TypeOf("") {
				// execute the template with the resolved value and our latest dataMap
				tpl, err := template.New("commandTemplate").Funcs(funcMap).Parse(resolvedValue.(string))
				if err != nil {
					return nil, fmt.Errorf("Error parsing template: %v", err)
				}

				// Add our self to the dataMap
				dataMap["self"] = projectMap

				// Execute the template
				var buf bytes.Buffer
				err = tpl.Execute(&buf, dataMap)
				if err != nil {
					return nil, fmt.Errorf("Error executing template: %v", err)
				}

				resolvedValue = buf.String()
			}

			projectMap["vars"].(map[string]interface{})[key] = resolvedValue
		}

		loadedConfig.Projects[projectName] = LoadedProject{
			Name:  projectName,
			Vars:  projectMap["vars"].(map[string]interface{}),
			Group: project.Group,
		}
	}

	for commandName, command := range config.Commands {
		loadedConfig.Commands[commandName] = LoadedCommand{
			Name:        commandName,
			Description: command.Description,
			Command:     command.Command,
			Group:       command.Group,
		}
	}

	return &loadedConfig, nil
}

func (command *LoadedCommand) InGroup(groups []string) bool {
	for _, group := range groups {
		for _, g := range command.Group {
			if g == group {
				return true
			}
		}
	}
	return false
}

func (config *LoadedConfig) GetCommand(name string, projectName string) (string, error) {
	loadedCommand, ok := config.Commands[name]
	if !ok {
		return "", fmt.Errorf("Command %s not found", name)
	}

	loadedProject, ok := config.Projects[projectName]
	if !ok {
		return "", fmt.Errorf("Project %s not found", projectName)
	}

	if !loadedCommand.InGroup(loadedProject.Group) {
		return "", nil
	}

	dataMap := LoadedConfigToMap(config)

	commands, ok := dataMap["commands"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Commands not found")
	}

	projects, ok := dataMap["projects"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Projects not found")
	}

	command, ok := commands[name].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Command %s not found", name)
	}

	project, ok := projects[projectName].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Project %s not found", projectName)
	}

	commandStr, ok := command["command"].(string)
	if !ok {
		return "", fmt.Errorf("Command %s is not a string", name)
	}

	commandDataMap := make(map[string]interface{})
	commandDataMap["vars"] = dataMap["vars"]
	commandDataMap["project"] = project
	commandDataMap["args"] = strings.Join(os.Args[1:], " ")

	// Define a custom template function map
	funcMap := template.FuncMap{
		"joinPath": joinPath,
	}

	// execute the template with the resolved value and our latest dataMap
	tpl, err := template.New("commandTemplate").Funcs(funcMap).Parse(commandStr)
	if err != nil {
		return "", fmt.Errorf("Error parsing template: %v", err)
	}

	// Execute the template
	var buf bytes.Buffer
	err = tpl.Execute(&buf, commandDataMap)
	if err != nil {
		return "", fmt.Errorf("Error executing template: %v", err)
	}

	return buf.String(), nil

}

func main() {
	config, err := LoadConfig("config.yaml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Final config")
	fmt.Println(config)

	// simulate executing a command where we pass in a new dataMap
	command, err := config.GetCommand("git", "repo1")
	if err != nil {
		fmt.Printf("Error getting command: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Final command")
	fmt.Println(command)

}
