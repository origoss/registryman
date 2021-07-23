/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/skopeo"
	"github.com/spf13/cobra"
)

var sourcePath string

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Uploads a repository from a local directory to a registry",
	Long: `The import command takes two arguments, the path to the 
local directory that contains the repository in .tar format, and also
the URL of the registry, where the repository will be pushed.
	`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("import called")
		projectName := args[0]
		configDir := args[1]

		config.SetLogger(logger)

		manifests, err := config.ReadManifests(configDir, nil)
		if err != nil {
			return err
		}

		project, err := manifests.GetProjectByName(projectName)
		if err != nil {
			return err
		}

		projectDestinationFullPath, err := project.GenerateProjectRepoName()
		if err != nil {
			return err
		}

		transfer, err := skopeo.New(project.Registry.GetUsername(), project.Registry.GetPassword())
		if err != nil {
			return err
		}

		sourceDirectoryPath := fmt.Sprintf("%s/%s", sourcePath, projectDestinationFullPath)

		if err := transfer.Import(sourceDirectoryPath, projectDestinationFullPath, logger); err != nil {
			return err
		}
		logger.Info("importing project finished", "project name", projectName)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	exportCmd.PersistentFlags().StringVarP(&sourcePath, "path", "f", "./exported-repositories", "The path for the saved repositories")
}
