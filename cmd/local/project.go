package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/att-cloudnative-labs/template-api/pkg/genesis/template"
)

// customProject Pass-through project structure
type customProject struct {
	TemplateSubFolder string
	TemplateName      string
	Options           map[string]string
}

func (d customProject) GetRequiredOptions() []template.Option {
	return []template.Option{}
}

func (d customProject) SetValidatedOptions(args map[string]string) error {
	return nil
}

func (d customProject) GetValidatedOptions() (map[string]string, error) {
	return d.Options, nil
}

func (d customProject) GetRoot() (string, error) {
	return d.TemplateSubFolder, nil
}

func (d customProject) GetName() string {
	return d.TemplateName
}

func (d customProject) OrganizeGroups() error {
	return nil
}

// terminateOnError If err is not nil, it prints message an exits with code 1
func terminateOnError(message string, err error) {
	if err != nil {
		fmt.Printf("%s: %s\n", message, err)
		os.Exit(1)
	}
}

// copyFile creates a copy of source into destination
func copyFile(source string, destination string) error {
	content, err := ioutil.ReadFile(source)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(destination, content, 0644)
}

// copyProject Copies the .genesis.yml file and all the base folder to target
func copyProject(target string) error {
	var dirMode os.FileMode = 700
	sep := string(os.PathSeparator)
	files := []string{".genesis.yml"}
	for _, file := range files {
		err := copyFile(file, fmt.Sprintf("%s%s%s", target, sep, file))
		if err != nil {
			return err
		}
	}
	templateFolder := "base"
	err := os.Mkdir(fmt.Sprintf("%s%s%s", target, sep, templateFolder), dirMode)
	if err != nil {
		return err
	}
	currDir, _ := os.Getwd()
	root := fmt.Sprintf("%s%s%s", currDir, sep, templateFolder)
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		stripped := target + sep + templateFolder + strings.Replace(path, root, "", -1)
		if info.IsDir() {
			err = os.MkdirAll(stripped, dirMode)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(path, stripped)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// folderWithSeparator if folderName does not end with the OS path separator
// (i.e. / in linux), it appends such character to the end of the string
func folderWithSeparator(folderName string) string {
	separator := string(os.PathSeparator)
	if len(folderName) < len(separator) {
		return folderName
	}
	wantedIndex := len(folderName) - len(separator)
	if folderName[wantedIndex:] != separator {
		return folderName + separator
	}
	return folderName
}
