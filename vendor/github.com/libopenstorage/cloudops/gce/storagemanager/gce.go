package storagemanager

import (
	"fmt"
	"math"
	"strings"

	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/pkg/storagedistribution"
	"github.com/libopenstorage/cloudops/unsupported"
)

type gceStorageManager struct {
	cloudops.StorageManager
	decisionMatrix *cloudops.StorageDecisionMatrix
}

const (
	// DriveTypeStandard is a constant for standard drive types
	DriveTypeStandard = "pd-standard"
	// DriveTypeSSD is a constant for ssd drive types
	DriveTypeSSD = "pd-ssd"
	// StandardIopsMultiplier is the amount with which a given pd-standard GiB size is multiplied
	// in order to get that drive's baseline IOPS performance
	StandardIopsMultiplier = 0.75
	// SSDIopsMultiplier is the amount with which a given ssd GiB size is multiplied
	// in order to get that drive's baseline IOPS performance
	SSDIopsMultiplier = 30
)

// NewStorageManager returns a GCE specific implementation of StorageManager interface.
func NewStorageManager(decisionMatrix cloudops.StorageDecisionMatrix) (cloudops.StorageManager, error) {
	return &gceStorageManager{
		StorageManager: unsupported.NewUnsupportedStorageManager(),
		decisionMatrix: &decisionMatrix}, nil
}

func (g *gceStorageManager) GetStorageDistribution(request *cloudops.StorageDistributionRequest) (*cloudops.StorageDistributionResponse, error) {
	response := &cloudops.StorageDistributionResponse{}
	for _, userRequest := range request.UserStorageSpec {
		// this hack is required because the gce drive type comes as urls:
		// https://www.googleapis.com/compute/v1/projects/portworx-eng/zones/us-east1-b/diskTypes/pd-standard
		// or  https://www.googleapis.com/compute/v1/projects/portworx-eng/zones/us-east1-b/diskTypes/pd-ssd
		var currentDriveType string
		if userRequest.DriveType != "" {
			// using the last part of drive type url for the StorageDistribution algorithm. Original type url is stored to be returned in response
			currentDriveType = userRequest.DriveType
			split := strings.Split(userRequest.DriveType, "/")
			userRequest.DriveType = split[len(split)-1]
		} else {
			currentDriveType = userRequest.DriveType
		}

		// for request, find how many instances per zone needs to have storage
		// and the storage spec for each of them
		instStorage, instancePerZone, row, err :=
			storagedistribution.GetStorageDistributionForPool(
				g.decisionMatrix,
				userRequest,
				request.InstancesPerZone,
				request.ZoneCount,
			)
		if err != nil {
			return nil, err
		}
		if currentDriveType == "" {
			currentDriveType = instStorage.DriveType
		}
		response.InstanceStorage = append(
			response.InstanceStorage,
			&cloudops.StoragePoolSpec{
				DriveCapacityGiB: instStorage.DriveCapacityGiB,
				DriveType:        currentDriveType,
				InstancesPerZone: instancePerZone,
				DriveCount:       instStorage.DriveCount,
				IOPS:             determineIOPSForPool(instStorage, row),
			},
		)

	}
	return response, nil
}

func (g *gceStorageManager) RecommendStoragePoolUpdate(request *cloudops.StoragePoolUpdateRequest) (*cloudops.StoragePoolUpdateResponse, error) {
	// this hack is required because the gce drive type comes as urls:
	// https://www.googleapis.com/compute/v1/projects/portworx-eng/zones/us-east1-b/diskTypes/pd-standard
	// or  https://www.googleapis.com/compute/v1/projects/portworx-eng/zones/us-east1-b/diskTypes/pd-ssd
	var currentDriveType string
	if request.CurrentDriveType != "" {
		currentDriveType = request.CurrentDriveType
		split := strings.Split(request.CurrentDriveType, "/")
		request.CurrentDriveType = split[len(split)-1]
	}

	resp, row, err := storagedistribution.GetStorageUpdateConfig(request, g.decisionMatrix)
	if err != nil {
		return nil, err
	}
	if resp == nil || len(resp.InstanceStorage) != 1 {
		return nil, fmt.Errorf("could not find a valid instance storage object")
	}
	resp.InstanceStorage[0].IOPS = determineIOPSForPool(resp.InstanceStorage[0], row)
	if currentDriveType != "" {
		resp.InstanceStorage[0].DriveType = currentDriveType
	}

	return resp, nil
}

func determineIOPSForPool(instStorage *cloudops.StoragePoolSpec, row *cloudops.StorageDecisionMatrixRow) uint64 {
	if instStorage.DriveType == DriveTypeStandard {
		return uint64(math.Ceil(float64(instStorage.DriveCapacityGiB) * StandardIopsMultiplier))
	} else if instStorage.DriveType == DriveTypeSSD {
		return uint64(math.Ceil(float64(instStorage.DriveCapacityGiB) * SSDIopsMultiplier))
	}
	return row.MinIOPS
}
func init() {
	cloudops.RegisterStorageManager(cloudops.GCE, NewStorageManager)
}
