package portworx

import (
	"context"
	"fmt"

	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/drivers/volume/portworx/schedops"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	// import ssh to invoke init
	_ "github.com/portworx/torpedo/drivers/node/ssh"
	// import scheduler k8s
	_ "github.com/portworx/torpedo/drivers/scheduler/k8s"
	// import portworx volume
	_ "github.com/portworx/torpedo/drivers/volume/portworx"
)

const (
	driverName            = "pxb"
	pxbRestPort           = 10001
	defaultPxbServicePort = 10002
	pxbServiceName        = "px-backup"
	pxbNameSpace          = "px-backup"
	schedulerDriverName   = "k8s"
	nodeDriverName        = "ssh"
	volumeDriverName      = "pxd"
	awsAccessKey          = "CT6R80D3ST0VW9NY6HYP"
	awsSecretKey          = "a0V6dPqu8C26KbAsa9qsIrfhsbvyGjjPPmZN2qD4"
)

type pxbackup struct {
	clusterManager         api.ClusterClient
	backupLocationManager  api.BackupLocationClient
	cloudCredentialManager api.CloudCredentialClient
	backupManger           api.BackupClient
	restoreManager         api.RestoreClient
	backupScheduleManager  api.BackupScheduleClient
	schedulePolicyManager  api.SchedulePolicyClient
	organizationManager    api.OrganizationClient
	healthManager          api.HealthClient

	schedulerDriver scheduler.Driver
	nodeDriver      node.Driver
	volumeDriver    volume.Driver
	schedOps        schedops.Driver
	refreshEndpoint bool
	token           string
}

func (d *pxbackup) String() string {
	return driverName
}

func (d *pxbackup) Init(schedulerDriverName string, nodeDriverName string, volumeDriverName string, token string) error {
	var err error

	logrus.Infof("using portworx px-backup driver under scheduler: %v", schedulerDriverName)

	d.nodeDriver, err = node.Get(nodeDriverName)
	d.token = token

	if err != nil {
		return err
	}

	d.schedulerDriver, err = scheduler.Get(schedulerDriverName)
	if err != nil {
		return fmt.Errorf("Error getting scheduler driver %v: %v", schedulerDriverName, err)
	}

	d.volumeDriver, err = volume.Get(volumeDriverName)
	if err != nil {
		return fmt.Errorf("Error getting volume driver %v: %v", volumeDriverName, err)
	}

	if err = d.setDriver(pxbServiceName, pxbNameSpace); err != nil {
		return fmt.Errorf("Error setting px-backup endpoint")
	}

	return nil

}

func (d *pxbackup) constructURL(ip string) string {
	return fmt.Sprintf("%s:%d", ip, defaultPxbServicePort)
}

func (d *pxbackup) testAndSetEndpoint(endpoint string) error {
	//	pxEndpoint := d.constructURL("10.233.105.67")
	pxEndpoint := d.constructURL(endpoint)
	fmt.Printf("testAndSetEndpoint-1: pxEndpoint = %v\n", pxEndpoint)
	conn, err := grpc.Dial(pxEndpoint, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("unable to get grpc connection\n")
		return err
	}

	d.healthManager = api.NewHealthClient(conn)
	_, err = d.healthManager.Status(context.Background(), &api.HealthStatusRequest{})
	fmt.Printf("HealthManager status %v\n", err)
	/*
		if err != nil {
			fmt.Printf("HealthManager API error\n")
			return err
		}*/

	d.clusterManager = api.NewClusterClient(conn)
	d.backupLocationManager = api.NewBackupLocationClient(conn)
	d.cloudCredentialManager = api.NewCloudCredentialClient(conn)
	d.backupManger = api.NewBackupClient(conn)
	d.restoreManager = api.NewRestoreClient(conn)
	d.backupScheduleManager = api.NewBackupScheduleClient(conn)
	d.schedulePolicyManager = api.NewSchedulePolicyClient(conn)
	d.organizationManager = api.NewOrganizationClient(conn)

	fmt.Printf("Using %v as endpoint for portworx backup driver", pxEndpoint)

	return nil
}

func (d *pxbackup) GetServiceEndpoint(serviceName string, nameSpace string) (string, error) {
	svc, err := core.Instance().GetService(serviceName, nameSpace)
	fmt.Printf("GetServiceEndpoint-1: %v\n", err)
	if err == nil {
		fmt.Printf("GetServiceEndpoint-2: %v\n", svc.Spec.ClusterIP)
		return svc.Spec.ClusterIP, nil
	}
	return "", err
}

func (d *pxbackup) setDriver(serviceName string, nameSpace string) error {
	var err error
	var endpoint string

	fmt.Printf("SetDriver-1\n")
	endpoint, err = d.GetServiceEndpoint(serviceName, nameSpace)
	fmt.Printf("setDriver-2: endpoint = %v err = %v\n", endpoint, err)
	if err == nil && endpoint != "" {
		if err = d.testAndSetEndpoint(endpoint); err == nil {
			d.refreshEndpoint = false
			fmt.Printf("seDriver-3: return")
			return nil
		}
	} else if err != nil && len(node.GetWorkerNodes()) == 0 {
		return err
	}

	d.refreshEndpoint = true
	logrus.Infof("Getting new backup driver")
	for _, n := range node.GetWorkerNodes() {
		for _, addr := range n.Addresses {
			fmt.Printf("setDriver-4: addr %v\n", addr)
			if err = d.testAndSetEndpoint(addr); err == nil {
				return nil
			}
		}
	}

	return fmt.Errorf("failed to get endpoint for portworx backup driver")

}

func (d *pxbackup) getOrganizationManager() api.OrganizationClient {
	if d.refreshEndpoint {
		d.setDriver(pxbServiceName, pxbNameSpace)
	}
	return d.organizationManager
}

func (d *pxbackup) getClusterManager() api.ClusterClient {
	if d.refreshEndpoint {
		d.setDriver(pxbServiceName, pxbNameSpace)
	}
	return d.clusterManager
}

func (d *pxbackup) getBackupLocationManager() api.BackupLocationClient {
	if d.refreshEndpoint {
		d.setDriver(pxbServiceName, pxbNameSpace)
	}
	return d.backupLocationManager
}

func (d *pxbackup) getCloudCredentialManager() api.CloudCredentialClient {
	if d.refreshEndpoint {
		d.setDriver(pxbServiceName, pxbNameSpace)
	}
	return d.cloudCredentialManager
}

func (d *pxbackup) getBackupManager() api.BackupClient {
	if d.refreshEndpoint {
		d.setDriver(pxbServiceName, pxbNameSpace)
	}
	return d.backupManger
}

func (d *pxbackup) getRestoreManager() api.RestoreClient {
	if d.refreshEndpoint {
		d.setDriver(pxbServiceName, pxbNameSpace)
	}
	return d.restoreManager
}

func (d *pxbackup) getBackupScheduleManager() api.BackupScheduleClient {
	if d.refreshEndpoint {
		d.setDriver(pxbServiceName, pxbNameSpace)
	}
	return d.backupScheduleManager
}

func (d *pxbackup) getSchedulePolicyManager() api.SchedulePolicyClient {
	if d.refreshEndpoint {
		d.setDriver(pxbServiceName, pxbNameSpace)
	}
	return d.schedulePolicyManager
}

func (d *pxbackup) CreateOrganization(orgName string) error {
	fmt.Printf("In CreateOrganization\n")
	org := d.getOrganizationManager()
	req := &api.OrganizationCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name: orgName,
		},
	}

	fmt.Printf("calling create organization\n")
	_, err := org.Create(context.Background(), req)
	if err != nil {
		fmt.Printf("Unable to create organization %v\n", err)
	}
	return nil
}

func (d *pxbackup) GetOrganization() error {
	org := d.getOrganizationManager()
	req := &api.OrganizationEnumerateRequest{}

	_, err := org.Enumerate(context.Background(), req)
	if err != nil {
		fmt.Printf("unable to get org %v\n", err)
	}
	return nil
}

func (d *pxbackup) CreateCloudCredential(orgID string, name string) error {
	fmt.Printf("In CreateCloudCredential\n")
	cc := d.getCloudCredentialManager()
	req := &api.CloudCredentialCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name:  name,
			OrgId: orgID,
		},
		CloudCredential: &api.CloudCredentialInfo{},
	}

	req.CloudCredential.Type = api.CloudCredentialInfo_AWS
	req.CloudCredential.Config = &api.CloudCredentialInfo_AwsConfig{
		AwsConfig: &api.AWSConfig{
			AccessKey: awsAccessKey,
			SecretKey: awsSecretKey,
		},
	}

	_, err := cc.Create(context.Background(), req)
	if err != nil {
		fmt.Printf("Unable to create credential %v\n", err)
	}
	return err
}

func (d *pxbackup) DeleteCloudCredential(orgID string, name string) error {
	fmt.Printf("In CreateCloudCredential\n")
	cc := d.getCloudCredentialManager()
	req := &api.CloudCredentialDeleteRequest{
		Name:  name,
		OrgId: orgID,
	}

	_, err := cc.Delete(context.Background(), req)
	if err != nil {
		fmt.Printf("Unable to delete credential %v\n", err)
	}
	return nil
}

func (d *pxbackup) CreateCluster(orgID string, name string, cloudCredntial string) error {
	return nil
}

func (d *pxbackup) GetCluster(OrgID string) error {
	return nil
}

func (d *pxbackup) DeleteCluster(orgID string, name string) error {
	return nil
}

func (d *pxbackup) CreateBackupLocation(orgID string, name string, cloudCredential string, path string) error {
	return nil
}

func (d *pxbackup) GetBackupLocation(orgID string, name string) error {
	return nil
}

func (d *pxbackup) DeleteBackupLocation(orgID string, name string) error {
	return nil
}

func (d *pxbackup) CreateBackup(orgID string, name string, backupLocation string, cluster string) error {
	return nil
}

func (d *pxbackup) GetBackupList(orgID string) error {
	return nil
}

func (d *pxbackup) InspectBackup(orgID string, name string) error {
	return nil
}

func (d *pxbackup) DeleteBackup(orgID string, name string) error {
	return nil
}

/*
func (d *pxbackup) CreateCluster(name string, orgID string, labels map[string]string) {
	fmt.Printf("In CreateCluster\n")
	cc := d.getCloudCredentialManager()
	req := &api.CloudCredentialCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name:  "abc",
			OrgId: "abc123",
		},
		CloudCredential: &api.CloudCredentialInfo{},
	}

	req.CloudCredential.Type = api.CloudCredentialInfo_AWS
	req.CloudCredential.Config = &api.CloudCredentialInfo_AwsConfig{
		AwsConfig: &api.AWSConfig{
			AccessKey: awsAccessKey,
			SecretKey: awsSecretKey,
		},
	}
	fmt.Printf("calling create cloud credential\n")
	_, err := cc.Create(context.Background(), req)
	if err != nil {
		fmt.Printf("Unable to create credential %v\n", err)
	}
	return
}
*/

func init() {
	fmt.Printf("BackupDriver init called\n")
	backup.Register(driverName, &pxbackup{})
}
