package pds

import (
	"crypto/tls"
	"fmt"
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	pdsapi "github.com/portworx/torpedo/drivers/pds/api"
	pdscontrolplane "github.com/portworx/torpedo/drivers/pds/controlplane"
	"github.com/portworx/torpedo/drivers/scheduler"
	unifiedPlatform "github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/pkg/errors"
	"github.com/portworx/torpedo/pkg/log"
	platformv2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/url"
	"os"
)

type LoadGenParams struct {
	LoadGenDepName    string
	PdsDeploymentName string
	Namespace         string
	FailOnError       string
	Mode              string
	TableName         string
	NumOfRows         string
	Iterations        string
	Timeout           string //example 60s
	ReplacePassword   string
	ClusterMode       string
	Replicas          int32
}

type Driver interface {

	//DeployPDSDataservices Deploys the given PDS dataservice and retruns the models deployment object
	DeployPDSDataservices() ([]*pds.ModelsDeployment, error)

	//CreateSchedulerContextForPDSApps Creates Context for the pds deployed applications
	CreateSchedulerContextForPDSApps(pdsApps []*pds.ModelsDeployment) ([]*scheduler.Context, error)

	//ValidateDataServiceDeployment Validate the PDS deployments
	ValidateDataServiceDeployment(deployment *pds.ModelsDeployment, namespace string) error

	//InsertDataAndReturnChecksum Inserts data and returns md5 hash for the data inserted
	InsertDataAndReturnChecksum(pdsDeployment *pds.ModelsDeployment, wkloadGenParams LoadGenParams) (string, *v1.Deployment, error)

	//ReadDataAndReturnChecksum Reads data and returns md5 hash for the data
	ReadDataAndReturnChecksum(pdsDeployment *pds.ModelsDeployment, wkloadGenParams LoadGenParams) (string, *v1.Deployment, error)
}

var (
	pdsschedulers = make(map[string]Driver)
)

func GetK8sContext() (*kubernetes.Clientset, *rest.Config, error) {
	kubeconfigPath := os.Getenv("KUBECONFIG")

	// Build the client configuration from the kubeconfig file.
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, nil, fmt.Errorf("Error creating client configuration from kubeconfig: %v\n", err)
	}
	// Create the Kubernetes client using the configuration.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("Error creating clientset: %v\n", err)
	}
	return clientset, config, nil

}

func InitUnifiedPlatformApiComponents(controlPlaneURL, accountID string, insecureDialOpt bool) (*unifiedPlatform.UnifiedPlatformComponents, error) {
	log.InfoD("Initializing Api components")

	// generate pds api client
	pdsApiConf := pdsv2.NewConfiguration()
	endpointURL, err := url.Parse(controlPlaneURL)
	log.Infof("controlPlane url is [%s]", endpointURL)
	if err != nil {
		return nil, err
	}
	pdsApiConf.Host = endpointURL.Host
	pdsApiConf.Scheme = endpointURL.Scheme
	pdsV2apiClient := pdsv2.NewAPIClient(pdsApiConf)

	//generate platform api client
	platformApiConf := platformv2.NewConfiguration()
	endpointURL, err = url.Parse(controlPlaneURL)
	if err != nil {
		return nil, err
	}
	platformApiConf.Host = endpointURL.Host
	platformApiConf.Scheme = endpointURL.Scheme
	platformV2apiClient := platformv2.NewAPIClient(platformApiConf)

	//Initialize the grpc client here and pass it to the NewUnifiedPlatformComponents method
	dialOpts := []grpc.DialOption{}
	if insecureDialOpt {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig := &tls.Config{}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}
	grpcClient, err := grpc.Dial(endpointURL.Host, dialOpts...)
	if err != nil {
		return nil, err
	}

	components := unifiedPlatform.NewUnifiedPlatformComponents(platformV2apiClient, pdsV2apiClient, grpcClient, accountID)

	return components, nil
}
func InitPdsApiComponents(ControlPlaneURL string) (*pdsapi.Components, *pdscontrolplane.ControlPlane, error) {
	log.InfoD("Initializing Api components")
	apiConf := pds.NewConfiguration()
	endpointURL, err := url.Parse(ControlPlaneURL)
	if err != nil {
		return nil, nil, err
	}
	apiConf.Host = endpointURL.Host
	apiConf.Scheme = endpointURL.Scheme

	apiClient := pds.NewAPIClient(apiConf)
	components := pdsapi.NewComponents(apiClient)
	controlplane := pdscontrolplane.NewControlPlane(ControlPlaneURL, components)

	return components, controlplane, nil
}

// Get returns a registered scheduler test provider.
func Get(name string) (Driver, error) {
	if d, ok := pdsschedulers[name]; ok {
		return d, nil
	}
	return nil, &errors.ErrNotFound{
		ID:   name,
		Type: "PdsDriver",
	}
}

func Register(name string, d Driver) error {
	if _, ok := pdsschedulers[name]; !ok {
		pdsschedulers[name] = d
	} else {
		return fmt.Errorf("pds driver: %s is already registered", name)
	}
	return nil
}
