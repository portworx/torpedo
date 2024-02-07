package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
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
	region      string
	clusterName string
	config      aws.Config
	eksClient   *eks.Client
	ec2Client   *ec2.Client
	ssmClient   *ssm.Client
	instances   []ec2types.Instance
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

//func (a *Aws) TestConnection(n node.Node, options node.ConnectionOpts) error {
//	var err error
//	instanceID, err := a.getNodeIDByPrivateIpAddress(n)
//	if err != nil {
//		return &node.ErrFailedToTestConnection{
//			Node:  n,
//			Cause: fmt.Sprintf("failed to get instance ID for connection due to: %v", err),
//		}
//	}
//	log.Infof("Node [%s] has instance ID [%v]", n.Name, instanceID)
//
//	command := "uptime"
//	param := make(map[string][]string)
//	param["commands"] = []string{command}
//	sendCommandInput := &ssm.SendCommandInput{
//		Comment:      aws.String(command),
//		DocumentName: aws.String("AWS-RunShellScript"),
//		Parameters:   param,
//		InstanceIds:  []string{instanceID},
//	}
//	log.Infof("sendCommandInput: [%+v]", sendCommandInput)
//	sendCommandOutput, err := a.ssmClient.SendCommand(context.TODO(), sendCommandInput)
//	if err != nil {
//		log.Infof("sendCommandOutput Error: [%+v]", err)
//		return &node.ErrFailedToTestConnection{
//			Node:  n,
//			Cause: fmt.Sprintf("failed to send command to instance %s: %v", instanceID, err),
//		}
//	}
//	log.Infof("sendCommandOutput: [%+v]", sendCommandOutput)
//	if sendCommandOutput.Command == nil || sendCommandOutput.Command.CommandId == nil {
//		return fmt.Errorf("no command returned after sending command to [%s]", instanceID)
//	}
//	listCmdInput := &ssm.ListCommandInvocationsInput{
//		CommandId: sendCommandOutput.Command.CommandId,
//	}
//	t := func() (interface{}, bool, error) {
//		return "", true, a.connect(n, listCmdInput)
//	}
//
//	if _, err := task.DoRetryWithTimeout(t, options.Timeout, options.TimeBeforeRetry); err != nil {
//		return &node.ErrFailedToTestConnection{
//			Node:  n,
//			Cause: err.Error(),
//		}
//	}
//	return err
//}
//
//func (a *Aws) connect(n node.Node, params *ssm.ListCommandInvocationsInput) error {
//	listCmdInvocationsOutput, err := a.ssmClient.ListCommandInvocations(context.TODO(), params)
//	if err != nil {
//		log.Infof("Error listing command invocations: %v", err)
//		return &node.ErrFailedToTestConnection{
//			Node:  n,
//			Cause: fmt.Sprintf("error listing command invocations: %v", err),
//		}
//	}
//
//	found := false
//	for _, cmd := range listCmdInvocationsOutput.CommandInvocations {
//		if cmd.Status == ssmtypes.CommandInvocationStatusSuccess {
//			return nil
//		} else if cmd.Status != ssmtypes.CommandInvocationStatusPending && cmd.Status != ssmtypes.CommandInvocationStatusInProgress {
//			found = true
//			break
//		}
//	}
//
//	if !found {
//		return fmt.Errorf("no completed command invocations found for node %v", n)
//	}
//
//	return &node.ErrFailedToTestConnection{
//		Node:  n,
//		Cause: fmt.Sprintf("failed to connect. Last command status was not successful"),
//	}
//}

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

//	func (a *Aws) RebootNode(n node.Node, options node.RebootNodeOpts) error {
//		var err error
//		instanceID, err := a.getNodeIDByPrivAddr(n)
//		if err != nil {
//			return &node.ErrFailedToRebootNode{
//				Node:  n,
//				Cause: fmt.Sprintf("failed to get instance ID due to: %v", err),
//			}
//		}
//		//Reboot the instance by its InstanceID
//		rebootInput := &ec2.RebootInstancesInput{
//			InstanceIds: []*string{
//				aws_pkg.String(instanceID),
//			},
//		}
//		_, err = a.svc.RebootInstances(rebootInput)
//		if err != nil {
//			return &node.ErrFailedToRebootNode{
//				Node:  n,
//				Cause: fmt.Sprintf("failed to reboot instance due to: %v", err),
//			}
//		}
//
//		return nil
//	}
//
//	func (a *Aws) ShutdownNode(n node.Node, options node.ShutdownNodeOpts) error {
//		var err error
//		instanceID, err := a.getNodeIDByPrivAddr(n)
//		if err != nil {
//			return &node.ErrFailedToShutdownNode{
//				Node:  n,
//				Cause: fmt.Sprintf("failed to get instance ID due to: %v", err),
//			}
//		}
//		//Reboot the instance by its InstanceID
//		stopInstanceInput := &ec2.StopInstancesInput{
//			InstanceIds: []*string{
//				aws_pkg.String(instanceID),
//			},
//		}
//		_, err = a.svc.StopInstances(stopInstanceInput)
//		if err != nil {
//			return &node.ErrFailedToShutdownNode{
//				Node:  n,
//				Cause: fmt.Sprintf("failed to stop instance due to: %v", err),
//			}
//		}
//
//		return nil
//	}
//
//	func (a *Aws) DeleteNode(n node.Node, timeout time.Duration) error {
//		var err error
//		instanceID, err := a.getNodeIDByPrivAddr(n)
//		if err != nil {
//			return &node.ErrFailedToDeleteNode{
//				Node:  n,
//				Cause: fmt.Sprintf("failed to get instance ID due to: %v", err),
//			}
//		}
//		//Terminate the instance by its InstanceID
//		stopInstanceInput := &ec2.TerminateInstancesInput{
//			InstanceIds: []*string{
//				aws_pkg.String(instanceID),
//			},
//		}
//		_, err = a.svc.TerminateInstances(stopInstanceInput)
//		if err != nil {
//			return &node.ErrFailedToDeleteNode{
//				Node:  n,
//				Cause: fmt.Sprintf("failed to terminate instance due to: %v", err),
//			}
//		}
//
//		return nil
//	}
//
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

//	func (a *Aws) SetClusterVersion(version string, timeout time.Duration) error {
//		log.Infof("Aws SetClusterVersion to %s", version)
//		clusterName := os.Getenv("AWS_CLUSTER_NAME")
//		if clusterName == "" {
//			return fmt.Errorf("env AWS_CLUSTER_NAME not found")
//		}
//		region := os.Getenv("AWS_REGION")
//		if region == "" {
//			return fmt.Errorf("env AWS_REGION not found")
//		}
//		cfg, err := config.LoadDefaultConfig(context.TODO(),
//			config.WithRegion(region),
//		)
//		if err != nil {
//			return fmt.Errorf("unable to load SDK config, %v", err)
//		}
//		eksClient := eks.NewFromConfig(cfg)
//		input := &eks.UpdateClusterVersionInput{
//			Name:    aws.String(clusterName),
//			Version: aws.String(version),
//		}
//		result, err := eksClient.UpdateClusterVersion(context.TODO(), input)
//		if err != nil {
//			return fmt.Errorf("error updating cluster version: %v", err)
//		}
//		log.Infof("UpdateClusterVersion Result: %v", result)
//		return nil
//	}

func init() {
	a := &Aws{
		SSH: *ssh.New(),
	}
	node.Register(DriverName, a)
}
