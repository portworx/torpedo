package azure

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/stretchr/testify/require"
)

func TestRetrieveDataDisks(t *testing.T) {
	var nilDiskSlice []compute.DataDisk
	testDisks := []compute.DataDisk{
		{
			Name: to.StringPtr("disk1"),
		},
		{
			Name: to.StringPtr("disk2"),
		},
	}

	testCases := []struct {
		name        string
		input       compute.VirtualMachineScaleSetVM
		expectedRes []compute.DataDisk
	}{
		{
			name:  "nil vm properties",
			input: compute.VirtualMachineScaleSetVM{},
			expectedRes: []compute.DataDisk{},
		},
		{
			name: "nil storage profile",
			input: compute.VirtualMachineScaleSetVM{
				VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{},
			},
			expectedRes: []compute.DataDisk{},
		},
		{
			name: "nil data disks reference",
			input: compute.VirtualMachineScaleSetVM{
				VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
					StorageProfile: &compute.StorageProfile{},
				},
			},
			expectedRes: []compute.DataDisk{},
		},
		{
			name: "nil data disks slice",
			input: compute.VirtualMachineScaleSetVM{
				VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
					StorageProfile: &compute.StorageProfile{
						DataDisks: &nilDiskSlice,
					},
				},
			},
			expectedRes: []compute.DataDisk{},
		},
		{
			name: "empty data disks slice",
			input: compute.VirtualMachineScaleSetVM{
				VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
					StorageProfile: &compute.StorageProfile{
						DataDisks: &([]compute.DataDisk{}),
					},
				},
			},
			expectedRes: []compute.DataDisk{},
		},
		{
			name: "test data disks",
			input: compute.VirtualMachineScaleSetVM{
				VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
					StorageProfile: &compute.StorageProfile{
						DataDisks: &testDisks,
					},
				},
			},
			expectedRes: testDisks,
		},
	}

	for _, tc := range testCases {
		res := retrieveDataDisks(tc.input)
		require.Equalf(t, tc.expectedRes, res, "TC: %s", tc.name)
	}
}
