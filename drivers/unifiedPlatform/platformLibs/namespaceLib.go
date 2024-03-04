package platformLibs

import (
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	defaultTestAccount = "demo-milestone-one"
	// DefaultTimeout default timeout
	timeOut      = 30 * time.Minute
	timeInterval = 10 * time.Second
)

var (
	nsInputs           *apiStructs.WorkFlowRequest
	pdsLabel           = "pds.portworx.com/available"
	namespaceNameIDMap = make(map[string]string)
)

// CreatePdsNamespace creates a PDS ns with label enabled
func CreatePdsNamespace() (*corev1.Namespace, bool, error) {
	namespace := "ns-" + strconv.Itoa(rand.Int())
	ns, err := k8sCore.GetNamespace(namespace)
	isAvailable := false
	if err != nil {
		log.Warnf("Namespace not found %v", err)
		if strings.Contains(err.Error(), "not found") {
			nsName := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   namespace,
					Labels: map[string]string{pdsLabel: "true"},
				},
			}
			log.InfoD("Creating namespace %v", namespace)
			ns, err = k8sCore.CreateNamespace(nsName)
			if err != nil {
				log.Errorf("Error while creating namespace %v", err)
				return nil, false, err
			}
			isAvailable = true
		}
		if !isAvailable {
			return nil, false, err
		}
	}
	isAvailable = false
	for key, value := range ns.Labels {
		log.Infof("key: %v values: %v", key, value)
		if key == pdsLabel && value == "true" {
			log.InfoD("key: %v values: %v", key, value)
			isAvailable = true
			break
		}
	}
	if !isAvailable {
		return nil, false, nil
	}
	return ns, true, nil
}

// GetPdsLabeledNamespaceId get the ns ID for given cluster and namespace name
func GetPdsLabeledNamespaceId(tenantId, deploymentTargetID, namespace string) (string, error) {
	var namespaceID string
	err = wait.Poll(timeInterval, timeOut, func() (bool, error) {
		nsInputs.ListNamespacesRequest.Label = pdsLabel
		nsInputs.ListNamespacesRequest.TenantId = tenantId
		namespaces, err := v2Components.Platform.ListNamespaces(nsInputs)
		for i := 0; i < len(namespaces); i++ {
			nsName := namespaces[i].Meta.Name
			nsStatus := namespaces[i].Status.Phase
			if *nsName == namespace {
				if nsStatus == "available" {
					namespaceID = *namespaces[i].Meta.Uid
					namespaceNameIDMap[*nsName] = namespaceID
					log.InfoD("Namespace Status - Name: %v , Id: %v , Status: %v", nsName, namespaceID, nsStatus)
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

// CreateAndFetchNamespaceId main function to create and return pds labeled ns ID
func CreateAndFetchNamespaceId(deploymentTargetId, tenantId string) (string, string, error) {
	nsObj, isCreated, err := CreatePdsNamespace()
	if isCreated {
		nsName := nsObj.Name
		nsId, err := GetPdsLabeledNamespaceId(tenantId, deploymentTargetId, nsName)
		if err != nil {
			return "", "", err
		}
		return nsName, nsId, nil
	} else {
		return "", "", err
	}
}
