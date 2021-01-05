// Copyright © 2019 AT&T
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

	"github.com/att-cloudnative-labs/template-api/pkg/genesis/template"
)

func main() {
	target := flag.String("target", "", "Target directory (automatic temporary directory if empty")
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

	err = os.Chdir(targetFolder + string(os.PathSeparator) + opts.Source)
	terminateOnError("Cannot switch to target directory", err)
	project := customProject{opts.Source, opts.TemplateName, opts.ConfigurationMap}
	err = tpl.GenerateFromTemplate(project, opts.ConfigurationMap)

	terminateOnError("Cannot produce project", err)
}
