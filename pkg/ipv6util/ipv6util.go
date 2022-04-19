package ipv6util

import (
	"bufio"
	"strings"

	"github.com/asaskevich/govalidator"
)

// ParseIpv6AddressInPxctlStatus takes output of `pxctl status` and return the list of IPs parsed
func ParseIpv6AddressInPxctlStatus(status string, nodeCount int) []string {
	ips := []string{}
	scanner := bufio.NewScanner(strings.NewReader(status))
	// iterate each line to check for two conditions where IPs are printed:
	// 1. `IP: <addr>`
	//    ex: IP: 0000:111:2222:3333:444:5555:6666:777
	// 2. (number of nodes) lines after `IP \t ID...`
	//    ex: IP					ID					SchedulerNodeName	Auth		StorageNode	Used	Capacity	Status	StorageStatus	Version		Kernel			OS
	//		0000:111:2222:3333:444:5555:6666:777	f703597a-9772-4bdb-b630-6395b3c98658	...	...
	// 		0000:111:2222:3333:444:5555:6666:777	cedc897f-a489-4c28-9c20-12b8b4c3d1d8	...	...
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimPrefix(line, "\t")

		if strings.HasPrefix(line, "IP") {
			// check for the first condition
			if strings.HasPrefix(line, "IP:") {
				ips = append(ips, strings.Split(line, " ")[1])
				continue
			}

			// check for the second condition
			for i := 0; i < nodeCount; i++ {
				if !scanner.Scan() {
					break
				}
				line := scanner.Text()
				line = strings.TrimPrefix(line, "\t")
				ips = append(ips, strings.Split(line, "\t")[0])
			}
		}
	}
	return ips
}

// IsAddressIPv6 checks the given address is a valid Ipv6 address
func IsAddressIPv6(addr string) bool {
	return govalidator.IsIPv6(addr)
}

// AreAddressesIPv6 checks the given addresses are valid Ipv6 addresses
func AreAddressesIPv6(addrs []string) bool {
	isIpv6 := true

	for _, addr := range addrs {
		isIpv6 = isIpv6 && IsAddressIPv6(addr)
	}
	return isIpv6
}
