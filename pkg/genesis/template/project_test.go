package template

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
	"testing"
)

func loadOptions(filename string) (options map[string]string, err error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "issue opening file: %s", filename)
	}

	err = yaml.Unmarshal(content, &options)
	if err != nil {
		return nil, errors.Wrapf(err, "issue unmarshalling file: %s", filename)
	}

	return options, nil
}

func TestGenesisTemplateApi_ReplaceFilters(t *testing.T) {
	optionsMap := make(map[string]string, 3)
	// upper case filter
	keyUpper := "{{mykey | upper}}"
	// lower case filter
	keyLower := "{{mykey | lower}}"
	// no filter
	keyNone := "{{mykey}}"
	value := "my_VALue"
	optionsMap[keyUpper] = value
	optionsMap[keyLower] = value
	optionsMap[keyNone] = value

	valUpper, err := ReplaceGenesisVariable(keyUpper, value)

	if err != nil {
		t.Errorf("problem replacing variables %+v.", err)
	}
	valLower, err := ReplaceGenesisVariable(keyLower, value)

	if err != nil {
		t.Errorf("problem replacing variables %+v.", err)
	}
	valNone, err := ReplaceGenesisVariable(keyNone, value)

	if err != nil {
		t.Errorf("problem replacing variables %+v.", err)
	}

	assert.Equal(t, valUpper, strings.ToUpper(value), "should have been converted to all upper case")
	assert.Equal(t, valLower, strings.ToLower(value), "should have been converted to all lower case")
	assert.Equal(t, valNone, value, "should have no transformation")

}

func TestRecursiveReplace_NoErrors(t *testing.T) {
	// setup
	optionsMap := make(map[string]string, 3)
	// title variable
	keyTitle := "title"
	// uppercase variable
	keyUpper := "uppercase"
	// lowercase variable
	keyLower := "lowercase"
	upperValue := "my_VALue"
	lowerValue := "my_lower_value"
	titleValue := "my Title"
	optionsMap[keyUpper] = upperValue
	optionsMap[keyLower] = lowerValue
	optionsMap[keyTitle] = titleValue
	// read file
	inputBytes, err := ioutil.ReadFile("./test_data/test_document_no_errors.txt")
	if err != nil {
		t.Errorf("problem reading file %+v", err)
	}

	output, err := RecursiveReplace(inputBytes, optionsMap)

	if err != nil {
		t.Errorf("problem with RecursiveReplace %+v", err)
	}

	indexOf := bytes.Index(output, []byte("{{"))
	assert.Equal(t, -1, indexOf, "there should be no double-brackets in the output")
	indexOf = bytes.Index(output, []byte("}}"))
	assert.Equal(t, -1, indexOf, "there should be no double-brackets in the output")
}

func TestRecursiveReplace_BadInputError(t *testing.T) {
	// setup
	optionsMap := make(map[string]string, 3)
	// title variable
	keyTitle := "title"
	// malformed variable
	keyBad := "bad_var"
	badValue := "my_VALue"
	titleValue := "my Title"
	optionsMap[keyBad] = badValue
	optionsMap[keyTitle] = titleValue

	// read file
	inputBytes, err := ioutil.ReadFile("./test_data/test_document_bad_input.txt")
	if err != nil {
		t.Errorf("problem reading file %+v", err)
	}

	_, err = RecursiveReplace(inputBytes, optionsMap)

	assert.NotNil(t, err, "there should be an error because the input is malformed")
}

func TestStringRecursiveReplace_NoError(t *testing.T) {
	// setup
	optionsMap := make(map[string]string, 3)
	// title variable
	keyTitle := "title"
	// uppercase variable
	keyUpper := "uppercase"
	// lowercase variable
	keyLower := "lowercase"
	upperValue := "my_VALue"
	lowerValue := "my_lower_value"
	titleValue := "my Title"
	optionsMap[keyUpper] = upperValue
	optionsMap[keyLower] = lowerValue
	optionsMap[keyTitle] = titleValue
	// read file

	inputString := "{{title | upper}} is the title. {{uppercase | upper}} is an uppercase filter, but {{lowercase | lower}} is lower case."
	output, err := StringRecursiveReplace(inputString, optionsMap)

	if err != nil {
		t.Errorf("problem with RecursiveReplace %+v", err)
	}

	indexOf := strings.Index(output, "{{")
	assert.Equal(t, -1, indexOf, "there should be no double-brackets in the output")
	indexOf = strings.Index(output, "}}")
	assert.Equal(t, -1, indexOf, "there should be no double-brackets in the output")
}
