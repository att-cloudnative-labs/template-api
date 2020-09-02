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

package cmd

import (
	"fmt"
	"github.com/att-cloudnative-labs/template-api/genesis_config"
	"github.com/att-cloudnative-labs/template-api/pkg/genesis"
	"github.com/att-cloudnative-labs/template-api/pkg/genesis/git_client"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile                     string
	optionsMap                  map[string]string
	targetRepoProjectKey        string
	targetRepoSlug              string
	targetRepoFunctionalDomain  string
	targetRepoProjectName       string
	templateProjectName         string
	templateProjectTemplateName string
	templateRepoJenkinsUrl      string
	userID                      string
	templateRepoCreateWebhook   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "genesis-api",
	Short: "Generate projects from self-describing templates.",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: Orchestrator,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Orchestrator(cmd *cobra.Command, args []string) {
	orchestrator := genesis.NewTemplateOrchestrator(genesis_config.AuthConfig)

	targetRepo := git_client.NewBitBucketRepoConfig(targetRepoProjectKey, targetRepoSlug, targetRepoFunctionalDomain, targetRepoProjectName)

	repoUrl, err := orchestrator.GenerateFromTemplateAndCommit(userID, templateProjectName, templateProjectTemplateName, templateRepoJenkinsUrl, optionsMap, targetRepo, templateRepoCreateWebhook)

	if err != nil {
		fmt.Printf("error occurred: %+v\n", err)
		return
	}

	fmt.Printf("Repo URL: %s", repoUrl)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Flags().StringToStringVar(&optionsMap, "options", map[string]string{}, "Pass in options to create your project.")

	rootCmd.Flags().StringVar(&targetRepoProjectKey, "targetProjectKey", "", "Project key for target repository")
	rootCmd.Flags().StringVar(&targetRepoSlug, "targetRepoSlug", "", "Project slug for target repository")
	rootCmd.Flags().StringVar(&targetRepoFunctionalDomain, "targetRepoFunctionalDomain", "", "Functional Domain for target repository")
	rootCmd.Flags().StringVar(&targetRepoProjectName, "targetRepoProjectName", "", "Then name of the target project for target repository")

	rootCmd.Flags().StringVar(&templateProjectName, "templateProjectName", "", "The name of the Template Project")
	rootCmd.Flags().StringVar(&templateProjectTemplateName, "templateName", "", "The name of the Template to generate")
	rootCmd.Flags().StringVar(&templateRepoJenkinsUrl, "templateRepoJenkinsUrl", "", "The Jenkins URL for webhook configuration")
	rootCmd.Flags().BoolVar(&templateRepoCreateWebhook, "templateRepoCreateWebhook", false, "Flag to generate webhook or not")
	rootCmd.Flags().StringVar(&userID, "userID", "", "The user ID of the person creating a project")
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.genesis-api.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	genesis_config.InitConfig()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".genesis-api" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".genesis-api")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
