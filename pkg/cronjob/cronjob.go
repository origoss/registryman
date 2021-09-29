package cronjob

import (
	"context"
	"fmt"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	batchv1 "k8s.io/api/batch/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CronJob struct {
	remoteRegistry globalregistry.Registry
	dir            string
	resource       *v1beta1.CronJob
}

var _ globalregistry.ReplicationRule = &CronJob{}

func new(cronJobFactory *CronJobFactory, repositories string, remoteRegistry globalregistry.Registry, args []string) *CronJob {
	var backOffLimit int32 = 0
	//volumeName := "skopeo-script"
	cronJobUniqueName := fmt.Sprintf("%s-job", cronJobFactory.project.GetName())

	envVar := v1.EnvVar{
		Name:  "REPOSITORIES",
		Value: repositories,
	}

	cronJob := &CronJob{
		remoteRegistry: remoteRegistry,
		resource: &v1beta1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cronJobUniqueName,
				Namespace: cronJobFactory.source.GetName(),
				Labels: map[string]string{
					"source":          "registryman-skopeo",
					"project":         cronJobFactory.project.GetName(),
					"remote-registry": remoteRegistry.GetName(),
				},
			},
			Spec: v1beta1.CronJobSpec{
				JobTemplate: v1beta1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Template: v1.PodTemplateSpec{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Name:            cronJobUniqueName,
										Image:           image,
										Command:         []string{"/bin/bash"},
										Args:            args,
										ImagePullPolicy: v1.PullAlways,
										Env:             append([]v1.EnvVar{}, envVar),
									},
								},
								RestartPolicy: v1.RestartPolicyNever,
							},
						},
						BackoffLimit: &backOffLimit,
					},
				},
				Schedule:          "*/1 * * * *",
				ConcurrencyPolicy: v1beta1.ForbidConcurrent,
			},
		},
	}

	return cronJob
	// TODO: Go converter package for dynamic cronjob version generation
}

// TODO: cj in registry's namespace
func (cj *CronJob) Deploy(ctx context.Context) error {
	_, err := clientSet.BatchV1beta1().CronJobs(cj.resource.Namespace).Create(ctx, cj.resource, metav1.CreateOptions{})
	return err
}

func (cj *CronJob) Delete(ctx context.Context) error {
	return clientSet.BatchV1beta1().CronJobs(cj.resource.Namespace).Delete(ctx, cj.resource.Name, metav1.DeleteOptions{})
}

func (cj *CronJob) Direction() string {
	return cj.dir
}

func (cj *CronJob) GetName() string {
	return cj.resource.Name
}

func (cj *CronJob) GetNamespace() string {
	return cj.resource.Namespace
}

func (cj *CronJob) GetProjectType() string {
	return ""
}

func (cj *CronJob) GetProjectName() string {
	return cj.resource.Labels["project"]
}

func (cj *CronJob) RemoteRegistry() globalregistry.Registry {
	return cj.remoteRegistry
}

func (cj *CronJob) Trigger() string {
	return ""
}
