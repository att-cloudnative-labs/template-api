package git_client

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"net/http"
	"os"
	"time"
)

// REST API URL: https://api.github.com/repos/{DOMAIN}/{REPO_NAME}
// clone URL: https://github.com/{DOMAIN}/{REPO_NAME}.git
// repo URL: https://github.com/{DOMAIN}/{REPO_NAME}
type GithubRepoConfig struct {
	Domain         string
	RepositoryName string
	Tags           []string
}

func (g *GithubRepoConfig) GetRepoDomain() string {
	return g.Domain
}

func (g *GithubRepoConfig) ConstructRestApiUrl(base string) string {
	return "https://api." + base + "/repos/" + g.Domain + "/" + g.RepositoryName
}

func NewGithubRepoConfig(domain, reponame string) (*GithubRepoConfig, error) {
	if domain == "" || reponame == "" {
		return &GithubRepoConfig{}, errors.New("domain and reponame are required")
	}
	return &GithubRepoConfig{
		Domain:         domain,
		RepositoryName: reponame,
	}, nil
}

func (g *GithubRepoConfig) ConstructRepoUrl(base string) string {
	if base == "" {
		return ""
	}
	return "git://" + base + "/" + g.Domain + "/" + g.RepositoryName
}

func (g *GithubRepoConfig) GetRepoName() string {
	return g.RepositoryName
}

func (g *GithubRepoConfig) SetRepoName(name string) {
	g.RepositoryName = name
}

func (g *GithubRepoConfig) Validate() bool {
	return g.Domain != "" && g.RepositoryName != ""
}

type GitHubClientConfig struct {
	GitHost             string `yaml:"git_host" json:"git_host"`
	Username            string `yaml:"username" json:"username"`
	Password            string `yaml:"password" json:"password"`
	AccessToken         string `yaml:"access_token" json:"access_token"`
	AuthenticationToken string
}

func NewGitClientConfig(gitHost, username, password, accessToken string) (GitHubClientConfig, error) {
	if gitHost == "" {
		return GitHubClientConfig{}, errors.New("gitHost must not be empty")
	}
	if username == "" {
		return GitHubClientConfig{}, errors.New("username must not be empty")
	}
	if accessToken == "" && password == "" {
		return GitHubClientConfig{}, errors.New("either accessToken or password must not be empty")
	}
	var authenticationToken string
	if password != "" {
		authenticationToken = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	} else {
		authenticationToken = base64.StdEncoding.EncodeToString([]byte(username + ":" + accessToken))
	}

	return GitHubClientConfig{
		GitHost:             gitHost,
		Username:            username,
		Password:            password,
		AccessToken:         accessToken,
		AuthenticationToken: authenticationToken,
	}, nil
}

type GitHubClient struct {
	Config     *GitHubClientConfig
	RestClient *http.Client
}

func NewGitHubClient(config *GitHubClientConfig) GitHubClient {
	return GitHubClient{
		Config:     config,
		RestClient: &http.Client{Timeout: time.Duration(1) * time.Second},
	}
}

func (client *GitHubClient) GetFileFromRepo(filename string, gitRepoConfig GitRepoConfig) (file *os.File, err error) {
	panic("implement me")
}

func (client *GitHubClient) InitialCommitProjectToRepo(baseDirectory string, gitRepoConfig GitRepoConfig) (err error) {
	panic("implement me")
}

func (client *GitHubClient) InitRepo(gitRepoConfig GitRepoConfig) (directory string, err error) {
	panic("implement me")
}

func (client *GitHubClient) CreateScmRepoUrl(config GitRepoConfig) string {
	panic("implement me")
}

// CreateNewRemoteRepo creates a new empty remote repository with the parameters specified
// in the GitRepoConfig, and returns any errors encountered
func (client *GitHubClient) CreateNewRemoteRepo(gitRepoConfig GitRepoConfig) (fullRepoUrl string, err error) {
	url := gitRepoConfig.ConstructRepoUrl(client.Config.GitHost)
	var repositoryUrl = ""

	// check if repo exists
	exists, err := client.RepoExists(gitRepoConfig)

	if err != nil {
		return repositoryUrl, err
	}

	if exists {
		return repositoryUrl, errors.New("the repository already exists at " + url)
	}

	repoName := gitRepoConfig.GetRepoName()

	if err != nil {
		return repositoryUrl, err
	}

	// initialize simple repository
	repository, err := git.PlainInit(repoName, false)

	if err != nil {
		return repositoryUrl, errors.New("unable to init repository: " + err.Error())
	}

	remote, err := repository.CreateRemote(&config.RemoteConfig{
		Name: gitRepoConfig.GetRepoName(),
		URLs: []string{url},
	})

	if err != nil {
		return repositoryUrl, err
	}

	fmt.Printf(remote.String())

	err = repository.Push(&git.PushOptions{})

	if err != nil {
		return repositoryUrl, err
	}

	return repositoryUrl, err
}

func (client *GitHubClient) CloneRepo(gitRepoConfig GitRepoConfig) (directoryPath string, err error) {
	exists, err := client.RepoExists(gitRepoConfig)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", ErrRepoNotExist
	}
	repoUrl := gitRepoConfig.ConstructRepoUrl(client.Config.GitHost) + ".git"
	randomHash := getRandomHash(10)

	directory := "/tmp/" + randomHash + "/"

	// https://github.com/att-cloudnative-labs/go-lambda-template
	repo, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL: repoUrl,
	})

	if err != nil {
		return "", err
	}

	_, err = repo.Worktree()

	if err != nil {
		return "", errors.Wrapf(err, "something happened while accessing the worktree")
	}

	return directory, nil
}

func (client *GitHubClient) CheckoutBranch(branchName string, gitRepoConfig GitRepoConfig) (directoryPath string, err error) {
	panic("implement me")
}

func (client *GitHubClient) CheckoutTag(tagName string, gitRepoConfig GitRepoConfig) (directoryPath string, err error) {
	panic("implement me")
}

func (client *GitHubClient) RepoExists(gitRepoConfig GitRepoConfig) (exists bool, err error) {
	apiUrl := gitRepoConfig.ConstructRestApiUrl(client.Config.GitHost)
	apiProjectUrl := apiUrl

	request, _ := http.NewRequest("GET", apiProjectUrl, bytes.NewBuffer([]byte("")))
	request.Header.Set("Authorization", "Basic "+client.Config.AuthenticationToken)

	response, err := client.RestClient.Do(request)

	if err != nil {
		fmt.Printf("an error occurred: %s", err)
	}

	defer response.Body.Close()

	code := response.StatusCode

	if code == http.StatusNotFound {
		return false, nil
	}

	if code == http.StatusUnauthorized {
		return false, errors.New("github account not authorized to access repository")
	}

	if code == http.StatusOK {
		return true, nil
	}

	return true, errors.New("unable to determine if the repository exists")
}

func (client *GitHubClient) CreateWebhook(url string, gitConfig GitRepoConfig) error {
	panic("implement me")
}
