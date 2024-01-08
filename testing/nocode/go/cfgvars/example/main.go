package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"gitm/cfgvars"
)

type Config struct {
	Vars []cfgvars.Variable `yaml:"vars"`
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
	vars, err := cfgvars.ParseToSlice(config.Vars, false)

	// Process the variables
	for _, variable := range vars {
		// You can now use the 'value' for further processing
		fmt.Printf("Variable %s: %v\n", variable.Name, variable.Value)
	}
}
