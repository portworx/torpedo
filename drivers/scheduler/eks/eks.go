package eks

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/scheduler"
	kube "github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/pkg/log"
	"os"
	"time"
)

const (
	// SchedName is the name of the kubernetes scheduler driver implementation
	SchedName = "eks"
	// defaultEKSUpgradeTimeout is the default timeout for EKS control plane upgrade
	defaultEKSUpgradeTimeout = 90 * time.Minute
	// defaultEKSUpgradeRetryInterval is the default retry interval for EKS control plane upgrade
	defaultEKSUpgradeRetryInterval = 5 * time.Minute
)

type EKS struct {
	kube.K8s
	clusterName     string
	region          string
	config          aws.Config
	eksClient       *eks.Client
	pxNodeGroupName string
}

// String returns the string name of this driver.
func (e *EKS) String() string {
	return SchedName
}

func (e *EKS) Init(schedOpts scheduler.InitOptions) (err error) {
	e.clusterName = os.Getenv("EKS_CLUSTER_NAME")
	if e.clusterName == "" {
		return fmt.Errorf("env EKS_CLUSTER_NAME not found")
	}
	e.region = os.Getenv("EKS_CLUSTER_REGION")
	if e.region == "" {
		return fmt.Errorf("env EKS_CLUSTER_REGION not found")
	}
	e.pxNodeGroupName = os.Getenv("EKS_PX_NODEGROUP_NAME")
	if e.pxNodeGroupName == "" {
		return fmt.Errorf("env EKS_PX_NODEGROUP_NAME not found")
	}
	e.config, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(e.region))
	if err != nil {
		return fmt.Errorf("unable to load default SDK config. Err: [%v]", err)
	}
	e.eksClient = eks.NewFromConfig(e.config)
	return nil
}

// GetCurrentVersion returns the current version of the EKS cluster
func (e *EKS) GetCurrentVersion() (string, error) {
	eksDescribeClusterOutput, err := e.eksClient.DescribeCluster(
		context.TODO(),
		&eks.DescribeClusterInput{
			Name: aws.String(e.clusterName),
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to describe EKS cluster [%s], Err: [%v]", e.clusterName, err)
	}
	if eksDescribeClusterOutput.Cluster == nil {
		return "", fmt.Errorf("failed to describe EKS cluster [%s], cluster not found", e.clusterName)
	}
	return aws.ToString(eksDescribeClusterOutput.Cluster.Version), nil
}

// UpgradeControlPlane upgrades the EKS control plane to the specified version
func (e *EKS) UpgradeControlPlane(version string) error {
	log.Infof("Upgrade EKS Control Plane to version [%s]", version)
	_, err := e.eksClient.UpdateClusterVersion(
		context.TODO(),
		&eks.UpdateClusterVersionInput{
			Name:    aws.String(e.clusterName),
			Version: aws.String(version),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to set cluser [%s] version to [%s], Err: [%v]", e.clusterName, version, err)
	}
	log.Infof("Initiated EKS Control Place upgrade to [%s] successfully", version)
	return nil
}

// WaitForControlPlaneToUpgrade waits for the EKS control plane to be upgraded to the specified version
func (e *EKS) WaitForControlPlaneToUpgrade(version string) error {
	log.Infof("Waiting for EKS Control Plane to be upgraded to [%s]", version)
	expectedUpgradeStatus := types.ClusterStatusActive
	t := func() (interface{}, bool, error) {
		eksDescribeClusterOutput, err := e.eksClient.DescribeCluster(
			context.TODO(),
			&eks.DescribeClusterInput{
				Name: aws.String(e.clusterName),
			},
		)
		if err != nil {
			return nil, false, err
		}
		if eksDescribeClusterOutput.Cluster == nil {
			return "", false, fmt.Errorf("failed to describe EKS cluster [%s], cluster not found", e.clusterName)
		}
		status := eksDescribeClusterOutput.Cluster.Status
		if status != expectedUpgradeStatus {
			return nil, true, fmt.Errorf("waiting for EKS Control Plane upgrade to [%s] to complete, expected status [%s], actual status [%s]", version, expectedUpgradeStatus, status)
		}
		log.Infof("Upgrade status for EKS Control Plane to [%s] is [%s]", version, status)
		return nil, false, nil
	}
	_, err := task.DoRetryWithTimeout(t, defaultEKSUpgradeTimeout, defaultEKSUpgradeRetryInterval)
	if err != nil {
		return fmt.Errorf("failed to upgrade EKS Control Plane to [%s], Err: [%v]", version, err)
	}
	log.Infof("Successfully upgraded EKS Control Plane to [%s]", version)
	return nil
}

// UpgradeNodeGroup upgrades the EKS node group to the specified version
func (e *EKS) UpgradeNodeGroup(nodeGroupName string, version string) error {
	log.Infof("Starting EKS Node Group upgrade [%s] to [%s]", nodeGroupName, version)
	_, err := e.eksClient.UpdateNodegroupVersion(context.TODO(), &eks.UpdateNodegroupVersionInput{
		ClusterName:   aws.String(e.clusterName),
		NodegroupName: aws.String(nodeGroupName),
		Version:       aws.String(version),
	})
	if err != nil {
		return fmt.Errorf("failed to upgrade EKS Node Group [%s] version to [%s], Err: [%v]", nodeGroupName, version, err)
	}
	log.Infof("Initiated EKS Node Group [%s] upgrade to version [%s] successfully", nodeGroupName, version)
	return nil
}

// WaitForNodeGroupToUpgrade waits for the EKS node group to be upgraded to the specified version
func (e *EKS) WaitForNodeGroupToUpgrade(nodeGroupName string, version string) error {
	log.Infof("Waiting for EKS Node Group [%s] to be upgraded to [%s]", nodeGroupName, version)
	expectedUpgradeStatus := types.NodegroupStatusActive
	t := func() (interface{}, bool, error) {
		eksDescribeNodegroupOutput, err := e.eksClient.DescribeNodegroup(
			context.TODO(),
			&eks.DescribeNodegroupInput{
				ClusterName:   aws.String(e.clusterName),
				NodegroupName: aws.String(nodeGroupName),
			},
		)
		if err != nil {
			return nil, false, err
		}
		status := eksDescribeNodegroupOutput.Nodegroup.Status
		if status != expectedUpgradeStatus {
			return nil, true, fmt.Errorf("waiting for EKS Node Group [%s] upgrade to [%s] to complete, expected status [%s], actual status [%s]", nodeGroupName, version, expectedUpgradeStatus, status)
		}
		log.Infof("Upgrade status for EKS Node Group [%s] to [%s] is [%s]", nodeGroupName, version, status)
		return nil, false, nil
	}
	_, err := task.DoRetryWithTimeout(t, defaultEKSUpgradeTimeout, defaultEKSUpgradeRetryInterval)
	if err != nil {
		return fmt.Errorf("failed to upgrade EKS Node Group [%s] version to [%s], Err: [%v]", nodeGroupName, version, err)
	}
	log.Infof("Successfully upgraded EKS Node Group [%s] to [%s]", nodeGroupName, version)
	return nil
}

// UpgradeScheduler upgrades the EKS cluster to the specified version
func (e *EKS) UpgradeScheduler(version string) error {
	currentVersion, err := e.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get current EKS cluster version, Err: [%v]", err)
	}
	log.Infof("Starting EKS cluster upgrade from [%s] to [%s]", currentVersion, version)

	// Upgrade Control Plane
	err = e.UpgradeControlPlane(version)
	if err != nil {
		return fmt.Errorf("failed to set EKS cluster version, Err: [%v]", err)
	}

	// Wait for control plane to be upgraded
	err = e.WaitForControlPlaneToUpgrade(version)
	if err != nil {
		return fmt.Errorf("failed to wait for EKS control plane to be upgraded to [%s], Err: %v", version, err)
	}

	// Upgrade Node Group
	err = e.UpgradeNodeGroup(e.pxNodeGroupName, version)
	if err != nil {
		return fmt.Errorf("failed to upgrade EKS node group [%s] to [%s], Err: %v", e.pxNodeGroupName, version, err)
	}

	// Wait for the portworx node group to be upgraded
	err = e.WaitForNodeGroupToUpgrade(e.pxNodeGroupName, version)
	if err != nil {
		return fmt.Errorf("failed to wait for EKS node group [%s] to be upgraded to [%s], Err: %v", e.pxNodeGroupName, version, err)
	}
	log.Infof("Successfully finished EKS cluster [%s] upgrade from [%s] to [%s]", e.clusterName, currentVersion, version)
	return nil
}

func init() {
	e := &EKS{}
	scheduler.Register(SchedName, e)
}
