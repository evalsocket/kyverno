package utils

import (
	"reflect"
	"regexp"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/minio/minio/pkg/wildcard"
	client "github.com/nirmata/kyverno/pkg/dclient"
	dclient "github.com/nirmata/kyverno/pkg/dclient"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//Contains Check if strint is contained in a list of string
func contains(list []string, element string, fn func(string, string) bool) bool {
	for _, e := range list {
		if fn(e, element) {
			return true
		}
	}
	return false
}

//ContainsNamepace check if namespace satisfies any list of pattern(regex)
func ContainsNamepace(patterns []string, ns string) bool {
	return contains(patterns, ns, compareNamespaces)
}

//ContainsString check if the string is contains in a list
func ContainsString(list []string, element string) bool {
	return contains(list, element, compareString)
}

func compareNamespaces(pattern, ns string) bool {
	return wildcard.Match(pattern, ns)
}

func compareString(str, name string) bool {
	return str == name
}

//NewKubeClient returns a new kubernetes client
func NewKubeClient(config *rest.Config) (kubernetes.Interface, error) {
	kclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return kclient, nil
}

//Btoi converts boolean to int
func Btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

//CRDInstalled to check if the CRD is installed or not
func CRDInstalled(discovery client.IDiscovery, log logr.Logger) bool {
	logger := log.WithName("CRDInstalled")
	check := func(kind string) bool {
		gvr := discovery.GetGVRFromKind(kind)
		if reflect.DeepEqual(gvr, (schema.GroupVersionResource{})) {
			logger.Info("CRD not installed", "kind", kind)
			return false
		}
		logger.Info("CRD found", "kind", kind)
		return true
	}
	if !check("ClusterPolicy") || !check("ClusterPolicyViolation") || !check("PolicyViolation") {
		return false
	}
	return true
}

//CleanupOldCrd deletes any existing NamespacedPolicyViolation resources in cluster
// If resource violates policy, new Violations will be generated
func CleanupOldCrd(client *dclient.Client, log logr.Logger) {
	logger := log.WithName("CleanupOldCrd")
	gvr := client.DiscoveryClient.GetGVRFromKind("NamespacedPolicyViolation")
	if !reflect.DeepEqual(gvr, (schema.GroupVersionResource{})) {
		if err := client.DeleteResource("CustomResourceDefinition", "", "namespacedpolicyviolations.kyverno.io", false); err != nil {
			logger.Error(err, "Failed to remove prevous CRD", "kind", "namespacedpolicyviolation")
		}
	}
}

// CompareKubernetesVersion compare kuberneates client version to user given version
func CompareKubernetesVersion(client *client.Client, log logr.Logger, k8smajor, k8sminor, k8ssub int) bool {
	logger := log.WithName("CompareKubernetesVersion")
	serverVersion, err := client.DiscoveryClient.GetServerVersion()
	if err != nil {
		logger.Error(err, "Failed to get kubernetes server version")
		return false
	}
	exp := regexp.MustCompile(`v(\d*).(\d*).(\d*)`)
	groups := exp.FindAllStringSubmatch(serverVersion.String(), -1)
	if len(groups) != 1 || len(groups[0]) != 4 {
		logger.Error(err, "Failed to extract kubernetes server version", "serverVersion", serverVersion)
		return false
	}
	// convert string to int
	// assuming the version are always intergers
	major, err := strconv.Atoi(groups[0][1])
	if err != nil {
		logger.Error(err, "Failed to extract kubernetes major server version", "serverVersion", serverVersion)
		return false
	}
	minor, err := strconv.Atoi(groups[0][2])
	if err != nil {
		logger.Error(err, "Failed to extract kubernetes minor server version", "serverVersion", serverVersion)
		return false
	}
	sub, err := strconv.Atoi(groups[0][3])
	if err != nil {
		logger.Error(err, "Failed to extract kubernetes sub minor server version", "serverVersion", serverVersion)
		return false
	}
	if major <= k8smajor && minor <= k8sminor && sub < k8ssub {
		logger.Info("Unsupported kubernetes server version %s. Kyverno is supported from version v1.12.7+", "serverVersion", serverVersion)
		return false
	}
	return true
}
