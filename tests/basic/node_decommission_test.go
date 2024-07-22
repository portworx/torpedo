package tests

import (
	"fmt"
	"github.com/libopenstorage/stork/pkg/apis/stork/v1alpha1"
	storkops "github.com/portworx/sched-ops/k8s/stork"
	"github.com/portworx/torpedo/pkg/log"
	"math/rand"
	"time"

	"github.com/libopenstorage/openstorage/api"
	. "github.com/onsi/ginkgo/v2"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/tests"
)

var _ = Describe("{DecommissionNode}", func() {

	JustBeforeEach(func() {
		StartTorpedoTest("DecommissionNode", "Validate node decommission", nil, 0)
	})
	var contexts []*scheduler.Context

	testName := "decommissionnode"
	stepLog := "has to decommission a node and check if node was decommissioned successfully"
	It(stepLog, func() {

		currNode := node.GetStorageDriverNodes()[0]
		err := Inst().V.SetClusterOpts(currNode, map[string]string{
			"--auto-fstrim": "on",
		})
		log.FailOnError(err, "error enabling auto fstrim on cluster")

		if Contains(Inst().AppList, "nginx-proxy-deployment") {
			var masterNode node.Node
			stepLog = "setup proxy server necessary for proxy volume"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				masterNodes := node.GetMasterNodes()
				if len(masterNodes) == 0 {
					log.FailOnError(fmt.Errorf("no master nodes found"), "Identifying master node of proxy server failed")
				}

				masterNode = masterNodes[0]
				err = SetupProxyServer(masterNode)
				log.FailOnError(err, fmt.Sprintf("error setting up proxy server on master node %s", masterNode.Name))

			})
			stepLog = "create storage class for proxy volumes"
			Step(stepLog, func() {
				log.InfoD(stepLog)
				addresses := masterNode.Addresses
				if len(addresses) == 0 {
					log.FailOnError(fmt.Errorf("no addresses found for node [%s]", masterNode.Name), "error getting ip addresses ")
				}
				err = CreateNFSProxyStorageClass("portworx-proxy-volume-volume", addresses[0], "/exports/testnfsexportdir")
				log.FailOnError(err, "error creating storage class for proxy volume")
			})
		}
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("%s-%d", testName, i))...)
		}
		ValidateApplications(contexts)

		var storageDriverNodes []node.Node
		Step(fmt.Sprintf("get storage driver nodes"), func() {
			storageDriverNodes = node.GetStorageDriverNodes()
			dash.VerifyFatal(len(storageDriverNodes) > 0, true, "Verify worker nodes")
		})

		nodeIndexMap := make(map[int]int)
		lenWorkerNodes := len(storageDriverNodes)
		chaosLevel := Inst().ChaosLevel
		// chaosLevel in this case is the number of worker nodes to be decommissioned
		// in case of being greater than that, it will assume the total no of worker nodes
		if chaosLevel > lenWorkerNodes {
			chaosLevel = lenWorkerNodes
		}

		Step(fmt.Sprintf("sort nodes randomly according to chaos level %d", chaosLevel), func() {
			for len(nodeIndexMap) != chaosLevel {
				index := rand.Intn(lenWorkerNodes)
				nodeIndexMap[index] = index
			}
		})

		// decommission nodes one at a time according to chaosLevel
		for nodeIndex := range nodeIndexMap {
			nodeToDecommission := storageDriverNodes[nodeIndex]

			fsTrimStatuses, err := Inst().V.GetAutoFsTrimStatus(nodeToDecommission.MgmtIp)
			log.FailOnError(err, fmt.Sprintf("error autofstrim status node %v status", nodeToDecommission.Name))
			for volId, fstrimStatus := range fsTrimStatuses {
				log.Infof("Volume %s fstrim status %v on node [%s]", volId, fstrimStatus, nodeToDecommission.Name)
			}

			//checking node status before decommission
			status, err := Inst().V.GetNodeStatus(nodeToDecommission)
			log.FailOnError(err, "error checking node [%s] status", nodeToDecommission.Name)
			if *status != api.Status_STATUS_OK {
				continue
			}

			nodeToDecommission, err = node.GetNodeByName(nodeToDecommission.Name) //This is required when multiple nodes are decommissioned sequentially
			log.FailOnError(err, fmt.Sprintf("node [%s] not found with name", nodeToDecommission.Name))
			stepLog = fmt.Sprintf("decommission node %s", nodeToDecommission.Name)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				var suspendedScheds []*v1alpha1.VolumeSnapshotSchedule
				defer func() {
					if len(suspendedScheds) > 0 {
						for _, sched := range suspendedScheds {
							makeSuspend := false
							sched.Spec.Suspend = &makeSuspend
							_, err := storkops.Instance().UpdateSnapshotSchedule(sched)
							log.FailOnError(err, "error resuming volumes snapshot schedule for volume [%s] ", sched.Name)
						}
					}
				}()
				err = PrereqForNodeDecomm(nodeToDecommission, suspendedScheds)
				log.FailOnError(err, "error performing prerequisites for node decommission")

				err := Inst().S.PrepareNodeToDecommission(nodeToDecommission, Inst().Provisioner)
				dash.VerifyFatal(err, nil, "Validate node decommission preparation")
				err = Inst().V.DecommissionNode(&nodeToDecommission)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validate node [%s] decommission init", nodeToDecommission.Name))
				stepLog = fmt.Sprintf("check if node %s was decommissioned", nodeToDecommission.Name)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					result := false
					t := func() (interface{}, bool, error) {
						status, err := Inst().V.GetNodeStatus(nodeToDecommission)
						if err != nil {
							return false, true, err
						}
						if *status == api.Status_STATUS_NONE {
							return true, false, nil
						}
						return false, true, fmt.Errorf("node %s not decomissioned yet", nodeToDecommission.Name)
					}
					decommissioned, err := task.DoRetryWithTimeout(t, defaultTimeout, defaultRetryInterval)
					log.FailOnError(err, "Failed to get decommissioned node status")
					result = decommissioned.(bool)

					dash.VerifyFatal(result, true, fmt.Sprintf("Validate node [%s] is decommissioned", nodeToDecommission.Name))

				})
			})
			stepLog = fmt.Sprintf("Rejoin node %s", nodeToDecommission.Name)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				//reboot required to remove encrypted dm devices if any
				err := Inst().N.RebootNode(nodeToDecommission, node.RebootNodeOpts{
					Force: true,
					ConnectionOpts: node.ConnectionOpts{
						Timeout:         defaultCommandTimeout,
						TimeBeforeRetry: defaultRetryInterval,
					},
				})
				log.FailOnError(err, fmt.Sprintf("error rebooting node %s", nodeToDecommission.Name))
				err = Inst().V.RejoinNode(&nodeToDecommission)
				dash.VerifyFatal(err, nil, "Validate node rejoin init")
				var rejoinedNode *api.StorageNode
				t := func() (interface{}, bool, error) {
					drvNodes, err := Inst().V.GetDriverNodes()
					if err != nil {
						return false, true, err
					}

					for _, n := range drvNodes {
						if n.Hostname == nodeToDecommission.Hostname {
							rejoinedNode = n
							return true, false, nil
						}
					}

					return false, true, fmt.Errorf("node %s not joined yet", nodeToDecommission.Name)
				}
				_, err = task.DoRetryWithTimeout(t, 20*time.Minute, defaultRetryInterval)
				log.FailOnError(err, fmt.Sprintf("error joining the node [%s]", nodeToDecommission.Name))
				dash.VerifyFatal(rejoinedNode != nil, true, fmt.Sprintf("verify node [%s] rejoined PX cluster", nodeToDecommission.Name))
				err = Inst().S.RefreshNodeRegistry()
				log.FailOnError(err, "error refreshing node registry")
				err = Inst().V.RefreshDriverEndpoints()
				log.FailOnError(err, "error refreshing storage drive endpoints")
				nodeToDecommission = node.Node{}
				for _, n := range node.GetStorageDriverNodes() {
					if n.Name == rejoinedNode.Hostname {
						nodeToDecommission = n
						break
					}
				}
				if nodeToDecommission.Name == "" {
					log.FailOnError(fmt.Errorf("rejoined node not found"), fmt.Sprintf("node [%s] not found in the node registry", rejoinedNode.Hostname))
				}
				err = Inst().V.WaitDriverUpOnNode(nodeToDecommission, Inst().DriverStartTimeout)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validate driver up on rejoined node [%s] after rejoining", nodeToDecommission.Name))
			})

		}

		Step("destroy apps", func() {
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
		PerformSystemCheck()

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{KvdbDecommissionNode}", func() {

	JustBeforeEach(func() {
		StartTorpedoTest("KvdbDecommissionNode", "Validate decommission of kvdb nodes", nil, 0)
	})
	var contexts []*scheduler.Context

	testName := "kvdbdecommissionnode"
	stepLog := "has to decommission a kvdb node and check if node was decommissioned successfully"
	It(stepLog, func() {
		log.InfoD(stepLog)
		contexts = make([]*scheduler.Context, 0)

		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			contexts = append(contexts, ScheduleApplications(fmt.Sprintf("%s-%d", testName, i))...)
		}

		ValidateApplications(contexts)

		var kvdbNodes []KvdbNode
		var newKVDBNodes []KvdbNode
		var err error
		Step(fmt.Sprintf("get kvdb nodes"), func() {
			kvdbNodes, err = GetAllKvdbNodes()
			log.FailOnError(err, "Failed to get list of KVDB nodes from the cluster")
			dash.VerifyFatal(len(kvdbNodes) == 3, true, "Verify kvdb nodes")
		})

		// decommission nodes one at a time according to chaosLevel
		for _, kvdbNode := range kvdbNodes {
			nodeToDecommission, err := node.GetNodeDetailsByNodeID(kvdbNode.ID)
			log.FailOnError(err, fmt.Sprintf("error getting node with id: %s", kvdbNode.ID))
			stepLog = fmt.Sprintf("decommission node %s", nodeToDecommission.Name)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				var suspendedScheds []*v1alpha1.VolumeSnapshotSchedule
				defer func() {
					if len(suspendedScheds) > 0 {
						for _, sched := range suspendedScheds {
							makeSuspend := false
							sched.Spec.Suspend = &makeSuspend
							_, err := storkops.Instance().UpdateSnapshotSchedule(sched)
							log.FailOnError(err, "error resuming volumes snapshot schedule for volume [%s]", sched.Name)
						}
					}
				}()
				err = PrereqForNodeDecomm(nodeToDecommission, suspendedScheds)
				log.FailOnError(err, "error performing prerequisites for node decommission")
				err := Inst().S.PrepareNodeToDecommission(nodeToDecommission, Inst().Provisioner)
				dash.VerifyFatal(err, nil, "Validate node decommission preparation")
				err = Inst().V.DecommissionNode(&nodeToDecommission)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validate node [%s] decommission init", nodeToDecommission.Name))
				stepLog = fmt.Sprintf("check if node %s was decommissioned", nodeToDecommission.Name)
				Step(stepLog, func() {
					log.InfoD(stepLog)
					t := func() (interface{}, bool, error) {
						status, err := Inst().V.GetNodeStatus(nodeToDecommission)
						if err != nil {
							return false, true, err
						}
						if *status == api.Status_STATUS_NONE {
							return true, false, nil
						}
						return false, true, fmt.Errorf("node %s not decomissioned yet", nodeToDecommission.Name)
					}
					decommissioned, err := task.DoRetryWithTimeout(t, defaultTimeout, defaultRetryInterval)
					log.FailOnError(err, "Failed to get decommissioned node status")
					dash.VerifyFatal(decommissioned.(bool), true, fmt.Sprintf("Validate node [%s] is decommissioned", nodeToDecommission.Name))
				})
			})
			err = Inst().S.RefreshNodeRegistry()
			log.FailOnError(err, "error refreshing node registry")
			err = Inst().V.RefreshDriverEndpoints()
			log.FailOnError(err, "error refreshing storage drive endpoints")

			t := func() (interface{}, bool, error) {

				newKVDBNodes, err = GetAllKvdbNodes()
				log.FailOnError(err, "Failed to get list of KVDB nodes from the cluster")
				if err != nil {
					return false, true, err
				}

				if len(newKVDBNodes) == 3 {
					return true, false, nil
				}

				return false, true, fmt.Errorf("current  number of KVDB nodes : %d", len(newKVDBNodes))
			}
			_, err = task.DoRetryWithTimeout(t, 4*time.Minute, defaultRetryInterval)
			dash.VerifyFatal(len(newKVDBNodes) == 3, true, "Verify kvdb nodes are updated")

			isLeaderHealthy := false
			for _, nKVDBNode := range newKVDBNodes {
				dash.VerifyFatal(nKVDBNode.IsHealthy, true, fmt.Sprintf("verify kvdb node %s is healthy", nKVDBNode.ID))
				if nKVDBNode.Leader && nKVDBNode.IsHealthy {
					isLeaderHealthy = true
				}
			}
			dash.VerifyFatal(isLeaderHealthy, true, "verify kvdb leader node exists and healthy")

			stepLog = fmt.Sprintf("Rejoin node %s", nodeToDecommission.Name)
			Step(stepLog, func() {
				log.InfoD(stepLog)
				err := Inst().V.RejoinNode(&nodeToDecommission)
				dash.VerifyFatal(err, nil, "Validate node rejoin init")
				var rejoinedNode *api.StorageNode
				t := func() (interface{}, bool, error) {
					drvNodes, err := Inst().V.GetDriverNodes()
					if err != nil {
						return false, true, err
					}

					for _, n := range drvNodes {
						log.Infof("checking for node %s", n.Hostname)
						if n.Hostname == nodeToDecommission.Hostname {
							rejoinedNode = n
							return true, false, nil
						}
					}

					return false, true, fmt.Errorf("node %s not joined yet", nodeToDecommission.Name)
				}
				_, err = task.DoRetryWithTimeout(t, 20*time.Minute, defaultRetryInterval)
				log.FailOnError(err, fmt.Sprintf("error joining the node [%s]", nodeToDecommission.Name))
				dash.VerifyFatal(rejoinedNode != nil, true, fmt.Sprintf("verify node [%s] rejoined PX cluster", nodeToDecommission.Name))
				err = Inst().S.RefreshNodeRegistry()
				log.FailOnError(err, "error refreshing node registry")
				err = Inst().V.RefreshDriverEndpoints()
				log.FailOnError(err, "error refreshing storage drive endpoints")
				nodeToDecommission = node.Node{}
				for _, n := range node.GetStorageDriverNodes() {
					if n.Name == rejoinedNode.Hostname {
						nodeToDecommission = n
						break
					}
				}
				if nodeToDecommission.Name == "" {
					log.FailOnError(fmt.Errorf("rejoined node not found"), fmt.Sprintf("node [%s] not found in the node registry", rejoinedNode.Hostname))
				}
				err = Inst().V.WaitDriverUpOnNode(nodeToDecommission, Inst().DriverStartTimeout)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Validate driver up on rejoined node [%s] after rejoining", nodeToDecommission.Name))
			})

		}

		Step("destroy apps", func() {
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})

	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})

var _ = Describe("{Decommission2Storage1KVDBNodeAtOnce}", func() {
	var (
		contexts            = make([]*scheduler.Context, 0)
		nodesToDecommission []node.Node
	)

	JustBeforeEach(func() {
		StartTorpedoTest("Decommission2Storage1KVDBNodeAtOnce", "Validate decommissioning of 2 Storage nodes and 1 KVDB node at once", nil, 86013)
	})

	// decommissionAndRejoinNode decommissions a node and rejoin it back to the cluster
	decommissionAndRejoinNode := func(n node.Node) (err error) {
		log.Infof("Decommissioning node [%s]", n.Name)

		var suspendedSchedules []*v1alpha1.VolumeSnapshotSchedule
		defer func() {
			if len(suspendedSchedules) > 0 {
				for _, schedule := range suspendedSchedules {
					makeSuspend := false
					schedule.Spec.Suspend = &makeSuspend
					_, err := storkops.Instance().UpdateSnapshotSchedule(schedule)
					if err != nil {
						err = fmt.Errorf("error resuming volumes snapshot schedule for volume [%s]. Err: [%v]", schedule.Name, err)
						return
					}
				}
			}
		}()
		err = PrereqForNodeDecomm(n, suspendedSchedules)
		if err != nil {
			err = fmt.Errorf("failed while performing prerequisites for node decommission. Err: [%v]", err)
			return
		}
		err = Inst().S.PrepareNodeToDecommission(n, Inst().Provisioner)
		if err != nil {
			err = fmt.Errorf("failed while preparing node [%s] for decommission. Err: [%v]", n.Name, err)
			return
		}
		err = Inst().V.DecommissionNode(&n)
		if err != nil {
			err = fmt.Errorf("failed while decommissioning node [%s]. Err: [%v]", n.Name, err)
			return
		}

		log.Infof("Checking if node [%s] was decommissioned", n.Name)
		t := func() (interface{}, bool, error) {
			status, err := Inst().V.GetNodeStatus(n)
			if err != nil {
				return false, true, err
			}
			if *status == api.Status_STATUS_NONE {
				return true, false, nil
			}
			return false, true, fmt.Errorf("node %s not decomissioned yet", n.Name)
		}
		decommissioned, err := task.DoRetryWithTimeout(t, defaultTimeout, defaultRetryInterval)
		if err != nil {
			err = fmt.Errorf("failed while getting decommissioned node status. Err: [%v]", err)
			return
		}
		if !decommissioned.(bool) {
			err = fmt.Errorf("failed to decommission node [%s]", n.Name)
			return
		}
		err = Inst().S.RefreshNodeRegistry()
		if err != nil {
			err = fmt.Errorf("failed while refreshing node registry. Err: [%v]", err)
			return
		}
		err = Inst().V.RefreshDriverEndpoints()
		if err != nil {
			err = fmt.Errorf("failed while refreshing storage drive endpoints. Err: [%v]", err)
			return
		}
		t = func() (interface{}, bool, error) {
			newKVDBNodes, err := GetAllKvdbNodes()
			if err != nil {
				return false, true, fmt.Errorf("failed to get all KVDB nodes. Err: [%v]", err)
			}
			if len(newKVDBNodes) == 3 {
				log.Infof("The new KVDB nodes are: [%v]", newKVDBNodes)
				for _, nKVDBNode := range newKVDBNodes {
					if nKVDBNode.Leader && !nKVDBNode.IsHealthy {
						return false, true, fmt.Errorf("leader kvdb node [%s] is not healthy", nKVDBNode.ID)
					}
				}
				for _, nKVDBNode := range newKVDBNodes {
					if !nKVDBNode.IsHealthy {
						return false, true, fmt.Errorf("kvdb node [%s] is not healthy", nKVDBNode.ID)
					}
				}
				return true, false, nil
			}
			return false, true, fmt.Errorf("actual number of KVDB nodes [%d], expected [%d]", len(newKVDBNodes), 3)
		}
		_, err = task.DoRetryWithTimeout(t, 8*time.Minute, defaultRetryInterval)
		if err != nil {
			err = fmt.Errorf("failed while verifying KVDB nodes are updated. Err: [%v]", err)
			return
		}

		log.Infof("Rejoining node [%s]", n.Name)
		err = Inst().V.RejoinNode(&n)
		if err != nil {
			err = fmt.Errorf("failed while rejoining node [%s]. Err: [%v]", n.Name, err)
			return
		}
		var rejoinedNode *api.StorageNode
		t = func() (interface{}, bool, error) {
			drvNodes, err := Inst().V.GetDriverNodes()
			if err != nil {
				return false, true, fmt.Errorf("failed to get driver nodes. Err: [%v]", err)
			}
			for _, n := range drvNodes {
				log.Infof("current node [%s]", n.Hostname)
				if n.Hostname == n.Hostname {
					rejoinedNode = n
					return true, false, nil
				}
			}
			return false, true, fmt.Errorf("node %s not rejoined yet", n.Name)
		}
		_, err = task.DoRetryWithTimeout(t, 20*time.Minute, defaultRetryInterval)
		if err != nil {
			err = fmt.Errorf("failed while rejoining the node [%s/%s]. Err: [%v]", n.Name, n.Hostname, err)
			return
		}
		if rejoinedNode == nil {
			err = fmt.Errorf("rejoined node not found")
			return
		}
		if rejoinedNode.Hostname == "" {
			err = fmt.Errorf("rejoined node [%v] hostname not found", rejoinedNode)
			return
		}
		err = Inst().S.RefreshNodeRegistry()
		if err != nil {
			err = fmt.Errorf("failed while refreshing node registry. Err: [%v]", err)
			return
		}
		err = Inst().V.RefreshDriverEndpoints()
		if err != nil {
			err = fmt.Errorf("failed while refreshing storage drive endpoints. Err: [%v]", err)
			return
		}
		for _, pxNode := range node.GetStorageDriverNodes() {
			if pxNode.Name == rejoinedNode.Hostname {
				err = Inst().V.WaitDriverUpOnNode(pxNode, Inst().DriverStartTimeout)
				if err != nil {
					err = fmt.Errorf("failed while waiting for driver up on rejoined node [%s]. Err: [%v]", pxNode.Name, err)
					return
				}
			}
		}
		return nil
	}

	stepLog := "validate decommissioning of 2 storage nodes and 1 kvdb node"
	It(stepLog, func() {
		log.InfoD(stepLog)

		stepLog := "Schedule applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for i := 0; i < Inst().GlobalScaleFactor; i++ {
				testName := "decommission2st1kvdb-nodes"
				contexts = append(contexts, ScheduleApplications(fmt.Sprintf("%s-%d", testName, i))...)
			}
			ValidateApplications(contexts)
		})

		stepLog = "Pick 2 storage nodes and 1 kvdb node to decommission"
		Step(stepLog, func() {
			log.InfoD(stepLog)

			kvdbNodes, err := GetAllKvdbNodes()
			log.FailOnError(err, "failed to get list of KVDB nodes from the cluster")
			dash.VerifyFatal(len(kvdbNodes), 3, "Verify if there are 3 kvdb nodes")

			storageNodes, err := GetStorageNodes()
			log.FailOnError(err, "failed to get list of storage nodes from the cluster")

			rand.Seed(time.Now().UnixNano())
			rand.Shuffle(len(storageNodes), func(i, j int) { storageNodes[i], storageNodes[j] = storageNodes[j], storageNodes[i] })
			selectedStorageNodes := storageNodes[:2]

			rand.Shuffle(len(kvdbNodes), func(i, j int) { kvdbNodes[i], kvdbNodes[j] = kvdbNodes[j], kvdbNodes[i] })
			selectedKvdbNode := kvdbNodes[0]

			kvdbNodeDetails, err := node.GetNodeDetailsByNodeID(selectedKvdbNode.ID)
			log.FailOnError(err, fmt.Sprintf("error getting node with id: %s", selectedKvdbNode.ID))

			nodesToDecommission = append(selectedStorageNodes, kvdbNodeDetails)
			log.InfoD("Nodes to decommission: [%+v]", nodesToDecommission)
		})

		Step("Decommission and rejoin nodes", func() {
			for _, n := range nodesToDecommission {
				err := decommissionAndRejoinNode(n)
				log.FailOnError(err, fmt.Sprintf("failed while decommissioning and rejoining node [%s]", n.Name))
			}
		})

		stepLog = "Destroy applications"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})
	})

	JustAfterEach(func() {
		defer EndTorpedoTest()
		AfterEachTest(contexts)
	})
})
