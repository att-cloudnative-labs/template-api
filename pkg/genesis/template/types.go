package template

import (
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

var (
	ErrRootUndefined = errors.Errorf("project root is undefined")
)

type Project interface {
	// GetTemplateFileName returns the name of the file to read as a template for the project.
	// This is effectively a contract that guarantees a constant value for a particular implementation of ProjectTemplate.
	GetTemplateFileName() string

	// GetClonedDirectoryPath
	GetClonedDirectoryPath() (string, error)

	// Set the directory path where this project instance has been cloned
	SetClonedDirectoryPath(directoryPath string) error
}

type GenesisProject struct {
	Projects            []GenesisTemplate `yaml:"projects"`
	clonedDirectoryPath string
}

func (p *GenesisProject) SetClonedDirectoryPath(directoryPath string) error {
	if directoryPath == "" {
		return errors.New("directory path is empty")
	}
	p.clonedDirectoryPath = directoryPath
	return nil
}

func (p GenesisProject) GetTemplateFileName() string {
	return ".genesis.yml"
}

func (p GenesisProject) GetClonedDirectoryPath() (string, error) {
	if p.clonedDirectoryPath == "" {
		return "", errors.New("clonedDirectoryPath is empty and must be set")
	}
	return p.clonedDirectoryPath, nil
}

// ProjectTemplate interface definition
type ProjectTemplate interface {
	// Get the options that are required to be provided for this ProjectTemplate instance
	GetRequiredOptions() []Option

	// Set all required options in a key-value map
	SetValidatedOptions(args map[string]string) error

	// Get the validated options map
	GetValidatedOptions() (map[string]string, error)

	// Get the root directory name
	GetRoot() (string, error)

	// Get the name of the template
	GetName() string

	// Add formFields to groups
	OrganizeGroups() error
}

type Language struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
}

type Runtime struct {
	Names     []string  `yaml:"names"`
	GroupName string    `yaml:"groupName" json:"groupName"`
	FormField FormField `yaml:"formField,omitempty" json:"formField,omitempty"`
}

type Option struct {
	Name      string    `yaml:"name" json:"name"`
	Default   string    `yaml:"default,omitempty" json:"default,omitempty"`
	Required  bool      `yaml:"required,omitempty" json:"required,omitempty"`
	GroupName string    `yaml:"groupName" json:"groupName"`
	FormField FormField `yaml:"formField,omitempty" json:"formField,omitempty"`
}

// variable replacement
type GenesisVariable struct {
	RawKey        string
	RawValue      string
	ParsedKey     string
	ParsedValue   string
	ParsedFilters []string
	Filters       []Filter
}

func NewGenesisVariable(rawKey, rawValue string) (GenesisVariable, error) {
	if strings.HasPrefix(rawKey, "{{") && strings.HasSuffix(rawKey, "}}") {
		// strip handlebars
		noHandles := strings.ReplaceAll(rawKey, "{{", "")
		noHandles = strings.ReplaceAll(noHandles, "}}", "")
		// find filters
		splitKey := strings.Split(noHandles, "|")

		parsedKey := splitKey[0]
		filterNames := make([]string, len(splitKey)-1)
		filters := make([]Filter, len(filterNames))
		if len(filterNames) > 0 {
			// filters are present
			for i, v := range splitKey {
				if i > 0 {
					filterName := strings.TrimSpace(v)
					filterNames[i-1] = filterName
					filters[i-1] = GetFilterMap(filterName)
				}
			}
		}

		// run filters
		parsedValue := rawValue
		// pre-define error so call to filter(parsedValue) does not reassign
		var err error
		for _, filter := range filters {
			if filter != nil {
				parsedValue, err = filter(parsedValue)
				if err != nil {
					return GenesisVariable{}, err
				}
			} else {
				fmt.Printf("filter is empty.\n")
			}
		}

		return GenesisVariable{
			RawKey:        rawKey,
			RawValue:      rawValue,
			ParsedKey:     parsedKey,
			ParsedValue:   parsedValue,
			ParsedFilters: filterNames,
			Filters:       filters,
		}, nil
	} else {
		return GenesisVariable{}, errors.Errorf("rawKey: %s is invalid. Expected to be surrounded with {{rawKey}}", rawKey)
	}
}

func (gVar *GenesisVariable) GetParsedValue() (string, error) {
	if gVar.ParsedValue == "" {
		return "", errors.Errorf("Parsed Value is empty. Is GenesisVariable properly initialized? %+v", &gVar)
	}
	return gVar.ParsedValue, nil
}

func ReplaceGenesisVariable(rawKey, rawValue string) (string, error) {
	gVar, err := NewGenesisVariable(rawKey, rawValue)
	if err != nil {
		return "", err
	}
	return gVar.GetParsedValue()
}

// temporary solution for mapping filter names to filter functions
func GetFilterMap(filterKey string) Filter {
	switch strings.ToLower(filterKey) {
	case "upper":
		return UpperCaseFilter
	case "lower":
		return LowerCaseFilter
	default:
		return DefaultFilter
	}
}

type Filter func(value string) (string, error)

func DefaultFilter(value string) (string, error) {
	return value, nil
}

func UpperCaseFilter(value string) (string, error) {
	return strings.ToUpper(value), nil
}

func LowerCaseFilter(value string) (string, error) {
	return strings.ToLower(value), nil
}

// genesis front-end objects
type FormGroup struct {
	DisplayName string      `yaml:"displayName" json:"displayName"`
	FormFields  []FormField `yaml:"formFields" json:"formFields"`
	ImageGroup  bool        `yaml:"imageGroup" json:"imageGroup"`
}

type FormField struct {
	Type                   FormFieldType  `json:"type" yaml:"type"`
	Label                  string         `json:"label" yaml:"label"`
	Placeholder            string         `json:"placeholder" yaml:"placeholder"`
	FormControlName        string         `yaml:"formControlName" json:"formControlName"`
	Hint                   string         `json:"hint" yaml:"hint"`
	Icon                   string         `json:"icon" yaml:"icon"`
	IsCheckedByDefault     bool           `json:"isCheckedByDefault" yaml:"isCheckedByDefault"`
	OptionsUrl             string         `json:"optionsUrl" yaml:"optionsUrl"`
	SelectOptions          []SelectOption `json:"selectOptions" yaml:"selectOptions"`
	Validation             string         `json:"validation" yaml:"validation"`
	ValidationErrorMessage string         `json:"validationErrorMessage" yaml:"validationErrorMessage"`
	MaxCharacters          string         `json:"maxCharacters" yaml:"maxCharacters"`
	ImageButtons           []ImageButton  `yaml:"imageButtons" json:"imageButtons"`
}

type SelectOption struct {
	Value        string `yaml:"value" json:"value"`
	DisplayValue string `yaml:"displayValue" json:"displayValue"`
	IconUrl      string `yaml:"iconUrl" json:"iconUrl"`
}

type ImageButton struct {
	Type        DisplayOptionType `yaml:"type" json:"type"`
	Name        string            `yaml:"name" json:"name"`
	DisplayName string            `yaml:"displayName" json:"displayName"`
	Description string            `yaml:"description" json:"description"`
	IconUrl     string            `yaml:"iconUrl" json:"iconUrl"`
	ImageUrl    string            `yaml:"imageUrl" json:"imageUrl"`
}

// approximate an Enum
type FormFieldType string

const (
	TEXT               FormFieldType = "TEXT"
	NUMBER             FormFieldType = "NUMBER"
	EMAIL              FormFieldType = "EMAIL"
	CHECKBOX           FormFieldType = "CHECKBOX"
	SELECT             FormFieldType = "SELECT"
	SELECT_TEMPLATE    FormFieldType = "SELECT_TEMPLATE"
	AUTOCOMPLETE       FormFieldType = "AUTOCOMPLETE"
	TEXT_AREA          FormFieldType = "TEXT AREA"
	COLOR              FormFieldType = "COLOR"
	DATE               FormFieldType = "DATE"
	DATETIME_LOCAL     FormFieldType = "DATETIME_LOCAL"
	MONTH              FormFieldType = "MONTH"
	PASSWORD           FormFieldType = "PASSWORD"
	SEARCH             FormFieldType = "SEARCH"
	TEL                FormFieldType = "TEL"
	TIME               FormFieldType = "TIME"
	URL                FormFieldType = "URL"
	WEEK               FormFieldType = "WEEK"
	SECTION            FormFieldType = "SECTION"
	IMAGE_BUTTON       FormFieldType = "IMAGE_BUTTON"
	IMAGE_BUTTON_GROUP FormFieldType = "IMAGE_BUTTON_GROUP"
)

type DisplayOptionType string

const (
	LANGUAGE             DisplayOptionType = "LANGUAGE"
	FRAMEWORK            DisplayOptionType = "FRAMEWORK"
	DEPLOYMENT_TYPE      DisplayOptionType = "DEPLOYMENT_TYPE"
	CLOUD_PROVIDER       DisplayOptionType = "CLOUD_PROVIDER"
	SOURCE_CONTROL       DisplayOptionType = "SOURCE_CONTROL"
	CI_SERVER            DisplayOptionType = "CI_SERVER"
	CD_SERVER            DisplayOptionType = "CD_SERVER"
	MODULE               DisplayOptionType = "MODULE"
	DEPENDENCY           DisplayOptionType = "DEPENDENCY"
	CONTAINER_MANAGEMENT DisplayOptionType = "CONTAINER_MANAGEMENT"
)

type GenesisTemplate struct {
	Name                string               `yaml:"name" json:"name"`
	Root                string               `yaml:"root" json:"root"`
	Language            Language             `yaml:"language" json:"language"`
	Runtime             Runtime              `yaml:"runtime" json:"runtime"`
	GitRepository       GenesisGitRepository `yaml:"git,omitempty" json:"git,omitempty"`
	Options             []Option             `yaml:"options,omitempty" json:"options,omitempty"`
	FormGroups          []FormGroup          `yaml:"formGroups" json:"formGroups"`
	validatedOptionsMap map[string]string
}

type GenesisGitRepository struct {
	Domain string `yaml:"domain" json:"domain"`
	Name   string `yaml:"name" json:"name"`
}

func (p *GenesisTemplate) OrganizeGroups() error {
	if p.FormGroups == nil {
		p.FormGroups = make([]FormGroup, 0)
	}
	for i, group := range p.FormGroups {
		if group.FormFields == nil {
			p.FormGroups[i].FormFields = make([]FormField, 0)
		}
		for _, option := range p.Options {
			if strings.ToLower(option.GroupName) == strings.ToLower(group.DisplayName) {
				p.FormGroups[i].FormFields = append(p.FormGroups[i].FormFields, option.FormField)
			}
		}
		if strings.ToLower(p.Runtime.GroupName) == strings.ToLower(group.DisplayName) {
			p.FormGroups[i].FormFields = append(p.FormGroups[i].FormFields, p.Runtime.FormField)
		}
	}

	// delete any empty elements of the slice
	for idx, group := range p.FormGroups {

		for i, rcount, rlen := 0, 0, len(group.FormFields); i < rlen; i++ {
			j := i - rcount
			if group.FormFields[i].Label == "" && group.FormFields[i].Type == "" && group.FormFields[i].Hint == "" {
				// delete
				p.FormGroups[idx].FormFields = append(p.FormGroups[idx].FormFields[:j], p.FormGroups[idx].FormFields[j+1:]...)
				rcount++
			}
		}
	}

	return nil
}

func (p *GenesisTemplate) GetRoot() (string, error) {
	if p.Root != "" {
		return p.Root, nil
	}

	return "", ErrRootUndefined
}

func (p *GenesisTemplate) GetRequiredOptions() []Option {
	var required []Option

	for _, option := range p.Options {
		if option.Required {
			required = append(required, option)
		}
	}

	return required
}

func (p *GenesisTemplate) SetValidatedOptions(args map[string]string) error {
	validArgs, err := p.validateOptions(args)
	if err != nil {
		return err
	}

	p.validatedOptionsMap = validArgs
	return nil
}

func (p *GenesisTemplate) GetName() string {
	return p.Name
}

func (p *GenesisTemplate) GetValidatedOptions() (map[string]string, error) {
	return p.validatedOptionsMap, nil
}

func (p *GenesisTemplate) validateOptions(args map[string]string) (map[string]string, error) {
	validArgs := make(map[string]string, len(args))
	for _, option := range p.Options {
		var defaultVal string
		if option.Default != "" {
			defaultVal = option.Default
		}
		val, ok := args[option.Name]
		if !ok && option.Required {
			return make(map[string]string, 0), errors.Errorf("Invalid request. %s is a required parameter and was not provided.", option.Name)
		} else if ok {
			validArgs[option.Name] = val
		} else if defaultVal != "" {
			validArgs[option.Name] = defaultVal
		}
	}

	return validArgs, nil
}
