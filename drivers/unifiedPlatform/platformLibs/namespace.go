package platformLibs

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/pkg/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
)

const portorxLabel = "platform.portworx.io/pds"

// CreatePDSNamespace creates a namespace with the given name if it does not exist
func CreatePDSNamespace(namespace string) error {
	k8sCore := core.Instance()
	_, err := k8sCore.GetNamespace(namespace)
	if err != nil {
		log.Infof("Namespace not found %v", err)
		if strings.Contains(err.Error(), "not found") {
			nsName := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name:   namespace,
					Labels: map[string]string{portorxLabel: "true"},
				},
			}
			log.Infof("Creating namespace %v", namespace)
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
			log.Infof("key: %v values: %v", key, value)
			isAvailable = true
			break
		}
	}
	if !isAvailable {
		return fmt.Errorf("Namespace does not have the label [%s]", portorxLabel)
	}

	return nil
}

func GetNamespace(tenantId, namespaceName string) (automationModels.V1Namespace, error) {
	allNamespaces, err := ListNamespaces(tenantId, "", "CREATED_AT", "DESC")
	if err != nil {
		return automationModels.V1Namespace{}, err
	}

	for _, eachNamespace := range allNamespaces.List.Namespaces {
		log.Infof("Namespace - [%s]", *eachNamespace.Meta.Name)
		if *eachNamespace.Meta.Name == namespaceName {
			return eachNamespace, nil
		}
	}
	return automationModels.V1Namespace{}, fmt.Errorf("Namespace [%s] not found in the list of namespaces", namespaceName)
}

func ListNamespaces(tenantId string, label string, sortBy string, sortOrder string) (*automationModels.PlatformNamespaceResponse, error) {
	request := &automationModels.PlatformNamespace{
		List: automationModels.PlatformListNamespace{
			TenantId:      tenantId,
			Label:         label,
			SortSortBy:    sortBy,
			SortSortOrder: sortOrder,
		},
	}

	listResponse, err := v2Components.Platform.ListNamespaces(request)
	if err != nil {
		return listResponse, err
	}

	totalPages, err := strconv.Atoi(*listResponse.List.Pagination.TotalPages)
	if err != nil {
		return listResponse, fmt.Errorf("Unable to get total pages")
	}
	totalRecords, err := strconv.Atoi(*listResponse.List.Pagination.TotalRecords)
	if err != nil {
		return listResponse, fmt.Errorf("Unable to get total records")
	}

	log.Infof("Namespaces have [%d] pages and [%d] rescords", totalPages, totalRecords)

	request.List.PaginationPageNumber = "1"
	request.List.PaginationPageSize = *listResponse.List.Pagination.TotalRecords

	listResponse, err = v2Components.Platform.ListNamespaces(request)
	log.Infof("Total records found - [%d]", len(listResponse.List.Namespaces))
	if err != nil {
		return listResponse, err
	}

	return listResponse, nil
}

// DeleteNamespace will delete the namespace from the control plane
func DeleteNamespace(id string) error {

	request := automationModels.PlatformNamespace{
		Delete: automationModels.PlatformNamespaceDelete{
			Id: id,
		},
	}

	err := v2Components.Platform.DeleteNamespace(&request)
	return err
}
