package ipv6util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidIpv6Address(t *testing.T) {
	// pxctl status tests
	addrs := ParseIPv6AddressInPxctlCommand(PXCTL_STATUS, sampleIpv6PxctlStatusOutput, nodeCount)
	assert.NotEmpty(t, addrs, "addresses are not expected to be empty. running command: %v", PXCTL_STATUS)
	isIpv6 := AreAddressesIPv6(addrs)
	assert.True(t, isIpv6, "running command %v. addresses are expected to be ipv6, got: %v", PXCTL_STATUS, addrs)

	addrs = ParseIPv6AddressInPxctlCommand(PXCTL_STATUS, sampleIpv4PxctlStatusOutput, nodeCount)
	assert.NotEmpty(t, addrs, "addresses are not expected to be empty. running command: %v", PXCTL_STATUS)
	isIpv6 = AreAddressesIPv6(addrs)
	assert.False(t, isIpv6, "running command %v. addresses are expected to be ipv4, got: %v", PXCTL_STATUS, addrs)

	// pxctl cluster list tests
	addrs = ParseIPv6AddressInPxctlCommand(PXCTL_CLUSTER_LIST, sampleIpv6PxctlClusterListOutput, nodeCount)
	assert.NotEmpty(t, addrs, "addresses are not expected to be empty. running command: %v", PXCTL_CLUSTER_LIST)
	isIpv6 = AreAddressesIPv6(addrs)
	assert.True(t, isIpv6, "running command %v. addresses are expected to be ipv6, got: %v", PXCTL_CLUSTER_LIST, addrs)

	// pxctl cluster inspect
	addrs = ParseIPv6AddressInPxctlCommand(PXCTL_CLUSTER_INSPECT, sampleIpv6PxctlClusterInspectOutput, nodeCount)
	assert.NotEmpty(t, addrs, "addresses are not expected to be empty. running command: %v", PXCTL_CLUSTER_INSPECT)
	isIpv6 = AreAddressesIPv6(addrs)
	assert.True(t, isIpv6, "running command %v. addresses are expected to be ipv6, got: %v", PXCTL_CLUSTER_INSPECT, addrs)
}
