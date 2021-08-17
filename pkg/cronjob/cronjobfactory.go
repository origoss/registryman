package cronjob

import (
	"fmt"
	"os"

	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	"github.com/kubermatic-labs/registryman/pkg/skopeo"
)

type CronJobFactory struct {
	source  globalregistry.Registry
	project globalregistry.Project
}

var _ globalregistry.ReplicationRuleManipulatorProject = &CronJobFactory{}

func NewCjFactory(source globalregistry.Registry, project globalregistry.Project) *CronJobFactory {
	return &CronJobFactory{
		source:  source,
		project: project,
	}
}

func (cjf *CronJobFactory) AssignReplicationRule(remoteRegistry globalregistry.Registry, trigger, direction string) (globalregistry.ReplicationRule, error) {
	transfer := skopeo.NewForOperator(cjf.source.GetUsername(), cjf.source.GetPassword())

	projectOfSourceRegistry := &config.ProjectOfRegistry{
		Registry: cjf.source,
		Project:  cjf.project,
	}

	projectOfDestinationRegistry := &config.ProjectOfRegistry{
		Registry: remoteRegistry,
		Project:  cjf.project,
	}

	projectFullPathOfSource, err := projectOfSourceRegistry.GenerateProjectRepoName()
	if err != nil {
		return nil, err
	}

	projectFullPathOfDestination, err := projectOfDestinationRegistry.GenerateProjectRepoName()
	if err != nil {
		return nil, err
	}

	projectWithRepositories, ok := projectOfSourceRegistry.Project.(globalregistry.ProjectWithRepositories)
	if !ok {
		return nil, fmt.Errorf("%s does not have repositories", projectFullPathOfSource)
	}

	repositories, err := projectWithRepositories.GetRepositories()
	if err != nil {
		return nil, err
	}

	for _, repoName := range repositories {
		repoFullPathOfSource := fmt.Sprintf("%s/%s", projectFullPathOfSource, repoName)

		skopeoCommand := transfer.Sync(repoFullPathOfSource, projectFullPathOfDestination, &[]string{remoteRegistry.GetUsername(), remoteRegistry.GetPassword()}, nil)

		skopeoCommand.Stderr = os.Stderr
		skopeoCommand.Stdout = os.Stdout

	}

	// Create Cron-job using the returned skopeo sync parameters
	// TODO: Cj config with envvars at creation time

	jobParams := &skopeo.JobParams{
		Command: &skopeo.Command{
			CmdType:     "sync",
			ProjectName: project.GetName(),
			ConfigPath:  "TODO",
		},
		KubeConfig: "TODO",
	}

	err = skopeo.CreateJob(jobParams)
	return nil, nil
}
