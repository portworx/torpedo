package ibm

import (
	"fmt"
	"time"

	bluemix "github.com/IBM-Cloud/bluemix-go"
	v2 "github.com/IBM-Cloud/bluemix-go/api/container/containerv2"
	"github.com/IBM-Cloud/bluemix-go/session"
	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/backoff"
	"github.com/libopenstorage/cloudops/unsupported"
)

const (
	labelWorkerPoolName = "ibm-cloud.kubernetes.io/worker-pool-name"
	labelWorkerPoolID   = "ibm-cloud.kubernetes.io/worker-pool-id"
	vpcProviderName     = "vpc-gen2"
)

type ibmOps struct {
	cloudops.Compute
	cloudops.Storage
	ibmClusterClient v2.ContainerServiceAPI
	inst             *instance
}

// instance stores the metadata of the running ibm instance
type instance struct {
	name            string
	hostname        string
	zone            string
	region          string
	resourceGroup   string
	clusterName     string
	clusterLocation string
	nodePoolID      string
}

// NewClient creates a new IBM operations client
func NewClient() (cloudops.Ops, error) {

	var i = new(instance)

	c := new(bluemix.Config)

	sess, err := session.New(c)
	if err != nil {
		return nil, fmt.Errorf("failed to get session. error: [%v]", err)
	}

	ibmClusterClient, err := v2.New(sess)
	if err != nil {
		return nil, fmt.Errorf("failed to get ibm cluster client. error: [%v]", err)
	}

	instanceName, clusterName, resourceGroup, err := getInfoFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster info. error: [%v] ", err)
	}

	i.name = instanceName
	i.clusterName = clusterName
	i.resourceGroup = resourceGroup

	return backoff.NewExponentialBackoffOps(
		&ibmOps{
			Compute:          unsupported.NewUnsupportedCompute(),
			Storage:          unsupported.NewUnsupportedStorage(),
			ibmClusterClient: ibmClusterClient,
			inst:             i,
		},
		isExponentialError,
		backoff.DefaultExponentialBackoff,
	), nil
}

func (i *ibmOps) Name() string {
	return string(cloudops.IBM)
}

func (i *ibmOps) InstanceID() string {
	return i.inst.name
}

func isExponentialError(err error) bool {
	return true
}

func (i *ibmOps) InspectInstance(instanceID string) (*cloudops.InstanceInfo, error) {
	target := v2.ClusterTargetHeader{
		ResourceGroup: i.inst.resourceGroup,
		Provider:      vpcProviderName,
	}
	workerDetails, err := i.ibmClusterClient.Workers().Get(i.inst.clusterName, instanceID, target)
	if err != nil {
		return nil, err
	}

	instanceInfo := &cloudops.InstanceInfo{
		CloudResourceInfo: cloudops.CloudResourceInfo{
			Name: workerDetails.ID,
			Labels: map[string]string{
				labelWorkerPoolName: workerDetails.PoolName,
				labelWorkerPoolID:   workerDetails.PoolID,
			},
			Zone:   workerDetails.Location,
			Region: workerDetails.Location,
		},
	}
	return instanceInfo, nil
}

func (i *ibmOps) InspectInstanceGroupForInstance(instanceID string) (*cloudops.InstanceGroupInfo, error) {
	instanceInfo, err := i.InspectInstance(instanceID)
	if err != nil {
		return nil, err
	}
	var instGroupInfo *cloudops.InstanceGroupInfo
	if workerPoolID, ok := instanceInfo.Labels[labelWorkerPoolID]; ok {
		target := v2.ClusterTargetHeader{
			ResourceGroup: i.inst.resourceGroup,
			Provider:      vpcProviderName,
		}
		workerPoolDetails, err := i.ibmClusterClient.WorkerPools().GetWorkerPool(i.inst.clusterName, workerPoolID, target)
		if err != nil {
			return nil, err
		}

		var zones []string
		for _, z := range workerPoolDetails.Zones {
			zones = append(zones, z.ID)
		}

		instGroupInfo = &cloudops.InstanceGroupInfo{
			CloudResourceInfo: cloudops.CloudResourceInfo{
				Name:   workerPoolDetails.PoolName,
				ID:     workerPoolDetails.ID,
				Labels: workerPoolDetails.Labels,
			},
			Zones: zones,
		}

		return instGroupInfo, nil
	}
	return instGroupInfo, fmt.Errorf("no [%s] label found for instance [%s]", labelWorkerPoolID, instanceID)
}

// IsDevMode checks if the pkg is invoked in
// developer mode where IBM credentials are set as env variables
func IsDevMode() bool {
	_, _, _, err := getInfoFromEnv()
	if err != nil {
		return false
	}
	return true
}

func getInfoFromEnv() (string, string, string, error) {
	instanceName, err := cloudops.GetEnvValueStrict("IBM_INSTANCE_NAME")
	if err != nil {
		return "", "", "", err
	}

	clusterName, err := cloudops.GetEnvValueStrict("IBM_CLUSTER_NAME")
	if err != nil {
		return "", "", "", err
	}

	resourceGroup, err := cloudops.GetEnvValueStrict("IBM_RESOURCE_GROUP")
	if err != nil {
		return "", "", "", err
	}
	return instanceName, clusterName, resourceGroup, nil
}

// SetInstanceGroupSize sets node count for a instance group.
// Count here is per availability zone
func (i *ibmOps) SetInstanceGroupSize(instanceGroupID string,
	count int64, timeout time.Duration) error {

	req := v2.ResizeWorkerPoolReq{
		Cluster:    i.inst.clusterName,
		Workerpool: instanceGroupID,
		Size:       count,
	}
	target := v2.ClusterTargetHeader{
		ResourceGroup: i.inst.resourceGroup,
		Provider:      vpcProviderName,
	}
	err := i.ibmClusterClient.WorkerPools().ResizeWorkerPool(req, target)
	if err != nil {
		return err
	}
	return nil
}

// GetInstanceGroupSize returns current node count of given instance group
func (i *ibmOps) GetInstanceGroupSize(instanceGroupID string) (int64, error) {
	target := v2.ClusterTargetHeader{
		ResourceGroup: i.inst.resourceGroup,
		Provider:      vpcProviderName,
	}
	workerPoolDetails, err := i.ibmClusterClient.WorkerPools().GetWorkerPool(i.inst.clusterName, instanceGroupID, target)
	if err != nil {
		return 0, err
	}
	return int64(workerPoolDetails.WorkerCount * len(workerPoolDetails.Zones)), nil
}
