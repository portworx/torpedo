package ibm

import (
	"fmt"
	"time"

	bluemix "github.com/IBM-Cloud/bluemix-go"
	v2 "github.com/IBM-Cloud/bluemix-go/api/container/containerv2"
	"github.com/IBM-Cloud/bluemix-go/session"
	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/unsupported"
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

	return &ibmOps{
		Compute:          unsupported.NewUnsupportedCompute(),
		Storage:          unsupported.NewUnsupportedStorage(),
		ibmClusterClient: ibmClusterClient,
		inst:             i,
	}, nil
	/*return backoff.NewExponentialBackoffOps(
		&ibmOps{
			Compute:          unsupported.NewUnsupportedCompute(),
			Storage:          unsupported.NewUnsupportedStorage(),
			ibmClusterClient: ibmClusterClient,
			inst:             i,
		},
		isExponentialError,
		backoff.DefaultExponentialBackoff,
	), nil
	*/
}

func (i *ibmOps) Name() string { return string(cloudops.IBM) }

func (i *ibmOps) InstanceID() string { return i.inst.name }

func isExponentialError(err error) bool {
	// TODO: revisit
	return true
}

func (i *ibmOps) InspectInstance(instanceID string) (*cloudops.InstanceInfo, error) {
	target := v2.ClusterTargetHeader{
		ResourceGroup: i.inst.resourceGroup,
		Provider:      "vpc-gen2",
	}
	fmt.Printf("target Header: %+v, cluter: %+v", target, i.inst.clusterName)
	workerDetails, err := i.ibmClusterClient.Workers().Get(i.inst.clusterName, instanceID, target)
	if err != nil {
		return nil, err
	}

	instanceInfo := &cloudops.InstanceInfo{
		CloudResourceInfo: cloudops.CloudResourceInfo{
			Name: workerDetails.ID,
			Labels: map[string]string{
				"ibm-cloud.kubernetes.io/worker-pool-name": workerDetails.PoolName,
				"ibm-cloud.kubernetes.io/worker-pool-id":   workerDetails.PoolID,
			},
			Zone:   workerDetails.Location,
			Region: workerDetails.Location,
		},
	}
	return instanceInfo, nil
}

func (i *ibmOps) InspectInstanceGroupForInstance(instanceID string) (*cloudops.InstanceGroupInfo, error) {
	return nil, nil
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
		Provider:      "vpc-gen2",
	}
	err := i.ibmClusterClient.WorkerPools().ResizeWorkerPool(req, target)
	if err != nil {
		return err
	}

	// TODO: Wait fot resize operation to complete.
	return nil
}
