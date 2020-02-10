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
	pxEndpoint := d.constructURL(endpoint)
	conn, err := grpc.Dial(pxEndpoint, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("unable to get grpc connection\n")
		return err
	}

	d.healthManager = api.NewHealthClient(conn)
	_, err = d.healthManager.Status(context.Background(), &api.HealthStatusRequest{})
	if err != nil {
		fmt.Printf("HealthManager API error %v\n", err)
		return err
	}

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
	if err == nil {
		return svc.Spec.ClusterIP, nil
	}
	return "", err
}

func (d *pxbackup) setDriver(serviceName string, nameSpace string) error {
	var err error
	var endpoint string

	endpoint, err = d.GetServiceEndpoint(serviceName, nameSpace)
	if err == nil && endpoint != "" {
		if err = d.testAndSetEndpoint(endpoint); err == nil {
			d.refreshEndpoint = false
			return nil
		}
	} else if err != nil && len(node.GetWorkerNodes()) == 0 {
		return err
	}

	d.refreshEndpoint = true
	logrus.Infof("Getting new backup driver")
	for _, n := range node.GetWorkerNodes() {
		for _, addr := range n.Addresses {
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
	org := d.getOrganizationManager()
	req := &api.OrganizationCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name: orgName,
		},
	}

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

func (d *pxbackup) InspectCloudCredential(orgID string, name string) (*api.CloudCredentialInspectResponse, error) {
	cc := d.getCloudCredentialManager()
	resp, err := cc.Inspect(
		context.Background(),
		&api.CloudCredentialInspectRequest{Name: name, OrgId: orgID},
	)
	return resp, err
}

func (d *pxbackup) EnumerateCloudCredential(orgID string) (*api.CloudCredentialEnumerateResponse, error) {
	cc := d.getCloudCredentialManager()
	resp, err := cc.Enumerate(
		context.Background(),
		&api.CloudCredentialEnumerateRequest{OrgId: orgID},
	)

	return resp, err
}

func (d *pxbackup) DeleteCloudCredential(orgID string, name string) error {
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

func (d *pxbackup) CreateCluster(name string, orgID string, labels map[string]string, cloudCredential string, pxToken string, config string) (*api.ClusterCreateResponse, error) {
	cluster := d.getClusterManager()
	req := &api.ClusterCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name:   name,
			OrgId:  orgID,
			Labels: labels,
		},
		Cluster: &api.ClusterInfo{
			PxConfig: &api.PXConfig{
				AccessToken: pxToken,
			},
			Kubeconfig:      config,
			CloudCredential: cloudCredential,
		},
	}
	resp, err := cluster.Create(context.Background(), req)
	if err != nil {
		fmt.Printf("Unable to create cluster %v\n", err)
	}
	return resp, err
}

func (d *pxbackup) InspectCluster(orgID string, name string) (*api.ClusterInspectResponse, error) {
	cluster := d.getClusterManager()
	resp, err := cluster.Inspect(
		context.Background(),
		&api.ClusterInspectRequest{OrgId: orgID, Name: name},
	)
	return resp, err
}

func (d *pxbackup) EnumerateCluster(orgID string) (*api.ClusterEnumerateResponse, error) {
	cluster := d.getClusterManager()
	resp, err := cluster.Enumerate(
		context.Background(),
		&api.ClusterEnumerateRequest{OrgId: orgID},
	)
	return resp, err
}

func (d *pxbackup) DeleteCluster(orgID string, name string) (*api.ClusterDeleteResponse, error) {
	cluster := d.getClusterManager()
	resp, err := cluster.Delete(
		context.Background(),
		&api.ClusterDeleteRequest{OrgId: orgID, Name: name},
	)
	return resp, err
}

func (d *pxbackup) CreateBackupLocation(orgID string, name string, cloudCredential string,
	path string, encKey string, provider string, s3Endpoint string,
	s3Region string, disableSsl bool, disablePathStyle bool) (*api.BackupLocationCreateResponse, error) {
	bl := d.getBackupLocationManager()
	req := &api.BackupLocationCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name:  name,
			OrgId: orgID,
		},
		BackupLocation: &api.BackupLocationInfo{
			Path:          path,
			EncryptionKey: encKey,
		},
	}

	switch provider {
	case "s3":
		req.BackupLocation.Config = &api.BackupLocationInfo_S3Config{
			S3Config: &api.S3Config{
				Endpoint:         s3Endpoint,
				Region:           s3Region,
				DisableSsl:       disableSsl,
				DisablePathStyle: disablePathStyle,
			},
		}
		req.BackupLocation.Type = api.BackupLocationInfo_S3
	case "azure":
		req.BackupLocation.Type = api.BackupLocationInfo_Azure
	case "google":
		req.BackupLocation.Type = api.BackupLocationInfo_Google
	default:
		fmt.Printf("provider needs to be either azure, google or s3")
	}
	req.BackupLocation.CloudCredential = cloudCredential
	status, err := bl.Create(context.Background(), req)
	return status, err
}

func (d *pxbackup) EnumerateBackupLocation(orgID string) (*api.BackupLocationEnumerateResponse, error) {
	bl := d.getBackupLocationManager()
	resp, err := bl.Enumerate(
		context.Background(),
		&api.BackupLocationEnumerateRequest{
			OrgId: orgID,
		},
	)
	return resp, err
}

func (d *pxbackup) InspectBackupLocation(orgID string, name string) (*api.BackupLocationInspectResponse, error) {
	bl := d.getBackupLocationManager()
	resp, err := bl.Inspect(
		context.Background(),
		&api.BackupLocationInspectRequest{
			Name:  name,
			OrgId: orgID,
		},
	)
	return resp, err
}

func (d *pxbackup) DeleteBackupLocation(orgID string, name string) (*api.BackupLocationDeleteResponse, error) {
	bl := d.getBackupLocationManager()
	status, err := bl.Delete(
		context.Background(),
		&api.BackupLocationDeleteRequest{
			Name:  name,
			OrgId: orgID,
		},
	)
	return status, err
}

func (d *pxbackup) CreateBackup(orgID string, name string, backupLocation string, clusterName string, owner string, nameSpaces []string, labels map[string]string) (*api.BackupCreateResponse, error) {
	backup := d.getBackupManager()
	req := &api.BackupCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name:   name,
			OrgId:  orgID,
			Owner:  owner,
			Labels: labels,
		},
		BackupLocation: backupLocation,
		Cluster:        clusterName,
		Namespaces:     nameSpaces,
	}

	resp, err := backup.Create(context.Background(), req)

	return resp, err
}

func (d *pxbackup) EnumerateBackup(orgID string) (*api.BackupEnumerateResponse, error) {
	backup := d.getBackupManager()
	resp, err := backup.Enumerate(
		context.Background(),
		&api.BackupEnumerateRequest{
			OrgId: orgID,
		},
	)
	return resp, err
}

func (d *pxbackup) InspectBackup(orgID string, name string) (*api.BackupInspectResponse, error) {
	backup := d.getBackupManager()
	resp, err := backup.Inspect(
		context.Background(),
		&api.BackupInspectRequest{
			OrgId: orgID,
			Name:  name,
		},
	)
	return resp, err
}

func (d *pxbackup) DeleteBackup(orgID string, name string) (*api.BackupDeleteResponse, error) {
	backup := d.getBackupManager()
	resp, err := backup.Delete(
		context.Background(),
		&api.BackupDeleteRequest{
			OrgId: orgID,
			Name:  name,
		},
	)
	return resp, err
}

func init() {
	fmt.Printf("BackupDriver init called\n")
	backup.Register(driverName, &pxbackup{})
}
