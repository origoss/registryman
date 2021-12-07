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

package skopeo

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	clientcmd "k8s.io/client-go/tools/clientcmd"
)

type Command struct {
	CmdType     string
	ProjectName string
	ConfigPath  string
}

type JobParams struct {
	*Command
	KubeConfig string
}

// Customizable cron-job
func CreateJob(cmdParams *JobParams) error {
	jobName := fmt.Sprintf("%s-%s", cmdParams.CmdType, "job")
	// skopeo image
	containerImage := "ubuntu:latest"
	entryCommand := []string{}
	entryCommand = append(entryCommand, cmdParams.CmdType, cmdParams.ProjectName, cmdParams.ConfigPath)

	clientSet, err := connectToK8s(cmdParams.KubeConfig)
	if err != nil {
		return err
	}

	err = launchK8sJob(clientSet, jobName, containerImage, &entryCommand)
	if err != nil {
		return err
	}

	return nil
}

func connectToK8s(customPath string) (*kubernetes.Clientset, error) {
	home, exists := os.LookupEnv("HOME")
	if !exists {
		home = "/root"
	}

	base := "config"
	if customPath != "" {
		base = filepath.Base(customPath)
	}

	configPath := filepath.Join(home, ".kube", base)

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create K8s config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create K8s clientset")
	}

	return clientset, nil
}

func launchK8sJob(clientset *kubernetes.Clientset, jobName string, image string, cmd *[]string) error {
	//jobs := clientset.BatchV1().Jobs("default")
	jobs := clientset.BatchV1().CronJobs("default")
	var backOffLimit int32 = 0

	cronJobSpec := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
		Spec: batchv1.CronJobSpec{
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:    jobName,
									Image:   image,
									Command: *cmd,
								},
							},
							RestartPolicy: v1.RestartPolicyNever,
						},
					},
					BackoffLimit: &backOffLimit,
				},
			},
			Schedule:          "0 * * * *",
			ConcurrencyPolicy: batchv1.ForbidConcurrent,
		},
	}

	_, err := jobs.Create(context.TODO(), cronJobSpec, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create K8s cron-job")
	}

	//print job details
	log.Println("created K8s cron-job successfully")
	return nil
}
