package template

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const genesisFileName = ".genesis.yml"

// The ProjectTemplateApi interface defines the behavior needed to retrieve, parse, and package
// a Genesis Template project.
type ProjectTemplateApi interface {
	// GetProjectNames returns the available template names in a Genesis template project.
	// The gitRepositoryUrl tells the function where to look for a .yml file.
	// It returns a slice of strings with each Project name, and any errors encountered
	GetProjectNames() ([]string, error)

	// GetProjectFromRepo retrieves a single template project from the Git repo at the given URL,
	// and any errors encountered.
	GetProjectFromRepo(projectName string) (ProjectTemplate, error)

	// GetProjectsFromRepo retrieves a list of genesis template projects from the Git repo at the
	// given URL, and any errors encountered.
	GetProjectsFromRepo() (GenesisProject, error)

	// GenerateFromTemplate creates a new project from the provided template, using the
	// variableReplacementMap to customize the Project, as needed.
	GenerateFromTemplate(project ProjectTemplate, variableReplacementMap map[string]string) error

	// ValidateGenesisProject goes out to gitRepositoryUrl and looks for a .yml file.
	// If it exists, then the method returns true. If not, false.
	// Also returns any errors encountered.
	ValidateGenesisProject() (bool, error)

	// Cleanup deletes temporary files and folders
	Cleanup() error
}

// Implement the ProjectTemplateApi
type GenesisTemplateApi struct {
	DirectoryPath string
}

func NewGenesisTemplateApi(directoryPath string) *GenesisTemplateApi {
	return &GenesisTemplateApi{
		DirectoryPath: directoryPath,
	}
}

func (gTemplateApi *GenesisTemplateApi) GetProjectNames() ([]string, error) {
	projects, err := gTemplateApi.GetProjectsFromRepo()
	if err != nil {
		return nil, err
	}

	var projectNames []string

	for _, project := range projects.Projects {
		projectNames = append(projectNames, project.Name)
	}

	return projectNames, nil
}

func (gTemplateApi *GenesisTemplateApi) GetProjectFromRepo(projectName string) (ProjectTemplate, error) {
	projects, err := gTemplateApi.GetProjectsFromRepo()
	if err != nil {
		return nil, err
	}
	var names = make([]interface{}, len(projects.Projects))
	for i, project := range projects.Projects {
		if project.Name == projectName {
			return &project, nil
		}
		names[i] = project.Name
	}
	msg := fmt.Sprintf("[name = %s]", names...)
	return nil, errors.Errorf("unable to find project template with name %s. Valid project names are: %s", projectName, msg)
}

func (gTemplateApi *GenesisTemplateApi) GetProjectsFromRepo() (GenesisProject, error) {
	file, err := ioutil.ReadFile(gTemplateApi.DirectoryPath + genesisFileName)

	if err != nil {
		return GenesisProject{}, errors.Wrapf(err, "Failed to read file with name %s", genesisFileName)
	}

	var projects GenesisProject
	err = yaml.Unmarshal(file, &projects)

	if err != nil {
		return GenesisProject{}, errors.Wrapf(err, "Failed to unmarshal genesis file file")
	}

	return projects, nil
}

func (gTemplateApi *GenesisTemplateApi) GenerateFromTemplate(project ProjectTemplate, variableReplacementMap map[string]string) error {
	err := project.SetValidatedOptions(variableReplacementMap)
	if err != nil {
		return err
	}

	// get the project files
	files, err := ioutil.ReadDir(gTemplateApi.DirectoryPath)
	if err != nil {
		return errors.Wrapf(err, "Failed to read project files from directory.")
	}

	root, err := project.GetRoot()
	if err != nil {
		return err
	}

	validatedOptions, err := project.GetValidatedOptions()
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() && file.Name() == root {
			directoryFuncs := make([]func() error, 0)

			err = filepath.Walk(gTemplateApi.DirectoryPath+file.Name(), func(path string, f os.FileInfo, err error) error {
				if f.IsDir() { // directory
					// create process closure with necessary parameters
					directoryFuncs = append(directoryFuncs, processDirectoryClosure(path, f, validatedOptions))
					return nil
				} else { // file
					err = processFileWithTokens(path, f, validatedOptions)
					if err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				return err
			}

			// process directory variable changes after files are handled
			for _, process := range directoryFuncs {
				err = process()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func processDirectoryClosure(path string, f os.FileInfo, options map[string]string) func() error {
	return func() error {
		oldName := f.Name()
		if strings.Contains(oldName, "{{") && strings.Contains(oldName, "}}") {
			// directory name should be replaced with one of the options passed in
			for key, value := range options {
				toFind := "{{" + key + "}}"
				if toFind == oldName {
					newPath := strings.Replace(path, oldName, value, 1)

					// handle Java packages
					if strings.Count(newPath, ".") > 0 {
						newPath = strings.ReplaceAll(newPath, ".", "/")
						err := os.MkdirAll(newPath, os.ModePerm)
						if err != nil {
							return errors.Wrapf(err, "unable to run MkdirAll on path %s", newPath)
						}
						err = copyDirectoryContents(path, newPath)
						if err != nil {
							return err
						}
						// delete old directory
						err = os.RemoveAll(path)
						if err != nil {
							return errors.Wrapf(err, "problem deleting the temp directory: %s", path)
						}
					} else {
						err := os.Rename(path, newPath)
						if err != nil {
							return errors.Wrapf(err, "unable to rename path %s to new path %s", path, newPath)
						}
					}
				} else if strings.Contains(oldName, toFind) {
					newPath := strings.Replace(path, toFind, value, 1)
					err := os.Rename(path, newPath)
					if err != nil {
						return errors.Wrapf(err, "unable to rename path %s to new path %s", path, newPath)
					}
					path = newPath
				}
			}
		}
		return nil
	}
}

func processFileRenameClosure(oldPath, newPath string) func() error {
	return func() error {
		err := os.Rename(oldPath, newPath)

		if err != nil {
			return errors.Wrapf(err, "unable to rename file path %s to new file path %s", oldPath, newPath)
		}

		return nil
	}
}

func copyDirectoryContents(oldPath, newPath string) error {
	err := filepath.Walk(oldPath, func(path string, info os.FileInfo, err error) error {
		if path == oldPath {
			return nil
		}
		remainingPath := strings.Replace(path, oldPath, "", 1)
		newRemainingPath := newPath + remainingPath
		if !info.IsDir() {
			file, err := ioutil.ReadFile(path)
			if err != nil {
				return errors.Wrapf(err, "unable to read file from path %s", path)
			}

			err = ioutil.WriteFile(newRemainingPath, file, os.ModePerm)
			if err != nil {
				return errors.Wrapf(err, "unable to write file to path %s", newRemainingPath)
			}
		} else {
			err := os.Mkdir(newRemainingPath, os.ModePerm)
			if err != nil {
				return errors.Wrapf(err, "unable to run Mkdir on path %s", newRemainingPath)
			}
		}
		return nil
	})
	return err
}

func FindTokens(document []byte, optionsMap map[string]string) ([]byte, error) {
	start := bytes.Index([]byte("{{"), document)
	end := bytes.Index([]byte("}}"), document)

	for {
		// {{key | filter}}
		token := document[start:end]
		split := bytes.Split(token, []byte("|"))
		key := bytes.TrimSpace(split[0])
		optionValue, ok := optionsMap[string(key)]
		if ok { // key is present in optionsMap
			replacement, err := ReplaceGenesisVariable(string(token), optionValue)
			if err != nil {
				return nil, err
			}
			return FindTokens(bytes.Replace(document, token, []byte(replacement), 1), optionsMap)
		}
	}

}

func RecursiveReplace(document []byte, optionsMap map[string]string) ([]byte, error) {
	start := bytes.Index(document, []byte("{{"))
	// base case - no opening tags
	if start == -1 {
		return document, nil
	}
	end := bytes.Index(document, []byte("}}"))
	// base case - no closing tags
	if end == -1 {
		return document, errors.Errorf("Malformed input: found opening tags, but not closing tags.")
	}

	// {{key | filter}}
	token := document[start : end+2]
	split := bytes.Split(token, []byte("|"))
	key := getTokenKey(split[0])
	optionValue, ok := optionsMap[string(key)]
	if ok { // key is present in optionsMap
		replacement, err := ReplaceGenesisVariable(string(token), optionValue)
		if err != nil {
			return nil, err
		}
		return RecursiveReplace(bytes.Replace(document, token, []byte(replacement), 1), optionsMap)
	} else { // key is not present, which leaves unresolved handlebars
		return nil, errors.Errorf("Malformed input: unresolved handlebars left in document for key %s", key)
	}
	// TODO - test this recursive method, and its cousin below
}

func StringRecursiveReplace(str string, optionsMap map[string]string) (string, error) {
	inputBytes := []byte(str)
	outputBytes, err := RecursiveReplace(inputBytes, optionsMap)
	if err != nil {
		return "", err
	}
	outputString := string(outputBytes)
	return outputString, nil
}

func getTokenKey(token []byte) []byte {
	stripped := bytes.ReplaceAll(token, []byte("{"), []byte(""))
	stripped = bytes.ReplaceAll(stripped, []byte("}"), []byte(""))
	stripped = bytes.ReplaceAll(stripped, []byte("|"), []byte(""))

	return bytes.TrimSpace(stripped)
}

func getTokenKeyString(token string) string {
	stripped := strings.ReplaceAll(token, "{", "")
	stripped = strings.ReplaceAll(stripped, "}", "")
	stripped = strings.ReplaceAll(stripped, "|", "")

	return strings.TrimSpace(stripped)
}

func processFileWithTokens(path string, f os.FileInfo, options map[string]string) error {

	readFile, err := ioutil.ReadFile(path)

	if err != nil {
		return errors.Wrapf(err, "unable to read file from path %s", path)
	}

	output := readFile
	var fileRenameFunc func() error
	for key, value := range options {
		toFind := "{{" + key + "}}"
		if strings.Contains(f.Name(), toFind) {
			lastIndex := strings.LastIndex(path, toFind)
			newPath := path[:lastIndex] + strings.Replace(path[lastIndex:], toFind, value, 1)
			// set up file rename for later
			fileRenameFunc = processFileRenameClosure(path, newPath)
		}
	}

	// replace all variables with filtering
	output, err = RecursiveReplace(output, options)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, output, os.ModePerm)

	if err != nil {
		return errors.Wrapf(err, "unable to write file to path %s", path)
	}

	// perform rename function after contents have been written
	if fileRenameFunc != nil {
		err := fileRenameFunc()
		if err != nil {
			return err
		}
	}
	return nil
}

func processFile(path string, f os.FileInfo, options map[string]string) error {

	readFile, err := ioutil.ReadFile(path)

	if err != nil {
		return errors.Wrapf(err, "unable to read file from path %s", path)
	}

	output := readFile
	updated := false
	var fileRenameFunc func() error
	for key, value := range options {
		toFind := "{{" + key + "}}"
		if bytes.Contains(output, []byte(toFind)) {
			output = bytes.ReplaceAll(output, []byte(toFind), []byte(value))
			updated = true
		}
		if strings.Contains(f.Name(), toFind) {
			lastIndex := strings.LastIndex(path, toFind)
			newPath := path[:lastIndex] + strings.Replace(path[lastIndex:], toFind, value, 1)
			// set up file rename for later
			fileRenameFunc = processFileRenameClosure(path, newPath)
		}
	}
	if updated {
		err = ioutil.WriteFile(path, output, os.ModePerm)

		if err != nil {
			return errors.Wrapf(err, "unable to write file to path %s", path)
		}
	}

	// perform rename function after contents have been written
	if fileRenameFunc != nil {
		err := fileRenameFunc()
		if err != nil {
			return err
		}
	}
	return nil
}

func (gTemplateApi *GenesisTemplateApi) Replace(validateOptions map[string]string) error {

	// TODO - do something with these values
	for key, value := range validateOptions {
		newVal, err := ReplaceGenesisVariable(key, value)
		if err != nil {
			return err
		}
		fmt.Printf("newVal = %s\n", newVal)
	}
	return nil
}

func (gTemplateApi *GenesisTemplateApi) ValidateGenesisProject() (bool, error) {
	file, err := ioutil.ReadFile(gTemplateApi.DirectoryPath + genesisFileName)
	if err != nil {
		return false, errors.Wrapf(err, "problem reading genesis file during validation")
	}
	var projects GenesisProject
	err = yaml.Unmarshal(file, &projects)
	if err != nil {
		return false, errors.Wrapf(err, "problem unmarshalling file into a genesis template")
	}
	return true, nil
}

func (gTemplateApi *GenesisTemplateApi) Cleanup() error {
	err := os.RemoveAll(gTemplateApi.DirectoryPath)
	if err != nil {
		return errors.Wrapf(err, "problem deleting the temp directory %s", gTemplateApi.DirectoryPath)
	}
	return nil
}
