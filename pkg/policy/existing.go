package policy

import (
	"reflect"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/minio/minio/pkg/wildcard"
	kyverno "github.com/nirmata/kyverno/pkg/api/kyverno/v1"
	"github.com/nirmata/kyverno/pkg/config"
	client "github.com/nirmata/kyverno/pkg/dclient"
	"github.com/nirmata/kyverno/pkg/engine"
	"github.com/nirmata/kyverno/pkg/engine/response"
	"github.com/nirmata/kyverno/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
)

func (pc *PolicyController) processExistingResources(policy kyverno.ClusterPolicy) []response.EngineResponse {
	logger := pc.log.WithValues("policy", policy.Name)
	// Parse through all the resources
	// drops the cache after configured rebuild time
	pc.rm.Drop()
	var engineResponses []response.EngineResponse
	// get resource that are satisfy the resource description defined in the rules
	resourceMap := listResources(pc.client, policy, pc.configHandler, logger)
	for _, resource := range resourceMap {
		// pre-processing, check if the policy and resource version has been processed before
		if !pc.rm.ProcessResource(policy.Name, policy.ResourceVersion, resource.GetKind(), resource.GetNamespace(), resource.GetName(), resource.GetResourceVersion()) {
			logger.V(4).Info("policy and resource already processed", "policyResourceVersion", policy.ResourceVersion, "resourceResourceVersion", resource.GetResourceVersion(), "kind", resource.GetKind(), "namespace", resource.GetNamespace(), "name", resource.GetName())
			continue
		}

		// skip reporting violation on pod which has annotation pod-policies.kyverno.io/autogen-applied
		ann := resource.GetAnnotations()
		if _, ok := ann[engine.PodTemplateAnnotation]; ok {
			if ann[engine.PodTemplateAnnotation] == "none" {
				if skipPodApplication(resource, logger) {
					continue
				}
			}
		}

		// apply the policy on each
		engineResponse := applyPolicy(policy, resource, logger)
		// get engine response for mutation & validation independently
		engineResponses = append(engineResponses, engineResponse...)
		// post-processing, register the resource as processed
		pc.rm.RegisterResource(policy.GetName(), policy.GetResourceVersion(), resource.GetKind(), resource.GetNamespace(), resource.GetName(), resource.GetResourceVersion())
	}
	return engineResponses
}

func listResources(client *client.Client, policy kyverno.ClusterPolicy, configHandler config.Interface, log logr.Logger) map[string]unstructured.Unstructured {
	// key uid
	resourceMap := map[string]unstructured.Unstructured{}

	for _, rule := range policy.Spec.Rules {
		// resources that match
		for _, k := range rule.MatchResources.Kinds {
			var namespaces []string
			if len(rule.MatchResources.Namespaces) > 0 {
				namespaces = append(namespaces, rule.MatchResources.Namespaces...)
				log.V(4).Info("namespaces included", "namespaces", rule.MatchResources.Namespaces)
			} else {
				log.V(4).Info("processing all namespaces", "rule", rule.Name)
				// get all namespaces
				namespaces = getAllNamespaces(client, log)
			}

			// get resources in the namespaces
			for _, ns := range namespaces {
				rMap := getResourcesPerNamespace(k, client, ns, rule, configHandler, log)
				mergeresources(resourceMap, rMap)
			}

		}
	}
	return resourceMap
}

func getResourcesPerNamespace(kind string, client *client.Client, namespace string, rule kyverno.Rule, configHandler config.Interface, log logr.Logger) map[string]unstructured.Unstructured {
	resourceMap := map[string]unstructured.Unstructured{}
	// merge include and exclude label selector values
	ls := rule.MatchResources.Selector
	//	ls := mergeLabelSectors(rule.MatchResources.Selector, rule.ExcludeResources.Selector)
	// list resources
	log.V(4).Info("list resources to be processed")
	list, err := client.ListResource(kind, namespace, ls)
	if err != nil {
		log.Error(err, "failed to list resources", "kind", kind)
		return nil
	}
	// filter based on name
	for _, r := range list.Items {
		// match name
		if rule.MatchResources.Name != "" {
			if !wildcard.Match(rule.MatchResources.Name, r.GetName()) {
				continue
			}
		}
		// Skip the filtered resources
		if configHandler.ToFilter(r.GetKind(), r.GetNamespace(), r.GetName()) {
			continue
		}

		//TODO check if the group version kind is present or not
		resourceMap[string(r.GetUID())] = r
	}

	// exclude the resources
	// skip resources to be filtered
	excludeResources(resourceMap, rule.ExcludeResources.ResourceDescription, configHandler, log)
	return resourceMap
}

func excludeResources(included map[string]unstructured.Unstructured, exclude kyverno.ResourceDescription, configHandler config.Interface, log logr.Logger) {
	if reflect.DeepEqual(exclude, (kyverno.ResourceDescription{})) {
		return
	}
	excludeName := func(name string) Condition {
		if exclude.Name == "" {
			return NotEvaluate
		}
		if wildcard.Match(exclude.Name, name) {
			return Skip
		}
		return Process
	}

	excludeNamespace := func(namespace string) Condition {
		if len(exclude.Namespaces) == 0 {
			return NotEvaluate
		}
		if utils.ContainsNamepace(exclude.Namespaces, namespace) {
			return Skip
		}
		return Process
	}

	excludeSelector := func(labelsMap map[string]string) Condition {
		if exclude.Selector == nil {
			return NotEvaluate
		}
		selector, err := metav1.LabelSelectorAsSelector(exclude.Selector)
		// if the label selector is incorrect, should be fail or
		if err != nil {
			log.Error(err, "failed to build label selector")
			return Skip
		}
		if selector.Matches(labels.Set(labelsMap)) {
			return Skip
		}
		return Process
	}

	findKind := func(kind string, kinds []string) bool {
		for _, k := range kinds {
			if k == kind {
				return true
			}
		}
		return false
	}

	excludeKind := func(kind string) Condition {
		if len(exclude.Kinds) == 0 {
			return NotEvaluate
		}

		if findKind(kind, exclude.Kinds) {
			return Skip
		}

		return Process
	}

	// check exclude condition for each resource
	for uid, resource := range included {
		// 0 -> dont check
		// 1 -> is not to be exclude
		// 2 -> to be exclude
		excludeEval := []Condition{}

		if ret := excludeName(resource.GetName()); ret != NotEvaluate {
			excludeEval = append(excludeEval, ret)
		}
		if ret := excludeNamespace(resource.GetNamespace()); ret != NotEvaluate {
			excludeEval = append(excludeEval, ret)
		}
		if ret := excludeSelector(resource.GetLabels()); ret != NotEvaluate {
			excludeEval = append(excludeEval, ret)
		}
		if ret := excludeKind(resource.GetKind()); ret != NotEvaluate {
			excludeEval = append(excludeEval, ret)
		}
		// exclude the filtered resources
		if configHandler.ToFilter(resource.GetKind(), resource.GetNamespace(), resource.GetName()) {
			delete(included, uid)
			continue
		}

		func() bool {
			for _, ret := range excludeEval {
				if ret == Process {
					// Process the resources
					continue
				}
			}
			// Skip the resource from processing
			delete(included, uid)
			return false
		}()
	}
}

//Condition defines condition type
type Condition int

const (
	//NotEvaluate to not evaluate condition
	NotEvaluate Condition = 0
	// Process to evaluate condition
	Process Condition = 1
	// Skip to ignore/skip the condition
	Skip Condition = 2
)

// merge b into a map
func mergeresources(a, b map[string]unstructured.Unstructured) {
	for k, v := range b {
		a[k] = v
	}
}

func getAllNamespaces(client *client.Client, log logr.Logger) []string {

	var namespaces []string
	// get all namespaces
	nsList, err := client.ListResource("Namespace", "", nil)
	if err != nil {
		log.Error(err, "failed to list namespaces")
		return namespaces
	}
	for _, ns := range nsList.Items {
		namespaces = append(namespaces, ns.GetName())
	}
	return namespaces
}

//NewResourceManager returns a new ResourceManager
func NewResourceManager(rebuildTime int64) *ResourceManager {
	rm := ResourceManager{
		data:        make(map[string]interface{}),
		time:        time.Now(),
		rebuildTime: rebuildTime,
	}
	// set time it was built
	return &rm
}

// ResourceManager stores the details on already processed resources for caching
type ResourceManager struct {
	// we drop and re-build the cache
	// based on the memory consumer of by the map
	data        map[string]interface{}
	mux         sync.RWMutex
	time        time.Time
	rebuildTime int64 // after how many seconds should we rebuild the cache
}

type resourceManager interface {
	ProcessResource(policy, pv, kind, ns, name, rv string) bool
	//TODO	removeResource(kind, ns, name string) error
	RegisterResource(policy, pv, kind, ns, name, rv string)
	// reload
	Drop()
}

//Drop drop the cache after every rebuild interval mins
//TODO: or drop based on the size
func (rm *ResourceManager) Drop() {
	timeSince := time.Since(rm.time)
	if timeSince > time.Duration(rm.rebuildTime)*time.Second {
		rm.mux.Lock()
		defer rm.mux.Unlock()
		rm.data = map[string]interface{}{}
		rm.time = time.Now()
	}
}

var empty struct{}

//RegisterResource stores if the policy is processed on this resource version
func (rm *ResourceManager) RegisterResource(policy, pv, kind, ns, name, rv string) {
	rm.mux.Lock()
	defer rm.mux.Unlock()
	// add the resource
	key := buildKey(policy, pv, kind, ns, name, rv)
	rm.data[key] = empty
}

//ProcessResource returns true if the policy was not applied on the resource
func (rm *ResourceManager) ProcessResource(policy, pv, kind, ns, name, rv string) bool {
	rm.mux.RLock()
	defer rm.mux.RUnlock()

	key := buildKey(policy, pv, kind, ns, name, rv)
	_, ok := rm.data[key]
	return !ok
}

func buildKey(policy, pv, kind, ns, name, rv string) string {
	return policy + "/" + pv + "/" + kind + "/" + ns + "/" + name + "/" + rv
}

func skipPodApplication(resource unstructured.Unstructured, log logr.Logger) bool {
	if resource.GetKind() != "Pod" {
		return false
	}

	annotation := resource.GetAnnotations()
	if _, ok := annotation[engine.PodTemplateAnnotation]; ok {
		log.V(4).Info("Policies already processed on pod controllers, skip processing policy on Pod", "kind", resource.GetKind(), "namespace", resource.GetNamespace(), "name", resource.GetName())
		return true
	}

	return false
}
