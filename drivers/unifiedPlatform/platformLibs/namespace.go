package platformLibs

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

const portorxLabel = "platform.portworx.io/pds"

// CreatePDSNamespace creates a namespace with the given name if it does not exist
func CreatePDSNamespace(namespace string) error {
	k8sCore := core.Instance()
	_, err := k8sCore.GetNamespace(namespace)
	if err != nil {
		log.Warnf("Namespace not found %v", err)
		if strings.Contains(err.Error(), "not found") {
			nsName := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   namespace,
					Labels: map[string]string{portorxLabel: "true"},
				},
			}
			log.InfoD("Creating namespace %v", namespace)
			_, err = k8sCore.CreateNamespace(nsName)
			if err != nil {
				return fmt.Errorf("Error while creating namespace [%s]", err.Error())
			}
		}
	} else {
		log.Infof("Namespace already exists")
	}

	return nil
}

// ValidateLabel validates if the namespace has the label
func ValidateLabel(namespace string) error {
	k8sCore := core.Instance()
	ns, err := k8sCore.GetNamespace(namespace)
	if err != nil {
		return fmt.Errorf("Some error occurred while getting namespace - Error - [%s]", err.Error())
	}
	isAvailable := false
	for key, value := range ns.Labels {
		log.Infof("key: %v values: %v", key, value)
		if key == portorxLabel && value == "true" {
			log.InfoD("key: %v values: %v", key, value)
			isAvailable = true
			break
		}
	}
	if !isAvailable {
		return fmt.Errorf("Namespace does not have the label [%s]", portorxLabel)
	}

	return nil
}

func DeleteNamespace(namespace string) error {
	k8sCore := core.Instance()
	err := k8sCore.DeleteNamespace(namespace)
	if err != nil {
		return fmt.Errorf("Error while deleting namespace [%s]", err.Error())
	}
	return nil
}

func ListNamespaces(tenantId string, clusterId string, projectId string, label string, sortBy string, sortOrder string) (*automationModels.PlatformNamespaceResponse, error) {
	request := &automationModels.PlatformNamespace{
		List: automationModels.PlatformListNamespace{
			TenantId:      tenantId,
			ClusterId:     clusterId,
			ProjectId:     projectId,
			Label:         label,
			SortSortBy:    sortBy,
			SortSortOrder: sortOrder,
		},
	}

	listResponse, err := v2Components.Platform.ListNamespaces(request)
	if err != nil {
		return listResponse, err
	}

	return listResponse, nil
}
