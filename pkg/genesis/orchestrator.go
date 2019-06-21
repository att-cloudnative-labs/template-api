package genesis

import (
	"fmt"
	"github.com/att-cloudnative-labs/template-api/genesis_config"
	"github.com/att-cloudnative-labs/template-api/pkg/genesis/git_client"
	"github.com/att-cloudnative-labs/template-api/pkg/genesis/template"
	"github.com/pkg/errors"
	"reflect"
)

const github = "github"
const bitbucket = "bitbucket"

type TemplateOrchestrator struct {
	RemoteTemplateMap map[string]git_client.GitRepoConfig
	GitClientMap      map[string]git_client.GitClient
}

type TemplateName struct {
	Name         string   `json:"name" yaml:"name"`
	ProjectNames []string `json:"projectNames" yaml:"projectNames"`
}

func NewTemplateOrchestrator(runtimeConfiguration *genesis_config.AppConfig) *TemplateOrchestrator {
	orchestrator := &TemplateOrchestrator{}
	orchestrator.RemoteTemplateMap = make(map[string]git_client.GitRepoConfig)
	orchestrator.GitClientMap = make(map[string]git_client.GitClient)
	orchestrator.initTemplates(runtimeConfiguration.BitBucketTemplateRepositories, runtimeConfiguration.GitHubTemplateRepositories)
	orchestrator.initClients(runtimeConfiguration)
	return orchestrator
}

func (templateOrchestrator *TemplateOrchestrator) initTemplates(bitBucketTemplates []genesis_config.BitBucketTemplateRepository, gitHubTemplates []genesis_config.GitHubTemplateRepository) {

	for _, bitBucketTemplate := range bitBucketTemplates {
		templateOrchestrator.RemoteTemplateMap[bitBucketTemplate.Name] = git_client.
			NewBitBucketRepoConfig(
				bitBucketTemplate.ProjectKey,
				bitBucketTemplate.RepositorySlug,
				bitBucketTemplate.FunctionalDomain,
				bitBucketTemplate.ProjectName,
			)
	}

	for _, gitHubTemplate := range gitHubTemplates {
		gitHubRepoConfig, err := git_client.NewGithubRepoConfig(gitHubTemplate.Domain, gitHubTemplate.RepoName)
		if err != nil {
			fmt.Printf("Error creating GithubRepoConfig: %+v\n", err)
		}
		templateOrchestrator.RemoteTemplateMap[gitHubTemplate.Name] = gitHubRepoConfig
	}
}

func (templateOrchestrator *TemplateOrchestrator) initClients(appConfig *genesis_config.AppConfig) {

	bitBucketClientConfig, err := git_client.NewBitBucketClientConfig(
		appConfig.BitBucketURL,
		appConfig.BitBucketUser,
		appConfig.BitBucketPassword,
		appConfig.BitBucketUserEmail,
		appConfig.BitBucketAuthToken,
		appConfig.BitBucketRestTimeout)

	if err != nil {
		fmt.Printf("Error while initializing BitBucketClientConfig: %+v\n", err)
	}

	gitHubConfig, err := git_client.NewGitClientConfig(
		"github.com",
		appConfig.GitHubUser,
		appConfig.GitHubPassword,
		appConfig.GitHubToken)

	if err != nil {
		fmt.Printf("Error while initializing GitHubClientConfig: %+v\n", err)
	}

	tempBitBucketClient := git_client.NewBitBucketClient(&bitBucketClientConfig)
	templateOrchestrator.GitClientMap[bitbucket] = &tempBitBucketClient

	tempGitHubClient := git_client.NewGitHubClient(&gitHubConfig)
	templateOrchestrator.GitClientMap[github] = &tempGitHubClient
}

// Get the names of the available Genesis Templates
func (templateOrchestrator *TemplateOrchestrator) GetTemplateNames() ([]TemplateName, error) {
	templateNames := make([]TemplateName, 0, len(templateOrchestrator.RemoteTemplateMap))
	for key := range templateOrchestrator.RemoteTemplateMap {
		// TODO - use goroutines to make async
		pt, err := templateOrchestrator.GetTemplates(key)
		if err != nil {
			return nil, err
		}
		templateName := TemplateName{Name: key, ProjectNames: make([]string, len(pt))}
		for i, p := range pt {
			templateName.ProjectNames[i] = p.GetName()
		}
		templateNames = append(templateNames, templateName)
	}
	return templateNames, nil
}

func (templateOrchestrator *TemplateOrchestrator) GetTemplates(projectName string) ([]template.ProjectTemplate, error) {
	templateRepoConfig := templateOrchestrator.RemoteTemplateMap[projectName]
	clientName, err := templateOrchestrator.getGitClient(templateRepoConfig)
	if err != nil {
		return nil, err
	}
	gitClient := templateOrchestrator.GitClientMap[clientName]

	dirName, err := gitClient.CloneRepo(templateRepoConfig)
	if err != nil {
		return nil, err
	}

	genesisTemplateApi := template.NewGenesisTemplateApi(dirName)

	genesisProject, err := genesisTemplateApi.GetProjectsFromRepo()

	if err != nil {
		return nil, err
	}
	// convert to interface to satisfy return signature
	projectTemplates := make([]template.ProjectTemplate, len(genesisProject.Projects))
	for i := range genesisProject.Projects {
		projectTemplates[i] = &genesisProject.Projects[i]
	}

	return projectTemplates, nil
}

func (templateOrchestrator *TemplateOrchestrator) GetTemplate(projectName, templateName string) (template.ProjectTemplate, error) {
	templateRepoConfig := templateOrchestrator.RemoteTemplateMap[projectName]
	clientName, err := templateOrchestrator.getGitClient(templateRepoConfig)
	if err != nil {
		return &template.GenesisTemplate{}, err
	}
	gitClient := templateOrchestrator.GitClientMap[clientName]

	dirName, err := gitClient.CloneRepo(templateRepoConfig)
	if err != nil {
		return &template.GenesisTemplate{}, err
	}

	genesisTemplateApi := template.NewGenesisTemplateApi(dirName)

	projectTemplate, err := genesisTemplateApi.GetProjectFromRepo(templateName)

	if err != nil {
		return &template.GenesisTemplate{}, err
	}

	return projectTemplate, nil
}

// Pulls a template repository, performs variable replacement, and commits new project to targetRepo
// Template and Target repositories can be from different Git Hosts (eg. Template in BitBucket and Target in GitHub)
func (templateOrchestrator *TemplateOrchestrator) GenerateFromTemplateAndCommit(templateKey, templateName, jenkinsUrl string, optionsMap map[string]string, targetRepo git_client.GitRepoConfig, createWebhook bool) (repoUrl string, err error) {

	ok := targetRepo.Validate()
	if !ok {
		return "", errors.Errorf("target repository configuration is invalid")
	}

	templateRepoConfig := templateOrchestrator.RemoteTemplateMap[templateKey]

	if templateRepoConfig == nil {
		return "", errors.Errorf("the template name [%s] is invalid", templateKey)
	}

	ok = templateRepoConfig.Validate()

	if !ok {
		return "", errors.Errorf("template repository configuration is invalid")
	}

	templateGitClientName, err := templateOrchestrator.getGitClient(templateRepoConfig)
	if err != nil {
		return "", err
	}
	templateGitClient := templateOrchestrator.GitClientMap[templateGitClientName]

	targetGitClientName, err := templateOrchestrator.getGitClient(targetRepo)
	if err != nil {
		return "", err
	}
	targetGitClient := templateOrchestrator.GitClientMap[targetGitClientName]

	dirName, err := templateGitClient.CloneRepo(templateRepoConfig)

	if err != nil {
		return "", err
	}

	genesisTemplateApi := template.NewGenesisTemplateApi(dirName)

	projectTemplate, err := genesisTemplateApi.GetProjectFromRepo(templateName)

	if err != nil {
		return "", err
	}

	err = genesisTemplateApi.GenerateFromTemplate(projectTemplate, optionsMap)

	if err != nil {
		return "", err
	}

	root, err := projectTemplate.GetRoot()

	if err != nil {
		return "", err
	}

	repoUrl, err = targetGitClient.CreateNewRemoteRepo(targetRepo)
	if err != nil {
		return "", err
	}

	err = targetGitClient.InitialCommitProjectToRepo(dirName+"/"+root, targetRepo)
	if err != nil {
		return "", err
	}

	if createWebhook {
		err = targetGitClient.CreateWebhook(jenkinsUrl, targetRepo)

		if err != nil {
			return "", err
		}
	}

	return repoUrl, nil
}

func (templateOrchestrator *TemplateOrchestrator) getGitClient(gitRepoConfig git_client.GitRepoConfig) (string, error) {
	bitBucketClz := reflect.TypeOf(git_client.BitBucketRepoConfig{}).Name()
	gitHubClz := reflect.TypeOf(git_client.GithubRepoConfig{}).Name()
	myClz := reflect.TypeOf(gitRepoConfig)

	switch myClz.Elem().Name() {
	case bitBucketClz:
		return bitbucket, nil
	case gitHubClz:
		return github, nil
	default:
		return "", errors.Errorf("git client is not supported for repo %s", gitRepoConfig.GetRepoName())
	}
}

// TODO - how to version template repositories, advertise supported versions, and expose versions to template api
