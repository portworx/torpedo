// +build unittest

package storagemanager

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/pkg/parser"
	"github.com/libopenstorage/openstorage/api"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	testSpecPath = "testspecs/azure-storage-decision-matrix.yaml"
)

var (
	storageManager cloudops.StorageManager
)

type updateTestInput struct {
	expectedErr error
	request     *cloudops.StoragePoolUpdateRequest
	response    *cloudops.StoragePoolUpdateResponse
}

func TestAzureStorageManager(t *testing.T) {
	t.Run("setup", setup)
	t.Run("storageDistribution", storageDistribution)
	t.Run("storageUpdate", storageUpdate)
}

func setup(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	decisionMatrix, err := parser.NewStorageDecisionMatrixParser().UnmarshalFromYaml(testSpecPath)
	require.NoError(t, err, "Unexpected error on yaml parser")

	storageManager, err = NewAzureStorageManager(*decisionMatrix)
	require.NoError(t, err, "Unexpected error on creating Azure storage manager")
}

func storageDistribution(t *testing.T) {
	testMatrix := []struct {
		expectedErr error
		request     *cloudops.StorageDistributionRequest
		response    *cloudops.StorageDistributionResponse
	}{
		{
			// Test1: always use the upper bound on IOPS if there is no drive type
			// that provides that exact amount of requested IOPS")
			request: &cloudops.StorageDistributionRequest{
				UserStorageSpec: []*cloudops.StorageSpec{
					&cloudops.StorageSpec{
						IOPS:        1000,
						MinCapacity: 1024,
						MaxCapacity: 4096,
					},
				},
				InstanceType:     "foo",
				InstancesPerZone: 3,
				ZoneCount:        2,
			},

			response: &cloudops.StorageDistributionResponse{
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 256,
						DriveType:        "Premium_LRS",
						InstancesPerZone: 2,
						DriveCount:       1,
						IOPS:             1100,
					},
				},
			},
			expectedErr: nil,
		},
		// Test2: choose the right size of the disk by updating the instances per zone
		//        in case of a conflict with two configurations providing the same IOPS
		//        and min capacity choose based of priority
		{
			request: &cloudops.StorageDistributionRequest{
				UserStorageSpec: []*cloudops.StorageSpec{
					&cloudops.StorageSpec{
						IOPS:        500,
						MinCapacity: 1024,
						MaxCapacity: 100000,
					},
				},
				InstanceType:     "foo",
				InstancesPerZone: 3,
				ZoneCount:        3,
			},
			response: &cloudops.StorageDistributionResponse{
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 37,
						DriveType:        "Standard_LRS",
						InstancesPerZone: 3,
						DriveCount:       3,
						IOPS:             500,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// Test3: user wants 1TiB on all the nodes
			request: &cloudops.StorageDistributionRequest{
				UserStorageSpec: []*cloudops.StorageSpec{
					&cloudops.StorageSpec{
						IOPS:        5000,
						MinCapacity: 9216,
						MaxCapacity: 90000,
					},
				},
				InstanceType:     "foo",
				InstancesPerZone: 3,
				ZoneCount:        3,
			},

			response: &cloudops.StorageDistributionResponse{
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 1024,
						DriveType:        "Premium_LRS",
						InstancesPerZone: 3,
						DriveCount:       1,
						IOPS:             5000,
					},
				},
			},
			expectedErr: nil,
		},

		{
			// Test4: choose the configuration which is closest to the requested IOPS
			request: &cloudops.StorageDistributionRequest{
				UserStorageSpec: []*cloudops.StorageSpec{
					&cloudops.StorageSpec{
						IOPS:        2000,
						MinCapacity: 16384,
						MaxCapacity: 100000,
					},
				},
				InstanceType:     "foo",
				InstancesPerZone: 2,
				ZoneCount:        2,
			},
			response: &cloudops.StorageDistributionResponse{
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 8192,
						DriveType:        "StandardSSD_LRS",
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             2000,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// Test5: choose upper bound IOPS when you cannot uniformly distribute storage
			// across nodes for the provided IOPS
			request: &cloudops.StorageDistributionRequest{
				UserStorageSpec: []*cloudops.StorageSpec{
					&cloudops.StorageSpec{
						IOPS:        2000,
						MinCapacity: 16384,
						MaxCapacity: 100000,
					},
				},
				InstanceType:     "foo",
				InstancesPerZone: 2,
				ZoneCount:        3,
			},
			response: &cloudops.StorageDistributionResponse{
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 8192,
						DriveType:        "StandardSSD_LRS",
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             2000,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// Test6: reduce the number of instances per zone if the IOPS and min capacity are not met
			request: &cloudops.StorageDistributionRequest{
				UserStorageSpec: []*cloudops.StorageSpec{
					&cloudops.StorageSpec{
						IOPS:        7500,
						MinCapacity: 4096,
						MaxCapacity: 100000,
					},
				},
				InstanceType:     "foo",
				InstancesPerZone: 2,
				ZoneCount:        2,
			},
			response: &cloudops.StorageDistributionResponse{
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 2048,
						DriveType:        "Premium_LRS",
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             7500,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// Test7: if storage cannot be distributed equally across zones return an error
			request: &cloudops.StorageDistributionRequest{
				UserStorageSpec: []*cloudops.StorageSpec{
					&cloudops.StorageSpec{
						IOPS:        7500,
						MinCapacity: 2048,
						MaxCapacity: 100000,
					},
				},
				InstanceType:     "foo",
				InstancesPerZone: 3,
				ZoneCount:        3,
			},
			response: &cloudops.StorageDistributionResponse{
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 2048,
						DriveType:        "Premium_LRS",
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             7500,
					},
				},
			},

			expectedErr: nil,
		},
		{
			// Test8: Multiple user storage specs in a single request
			request: &cloudops.StorageDistributionRequest{
				UserStorageSpec: []*cloudops.StorageSpec{
					&cloudops.StorageSpec{
						IOPS:        500,
						MinCapacity: 1000,
						MaxCapacity: 100000,
						DriveType:   "StandardSSD_LRS",
					},
					&cloudops.StorageSpec{
						IOPS:        5000,
						MinCapacity: 9216,
						MaxCapacity: 90000,
					},
				},
				InstanceType:     "foo",
				InstancesPerZone: 3,
				ZoneCount:        3,
			},
			response: &cloudops.StorageDistributionResponse{
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 128,
						DriveType:        "StandardSSD_LRS",
						InstancesPerZone: 3,
						DriveCount:       1,
						IOPS:             500,
					},
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 1024,
						DriveType:        "Premium_LRS",
						InstancesPerZone: 3,
						DriveCount:       1,
						IOPS:             5000,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// Test9: Fail the request even if one of the user specs fails
			request: &cloudops.StorageDistributionRequest{
				UserStorageSpec: []*cloudops.StorageSpec{
					&cloudops.StorageSpec{
						IOPS:        500,
						MinCapacity: 10,
						MaxCapacity: 30,
					},
					&cloudops.StorageSpec{
						IOPS:        7500,
						MinCapacity: 2048,
						MaxCapacity: 100000,
					},
				},
				InstanceType:     "foo",
				InstancesPerZone: 3,
				ZoneCount:        3,
			},
			expectedErr: &cloudops.ErrStorageDistributionCandidateNotFound{},
		},
		{
			// Test10: Install with lower sized disks
			request: &cloudops.StorageDistributionRequest{
				UserStorageSpec: []*cloudops.StorageSpec{
					&cloudops.StorageSpec{
						IOPS:        300,
						MinCapacity: 150,
						MaxCapacity: 300,
					},
				},
				InstanceType:     "foo",
				InstancesPerZone: 1,
				ZoneCount:        3,
			},
			response: &cloudops.StorageDistributionResponse{
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 50,
						DriveType:        "Standard_LRS",
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             500,
					},
				},
			},
			expectedErr: nil,
		},
	}

	for j, test := range testMatrix {
		fmt.Println("Executing test case: ", j+1)
		response, err := storageManager.GetStorageDistribution(test.request)
		if test.expectedErr == nil {
			require.NoError(t, err, "Unexpected error on GetStorageDistribution")
			require.NotNil(t, response, "got nil response from GetStorageDistribution")
			require.Equal(t, len(test.response.InstanceStorage), len(response.InstanceStorage), "unequal response lengths")
			for i := range test.response.InstanceStorage {
				require.True(t, reflect.DeepEqual(*response.InstanceStorage[i], *test.response.InstanceStorage[i]),
					"Test Case %v Expected Response: %+v . Actual Response %+v", j+1,
					test.response.InstanceStorage[i], response.InstanceStorage[i])
			}
		} else {
			require.NotNil(t, err, "GetStorageDistribution should have returned an error")
			require.Equal(t, test.expectedErr, err, "received unexpected type of error")
		}
	}

}

func storageUpdate(t *testing.T) {
	testMatrix := []updateTestInput{
		{
			// ***** TEST: 1
			//        Instance has 3 x 256 GiB
			//        Update from 768GiB to 1536 GiB by resizing disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     1536,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				CurrentDriveSize:    256,
				CurrentDriveType:    "Premium_LRS",
				CurrentIOPS:         1000,
				CurrentDriveCount:   3,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 512,
						DriveType:        "Premium_LRS",
						DriveCount:       3,
						IOPS:             1100,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST: 2
			//        Instance has 2 x 350 GiB
			//        Update from 700GiB to 800 GiB by resizing disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     800,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				CurrentDriveSize:    350,
				CurrentDriveType:    "Premium_LRS",
				CurrentDriveCount:   2,
				TotalDrivesOnNode:   2,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 400,
						DriveType:        "Premium_LRS",
						DriveCount:       2,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST: 3
			//        Instance has 3 x 300 GiB
			//        Update from 900GiB to 1200 GiB by resizing disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     1200,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				CurrentDriveSize:    300,
				CurrentDriveType:    "StandardSSD_LRS",
				CurrentDriveCount:   3,
				TotalDrivesOnNode:   3,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 400,
						DriveType:        "StandardSSD_LRS",
						DriveCount:       3,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST: 4
			//		  Instances has 2 x 1024 GiB
			//        Update from 2048 GiB to  4096 GiB by adding disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     4096,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				CurrentDriveSize:    1024,
				CurrentDriveType:    "Premium_LRS",
				CurrentDriveCount:   2,
				TotalDrivesOnNode:   2,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 1024,
						DriveType:        "Premium_LRS",
						DriveCount:       2,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST: 5
			//		  Instances has 2 x 1024 GiB
			//        Update from 2048 GiB to  3072 GiB by adding disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     3072,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				CurrentDriveSize:    1024,
				CurrentDriveType:    "Standard_LRS",
				CurrentDriveCount:   2,
				TotalDrivesOnNode:   2,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 1024,
						DriveType:        "Standard_LRS",
						DriveCount:       1,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST: 6
			//		  Instances has 3 x 600 GiB
			//        Update from 1800 GiB to 2000 GiB by adding disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     2000,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				CurrentDriveSize:    600,
				CurrentDriveType:    "Premium_LRS",
				CurrentDriveCount:   3,
				TotalDrivesOnNode:   3,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 600,
						DriveType:        "Premium_LRS",
						DriveCount:       1,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST: 7
			//		  Instances has no existing drives
			//        Update from 0 GiB to 700 GiB by adding disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     700,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				TotalDrivesOnNode:   0,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 700,
						DriveType:        "Premium_LRS",
						DriveCount:       1,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST: 8
			//		  Instances has 4 x 100 GiB and one extra drive
			//        Update from 400 GiB to 800 GiB by adding disks. Should fail as we support max 8 drives
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     800,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				CurrentDriveSize:    100,
				CurrentDriveType:    "Premium_LRS",
				CurrentDriveCount:   4,
				TotalDrivesOnNode:   5,
			},
			response: nil,
			expectedErr: &cloudops.ErrStorageDistributionCandidateNotFound{
				Reason: "node has reached it's maximum supported drive count: 8",
			},
		},
		/*{
			// ***** TEST: 8
			//		  Instances has no existing drives
			//        Update from 0 GiB to 8193 GiB by adding disks. 8193 is higher
			//        than the maximum drive in the matrix
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     8196,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				TotalDrivesOnNode:   0,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 4098,
						DriveType:        "thin",
						DriveCount:       2,
					},
				},
			},
			expectedErr: nil,
		},*/
		{
			// ***** TEST: 9
			//        Instance has 1 x 150 GiB
			//        Update from 150GiB to 170 GiB by resizing disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     280,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				CurrentDriveSize:    256,
				CurrentDriveType:    "Standard_LRS",
				CurrentDriveCount:   1,
				TotalDrivesOnNode:   1,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 280,
						DriveType:        "Standard_LRS",
						DriveCount:       1,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST: 10 -> lower sized disks
			//        Instance has 1 x 200 GiB
			//        Update from 200GiB to 400 GiB by adding disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     400,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				CurrentDriveSize:    200,
				CurrentDriveType:    "Premium_LRS",
				CurrentDriveCount:   1,
				TotalDrivesOnNode:   1,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 200,
						DriveType:        "Premium_LRS",
						DriveCount:       1,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST: 11 -> ask for one more GiB
			//        Instance has 2 x 200 GiB
			//        Update from 400 GiB to 401 GiB by adding disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     401,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				CurrentDriveSize:    200,
				CurrentDriveType:    "Premium_LRS",
				CurrentDriveCount:   2,
				TotalDrivesOnNode:   2,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 200,
						DriveType:        "Premium_LRS",
						DriveCount:       1,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST: 12 instance is already at higher capacity than requested
			//        Instance has 3 x 200 GiB
			//        Update from 600 GiB to 401 GiB by adding disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     401,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				CurrentDriveSize:    200,
				CurrentDriveType:    "Premium_LRS",
				CurrentDriveCount:   3,
				TotalDrivesOnNode:   3,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage:     nil,
			},
			expectedErr: &cloudops.ErrCurrentCapacityHigherThanDesired{Current: 600, Desired: 401},
		},
		{
			// ***** TEST: 13 instance is already at higher capacity than requested
			//        Instance has 2 x 105 GiB
			//        Update from 210 GiB to 215 GiB by adding disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     215,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				CurrentDriveSize:    105,
				CurrentDriveType:    "Premium_LRS",
				CurrentDriveCount:   2,
				TotalDrivesOnNode:   2,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 108,
						DriveType:        "Premium_LRS",
						DriveCount:       2,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST: 14 delta is a float number 2.1 per disk (should be rounded up to 3)
			//        Instance has 9 x 28 GiB
			//        Update from 252 GiB to 271 GiB by adding disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     271,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				CurrentDriveSize:    28,
				CurrentDriveType:    "Premium_LRS",
				CurrentDriveCount:   9,
				TotalDrivesOnNode:   9,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 31,
						DriveType:        "Premium_LRS",
						DriveCount:       9,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST 15: delta is a float number 53.6 per disk (should be rounded up to 54)
			//        Instance has 5 x 137 GiB
			//        Update from 685 GiB to 953 GiB by adding disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     953,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				CurrentDriveSize:    137,
				CurrentDriveType:    "Premium_LRS",
				CurrentDriveCount:   5,
				TotalDrivesOnNode:   5,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 191,
						DriveType:        "Premium_LRS",
						DriveCount:       5,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST: 16
			//		  Instances has 2 x 5 GiB
			//        Update from 10 GiB to  15 GiB by adding disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     15,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				CurrentDriveSize:    5,
				CurrentDriveType:    "Premium_LRS",
				CurrentDriveCount:   2,
				TotalDrivesOnNode:   2,
			},
			response: nil,
			expectedErr: &cloudops.ErrStorageDistributionCandidateNotFound{
				Reason: "found no candidates for adding a new disk of existing size: 5 GiB. Only drives in following " +
					"size ranges are supported: [[1024 GiB -> 8192 GiB (Premium_LRS)] [128 GiB -> 256 GiB (Premium_LRS)]" +
					" [16384 GiB -> 32768 GiB (Premium_LRS)] [2048 GiB -> 8192 GiB (Premium_LRS)] [256 GiB -> 8192 GiB (Premium_LRS)] [32 GiB -> 64 GiB (Premium_LRS)] [32768 GiB -> 32768 GiB (Premium_LRS)] [512 GiB -> 8192 GiB (Premium_LRS)] [64 GiB -> 128 GiB (Premium_LRS)] [8192 GiB -> 16384 GiB (Premium_LRS)]]",
			},
		},
	}

	for i, test := range testMatrix {
		fmt.Println("Executing test case: ", i+1)
		response, err := storageManager.RecommendStoragePoolUpdate(test.request)
		if test.expectedErr == nil {
			require.Nil(t, err, "RecommendStoragePoolUpdate returned an error")
			require.NotNil(t, response, "RecommendStoragePoolUpdate returned empty response")
			require.Equal(t, len(test.response.InstanceStorage), len(response.InstanceStorage), "length of expected and actual response not equal")
			// ensure response contains test.response
			for _, instStorage := range response.InstanceStorage {
				matched := false
				for _, expectedInstStorage := range test.response.InstanceStorage {
					matched = (expectedInstStorage.DriveCapacityGiB == instStorage.DriveCapacityGiB) &&
						(expectedInstStorage.DriveType == instStorage.DriveType) &&
						(expectedInstStorage.DriveCount == instStorage.DriveCount)

					if expectedInstStorage.IOPS > 0 {
						matched = matched && (expectedInstStorage.IOPS >= instStorage.IOPS)
					}

					if matched {
						break
					}

				}
				require.True(t, matched, fmt.Sprintf("response didn't match. expected: %v actual: %v", test.response.InstanceStorage[0], response.InstanceStorage[0]))
			}
		} else {
			require.NotNil(t, err, "RecommendInstanceStorageUpdate should have returned an error")
			require.Equal(t, test.expectedErr.Error(), err.Error(), "received unexpected type of error")
		}
	}
}

func logUpdateTestInput(test updateTestInput) {
	logrus.Infof("### RUNNING TEST")
	logrus.Infof("### REQUEST:  new capacity: %d GiB op_type: %v",
		test.request.DesiredCapacity, test.request.ResizeOperationType)
	logrus.Infof("### RESPONSE: op_type: %v", test.response.ResizeOperationType)
	for _, responseInstStorage := range test.response.InstanceStorage {
		logrus.Infof("              instStorage: %d X %d GiB %s drives", responseInstStorage.DriveCount,
			responseInstStorage.DriveCapacityGiB, responseInstStorage.DriveType)
	}
}
