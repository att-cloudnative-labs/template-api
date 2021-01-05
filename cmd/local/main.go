// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/att-cloudnative-labs/template-api/pkg/genesis/template"
)

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
	sep := string(os.PathSeparator)
	files := []string{".genesis.yml"}
	for _, file := range files {
		err := copyFile(file, fmt.Sprintf("%s%s.genesis.yml", target, sep))
		if err != nil {
			return err
		}
	}
	templateFolder := "base"
	err := os.Mkdir(fmt.Sprintf("%s%s%s", target, sep, templateFolder), 0700)
	if err != nil {
		return err
	}
	currDir, _ := os.Getwd()
	root := fmt.Sprintf("%s%s%s", currDir, sep, templateFolder)
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		stripped := target + sep + templateFolder + strings.Replace(path, root, "", -1)
		if info.IsDir() {
			err = os.MkdirAll(stripped, 0700)
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

// getOptionsFrom Returns a map of values from a yaml file
func getOptionsFrom(settingsFile string) (map[string]string, error) {
	content, err := ioutil.ReadFile(settingsFile)
	if err != nil {
		return nil, err
	}
	var options map[string]string
	err = yaml.Unmarshal(content, &options)
	return options, err
}

func main() {
	target := flag.String("target", "", "Target directory (automatic temporary directory if empty")
	source := flag.String("source", "", "Templates directory")
	workDir := flag.String("wd", ".", "Initial working directory (root of .genesis.yml)")
	options := flag.String("options", "options.test.yaml", "Options file")
	flag.Parse()
	opts, err := getOptionsFrom(*options)
	terminateOnError("Cannot read project properties", err)
	targetFolder := *target
	if targetFolder == "" {
		tmpFolder, err := ioutil.TempDir(os.TempDir(), "genesis-")
		terminateOnError("Cannot create temporary folder", err)
		targetFolder = tmpFolder
	}
	if *workDir != "." {
		err := os.Chdir(*workDir)
		terminateOnError("Cannot switch to initial working directory", err)
	}
	fmt.Printf("Creating project in %s\n", targetFolder)

	err = copyProject(targetFolder)
	terminateOnError("Cannot create copy of project", err)
	tpl := template.NewGenesisTemplateApi(folderWithSeparator(targetFolder))

	os.Chdir(targetFolder + string(os.PathSeparator) + *source)
	err = tpl.GenerateFromTemplate(CustomProject{targetFolder, *source, opts}, opts)

	terminateOnError("Cannot produce project", err)
}
