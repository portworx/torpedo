package storagemanager

import (
	"fmt"

	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/pkg/storagedistribution"
	"github.com/libopenstorage/cloudops/unsupported"
)

type awsStorageManager struct {
	cloudops.StorageManager
	decisionMatrix *cloudops.StorageDecisionMatrix
}

const (
	// DriveTypeGp3 is a constant for gp3 drive types
	DriveTypeGp3 = "gp3"
	// DriveTypeGp2 is a constant for gp2 drive types
	DriveTypeGp2 = "gp2"
	// DriveTypeIo1 is a constant for io1 drive types
	DriveTypeIo1 = "io1"
	// Gp2IopsMultiplier is the amount with which a given gp2 GiB size is multiplied
	// in order to get that drive's baseline IOPS performance
	Gp2IopsMultiplier = 3
)

// NewAWSStorageManager returns an aws implementation for Storage Management
func NewAWSStorageManager(
	decisionMatrix cloudops.StorageDecisionMatrix,
) (cloudops.StorageManager, error) {
	return &awsStorageManager{
		StorageManager: unsupported.NewUnsupportedStorageManager(),
		decisionMatrix: &decisionMatrix}, nil
}

func (a *awsStorageManager) GetStorageDistribution(
	request *cloudops.StorageDistributionRequest,
) (*cloudops.StorageDistributionResponse, error) {
	response := &cloudops.StorageDistributionResponse{}
	for _, userRequest := range request.UserStorageSpec {
		// for for request, find how many instances per zone needs to have storage
		// and the storage spec for each of them
		instStorage, instancePerZone, row, err :=
			storagedistribution.GetStorageDistributionForPool(
				a.decisionMatrix,
				userRequest,
				request.InstancesPerZone,
				request.ZoneCount,
			)
		if err != nil {
			return nil, err
		}
		response.InstanceStorage = append(
			response.InstanceStorage,
			&cloudops.StoragePoolSpec{
				DriveCapacityGiB: instStorage.DriveCapacityGiB,
				DriveType:        instStorage.DriveType,
				InstancesPerZone: instancePerZone,
				DriveCount:       instStorage.DriveCount,
				IOPS:             determineIOPSForPool(instStorage, row, userRequest.IOPS),
			},
		)

	}
	return response, nil
}

func (a *awsStorageManager) RecommendStoragePoolUpdate(
	request *cloudops.StoragePoolUpdateRequest) (*cloudops.StoragePoolUpdateResponse, error) {
	resp, row, err := storagedistribution.GetStorageUpdateConfig(request, a.decisionMatrix)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.InstanceStorage) != 1 {
		return nil, fmt.Errorf("could not find a valid instance storage object")
	}
	resp.InstanceStorage[0].IOPS = determineIOPSForPool(resp.InstanceStorage[0], row, request.CurrentIOPS /*we do not support updating IOPS yet*/)
	return resp, nil
}

func determineIOPSForPool(instStorage *cloudops.StoragePoolSpec, row *cloudops.StorageDecisionMatrixRow, currentIOPS uint64) uint64 {
	if instStorage.DriveType == DriveTypeGp2 {
		return instStorage.DriveCapacityGiB * Gp2IopsMultiplier
	} else if instStorage.DriveType == DriveTypeIo1 {
		// For io1 volumes we need to specify the requested iops as the provisioned iops
		return currentIOPS
	}
	return row.MinIOPS
}

func init() {
	cloudops.RegisterStorageManager(cloudops.AWS, NewAWSStorageManager)
}
