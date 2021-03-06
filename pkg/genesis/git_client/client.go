package git_client

import (
	"errors"
	"math/rand"
	"os"
	"time"
	"unsafe"
)

// common errors
var (
	ErrRepoNotExist       = errors.New("repository does not exist")
	ErrCredentialsInvalid = errors.New("provided credentials are invalid")
)

// interface and common package functions

// Interface defines behavior that a Git Client must expose
type GitClient interface {
	// RepoExists Returns true if the repository exists, and false otherwise,
	// as well as any error encountered.
	RepoExists(gitRepoConfig GitRepoConfig) (exists bool, err error)
	// GetFileFromRepo Reads a single file from a git repository. Return a pointer to the file,
	// and any errors encountered.
	GetFileFromRepo(filename string, gitRepoConfig GitRepoConfig) (file *os.File, err error)
	// CloneRepo Clones a remote repository. Return the path to that repository, and
	// any errors encountered.
	CloneRepo(gitRepoConfig GitRepoConfig) (directoryPath string, err error)
	// CheckoutBranch Checks out a particular branch. Return any errors encountered
	CheckoutBranch(branchName string, gitRepoConfig GitRepoConfig) (directoryPath string, err error)
	// CheckoutTag Checks out a particular tag. Return any errors encountered
	CheckoutTag(tagName string, gitRepoConfig GitRepoConfig) (directoryPath string, err error)
	// CreateNewRemoteRepo Creates a new git repository. Return any errors encountered.
	CreateNewRemoteRepo(gitRepoConfig GitRepoConfig) (fullRepoUrl string, err error)
	// InitRepo Initializes a repository for the given config
	InitRepo(gitRepoConfig GitRepoConfig) (directory string, err error)
	// InitialCommitProjectToRepo Commits a project to new repo with an initialize commit message.
	InitialCommitProjectToRepo(baseDirectory string, gitRepoConfig GitRepoConfig) (err error)
	// CreateScmRepoUrl Constructs a URL suitable for pushing git commits to
	CreateScmRepoUrl(config GitRepoConfig) string
	// CreateWebhook Adds a webhook
	CreateWebhook(url string, gitConfig GitRepoConfig) error
	// ListAllReposForProjectKey queries the BitBucket REST API to retrieve a list of repository names, or an error
	ListAllReposForProjectKey(projectKey string) ([]string, error)
	// AddAdminRights adds the given userID to the list of admins for a repository
	AddAdminRights(userID string, gitRepoConfig GitRepoConfig) error
}

// Defines behavior for a set of git repo configuration
type GitRepoConfig interface {
	// Returns true if the repository configuration is valid, and false otherwise
	Validate() bool
	// construct the URL used to navigate a browser to the configured repository
	ConstructRepoUrl(base string) string
	// Returns the correct URL to use for a REST API Call
	ConstructRestApiUrl(base string) string
	// Returns the provider-specific domain value
	GetRepoDomain() string
	// Returns the repository name
	GetRepoName() string
	// Set the repository name
	SetRepoName(name string)
	// Get the available repo tags
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// Utility function to generate a random hash made of lower and upper case letters
func getRandomHash(n int) string {
	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
