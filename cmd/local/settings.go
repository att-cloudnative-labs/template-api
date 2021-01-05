package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Settings schema for configuration files (i.e. options.test.yaml)
type Settings struct {
	Source           string            `yaml:"source"`
	TemplateName     string            `yaml:"template_name"`
	ConfigurationMap map[string]string `yaml:"settings"`
}

// getOptionsFrom Returns a map of values from a yaml file
func getOptionsFrom(settingsFile string) (Settings, error) {
	content, err := ioutil.ReadFile(settingsFile)
	if err != nil {
		return Settings{}, err
	}
	var options Settings
	err = yaml.Unmarshal(content, &options)
	return options, err
}
