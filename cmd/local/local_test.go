package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func Test_folderWithSeparator(t *testing.T) {
	const sep = string(os.PathSeparator)
	type testCase struct {
		Input    string
		Expected string
	}
	cases := []testCase{
		{"myfolder", "myfolder" + sep},
		{"myfolder" + sep, "myfolder" + sep},
		{"myfolder" + sep + "sub", "myfolder" + sep + "sub" + sep},
		{"myfolder" + sep + "sub" + sep, "myfolder" + sep + "sub" + sep},
		{"", ""},
	}
	for _, testCase := range cases {
		value := folderWithSeparator(testCase.Input)
		if value != testCase.Expected {
			t.Errorf("Expected '%s' to equal '%s'", value, testCase.Expected)
		}
	}
}

func Test_getOptionsFrom(t *testing.T) {
	content := []byte(`
source: base
template_name: Demo template
settings:
  a: Value 1
  c: Value 2
`)
	file, err := ioutil.TempFile(os.TempDir(), "genesis-test")
	if err != nil {
		t.Errorf("Cannot perform test because %s", err)
	}
	file.Close()
	ioutil.WriteFile(file.Name(), content, 0644)
	settings, err := getOptionsFrom(file.Name())
	if err != nil {
		t.Errorf("Cannot perform test because %s", err)
	}
	if settings.Source != "base" {
		t.Errorf("Expected source field '%s' to equal base", settings.Source)
	}
	if settings.TemplateName != "Demo template" {
		t.Errorf("Expected template_name field '%s' to equal 'Demo template'", settings.TemplateName)
	}
	if settings.ConfigurationMap["a"] != "Value 1" {
		t.Errorf("Expected a settings field '%s' to equal 'Value 1'", settings.ConfigurationMap["a"])
	}
}
