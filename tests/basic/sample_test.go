package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
)

// Arpit jira https://purestorage.atlassian.net/browse/PTX-24859

var _ = Describe("{VerifyNoNodeRestartUponPxPodRestart}", func() {

	JustBeforeEach(func() {
		StartTorpedoTest("VerifyNoNodeRestartUponPxPodRestart", "Verify that px serivce remain up even if px pod got deleted ", nil, 0)
	})

	err = DeletePXPods("portworx")

	//var contexts []*scheduler.Context
	//Delete perticular pod

	stepLog = "Validate PX on all nodes"
	//storageNodeList :=
	Step(stepLog, func() {
		log.InfoD(stepLog)
		for _, node := range node.GetStorageNodes() {
			status, err := IsPxRunningOnNode(&node)
			log.FailOnError(err, fmt.Sprintf("Failed to check if PX is running on node [%s]", node.Name))
			dash.VerifySafely(status, true, fmt.Sprintf("PX is not running on node [%s]", node.Name))
		}
	})
})
