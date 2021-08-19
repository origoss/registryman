package cronjob

import (
	"fmt"
	"os"

	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	"github.com/kubermatic-labs/registryman/pkg/skopeo"
	"github.com/spf13/pflag"
	kubernetes "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type CronJobFactory struct {
	source  globalregistry.Registry
	project globalregistry.Project
}

var _ globalregistry.ReplicationRuleManipulatorProject = &CronJobFactory{}

const image = "registryman-skopeo:latest"

var clientSet *kubernetes.Clientset
var clientConfig *rest.Config
var kubeConfig clientcmd.ClientConfig

func init() {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// if you want to change the loading rules (which files in which order), you can do so here

	configOverrides := &clientcmd.ConfigOverrides{}
	clientcmd.BindOverrideFlags(configOverrides, pflag.CommandLine,
		clientcmd.RecommendedConfigOverrideFlags(""))
	// if you want to change override values or bind them to flags, there are methods to help you

	kubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
}

func connectToKube() (*kubernetes.Clientset, error) {
	var err error
	clientConfig, err = kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	kubeClient := kubernetes.NewForConfigOrDie(clientConfig)
	return kubeClient, nil
}

func NewCjFactory(source globalregistry.Registry, project globalregistry.Project) (*CronJobFactory, error) {
	resultClientSet, err := connectToKube()
	if err != nil {
		return nil, err
	}
	clientSet = resultClientSet

	return &CronJobFactory{
		source:  source,
		project: project,
	}, nil
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
