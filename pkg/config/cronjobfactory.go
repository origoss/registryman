package config

import (
	"context"
	"fmt"

	"github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/config/registry"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	"github.com/kubermatic-labs/registryman/pkg/skopeo"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type CronJobFactory struct {
	source  globalregistry.Registry
	project globalregistry.Project
}

var _ globalregistry.ReplicationRuleManipulatorProject = &CronJobFactory{}

const image = "quay.io/skopeo/stable"

var clientSet *kubernetes.Clientset

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

func (cjf *CronJobFactory) AssignReplicationRule(ctx context.Context, remoteRegistry globalregistry.Registry, trigger globalregistry.ReplicationTrigger, direction string) (globalregistry.ReplicationRule, error) {
	transfer := skopeo.NewForOperator(cjf.source.GetUsername(), cjf.source.GetPassword())

	projectOfDestinationRegistry := &ProjectOfRegistry{
		Registry: remoteRegistry,
		Project:  cjf.project,
	}

	projectFullPathOfDestination, err := projectOfDestinationRegistry.GenerateProjectRepoName()
	if err != nil {
		return nil, err
	}

	skopeoCommand := transfer.Sync(true, "", projectFullPathOfDestination, remoteRegistry.GetUsername(), remoteRegistry.GetPassword())
	fmt.Println(skopeoCommand)

	// TODO: go text pkg
	scriptContents := fmt.Sprintf("%s%s\n%s",
		`repoArray=(${REPOSITORIES})
for repo in "${repoArray[@]}"
do
	`, concatenateArrayOfStrings(skopeoCommand.Args), "done")

	finalArgs := []string{"-c", scriptContents}

	labels := map[string]string{
		"generator": "registryman-skopeo",
		"project":   cjf.project.GetName(),
		//"direction":       direction,
		"remote-registry": remoteRegistry.GetName()}

	repositories, err := cjf.getRepositories(ctx, remoteRegistry)
	if err != nil {
		return nil, err
	}

	configMap, err := cjf.ApplyConfigMapForRepositories(ctx, repositories, labels)
	if err != nil {
		return nil, err
	}

	cronJob := create(labels, configMap.Name, direction, remoteRegistry, finalArgs, trigger)
	err = writeCronJob(ctx, cronJob)

	return cronJob, err
}

func (cjf *CronJobFactory) getRepositories(ctx context.Context, remoteRegistry globalregistry.Registry) ([]string, error) {
	projectOfSourceRegistry := &ProjectOfRegistry{
		Registry: cjf.source,
		Project:  cjf.project,
	}

	projectFullPathOfSource, err := projectOfSourceRegistry.GenerateProjectRepoName()
	if err != nil {
		return nil, err
	}

	projectWithRepositories, ok := projectOfSourceRegistry.Project.(globalregistry.ProjectWithRepositories)
	if !ok {
		return nil, fmt.Errorf("%s does not have repositories", projectFullPathOfSource)
	}

	return projectWithRepositories.GetRepositories(ctx)
}

func (cjf *CronJobFactory) GetAllCronJobsForProject(ctx context.Context, project globalregistry.Project, sourceRegistryName string) ([]CronJob, error) {
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		return nil, fmt.Errorf("cannot get Kubernetes namespace: %w", err)
	}

	cronJobList, err := clientSet.BatchV1beta1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var results []CronJob
	for _, cj := range cronJobList.Items {
		if cj.Labels["project"] == project.GetName() && cj.Labels["remote-registry"] != sourceRegistryName {
			remoteRegistryByName, err := getRegistryByName(ctx, cj.Labels["remote-registry"])
			if err != nil {
				return nil, err
			}
			cj.TypeMeta = metav1.TypeMeta{
				Kind:       "CronJob",
				APIVersion: "v1beta1",
			}

			cjObject := CronJob{
				remoteRegistry: remoteRegistryByName,
				resource:       &cj,
				//dir:            cj.Labels["direction"],
				dir: "Push",
				replTrigger: &replicationTrigger{
					triggerType:     v1alpha1.CronReplicationTriggerType,
					triggerSchedule: cj.Spec.Schedule,
				},
			}
			results = append(results, cjObject)
		}
	}

	return results, nil
}

func (cjf *CronJobFactory) ApplyConfigMapForRepositories(ctx context.Context, repositories []string, labels map[string]string) (*v1.ConfigMap, error) {
	projectOfSourceRegistry := &ProjectOfRegistry{
		Registry: cjf.source,
		Project:  cjf.project,
	}

	projectFullPathOfSource, err := projectOfSourceRegistry.GenerateProjectRepoName()
	if err != nil {
		return nil, err
	}

	var fullPathOfSourceRepositories []string

	for _, repoName := range repositories {
		repoFullPathOfSource := fmt.Sprintf("%s/%s", projectFullPathOfSource, repoName)
		fullPathOfSourceRepositories = append(fullPathOfSourceRepositories, repoFullPathOfSource)
	}

	configMapData := map[string]string{
		"REPOSITORIES": concatenateArrayOfStrings(fullPathOfSourceRepositories),
	}

	configMap := createConfigMapForEnvvar(labels, configMapData)

	manifestManipulator, err := createManifestManipulator(ctx)
	if err != nil {
		return nil, err
	}

	err = manifestManipulator.WriteResource(ctx, configMap)

	return configMap, err
}

func getRegistryByName(ctx context.Context, name string) (globalregistry.Registry, error) {
	manipulatorCtx := ctx.Value(ResourceManipulatorKey)
	if manipulatorCtx == nil {
		return nil, fmt.Errorf("context shall contain ResourceManipulatorKey")
	}
	kube_aos := manipulatorCtx.(ApiObjectStore)
	registries := kube_aos.GetRegistries(ctx)

	for _, r := range registries {
		if r.Name == name {
			registryFound := registry.New(r, kube_aos)
			return registryFound, nil
		}
	}
	return nil, fmt.Errorf("registry not found with name %s", name)

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

func createManifestManipulator(ctx context.Context) (ManifestManipulator, error) {
	manipulatorCtx := ctx.Value(ResourceManipulatorKey)
	if manipulatorCtx == nil {
		return nil, fmt.Errorf("context shall contain ResourceManipulatorKey")
	}
	manifestManipulator, ok := manipulatorCtx.(ManifestManipulator)
	if !ok {
		return nil, fmt.Errorf("manipulatorCtx is not a proper ManifestManipulator")
	}
	return manifestManipulator, nil
}

func writeCronJob(ctx context.Context, cronJob *CronJob) error {
	manifestManipulator, err := createManifestManipulator(ctx)
	if err != nil {
		return err
	}
	return manifestManipulator.WriteResource(ctx, cronJob.resource)
}
