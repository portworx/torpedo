package targetcluster

import (
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	pdsapi "github.com/portworx/torpedo/drivers/pds/api"
	"github.com/portworx/torpedo/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"net/url"
	"os"
	"strings"
	"time"
)

type K8sType struct{}

// PDS vars
var (
	components         *pdsapi.Components
	deployment         *pds.ModelsDeployment
	apiClient          *pds.APIClient
	ns                 *corev1.Namespace
	err                error
	isavailable        bool
	namespaceNameIDMap = make(map[string]string)
)

// PDS const
const (
	timeOut      = 30 * time.Minute
	timeInterval = 10 * time.Second
	pxLabel      = "pds.portworx.com/available"
)

// GetAndExpectStringEnvVar parses a string from env variable.
func (k8s *K8sType) GetAndExpectStringEnvVar(varName string) string {
	varValue := os.Getenv(varName)
	return varValue
}

// CreatePDSNamespace checks if the namespace is available in the cluster and pds is enabled on it
func (k8s *K8sType) CreatePDSNamespace(namespace string) (*corev1.Namespace, bool, error) {
	ns, err = k8sCore.GetNamespace(namespace)
	isavailable = false
	if err != nil {
		log.Warnf("Namespace not found %v", err)
		if strings.Contains(err.Error(), "not found") {
			nsName := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   namespace,
					Labels: map[string]string{pxLabel: "true"},
				},
			}
			log.InfoD("Creating namespace %v", namespace)
			ns, err = k8sCore.CreateNamespace(nsName)
			if err != nil {
				log.Errorf("Error while creating namespace %v", err)
				return nil, false, err
			}
			isavailable = true
		}
		if !isavailable {
			return nil, false, err
		}
	}
	isavailable = false
	for key, value := range ns.Labels {
		log.Infof("key: %v values: %v", key, value)
		if key == pxLabel && value == "true" {
			log.InfoD("key: %v values: %v", key, value)
			isavailable = true
			break
		}
	}
	if !isavailable {
		return nil, false, nil
	}
	return ns, true, nil
}

// GetnameSpaceID returns the namespace ID
func (k8s *K8sType) GetnameSpaceID(namespace string, deploymentTargetID string) (string, error) {
	var namespaceID string

	err = wait.Poll(timeInterval, timeOut, func() (bool, error) {
		namespaces, err := components.Namespace.ListNamespaces(deploymentTargetID)
		for i := 0; i < len(namespaces); i++ {
			if namespaces[i].GetName() == namespace {
				if namespaces[i].GetStatus() == "available" {
					namespaceID = namespaces[i].GetId()
					namespaceNameIDMap[namespaces[i].GetName()] = namespaces[i].GetId()
					log.InfoD("Namespace Status - Name: %v , Id: %v , Status: %v", namespaces[i].GetName(), namespaces[i].GetId(), namespaces[i].GetStatus())
					return true, nil
				}
			}
		}
		if err != nil {
			log.Errorf("An Error Occured while listing namespaces %v", err)
			return false, err
		}
		return false, nil
	})
	return namespaceID, nil
}

func K8sInit(ControlPlaneURL string) (*K8sType, error) {
	apiConf := pds.NewConfiguration()
	endpointURL, err := url.Parse(ControlPlaneURL)
	if err != nil {
		return nil, err
	}
	apiConf.Host = endpointURL.Host
	apiConf.Scheme = endpointURL.Scheme

	apiClient = pds.NewAPIClient(apiConf)
	components = pdsapi.NewComponents(apiClient)

	return &K8sType{}, nil
}
