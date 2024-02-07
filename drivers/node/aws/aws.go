package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/node/ssh"
	"github.com/portworx/torpedo/pkg/log"
	"os"
	"time"
)

const (
	// DriverName is the name of the aws driver
	DriverName = "aws"
)

type Aws struct {
	ssh.SSH
	region            string
	clusterName       string
	config            aws.Config
	eksClient         *eks.Client
	ec2Client         *ec2.Client
	ssmClient         *ssm.Client
	autoscalingClient *autoscaling.Client
	instances         []ec2types.Instance
}

func (a *Aws) String() string {
	return DriverName
}

func (a *Aws) Init(nodeOpts node.InitOptions) error {
	err := a.SSH.Init(nodeOpts)
	if err != nil {
		return err
	}
	a.region = os.Getenv("AWS_REGION")
	if a.region == "" {
		return fmt.Errorf("env AWS_REGION not found")
	}
	a.clusterName = os.Getenv("AWS_CLUSTER_NAME")
	if a.clusterName == "" {
		return fmt.Errorf("env AWS_CLUSTER_NAME not found")
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(a.region))
	if err != nil {
		return fmt.Errorf("unable to load default SDK config. Err: [%v]", err)
	}
	a.config = cfg
	a.eksClient = eks.NewFromConfig(cfg)
	a.ec2Client = ec2.NewFromConfig(cfg)
	a.ssmClient = ssm.NewFromConfig(cfg)
	a.autoscalingClient = autoscaling.NewFromConfig(cfg)
	a.instances, err = a.getAllInstances()
	if err != nil {
		return err
	}
	nodes := node.GetWorkerNodes()
	log.Infof("Found %d worker nodes", len(nodes))
	for _, n := range nodes {
		if err := a.TestConnection(n, node.ConnectionOpts{
			Timeout:         1 * time.Minute,
			TimeBeforeRetry: 10 * time.Second,
		}); err != nil {
			return &node.ErrFailedToTestConnection{
				Node:  n,
				Cause: err.Error(),
			}
		}
	}
	return nil
}

func (a *Aws) TestConnection(n node.Node, options node.ConnectionOpts) error {
	var err error
	instanceID, err := a.getNodeIDByPrivateIpAddress(n)
	if err != nil {
		return &node.ErrFailedToTestConnection{
			Node:  n,
			Cause: fmt.Sprintf("failed to get instance ID for connection due to: %v", err),
		}
	}
	log.Infof("Node [%s] has instance ID [%v]", n.Name, instanceID)

	command := "uptime"
	param := make(map[string][]string)
	param["commands"] = []string{command}
	sendCommandInput := &ssm.SendCommandInput{
		Comment:      aws.String(command),
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters:   param,
		InstanceIds:  []string{instanceID},
	}
	log.Infof("Sending command to node [%s] with instance ID [%v]", n.Name, instanceID)
	sendCommandOutput, err := a.ssmClient.SendCommand(context.TODO(), sendCommandInput)
	if err != nil {
		return &node.ErrFailedToTestConnection{
			Node:  n,
			Cause: fmt.Sprintf("failed to send command to instance %s: %v", instanceID, err),
		}
	}

	if sendCommandOutput.Command == nil || sendCommandOutput.Command.CommandId == nil {
		return fmt.Errorf("no command returned after sending command to [%s]", instanceID)
	}

	// Wait for command to execute and check the status
	t := func() (interface{}, bool, error) {
		listCmdInput := &ssm.ListCommandInvocationsInput{
			CommandId: sendCommandOutput.Command.CommandId,
		}
		listCmdInvocationsOutput, err := a.ssmClient.ListCommandInvocations(context.TODO(), listCmdInput)
		if err != nil {
			return nil, false, fmt.Errorf("error listing command invocations: %v", err)
		}

		for _, cmd := range listCmdInvocationsOutput.CommandInvocations {
			if cmd.Status == ssmtypes.CommandInvocationStatusSuccess {
				return nil, true, nil // Success
			} else if cmd.Status != ssmtypes.CommandInvocationStatusPending && cmd.Status != ssmtypes.CommandInvocationStatusInProgress {
				// Found a non-pending and non-in-progress status, so command execution is finished but not successful
				return nil, true, fmt.Errorf("command execution finished with status %s", cmd.Status)
			}
		}

		// If no command invocations were found to be successful or failed, retry
		return nil, false, nil
	}

	if _, err := task.DoRetryWithTimeout(t, options.Timeout, options.TimeBeforeRetry); err != nil {
		return &node.ErrFailedToTestConnection{
			Node:  n,
			Cause: err.Error(),
		}
	}
	return nil
}

func (a *Aws) RebootNode(n node.Node, options node.RebootNodeOpts) error {
	instanceID, err := a.getNodeIDByPrivateIpAddress(n)
	if err != nil {
		return &node.ErrFailedToRebootNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to get instance ID due to: %v", err),
		}
	}
	rebootInput := &ec2.RebootInstancesInput{
		InstanceIds: []string{instanceID},
	}
	_, err = a.ec2Client.RebootInstances(context.Background(), rebootInput)
	if err != nil {
		return &node.ErrFailedToRebootNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to reboot instance due to: %v", err),
		}
	}
	return nil
}

func (a *Aws) ShutdownNode(n node.Node, options node.ShutdownNodeOpts) error {
	instanceID, err := a.getNodeIDByPrivateIpAddress(n)
	if err != nil {
		return &node.ErrFailedToShutdownNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to get instance ID due to: %v", err),
		}
	}

	stopInstanceInput := &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	}
	_, err = a.ec2Client.StopInstances(context.Background(), stopInstanceInput)
	if err != nil {
		return &node.ErrFailedToShutdownNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to stop instance due to: %v", err),
		}
	}

	return nil
}

func (a *Aws) DeleteNode(n node.Node, timeout time.Duration) error {
	instanceID, err := a.getNodeIDByPrivateIpAddress(n)
	if err != nil {
		return &node.ErrFailedToDeleteNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to get instance ID due to: %v", err),
		}
	}
	terminateInstanceInput := &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	}
	_, err = a.ec2Client.TerminateInstances(context.Background(), terminateInstanceInput)
	if err != nil {
		return &node.ErrFailedToDeleteNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to terminate instance due to: %v", err),
		}
	}

	return nil
}

// // TODO add AWS implementation for this
//
//	func (a *Aws) FindFiles(path string, n node.Node, options node.FindOpts) (string, error) {
//		return "", nil
//	}
//
// // TODO implement for AWS
//
//	func (a *Aws) Systemctl(n node.Node, service string, options node.SystemctlOpts) error {
//		return nil
//	}

func (a *Aws) getAllInstances() ([]ec2types.Instance, error) {
	var instances []ec2types.Instance
	clusterTagKey := "tag:kubernetes.io/cluster/" + a.clusterName
	params := &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{
			{
				Name:   &clusterTagKey,
				Values: []string{"owned", "shared"},
			},
		},
	}
	paginator := ec2.NewDescribeInstancesPaginator(a.ec2Client, params)
	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to list instances in [%s]. Err: [%v]", a.region, err)
		}
		for _, resv := range resp.Reservations {
			for _, ins := range resv.Instances {
				instances = append(instances, ins)
			}
		}
	}
	return instances, nil
}

func (a *Aws) getNodeIDByPrivateIpAddress(n node.Node) (string, error) {
	for _, i := range a.instances {
		for _, addr := range n.Addresses {
			log.Infof("%#v %#v, %#v, %#v", *i.InstanceId, aws.ToString(i.PrivateIpAddress) == addr, aws.ToString(i.PrivateIpAddress), addr)
			if aws.ToString(i.PrivateIpAddress) == addr {
				return aws.ToString(i.InstanceId), nil
			}
		}
	}
	return "", fmt.Errorf("failed to get node [%s] instanceID by privateIP address", n.Name)
}

func (a *Aws) GetASGClusterSize() (int64, error) {
	nodeGroups, err := a.eksClient.ListNodegroups(context.TODO(), &eks.ListNodegroupsInput{
		ClusterName: aws.String(a.clusterName),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list node groups for cluster '%s': %v", a.clusterName, err)
	}
	log.Infof("Found %d node groups", len(nodeGroups.Nodegroups))
	totalSize := int32(0)
	for _, nodeGroupName := range nodeGroups.Nodegroups {
		nodeGroup, err := a.eksClient.DescribeNodegroup(context.TODO(), &eks.DescribeNodegroupInput{
			ClusterName:   aws.String(a.clusterName),
			NodegroupName: aws.String(nodeGroupName),
		})
		if err != nil {
			return 0, fmt.Errorf("failed to describe node group '%s': %v", nodeGroupName, err)
		}
		asgName := nodeGroup.Nodegroup.Resources.AutoScalingGroups[0].Name
		log.Infof("Found ASG '%s' for node group '%s'", *asgName, nodeGroupName)

		// Now query the Auto Scaling API to get the size of this ASG
		asg, err := a.autoscalingClient.DescribeAutoScalingGroups(context.TODO(), &autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []string{*asgName},
		})
		if err != nil {
			return 0, fmt.Errorf("failed to describe ASG '%s': %v", asgName, err)
		}
		log.Infof("Found %d ASGs", len(asg.AutoScalingGroups))
		if len(asg.AutoScalingGroups) > 0 {
			totalSize += aws.ToInt32(asg.AutoScalingGroups[0].DesiredCapacity)
		}
	}
	return int64(totalSize), nil
}

func (a *Aws) GetZones() ([]string, error) {
	//resp, err := a.ec2Client.DescribeAvailabilityZones(context.TODO(), &ec2.DescribeAvailabilityZonesInput{
	//	Filters: []ec2types.Filter{
	//		{
	//			Name:   aws.String("state"),
	//			Values: []string{"available"},
	//		},
	//	},
	//})
	//if err != nil {
	//	return nil, fmt.Errorf("failed to describe availability zones: %v", err)
	//}
	//log.Info("Found %d availability zones", len(resp.AvailabilityZones))
	//zones := make([]string, len(resp.AvailabilityZones))
	//for i, zone := range resp.AvailabilityZones {
	//	zones[i] = *zone.ZoneName
	//	log.Infof("Found zone: %s", zones[i])
	//}
	//return zones, nil

	filters := []ec2types.Filter{
		{
			Name:   aws.String("tag:kubernetes.io/cluster/" + a.clusterName),
			Values: []string{"owned", "shared"}, // Adjust based on your tagging strategy
		},
		{
			Name:   aws.String("instance-state-name"),
			Values: []string{"running"}, // Only consider running instances
		},
	}

	// Describe instances with the specified filters
	resp, err := a.ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		Filters: filters,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %v", err)
	}

	// Use a map to track unique zone names
	zoneMap := make(map[string]bool)
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			if instance.Placement != nil && instance.Placement.AvailabilityZone != nil {
				zoneMap[*instance.Placement.AvailabilityZone] = true
			}
		}
	}

	// Convert the map keys to a slice
	zones := make([]string, 0, len(zoneMap))
	for zone := range zoneMap {
		zones = append(zones, zone)
	}

	return zones, nil
}

func (a *Aws) SetClusterVersion(version string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Initiate control plane upgrade
	input := &eks.UpdateClusterVersionInput{
		Name:    aws.String(a.clusterName),
		Version: aws.String(version),
	}
	_, err := a.eksClient.UpdateClusterVersion(ctx, input)
	if err != nil {
		return fmt.Errorf("error initiating cluster version update to %v: %v", version, err)
	}
	log.Infof("UpdateClusterVersion initiated for control plane to version %s", version)

	// Wait for control plane upgrade to complete
	if err := a.waitForClusterUpdate(ctx, version); err != nil {
		return err
	}

	// After successful control plane upgrade, proceed with node group upgrades
	return a.upgradeNodeGroups(ctx, version)
}

// waitForClusterUpdate waits for the EKS cluster's control plane to finish updating
func (a *Aws) waitForClusterUpdate(ctx context.Context, version string) error {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for cluster update to complete")
		case <-ticker.C:
			descInput := &eks.DescribeClusterInput{
				Name: aws.String(a.clusterName),
			}
			cluster, err := a.eksClient.DescribeCluster(ctx, descInput)
			if err != nil {
				log.Infof("Error fetching cluster status: %v", err)
				continue
			}
			if cluster.Cluster == nil {
				return fmt.Errorf("cluster '%v' not found", a.clusterName)
			}
			status := cluster.Cluster.Status
			log.Infof("Current cluster status: %v", status)
			if status == types.ClusterStatusActive {
				log.Infof("Cluster %v successfully updated to version %v", a.clusterName, version)
				return nil // Successfully updated
			} else if status == types.ClusterStatusFailed {
				return fmt.Errorf("cluster update to version %v failed", version)
			}
			// If status is neither ACTIVE nor FAILED, continue polling
		}
	}
}

// upgradeNodeGroups upgrades all node groups in the EKS cluster to the specified version
func (a *Aws) upgradeNodeGroups(ctx context.Context, version string) error {
	nodeGroups, err := a.eksClient.ListNodegroups(ctx, &eks.ListNodegroupsInput{
		ClusterName: aws.String(a.clusterName),
	})
	if err != nil {
		return fmt.Errorf("error listing node groups: %v", err)
	}

	for _, nodeGroupName := range nodeGroups.Nodegroups {
		if nodeGroupName != "ng-torpedo" {
			log.Infof("Upgrading node group %s to version %s", nodeGroupName, version)
			_, err := a.eksClient.UpdateNodegroupVersion(ctx, &eks.UpdateNodegroupVersionInput{
				ClusterName:   aws.String(a.clusterName),
				NodegroupName: aws.String(nodeGroupName),
				Version:       aws.String(version),
			})
			if err != nil {
				return fmt.Errorf("error upgrading node group %s: %v", nodeGroupName, err)
			}
			log.Infof("Node group %s upgrade initiated to version %s", nodeGroupName, version)
			// Optionally implement waiting for node group upgrade to complete here
		}
	}

	return nil
}

func init() {
	a := &Aws{
		SSH: *ssh.New(),
	}
	node.Register(DriverName, a)
}
