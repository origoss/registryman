package cronjob

import (
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	batchv1 "k8s.io/api/batch/v1"
)

type CronJob struct {
	projectName string
	Dir         string
	spec        *batchv1.CronJob
}

var _ globalregistry.ReplicationRule = &CronJob{}

func (cj *CronJob) Delete() error {

	return nil
}

func (cj *CronJob) Direction() string {

	return ""
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

	return ""
}

func (cj *CronJob) RemoteRegistry() globalregistry.Registry {

	return nil
}

func (cj *CronJob) Trigger() string {

	return ""
}
