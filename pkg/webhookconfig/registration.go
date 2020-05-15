package webhookconfig

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/nirmata/kyverno/pkg/config"
	client "github.com/nirmata/kyverno/pkg/dclient"
	admregapi "k8s.io/api/admissionregistration/v1beta1"
	errorsapi "k8s.io/apimachinery/pkg/api/errors"
	rest "k8s.io/client-go/rest"
)

const (
	//MutatingWebhookConfigurationKind defines the kind for MutatingWebhookConfiguration
	MutatingWebhookConfigurationKind string = "MutatingWebhookConfiguration"
	//ValidatingWebhookConfigurationKind defines the kind for ValidatingWebhookConfiguration
	ValidatingWebhookConfigurationKind string = "ValidatingWebhookConfiguration"
)

// WebhookRegistrationClient is client for registration webhooks on cluster
type WebhookRegistrationClient struct {
	client       *client.Client
	clientConfig *rest.Config
	// serverIP should be used if running Kyverno out of clutser
	serverIP       string
	timeoutSeconds int32
	log            logr.Logger
}

// NewWebhookRegistrationClient creates new WebhookRegistrationClient instance
func NewWebhookRegistrationClient(
	clientConfig *rest.Config,
	client *client.Client,
	serverIP string,
	webhookTimeout int32,
	log logr.Logger) *WebhookRegistrationClient {
	return &WebhookRegistrationClient{
		clientConfig:   clientConfig,
		client:         client,
		serverIP:       serverIP,
		timeoutSeconds: webhookTimeout,
		log:            log.WithName("WebhookRegistrationClient"),
	}
}

// Register creates admission webhooks configs on cluster
func (wrc *WebhookRegistrationClient) Register() error {
	logger := wrc.log.WithName("Register")
	if wrc.serverIP != "" {
		logger.Info("Registering webhook", "url", fmt.Sprintf("https://%s", wrc.serverIP))
	}

	// For the case if cluster already has this configs
	// remove previously create webhookconfigurations if any
	// webhook configurations are created dynamically based on the policy resources
	wrc.removeWebhookConfigurations()

	// create Verify mutating webhook configuration resource
	// that is used to check if admission control is enabled or not
	if err := wrc.createVerifyMutatingWebhookConfiguration(); err != nil {
		return err
	}

	// Static Webhook configuration on Policy CRD
	// create Policy CRD validating webhook configuration resource
	// used for validating Policy CR
	if err := wrc.createPolicyValidatingWebhookConfiguration(); err != nil {
		return err
	}
	// create Policy CRD validating webhook configuration resource
	// used for defauling values in Policy CR
	if err := wrc.createPolicyMutatingWebhookConfiguration(); err != nil {
		return err
	}

	return nil
}

// RemoveWebhookConfigurations removes webhook configurations for reosurces and policy
// called during webhook server shutdown
func (wrc *WebhookRegistrationClient) RemoveWebhookConfigurations(cleanUp chan<- struct{}) {
	//TODO: dupliate, but a placeholder to perform more error handlind during cleanup
	wrc.removeWebhookConfigurations()
	// close channel to notify cleanup is complete
	close(cleanUp)
}

//CreateResourceMutatingWebhookConfiguration create a Mutatingwebhookconfiguration resource for all resource type
// used to forward request to kyverno webhooks to apply policeis
// Mutationg webhook is be used for Mutating purpose
func (wrc *WebhookRegistrationClient) CreateResourceMutatingWebhookConfiguration() error {
	logger := wrc.log
	var caData []byte
	var config *admregapi.MutatingWebhookConfiguration

	// read CA data from
	// 1) secret(config)
	// 2) kubeconfig
	if caData = wrc.readCaData(); caData == nil {
		return errors.New("Unable to extract CA data from configuration")
	}
	// if serverIP is specified we assume its debug mode
	if wrc.serverIP != "" {
		// debug mode
		// clientConfig - URL
		config = wrc.constructDebugMutatingWebhookConfig(caData)
	} else {
		// clientConfig - service
		config = wrc.constructMutatingWebhookConfig(caData)
	}
	_, err := wrc.client.CreateResource(MutatingWebhookConfigurationKind, "", *config, false)
	if errorsapi.IsAlreadyExists(err) {
		logger.V(4).Info("resource mutating webhook configuration already exists. not creating one", "name", config.Name)
		return nil
	}
	if err != nil {
		logger.Error(err, "failed to create resource mutating webhook configuration", "name", config.Name)
		return err
	}
	return nil
}

//CreateResourceValidatingWebhookConfiguration ...
func (wrc *WebhookRegistrationClient) CreateResourceValidatingWebhookConfiguration() error {
	var caData []byte
	var config *admregapi.ValidatingWebhookConfiguration

	if caData = wrc.readCaData(); caData == nil {
		return errors.New("Unable to extract CA data from configuration")
	}
	// if serverIP is specified we assume its debug mode
	if wrc.serverIP != "" {
		// debug mode
		// clientConfig - URL
		config = wrc.constructDebugValidatingWebhookConfig(caData)
	} else {
		// clientConfig - service
		config = wrc.constructValidatingWebhookConfig(caData)
	}
	logger := wrc.log.WithValues("kind", ValidatingWebhookConfigurationKind, "name", config.Name)

	_, err := wrc.client.CreateResource(ValidatingWebhookConfigurationKind, "", *config, false)
	if errorsapi.IsAlreadyExists(err) {
		logger.V(4).Info("resource already exists. not create one")
		return nil
	}
	if err != nil {
		logger.Error(err, "failed to create resource")
		return err
	}
	return nil
}

//registerPolicyValidatingWebhookConfiguration create a Validating webhook configuration for Policy CRD
func (wrc *WebhookRegistrationClient) createPolicyValidatingWebhookConfiguration() error {
	var caData []byte
	var config *admregapi.ValidatingWebhookConfiguration

	// read CA data from
	// 1) secret(config)
	// 2) kubeconfig
	if caData = wrc.readCaData(); caData == nil {
		return errors.New("Unable to extract CA data from configuration")
	}

	// if serverIP is specified we assume its debug mode
	if wrc.serverIP != "" {
		// debug mode
		// clientConfig - URL
		config = wrc.contructDebugPolicyValidatingWebhookConfig(caData)
	} else {
		// clientConfig - service
		config = wrc.contructPolicyValidatingWebhookConfig(caData)
	}
	logger := wrc.log.WithValues("kind", ValidatingWebhookConfigurationKind, "name", config.Name)

	// create validating webhook configuration resource
	if _, err := wrc.client.CreateResource(ValidatingWebhookConfigurationKind, "", *config, false); err != nil {
		return err
	}
	logger.V(4).Info("created resource")
	return nil
}

func (wrc *WebhookRegistrationClient) createPolicyMutatingWebhookConfiguration() error {
	var caData []byte
	var config *admregapi.MutatingWebhookConfiguration
	// read CA data from
	// 1) secret(config)
	// 2) kubeconfig
	if caData = wrc.readCaData(); caData == nil {
		return errors.New("Unable to extract CA data from configuration")
	}

	// if serverIP is specified we assume its debug mode
	if wrc.serverIP != "" {
		// debug mode
		// clientConfig - URL
		config = wrc.contructDebugPolicyMutatingWebhookConfig(caData)
	} else {
		// clientConfig - service
		config = wrc.contructPolicyMutatingWebhookConfig(caData)
	}

	// create mutating webhook configuration resource
	if _, err := wrc.client.CreateResource(MutatingWebhookConfigurationKind, "", *config, false); err != nil {
		return err
	}
	wrc.log.V(4).Info("reated Mutating Webhook Configuration", "name", config.Name)
	return nil
}

func (wrc *WebhookRegistrationClient) createVerifyMutatingWebhookConfiguration() error {
	var caData []byte
	var config *admregapi.MutatingWebhookConfiguration

	// read CA data from
	// 1) secret(config)
	// 2) kubeconfig
	if caData = wrc.readCaData(); caData == nil {
		return errors.New("Unable to extract CA data from configuration")
	}

	// if serverIP is specified we assume its debug mode
	if wrc.serverIP != "" {
		// debug mode
		// clientConfig - URL
		config = wrc.constructDebugVerifyMutatingWebhookConfig(caData)
	} else {
		// clientConfig - service
		config = wrc.constructVerifyMutatingWebhookConfig(caData)
	}

	// create mutating webhook configuration resource
	if _, err := wrc.client.CreateResource(MutatingWebhookConfigurationKind, "", *config, false); err != nil {
		return err
	}

	wrc.log.V(4).Info("reated Mutating Webhook Configuration", "name", config.Name)
	return nil
}

// DeregisterAll deletes webhook configs from cluster
// This function does not fail on error:
// Register will fail if the config exists, so there is no need to fail on error
func (wrc *WebhookRegistrationClient) removeWebhookConfigurations() {
	startTime := time.Now()
	wrc.log.Info("Started cleaning up webhookconfigurations")
	defer func() {
		wrc.log.V(4).Info("Finished cleaning up webhookcongfigurations", "processingTime", time.Since(startTime))
	}()

	var wg sync.WaitGroup

	wg.Add(5)
	// mutating and validating webhook configuration for Kubernetes resources
	go wrc.removeResourceMutatingWebhookConfiguration(&wg)
	go wrc.removeResourceValidatingWebhookConfiguration(&wg)
	// mutating and validating webhook configurtion for Policy CRD resource
	go wrc.removePolicyMutatingWebhookConfiguration(&wg)
	go wrc.removePolicyValidatingWebhookConfiguration(&wg)
	// mutating webhook configuration for verifying webhook
	go wrc.removeVerifyWebhookMutatingWebhookConfig(&wg)

	// wait for the removal go routines to return
	wg.Wait()
}

// wrapper to handle wait group
// TODO: re-work with RemoveResourceMutatingWebhookConfiguration, as the only difference is wg handling
func (wrc *WebhookRegistrationClient) removeResourceMutatingWebhookConfiguration(wg *sync.WaitGroup) {
	defer wg.Done()
	if err := wrc.RemoveResourceMutatingWebhookConfiguration(); err != nil {
		wrc.log.Error(err, "failed to remove resource mutating webhook configuration")
	}
}
func (wrc *WebhookRegistrationClient) removeResourceValidatingWebhookConfiguration(wg *sync.WaitGroup) {
	defer wg.Done()
	if err := wrc.RemoveResourceValidatingWebhookConfiguration(); err != nil {
		wrc.log.Error(err, "failed to remove resource validation webhook configuration")
	}
}

// delete policy mutating webhookconfigurations
// handle wait group
func (wrc *WebhookRegistrationClient) removePolicyMutatingWebhookConfiguration(wg *sync.WaitGroup) {
	defer wg.Done()
	// Mutating webhook configuration
	var mutatingConfig string
	if wrc.serverIP != "" {
		mutatingConfig = config.PolicyMutatingWebhookConfigurationDebugName
	} else {
		mutatingConfig = config.PolicyMutatingWebhookConfigurationName
	}
	logger := wrc.log.WithValues("name", mutatingConfig)
	logger.V(4).Info("removing mutating webhook configuration")
	err := wrc.client.DeleteResource(MutatingWebhookConfigurationKind, "", mutatingConfig, false)
	if errorsapi.IsNotFound(err) {
		logger.Error(err, "policy mutating webhook configuration does not exist, not deleting")
	} else if err != nil {
		logger.Error(err, "failed to delete policy mutating webhook configuration")
	} else {
		logger.V(4).Info("successfully deleted policy mutating webhook configutation")
	}
}

// delete policy validating webhookconfigurations
// handle wait group
func (wrc *WebhookRegistrationClient) removePolicyValidatingWebhookConfiguration(wg *sync.WaitGroup) {
	defer wg.Done()
	// Validating webhook configuration
	var validatingConfig string
	if wrc.serverIP != "" {
		validatingConfig = config.PolicyValidatingWebhookConfigurationDebugName
	} else {
		validatingConfig = config.PolicyValidatingWebhookConfigurationName
	}
	logger := wrc.log.WithValues("name", validatingConfig)
	logger.V(4).Info("removing validating webhook configuration")
	err := wrc.client.DeleteResource(ValidatingWebhookConfigurationKind, "", validatingConfig, false)
	if errorsapi.IsNotFound(err) {
		logger.Error(err, "policy validating webhook configuration does not exist, not deleting")
	} else if err != nil {
		logger.Error(err, "failed to delete policy validating webhook configuration")
	} else {
		logger.V(4).Info("successfully deleted policy validating webhook configutation")
	}
}

// GetWebhookTimeOut returns the value of webhook timeout
func (wrc *WebhookRegistrationClient) GetWebhookTimeOut() int32 {
	return wrc.timeoutSeconds
}
