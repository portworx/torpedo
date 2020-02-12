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
	//_ "github.com/portworx/torpedo/drivers/node/ssh"
	// import scheduler k8s
	//_ "github.com/portworx/torpedo/drivers/scheduler/k8s"
	// import portworx volume
	//_ "github.com/portworx/torpedo/drivers/volume/portworx"
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
)

type portworx struct {
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

func (p *portworx) String() string {
	return driverName
}

func (p *portworx) Init(schedulerDriverName string, nodeDriverName string, volumeDriverName string, token string) error {
	var err error

	logrus.Infof("using portworx px-backup driver under scheduler: %v", schedulerDriverName)

	p.nodeDriver, err = node.Get(nodeDriverName)
	if err != nil {
		return err
	}
	p.token = token

	p.schedulerDriver, err = scheduler.Get(schedulerDriverName)
	if err != nil {
		return fmt.Errorf("Error getting scheduler driver %v: %v", schedulerDriverName, err)
	}

	p.volumeDriver, err = volume.Get(volumeDriverName)
	if err != nil {
		return fmt.Errorf("Error getting volume driver %v: %v", volumeDriverName, err)
	}

	if err = p.setDriver(pxbServiceName, pxbNameSpace); err != nil {
		return fmt.Errorf("Error setting px-backup endpoint: %v", err)
	}

	return err

}

func (p *portworx) constructURL(ip string) string {
	return fmt.Sprintf("%s:%d", ip, defaultPxbServicePort)
}

func (p *portworx) testAndSetEndpoint(endpoint string) error {
	pxEndpoint := p.constructURL(endpoint)
	conn, err := grpc.Dial(pxEndpoint, grpc.WithInsecure())
	if err != nil {
		logrus.Errorf("unable to get grpc connection: %v", err)
		return err
	}

	p.healthManager = api.NewHealthClient(conn)
	_, err = p.healthManager.Status(context.Background(), &api.HealthStatusRequest{})
	if err != nil {
		logrus.Errorf("HealthManager API error: %v", err)
		return err
	}

	p.clusterManager = api.NewClusterClient(conn)
	p.backupLocationManager = api.NewBackupLocationClient(conn)
	p.cloudCredentialManager = api.NewCloudCredentialClient(conn)
	p.backupManger = api.NewBackupClient(conn)
	p.restoreManager = api.NewRestoreClient(conn)
	p.backupScheduleManager = api.NewBackupScheduleClient(conn)
	p.schedulePolicyManager = api.NewSchedulePolicyClient(conn)
	p.organizationManager = api.NewOrganizationClient(conn)

	logrus.Infof("Using %v as endpoint for portworx backup driver", pxEndpoint)

	return err
}

func (p *portworx) GetServiceEndpoint(serviceName string, nameSpace string) (string, error) {
	svc, err := core.Instance().GetService(serviceName, nameSpace)
	if err == nil {
		return svc.Spec.ClusterIP, nil
	}
	return "", err
}

func (p *portworx) setDriver(serviceName string, nameSpace string) error {
	var err error
	var endpoint string

	endpoint, err = p.GetServiceEndpoint(serviceName, nameSpace)
	if err == nil && endpoint != "" {
		if err = p.testAndSetEndpoint(endpoint); err == nil {
			p.refreshEndpoint = false
			return nil
		}
	} else if err != nil && len(node.GetWorkerNodes()) == 0 {
		return err
	}

	p.refreshEndpoint = true
	for _, n := range node.GetWorkerNodes() {
		for _, addr := range n.Addresses {
			if err = p.testAndSetEndpoint(addr); err == nil {
				return nil
			}
		}
	}

	return fmt.Errorf("failed to get endpoint for portworx backup driver: %v", err)
}

func (p *portworx) getOrganizationManager() api.OrganizationClient {
	if p.refreshEndpoint {
		p.setDriver(pxbServiceName, pxbNameSpace)
	}
	return p.organizationManager
}

func (p *portworx) getClusterManager() api.ClusterClient {
	if p.refreshEndpoint {
		p.setDriver(pxbServiceName, pxbNameSpace)
	}
	return p.clusterManager
}

func (p *portworx) getBackupLocationManager() api.BackupLocationClient {
	if p.refreshEndpoint {
		p.setDriver(pxbServiceName, pxbNameSpace)
	}
	return p.backupLocationManager
}

func (p *portworx) getCloudCredentialManager() api.CloudCredentialClient {
	if p.refreshEndpoint {
		p.setDriver(pxbServiceName, pxbNameSpace)
	}
	return p.cloudCredentialManager
}

func (p *portworx) getBackupManager() api.BackupClient {
	if p.refreshEndpoint {
		p.setDriver(pxbServiceName, pxbNameSpace)
	}
	return p.backupManger
}

func (p *portworx) getRestoreManager() api.RestoreClient {
	if p.refreshEndpoint {
		p.setDriver(pxbServiceName, pxbNameSpace)
	}
	return p.restoreManager
}

func (p *portworx) getBackupScheduleManager() api.BackupScheduleClient {
	if p.refreshEndpoint {
		p.setDriver(pxbServiceName, pxbNameSpace)
	}
	return p.backupScheduleManager
}

func (p *portworx) getSchedulePolicyManager() api.SchedulePolicyClient {
	if p.refreshEndpoint {
		p.setDriver(pxbServiceName, pxbNameSpace)
	}
	return p.schedulePolicyManager
}

func (p *portworx) getOrgID(params *backup.Request) string {
	if params != nil && params.OrgID != "" {
		return params.OrgID
	}
	return ""
}

func (p *portworx) getCloudCredentialName(params *backup.Request) string {
	if params != nil && params.CloudCredentialName != "" {
		return params.CloudCredentialName
	}
	return ""
}

func (p *portworx) getBackupLocationName(params *backup.Request) string {
	if params != nil && params.BackupLocationName != "" {
		return params.BackupLocationName
	}
	return ""
}

func (p *portworx) getClusterName(params *backup.Request) string {
	if params != nil && params.ClusterName != "" {
		return params.ClusterName
	}
	return ""
}

func (p *portworx) getBackupName(params *backup.Request) string {
	if params != nil && params.BackupName != "" {
		return params.BackupName
	}
	return ""
}

func (p *portworx) getNameSpaces(params *backup.Request) []string {
	return params.NameSpaces
}

func (p *portworx) getLabels(params *backup.Request) map[string]string {
	return params.Labels
}

func (p *portworx) getToken(params *backup.Request) string {
	if params != nil && params.Token != "" {
		return params.Token
	}
	return ""
}

func (p *portworx) getKubeconfig(params *backup.Request) string {
	if params != nil && params.Kubeconfig != "" {
		return params.Kubeconfig
	}
	return ""
}

func (p *portworx) getBackupLocationPath(params *backup.Request) string {
	if params != nil && params.BackupLocationPath != "" {
		return params.BackupLocationPath
	}
	return ""
}

func (p *portworx) getEncKey(params *backup.Request) string {
	if params != nil && params.EncKey != "" {
		return params.EncKey
	}
	return ""
}

func (p *portworx) getProvider(params *backup.Request) string {
	if params != nil && params.Provider != "" {
		return params.Provider
	}
	return ""
}

func (p *portworx) getS3Endpoint(params *backup.Request) string {
	if params != nil && params.S3Endpoint != "" {
		return params.S3Endpoint
	}
	return ""
}

func (p *portworx) getS3Region(params *backup.Request) string {
	if params != nil && params.S3Region != "" {
		return params.S3Region
	}
	return ""
}

func (p *portworx) getOwner(params *backup.Request) string {
	if params != nil && params.Owner != "" {
		return params.Owner
	}
	return ""
}

func (p *portworx) getDisableSsl(params *backup.Request) bool {
	if params != nil {
		return params.DisableSsl
	}
	return true
}

func (p *portworx) getDisablePathStyle(params *backup.Request) bool {
	if params != nil {
		return params.DisablePathStyle
	}
	return true
}

func (p *portworx) getAccessKey(params *backup.Request) string {
	if params != nil && params.AccessKey != "" {
		return params.AccessKey
	}
	return ""
}

func (p *portworx) getSecretKey(params *backup.Request) string {
	if params != nil && params.SecretKey != "" {
		return params.SecretKey
	}
	return ""
}

func (p *portworx) CreateOrganization(params *backup.Request) (*api.OrganizationCreateResponse, error) {
	org := p.getOrganizationManager()
	req := &api.OrganizationCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name: p.getOrgID(params),
		},
	}
	return org.Create(context.Background(), req)
}

func (p *portworx) EnumerateOrganization() (*api.OrganizationEnumerateResponse, error) {
	org := p.getOrganizationManager()
	req := &api.OrganizationEnumerateRequest{}

	return org.Enumerate(context.Background(), req)
}

func (p *portworx) CreateCloudCredential(params *backup.Request) (*api.CloudCredentialCreateResponse, error) {
	cc := p.getCloudCredentialManager()
	req := &api.CloudCredentialCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name:  p.getCloudCredentialName(params),
			OrgId: p.getOrgID(params),
		},
		CloudCredential: &api.CloudCredentialInfo{},
	}

	req.CloudCredential.Type = api.CloudCredentialInfo_AWS
	req.CloudCredential.Config = &api.CloudCredentialInfo_AwsConfig{
		AwsConfig: &api.AWSConfig{
			AccessKey: p.getAccessKey(params),
			SecretKey: p.getSecretKey(params),
		},
	}
	return cc.Create(context.Background(), req)
}

func (p *portworx) InspectCloudCredential(params *backup.Request) (*api.CloudCredentialInspectResponse, error) {
	cc := p.getCloudCredentialManager()
	resp, err := cc.Inspect(
		context.Background(),
		&api.CloudCredentialInspectRequest{
			Name:  p.getCloudCredentialName(params),
			OrgId: p.getOrgID(params),
		},
	)
	return resp, err
}

func (p *portworx) EnumerateCloudCredential(params *backup.Request) (*api.CloudCredentialEnumerateResponse, error) {
	cc := p.getCloudCredentialManager()
	resp, err := cc.Enumerate(
		context.Background(),
		&api.CloudCredentialEnumerateRequest{OrgId: p.getOrgID(params)},
	)
	return resp, err
}

func (p *portworx) DeleteCloudCredential(params *backup.Request) (*api.CloudCredentialDeleteResponse, error) {
	cc := p.getCloudCredentialManager()
	req := &api.CloudCredentialDeleteRequest{
		Name:  p.getCloudCredentialName(params),
		OrgId: p.getOrgID(params),
	}
	return cc.Delete(context.Background(), req)
}

func (p *portworx) CreateCluster(params *backup.Request) (*api.ClusterCreateResponse, error) {
	cluster := p.getClusterManager()
	req := &api.ClusterCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name:   p.getClusterName(params),
			OrgId:  p.getOrgID(params),
			Labels: p.getLabels(params),
		},
		Cluster: &api.ClusterInfo{
			PxConfig: &api.PXConfig{
				AccessToken: p.getToken(params),
			},
			Kubeconfig:      p.getKubeconfig(params),
			CloudCredential: p.getCloudCredentialName(params),
		},
	}
	return cluster.Create(context.Background(), req)
}

func (p *portworx) InspectCluster(params *backup.Request) (*api.ClusterInspectResponse, error) {
	cluster := p.getClusterManager()
	return cluster.Inspect(
		context.Background(),
		&api.ClusterInspectRequest{
			OrgId: p.getOrgID(params),
			Name:  p.getClusterName(params),
		},
	)
}

func (p *portworx) EnumerateCluster(params *backup.Request) (*api.ClusterEnumerateResponse, error) {
	cluster := p.getClusterManager()
	return cluster.Enumerate(
		context.Background(),
		&api.ClusterEnumerateRequest{OrgId: p.getOrgID(params)},
	)
}

func (p *portworx) DeleteCluster(params *backup.Request) (*api.ClusterDeleteResponse, error) {
	cluster := p.getClusterManager()
	return cluster.Delete(
		context.Background(),
		&api.ClusterDeleteRequest{
			OrgId: p.getOrgID(params),
			Name:  p.getClusterName(params),
		},
	)
}

func (p *portworx) CreateBackupLocation(params *backup.Request) (*api.BackupLocationCreateResponse, error) {
	bl := p.getBackupLocationManager()
	req := &api.BackupLocationCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name:  p.getBackupLocationName(params),
			OrgId: p.getOrgID(params),
		},
		BackupLocation: &api.BackupLocationInfo{
			Path:          p.getBackupLocationPath(params),
			EncryptionKey: p.getEncKey(params),
		},
	}

	switch p.getProvider(params) {
	case "s3":
		req.BackupLocation.Config = &api.BackupLocationInfo_S3Config{
			S3Config: &api.S3Config{
				Endpoint:         p.getS3Endpoint(params),
				Region:           p.getS3Region(params),
				DisableSsl:       p.getDisableSsl(params),
				DisablePathStyle: p.getDisablePathStyle(params),
			},
		}
		req.BackupLocation.Type = api.BackupLocationInfo_S3
	case "azure":
		req.BackupLocation.Type = api.BackupLocationInfo_Azure
	case "google":
		req.BackupLocation.Type = api.BackupLocationInfo_Google
	default:
		logrus.Errorf("provider needs to be either azure, google or s3")
	}
	req.BackupLocation.CloudCredential = p.getCloudCredentialName(params)
	return bl.Create(context.Background(), req)
}

func (p *portworx) EnumerateBackupLocation(params *backup.Request) (*api.BackupLocationEnumerateResponse, error) {
	bl := p.getBackupLocationManager()
	return bl.Enumerate(
		context.Background(),
		&api.BackupLocationEnumerateRequest{
			OrgId: p.getOrgID(params),
		},
	)
}

func (p *portworx) InspectBackupLocation(params *backup.Request) (*api.BackupLocationInspectResponse, error) {
	bl := p.getBackupLocationManager()
	return bl.Inspect(
		context.Background(),
		&api.BackupLocationInspectRequest{
			Name:  p.getBackupLocationName(params),
			OrgId: p.getOrgID(params),
		},
	)
}

func (p *portworx) DeleteBackupLocation(params *backup.Request) (*api.BackupLocationDeleteResponse, error) {
	bl := p.getBackupLocationManager()
	return bl.Delete(
		context.Background(),
		&api.BackupLocationDeleteRequest{
			Name:  p.getBackupLocationName(params),
			OrgId: p.getOrgID(params),
		},
	)
}

func (p *portworx) CreateBackup(params *backup.Request) (*api.BackupCreateResponse, error) {
	backup := p.getBackupManager()
	req := &api.BackupCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name:   p.getBackupName(params),
			OrgId:  p.getOrgID(params),
			Owner:  p.getOwner(params),
			Labels: p.getLabels(params),
		},
		BackupLocation: p.getBackupLocationName(params),
		Cluster:        p.getClusterName(params),
		Namespaces:     p.getNameSpaces(params),
	}

	return backup.Create(context.Background(), req)
}

func (p *portworx) EnumerateBackup(params *backup.Request) (*api.BackupEnumerateResponse, error) {
	backup := p.getBackupManager()
	return backup.Enumerate(
		context.Background(),
		&api.BackupEnumerateRequest{
			OrgId: p.getOrgID(params),
		},
	)
}

func (p *portworx) InspectBackup(params *backup.Request) (*api.BackupInspectResponse, error) {
	backup := p.getBackupManager()
	return backup.Inspect(
		context.Background(),
		&api.BackupInspectRequest{
			OrgId: p.getOrgID(params),
			Name:  p.getBackupName(params),
		},
	)
}

func (p *portworx) DeleteBackup(params *backup.Request) (*api.BackupDeleteResponse, error) {
	backup := p.getBackupManager()
	return backup.Delete(
		context.Background(),
		&api.BackupDeleteRequest{
			OrgId: p.getOrgID(params),
			Name:  p.getBackupName(params),
		},
	)
}

func init() {
	backup.Register(driverName, &portworx{})
}
