package stworkflows

import (
	"fmt"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	"github.com/portworx/torpedo/pkg/log"
	"time"
)

type WorkflowNamespace struct {
	Namespaces    map[string]string
	TargetCluster WorkflowTargetCluster
}

const (
	retryTimeout  = 10 * time.Minute
	retryInterval = 30 * time.Second
)

func (workflowNamespace *WorkflowNamespace) CreateNamespaces(namespace string) (*WorkflowNamespace, error) {
	err := platformLibs.CreatePDSNamespace(namespace)
	if err != nil {
		return workflowNamespace, err
	}
	err = platformLibs.ValidateLabel(namespace)
	if err != nil {
		return workflowNamespace, err
	}

	uid, err := workflowNamespace.GetNamespaceUID(namespace)
	if err != nil {
		return workflowNamespace, err
	}

	workflowNamespace.Namespaces[namespace] = uid

	log.Infof("Namespace Name - [%s], UID - [%s]", workflowNamespace.Namespaces[namespace], workflowNamespace.Namespaces[namespace])

	return workflowNamespace, nil
}

func (workflowNamespace *WorkflowNamespace) DeleteNamespace(namespace string) error {
	err := platformLibs.DeleteNamespace(namespace)
	if err != nil {
		return err
	}

	// TODO: This needs to be enabled once deletion sync is fixed from platform side
	//err = workflowNamespace.ValidateNamespaceDeletion(namespace)
	//if err != nil {
	//	return err
	//}

	delete(workflowNamespace.Namespaces, namespace)

	return nil
}

func (workflowNamespace *WorkflowNamespace) ValidateNamespaceDeletion(namespaceName string) error {
	checkForNs := func() (interface{}, bool, error) {
		_, err := workflowNamespace.GetNamespaceUID(namespaceName)
		if err != nil {
			return nil, false, nil
		}
		return nil, true, fmt.Errorf("Namespace [%s] still exists", namespaceName)
	}

	_, err := task.DoRetryWithTimeout(checkForNs, retryTimeout, retryInterval)

	return err
}

func (workflowNamespace *WorkflowNamespace) ListNamespaces(tenantId string, label string, sortBy string, sortOrder string) (*automationModels.PlatformNamespaceResponse, error) {
	allNamespaces, err := platformLibs.ListNamespaces(
		tenantId,
		label,
		sortBy,
		sortOrder,
	)

	if err != nil {
		return allNamespaces, err
	}
	return allNamespaces, nil
}

func (workflowNamespace *WorkflowNamespace) GetNamespaceUID(namespace string) (string, error) {
	waitForNSToReflect := func() (interface{}, bool, error) {
		allNamespaces, err := workflowNamespace.ListNamespaces(
			workflowNamespace.TargetCluster.Platform.TenantId, "", "", "")

		if err != nil {
			return "", true, err
		}

		for _, eachNamespace := range allNamespaces.List.Namespaces {
			log.Infof("Namespace - [%s]", *eachNamespace.Meta.Name)
			if *eachNamespace.Meta.Name == namespace {
				return *eachNamespace.Meta.Uid, false, nil
			}
		}

		return "", true, fmt.Errorf("Namespace [%s] not found in the list of namespaces", namespace)
	}

	ns, err := task.DoRetryWithTimeout(waitForNSToReflect, retryTimeout, retryInterval)
	if err != nil {
		return "", fmt.Errorf("Namespace [%s] not found in the list of namespaces", namespace)
	}

	return ns.(string), nil
}
