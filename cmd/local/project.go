package main

import "github.com/att-cloudnative-labs/template-api/pkg/genesis/template"

type CustomProject struct {
	Root              string
	TemplateSubFolder string
	Options           map[string]string
}

func (d CustomProject) GetRequiredOptions() []template.Option {
	return []template.Option{}
}

func (d CustomProject) SetValidatedOptions(args map[string]string) error {
	return nil
}

func (d CustomProject) GetValidatedOptions() (map[string]string, error) {
	return d.Options, nil
}

func (d CustomProject) GetRoot() (string, error) {
	return d.TemplateSubFolder, nil
}

func (d CustomProject) GetName() string {
	return "GoATT Base Template"
}

func (d CustomProject) OrganizeGroups() error {
	return nil
}
