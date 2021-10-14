package cronjob

import (
	"context"
	"fmt"
	"os"

	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	"github.com/kubermatic-labs/registryman/pkg/skopeo"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type CronJobFactory struct {
	source  globalregistry.Registry
	project globalregistry.Project
}

var _ globalregistry.ReplicationRuleManipulatorProject = &CronJobFactory{}

const image = "quay.io/skopeo/stable"

var clientSet *kubernetes.Clientset
var clientConfig *rest.Config
var kubeConfig clientcmd.ClientConfig

func init() {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// if you want to change the loading rules (which files in which order), you can do so here

	configOverrides := &clientcmd.ConfigOverrides{}
	// clientcmd.BindOverrideFlags(configOverrides, pflag.CommandLine,
	// 	clientcmd.RecommendedConfigOverrideFlags(""))
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

func (cjf *CronJobFactory) AssignReplicationRule(ctx context.Context, remoteRegistry globalregistry.Registry, trigger, direction string) (globalregistry.ReplicationRule, error) {
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

	repositories, err := projectWithRepositories.GetRepositories(ctx)
	if err != nil {
		return nil, err
	}

	skopeoCommand := transfer.Sync(true, "", projectFullPathOfDestination, &[]string{remoteRegistry.GetUsername(), remoteRegistry.GetPassword()}, nil)
	fmt.Println(skopeoCommand)

	skopeoCommand.Stderr = os.Stderr
	skopeoCommand.Stdout = os.Stdout

	// TODO: go text pkg
	scriptContents := fmt.Sprintf("%s%s\n%s",
		`repoArray=(${REPOSITORIES})
for repo in "${repoArray[@]}"
do
	`, concatenateArrayOfStrings(skopeoCommand.Args), "done")

	var fullPathOfSourceRepositories []string

	for _, repoName := range repositories {
		repoFullPathOfSource := fmt.Sprintf("%s/%s", projectFullPathOfSource, repoName)
		fullPathOfSourceRepositories = append(fullPathOfSourceRepositories, repoFullPathOfSource)
	}

	finalArgs := []string{"-c", scriptContents}

	labels := map[string]string{
		"generator":       "registryman-skopeo",
		"project":         cjf.project.GetName(),
		"remote-registry": remoteRegistry.GetName()}

	manipulatorCtx := ctx.Value(config.ResourceManipulatorKey)
	if manipulatorCtx == nil {
		return nil, fmt.Errorf("context shall contain ResourceManipulatorKey")
	}
	manifestManipulator, ok := manipulatorCtx.(config.ManifestManipulator)
	if !ok {
		return nil, fmt.Errorf("manipulatorCtx is not a proper ManifestManipulator")
	}

	configMapData := map[string]string{
		"REPOSITORIES": concatenateArrayOfStrings(fullPathOfSourceRepositories),
	}
	configMap := createConfigMapForEnvvar(labels, configMapData)

	cronJob := create(labels, remoteRegistry, finalArgs, configMap.Name)

	err = manifestManipulator.WriteResource(ctx, configMap)
	if err != nil {
		return nil, err
	}

	err = manifestManipulator.WriteResource(ctx, cronJob.resource)

	return nil, err
}

func (cjf *CronJobFactory) GetAllCronJobs(registryName string, ctx context.Context) (*[]CronJob, error) {
	cronJobList, err := clientSet.BatchV1beta1().CronJobs(registryName).List(ctx, v1.ListOptions{})

	if err != nil {
		return nil, err
	}

	var results []CronJob
	for _, cj := range cronJobList.Items {
		cjObject := CronJob{
			resource: &cj,
		}
		results = append(results, cjObject)
	}

	return &results, nil
}

func concatenateArrayOfStrings(arg []string) string {
	result := ""
	for i, word := range arg {
		if i != 0 {
			result = fmt.Sprintf("%s %s", result, word)
		} else {
			result = fmt.Sprintf("%s%s", result, word)
		}
	}
	return result
}
