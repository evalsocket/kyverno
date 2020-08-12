package policyreport

import (
	"errors"
	"reflect"
	"fmt"
	"context"

	"github.com/nirmata/kyverno/pkg/constant"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"github.com/go-logr/logr"
	kyverno "github.com/nirmata/kyverno/pkg/api/kyverno/v1"
	policyreportv1alpha1 "github.com/nirmata/kyverno/pkg/client/clientset/versioned/typed/policyreport/v1alpha1"
	policyreportlister "github.com/nirmata/kyverno/pkg/client/listers/policyreport/v1alpha1"
	client "github.com/nirmata/kyverno/pkg/dclient"
	"github.com/nirmata/kyverno/pkg/policystatus"
)

const nsWorkQueueName = "policy-report-namespace"
const nsWorkQueueRetryLimit = 3

//namespacedPR ...
type namespacedPR struct {
	// dynamic client
	dclient *client.Client
	// get/list namespaced policy violation
	nsprLister policyreportlister.KyvernoKyvernoPolicyReportLister
	// policy violation interface
	policyreportInterface policyreportv1alpha1.PolicyV1alpha1Interface
	// logger
	log logr.Logger
	// update policy status with violationCount
	policyStatusListener policystatus.Listener

	dataStore            *dataStore

	queue  workqueue.RateLimitingInterface

}


func newNamespacedPR(log logr.Logger, dclient *client.Client,
	nsprLister policyreportlister.KyvernoKyvernoPolicyReportLister,
	policyreportInterface policyreportv1alpha1.PolicyV1alpha1Interface,
	policyStatus policystatus.Listener,
) *namespacedPR {
	nspr := namespacedPR{
		dclient:              dclient,
		nsprLister:           nsprLister,
		policyreportInterface:     policyreportInterface,
		log:                  log,
		policyStatusListener: policyStatus,
		dataStore:            newDataStore(),
		queue:  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), nsWorkQueueName),
	}
	return &nspr
}


func (nspr *namespacedPR) enqueue(info Info) {
	// add to data map
	keyHash := info.toKey()
	// add to
	// queue the key hash
	nspr.dataStore.add(keyHash, info)
	nspr.queue.Add(keyHash)
}

//Add queues a policy violation create request
func (nspr *namespacedPR) Add(infos ...Info) {
	for _, info := range infos {
		nspr.enqueue(info)
	}
}

// Run starts the workers
func (nspr *namespacedPR) Run(workers int, stopCh <-chan struct{}) {
	logger := nspr.log
	defer utilruntime.HandleCrash()
	logger.Info("start")
	defer logger.Info("shutting down")

	for i := 0; i < workers; i++ {
		go wait.Until(nspr.runWorker, constant.PolicyViolationControllerResync, stopCh)
	}
	<-stopCh
}

func (nspr *namespacedPR) runWorker() {
	for nspr.processNextWorkItem() {
	}
}

func (nspr *namespacedPR) handleErr(err error, key interface{}) {
	logger := nspr.log
	if err == nil {
		nspr.queue.Forget(key)
		return
	}

	// retires requests if there is error
	if nspr.queue.NumRequeues(key) < nsWorkQueueRetryLimit {
		logger.Error(err, "failed to sync policy violation", "key", key)
		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		nspr.queue.AddRateLimited(key)
		return
	}
	nspr.queue.Forget(key)
	// remove from data store
	if keyHash, ok := key.(string); ok {
		nspr.dataStore.delete(keyHash)
	}
	logger.Error(err, "dropping key out of the queue", "key", key)
}

func (nspr *namespacedPR) processNextWorkItem() bool {
	logger := nspr.log
	obj, shutdown := nspr.queue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer nspr.queue.Done(obj)
		var keyHash string
		var ok bool

		if keyHash, ok = obj.(string); !ok {
			nspr.queue.Forget(obj)
			logger.Info("incorrect type; expecting type 'string'", "obj", obj)
			return nil
		}

		// lookup data store
		info := nspr.dataStore.lookup(keyHash)
		if reflect.DeepEqual(info, Info{}) {
			// empty key
			nspr.queue.Forget(obj)
			logger.Info("empty key")
			return nil
		}

		err := nspr.syncHandler(info)
		nspr.handleErr(err, obj)
		return nil
	}(obj)

	if err != nil {
		logger.Error(err, "failed to process item")
		return true
	}

	return true
}

func (nspr *namespacedPR) syncHandler(info Info) error {
	logger := nspr.log
	failure := false
	builder := newPvBuilder()

	pv := builder.generate(info)

	if info.FromSync {
		pv.Annotations = map[string]string{
			"fromSync": "true",
		}
	}

	// Create Policy Violations
	logger.V(4).Info("creating policy violation", "key", info.toKey())
	if err := nspr.create(pv); err != nil {
		failure = true
		logger.Error(err, "failed to create policy violation")
	}

	if failure {
		// even if there is a single failure we requeue the request
		return errors.New("Failed to process some policy violations, re-queuing")
	}
	return nil
}


func (nspr *namespacedPR) create(pv kyverno.KyvernoKyvernoPolicyReportTemplate) error {
	policyName := fmt.Sprintf("kyverno-policyreport",)
	clusterpr,err:= nspr.policyreportInterface.KyvernoKyvernoPolicyReports(pv.Namespace).Get(context.Background(),policyName,v1.GetOptions{});
	if err != nil {
		return err
	}
	cpr := PolicyViolationsToKyvernoKyvernoPolicyReport(&pv,clusterpr)
	cpr,err = nspr.policyreportInterface.KyvernoKyvernoPolicyReports(pv.Namespace).Update(context.Background(),cpr,v1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}
