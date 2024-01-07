package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type RepoConfig struct {
	Name         string `yaml:"name"`
	URL          string `yaml:"url"`
	Path         string `yaml:"path"`
	UsesLFS      bool   `yaml:"uses-lfs"`
	UsesSubrepos bool   `yaml:"uses-submodules"`
}

type LoggingConfig struct {
	Verbose      *bool   `yaml:"verbose"`
	Timestamps   *bool   `yaml:"timestamps"`
	StdoutPrefix *string `yaml:"stdout-prefix"`
	StderrPrefix *string `yaml:"stderr-prefix"`
	Commands     *bool   `yaml:"commands"`
	Duration     *bool   `yaml:"duration"`
	Begin        *bool   `yaml:"begin"`

	Color          *bool   `yaml:"color"`
	CommandColor   *string `yaml:"command-color"`
	StdoutColor    *string `yaml:"stdout-color"`
	StderrColor    *string `yaml:"stderr-color"`
	TimestampColor *string `yaml:"timestamp-color"`
	DurationColor  *string `yaml:"duration-color"`
	BeginColor     *string `yaml:"begin-color"`
}

type GitmConfig struct {
	Logging LoggingConfig `yaml:"logging"`
	Repos   []RepoConfig  `yaml:"repos"`
}

func findConfigFiles(dir string) []string {
	var configFiles []string

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), "gitm.yml") || strings.HasSuffix(info.Name(), "gitm.yaml") || strings.HasSuffix(info.Name(), ".gitm.yml") || strings.HasSuffix(info.Name(), ".gitm.yaml") {
			configFiles = append(configFiles, path)
		}

		return nil
	})

	return configFiles
}

func cfgFind() (string, error) {
	dir, _ := os.Getwd()

	for {
		// files, _ := filepath.Glob(filepath.Join(dir, "{gitm,.gitm}.{yml,yaml}"))
		files := findConfigFiles(dir)
		if len(files) > 0 {
			return files[0], nil
		}

		newDir := filepath.Dir(dir)
		if newDir == dir {
			break
		}

		dir = newDir
	}

	return "", fmt.Errorf("config file not found")
}

func loadConfig(configFile string) (*GitmConfig, error) {
	var err error
	if configFile == "" {
		// recursively search up for config file
		configFile, err = cfgFind()
		if err != nil {
			return nil, err
		}

		if configFile == "" {
			return nil, fmt.Errorf("config file not found")
		}
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config GitmConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
