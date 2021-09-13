package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type RepoConfig struct {
	Name      string `json:"name" yaml:"name"`
	URL       string `json:"url" yaml:"url"`
	Path      string `json:"path" yaml:"path"`
	LFS       bool   `json:"lfs" yaml:"lfs"`
	Depth     int    `json:"depth" yaml:"depth"`
	Autostash bool   `json:"autostash" yaml:"autostash"`

	Timeout int `json:"timeout" yaml:"timeout"`
}

//Config holds the configuration struct
//generate go struct with https://yaml.to-go.online/
//search regex `yaml:"(.*)"` replace with `json:"$1" yaml:"$1"`
type Config struct {
	Debug bool         `json:"debug" yaml:"debug"`
	Mock  bool         `json:"mock" yaml:"mock"`
	Repos []RepoConfig `json:"repos" yaml:"repos"`
}

func (g *Gitem) LoadConfig(cfg_path string) error {
	data, err := ioutil.ReadFile(cfg_path)
	if err != nil {
		return err
	}

	// TODO detect extension and use json
	err = yaml.Unmarshal([]byte(data), &g.config)
	if err != nil {
		return err
	}

	for i := range g.config.Repos {
		if g.config.Repos[i].Name == "" {
			g.config.Repos[i].Name = g.config.Repos[i].URL
		}
	}

	return nil
}
