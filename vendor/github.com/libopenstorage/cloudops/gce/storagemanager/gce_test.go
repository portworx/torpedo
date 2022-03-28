package storagemanager

import (
	"fmt"
	"github.com/libopenstorage/openstorage/api"
	"reflect"
	"testing"

	"github.com/libopenstorage/cloudops"
	"github.com/libopenstorage/cloudops/pkg/parser"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	testSpecPath = "testspecs/gce.yaml"
)

var (
	storageManager cloudops.StorageManager
)

type updateTestInput struct {
	expectedErr error
	request     *cloudops.StoragePoolUpdateRequest
	response    *cloudops.StoragePoolUpdateResponse
}

func TestGCEStorageManager(t *testing.T) {
	t.Run("setup", setup)
	t.Run("storageDistribution", storageDistribution)
	t.Run("storageUpdate", storageUpdate)
}

func setup(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	decisionMatrix, err := parser.NewStorageDecisionMatrixParser().UnmarshalFromYaml(testSpecPath)
	require.NoError(t, err, "Unexpected error on yaml parser")

	storageManager, err = NewStorageManager(*decisionMatrix)
	require.NoError(t, err, "Unexpected error on creating GCE storage manager")
}

func storageDistribution(t *testing.T) {
	testMatrix := []struct {
		expectedErr error
		request     *cloudops.StorageDistributionRequest
		response    *cloudops.StorageDistributionResponse
	}{
		{
			// Test1: chose the right size of disk giving the highest priority to IOPS
			request: &cloudops.StorageDistributionRequest{
				UserStorageSpec: []*cloudops.StorageSpec{
					&cloudops.StorageSpec{
						IOPS:        1000,
						MinCapacity: 1024,
						MaxCapacity: 4096,
					},
				},
				InstanceType:     "foo",
				InstancesPerZone: 1,
				ZoneCount:        1,
			},

			response: &cloudops.StorageDistributionResponse{
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 1267,
						DriveType:        DriveTypeStandard,
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             951,
					},
				},
			},
			expectedErr: nil,
		},
		// Test2: choose the right size of the disk by updating the instances per zone
		//        in case of a conflict with two configurations providing the same IOPS
		//        and min capacity choose based of priority.
		{
			request: &cloudops.StorageDistributionRequest{
				UserStorageSpec: []*cloudops.StorageSpec{
					&cloudops.StorageSpec{
						IOPS:        500,
						MinCapacity: 1024,
						MaxCapacity: 100000,
					},
				},
				InstanceType:     DriveTypeStandard,
				InstancesPerZone: 3,
				ZoneCount:        3,
			},
			response: &cloudops.StorageDistributionResponse{
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 600,
						DriveType:        DriveTypeStandard,
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             450,
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
						IOPS:        2900,
						MinCapacity: 9216,
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
						DriveCapacityGiB: 3800,
						DriveType:        DriveTypeStandard,
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             2850,
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
						IOPS:        5700,
						MinCapacity: 8000,
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
						DriveCapacityGiB: 7534,
						DriveType:        DriveTypeStandard,
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             5651,
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
						IOPS:        800,
						MinCapacity: 2096,
						MaxCapacity: 10000,
					},
				},
				InstanceType:     "foo",
				InstancesPerZone: 2,
				ZoneCount:        3,
			},
			response: &cloudops.StorageDistributionResponse{
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 1000,
						DriveType:        DriveTypeStandard,
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             750,
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
						DriveCapacityGiB: 9934,
						DriveType:        DriveTypeStandard,
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             7451,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// Test7: provision an io1 drive if the IOPS is not achievable
			// by the provided size
			request: &cloudops.StorageDistributionRequest{
				UserStorageSpec: []*cloudops.StorageSpec{
					&cloudops.StorageSpec{
						IOPS:        7500,
						MinCapacity: 1000,
						MaxCapacity: 2000,
						DriveType:   genDriveType(DriveTypeSSD),
					},
				},
				InstanceType:     "foo",
				InstancesPerZone: 3,
				ZoneCount:        3,
			},
			response: &cloudops.StorageDistributionResponse{
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 249,
						DriveType:        genDriveType(DriveTypeSSD),
						InstancesPerZone: 2,
						DriveCount:       1,
						IOPS:             7470,
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
						DriveType:   genDriveType(DriveTypeStandard),
					},
					&cloudops.StorageSpec{
						IOPS:        5000,
						MinCapacity: 9216,
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
						DriveCapacityGiB: 600,
						DriveType:        genDriveType(DriveTypeStandard),
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             450,
					},
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 6600,
						DriveType:        DriveTypeStandard, // no drive type in this request, so response contains pd-standard only, url will be generated later in porx
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             4950,
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
						DriveCapacityGiB: 10,
						DriveType:        DriveTypeSSD,
						InstancesPerZone: 1,
						DriveCount:       5,
						IOPS:             300,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// Test11: Install for specific drive type
			request: &cloudops.StorageDistributionRequest{
				UserStorageSpec: []*cloudops.StorageSpec{
					&cloudops.StorageSpec{
						IOPS:        2500,
						MinCapacity: 150,
						MaxCapacity: 300,
						DriveType:   genDriveType(DriveTypeSSD),
					},
				},
				InstanceType:     "foo",
				InstancesPerZone: 1,
				ZoneCount:        3,
			},
			response: &cloudops.StorageDistributionResponse{
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 82,
						DriveType:        genDriveType(DriveTypeSSD),
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             2460,
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
				CurrentDriveType:    genDriveType(DriveTypeStandard),
				CurrentIOPS:         192,
				CurrentDriveCount:   3,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 512,
						DriveType:        genDriveType(DriveTypeStandard),
						DriveCount:       3,
						IOPS:             384,
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
				CurrentDriveType:    genDriveType(DriveTypeStandard),
				CurrentDriveCount:   2,
				TotalDrivesOnNode:   2,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 400,
						DriveType:        genDriveType(DriveTypeStandard),
						DriveCount:       2,
						IOPS:             300,
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
				CurrentDriveType:    genDriveType(DriveTypeStandard),
				CurrentDriveCount:   3,
				TotalDrivesOnNode:   3,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 400,
						DriveType:        genDriveType(DriveTypeStandard),
						DriveCount:       3,
						IOPS:             300,
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
				CurrentDriveType:    genDriveType(DriveTypeStandard),
				CurrentDriveCount:   2,
				TotalDrivesOnNode:   2,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 1024,
						DriveType:        genDriveType(DriveTypeStandard),
						DriveCount:       2,
						IOPS:             768,
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
				CurrentDriveType:    genDriveType(DriveTypeStandard),
				CurrentDriveCount:   2,
				TotalDrivesOnNode:   2,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 1024,
						DriveType:        genDriveType(DriveTypeStandard),
						DriveCount:       1,
						IOPS:             768,
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
				CurrentDriveType:    genDriveType(DriveTypeStandard),
				CurrentDriveCount:   3,
				TotalDrivesOnNode:   3,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 600,
						DriveType:        genDriveType(DriveTypeStandard),
						DriveCount:       1,
						IOPS:             450,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST: 7
			// 		  Instances has no existing drives
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
						DriveType:        DriveTypeStandard, // don't return a drive type as url, AddDrive function will manage this later
						DriveCount:       1,
						IOPS:             525,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST: 8
			//        Instance has 1 x 256 GiB
			//        Update from 256GiB to 280 GiB by resizing disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     280,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				CurrentDriveSize:    256,
				CurrentDriveType:    genDriveType(DriveTypeStandard),
				CurrentDriveCount:   1,
				TotalDrivesOnNode:   1,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 280,
						DriveType:        genDriveType(DriveTypeStandard),
						DriveCount:       1,
						IOPS:             210,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// ***** TEST: 9 -> lower sized disks
			//        Instance has 1 x 200 GiB
			//        Update from 200GiB to 400 GiB by adding disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     400,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				CurrentDriveSize:    200,
				CurrentDriveType:    genDriveType(DriveTypeStandard),
				CurrentDriveCount:   1,
				TotalDrivesOnNode:   1,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 200,
						DriveType:        genDriveType(DriveTypeStandard),
						DriveCount:       1,
						IOPS:             150,
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
				CurrentDriveType:    genDriveType(DriveTypeStandard),
				CurrentDriveCount:   2,
				TotalDrivesOnNode:   2,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 200,
						DriveType:        genDriveType(DriveTypeStandard),
						DriveCount:       1,
						IOPS:             150,
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
				CurrentDriveType:    genDriveType(DriveTypeStandard),
				CurrentDriveCount:   3,
				TotalDrivesOnNode:   3,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage:     nil,
			},
			expectedErr: &cloudops.ErrCurrentCapacityHigherThanDesired{Current: 600, Desired: 401},
		},
	}

	for j, test := range testMatrix {
		fmt.Println("Executing test case: ", j+1)
		response, err := storageManager.RecommendStoragePoolUpdate(test.request)

		if test.expectedErr == nil {
			require.Nil(t, err, "RecommendStoragePoolUpdate returned an error")
			require.NotNil(t, response, "RecommendStoragePoolUpdate returned empty response")
			require.Equal(t, len(test.response.InstanceStorage), len(response.InstanceStorage), "length of expected and actual response not equal")
			for i := range test.response.InstanceStorage {
				require.True(t, reflect.DeepEqual(*response.InstanceStorage[i], *test.response.InstanceStorage[i]),
					"Test Case %v Expected Response: %+v . Actual Response %+v", j+1,
					test.response.InstanceStorage[i], response.InstanceStorage[i])
			}
		} else {
			require.NotNil(t, err, "RecommendInstanceStorageUpdate should have returned an error")
			require.Equal(t, test.expectedErr.Error(), err.Error(), "received unexpected type of error")
		}
	}
}

func genDriveType(dType string) string {
	// the gce drive path comes as  https://www.googleapis.com/compute/v1/projects/portworx-eng/zones/us-east1-b/diskTypes/pd-standard
	// or  https://www.googleapis.com/compute/v1/projects/portworx-eng/zones/us-east1-b/diskTypes/pd-ssd
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/portworx-eng/zones/us-east1-b/diskTypes/%s", dType)
}
