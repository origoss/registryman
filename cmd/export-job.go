/*
   Copyright 2021 The Kubermatic Kubernetes Platform contributors.

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
	"context"
	"fmt"
	"time"

	"github.com/kubermatic-labs/registryman/pkg/skopeo"
	"github.com/spf13/cobra"
)

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("job called")
		kubeConfig := args[0]

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		jobParams := &skopeo.JobParams{
			Command: &skopeo.Command{
				CmdType:     "export",
				ProjectName: "os-images",
				ConfigPath:  "registryman-test/testdata/state5",
			},
			KubeConfig: kubeConfig,
		}

		return skopeo.CreateJob(ctx, jobParams)
	},
}

func init() {
	rootCmd.AddCommand(jobCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// jobCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// jobCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
