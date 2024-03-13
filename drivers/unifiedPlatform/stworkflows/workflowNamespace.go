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
	Namespaces    map[string]map[string]string
	TargetCluster WorkflowTargetCluster
}

const (
	NamespaceName = "namesace_name"
	NamespaceUID  = "namespace_uid"
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

	uid, err := workflowNamespace.GetNamespaeUID(namespace)
	if err != nil {
		return workflowNamespace, err
	}

	workflowNamespace.Namespaces[namespace] = make(map[string]string)
	workflowNamespace.Namespaces[namespace][NamespaceName] = namespace
	workflowNamespace.Namespaces[namespace][NamespaceUID] = uid

	log.Infof("Namespace Name - [%s], UID - [%s]", workflowNamespace.Namespaces[namespace][NamespaceName], workflowNamespace.Namespaces[namespace][NamespaceUID])

	return workflowNamespace, nil
}

func (workflowNamespace *WorkflowNamespace) DeleteNamespace(namespace string) error {
	err := platformLibs.DeleteNamespace(namespace)
	if err != nil {
		return err
	}

	err = workflowNamespace.ValidateNamespaceDeletion(namespace)
	if err != nil {
		return err
	}

	delete(workflowNamespace.Namespaces, namespace)

	return nil
}

func (workflowNamespace *WorkflowNamespace) ValidateNamespaceDeletion(namespaceName string) error {
	checkForNs := func() (interface{}, bool, error) {
		_, err := workflowNamespace.GetNamespaeUID(namespaceName)
		if err != nil {
			return nil, false, nil
		}
		return nil, true, fmt.Errorf("Namespace [%s] still exists", namespaceName)
	}

	_, err := task.DoRetryWithTimeout(checkForNs, retryTimeout, retryInterval)

	return err
}

func (workflowNamespace *WorkflowNamespace) ListNamespaces(tenantId string, clusterId string, projectId string, label string, sortBy string, sortOrder string) (*automationModels.PlatformNamespaceResponse, error) {
	allNamespaces, err := platformLibs.ListNamespaces(
		tenantId,
		clusterId,
		projectId,
		label,
		sortBy,
		sortOrder,
	)

	if err != nil {
		return allNamespaces, err
	}
	return allNamespaces, nil
}

func (workflowNamespace *WorkflowNamespace) GetNamespaeUID(namespace string) (string, error) {
	allNamespaces, err := workflowNamespace.ListNamespaces(
		workflowNamespace.TargetCluster.Platform.TenantId,
		workflowNamespace.TargetCluster.ClusterUID,
		"", "", "", "")

	if err != nil {
		return "", err
	}

	for _, eachNamespace := range allNamespaces.List.Namespaces {
		log.Infof("Namespace - [%s]", *eachNamespace.Meta.Name)
		if *eachNamespace.Meta.Name == namespace {
			return *eachNamespace.Meta.Uid, nil
		}
	}
	return "", fmt.Errorf("Namespace [%s] not found in the list of namespaces", namespace)
}
