package platform

import (
	"fmt"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/platformLibs"
	k8utils "github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	"strings"
	"time"
)

type WorkflowNamespace struct {
	Namespaces    map[string]string
	TargetCluster *WorkflowTargetCluster
}

const (
	retryTimeout            = 20 * time.Minute
	namespaceRemovalTimeout = 5 * time.Minute
	retryInterval           = 10 * time.Second
	tombstoned              = "TOMBSTONED"
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

	uid, err := workflowNamespace.ValidateNamespaceUID(namespace)
	if err != nil {
		return workflowNamespace, err
	}

	workflowNamespace.Namespaces[namespace] = uid

	log.Infof("Namespace Name - [%s], UID - [%s]", namespace, workflowNamespace.Namespaces[namespace])

	return workflowNamespace, nil
}

func (workflowNamespace *WorkflowNamespace) DeleteNamespace(namespace string) error {
	err := k8utils.DeleteNamespace(namespace)

	if err != nil {
		return err
	}
	log.Infof("Delete [%s] from the cluster", namespace)

	id, err := workflowNamespace.ValidateNamespaceUID(namespace)
	if err != nil {
		return err
	}

	log.Infof("Waiting for [%s] to get into TOMBSTONED phase.", namespace)

	// Wait for namespaces to go in 'TOMBSTONED' phase and call delete API
	waitforNamepsaceToBeTombstoned := func() (interface{}, bool, error) {
		allNamespaces, err := workflowNamespace.ListNamespaces(
			workflowNamespace.TargetCluster.Project.Platform.TenantId, "", "CREATED_AT", "DESC")

		if err != nil {
			return nil, false, fmt.Errorf("Some error occurred while polling for namespaces. Error - [%s]", err.Error())
		}

		for _, eachNamespace := range allNamespaces.List.Namespaces {
			if *eachNamespace.Meta.Name == namespace {
				if *eachNamespace.Status.Phase != tombstoned {
					return nil, true, fmt.Errorf("Waiting for [%s] to be %s, Current Phase - [%s]", namespace, tombstoned, *eachNamespace.Status.Phase)
				} else {
					err := platformLibs.DeleteNamespace(id)
					time.Sleep(2 * time.Second) // Explicit delay for delete to remove entry from DB
					if err != nil {
						return nil, false, fmt.Errorf("Some error occurred while deleting [%s]. Error - [%s]", namespace, err.Error())
					} else {
						err := workflowNamespace.ValidateNamespaceDeletion(id)
						if err != nil {
							return nil, false, err
						}
						delete(workflowNamespace.Namespaces, namespace)
						log.Infof("[%s] deleted successfully", namespace)
					}
				}
			}
		}

		return nil, false, nil

	}

	_, err = task.DoRetryWithTimeout(waitforNamepsaceToBeTombstoned, retryTimeout, retryInterval)

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

func (workflowNamespace *WorkflowNamespace) GetNamespace(namespace string) (automationModels.V1Namespace, error) {
	allNamespaces, err := workflowNamespace.ListNamespaces(
		workflowNamespace.TargetCluster.Project.Platform.TenantId, "", "CREATED_AT", "DESC")

	if err != nil {
		return automationModels.V1Namespace{}, err
	}

	for _, eachNamespace := range allNamespaces.List.Namespaces {
		log.Infof("Namespace - [%s]", *eachNamespace.Meta.Name)
		if *eachNamespace.Meta.Name == namespace {
			return eachNamespace, nil
		}
	}

	return automationModels.V1Namespace{}, fmt.Errorf("Namespace [%s] not found in the list of namespaces", namespace)
}

func (workflowNamespace *WorkflowNamespace) ValidateNamespaceUID(namespace string) (string, error) {
	waitForNSToReflect := func() (interface{}, bool, error) {
		allNamespaces, err := workflowNamespace.ListNamespaces(
			workflowNamespace.TargetCluster.Project.Platform.TenantId, "", "CREATED_AT", "DESC")

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

func (workflowNamespace *WorkflowNamespace) Purge(ignoreError bool) error {

	// Delete all namespaces from k8s cluster

	for namespace, _ := range workflowNamespace.Namespaces {
		err := k8utils.DeleteNamespace(namespace)
		if err != nil {
			if ignoreError && strings.Contains(err.Error(), "not found") == true {
				log.Warnf(err.Error())
			} else {
				return fmt.Errorf("Error occurred during deleting namespace from cluster. Error - [%s]", err.Error())
			}
		}
		log.Infof("Deleted [%s] from the cluster", namespace)
	}

	// Wait for all namespaces to go in 'TOMBSTONED' phase and call delete API

	waitforNamepsaceToBeTombstoned := func() (interface{}, bool, error) {
		allNamespaces, err := workflowNamespace.ListNamespaces(
			workflowNamespace.TargetCluster.Project.Platform.TenantId, "", "CREATED_AT", "DESC")

		if err != nil {
			return nil, false, fmt.Errorf("Some error occurred while polling for namespaces. Error - [%s]", err.Error())
		}

		if len(workflowNamespace.Namespaces) > 0 {
			for _, eachNamespace := range allNamespaces.List.Namespaces {
				for namespace, id := range workflowNamespace.Namespaces {
					if *eachNamespace.Meta.Name == namespace {
						if *eachNamespace.Status.Phase != tombstoned {
							return nil, true, fmt.Errorf("Waiting for [%s] to be %s, Current Phase - [%s]", namespace, tombstoned, *eachNamespace.Status.Phase)
						} else {
							err := platformLibs.DeleteNamespace(id)
							time.Sleep(2 * time.Second) // Explicit delay for delete to remove entry from DB
							if err != nil {
								return nil, false, fmt.Errorf("Some error occurred while deleting [%s]. Error - [%s]", namespace, err.Error())
							} else {
								err := workflowNamespace.ValidateNamespaceDeletion(id)
								if err != nil {
									return nil, false, err
								}
								log.Infof("[%s] deleted successfully", namespace)
								delete(workflowNamespace.Namespaces, namespace)
							}
						}
					}
				}
			}
		}

		return nil, false, nil

	}

	_, err := task.DoRetryWithTimeout(waitforNamepsaceToBeTombstoned, retryTimeout, retryInterval)

	return err

}

func (workflowNamespace *WorkflowNamespace) ValidateNamespaceDeletion(id string) error {

	waitforNamepsaceToBeTombstoned := func() (interface{}, bool, error) {
		allNamespaces, err := workflowNamespace.ListNamespaces(
			workflowNamespace.TargetCluster.Project.Platform.TenantId, "", "CREATED_AT", "DESC")

		if err != nil {
			return nil, false, fmt.Errorf("Some error occurred while polling for namespaces. Error - [%s]", err.Error())
		}

		for _, eachNamespace := range allNamespaces.List.Namespaces {
			if *eachNamespace.Meta.Uid == id {
				return nil, true, fmt.Errorf("Namespace [%s] found after deletion, CurrentPhase - [%s]", id, *eachNamespace.Status.Phase)
			}
		}

		return nil, false, nil
	}

	_, err := task.DoRetryWithTimeout(waitforNamepsaceToBeTombstoned, namespaceRemovalTimeout, retryInterval)

	return err
}
