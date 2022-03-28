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
	testSpecPath = "testspecs/aws.yaml"
)

var (
	storageManager cloudops.StorageManager
)

type updateTestInput struct {
	expectedErr error
	request     *cloudops.StoragePoolUpdateRequest
	response    *cloudops.StoragePoolUpdateResponse
}

func TestAWSStorageManager(t *testing.T) {
	t.Run("setup", setup)
	t.Run("storageDistribution", storageDistribution)
	t.Run("storageUpdate", storageUpdate)
}

func setup(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	decisionMatrix, err := parser.NewStorageDecisionMatrixParser().UnmarshalFromYaml(testSpecPath)
	require.NoError(t, err, "Unexpected error on yaml parser")

	storageManager, err = NewAWSStorageManager(*decisionMatrix)
	require.NoError(t, err, "Unexpected error on creating AWS storage manager")
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
						DriveCapacityGiB: 316,
						DriveType:        "gp2",
						InstancesPerZone: 2,
						DriveCount:       1,
						IOPS:             948,
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
						DriveCapacityGiB: 150,
						DriveType:        "gp2",
						InstancesPerZone: 3,
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
						DriveCapacityGiB: 1024,
						DriveType:        "gp2",
						InstancesPerZone: 3,
						DriveCount:       1,
						IOPS:             3072,
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
						DriveCapacityGiB: 2000,
						DriveType:        "gp2",
						InstancesPerZone: 2,
						DriveCount:       1,
						IOPS:             6000,
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
						DriveCapacityGiB: 349,
						DriveType:        "gp2",
						InstancesPerZone: 2,
						DriveCount:       1,
						IOPS:             1047,
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
						DriveCapacityGiB: 2483,
						DriveType:        "gp2",
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             7449,
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
						DriveType:   "io1",
					},
				},
				InstanceType:     "foo",
				InstancesPerZone: 3,
				ZoneCount:        3,
			},
			response: &cloudops.StorageDistributionResponse{
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 200,
						DriveType:        "io1",
						InstancesPerZone: 2,
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
						DriveType:   "gp2",
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
						DriveCapacityGiB: 150,
						DriveType:        "gp2",
						InstancesPerZone: 3,
						DriveCount:       1,
						IOPS:             450,
					},
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 1650,
						DriveType:        "gp2",
						InstancesPerZone: 2,
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
						IOPS:        10000,
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
						DriveCapacityGiB: 83,
						DriveType:        "gp2",
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             249,
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
						DriveType:   "io1",
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
						DriveType:        "io1",
						InstancesPerZone: 1,
						DriveCount:       1,
						IOPS:             2500,
					},
				},
			},
			expectedErr: nil,
		},
		{
			// Test12: "happy-path" test for GP3
			request: &cloudops.StorageDistributionRequest{
				UserStorageSpec: []*cloudops.StorageSpec{
					&cloudops.StorageSpec{
						IOPS:        3000,
						MinCapacity: 9216,
						MaxCapacity: 100000,
						DriveType:   "gp3",
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
						DriveType:        "gp3",
						InstancesPerZone: 3,
						DriveCount:       8,
						IOPS:             3000,
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
				CurrentDriveType:    "gp2",
				CurrentIOPS:         768,
				CurrentDriveCount:   3,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 512,
						DriveType:        "gp2",
						DriveCount:       3,
						IOPS:             1536,
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
				CurrentDriveType:    "gp2",
				CurrentDriveCount:   2,
				TotalDrivesOnNode:   2,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 400,
						DriveType:        "gp2",
						DriveCount:       2,
						IOPS:             1200,
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
				CurrentDriveType:    "gp2",
				CurrentDriveCount:   3,
				TotalDrivesOnNode:   3,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 400,
						DriveType:        "gp2",
						DriveCount:       3,
						IOPS:             1200,
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
				CurrentDriveType:    "gp2",
				CurrentDriveCount:   2,
				TotalDrivesOnNode:   2,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 1024,
						DriveType:        "gp2",
						DriveCount:       2,
						IOPS:             3072,
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
				CurrentDriveType:    "gp2",
				CurrentDriveCount:   2,
				TotalDrivesOnNode:   2,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 1024,
						DriveType:        "gp2",
						DriveCount:       1,
						IOPS:             3072,
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
				CurrentDriveType:    "gp2",
				CurrentDriveCount:   3,
				TotalDrivesOnNode:   3,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 600,
						DriveType:        "gp2",
						DriveCount:       1,
						IOPS:             1800,
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
						DriveType:        "gp2",
						DriveCount:       1,
						IOPS:             2100,
					},
				},
			},
			expectedErr: nil,
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
				CurrentDriveType:    "gp2",
				CurrentDriveCount:   1,
				TotalDrivesOnNode:   1,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 280,
						DriveType:        "gp2",
						DriveCount:       1,
						IOPS:             840,
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
				CurrentDriveType:    "gp2",
				CurrentDriveCount:   1,
				TotalDrivesOnNode:   1,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 200,
						DriveType:        "gp2",
						DriveCount:       1,
						IOPS:             600,
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
				CurrentDriveType:    "gp2",
				CurrentDriveCount:   2,
				TotalDrivesOnNode:   2,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_ADD_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					&cloudops.StoragePoolSpec{
						DriveCapacityGiB: 200,
						DriveType:        "gp2",
						DriveCount:       1,
						IOPS:             600,
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
				CurrentDriveType:    "gp2",
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
			// ***** TEST: 13
			//        GP3 Instance has 2 x 350 GiB
			//        Update from 700GiB to 800 GiB by resizing disks
			request: &cloudops.StoragePoolUpdateRequest{
				DesiredCapacity:     800,
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				CurrentDriveSize:    350,
				CurrentDriveType:    "gp3",
				CurrentDriveCount:   2,
				TotalDrivesOnNode:   2,
			},
			response: &cloudops.StoragePoolUpdateResponse{
				ResizeOperationType: api.SdkStoragePool_RESIZE_TYPE_RESIZE_DISK,
				InstanceStorage: []*cloudops.StoragePoolSpec{
					{
						DriveCapacityGiB: 400,
						DriveType:        "gp3",
						DriveCount:       2,
						IOPS:             3000,
					},
				},
			},
			expectedErr: nil,
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
