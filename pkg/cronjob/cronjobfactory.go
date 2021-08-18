package cronjob

import (
	"fmt"
	"os"

	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	"github.com/kubermatic-labs/registryman/pkg/skopeo"
	kubernetes "k8s.io/client-go/kubernetes"
)

type CronJobFactory struct {
	source  globalregistry.Registry
	project globalregistry.Project
}

var _ globalregistry.ReplicationRuleManipulatorProject = &CronJobFactory{}

const image = "registryman-skopeo:latest"

var clientset *kubernetes.Clientset

func NewCjFactory(source globalregistry.Registry, project globalregistry.Project) *CronJobFactory {
	// TODO: connect to K8s -> initialize clientset
	// clientset, err := kubernetes.NewForConfig(config)
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

		fmt.Println(skopeoCommand)

		// Create Cron-job config using the returned skopeo sync parameters
		// TODO: Cj config with envvars at creation time
		cronJob := new(cjf.project.GetName(), "default", repoName, &skopeoCommand.Args, &remoteRegistry)

		if err := cronJob.Deploy(); err != nil {
			return nil, err
		}

	}

	return nil, nil
}
