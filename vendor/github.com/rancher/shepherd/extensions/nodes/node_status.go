package nodes

import (
	"context"
	"strings"
	"time"

	"github.com/rancher/norman/types"
	rkev1 "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	activeState              = "active"
	runningState             = "running"
	errorState               = "error"
	machineSteveResourceType = "cluster.x-k8s.io.machine"
	machineSteveAnnotation   = "cluster.x-k8s.io/machine"
	fleetNamespace           = "fleet-default"
	etcdLabel                = "rke.cattle.io/etcd-role"
	clusterLabel             = "cluster.x-k8s.io/cluster-name"

	PollInterval = time.Duration(5 * time.Second)
	PollTimeout  = time.Duration(15 * time.Minute)

	oneSecondInterval = time.Duration(1 * time.Second)
	fiveMinuteTimeout = time.Duration(5 * time.Minute)

	httpNotFound = "404 Not Found"
)

// AllManagementNodeReady is a helper method that will loop and check if the node is ready in the RKE1 cluster.
// It will return an error if the node is not ready after set amount of time.
func AllManagementNodeReady(client *rancher.Client, ClusterID string, timeout time.Duration) error {
	ctx := context.Background()
	err := wait.PollUntilContextTimeout(
		ctx, PollInterval, timeout, true, func(ctx context.Context) (bool, error) {
			nodes, err := client.Management.Node.ListAll(&types.ListOpts{
				Filters: map[string]interface{}{
					"clusterId": ClusterID,
				},
			})
			if err != nil {
				return false, nil
			}

			for _, node := range nodes.Data {
				node, err := client.Management.Node.ByID(node.ID)
				if err != nil {
					return false, nil
				}

				if node.State == errorState {
					logrus.Warnf("node %s is in error state", node.Name)

					return false, nil
				}

				if node.State != activeState {
					return false, nil
				}
			}

			logrus.Infof("All nodes in the cluster are in an active state!")

			return true, nil
		})

	return err
}

// AllMachineReady is a helper method that will loop and check if
// the machine object of every node in a cluster is ready. Typically Used for RKE2/K3s Clusters.
// It will return an error if the machine object is not ready after set amount of time.
func AllMachineReady(client *rancher.Client, clusterID string, timeout time.Duration) error {
	ctx := context.Background()
	err := wait.PollUntilContextTimeout(
		ctx, PollInterval, timeout, true, func(ctx context.Context) (bool, error) {
			nodes, err := client.Management.Node.List(&types.ListOpts{Filters: map[string]interface{}{
				"clusterId": clusterID,
			}})
			if err != nil {
				return false, err
			}

			for _, node := range nodes.Data {
				machine, err := client.Steve.
					SteveType(machineSteveResourceType).
					ByID(fleetNamespace + "/" + node.Annotations[machineSteveAnnotation])
				if err != nil {
					return false, err
				}

				if machine.State == nil {
					logrus.Infof("Machine: %s state is nil", machine.Name)
					return false, nil
				}

				if machine.State.Error {
					logrus.Warnf("Machine: %s is in error state: %s", machine.Name, machine.State.Message)
					return false, nil
				}

				if machine.State.Name != runningState {
					return false, nil
				}
			}

			logrus.Infof("All nodes in the cluster are running!")

			return true, nil
		})
	return err
}

// AllNodeDeleted is a helper method that will loop and check if the node is deleted in the cluster.
func AllNodeDeleted(client *rancher.Client, ClusterID string) error {
	ctx := context.Background()
	err := wait.PollUntilContextTimeout(
		ctx, oneSecondInterval, fiveMinuteTimeout, true, func(ctx context.Context) (bool, error) {
			nodes, err := client.Management.Node.ListAll(&types.ListOpts{
				Filters: map[string]interface{}{
					"clusterId": ClusterID,
				},
			})
			if err != nil {
				return false, err
			}

			if len(nodes.Data) == 0 {
				logrus.Infof("All nodes in the cluster are deleted!")
				return true, nil
			}

			return false, nil
		})

	return err
}

// IsNodeReplaced is a helper method that will loop and check if the node matching its type is replaced in a cluster.
// It will return an error if the node is not replaced after set amount of time.
func IsNodeReplaced(client *rancher.Client, oldMachineID string, clusterID string, numOfNodesBeforeDeletion int) (bool, error) {
	numOfNodesAfterDeletion := 0
	isOldMachineDeleted := true

	ctx := context.Background()
	err := wait.PollUntilContextTimeout(
		ctx, oneSecondInterval, PollTimeout, true, func(ctx context.Context) (bool, error) {
			machines, err := client.Management.Node.ListAll(&types.ListOpts{Filters: map[string]interface{}{
				"clusterId": clusterID,
			}})
			if err != nil {
				return false, err
			}

			numOfNodesAfterDeletion = 0

			for _, machine := range machines.Data {
				if machine.ID == oldMachineID {
					isOldMachineDeleted = false
					return false, nil
				}

				numOfNodesAfterDeletion++
			}

			return isOldMachineDeleted, nil
		})

	logrus.Infof("Node has been successfully replaced!")

	return numOfNodesAfterDeletion == numOfNodesBeforeDeletion, err
}

// Isv1NodeConditionMet checks the condition reasons of a given machine in a cluster and waits for it to be true.
// Otherwise, an error is returned.
func Isv1NodeConditionMet(client *rancher.Client, machineID, clusterID, conditionReason string) error {
	steveclient, err := client.Steve.ProxyDownstream(clusterID)
	if err != nil {
		return err
	}

	v1NodeStatus := &rkev1.RKEMachineStatus{}

	ctx := context.Background()
	err = wait.PollUntilContextTimeout(
		ctx, PollInterval, PollTimeout, true, func(ctx context.Context) (bool, error) {
			refreshedMachine, err := steveclient.SteveType("node").ByID(machineID)
			if err != nil {
				if strings.Contains(err.Error(), httpNotFound) {
					return true, nil
				}

				return false, err
			}

			err = v1.ConvertToK8sType(refreshedMachine.Status, v1NodeStatus)
			if err != nil {
				return false, err
			}

			for _, condition := range v1NodeStatus.Conditions {
				if condition.Reason == conditionReason {
					logrus.Infof("Node is in condition: %s", conditionReason)
					return true, nil
				}
			}

			return false, nil
		})

	return err
}
