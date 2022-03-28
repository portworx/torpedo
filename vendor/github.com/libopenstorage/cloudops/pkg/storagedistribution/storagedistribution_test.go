package storagedistribution

import (
	"github.com/libopenstorage/cloudops"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCalculateDriveCapacity(t *testing.T) {
	testCases := []*cloudops.StoragePoolUpdateRequest{
		&cloudops.StoragePoolUpdateRequest{
			// 53.6 per disk (should be rounded up to 54)
			DesiredCapacity:   uint64(953),
			CurrentDriveSize:  uint64(137),
			CurrentDriveCount: uint64(5),
		},
		&cloudops.StoragePoolUpdateRequest{
			// 53.4 per disk (should be rounded up to 54)
			DesiredCapacity:   uint64(952),
			CurrentDriveSize:  uint64(137),
			CurrentDriveCount: uint64(5),
		},
		&cloudops.StoragePoolUpdateRequest{
			// 53.2 per disk (should be rounded up to 54)
			DesiredCapacity:   uint64(951),
			CurrentDriveSize:  uint64(137),
			CurrentDriveCount: uint64(5),
		},
		&cloudops.StoragePoolUpdateRequest{
			// 2.1 per disk (should be rounded up to 3)
			DesiredCapacity:   uint64(271),
			CurrentDriveSize:  uint64(28),
			CurrentDriveCount: uint64(9),
		},
	}

	for _, request := range testCases {
		result := calculateDriveCapacity(request)
		if request.DesiredCapacity > request.CurrentDriveCount*request.CurrentDriveSize {
			require.True(t, (result+request.CurrentDriveSize)*request.CurrentDriveCount >= request.DesiredCapacity)
		}
	}
}
