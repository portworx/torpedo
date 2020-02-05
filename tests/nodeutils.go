package tests

import (
	"fmt"
	"time"

	"github.com/portworx/torpedo/drivers/node"
	"github.com/sirupsen/logrus"
)

// GetPoolsWithSmallestSize returns storage pools on nodes which have the smallest size
func GetPoolsWithSmallestSize() ([]node.StoragePool, uint64) {
	var (
		smallestPoolSize uint64
		pools            = make([]node.StoragePool, 0)
	)
	for _, workerNode := range node.GetWorkerNodes() {
		for _, p := range workerNode.StoragePools {
			if smallestPoolSize == 0 {
				smallestPoolSize = p.TotalSize
			} else {
				if p.TotalSize < smallestPoolSize {
					smallestPoolSize = p.TotalSize
				}
			}
		}
	}

	return pools, smallestPoolSize
}

// AddLabelsOnNode adds labels on the node
func AddLabelsOnNode(n node.Node, labels map[string]string) error {
	for labelKey, labelValue := range labels {
		if err := Inst().S.AddLabelOnNode(n, labelKey, labelValue); err != nil {
			return err
		}
	}
	return nil
}

// PerformSystemCheck check if core files are present on each node
func PerformSystemCheck() {
	context(fmt.Sprintf("checking for core files..."), func() {
		Step(fmt.Sprintf("verifying if core files are present on each node"), func() {
			nodes := node.GetWorkerNodes()
			expect(nodes).NotTo(beEmpty())
			for _, n := range nodes {
				if !n.IsStorageDriverInstalled {
					continue
				}
				logrus.Infof("looking for core files on node %s", n.Name)
				file, err := Inst().N.SystemCheck(n, node.ConnectionOpts{
					Timeout:         2 * time.Minute,
					TimeBeforeRetry: 10 * time.Second,
				})
				expect(err).NotTo(haveOccurred())
				expect(file).To(beEmpty())
			}
		})
	})
}

func runCmd(cmd string, n node.Node) {
	_, err := Inst().N.RunCommand(n, cmd, node.ConnectionOpts{
		Timeout:         defaultTimeout,
		TimeBeforeRetry: defaultRetryInterval,
		Sudo:            true,
	})
	if err != nil {
		logrus.Warnf("failed to run cmd: %s. err: %v", cmd, err)
	}
}
