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
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
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

type AWS struct {
	ssh.SSH
	region            string
	clusterName       string
	config            aws.Config
	eksClient         *eks.Client
	ec2Client         *ec2.Client
	ssmClient         *ssm.Client
	autoscalingClient *autoscaling.Client
	ec2Instances      []ec2types.Instance
}

func (a *AWS) String() string {
	return DriverName
}

func (a *AWS) Init(nodeOpts node.InitOptions) error {
	err := a.SSH.Init(nodeOpts)
	if err != nil {
		return err
	}
	a.config, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion(a.region))
	if err != nil {
		return fmt.Errorf("failed to load default AWS SDK config for region [%s] due to [%v]", a.region, err)
	}
	a.eksClient = eks.NewFromConfig(a.config)
	a.ec2Client = ec2.NewFromConfig(a.config)
	a.ssmClient = ssm.NewFromConfig(a.config)
	a.autoscalingClient = autoscaling.NewFromConfig(a.config)
	workerNodes := node.GetWorkerNodes()
	log.Infof("There are [%d] worker nodes in [%s] cluster in [%s]", len(workerNodes), a.clusterName, a.region)
	testConnectionOpts := node.ConnectionOpts{
		Timeout:         1 * time.Minute,
		TimeBeforeRetry: 10 * time.Second,
	}
	for _, n := range workerNodes {
		log.Infof("Testing connection to worker node [%s]", n.Name)
		err = a.TestConnection(n, testConnectionOpts)
		if err != nil {
			return err
		}
	}

	a.ec2Instances, err = a.getAllInstances()
	if err != nil {
		return err
	}
	return nil
}

func (a *AWS) TestConnection(n node.Node, options node.ConnectionOpts) error {
	instanceID, err := a.getNodeIDByPrivateIpAddress(n)
	if err != nil {
		return &node.ErrFailedToTestConnection{
			Node:  n,
			Cause: fmt.Sprintf("failed to get instanceID for connection due to [%v]", err),
		}
	}
	log.Infof("Node [%s] has instanceID [%v]", n.Name, instanceID)
	command := "uptime"
	param := make(map[string][]string)
	param["commands"] = []string{command}
	sendCommandInput := &ssm.SendCommandInput{
		Comment:      aws.String(command),
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters:   param,
		InstanceIds:  []string{instanceID},
	}
	log.Infof("Sending command [%s] to node [%s] with instanceID [%v]", command, n.Name, instanceID)
	sendCommandOutput, err := a.ssmClient.SendCommand(context.TODO(), sendCommandInput)
	if err != nil {
		return &node.ErrFailedToTestConnection{
			Node:  n,
			Cause: fmt.Sprintf("failed to send command to instanceID [%s] due to [%v]", instanceID, err),
		}
	}
	if sendCommandOutput.Command == nil || sendCommandOutput.Command.CommandId == nil {
		return fmt.Errorf("no response received after sending command to instanceID [%s]", instanceID)
	}
	t := func() (interface{}, bool, error) {
		listCmdInvocationsOutput, err := a.ssmClient.ListCommandInvocations(
			context.TODO(),
			&ssm.ListCommandInvocationsInput{
				CommandId: sendCommandOutput.Command.CommandId,
			},
		)
		if err != nil {
			return nil, false, fmt.Errorf("error listing command invocations: %v", err)
		}
		for _, cmd := range listCmdInvocationsOutput.CommandInvocations {
			switch cmd.Status {
			case ssmtypes.CommandInvocationStatusSuccess:
				return nil, false, nil
			default:
				return nil, true, fmt.Errorf("current status of the commandID [%s] is [%s], expected [%s]", *cmd.CommandId, cmd.Status, ssmtypes.CommandInvocationStatusSuccess)
			}
		}
		return nil, false, fmt.Errorf("no command invocations found for commandID [%s]", *sendCommandOutput.Command.CommandId)
	}
	_, err = task.DoRetryWithTimeout(t, options.Timeout, options.TimeBeforeRetry)
	if err != nil {
		return &node.ErrFailedToTestConnection{
			Node:  n,
			Cause: err.Error(),
		}
	}
	return nil
}

func (a *AWS) RebootNode(n node.Node, options node.RebootNodeOpts) error {
	instanceID, err := a.getNodeIDByPrivateIpAddress(n)
	if err != nil {
		return &node.ErrFailedToRebootNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to get instanceID due to: %v", err),
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

func (a *AWS) ShutdownNode(n node.Node, options node.ShutdownNodeOpts) error {
	instanceID, err := a.getNodeIDByPrivateIpAddress(n)
	if err != nil {
		return &node.ErrFailedToShutdownNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to get instanceID due to: %v", err),
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

func (a *AWS) DeleteNode(n node.Node, timeout time.Duration) error {
	instanceID, err := a.getNodeIDByPrivateIpAddress(n)
	if err != nil {
		return &node.ErrFailedToDeleteNode{
			Node:  n,
			Cause: fmt.Sprintf("failed to get instanceID due to: %v", err),
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

// TODO add AWS implementation for this

func (a *AWS) FindFiles(path string, n node.Node, options node.FindOpts) (string, error) {
	return "", nil
}

// TODO implement for AWS

func (a *AWS) Systemctl(n node.Node, service string, options node.SystemctlOpts) error {
	return nil
}

func (a *AWS) getAllInstances() ([]ec2types.Instance, error) {
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

func (a *AWS) getNodeIDByPrivateIpAddress(n node.Node) (string, error) {
	for _, i := range a.ec2Instances {
		for _, addr := range n.Addresses {
			if aws.ToString(i.PrivateIpAddress) == addr {
				return aws.ToString(i.InstanceId), nil
			}
		}
	}
	return "", fmt.Errorf("failed to get node [%s] instanceID by privateIP address [%s]", n.Name, n.Addresses)
}

func (a *AWS) GetASGClusterSize() (int64, error) {
	nodeGroups, err := a.eksClient.ListNodegroups(context.TODO(), &eks.ListNodegroupsInput{
		ClusterName: aws.String(a.clusterName),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list node groups for cluster '%s': %v", a.clusterName, err)
	}
	log.Infof("Found %d node groups", len(nodeGroups.Nodegroups))
	totalSize := int32(0)
	for _, nodeGroupName := range nodeGroups.Nodegroups {
		if nodeGroupName != "ng-torpedo" {
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
	}
	return int64(totalSize), nil
}

func (a *AWS) GetZones() ([]string, error) {
	filters := []ec2types.Filter{
		{
			Name:   aws.String("tag:kubernetes.io/cluster/" + a.clusterName),
			Values: []string{"owned", "shared"},
		},
		{
			Name:   aws.String("instance-state-name"),
			Values: []string{"running"},
		},
	}
	resp, err := a.ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		Filters: filters,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %v", err)
	}
	zoneMap := make(map[string]bool)
	for _, reservation := range resp.Reservations {
		for _, instance := range reservation.Instances {
			if instance.Placement != nil && instance.Placement.AvailabilityZone != nil {
				zoneMap[*instance.Placement.AvailabilityZone] = true
			}
		}
	}
	zones := make([]string, 0, len(zoneMap))
	for zone := range zoneMap {
		zones = append(zones, zone)
	}

	return zones, nil
}

func (a *AWS) SetClusterVersion(version string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	input := &eks.UpdateClusterVersionInput{
		Name:    aws.String(a.clusterName),
		Version: aws.String(version),
	}
	_, err := a.eksClient.UpdateClusterVersion(ctx, input)
	if err != nil {
		return fmt.Errorf("error initiating cluster version update to %v: %v", version, err)
	}
	log.Infof("UpdateClusterVersion initiated for control plane to version %s", version)
	if err := a.waitForClusterUpdate(ctx, version); err != nil {
		return err
	}
	return a.upgradeNodeGroups(ctx, version)
}

// waitForClusterUpdate waits for the EKS cluster's control plane to finish updating
func (a *AWS) waitForClusterUpdate(ctx context.Context, version string) error {
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
			if status == ekstypes.ClusterStatusActive {
				log.Infof("Cluster %v successfully updated to version %v", a.clusterName, version)
				return nil
			} else if status == ekstypes.ClusterStatusFailed {
				return fmt.Errorf("cluster update to version %v failed", version)
			}
		}
	}
}

// upgradeNodeGroups upgrades all node groups in the EKS cluster to the specified version
func (a *AWS) upgradeNodeGroups(ctx context.Context, version string) error {
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
		}
	}

	return nil
}

func init() {
	a := &AWS{
		SSH:         *ssh.New(),
		region:      os.Getenv("AWS_REGION"),
		clusterName: os.Getenv("AWS_CLUSTER_NAME"),
	}
	_ = node.Register(DriverName, a)
}
