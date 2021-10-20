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
var _ globalregistry.DestructibleReplicationRule = &CronJob{}

func create(labels map[string]string, schedule, configMapName, direction string, remoteRegistry globalregistry.Registry, args []string) *CronJob {
	var backOffLimit int32 = 1
	cronJobUniqueName := fmt.Sprintf("%s-job", labels["project"])
	startingDeadlineSecPtr := new(int64)
	*startingDeadlineSecPtr = 200

	cronJob := &CronJob{
		remoteRegistry: remoteRegistry,
		dir:            direction,
		resource: &v1beta1.CronJob{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CronJob",
				APIVersion: "v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:   cronJobUniqueName,
				Labels: labels,
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
										EnvFrom: []v1.EnvFromSource{
											{
												ConfigMapRef: &v1.ConfigMapEnvSource{
													LocalObjectReference: v1.LocalObjectReference{
														Name: configMapName,
													},
												},
											},
										},
									},
								},
								RestartPolicy: v1.RestartPolicyNever,
							},
						},
						BackoffLimit: &backOffLimit,
					},
				},
				Schedule:                schedule,
				ConcurrencyPolicy:       v1beta1.ForbidConcurrent,
				StartingDeadlineSeconds: startingDeadlineSecPtr,
			},
		},
	}

	return cronJob
	// TODO: Go converter package for dynamic cronjob version generation
}

func createConfigMapForEnvvar(labels, data map[string]string) *v1.ConfigMap {
	configMapUniqueName := fmt.Sprintf("%s-cm", labels["project"])
	configMap := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   configMapUniqueName,
			Labels: labels,
		},
		Data: data,
	}
	return configMap
}

func (cj *CronJob) Resource() *v1beta1.CronJob {
	return cj.resource
}

func (cj *CronJob) Direction() string {
	return cj.dir
}

func (cj *CronJob) GetName() string {
	return cj.resource.Name
}

func (cj *CronJob) GetProjectName() string {
	return cj.resource.Labels["project"]
}

func (cj *CronJob) RemoteRegistry() globalregistry.Registry {
	return cj.remoteRegistry
}

func (cj *CronJob) Trigger() string {
	return fmt.Sprintf("%s %s", "cron", cj.resource.Spec.Schedule)
}

func (cj *CronJob) Type() globalregistry.ReplicationType {
	return globalregistry.SkopeoReplication
}

func (cj *CronJob) Delete(ctx context.Context) error {
	manifestManipulator, err := createManifestManipulator(ctx)
	if err != nil {
		return err
	}
	return manifestManipulator.RemoveResource(ctx, cj.resource)
}
