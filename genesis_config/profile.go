package genesis_config

import (
	"fmt"
	"github.com/spf13/viper"
)

var AuthConfig *AppConfig

type AppConfig struct {
	BitBucketURL         string `mapstructure:"bitbucket_url"`
	BitBucketUser        string `mapstructure:"bitbucket_user"`
	BitBucketPassword    string `mapstructure:"bitbucket_password"`
	BitBucketAuthToken   string `mapstructure:"bitbucket_token"`
	BitBucketRestTimeout int    `mapstructure:"bitbucket_timeout"`
	BitBucketUserEmail   string `mapstructure:"bitbucket_user_email"`
	GitHubUser           string `mapstructure:"github_user"`
	GitHubPassword       string `mapstructure:"github_password"`
	GitHubToken          string `mapstructure:"github_token"`
	// TODO - reconfigure to enable override with environment variables
	GitHubTemplateRepositories    []GitHubTemplateRepository    `mapstructure:"github_template_repositories"`
	BitBucketTemplateRepositories []BitBucketTemplateRepository `mapstructure:"bitbucket_template_repositories"`
	Port                          string                        `mapstructure:"port"`
}

type GitHubTemplateRepository struct {
	Name     string `mapstructure:"name"`
	Domain   string `mapstructure:"domain"`
	RepoName string `mapstructure:"repo_name"`
}

type BitBucketTemplateRepository struct {
	Name             string `mapstructure:"name"`
	ProjectKey       string `mapstructure:"project_key"`
	RepositorySlug   string `mapstructure:"repository_slug"`
	FunctionalDomain string `mapstructure:"functional_domain"`
	ProjectName      string `mapstructure:"project_name"`
}

func InitConfig() {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./")
	v.AddConfigPath("./config")
	v.AddConfigPath("./genesis_config")

	v.AutomaticEnv()

	err := v.ReadInConfig()
	if err != nil {
		fmt.Printf("Error reading in config. %+v\n", err)
	}

	v.SetDefault("bitbucket_timeout", 3)

	err = v.Unmarshal(&AuthConfig)
	if err != nil {
		fmt.Printf("Error unmarshalling config. %+v\n", err)
	}
}
