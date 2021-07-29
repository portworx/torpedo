package tests

import (
	"bytes"
	"fmt"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/backup"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math"
	"math/rand"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"text/template"
	"time"

	"container/ring"

	"github.com/onsi/ginkgo"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"

	"github.com/portworx/torpedo/pkg/email"
	"github.com/sirupsen/logrus"
)

const (
	subject = "Torpedo Longevity Report"
	from    = "wilkins@portworx.com"

	// EmailRecipientsConfigMapField is field in config map whose value is comma
	// seperated list of email IDs which will receive email notifications about longevity
	EmailRecipientsConfigMapField = "emailRecipients"
	// DefaultEmailRecipient is list of email IDs that will receive email
	// notifications when no EmailRecipientsConfigMapField field present in configMap
	DefaultEmailRecipient = "test@portworx.com"
	// SendGridEmailAPIKeyField is field in config map which stores the SendGrid Email API key
	SendGridEmailAPIKeyField = "sendGridAPIKey"
)
const (
	validateReplicationUpdateTimeout = 2 * time.Hour
	errorChannelSize                 = 10
)

// EmailRecipients list of email IDs to send email to
var EmailRecipients []string

// SendGridEmailAPIKey holds API key used to interact
// with SendGrid Email APIs
var SendGridEmailAPIKey string

//BackupCounter holds the iteration of TriggerBackup
var BackupCounter = 0

// RestoreCounter holds the iteration of TriggerRestore
var RestoreCounter = 0

var newNamespaceCounter = 0

// Event describes type of test trigger
type Event struct {
	ID   string
	Type string
}

// EventRecord recodes which event took
// place at what time with what outcome
type EventRecord struct {
	Event   Event
	Start   string
	End     string
	Outcome []error
}

// eventRing is circular buffer to store
// events for sending email notifications
var eventRing *ring.Ring

// emailRecords stores events for rendering
// email template
type emailRecords struct {
	Records []EventRecord
}

// GenerateUUID generates unique ID
func GenerateUUID() string {
	uuidbyte, _ := exec.Command("uuidgen").Output()
	return strings.TrimSpace(string(uuidbyte))
}

// UpdateOutcome updates outcome based on error
func UpdateOutcome(event *EventRecord, err error) {
	if err != nil {
		event.Outcome = append(event.Outcome, err)
	}
}

// ProcessErrorWithMessage updates outcome and expects no error
func ProcessErrorWithMessage(event *EventRecord, err error, desc string) {
	UpdateOutcome(event, err)
	expect(err).NotTo(haveOccurred(), desc)
}

const (
	// HAIncrease performs repl-add
	HAIncrease = "haIncrease"
	// HADecrease performs repl-reduce
	HADecrease = "haDecrease"
	// AppTaskDown deletes application task for all contexts
	AppTaskDown = "appTaskDown"
	// RestartVolDriver restart volume driver
	RestartVolDriver = "restartVolDriver"
	// CrashVolDriver crashes volume driver
	CrashVolDriver = "crashVolDriver"
	// RebootNode reboots all nodes one by one
	RebootNode = "rebootNode"
	// EmailReporter notifies via email outcome of past events
	EmailReporter = "emailReporter"
	// BackupAllApps Perform backups of all deployed apps
	BackupAllApps = "backupAllApps"
	// InspectScheduledBackups Create one namespace, delete one namespace, inspect next scheduled backup
	InspectScheduledBackups = "inspectScheduledBackups"
	// TestInspectBackup Inspect a backup
	TestInspectBackup = "inspectBackup"
	// TestInspectRestore Inspect a restore
	TestInspectRestore = "inspectRestore"
	// TestDeleteBackup Delete a backup
	TestDeleteBackup = "deleteBackup"
	//BackupSpecificResource backs up a specified resource
	BackupSpecificResource = "backupSpecificResource"
	// RestoreNamespace restores a single namespace from a backup
	RestoreNamespace = "restoreNamespace"
	//BackupSpecificResourceOnCluster backs up all of a resource type on the cluster
	BackupSpecificResourceOnCluster = "backupSpecificResourceOnCluster"
	//BackupUsingLabelOnCluster backs up resources on a cluster using a specific label
	BackupUsingLabelOnCluster = "backupUsingLabelOnCluster"
)

// TriggerHAIncrease performs repl-add on all volumes of given contexts
func TriggerHAIncrease(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	defer ginkgo.GinkgoRecover()
	event := &EventRecord{
		Event: Event{
			ID:   GenerateUUID(),
			Type: HAIncrease,
		},
		Start:   time.Now().Format(time.RFC1123),
		Outcome: []error{},
	}

	defer func() {
		event.End = time.Now().Format(time.RFC1123)
		*recordChan <- event
	}()

	expReplMap := make(map[*volume.Volume]int64)
	Step("get volumes for all apps in test and increase replication factor", func() {
		time.Sleep(10 * time.Minute)
		for _, ctx := range contexts {
			var appVolumes []*volume.Volume
			var err error
			Step(fmt.Sprintf("get volumes for %s app", ctx.App.Key), func() {
				appVolumes, err = Inst().S.GetVolumes(ctx)
				UpdateOutcome(event, err)
				expect(appVolumes).NotTo(beEmpty())
			})
			opts := volume.Options{
				ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout,
			}
			for _, v := range appVolumes {
				MaxRF := Inst().V.GetMaxReplicationFactor()

				Step(
					fmt.Sprintf("repl increase volume driver %s on app %s's volume: %v",
						Inst().V.String(), ctx.App.Key, v),
					func() {
						errExpected := false
						currRep, err := Inst().V.GetReplicationFactor(v)
						UpdateOutcome(event, err)
						expect(err).NotTo(haveOccurred())
						// GetMaxReplicationFactory is hardcoded to 3
						// if it increases repl 3 to an aggregated 2 volume, it will fail
						// because it would require 6 worker nodes, since
						// number of nodes required = aggregation level * replication factor
						currAggr, err := Inst().V.GetAggregationLevel(v)
						UpdateOutcome(event, err)
						expect(err).NotTo(haveOccurred())
						if currAggr > 1 {
							MaxRF = int64(len(node.GetWorkerNodes())) / currAggr
						}
						if currRep == MaxRF {
							errExpected = true
						}
						expReplMap[v] = int64(math.Min(float64(MaxRF), float64(currRep)+1))
						err = Inst().V.SetReplicationFactor(v, currRep+1, opts)
						if !errExpected {
							UpdateOutcome(event, err)
							expect(err).NotTo(haveOccurred())
						} else {
							if !expect(err).To(haveOccurred()) {
								UpdateOutcome(event, fmt.Errorf("Expected HA increase to fail since new repl factor is greater than %v but it did not", MaxRF))
							}
						}
					})
				Step(
					fmt.Sprintf("validate successful repl increase on app %s's volume: %v",
						ctx.App.Key, v),
					func() {
						newRepl, err := Inst().V.GetReplicationFactor(v)
						UpdateOutcome(event, err)
						expect(err).NotTo(haveOccurred())
						if newRepl != expReplMap[v] {
							err = fmt.Errorf("volume has invalid repl value. Expected:%d Actual:%d", expReplMap[v], newRepl)
							UpdateOutcome(event, err)
						}
						expect(newRepl).To(equal(expReplMap[v]))
					})
			}
			Step(fmt.Sprintf("validating context after increasing HA for app: %s",
				ctx.App.Key), func() {
				errorChan := make(chan error, errorChannelSize)
				ctx.SkipVolumeValidation = false
				ValidateContext(ctx, &errorChan)
				for err := range errorChan {
					UpdateOutcome(event, err)
				}
			})
		}
	})
}

// TriggerHADecrease performs repl-reduce on all volumes of given contexts
func TriggerHADecrease(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	defer ginkgo.GinkgoRecover()
	event := &EventRecord{
		Event: Event{
			ID:   GenerateUUID(),
			Type: HADecrease,
		},
		Start:   time.Now().Format(time.RFC1123),
		Outcome: []error{},
	}

	defer func() {
		event.End = time.Now().Format(time.RFC1123)
		*recordChan <- event
	}()

	expReplMap := make(map[*volume.Volume]int64)
	Step("get volumes for all apps in test and decrease replication factor", func() {
		for _, ctx := range contexts {
			var appVolumes []*volume.Volume
			var err error
			Step(fmt.Sprintf("get volumes for %s app", ctx.App.Key), func() {
				appVolumes, err = Inst().S.GetVolumes(ctx)
				UpdateOutcome(event, err)
				expect(appVolumes).NotTo(beEmpty())
			})
			opts := volume.Options{
				ValidateReplicationUpdateTimeout: validateReplicationUpdateTimeout,
			}
			for _, v := range appVolumes {
				MinRF := Inst().V.GetMinReplicationFactor()

				Step(
					fmt.Sprintf("repl decrease volume driver %s on app %s's volume: %v",
						Inst().V.String(), ctx.App.Key, v),
					func() {
						errExpected := false
						currRep, err := Inst().V.GetReplicationFactor(v)
						UpdateOutcome(event, err)
						expect(err).NotTo(haveOccurred())

						if currRep == MinRF {
							errExpected = true
						}
						expReplMap[v] = int64(math.Max(float64(MinRF), float64(currRep)-1))

						err = Inst().V.SetReplicationFactor(v, currRep-1, opts)
						if !errExpected {
							UpdateOutcome(event, err)
							expect(err).NotTo(haveOccurred())
						} else {
							if !expect(err).To(haveOccurred()) {
								UpdateOutcome(event, fmt.Errorf("Expected HA reduce to fail since new repl factor is less than %v but it did not", MinRF))
							}
						}

					})
				Step(
					fmt.Sprintf("validate successful repl decrease on app %s's volume: %v",
						ctx.App.Key, v),
					func() {
						newRepl, err := Inst().V.GetReplicationFactor(v)
						UpdateOutcome(event, err)
						expect(err).NotTo(haveOccurred())
						if newRepl != expReplMap[v] {
							UpdateOutcome(event, fmt.Errorf("volume has invalid repl value. Expected:%d Actual:%d", expReplMap[v], newRepl))
						}
						expect(newRepl).To(equal(expReplMap[v]))
					})
			}
			Step(fmt.Sprintf("validating context after reducing HA for app: %s",
				ctx.App.Key), func() {
				errorChan := make(chan error, errorChannelSize)
				ctx.SkipVolumeValidation = false
				ValidateContext(ctx, &errorChan)
				for err := range errorChan {
					UpdateOutcome(event, err)
				}
			})
		}
	})
}

// TriggerAppTaskDown deletes application task for all contexts
func TriggerAppTaskDown(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	defer ginkgo.GinkgoRecover()
	event := &EventRecord{
		Event: Event{
			ID:   GenerateUUID(),
			Type: AppTaskDown,
		},
		Start:   time.Now().Format(time.RFC1123),
		Outcome: []error{},
	}

	defer func() {
		event.End = time.Now().Format(time.RFC1123)
		*recordChan <- event
	}()

	for _, ctx := range contexts {
		Step(fmt.Sprintf("delete tasks for app: [%s]", ctx.App.Key), func() {
			err := Inst().S.DeleteTasks(ctx, nil)
			UpdateOutcome(event, err)
			expect(err).NotTo(haveOccurred())
		})

		Step(fmt.Sprintf("validating context after delete tasks for app: [%s]",
			ctx.App.Key), func() {
			errorChan := make(chan error, errorChannelSize)
			ctx.SkipVolumeValidation = false
			ValidateContext(ctx, &errorChan)
			for err := range errorChan {
				UpdateOutcome(event, err)
			}
		})
	}
}

// TriggerCrashVolDriver crashes vol driver
func TriggerCrashVolDriver(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	defer ginkgo.GinkgoRecover()
	event := &EventRecord{
		Event: Event{
			ID:   GenerateUUID(),
			Type: CrashVolDriver,
		},
		Start:   time.Now().Format(time.RFC1123),
		Outcome: []error{},
	}

	defer func() {
		event.End = time.Now().Format(time.RFC1123)
		*recordChan <- event
	}()
	Step("crash volume driver in all nodes", func() {
		for _, appNode := range node.GetStorageDriverNodes() {
			Step(
				fmt.Sprintf("crash volume driver %s on node: %v",
					Inst().V.String(), appNode.Name),
				func() {
					CrashVolDriverAndWait([]node.Node{appNode})
				})
		}
	})
}

// TriggerRestartVolDriver restarts volume driver and validates app
func TriggerRestartVolDriver(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	defer ginkgo.GinkgoRecover()
	event := &EventRecord{
		Event: Event{
			ID:   GenerateUUID(),
			Type: RestartVolDriver,
		},
		Start:   time.Now().Format(time.RFC1123),
		Outcome: []error{},
	}

	defer func() {
		event.End = time.Now().Format(time.RFC1123)
		*recordChan <- event
	}()
	Step("get nodes bounce volume driver", func() {
		for _, appNode := range node.GetStorageDriverNodes() {
			Step(
				fmt.Sprintf("stop volume driver %s on node: %s",
					Inst().V.String(), appNode.Name),
				func() {
					StopVolDriverAndWait([]node.Node{appNode})
				})

			Step(
				fmt.Sprintf("starting volume %s driver on node %s",
					Inst().V.String(), appNode.Name),
				func() {
					StartVolDriverAndWait([]node.Node{appNode})
				})

			Step("Giving few seconds for volume driver to stabilize", func() {
				time.Sleep(20 * time.Second)
			})

			for _, ctx := range contexts {
				Step(fmt.Sprintf("RestartVolDriver: validating app [%s]", ctx.App.Key), func() {
					errorChan := make(chan error, errorChannelSize)
					ValidateContext(ctx, &errorChan)
					for err := range errorChan {
						UpdateOutcome(event, err)
					}
				})
			}
		}
	})
}

// TriggerRebootNodes reboots node on which apps are running
func TriggerRebootNodes(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	defer ginkgo.GinkgoRecover()
	event := &EventRecord{
		Event: Event{
			ID:   GenerateUUID(),
			Type: RebootNode,
		},
		Start:   time.Now().Format(time.RFC1123),
		Outcome: []error{},
	}

	defer func() {
		event.End = time.Now().Format(time.RFC1123)
		*recordChan <- event
	}()

	Step("get all nodes and reboot one by one", func() {
		nodesToReboot := node.GetWorkerNodes()

		// Reboot node and check driver status
		Step(fmt.Sprintf("reboot node one at a time from the node(s): %v", nodesToReboot), func() {
			// TODO: Below is the same code from existing nodeReboot test
			for _, n := range nodesToReboot {
				if n.IsStorageDriverInstalled {
					Step(fmt.Sprintf("reboot node: %s", n.Name), func() {
						err := Inst().N.RebootNode(n, node.RebootNodeOpts{
							Force: true,
							ConnectionOpts: node.ConnectionOpts{
								Timeout:         1 * time.Minute,
								TimeBeforeRetry: 5 * time.Second,
							},
						})
						expect(err).NotTo(haveOccurred())
						UpdateOutcome(event, err)
					})

					Step(fmt.Sprintf("wait for node: %s to be back up", n.Name), func() {
						err := Inst().N.TestConnection(n, node.ConnectionOpts{
							Timeout:         15 * time.Minute,
							TimeBeforeRetry: 10 * time.Second,
						})
						expect(err).NotTo(haveOccurred())
						UpdateOutcome(event, err)
					})

					Step(fmt.Sprintf("wait for volume driver to stop on node: %v", n.Name), func() {
						err := Inst().V.WaitDriverDownOnNode(n)
						expect(err).NotTo(haveOccurred())
						UpdateOutcome(event, err)
					})

					Step(fmt.Sprintf("wait to scheduler: %s and volume driver: %s to start",
						Inst().S.String(), Inst().V.String()), func() {

						err := Inst().S.IsNodeReady(n)
						expect(err).NotTo(haveOccurred())
						UpdateOutcome(event, err)

						err = Inst().V.WaitDriverUpOnNode(n, Inst().DriverStartTimeout)
						expect(err).NotTo(haveOccurred())
						UpdateOutcome(event, err)
					})

					Step("validate apps", func() {
						for _, ctx := range contexts {
							Step(fmt.Sprintf("RebootNode: validating app [%s]", ctx.App.Key), func() {
								errorChan := make(chan error, errorChannelSize)
								ValidateContext(ctx, &errorChan)
								for err := range errorChan {
									UpdateOutcome(event, err)
								}
							})
						}
					})
				}
			}
		})
	})
}

// TriggerBackupApps takes backups of all namespaces of deployed apps
func TriggerBackupApps(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	defer ginkgo.GinkgoRecover()
	event := &EventRecord{
		Event: Event{
			ID:   GenerateUUID(),
			Type: BackupAllApps,
		},
		Start:   time.Now().Format(time.RFC1123),
		Outcome: []error{},
	}

	defer func() {
		event.End = time.Now().Format(time.RFC1123)
		*recordChan <- event
	}()
	Step("Update admin secret", func() {
		err := backup.UpdatePxBackupAdminSecret()
		ProcessErrorWithMessage(event, err, "Unable to update PxBackupAdminSecret")
	})
	BackupCounter++
	bkpNamespaces := make([]string, 0)
	labelSelectors := make(map[string]string)
	for _, ctx := range contexts {
		namespace := ctx.GetID()
		bkpNamespaces = append(bkpNamespaces, namespace)
	}
	Step("Backup all namespaces", func() {
		bkpNamespaceErrors := make(map[string]error)
		sourceClusterConfigPath, err := getSourceClusterConfigPath()
		UpdateOutcome(event, err)
		SetClusterContext(sourceClusterConfigPath)
		for _, namespace := range bkpNamespaces {
			backupName := fmt.Sprintf("%s-%s-%d", backupNamePrefix, namespace, BackupCounter)
			Step(fmt.Sprintf("Create backup full name %s:%s:%s",
				sourceClusterName, namespace, backupName), func() {
				err = CreateBackupGetErr(backupName,
					sourceClusterName, backupLocationName, backupLocationUID,
					[]string{namespace}, labelSelectors, orgID)
				if err != nil {
					bkpNamespaceErrors[namespace] = err
				}
				UpdateOutcome(event, err)
			})
		}
		for _, namespace := range bkpNamespaces {
			backupName := fmt.Sprintf("%s-%s-%d", backupNamePrefix, namespace, BackupCounter)
			err, ok := bkpNamespaceErrors[namespace]
			if ok {
				logrus.Warningf("Skipping waiting for backup %s because %s", backupName, err)
				continue
			}
			Step(fmt.Sprintf("Wait for backup %s to complete", backupName), func() {
				ctx, err := backup.GetPxCentralAdminCtx()
				if err != nil {
					logrus.Errorf("Failed to fetch px-central-admin ctx: [%v]", err)
					bkpNamespaceErrors[namespace] = err
					UpdateOutcome(event, err)
				} else {
					err = Inst().Backup.WaitForBackupCompletion(
						ctx,
						backupName, orgID,
						backupRestoreCompletionTimeoutMin*time.Minute,
						retrySeconds*time.Second)
					if err == nil {
						logrus.Infof("Backup [%s] completed successfully", backupName)
					} else {
						logrus.Errorf("Failed to wait for backup [%s] to complete. Error: [%v]",
							backupName, err)
						bkpNamespaceErrors[namespace] = err
						UpdateOutcome(event, err)
					}
				}
			})
		}
	})
	logrus.Infof("Finished TriggerBackupApps")
}

// TriggerInspectScheduledBackup creates scheduled backup if it doesn't exist and makes sure backups are correct otherwise
func TriggerInspectScheduledBackup(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	defer ginkgo.GinkgoRecover()
	event := &EventRecord{
		Event: Event{
			ID:   GenerateUUID(),
			Type: InspectScheduledBackups,
		},
		Start:   time.Now().Format(time.RFC1123),
		Outcome: []error{},
	}

	defer func() {
		event.End = time.Now().Format(time.RFC1123)
		*recordChan <- event
	}()

	err := backup.UpdatePxBackupAdminSecret()
	ProcessErrorWithMessage(event, err, "Unable to update PxBackupAdminSecret")

	err = DeleteNamespace()
	ProcessErrorWithMessage(event, err, "Failed to delete namespace")

	err = CreateNamespace()
	ProcessErrorWithMessage(event, err, "Failed to create namespace")

	logrus.Infof("Enumerating backups")
	bkpEnumerateReq := &api.BackupEnumerateRequest{
		OrgId: orgID}
	ctx, err := backup.GetPxCentralAdminCtx()
	ProcessErrorWithMessage(event, err, "Failed to get px-central admin context")
	curBackups, err := Inst().Backup.EnumerateBackup(ctx, bkpEnumerateReq)
	ProcessErrorWithMessage(event, err, "Enumerate backup request failed")

	waitForNBackups := curBackups.GetTotalCount() + 1
	for _, bkp := range curBackups.GetBackups() {
		if bkp.GetStatus().GetStatus() == api.BackupInfo_StatusInfo_DeletePending ||
			bkp.GetStatus().GetStatus() == api.BackupInfo_StatusInfo_Deleting {
			waitForNBackups--
		}
	}

	_, err = InspectScheduledBackup()
	if ObjectExists(err) {
		err = CreateScheduledBackup()
		ProcessErrorWithMessage(event, err, "Create scheduled backup failed")
	} else if err != nil {
		ProcessErrorWithMessage(event, err, "Inspecting scheduled backup failed")
	}

	logrus.Infof("Waiting for another backup")

	// Wait for 1 more backup to get created
	err = Inst().Backup.BackupScheduleWaitForNBackupsCompletion(
		ctx,
		orgID,
		orgID,
		int(waitForNBackups),
		scheduledBackupInterval*2,
		defaultRetryInterval,
	)
	ProcessErrorWithMessage(event, err, "Failed to wait for backup to be created")

	logrus.Infof("Verify namespaces")
	// Verify that all namespaces are present in latest backup
	curBackups, err = Inst().Backup.EnumerateBackup(ctx, bkpEnumerateReq)
	ProcessErrorWithMessage(event, err, "Enumerate backup request failed")
	latestBkp := curBackups.GetBackups()[0]
	latestBkpNamespaces := latestBkp.GetNamespaces()
	namespacesList, err := core.Instance().ListNamespaces(nil)
	ProcessErrorWithMessage(event, err, "List namespaces failed")

	if len(namespacesList.Items) != len(latestBkpNamespaces) {
		err = fmt.Errorf("backup backed up %d namespaces, but %d namespaces exist", len(latestBkpNamespaces),
			len(namespacesList.Items))
		ProcessErrorWithMessage(event, err, "Scheduled backup backed up wrong namespaces")
	}

	var namespaces []string

	for _, ns := range namespacesList.Items {
		namespaces = append(namespaces, ns.GetName())
	}

	sort.Strings(namespaces)
	sort.Strings(latestBkpNamespaces)

	for i, ns := range namespaces {
		if latestBkpNamespaces[i] != ns {
			err = fmt.Errorf("namespace %s not present in backup", ns)
			ProcessErrorWithMessage(event, err, "Scheduled backup backed up wrong namespaces")
		}
	}
}

// TriggerBackupSpecificResource backs up a specific resource in a namespace
// Creates config maps in the the specified namespaces and backups up only these config maps
func TriggerBackupSpecificResource(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	defer ginkgo.GinkgoRecover()
	event := &EventRecord{
		Event: Event{
			ID:   GenerateUUID(),
			Type: BackupSpecificResource,
		},
		Start:   time.Now().Format(time.RFC1123),
		Outcome: []error{},
	}
	namespaceResourceMap := make(map[string][]string)
	err := backup.UpdatePxBackupAdminSecret()
	ProcessErrorWithMessage(event, err, "Unable to update PxBackupAdminSecret")
	if err != nil {
		return
	}
	sourceClusterConfigPath, err := getSourceClusterConfigPath()
	UpdateOutcome(event, err)
	if err != nil {
		return
	}
	SetClusterContext(sourceClusterConfigPath)
	BackupCounter++
	bkpNamespaces := make([]string, 0)
	labelSelectors := make(map[string]string)
	bkpNamespaceErrors := make(map[string]error)
	for _, ctx := range contexts {
		namespace := ctx.GetID()
		bkpNamespaces = append(bkpNamespaces, namespace)
	}
	Step("Create config maps", func() {
		configMapCount := 2
		for _, namespace := range bkpNamespaces {
			for i := 0; i < configMapCount; i++ {
				configName := fmt.Sprintf("%s-%d-%d", namespace, BackupCounter, i)
				cm := &v1.ConfigMap{
					ObjectMeta: meta_v1.ObjectMeta{
						Name:      configName,
						Namespace: namespace,
					},
				}
				_, err := core.Instance().CreateConfigMap(cm)
				ProcessErrorWithMessage(event, err, fmt.Sprintf("Unable to create config map [%s]", configName))
				if err == nil {
					namespaceResourceMap[namespace] = append(namespaceResourceMap[namespace], configName)
				}
			}
		}
	})
	defer func() {
		Step("Clean up config maps", func() {
			for _, namespace := range bkpNamespaces {
				for _, configName := range namespaceResourceMap[namespace] {
					err := core.Instance().DeleteConfigMap(configName, namespace)
					ProcessErrorWithMessage(event, err, fmt.Sprintf("Unable to delete config map [%s]", configName))
				}
			}
		})
		event.End = time.Now().Format(time.RFC1123)
		*recordChan <- event
	}()
	bkpNames := make([]string, 0)
	Step("Create backups", func() {
		for _, namespace := range bkpNamespaces {
			backupName := fmt.Sprintf("%s-%s-%d", backupNamePrefix, namespace, BackupCounter)
			bkpNames = append(bkpNames, namespace)
			logrus.Infof("Create backup full name %s:%s:%s", sourceClusterName, namespace, backupName)
			backupCreateRequest := GetBackupCreateRequest(backupName, sourceClusterName, backupLocationName, backupLocationUID,
				[]string{namespace}, labelSelectors, orgID)
			backupCreateRequest.Name = backupName
			backupCreateRequest.ResourceTypes = []string{"ConfigMap"}
			err = CreateBackupFromRequest(backupName, orgID, backupCreateRequest)
			UpdateOutcome(event, err)
			if err != nil {
				bkpNamespaceErrors[namespace] = err
			}
		}
	})
	for _, namespace := range bkpNames {
		backupName := fmt.Sprintf("%s-%s-%d", backupNamePrefix, namespace, BackupCounter)
		err, ok := bkpNamespaceErrors[namespace]
		if ok {
			logrus.Warningf("Skipping waiting for backup [%s] because [%s]", backupName, err)
			continue
		}
		Step(fmt.Sprintf("Wait for backup [%s] to complete", backupName), func() {
			ctx, err := backup.GetPxCentralAdminCtx()
			if err != nil {
				bkpNamespaceErrors[namespace] = err
				ProcessErrorWithMessage(event, err, fmt.Sprintf("Failed to fetch px-central-admin ctx: [%v]", err))
			} else {
				err = Inst().Backup.WaitForBackupCompletion(
					ctx,
					backupName, orgID,
					backupRestoreCompletionTimeoutMin*time.Minute,
					retrySeconds*time.Second)
				if err == nil {
					logrus.Infof("Backup [%s] completed successfully", backupName)
				} else {
					bkpNamespaceErrors[namespace] = err
					ProcessErrorWithMessage(event, err, fmt.Sprintf("Failed to wait for backup [%s] to complete. Error: [%v]", backupName, err))
				}
			}
		})
	}
	Step("Check that only config maps are backed up", func() {
		for _, namespace := range bkpNames {
			backupName := fmt.Sprintf("%s-%s-%d", backupNamePrefix, namespace, BackupCounter)
			err, ok := bkpNamespaceErrors[namespace]
			if ok {
				logrus.Warningf("Skipping inspecting backup [%s] because [%s]", backupName, err)
				continue
			}
			bkpInspectResp, err := InspectBackup(backupName)
			UpdateOutcome(event, err)
			backupObj := bkpInspectResp.GetBackup()
			cmList, err := core.Instance().GetConfigMaps(namespace, nil)
			//kube-root-ca.crt exists in every namespace but does not get backed up, so we subtract 1 from the count
			if backupObj.GetResourceCount() != uint64(len(cmList.Items)-1) {
				errMsg := fmt.Sprintf("Backup [%s] has an incorrect number of objects, expected [%d], actual [%d]", backupName, len(cmList.Items)-1, backupObj.GetResourceCount())
				err = fmt.Errorf(errMsg)
				ProcessErrorWithMessage(event, err, errMsg)
			}
			for _, resource := range backupObj.GetResources() {
				if resource.GetKind() != "ConfigMap" {
					errMsg := fmt.Sprintf("Backup [%s] contains non configMap resource, expected [configMap], actual [%v]", backupName, resource.GetKind())
					err = fmt.Errorf(errMsg)
					ProcessErrorWithMessage(event, err, errMsg)
				}
			}
		}
	})
	Step("Clean up config maps", func() {
		for _, namespace := range bkpNamespaces {
			for i := 0; i < configMapCount; i++ {
				configName := fmt.Sprintf("%s-%d-%d", namespace, BackupCounter, i)
				err := core.Instance().DeleteConfigMap(configName, namespace)
				ProcessErrorWithMessage(event, err, fmt.Sprintf("Unable to delete config map [%s]", configName))
			}
		}
	})
}

// TriggerInspectBackup inspects backup and checks for errors
func TriggerInspectBackup(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	defer ginkgo.GinkgoRecover()
	event := &EventRecord{
		Event: Event{
			ID:   GenerateUUID(),
			Type: TestInspectBackup,
		},
		Start:   time.Now().Format(time.RFC1123),
		Outcome: []error{},
	}

	defer func() {
		event.End = time.Now().Format(time.RFC1123)
		*recordChan <- event
	}()

	logrus.Infof("Enumerating backups")
	bkpEnumerateReq := &api.BackupEnumerateRequest{
		OrgId: orgID}
	ctx, err := backup.GetPxCentralAdminCtx()
	ProcessErrorWithMessage(event, err, "InspectBackup failed: Failed to get px-central admin context")
	curBackups, err := Inst().Backup.EnumerateBackup(ctx, bkpEnumerateReq)
	ProcessErrorWithMessage(event, err, "InspectBackup failed: Enumerate backup request failed")

	if len(curBackups.GetBackups()) == 0 {
		return
	}

	backupToInspect := curBackups.GetBackups()[0]

	backupInspectRequest := &api.BackupInspectRequest{
		Name:  backupToInspect.GetName(),
		OrgId: backupToInspect.GetOrgId(),
	}
	_, err = Inst().Backup.InspectBackup(ctx, backupInspectRequest)
	desc := fmt.Sprintf("InspectBackup failed: Inspect backup %s failed", backupToInspect.GetName())
	ProcessErrorWithMessage(event, err, desc)

}

// TriggerInspectRestore inspects restore and checks for errors
func TriggerInspectRestore(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	defer ginkgo.GinkgoRecover()
	event := &EventRecord{
		Event: Event{
			ID:   GenerateUUID(),
			Type: TestInspectRestore,
		},
		Start:   time.Now().Format(time.RFC1123),
		Outcome: []error{},
	}

	defer func() {
		event.End = time.Now().Format(time.RFC1123)
		*recordChan <- event
	}()

	logrus.Infof("Enumerating restores")
	restoreEnumerateReq := &api.RestoreEnumerateRequest{
		OrgId: orgID}
	ctx, err := backup.GetPxCentralAdminCtx()
	ProcessErrorWithMessage(event, err, "InspectRestore failed: Failed to get px-central admin context")
	curRestores, err := Inst().Backup.EnumerateRestore(ctx, restoreEnumerateReq)
	ProcessErrorWithMessage(event, err, "InspectRestore failed: Enumerate restore request failed")

	if len(curRestores.GetRestores()) == 0 {
		return
	}

	restoreToInspect := curRestores.GetRestores()[0]

	restoreInspectRequest := &api.RestoreInspectRequest{
		Name:  restoreToInspect.GetName(),
		OrgId: restoreToInspect.GetOrgId(),
	}
	_, err = Inst().Backup.InspectRestore(ctx, restoreInspectRequest)
	desc := fmt.Sprintf("InspectRestore failed: Inspect restore %s failed", restoreToInspect.GetName())
	ProcessErrorWithMessage(event, err, desc)
}

// TriggerRestoreNamespace restores a namespace to a new namespace
func TriggerRestoreNamespace(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	defer ginkgo.GinkgoRecover()
	event := &EventRecord{
		Event: Event{
			ID:   GenerateUUID(),
			Type: RestoreNamespace,
		},
		Start:   time.Now().Format(time.RFC1123),
		Outcome: []error{},
	}

	defer func() {
		sourceClusterConfigPath, err := getSourceClusterConfigPath()
		UpdateOutcome(event, err)
		SetClusterContext(sourceClusterConfigPath)
		event.End = time.Now().Format(time.RFC1123)
		*recordChan <- event
	}()

	RestoreCounter++
	namespacesList, err := core.Instance().ListNamespaces(nil)
	ProcessErrorWithMessage(event, err, "Restore namespace failed: List namespaces failed")

	destClusterConfigPath, err := getDestinationClusterConfigPath()
	ProcessErrorWithMessage(event, err, "Restore namespace failed: getDestinationClusterConfigPath failed")
	SetClusterContext(destClusterConfigPath)

	logrus.Infof("Enumerating backups")
	bkpEnumerateReq := &api.BackupEnumerateRequest{
		OrgId: orgID}
	ctx, err := backup.GetPxCentralAdminCtx()
	ProcessErrorWithMessage(event, err, "Restore namespace failed: Failed to get px-central admin context")
	curBackups, err := Inst().Backup.EnumerateBackup(ctx, bkpEnumerateReq)
	ProcessErrorWithMessage(event, err, "Restore namespace failed: Enumerate backup request failed")

	// Get a completed backup
	var backupToRestore *api.BackupObject
	backupToRestore = nil
	for _, bkp := range curBackups.GetBackups() {
		if bkp.GetStatus().GetStatus() == api.BackupInfo_StatusInfo_PartialSuccess ||
			bkp.GetStatus().GetStatus() == api.BackupInfo_StatusInfo_Success {
			backupToRestore = bkp
			break
		}
	}
	// If there is nothing to restore, return
	if backupToRestore == nil {
		return
	}
	restoreName := fmt.Sprintf("%s-%d", backupToRestore.GetName(), RestoreCounter)

	// Pick one namespace to restore
	// In case destination cluster == source cluster, restore to a new namespace
	namespaceMapping := make(map[string]string)
	namespaces := backupToRestore.GetNamespaces()
	if len(namespaces) > 0 {
		ns := namespaces[0]
		namespaceMapping[ns] = fmt.Sprintf("%s-restore-%s-%d", ns, Inst().InstanceID, RestoreCounter)
	}

	restoreCreateRequest := &api.RestoreCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name:  restoreName,
			OrgId: orgID,
		},
		Backup:           backupToRestore.GetName(),
		Cluster:          destinationClusterName,
		NamespaceMapping: namespaceMapping,
	}
	_, err = Inst().Backup.CreateRestore(ctx, restoreCreateRequest)
	desc := fmt.Sprintf("Restore namespace failed: Create restore %s failed", restoreName)
	ProcessErrorWithMessage(event, err, desc)

	err = Inst().Backup.WaitForRestoreCompletion(ctx, restoreName, orgID,
		backupRestoreCompletionTimeoutMin*time.Minute,
		retrySeconds*time.Second)
	desc = fmt.Sprintf("Restore namespace failed: Failed to wait for restore [%s] to complete.", restoreName)
	ProcessErrorWithMessage(event, err, desc)

	// Validate that one namespace is restored
	newNamespacesList, err := core.Instance().ListNamespaces(nil)
	ProcessErrorWithMessage(event, err, "Restore namespace failed: List namespaces failed")

	if len(newNamespacesList.Items) != len(namespacesList.Items)+1 {
		err = fmt.Errorf("restored %d namespaces instead of 1, %s",
			len(newNamespacesList.Items)-len(namespacesList.Items), restoreName)
		ProcessErrorWithMessage(event, err, "RestoreNamespace restored incorrect namespaces")
	}

	nsFound := false
	for _, ns := range newNamespacesList.Items {
		if ns.GetName() == namespaces[0] {
			nsFound = true
		}
	}
	if !nsFound {
		err = fmt.Errorf("namespace %s not found", namespaces[0])
		ProcessErrorWithMessage(event, err, "RestoreNamespace restored incorrect namespaces")
	}
}

// TriggerDeleteBackup deletes a backup
func TriggerDeleteBackup(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	defer ginkgo.GinkgoRecover()
	event := &EventRecord{
		Event: Event{
			ID:   GenerateUUID(),
			Type: TestDeleteBackup,
		},
		Start:   time.Now().Format(time.RFC1123),
		Outcome: []error{},
	}

	defer func() {
		event.End = time.Now().Format(time.RFC1123)
		*recordChan <- event
	}()

	logrus.Infof("Enumerating backups")
	bkpEnumerateReq := &api.BackupEnumerateRequest{
		OrgId: orgID}
	ctx, err := backup.GetPxCentralAdminCtx()
	ProcessErrorWithMessage(event, err, "DeleteBackup failed: Failed to get px-central admin context")
	curBackups, err := Inst().Backup.EnumerateBackup(ctx, bkpEnumerateReq)
	ProcessErrorWithMessage(event, err, "DeleteBackup failed: Enumerate backup request failed")

	if len(curBackups.GetBackups()) == 0 {
		return
	}

	backupToDelete := curBackups.GetBackups()[0]
	err = DeleteBackupAndDependencies(backupToDelete.GetName(), orgID, backupToDelete.GetCluster())
	desc := fmt.Sprintf("DeleteBackup failed: Delete backup %s on cluster %s failed",
		backupToDelete.GetName(), backupToDelete.GetCluster())
	ProcessErrorWithMessage(event, err, desc)

}

// TriggerBackupSpecificResourceOnCluster backs up all PVCs on the source cluster
func TriggerBackupSpecificResourceOnCluster(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	defer ginkgo.GinkgoRecover()
	event := &EventRecord{
		Event: Event{
			ID:   GenerateUUID(),
			Type: BackupSpecificResourceOnCluster,
		},
		Start:   time.Now().Format(time.RFC1123),
		Outcome: []error{},
	}

	defer func() {
		event.End = time.Now().Format(time.RFC1123)
		*recordChan <- event
	}()
	err := backup.UpdatePxBackupAdminSecret()
	ProcessErrorWithMessage(event, err, "Unable to update PxBackupAdminSecret")
	if err != nil {
		return
	}
	sourceClusterConfigPath, err := getSourceClusterConfigPath()
	UpdateOutcome(event, err)
	if err != nil {
		return
	}
	SetClusterContext(sourceClusterConfigPath)
	BackupCounter++
	backupName := fmt.Sprintf("%s-%s-%d", backupNamePrefix, Inst().InstanceID, BackupCounter)
	namespaces := make([]string, 0)
	labelSelectors := make(map[string]string)
	totalPVC := 0
	Step("Backup all persistent volume claims on source cluster", func() {
		nsList, err := core.Instance().ListNamespaces(labelSelectors)
		UpdateOutcome(event, err)
		if err == nil {
			for _, ns := range nsList.Items {
				namespaces = append(namespaces, ns.Name)
			}
			backupCreateRequest := GetBackupCreateRequest(backupName, sourceClusterName, backupLocationName, backupLocationUID,
				namespaces, labelSelectors, orgID)
			backupCreateRequest.Name = backupName
			backupCreateRequest.ResourceTypes = []string{"PersistentVolumeClaim"}
			err = CreateBackupFromRequest(backupName, orgID, backupCreateRequest)
			UpdateOutcome(event, err)
		}
	})
	if err != nil {
		return
	}
	Step("Wait for backup to complete", func() {
		ctx, err := backup.GetPxCentralAdminCtx()
		if err != nil {
			ProcessErrorWithMessage(event, err, fmt.Sprintf("Failed to fetch px-central-admin ctx: [%v]", err))
		} else {
			err = Inst().Backup.WaitForBackupCompletion(
				ctx,
				backupName, orgID,
				backupRestoreCompletionTimeoutMin*time.Minute,
				retrySeconds*time.Second)
			if err == nil {
				logrus.Infof("Backup [%s] completed successfully", backupName)
			} else {
				ProcessErrorWithMessage(event, err, fmt.Sprintf("Failed to wait for backup [%s] to complete. Error: [%v]", backupName, err))
			}
		}
	})
	if err != nil {
		return
	}
	Step("Check PVCs in backup", func() {
		bkpInspectResp, err := InspectBackup(backupName)
		UpdateOutcome(event, err)
		if err == nil {
			backupObj := bkpInspectResp.GetBackup()
			pvcList, err := core.Instance().GetPersistentVolumes()
			UpdateOutcome(event, err)
			if err == nil {
				if backupObj.GetResourceCount() != uint64(len(pvcList.Items))*2 { //Each backed up PVC should give a PVC and a PV, hence x2
					errMsg := fmt.Sprintf("Backup %s has incorrect number of objects, expected [%d], actual [%d]", backupName, totalPVC, backupObj.GetResourceCount())
					err = fmt.Errorf(errMsg)
					ProcessErrorWithMessage(event, err, errMsg)
				}
				for _, resource := range backupObj.GetResources() {
					if resource.GetKind() != "PersistentVolumeClaim" && resource.GetKind() != "PersistentVolume" {
						errMsg := fmt.Sprintf("Backup %s contains non PersistentVolumeClaim resource of type [%v]", backupName, resource.GetKind())
						err = fmt.Errorf(errMsg)
						ProcessErrorWithMessage(event, err, errMsg)
					}
				}
			}
		}
	})
}

//TriggerBackupByLabel gives a label to random resources on the cluster and tries to back up only resources with that label
func TriggerBackupByLabel(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	defer ginkgo.GinkgoRecover()
	event := &EventRecord{
		Event: Event{
			ID:   GenerateUUID(),
			Type: BackupUsingLabelOnCluster,
		},
		Start:   time.Now().Format(time.RFC1123),
		Outcome: []error{},
	}
	labelKey := "backup-by-label"
	labelValue := uuid.New()
	defer func() {
		Step("Delete the temporary labels", func() {
			nsList, err := core.Instance().ListNamespaces(nil)
			UpdateOutcome(event, err)
			for _, ns := range nsList.Items {
				pvcList, err := core.Instance().GetPersistentVolumeClaims(ns.Name, nil)
				UpdateOutcome(event, err)
				for _, pvc := range pvcList.Items {
					pvcPointer, err := core.Instance().GetPersistentVolumeClaim(pvc.Name, ns.Name)
					UpdateOutcome(event, err)
					if err == nil {
						DeleteLabelFromResource(pvcPointer, labelKey)
					}
				}
				cmList, err := core.Instance().GetConfigMaps(ns.Name, nil)
				UpdateOutcome(event, err)
				for _, cm := range cmList.Items {
					cmPointer, err := core.Instance().GetConfigMap(cm.Name, ns.Name)
					UpdateOutcome(event, err)
					if err == nil {
						DeleteLabelFromResource(cmPointer, labelKey)
					}
				}
				secretList, err := core.Instance().GetSecrets(ns.Name, nil)
				UpdateOutcome(event, err)
				for _, secret := range secretList.Items {
					secretPointer, err := core.Instance().GetConfigMap(secret.Name, ns.Name)
					UpdateOutcome(event, err)
					if err == nil {
						DeleteLabelFromResource(secretPointer, labelKey)
					}
				}
			}
		})
		event.End = time.Now().Format(time.RFC1123)
		*recordChan <- event
	}()
	err := backup.UpdatePxBackupAdminSecret()
	ProcessErrorWithMessage(event, err, "Unable to update PxBackupAdminSecret")
	if err != nil {
		return
	}
	sourceClusterConfigPath, err := getSourceClusterConfigPath()
	UpdateOutcome(event, err)
	if err != nil {
		return
	}
	SetClusterContext(sourceClusterConfigPath)
	BackupCounter++
	backupName := fmt.Sprintf("%s-%s-%d", backupNamePrefix, Inst().InstanceID, BackupCounter)
	namespaces := make([]string, 0)
	labelSelectors := make(map[string]string)
	labeledResources := make(map[string]bool)
	Step("Add labels to random resources", func() {
		nsList, err := core.Instance().ListNamespaces(nil)
		UpdateOutcome(event, err)
		for _, ns := range nsList.Items {
			namespaces = append(namespaces, ns.Name)
			pvcList, err := core.Instance().GetPersistentVolumeClaims(ns.Name, nil)
			UpdateOutcome(event, err)
			for _, pvc := range pvcList.Items {
				pvcPointer, err := core.Instance().GetPersistentVolumeClaim(pvc.Name, ns.Name)
				UpdateOutcome(event, err)
				if err == nil {
					dice := rand.Intn(4)
					if dice == 1 {
						err = AddLabelToResource(pvcPointer, labelKey, labelValue)
						UpdateOutcome(event, err)
						if err == nil {
							resourceName := fmt.Sprintf("%s/%s/PersistentVolumeClaim", ns.Name, pvc.Name)
							labeledResources[resourceName] = true
						}
					}
				}
			}
			cmList, err := core.Instance().GetConfigMaps(ns.Name, nil)
			UpdateOutcome(event, err)
			for _, cm := range cmList.Items {
				cmPointer, err := core.Instance().GetConfigMap(cm.Name, ns.Name)
				UpdateOutcome(event, err)
				if err == nil {
					dice := rand.Intn(4)
					if dice == 1 {
						err = AddLabelToResource(cmPointer, labelKey, labelValue)
						UpdateOutcome(event, err)
						if err == nil {
							resourceName := fmt.Sprintf("%s/%s/ConfigMap", ns.Name, cm.Name)
							labeledResources[resourceName] = true
						}
					}
				}
			}
			secretList, err := core.Instance().GetSecrets(ns.Name, nil)
			UpdateOutcome(event, err)
			for _, secret := range secretList.Items {
				secretPointer, err := core.Instance().GetSecret(secret.Name, ns.Name)
				UpdateOutcome(event, err)
				if err == nil {
					dice := rand.Intn(4)
					if dice == 1 {
						err = AddLabelToResource(secretPointer, labelKey, labelValue)
						UpdateOutcome(event, err)
						if err == nil {
							resourceName := fmt.Sprintf("%s/%s/Secret", ns.Name, secret.Name)
							labeledResources[resourceName] = true
						}
					}
				}
			}
		}
	})
	Step(fmt.Sprintf("Backup using label [%s=%s]", labelKey, labelValue), func() {
		labelSelectors[labelKey] = labelValue
		backupCreateRequest := GetBackupCreateRequest(backupName, sourceClusterName, backupLocationName, backupLocationUID,
			namespaces, labelSelectors, orgID)
		backupCreateRequest.Name = backupName
		err = CreateBackupFromRequest(backupName, orgID, backupCreateRequest)
		UpdateOutcome(event, err)
		if err != nil {
			return
		}
	})
	Step("Wait for backup to complete", func() {
		ctx, err := backup.GetPxCentralAdminCtx()
		if err != nil {
			ProcessErrorWithMessage(event, err, fmt.Sprintf("Failed to fetch px-central-admin ctx: [%v]", err))
		} else {
			err = Inst().Backup.WaitForBackupCompletion(
				ctx,
				backupName, orgID,
				backupRestoreCompletionTimeoutMin*time.Minute,
				retrySeconds*time.Second)
			if err == nil {
				logrus.Infof("Backup [%s] completed successfully", backupName)
			} else {
				ProcessErrorWithMessage(event, err, fmt.Sprintf("Failed to wait for backup [%s] to complete. Error: [%v]", backupName, err))
				return
			}
		}
	})
	Step("Check that we only backed up objects with specified labels", func() {
		bkpInspectResp, err := InspectBackup(backupName)
		UpdateOutcome(event, err)
		if err != nil {
			return
		}
		backupObj := bkpInspectResp.GetBackup()
		for _, resource := range backupObj.GetResources() {
			if resource.GetKind() == "PersistentVolume" { //PV are automatically backed up with PVCs
				continue
			}
			resourceName := fmt.Sprintf("%s/%s/%s", resource.Namespace, resource.Namespace, resource.GetKind())
			if _, ok := labeledResources[resourceName]; !ok {
				err = fmt.Errorf("Backup [%s] has a resource [%s]that shouldn't be there", backupName, resourceName)
				UpdateOutcome(event, err)
			}
		}
	})
}

// CollectEventRecords collects eventRecords from channel
// and stores in buffer for future email notifications
func CollectEventRecords(recordChan *chan *EventRecord) {
	eventRing = ring.New(100)
	for eventRecord := range *recordChan {
		eventRing.Value = eventRecord
		eventRing = eventRing.Next()
	}
}

// TriggerEmailReporter sends email with all reported errors
func TriggerEmailReporter(contexts []*scheduler.Context, recordChan *chan *EventRecord) {
	// emailRecords stores events to be notified
	emailRecords := emailRecords{}
	logrus.Infof("Generating email report: %s", time.Now().Format(time.RFC1123))
	for i := 0; i < eventRing.Len(); i++ {
		record := eventRing.Value
		if record != nil {
			emailRecords.Records = append(emailRecords.Records, *record.(*EventRecord))
			eventRing.Value = nil
		}
		eventRing = eventRing.Next()
	}

	content, err := prepareEmailBody(emailRecords)
	if err != nil {
		logrus.Errorf("Failed to prepare email body. Error: [%v]", err)
	}

	emailDetails := &email.Email{
		Subject:        subject,
		Content:        content,
		From:           from,
		To:             EmailRecipients,
		SendGridAPIKey: SendGridEmailAPIKey,
	}

	err = emailDetails.SendEmail()
	if err != nil {
		logrus.Errorf("Failed to send out email, because of Error: %q", err)
	}
}

func prepareEmailBody(eventRecords emailRecords) (string, error) {
	var err error
	t := template.New("t").Funcs(templateFuncs)
	t, err = t.Parse(htmlTemplate)
	if err != nil {
		logrus.Errorf("Cannot parse HTML template Err: %v", err)
		return "", err
	}
	var buf []byte
	buffer := bytes.NewBuffer(buf)
	err = t.Execute(buffer, eventRecords)
	if err != nil {
		logrus.Errorf("Cannot generate body from values, Err: %v", err)
		return "", err
	}

	return buffer.String(), nil
}

var templateFuncs = template.FuncMap{"rangeStruct": rangeStructer}

func rangeStructer(args ...interface{}) []interface{} {
	if len(args) == 0 {
		return nil
	}

	v := reflect.ValueOf(args[0])
	if v.Kind() != reflect.Struct {
		return nil
	}

	out := make([]interface{}, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		out[i] = v.Field(i).Interface()
	}

	return out
}

var htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<style>
table {
  border-collapse: collapse;
  width: 100%;
}
th {
   background-color: #0ca1f0;
   text-align: center;
   padding: 3px;
}
td {
  text-align: left;
  padding: 3px;
}
tbody tr:nth-child(even) {
	background-color: #bac5ca;
}
tbody tr:last-child {
  background-color: #79ab78;
}
</style>
</head>
<body>
<h1>Torpedo Longevity Report</h1>
<hr/>
<h3>Event Details</h3>
<table border=1>
<tr>
   <td align="center"><h4>Event </h4></td>
   <td align="center"><h4>Start Time </h4></td>
   <td align="center"><h4>End Time </h4></td>
   <td align="center"><h4>Errors </h4></td>
 </tr>
{{range .Records}}<tr>
{{range rangeStruct .}}	<td>{{.}}</td>
{{end}}</tr>
{{end}}
</table>
<hr/>
</table>
</body>
</html>
`
