package cronjob

import (
	"context"
	"fmt"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CronJob struct {
	projectName    string
	remoteRegistry *globalregistry.Registry
	dir            string
	spec           *batchv1.CronJob
}

var _ globalregistry.ReplicationRule = &CronJob{}

func new(projectName, nameSpace, repo string, cmd *[]string, remoteRegistry *globalregistry.Registry) *CronJob {
	var backOffLimit int32 = 0
	cronJobUniqueName := fmt.Sprintf("%s-%s-job", projectName, repo)

	cronJob := &CronJob{
		projectName:    projectName,
		remoteRegistry: remoteRegistry,
		spec: &batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cronJobUniqueName,
				Namespace: nameSpace,
			},
			Spec: batchv1.CronJobSpec{
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Name:    cronJobUniqueName,
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
		},
	}

	return cronJob
}

func (cj *CronJob) Deploy() error {
	cronJobInterface := clientSet.BatchV1().CronJobs(cj.spec.Namespace)
	_, err := cronJobInterface.Create(context.TODO(), cj.spec, metav1.CreateOptions{})
	return err
}

func (cj *CronJob) Delete() error {
	cronJobInterface := clientSet.BatchV1().CronJobs(cj.spec.Namespace)
	return cronJobInterface.Delete(context.TODO(), cj.spec.Name, metav1.DeleteOptions{})
}

func (cj *CronJob) Direction() string {
	return cj.dir
}

func (cj *CronJob) GetName() string {
	return cj.spec.Name
}

func (cj *CronJob) GetNamespace() string {
	return cj.spec.Namespace
}

func (cj *CronJob) GetProjectType() string {
	return ""
}

func (cj *CronJob) GetProjectName() string {
	return cj.projectName
}

func (cj *CronJob) RemoteRegistry() globalregistry.Registry {
	return *cj.remoteRegistry
}

func (cj *CronJob) Trigger() string {
	return ""
}
