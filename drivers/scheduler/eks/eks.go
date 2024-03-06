package eks

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/scheduler"
	kube "github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/pkg/log"
	"os"
	"strings"
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
	// This implementation assumes the EKS cluster has two node groups: one group for
	// Torpedo and another group for Portworx.
	err = e.K8s.Init(schedOpts)
	if err != nil {
		return err
	}
	torpedoNodeName := ""
	pods, err := core.Instance().GetPods("default", nil)
	if err != nil {
		log.Errorf("failed to get pods from default namespace. Err: [%v]", err)
	}
	if pods != nil {
		for _, pod := range pods.Items {
			if pod.Name == "torpedo" {
				torpedoNodeName = pod.Spec.NodeName
			}
		}
	}
	nodes, err := core.Instance().GetNodes()
	if err != nil {
		log.Errorf("failed to get nodes. Err: [%v]", err)
	}
	e.region = os.Getenv("EKS_CLUSTER_REGION")
	if e.region == "" {
		nodeRegionLabel := "topology.kubernetes.io/region"
		log.Warnf("env EKS_CLUSTER_REGION not found. Using node label [%s] to determine region", nodeRegionLabel)
		if torpedoNodeName != "" && nodes != nil  {
			for _, node := range nodes.Items {
				if node.Name != torpedoNodeName {
					e.region = node.Labels[nodeRegionLabel]
					log.Infof("Using node label [%s] to determine region [%s]", nodeRegionLabel, e.region)
					break
				}
			}
		}
		if e.region == "" {
			return fmt.Errorf("env EKS_CLUSTER_REGION or node label [%s] not found", nodeRegionLabel)
		}
	}
	e.pxNodeGroupName = os.Getenv("EKS_PX_NODEGROUP_NAME")
	if e.pxNodeGroupName == "" {
		nodeGroupLabel := "eks.amazonaws.com/nodegroup"
		log.Warnf("env EKS_PX_NODEGROUP_NAME not found. Using node label [%s] to determine Portworx node group", nodeGroupLabel)
		if torpedoNodeName != "" && nodes != nil  {
			for _, node := range nodes.Items {
				if node.Name != torpedoNodeName {
					e.pxNodeGroupName = node.Labels[nodeGroupLabel]
					log.Infof("Using node label [%s] to determine Portworx node group [%s]", nodeGroupLabel, e.pxNodeGroupName)
					break
				}
			}
		}
		if e.pxNodeGroupName == "" {
			return fmt.Errorf("env EKS_PX_NODEGROUP_NAME or node label [%s] not found", nodeGroupLabel)
		}
	}
	e.clusterName = os.Getenv("EKS_CLUSTER_NAME")
	if e.clusterName == "" {
		ec2InstanceLabel := "kubernetes.io/cluster/"
		for _, node := range nodes.Items {
			providerID := node.Spec.ProviderID
			// In EKS, nodes have a ProviderID formatted as aws:///<region>/<instance-id>
			splitID := strings.Split(providerID, "/")
			if len(splitID) < 5 {
				return fmt.Errorf("unexpected format of providerID: %s", providerID)
			}
			instanceID := splitID[4]
			e.config, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(e.region))
			if err != nil {
				return fmt.Errorf("unable to load config for region %s, %v", e.region, err)
			}
			ec2Client := ec2.NewFromConfig(e.config)
			result, err := ec2Client.DescribeInstances(
				context.TODO(),
				&ec2.DescribeInstancesInput{
					InstanceIds: []string{instanceID},
				},
			)
			if err != nil {
				return fmt.Errorf("failed to describe instance %s, %v", instanceID, err)
			}
			for _, reservation := range result.Reservations {
				for _, instance := range reservation.Instances {
					for _, tag := range instance.Tags {
						if strings.HasPrefix(*tag.Key, ec2InstanceLabel) {
							e.clusterName = strings.TrimPrefix(*tag.Key, ec2InstanceLabel)
							log.Infof("Instance [%s] is part of EKS cluster [%s] in region [%s]", instanceID, e.clusterName, e.region)
							break
						}
					}
				}
			}
		}
		if e.clusterName == "" {
			return fmt.Errorf("env EKS_CLUSTER_NAME or EC2 instance label [%s] not found", ec2InstanceLabel)
		}
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
			return nil, false, fmt.Errorf("failed to describe EKS cluster [%s], cluster not found", e.clusterName)
		}
		status := eksDescribeClusterOutput.Cluster.Status
		currentVersion := aws.ToString(eksDescribeClusterOutput.Cluster.Version)
		if status == expectedUpgradeStatus && currentVersion == version {
			log.Infof("EKS Control Plane upgrade to [%s] completed successfully. Current status: [%s], version: [%s].", version, status, currentVersion)
			return nil, false, nil
		} else {
			return nil, true, fmt.Errorf("waiting for EKS Control Plane upgrade to [%s] to complete, expected status [%s], actual status [%s], current version [%s]", version, expectedUpgradeStatus, status, currentVersion)
		}
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
		if eksDescribeNodegroupOutput.Nodegroup == nil {
			return nil, false, fmt.Errorf("failed to describe EKS Node Group [%s], node group not found", nodeGroupName)
		}
		status := eksDescribeNodegroupOutput.Nodegroup.Status
		releaseVersion := aws.ToString(eksDescribeNodegroupOutput.Nodegroup.ReleaseVersion)
		// The release version comparison using strings.HasPrefix is necessary because
		// EKS appends a suffix to the version (e.g., "1.27.9-20240213").
		if status == expectedUpgradeStatus && strings.HasPrefix(releaseVersion, version) {
			log.Infof("EKS Node Group [%s] successfully upgraded to version [%s]. Current status: [%s], release version: [%s].", nodeGroupName, version, status, releaseVersion)
			return nil, false, nil
		} else {
			return nil, true, fmt.Errorf("waiting for EKS Node Group [%s] upgrade to [%s] to complete, expected status [%s], actual status [%s], current release version [%s]", nodeGroupName, version, expectedUpgradeStatus, status, releaseVersion)
		}
	}
	_, err := task.DoRetryWithTimeout(t, defaultEKSUpgradeTimeout, defaultEKSUpgradeRetryInterval)
	if err != nil {
		return fmt.Errorf("failed to upgrade EKS Node Group [%s] to version [%s], Err: [%v]", nodeGroupName, version, err)
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
