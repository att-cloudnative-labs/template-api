package git_client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	gitHttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type BitBucketRepoRequest struct {
	Name     string `json:"name"`
	ScmId    string `json:"scmId"`
	Forkable bool   `json:"forkable"`
}

type BitBucketWebHookRequest struct {
	Title              string `json:"title"`
	URL                string `json:"url"`
	CommittersToIgnore string `json:"committersToIgnore"`
	BranchesToIgnore   string `json:"branchesToIgnore"`
	Enabled            bool   `json:"enabled"`
}

// REST API URL:
// clone URL:
// repo URL:
// BitBucket implementation of the GitClient interfaces
type BitBucketRepoConfig struct {
	ProjectKey       string
	RepositorySlug   string
	FunctionalDomain string
	ProjectName      string
	Tags             []string
}

func (config *BitBucketRepoConfig) GetRepoDomain() string {
	return config.ProjectKey
}

func (config *BitBucketRepoConfig) Validate() bool {
	return config.ProjectKey != "" && config.RepositorySlug != "" && config.FunctionalDomain != "" && config.ProjectName != ""
}

func NewBitBucketRepoConfig(projectKey, repositorySlug, functionalDomain, projectName string, tags ...string) *BitBucketRepoConfig {
	return &BitBucketRepoConfig{
		ProjectKey:       projectKey,
		RepositorySlug:   repositorySlug,
		FunctionalDomain: functionalDomain,
		ProjectName:      projectName,
		Tags:             tags,
	}
}

//https://{user}@{base}/rest/api/1.0/projects/{PROJECT_KEY}/repos
func (config *BitBucketRepoConfig) ConstructRestApiUrl(base string) string {
	return "https://" + base + "/rest/api/1.0" + "/projects/" + config.ProjectKey + "/repos"
}

//https://{base}/projects/{PROJECT_KEY}/repos
func (config *BitBucketRepoConfig) ConstructRepoUrl(base string) string {
	return "https://" + base + "/projects/" + config.ProjectKey + "/repos"
}

func (config *BitBucketRepoConfig) GetRepoName() string {
	return config.RepositorySlug
}

func (config *BitBucketRepoConfig) SetRepoName(name string) {
	config.RepositorySlug = name
}

type BitBucketClientConfig struct {
	GitHost             string `yaml:"git_host" json:"git_host"`
	Username            string `yaml:"username" json:"username"`
	Password            string `yaml:"password" json:"password"`
	AccessToken         string `yaml:"access_token" json:"access_token"`
	Email               string `yaml:"email" json:"email"`
	RestTimeout         int    `yaml:"timeout" json:"timeout"`
	AuthenticationToken string
}

func NewBitBucketClientConfig(gitHost, username, password, email, accessToken string, timeout int) (BitBucketClientConfig, error) {
	if gitHost == "" {
		return BitBucketClientConfig{}, errors.Errorf("gitHost must not be empty")
	}
	if username == "" {
		return BitBucketClientConfig{}, errors.Errorf("username must not be empty")
	}
	if email == "" {
		return BitBucketClientConfig{}, errors.Errorf("email must not be empty")
	}
	if accessToken == "" && password == "" {
		return BitBucketClientConfig{}, errors.Errorf("either accessToken or password must not be empty")
	}
	var authenticationToken string
	if password != "" {
		authenticationToken = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	} else {
		authenticationToken = base64.StdEncoding.EncodeToString([]byte(username + ":" + accessToken))
	}

	if timeout < 0 || timeout > 10 {
		return BitBucketClientConfig{}, errors.Errorf("timeout must be set between 1 and 10")
	}

	return BitBucketClientConfig{
		GitHost:             gitHost,
		Username:            username,
		Password:            password,
		AccessToken:         accessToken,
		Email:               email,
		RestTimeout:         timeout,
		AuthenticationToken: authenticationToken,
	}, nil
}

type BitBucketClient struct {
	Config     *BitBucketClientConfig
	RestClient *http.Client
}

// https://{host}/rest/api/1.0/projects/{projectKey}/repos?limit=1000
func (gitClient *BitBucketClient) ListAllReposForProjectKey(projectKey string) ([]string, error) {
	url := "https://" + gitClient.Config.GitHost + "/rest/api/1.0/projects/" + projectKey + "/repos?limit=1000"

	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Authorization", "Basic "+gitClient.Config.AuthenticationToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := gitClient.RestClient.Do(request)

	if response != nil {
		defer func() {
			if err := response.Body.Close(); err != nil {
				fmt.Printf("error closing response body %+v\n", err)
			}
		}()
	}

	if err != nil {
		return nil, errors.Wrapf(err, "Problem performing request to BitBucket REST API for retrieving repos in project: %s", projectKey)
	}
	var bitBucketResponse BitBucketRepoResponse

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "Problem reading result body")
	}

	err = json.Unmarshal(body, &bitBucketResponse)
	if err != nil {
		return nil, errors.Wrapf(err, "Problem unmarshalling body")
	}

	return bitBucketResponse.GetRepositoryNames(), nil
}

func (gitClient *BitBucketClient) CreateWebhook(url string, gitConfig GitRepoConfig) error {
	webhookUrl := "https://" + gitClient.Config.GitHost + "/rest/webhook/1.0/projects/" + gitConfig.GetRepoDomain() + "/repos/" + gitConfig.GetRepoName() + "/configurations"
	payload := BitBucketWebHookRequest{
		Title:              "Jenkins Webhook",
		URL:                url,
		CommittersToIgnore: "",
		BranchesToIgnore:   "",
		Enabled:            true,
	}
	jsonValue, err := json.Marshal(payload)

	if err != nil {
		return errors.Wrapf(err, "problem marshaling webhook payload")
	}

	request, _ := http.NewRequest("PUT", webhookUrl, bytes.NewBuffer([]byte(jsonValue)))
	request.Header.Set("Authorization", "Basic "+gitClient.Config.AuthenticationToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := gitClient.RestClient.Do(request)

	if response != nil {
		defer func() {
			if err := response.Body.Close(); err != nil {
				fmt.Printf("error closing response body %+v\n", err)
			}
		}()
	}

	if err != nil {
		return errors.Wrapf(err, "something happened while creating webhook")
	}

	return nil
}

func NewBitBucketClient(config *BitBucketClientConfig) BitBucketClient {
	return BitBucketClient{
		Config:     config,
		RestClient: &http.Client{Timeout: time.Duration(config.RestTimeout) * time.Second},
	}
}

//https://{user}@egbitbucket.dtvops.net/scm/{domain}/{repo_name}.git
func (gitClient *BitBucketClient) CreateScmRepoUrl(gitConfig GitRepoConfig) string {
	return "https://" + gitClient.Config.Username + "@" + gitClient.Config.GitHost + "/scm/" + gitConfig.GetRepoDomain() + "/" + gitConfig.GetRepoName() + ".git"
}

func (gitClient *BitBucketClient) CreateNewRemoteRepo(gitRepoConfig GitRepoConfig) (fullRepoUrl string, err error) {
	apiUrl := gitRepoConfig.ConstructRestApiUrl(gitClient.Config.GitHost)
	var repositoryUrl = ""

	// check if repo exists
	exists, err := gitClient.RepoExists(gitRepoConfig)

	if err != nil {
		return repositoryUrl, err
	}

	// create repo
	repoUrl := gitRepoConfig.ConstructRepoUrl(gitClient.Config.GitHost) + "/" + gitRepoConfig.GetRepoName()

	// don't attempt to create repo if it exists
	if exists {
		return repositoryUrl, errors.Errorf("the repository already exists at %s", repoUrl)
	}

	newRepoObject := BitBucketRepoRequest{
		Name:     gitRepoConfig.GetRepoName(),
		ScmId:    "git",
		Forkable: true,
	}
	jsonValue, _ := json.Marshal(newRepoObject)

	request, _ := http.NewRequest("POST", apiUrl, bytes.NewBuffer([]byte(jsonValue)))
	request.Header.Set("Authorization", "Basic "+gitClient.Config.AuthenticationToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := gitClient.RestClient.Do(request)

	if response != nil {
		defer func() {
			if err := response.Body.Close(); err != nil {
				fmt.Printf("error closing response body %+v\n", err)
			}
		}()
	}

	if err != nil {
		return "", errors.Wrapf(err, "something happened while creating repository %s", gitRepoConfig.GetRepoName())
	}

	return repoUrl, nil
}

func (gitClient *BitBucketClient) InitRepo(gitRepoConfig GitRepoConfig) (directory string, err error) {
	randomHash := getRandomHash(10)

	directory = "/tmp/" + randomHash + "/"

	// if directory is not empty, run git init
	_, err = git.PlainInit(directory, false)
	if err != nil {
		return "", errors.Wrapf(err, "something happened while running `git init` for project %s", gitRepoConfig.GetRepoName())
	}

	return directory, nil
}

func (gitClient *BitBucketClient) InitialCommitProjectToRepo(baseDirectory string, gitRepoConfig GitRepoConfig) error {
	// check if directory exists
	if _, err := os.Stat(baseDirectory); os.IsNotExist(err) {
		// baseDirectory does not exist
		return errors.Wrapf(err, "Base directory %s does not exist", baseDirectory)
	}

	// if directory is not empty, run git init
	// git init
	repository, err := git.PlainInit(baseDirectory, false)
	if err != nil {
		return errors.Wrapf(err, "something happened while running `git init` in %s", baseDirectory)
	}

	// add new remote
	// git remote add {Name} {URL}
	remote, err := repository.CreateRemote(&config.RemoteConfig{
		Name: gitRepoConfig.GetRepoName(),
		URLs: []string{gitClient.CreateScmRepoUrl(gitRepoConfig)},
	})

	if err != nil {
		return errors.Wrapf(err, "something happened while creating remote repository %s", gitRepoConfig.GetRepoName())
	}

	// get the worktree from the repository
	worktree, err := repository.Worktree()
	if err != nil {
		return errors.Wrapf(err, "something happened while getting the worktree from the repository object")
	}

	var commitMessage = "Initial Commit by Genesis API"

	// recursively add files to commit
	// git add
	err = addFilesToGit(baseDirectory, worktree)

	if err != nil {
		return err
	}

	// git commit
	commit, err := worktree.Commit(commitMessage, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Genesis API",
			Email: gitClient.Config.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return errors.Wrapf(err, "something happened while running `git commit` for project")
	}
	_, err = repository.CommitObject(commit)
	if err != nil {
		return errors.Wrapf(err, "something happened while running `git commit` for project")
	}

	// git push
	err = repository.Push(&git.PushOptions{
		RemoteName: remote.Config().Name,
		Auth:       &gitHttp.BasicAuth{Username: gitClient.Config.Username, Password: gitClient.Config.Password},
	})
	if err != nil {
		return errors.Wrapf(err, "something happened while running `git push` for project %s", gitRepoConfig.GetRepoName())
	}

	return nil
}

func (gitClient *BitBucketClient) RepoExists(gitRepoConfig GitRepoConfig) (exists bool, err error) {
	apiUrl := gitRepoConfig.ConstructRestApiUrl(gitClient.Config.GitHost)
	apiProjectUrl := apiUrl + "/" + gitRepoConfig.GetRepoName()

	request, _ := http.NewRequest("GET", apiProjectUrl, bytes.NewBuffer([]byte("")))
	request.Header.Set("Authorization", "Basic "+gitClient.Config.AuthenticationToken)

	response, err := gitClient.RestClient.Do(request)

	if response != nil {
		defer func() {
			if err := response.Body.Close(); err != nil {
				fmt.Printf("error closing response body %+v\n", err)
			}
		}()
	}

	if err != nil {
		return false, errors.Wrapf(err, "unable to determine if the repository exists")
	}

	code := response.StatusCode

	if code == http.StatusOK {
		return true, nil
	}

	if code == http.StatusNotFound {
		return false, nil
	}

	if code == http.StatusUnauthorized {
		return false, errors.Errorf("configured bitbucket account not authorized to access repository")
	}

	return true, errors.Wrapf(err, "unable to determine if the repository exists")
}

func (gitClient *BitBucketClient) GetFileFromRepo(filename string, gitRepoConfig GitRepoConfig) (file *os.File, err error) {
	panic("implement me")
}

func (gitClient *BitBucketClient) CloneRepo(gitRepoConfig GitRepoConfig) (directoryPath string, err error) {

	directory, repoUrl, err := getTempDirectoryAndRepoUrl(gitClient, gitRepoConfig)

	if err != nil {
		return "", err
	}

	repo, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:  repoUrl,
		Auth: &gitHttp.BasicAuth{Username: gitClient.Config.Username, Password: gitClient.Config.Password},
	})

	err = verifyRepo(repo, err)

	if err != nil {
		return "", err
	}

	return directory, nil
}

func (gitClient *BitBucketClient) CheckoutBranch(branchName string, gitRepoConfig GitRepoConfig) (directoryPath string, err error) {

	directory, repoUrl, err := getTempDirectoryAndRepoUrl(gitClient, gitRepoConfig)

	if err != nil {
		return "", err
	}

	repo, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:           repoUrl,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)),
		SingleBranch:  true,
		Auth:          &gitHttp.BasicAuth{Username: gitClient.Config.Username, Password: gitClient.Config.Password},
	})

	err = verifyRepo(repo, err)

	if err != nil {
		return "", err
	}

	return directory, nil
}

func (gitClient *BitBucketClient) CheckoutTag(tagName string, gitRepoConfig GitRepoConfig) (directoryPath string, err error) {
	directory, repoUrl, err := getTempDirectoryAndRepoUrl(gitClient, gitRepoConfig)

	if err != nil {
		return "", err
	}

	repo, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:           repoUrl,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/tags/%s", tagName)),
		SingleBranch:  true,
		Auth:          &gitHttp.BasicAuth{Username: gitClient.Config.Username, Password: gitClient.Config.Password},
	})

	err = verifyRepo(repo, err)

	if err != nil {
		return "", err
	}

	return directory, nil
}

func getTempDirectoryAndRepoUrl(gitClient *BitBucketClient, gitRepoConfig GitRepoConfig) (directoryPath, repoUrl string, err error) {
	exists, err := gitClient.RepoExists(gitRepoConfig)
	if err != nil {
		return "", "", err
	}

	if !exists {
		return "", "", ErrRepoNotExist
	}
	repoUrl = gitClient.CreateScmRepoUrl(gitRepoConfig)
	randomHash := getRandomHash(10)

	directoryPath = "/tmp/" + randomHash + "/"

	return directoryPath, repoUrl, nil
}

func verifyRepo(repo *git.Repository, err error) error {

	if err != nil {
		return errors.Wrapf(err, "something happened while cloning repo")
	}

	_, err = repo.Worktree()

	if err != nil {
		return errors.Wrapf(err, "something happened while accessing the worktree")
	}

	return nil
}

func addFilesToGit(directoryPath string, worktree *git.Worktree) error {
	files, err := ioutil.ReadDir(directoryPath)

	if err != nil {
		return errors.Wrapf(err, "something happened while reading the directory path %s", directoryPath)
	}

	for _, file := range files {
		err = filepath.Walk(directoryPath+"/"+file.Name(), func(path string, f os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if f.Name() != ".git" {
				// do work
				_, err = worktree.Add(f.Name())
				if err != nil {
					return errors.Wrapf(err, "something happened while adding file %s to worktree", f.Name())
				}
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error encountered but not fatal. Error: %s\n", err)
		}
	}
	return nil
}
