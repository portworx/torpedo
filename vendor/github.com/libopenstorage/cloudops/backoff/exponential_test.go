package backoff

import (
	"reflect"
	"testing"
)

func TestVolumeIdsToString(t *testing.T) {
	expectedVolumeIdsStr := []string{"123", "456"}
	var volumeIds []*string
	for i := range expectedVolumeIdsStr {
		volumeIds = append(volumeIds, &expectedVolumeIdsStr[i])
	}
	result := volumeIdsStringDereference(volumeIds)

	if isEqual := reflect.DeepEqual(expectedVolumeIdsStr, result); !isEqual {
		t.Error("volumeIds doesn't match dereferenced value. got: ", result, "; expected: ", expectedVolumeIdsStr)
	}

}
