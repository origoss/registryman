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

func CreateJob(cmdParams *JobParams) {
	jobName := fmt.Sprintf("%s-%s", cmdParams.CmdType, "job")
	containerImage := "ubuntu:latest"
	entryCommand := []string{}
	entryCommand = append(entryCommand, cmdParams.CmdType, cmdParams.ProjectName, cmdParams.ConfigPath)

	clientSet := connectToK8s(cmdParams.KubeConfig)
	launchK8sJob(clientSet, jobName, containerImage, &entryCommand)
}

func connectToK8s(customPath string) *kubernetes.Clientset {
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
		log.Panicln("failed to create K8s config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panicln("Failed to create K8s clientset")
	}

	return clientset
}

func launchK8sJob(clientset *kubernetes.Clientset, jobName string, image string, cmd *[]string) {
	jobs := clientset.BatchV1().Jobs("default")
	var backOffLimit int32 = 0

	jobSpec := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: "default",
		},
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
	}

	_, err := jobs.Create(context.TODO(), jobSpec, metav1.CreateOptions{})
	if err != nil {
		log.Fatalln("Failed to create K8s job.")
	}

	//print job details
	log.Println("Created K8s job successfully")
}
