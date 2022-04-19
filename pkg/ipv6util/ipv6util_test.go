package ipv6util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	sampleIpv6PxctlStatusOutput = `
Status: PX is operational
Telemetry: Disabled or Unhealthy
License: Trial
Node ID: f703597a-9772-4bdb-b630-6395b3c98658
	IP: 0000:111:2222:3333:444:5555:6666:777
 	Local Storage Pool: 1 pool
	POOL	IO_PRIORITY	RAID_LEVEL	USABLE	USED	STATUS	ZONE	REGION
	0	HIGH		raid0		100 GiB	6.2 GiB	Online	default	default
	Local Storage Devices: 1 device
	Device	Path		Media Type		Size		Last-Scan
	0:1	/dev/sdb	STORAGE_MEDIUM_SSD	100 GiB		some time
	* Internal kvdb on this node is sharing this storage device /dev/sdb  to store its data.
	total		-	100 GiB
	Cache Devices:
	 * No cache devices
Cluster Summary
	Cluster ID: px-cluster-2c8df3fc-a2b9-4c31-8b9d-5fddcb4646e1
	Cluster UUID: f2c71ae5-c076-4e33-be1c-001c0d558274
	Scheduler: kubernetes
	Nodes: 6 node(s) with storage (6 online)
	IP					ID					SchedulerNodeName	Auth		StorageNode	Used	Capacity	Status	StorageStatus	Version		Kernel			OS
	0000:111:2222:3333:444:5555:6666:666	f703597a-9772-4bdb-b630-6395b3c98658	node05			Disabled	Yes		6.2 GiB	100 GiB		Online	Up (This node)	2.11.0-5fbb8c2	3.10.0-1160.53.1.el7.x86_64	CentOS Linux 7 (Core)
	0000:111:2222:3333:444:5555:6666:555	cedc897f-a489-4c28-9c20-12b8b4c3d1d8	node01			Disabled	Yes		6.7 GiB	100 GiB		Online	Up		2.11.0-5fbb8c2	3.10.0-1160.53.1.el7.x86_64	CentOS Linux 7 (Core)
	0000:111:2222:3333:444:5555:6666:444	956aafc1-a52d-41f3-afb1-6427e2a3b0ef	node04			Disabled	Yes		6.3 GiB	100 GiB		Online	Up		2.11.0-5fbb8c2	3.10.0-1160.53.1.el7.x86_64	CentOS Linux 7 (Core)
	0000:111:2222:3333:444:5555:6666:333	6d801e0f-a7e7-4063-8f2f-50b43c1d9608	node03			Disabled	Yes		6.6 GiB	100 GiB		Online	Up		2.11.0-5fbb8c2	3.10.0-1160.53.1.el7.x86_64	CentOS Linux 7 (Core)
	0000:111:2222:3333:444:5555:6666:222	28dee5d4-7724-41eb-a86d-929a3f88456e	node06			Disabled	Yes		6.3 GiB	100 GiB		Online	Up		2.11.0-5fbb8c2	3.10.0-1160.53.1.el7.x86_64	CentOS Linux 7 (Core)
	0000:111:2222:3333:444:5555:6666:111	0e88d11f-6fb1-4898-b76a-e38c200fa7ae	node02			Disabled	Yes		6.1 GiB	100 GiB		Online	Up		2.11.0-5fbb8c2	3.10.0-1160.53.1.el7.x86_64	CentOS Linux 7 (Core)
	Warnings:
		 WARNING: Internal Kvdb is not using dedicated drive on nodes [0000:111:2222:3333:444:5555:6666:666 0000:111:2222:3333:444:5555:6666:444 0000:111:2222:3333:444:5555:6666:333]. This configuration is not recommended for production clusters.
Global Storage Pool
	Total Used    	:  38 GiB
	Total Capacity	:  600 GiB
`
	sampleIpv4PxctlStatusOutput = `
Status: PX is operational
Telemetry: Disabled or Unhealthy
License: Trial
Node ID: f703597a-9772-4bdb-b630-6395b3c98658
	IP: 192.168.121.111
 	Local Storage Pool: 1 pool
	POOL	IO_PRIORITY	RAID_LEVEL	USABLE	USED	STATUS	ZONE	REGION
	0	HIGH		raid0		100 GiB	6.2 GiB	Online	default	default
	Local Storage Devices: 1 device
	Device	Path		Media Type		Size		Last-Scan
	0:1	/dev/sdb	STORAGE_MEDIUM_SSD	100 GiB		some time
	* Internal kvdb on this node is sharing this storage device /dev/sdb  to store its data.
	total		-	100 GiB
	Cache Devices:
	 * No cache devices
Cluster Summary
	Cluster ID: px-cluster-2c8df3fc-a2b9-4c31-8b9d-5fddcb4646e1
	Cluster UUID: f2c71ae5-c076-4e33-be1c-001c0d558274
	Scheduler: kubernetes
	Nodes: 6 node(s) with storage (6 online)
	IP					ID					SchedulerNodeName	Auth		StorageNode	Used	Capacity	Status	StorageStatus	Version		Kernel			OS
	192.168.121.111	f703597a-9772-4bdb-b630-6395b3c98658	node05			Disabled	Yes		6.2 GiB	100 GiB		Online	Up (This node)	2.11.0-5fbb8c2	3.10.0-1160.53.1.el7.x86_64	CentOS Linux 7 (Core)
	192.168.121.222	cedc897f-a489-4c28-9c20-12b8b4c3d1d8	node01			Disabled	Yes		6.7 GiB	100 GiB		Online	Up		2.11.0-5fbb8c2	3.10.0-1160.53.1.el7.x86_64	CentOS Linux 7 (Core)
	192.168.121.333	956aafc1-a52d-41f3-afb1-6427e2a3b0ef	node04			Disabled	Yes		6.3 GiB	100 GiB		Online	Up		2.11.0-5fbb8c2	3.10.0-1160.53.1.el7.x86_64	CentOS Linux 7 (Core)
	192.168.121.444	6d801e0f-a7e7-4063-8f2f-50b43c1d9608	node03			Disabled	Yes		6.6 GiB	100 GiB		Online	Up		2.11.0-5fbb8c2	3.10.0-1160.53.1.el7.x86_64	CentOS Linux 7 (Core)
	192.168.121.555	28dee5d4-7724-41eb-a86d-929a3f88456e	node06			Disabled	Yes		6.3 GiB	100 GiB		Online	Up		2.11.0-5fbb8c2	3.10.0-1160.53.1.el7.x86_64	CentOS Linux 7 (Core)
	192.168.121.666	0e88d11f-6fb1-4898-b76a-e38c200fa7ae	node02			Disabled	Yes		6.1 GiB	100 GiB		Online	Up		2.11.0-5fbb8c2	3.10.0-1160.53.1.el7.x86_64	CentOS Linux 7 (Core)
Global Storage Pool
	Total Used    	:  38 GiB
	Total Capacity	:  600 GiB
`

	nodeCount = 6
)

func TestValidIpv6Address(t *testing.T) {
	addrs := ParseIpv6AddressInPxctlStatus(sampleIpv6PxctlStatusOutput, nodeCount)
	isIpv6 := AreAddressesIPv6(addrs)
	assert.True(t, isIpv6, "addresses are expected to be ipv6")

	addrs = ParseIpv6AddressInPxctlStatus(sampleIpv4PxctlStatusOutput, nodeCount)
	isIpv6 = AreAddressesIPv6(addrs)
	assert.False(t, isIpv6, "addresses are expected to be ipv4")

}
