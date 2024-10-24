package eks

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	aws_pkg "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	kube "github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/pkg/log"
)

const (
	// SchedName is the name of the EKS scheduler driver implementation
	SchedName = "eks"
	// defaultEKSUpgradeTimeout is the default timeout for EKS control plane upgrade
	defaultEKSUpgradeTimeout = 90 * time.Minute
	// defaultEKSUpgradeRetryInterval is the default retry interval for EKS control plane upgrade
	defaultEKSUpgradeRetryInterval = 5 * time.Minute
)

type EKS struct {
	kube.K8s
	clusterName       string
	region            string
	config            aws.Config
	eksClient         *eks.Client
	ec2Client         *ec2.Client
	svcSsm            *ssm.SSM
	pxNodeGroupName   string
	isConfigured      bool
	instances         []ec2types.Instance
	autoscalingClient *autoscaling.Client
}

// String returns the string name of this driver.
func (e *EKS) String() string {
	return SchedName
}

func (e *EKS) Init(schedOpts scheduler.InitOptions) (err error) {
	err = e.K8s.Init(schedOpts)
	if err != nil {
		return err
	}
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
	log.Infof("Upgrading EKS cluster [%s] control plane to version [%s]", e.clusterName, version)
	_, err := e.eksClient.UpdateClusterVersion(
		context.TODO(),
		&eks.UpdateClusterVersionInput{
			Name:    aws.String(e.clusterName),
			Version: aws.String(version),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to set EKS cluster [%s] control plane version to [%s], Err: [%v]", e.clusterName, version, err)
	}
	log.Infof("Initiated EKS cluster [%s] control plane upgrade to [%s] successfully", e.clusterName, version)
	return nil
}

// WaitForControlPlaneToUpgrade waits for the EKS control plane to be upgraded to the specified version
func (e *EKS) WaitForControlPlaneToUpgrade(version string) error {
	log.Infof("Waiting for EKS cluster [%s] control plane to be upgraded to [%s]", e.clusterName, version)
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
			log.Infof("EKS cluster [%s] control plane upgrade to [%s] completed successfully. Current status: [%s], version: [%s].", e.clusterName, version, status, currentVersion)
			return nil, false, nil
		} else {
			return nil, true, fmt.Errorf("waiting for EKS cluster [%s] control plane upgrade to [%s] to complete, expected status [%s], actual status [%s], current version [%s]", e.clusterName, version, expectedUpgradeStatus, status, currentVersion)
		}
	}
	_, err := task.DoRetryWithTimeout(t, defaultEKSUpgradeTimeout, defaultEKSUpgradeRetryInterval)
	if err != nil {
		return fmt.Errorf("failed to upgrade EKS cluster [%s] control plane to [%s], Err: [%v]", e.clusterName, version, err)
	}
	log.Infof("Successfully upgraded EKS cluster [%s] control plane to [%s]", e.clusterName, version)
	return nil
}

// UpgradeNodeGroup upgrades the EKS node group to the specified version
func (e *EKS) UpgradeNodeGroup(nodeGroupName string, version string) error {
	log.Infof("Starting EKS cluster [%s] node group upgrade [%s] to [%s]", e.clusterName, nodeGroupName, version)
	_, err := e.eksClient.UpdateNodegroupVersion(context.TODO(), &eks.UpdateNodegroupVersionInput{
		ClusterName:   aws.String(e.clusterName),
		NodegroupName: aws.String(nodeGroupName),
		Version:       aws.String(version),
	})
	if err != nil {
		return fmt.Errorf("failed to upgrade EKS cluster [%s] node group [%s] version to [%s], Err: [%v]", e.clusterName, nodeGroupName, version, err)
	}
	log.Infof("Initiated EKS cluster [%s] node group [%s] upgrade to version [%s] successfully", e.clusterName, nodeGroupName, version)
	return nil
}

// SetMaxUnavailableForUpgradeInNodeGroup updates Max Unavailable values for upgrades for a given node group
func (e *EKS) SetMaxUnavailableForUpgradeInNodeGroup(nodeGroupName string) error {
	updateNodeGroup := false

	log.Info("Checking if we need to configure EKS Node Group Max Unavailable values for surge upgrade..")
	// Get Max Unavailable values from ENV vars, if present
	maxUnavailableNodes := os.Getenv("EKS_MAX_UNAVAILABLE_NODES_VALUE")
	maxUnavailablePercentage := os.Getenv("EKS_MAX_UNAVAILABLE_PERCENTAGE_VALUE")

	// MaxUnavailable:           aws.Int32(1),  // Set the maximum number of nodes that can be unavailable during the update
	// MaxUnavailablePercentage: aws.Int32(25), // Set the maximum percentage of nodes that can be unavailable during the update
	updateConfig := &types.NodegroupUpdateConfig{}

	if maxUnavailableNodes != "" {
		log.Infof("Setting MaxUnavailable to [%s]", maxUnavailableNodes)
		maxUnavailableNodesInt, err := strconv.Atoi(maxUnavailableNodes)
		if err != nil {
			return fmt.Errorf("failed to convert maxUnavailableNode string [%s] to int, Err: %v", maxUnavailableNodes, err)
		}
		updateConfig.MaxUnavailable = aws.Int32(int32(maxUnavailableNodesInt))
		updateNodeGroup = true
	}

	if maxUnavailablePercentage != "" {
		log.Infof("Setting MaxUnavailablePercentage to [%s]", maxUnavailablePercentage)
		maxUnavailablePercentageInt, err := strconv.Atoi(maxUnavailablePercentage)
		if err != nil {
			return fmt.Errorf("failed to convert maxUnavailablePercentage string [%s] to int, Err: %v", maxUnavailablePercentage, err)
		}
		updateConfig.MaxUnavailablePercentage = aws.Int32(int32(maxUnavailablePercentageInt))
		updateNodeGroup = true
	}

	if !updateNodeGroup {
		log.Info("Skipping updating Node Group Max Unavailable values as none were passed..")
		return nil
	}

	log.Infof("Updating Node Group [%s] with Max Unavailable values for surge upgrade for EKS cluster [%s]", nodeGroupName, e.clusterName)
	_, err := e.eksClient.UpdateNodegroupConfig(context.TODO(), &eks.UpdateNodegroupConfigInput{
		ClusterName:   aws.String(e.clusterName),
		NodegroupName: aws.String(nodeGroupName),
		UpdateConfig:  updateConfig,
	})
	if err != nil {
		return fmt.Errorf("failed to configure Max Unavailable for EKS cluster [%s] node group [%s], Err: [%v]", e.clusterName, nodeGroupName, err)
	}

	if err := e.WaitForNodeGroupUpdate(nodeGroupName); err != nil {
		return err
	}

	log.Infof("Successfully configured EKS cluster [%s] node group [%s] with Max Unavailable values for surge upgrade", e.clusterName, nodeGroupName)
	return nil
}

// WaitForNodeGroupUpdate waits for Node Group to be updated and have expected status
func (e *EKS) WaitForNodeGroupUpdate(nodeGroupName string) error {
	log.Infof("Waiting for EKS cluster [%s] Node Group [%s] to be updated", e.clusterName, nodeGroupName)
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
			return nil, false, fmt.Errorf("failed to describe EKS cluster [%s] Node Group [%s], node group not found", e.clusterName, nodeGroupName)
		}
		status := eksDescribeNodegroupOutput.Nodegroup.Status
		if status == expectedUpgradeStatus {
			return nil, false, nil
		} else {
			return nil, true, fmt.Errorf("waiting for EKS cluster [%s] Node Group [%s] update to complete, expected status [%s], actual status [%s]", e.clusterName, nodeGroupName, expectedUpgradeStatus, status)
		}
	}
	_, err := task.DoRetryWithTimeout(t, defaultEKSUpgradeTimeout, defaultEKSUpgradeRetryInterval)
	if err != nil {
		return fmt.Errorf("failed to update EKS cluster [%s] Node Group [%s], Err: [%v]", e.clusterName, nodeGroupName, err)
	}

	log.Infof("Successfully updated EKS cluster [%s] Node Group [%s]", e.clusterName, nodeGroupName)
	return nil
}

// WaitForNodeGroupToUpgrade waits for the EKS node group to be upgraded to the specified version
func (e *EKS) WaitForNodeGroupToUpgrade(nodeGroupName string, version string) error {
	log.Infof("Waiting for EKS cluster [%s] node group [%s] to be upgraded to [%s]", e.clusterName, nodeGroupName, version)
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
			return nil, false, fmt.Errorf("failed to describe EKS cluster [%s] node group [%s], node group not found", e.clusterName, nodeGroupName)
		}
		status := eksDescribeNodegroupOutput.Nodegroup.Status
		releaseVersion := aws.ToString(eksDescribeNodegroupOutput.Nodegroup.ReleaseVersion)
		// The release version comparison using strings.HasPrefix is necessary because
		// EKS appends a suffix to the version (e.g., "1.27.9-20240213").
		if status == expectedUpgradeStatus && strings.HasPrefix(releaseVersion, version) {
			log.Infof("EKS cluster [%s] node group [%s] successfully upgraded to version [%s]. Current status: [%s], release version: [%s].", e.clusterName, nodeGroupName, version, status, releaseVersion)
			return nil, false, nil
		} else {
			return nil, true, fmt.Errorf("waiting for EKS cluster [%s] node group [%s] upgrade to [%s] to complete, expected status [%s], actual status [%s], current release version [%s]", e.clusterName, nodeGroupName, version, expectedUpgradeStatus, status, releaseVersion)
		}
	}
	_, err := task.DoRetryWithTimeout(t, defaultEKSUpgradeTimeout, defaultEKSUpgradeRetryInterval)
	if err != nil {
		return fmt.Errorf("failed to upgrade EKS cluster [%s] node group [%s] to version [%s], Err: [%v]", e.clusterName, nodeGroupName, version, err)
	}
	log.Infof("Successfully upgraded EKS cluster [%s] node group [%s] to [%s]", e.clusterName, nodeGroupName, version)
	return nil
}

// configureEKSClient configures the EKS client
func (e *EKS) configureEKSClient() error {
	// This implementation assumes the EKS cluster has two node groups: one group for
	// Torpedo and another group for Portworx.
	if !e.isConfigured {
		log.Infof("Configuring EKS client")
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
		torpedoNodeGroupName := ""
		nodeGroupLabel := "eks.amazonaws.com/nodegroup"
		nodes, err := core.Instance().GetNodes()
		if err != nil {
			log.Errorf("failed to get nodes. Err: [%v]", err)
		}
		if nodes != nil {
			for _, n := range nodes.Items {
				if n.Name == torpedoNodeName {
					torpedoNodeGroupName = n.Labels[nodeGroupLabel]
					break
				}
			}
		}
		e.pxNodeGroupName = os.Getenv("EKS_PX_NODEGROUP_NAME")
		if e.pxNodeGroupName == "" {
			log.Warnf("env EKS_PX_NODEGROUP_NAME not set. Using node label [%s] to determine Portworx node group", nodeGroupLabel)
			if torpedoNodeGroupName != "" && nodes != nil {
				for _, n := range nodes.Items {
					if n.Labels[nodeGroupLabel] != torpedoNodeGroupName {
						e.pxNodeGroupName = n.Labels[nodeGroupLabel]
						log.Infof("Used node label [%s] to determine Portworx node group [%s]", nodeGroupLabel, e.pxNodeGroupName)
						break
					}
				}
			}
			if e.pxNodeGroupName == "" {
				return fmt.Errorf("env EKS_PX_NODEGROUP_NAME or node label [%s] not set", nodeGroupLabel)
			}
		}
		e.region = os.Getenv("EKS_CLUSTER_REGION")
		if e.region == "" {
			nodeRegionLabel := "topology.kubernetes.io/region"
			log.Warnf("env EKS_CLUSTER_REGION not set. Using node label [%s] to determine region", nodeRegionLabel)
			if torpedoNodeGroupName != "" && nodes != nil {
				for _, n := range nodes.Items {
					if n.Labels[nodeGroupLabel] != torpedoNodeGroupName {
						e.region = n.Labels[nodeRegionLabel]
						log.Infof("Used node label [%s] to determine region [%s]", nodeRegionLabel, e.region)
						break
					}
				}
			}
			if e.region == "" {
				return fmt.Errorf("env EKS_CLUSTER_REGION or node label [%s] not set", nodeRegionLabel)
			}
		}
		e.config, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(e.region))
		if err != nil {
			return fmt.Errorf("unable to load config for region [%s]. Err: [%v]", e.region, err)
		}
		e.eksClient = eks.NewFromConfig(e.config)
		e.ec2Client = ec2.NewFromConfig(e.config)
		e.autoscalingClient = autoscaling.NewFromConfig(e.config)
		e.clusterName = os.Getenv("EKS_CLUSTER_NAME")
		if e.clusterName == "" {
			ec2InstanceLabel := "kubernetes.io/cluster/"
			if nodes != nil {
				for _, n := range nodes.Items {
					providerID := n.Spec.ProviderID
					// In EKS, nodes have a ProviderID formatted as aws:///<region>/<instance-id>
					splitID := strings.Split(providerID, "/")
					if len(splitID) < 5 {
						return fmt.Errorf("unexpected format of provider ID [%s]", providerID)
					}
					instanceID := splitID[4]
					result, err := e.ec2Client.DescribeInstances(
						context.TODO(),
						&ec2.DescribeInstancesInput{
							InstanceIds: []string{instanceID},
						},
					)
					if err != nil {
						return fmt.Errorf("failed to describe EC2 instance [%s]. Err: [%v]", instanceID, err)
					}
					for _, reservation := range result.Reservations {
						for _, instance := range reservation.Instances {
							for _, tag := range instance.Tags {
								if strings.HasPrefix(aws.ToString(tag.Key), ec2InstanceLabel) {
									e.clusterName = strings.TrimPrefix(aws.ToString(tag.Key), ec2InstanceLabel)
									log.Infof("Instance [%s] is part of EKS cluster [%s] in region [%s]", instanceID, e.clusterName, e.region)
									break
								}
							}
						}
					}
				}
			}
			if e.clusterName == "" {
				return fmt.Errorf("env EKS_CLUSTER_NAME or EC2 instance label [%s] not set", ec2InstanceLabel)
			}
			e.isConfigured = true
		}
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))
		e.svcSsm = ssm.New(sess, aws_pkg.NewConfig().WithRegion(e.region))
	} else {
		log.Infof("EKS client already configured")
	}
	return nil
}

// UpgradeScheduler upgrades the EKS cluster to the specified version
func (e *EKS) UpgradeScheduler(version string) error {
	err := e.configureEKSClient()
	if err != nil {
		return fmt.Errorf("failed to configure EKS client. Err: [%v]", err)
	}
	currentVersion, err := e.GetCurrentVersion()
	if err != nil {
		return fmt.Errorf("failed to get EKS cluster [%s] current version, Err: [%v]", e.clusterName, err)
	}
	log.Infof("Starting EKS cluster [%s] upgrade from [%s] to [%s]", e.clusterName, currentVersion, version)

	// Upgrade Control Plane
	err = e.UpgradeControlPlane(version)
	if err != nil {
		return fmt.Errorf("failed to set EKS cluster [%s] control plane version to [%s], Err: [%v]", e.clusterName, version, err)
	}

	// Wait for control plane to be upgraded
	err = e.WaitForControlPlaneToUpgrade(version)
	if err != nil {
		return fmt.Errorf("failed to wait for EKS cluster [%s] control plane to be upgraded to [%s], Err: [%v]", e.clusterName, version, err)
	}

	// Update Max Unavailable values for Node Group
	err = e.SetMaxUnavailableForUpgradeInNodeGroup(e.pxNodeGroupName)
	if err != nil {
		return fmt.Errorf("failed to configure EKS cluster [%s] node group [%s] Max Unavailable values, Err: [%v]", e.clusterName, e.pxNodeGroupName, err)
	}

	// Upgrade Node Group
	err = e.UpgradeNodeGroup(e.pxNodeGroupName, version)
	if err != nil {
		return fmt.Errorf("failed to upgrade EKS cluster [%s] node group [%s] to [%s]. Err: [%v]", e.clusterName, e.pxNodeGroupName, version, err)
	}

	// Wait for the portworx node group to be upgraded
	err = e.WaitForNodeGroupToUpgrade(e.pxNodeGroupName, version)
	if err != nil {
		return fmt.Errorf("failed to wait for EKS cluster [%s] node group [%s] to be upgraded to [%s]. Err: [%v]", e.clusterName, e.pxNodeGroupName, version, err)
	}
	log.Infof("Successfully finished EKS cluster [%s] upgrade from [%s] to [%s]", e.clusterName, currentVersion, version)
	return nil
}

// getAllInstances returns all instances in the EKS cluster
func (e *EKS) getAllInstances() ([]ec2types.Instance, error) {
	err := e.configureEKSClient()
	if err != nil {
		return nil, fmt.Errorf("failed to configure EKS client. Err: [%v]", err)
	}
	var instances []ec2types.Instance
	params := &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("tag:kubernetes.io/cluster/" + e.clusterName),
				Values: []string{"owned"},
			},
			{
				Name:   aws.String("tag:eks:nodegroup-name"),
				Values: []string{e.pxNodeGroupName},
			},
		},
	}
	paginator := ec2.NewDescribeInstancesPaginator(e.ec2Client, params)
	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to list instances in [%s]. Err: [%v]", e.region, err)
		}
		for _, resv := range resp.Reservations {
			for _, ins := range resv.Instances {
				instances = append(instances, ins)
			}
		}
	}
	return instances, nil
}

// GetZones returns the availability zones of the EKS cluster
func (e *EKS) GetZones() ([]string, error) {
	err := e.configureEKSClient()
	if err != nil {
		return nil, fmt.Errorf("failed to configure EKS client. Err: [%v]", err)
	}
	instances, err := e.getAllInstances()
	if err != nil {
		return nil, fmt.Errorf("failed to get instances in EKS cluster [%s]. Err: [%v]", e.clusterName, err)
	}
	zoneMap := make(map[string]bool)
	for _, instance := range instances {
		if instance.Placement != nil && instance.Placement.AvailabilityZone != nil {
			zoneMap[*instance.Placement.AvailabilityZone] = true
		}
	}
	zones := make([]string, 0, len(zoneMap))
	for zone := range zoneMap {
		zones = append(zones, zone)
	}
	return zones, nil
}

// getInstanceIDByPrivateIpAddress returns the instance ID of the node with the specified private IP address
func (e *EKS) getInstanceIDByPrivateIpAddress(n node.Node) (string, error) {
	err := e.configureEKSClient()
	if err != nil {
		return "", fmt.Errorf("failed to configure EKS client. Err: [%v]", err)
	}
	instances, err := e.getAllInstances()
	if err != nil {
		return "", fmt.Errorf("failed to get instances in EKS cluster [%s]. Err: [%v]", e.clusterName, err)
	}
	for _, i := range instances {
		for _, addr := range n.Addresses {
			if aws.ToString(i.PrivateIpAddress) == addr {
				return aws.ToString(i.InstanceId), nil
			}
		}
	}
	return "", fmt.Errorf("failed to get node [%s] instanceID by privateIP address", n.Name)
}

// DeleteNode deletes the specified node from the EKS cluster
func (e *EKS) DeleteNode(n node.Node) error {
	err := e.configureEKSClient()
	if err != nil {
		return fmt.Errorf("failed to configure EKS client. Err: [%v]", err)
	}
	instanceID, err := e.getInstanceIDByPrivateIpAddress(n)
	if err != nil {
		return &node.ErrFailedToDeleteNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to get instance ID due to: %v", err),
		}
	}
	terminateInstanceInput := &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}
	_, err = e.ec2Client.TerminateInstances(context.Background(), terminateInstanceInput)
	if err != nil {
		return &node.ErrFailedToDeleteNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to terminate instance due to: %v", err),
		}
	}
	return nil
}

// GetASGName returns the name of the autoscaling group for the EKS cluster
func (e *EKS) GetASGName() (string, error) {
	err := e.configureEKSClient()
	if err != nil {
		return "", fmt.Errorf("failed to configure EKS client. Err: [%v]", err)
	}
	nodeGroup, err := e.eksClient.DescribeNodegroup(context.TODO(), &eks.DescribeNodegroupInput{
		ClusterName:   aws.String(e.clusterName),
		NodegroupName: aws.String(e.pxNodeGroupName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe node group [%s]. Err: [%v]", e.pxNodeGroupName, err)
	}
	if len(nodeGroup.Nodegroup.Resources.AutoScalingGroups) == 0 {
		return "", fmt.Errorf("no auto scaling groups found for node group [%s]", e.pxNodeGroupName)
	}
	asgName := nodeGroup.Nodegroup.Resources.AutoScalingGroups[0].Name
	log.Infof("Found ASG [%s] for node group [%s]", asgName, e.pxNodeGroupName)
	return aws.ToString(asgName), nil
}

// GetASGClusterSize returns the total size of the EKS cluster autoscaling group
func (e *EKS) GetASGClusterSize() (int64, error) {
	err := e.configureEKSClient()
	if err != nil {
		return 0, fmt.Errorf("failed to configure EKS client. Err: [%v]", err)
	}
	asgName, err := e.GetASGName()
	if err != nil {
		return 0, err
	}
	asg, err := e.autoscalingClient.DescribeAutoScalingGroups(context.TODO(), &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{asgName},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to describe ASG [%s]. Err: [%v]", asgName, err)
	}
	log.Infof("Found %d ASGs", len(asg.AutoScalingGroups))
	if len(asg.AutoScalingGroups) > 0 {
		return int64(aws.ToInt32(asg.AutoScalingGroups[0].DesiredCapacity)), nil
	}
	return 0, fmt.Errorf("no information found for ASG [%s]", asgName)
}

// SetASGClusterSize sets the size of the EKS cluster autoscaling group
func (e *EKS) SetASGClusterSize(perZoneCount int64, timeout time.Duration) error {
	err := e.configureEKSClient()
	if err != nil {
		return fmt.Errorf("failed to configure EKS client. Err: [%v]", err)
	}
	asgName, err := e.GetASGName()
	if err != nil {
		return err
	}
	minSize := perZoneCount
	maxSize := perZoneCount
	desiredCapacity := perZoneCount
	_, err = e.autoscalingClient.UpdateAutoScalingGroup(context.TODO(), &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(asgName),
		MinSize:              aws.Int32(int32(minSize)),
		MaxSize:              aws.Int32(int32(maxSize)),
		DesiredCapacity:      aws.Int32(int32(desiredCapacity)),
	})
	if err != nil {
		return fmt.Errorf("failed to update auto scaling group [%s] settings. Min: [%d], Max: [%d], Desired: [%d], Err: [%v]", asgName, minSize, maxSize, desiredCapacity, err)
	}
	t := func() (interface{}, bool, error) {
		asg, err := e.autoscalingClient.DescribeAutoScalingGroups(context.TODO(), &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []string{asgName},
		})
		if err != nil {
			return nil, false, fmt.Errorf("failed to describe ASG [%s]. Err: [%v]", asgName, err)
		}
		if len(asg.AutoScalingGroups) == 0 {
			return nil, false, fmt.Errorf("ASG [%s] not found", asgName)
		}
		currentCapacity := aws.ToInt32(asg.AutoScalingGroups[0].DesiredCapacity)
		if currentCapacity == int32(desiredCapacity) {
			return nil, false, nil
		}
		return nil, true, fmt.Errorf("waiting for ASG [%s] to scale. Current desired capacity: [%d], expected: [%d]", asgName, currentCapacity, desiredCapacity)
	}
	_, err = task.DoRetryWithTimeout(t, timeout, defaultEKSUpgradeRetryInterval)
	if err != nil {
		return fmt.Errorf("timeout waiting for ASG [%s] to reach desired capacity [%d]. Err: [%v]", asgName, desiredCapacity, err)
	}
	log.Infof("Successfully scaled ASG [%s] to desired capacity [%d]", asgName, desiredCapacity)
	return nil
}

// StopKubelet allows to stop kubelet on a give node
func (e *EKS) StopKubelet(n node.Node, options node.SystemctlOpts) error {
	var err error

	instanceID, err := e.getInstanceIDByPrivateIpAddress(n)
	if err != nil {
		return err
	}
	log.Infof("Got Insatacne id:%v", instanceID)

	command := "sudo systemctl stop kubelet"
	param := make(map[string][]*string)
	param["commands"] = []*string{
		aws_pkg.String(command),
	}
	sendCommandInput := &ssm.SendCommandInput{
		Comment:      aws_pkg.String(command),
		DocumentName: aws_pkg.String("AWS-RunShellScript"),
		Parameters:   param,
		InstanceIds: []*string{
			aws_pkg.String(instanceID),
		},
	}
	log.Infof("sendCommandInput:%+v", sendCommandInput)
	sendCommandOutput, err := e.svcSsm.SendCommand(sendCommandInput)
	log.Infof("sendCommandOutput:%+v", sendCommandOutput)
	if err != nil {
		log.Infof("sendCommandOutput Err:%+v", err)
		return &node.ErrFailedToRunCommand{
			Node:  n,
			Cause: fmt.Sprintf("failed to send command to instance %s: %v", instanceID, err),
		}
	}
	if sendCommandOutput.Command == nil || sendCommandOutput.Command.CommandId == nil {
		return fmt.Errorf("no command returned after sending command to %s", instanceID)
	}
	listCmdsInput := &ssm.ListCommandInvocationsInput{
		CommandId: sendCommandOutput.Command.CommandId,
	}
	t := func() (interface{}, bool, error) {
		return "", true, e.connect(n, listCmdsInput)
	}

	if _, err := task.DoRetryWithTimeout(t, options.Timeout, options.TimeBeforeRetry); err != nil {
		return &node.ErrFailedToRunCommand{
			Node:  n,
			Cause: err.Error(),
		}
	}
	return err
}

// StartKubelet allows to start kubelet on a give node
func (e *EKS) StartKubelet(n node.Node, options node.SystemctlOpts) error {
	var err error
	instanceID, err := e.getInstanceIDByPrivateIpAddress(n)
	if err != nil {
		return err
	}
	log.Infof("Got Insatacne id:%v", instanceID)
	if err != nil {
		return &node.ErrFailedToTestConnection{
			Node:  n,
			Cause: fmt.Sprintf("failed to get instance ID for connection due to: %v", err),
		}
	}
	command := "sudo systemctl start kubelet"
	param := make(map[string][]*string)
	param["commands"] = []*string{
		aws_pkg.String(command),
	}
	sendCommandInput := &ssm.SendCommandInput{
		Comment:      aws_pkg.String(command),
		DocumentName: aws_pkg.String("AWS-RunShellScript"),
		Parameters:   param,
		InstanceIds: []*string{
			aws_pkg.String(instanceID),
		},
	}
	log.Infof("sendCommandInput:%+v", sendCommandInput)
	sendCommandOutput, err := e.svcSsm.SendCommand(sendCommandInput)
	log.Infof("sendCommandOutput:%+v", sendCommandOutput)
	if err != nil {
		log.Infof("sendCommandOutput Err:%+v", err)
		return &node.ErrFailedToRunCommand{
			Node:  n,
			Cause: fmt.Sprintf("failed to send command to instance %s: %v", instanceID, err),
		}
	}
	if sendCommandOutput.Command == nil || sendCommandOutput.Command.CommandId == nil {
		return fmt.Errorf("no command returned after sending command to %s", instanceID)
	}
	listCmdsInput := &ssm.ListCommandInvocationsInput{
		CommandId: sendCommandOutput.Command.CommandId,
	}
	t := func() (interface{}, bool, error) {
		return "", true, e.connect(n, listCmdsInput)
	}

	if _, err := task.DoRetryWithTimeout(t, options.Timeout, options.TimeBeforeRetry); err != nil {
		return &node.ErrFailedToRunCommand{
			Node:  n,
			Cause: err.Error(),
		}
	}
	return err
}

func (e *EKS) connect(n node.Node, listCmdsInput *ssm.ListCommandInvocationsInput) error {
	var status string
	listCmdInvsOutput, _ := e.svcSsm.ListCommandInvocations(listCmdsInput)
	for _, cmd := range listCmdInvsOutput.CommandInvocations {
		status = strings.TrimSpace(*cmd.StatusDetails)
		if status == "Success" {
			return nil
		}
	}
	return &node.ErrFailedToTestConnection{
		Node:  n,
		Cause: fmt.Sprintf("Failed to connect. Command status is %s", status),
	}
}

func init() {
	e := &EKS{}
	scheduler.Register(SchedName, e)
}
