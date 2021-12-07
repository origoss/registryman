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

package config

import (
	"context"
	"fmt"

	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	batchv1 "k8s.io/api/batch/v1"
	v1beta1 "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type replicationTrigger struct {
	triggerType     api.ReplicationTriggerType
	triggerSchedule string
}

func (rt replicationTrigger) TriggerType() api.ReplicationTriggerType {
	return rt.triggerType
}

func (rt replicationTrigger) TriggerSchedule() string {
	return rt.triggerSchedule
}

type cronJob struct {
	remoteRegistry globalregistry.Registry
	dir            string
	resource       *v1beta1.CronJob
	replTrigger    *replicationTrigger
}

var _ globalregistry.ReplicationRule = &cronJob{}
var _ globalregistry.DestructibleReplicationRule = &cronJob{}
var _ globalregistry.ReplicationTrigger = replicationTrigger{}

func create(labels map[string]string, configMapName, direction string, remoteRegistry globalregistry.Registry, args []string, trigger globalregistry.ReplicationTrigger) *cronJob {
	var backOffLimit int32 = 1
	cronJobUniqueName := fmt.Sprintf("%s-job", labels["project"])
	startingDeadlineSecPtr := new(int64)
	*startingDeadlineSecPtr = 200

	cronJobConfig := &cronJob{
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
				Schedule:                trigger.TriggerSchedule(),
				ConcurrencyPolicy:       v1beta1.ForbidConcurrent,
				StartingDeadlineSeconds: startingDeadlineSecPtr,
			},
		},
	}

	return cronJobConfig
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

func (cj *cronJob) Resource() *v1beta1.CronJob {
	return cj.resource
}

func (cj *cronJob) Direction() string {
	return cj.dir
}

func (cj *cronJob) GetName() string {
	return cj.resource.Name
}

func (cj *cronJob) GetProjectName() string {
	return cj.resource.Labels["project"]
}

func (cj *cronJob) RemoteRegistry() globalregistry.Registry {
	return cj.remoteRegistry
}

func (cj *cronJob) Trigger() globalregistry.ReplicationTrigger {
	return cj.replTrigger
}

func (cj *cronJob) Type() globalregistry.ReplicationType {
	return globalregistry.SkopeoReplication
}

func (cj *cronJob) Delete(ctx context.Context) error {
	manifestManipulator, err := createManifestManipulator(ctx)
	if err != nil {
		return err
	}
	return manifestManipulator.RemoveResource(ctx, cj.resource)
}
