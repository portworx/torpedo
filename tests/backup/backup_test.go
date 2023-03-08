package tests

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/backup/portworx"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/pkg/log"

	. "github.com/portworx/torpedo/tests"
)

// This testcase verifies if the backup pods are in Ready state or not
var _ = Describe("{BackupClusterVerification}", func() {
	JustBeforeEach(func() {
		log.Infof("No pre-setup required for this testcase")
		StartTorpedoTest("Backup: BackupClusterVerification", "Validating backup cluster pods", nil, 0)
	})
	It("Backup Cluster Verification", func() {
		Step("Check the status of backup pods", func() {
			log.InfoD("Check the status of backup pods")
			err := Inst().Backup.ValidateBackupCluster()
			dash.VerifyFatal(err, nil, "Backup Cluster Verification successful")
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		log.Infof("No cleanup required for this testcase")
	})
})

// This is a sample test case to verify User/Group Management and role mapping
var _ = Describe("{UserGroupManagement}", func() {
	JustBeforeEach(func() {
		log.Infof("No pre-setup required for this testcase")
		StartTorpedoTest("Backup: UserGroupManagement", "Creating users and adding them to groups", nil, 0)
	})
	It("User and group role mappings", func() {
		Step("Create Users", func() {
			err := backup.AddUser("testuser1", "test", "user1", "testuser1@localhost.com", "Password1")
			log.FailOnError(err, "Failed to create user")
		})
		Step("Create Groups", func() {
			err := backup.AddGroup("testgroup1")
			log.FailOnError(err, "Failed to create group")
		})
		Step("Add users to group", func() {
			err := backup.AddGroupToUser("testuser1", "testgroup1")
			log.FailOnError(err, "Failed to assign group to user")
		})
		Step("Assign role to groups", func() {
			err := backup.AddRoleToGroup("testgroup1", backup.ApplicationOwner, "testing from torpedo")
			log.FailOnError(err, "Failed to assign group to user")
		})
		Step("Verify Application Owner role permissions for user", func() {
			isUserRoleMapped, err := ValidateUserRole("testuser1", backup.ApplicationOwner)
			log.FailOnError(err, "User does not contain the expected role")
			dash.VerifyFatal(isUserRoleMapped, true, "Verifying the user role mapping")
		})
		Step("Update role to groups", func() {
			err := backup.DeleteRoleFromGroup("testgroup1", backup.ApplicationOwner, "removing role from testgroup1")
			log.FailOnError(err, "Failed to delete role from group")
			err = backup.AddRoleToGroup("testgroup1", backup.ApplicationUser, "testing from torpedo")
			log.FailOnError(err, "Failed to add role to group")
		})
		Step("Verify Application User role permissions for user", func() {
			isUserRoleMapped, err := ValidateUserRole("testuser1", backup.ApplicationUser)
			log.FailOnError(err, "User does not contain the expected role")
			dash.VerifyFatal(isUserRoleMapped, true, "Verifying the user role mapping")
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		log.Infof("Cleanup started")
		err := backup.DeleteUser("testuser1")
		dash.VerifySafely(err, nil, "Delete user testuser1")
		err = backup.DeleteGroup("testgroup1")
		dash.VerifySafely(err, nil, "Delete group testgroup1")
		log.Infof("Cleanup done")
	})
})

var _ = Describe("{BkpRstrDiffK8sVerSimultaneousDiffNS}", func() {

	var (
		appList                          = Inst().AppList
		backupNames                      []string
		sourceClusterAppsContexts        []*scheduler.Context    // Each Context is for one Namespace which corresponds to one App
		destinationClusterAppsContexts   []*scheduler.Context    // Each Context is for one Namespace which corresponds to one App
		backupContexts                   []*BackupRestoreContext // Each Context is for one backup
		restoreContexts                  []*BackupRestoreContext // Each Context is for one restore
		restoreLaterContexts             []*BackupRestoreContext // Each Context is for one restore-later
		preRuleNameList                  []string
		postRuleNameList                 []string
		appContexts                      []*scheduler.Context // All app contexts (namespaces) in this loop of the scaling
		sourceNamespaces                 []string
		destinationNamespaces            []string
		namespaceMapping                 map[string]string
		clusterUid                       string
		clusterStatus                    api.ClusterInfo_StatusInfo_Status
		cloudCredName                    string
		cloudCredUID                     string
		backupLocationUID                string
		bkpLocationName                  string
		srcMaj, srcMin, destMaj, destMin int64
	)

	backupLocationMap := make(map[string]string)
	labelSelectors := make(map[string]string)
	sourceNamespaces = make([]string, 0)
	destinationNamespaces = make([]string, 0)
	namespaceMapping = make(map[string]string)
	providers := getProviders()

	JustBeforeEach(func() {

		StartTorpedoTest("BkpRstrDiffK8sVerSimultaneousDiffNS", "Backup on NS(=yy), K8-version(=x) ;; Restore on NS=yy, K8s-version=x+?(=z) [success]; Simultaneously, Restore on K8s-version=z, NS=abc [partial success]", nil, 0)

		log.InfoD("Verifying if the pre/post rules for the required apps are present in the AppParameters or not ")
		for i := 0; i < len(appList); i++ {
			if Contains(postRuleApp, appList[i]) {
				if _, ok := portworx.AppParameters[appList[i]]["post"]; ok {
					dash.VerifyFatal(ok, true, "Post Rule details mentioned for the apps")
				}
			}
			if Contains(preRuleApp, appList[i]) {
				if _, ok := portworx.AppParameters[appList[i]]["pre"]; ok {
					dash.VerifyFatal(ok, true, "Pre Rule details mentioned for the apps")
				}
			}
		}

	})

	It("Backup on NS(=yy), K8-version(=x) ;; Restore on NS=yy, K8s-version=x+?(=z) [success]; Simultaneously, Restore on K8s-version=z, NS=abc [partial success]", func() {

		Step("Verify if 'test app spec'", func() {

			log.InfoD("Allowed apps are %v", appsWithCRDsAndWebhooks)
			for i := 0; i < len(appList); i++ {
				contains := Contains(appsWithCRDsAndWebhooks, appList[i])
				dash.VerifyFatal(contains, true,
					fmt.Sprintf("check if app [%s] can be used as a dummy spec for *this* test", appList[i]))
			}
		})

		Step("Verify K8s version of Src and Dest Cluster", func() {
			log.InfoD("Verify K8s version of Src and Dest Cluster")

			Step("Register cluster for backup", func() {
				log.InfoD("Register cluster for backup")
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "Fetching px-central-admin ctx")
				err = CreateSourceAndDestClusters(orgID, "", "", ctx)
				dash.VerifyFatal(err, nil, "Creating source and destination cluster")
				clusterStatus, clusterUid = Inst().Backup.RegisterBackupCluster(orgID, SourceClusterName, "")
				dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, "Verifying backup cluster")
			})

			Step("Switch Context (\"Source\")", func() {
				sourceClusterConfigPath, err := GetSourceClusterConfigPath()
				log.FailOnError(err, "Failed to get kubeconfig path for source cluster. Error: [%v]", err)

				err = Inst().S.SetConfig(sourceClusterConfigPath)
				log.FailOnError(err, "Failed to switch to context to source cluster [%v]", sourceClusterConfigPath)

				srcMaj, srcMin, err = k8s.ClusterVersion()
				log.FailOnError(err, "Failed to get Source Cluster Version")
			})

			Step("Switch Context (\"destination\")", func() {
				destinationClusterConfigPath, err := GetDestinationClusterConfigPath()
				log.FailOnError(err, "Failed to get kubeconfig path for destination cluster. Error: [%v]", err)

				err = Inst().S.SetConfig(destinationClusterConfigPath)
				log.FailOnError(err, "Failed to switch to context to destination cluster [%v]", destinationClusterConfigPath)

				destMaj, destMin, err = k8s.ClusterVersion()
				log.FailOnError(err, "Failed to get Destination Cluster Version")
			})

			Step("Compare Source and Destination cluster version numbers", func() {

				if srcMaj != 0 && destMaj != 0 {

					log.InfoD("Source Cluster version: %d.%d ; Destination Cluster version: %d.%d",
						srcMaj, srcMin, destMaj, destMin)

					const multiple = 100
					isValid := (destMaj*multiple + destMin) >= (srcMaj*multiple + srcMin)

					dash.VerifyFatal(isValid, true,
						"This test is only meant for cases where the Source Cluster version is LESSER than the Destination Cluster version.")

				} else {
					err := fmt.Errorf("Cannot compare Source and Destination CLuster versions due to invalid data in Source Cluster version (%d.%d), Destination Cluster (%d.%d)", srcMaj, srcMin, destMaj, destMin)
					log.FailOnError(err, "Failed to validate K8s versions")
				}

			})

			// Step("SetClusterContext \"\" (setting it back to default)") is not needed
			// since the next statement (in the next step) is anyway a sontect switch to source
			// err := SetClusterContext("")
			// log.FailOnError(err, "Failed to SetClusterContext to default cluster")

		})

		Step("Deploy the applications on Src cluster", func() {
			log.InfoD("Deploy the applications on Src cluster")

			Step("Deploy applications", func() {
				log.InfoD("Deploy applications")

				log.InfoD("Switching to source context")
				err := SetSourceKubeConfig()
				log.FailOnError(err, "Failed to switch to context to source cluster")

				sourceClusterAppsContexts = make([]*scheduler.Context, 0)
				for i := 0; i < Inst().GlobalScaleFactor; i++ {
					taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
					appContexts = ScheduleApplications(taskName)
					for _, appCtx := range appContexts {
						appCtx.ReadinessTimeout = appReadinessTimeout
						namespace := GetAppNamespace(appCtx, taskName)
						// (sourceNamespaces, sourceClusterAppsContexts) will always correspoond
						sourceNamespaces = append(sourceNamespaces, namespace)
						sourceClusterAppsContexts = append(sourceClusterAppsContexts, appCtx)
					}
				}
			})

			Step("Validate applications", func() {
				ValidateApplications(sourceClusterAppsContexts)

				log.InfoD("Switching to default context")
				err := SetClusterContext("")
				log.FailOnError(err, "Failed to SetClusterContext to default cluster")
			})

			log.Warn("Waiting for 2 minutes, hoping for any Custom Resources to finish starting up.\nThis logic will have to be changed to polling in the future, as just randomly waiting, without verification can easily lead to errors. Make sure that any errors you're seeing is not due to this.")
			time.Sleep(time.Minute * 2)

		})

		Step("Creating rules for backup", func() {
			log.InfoD("Creating pre rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				preRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], orgID, "pre")
				log.FailOnError(err, "Creating pre rule for deployed apps failed")
				dash.VerifyFatal(preRuleStatus, true, "Verifying pre rule for backup")

				if ruleName != "" {
					preRuleNameList = append(preRuleNameList, ruleName)
				}
			}
			log.InfoD("Creating post rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				postRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], orgID, "post")
				log.FailOnError(err, "Creating post rule for deployed apps failed")
				dash.VerifyFatal(postRuleStatus, true, "Verifying Post rule for backup")
				if ruleName != "" {
					postRuleNameList = append(postRuleNameList, ruleName)
				}
			}
		})

		Step("Creating bucket,backup location and cloud setting", func() {
			log.InfoD("Creating bucket,backup location and cloud setting")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				bkpLocationName = fmt.Sprintf("%s-%s-bl", provider, getGlobalBucketName(provider))
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = bkpLocationName
				CreateCloudCredential(provider, cloudCredName, cloudCredUID, orgID)
				err := CreateBackupLocation(provider, bkpLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), orgID, "")
				dash.VerifyFatal(err, nil, "Creating backup location")
			}
		})

		// Step("Register cluster for backup" is not needed as it has been executed as psrt of K8s version validation

		Step("Taking backup of application fron source cluster", func() {
			log.InfoD("Taking backup of applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			dash.VerifyFatal(err, nil, "Getting context")
			backupNames = make([]string, len(sourceNamespaces))
			backupContexts = make([]*BackupRestoreContext, len(sourceNamespaces))
			for i, namespace := range sourceNamespaces {
				backupName := fmt.Sprintf("%s-%s-%v", BackupNamePrefix, namespace, time.Now().Unix())
				backupNames[i] = backupName
				backupCtx, err := CreateBackupAndGetBackupCtx(backupName, SourceClusterName, bkpLocationName, backupLocationUID, []string{namespace}, labelSelectors, orgID, clusterUid, "", "", "", "", ctx, []*scheduler.Context{sourceClusterAppsContexts[i]})
				dash.VerifyFatal(err, nil, "Verifying backup creation")
				backupContexts[i] = backupCtx
			}
		})

		Step("Restoring the backed up applications on destination cluster", func() {

			log.InfoD("Restoring the backed up applications on destination cluster")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			for i, sourceNamespace := range sourceNamespaces {
				var initialRestoreName string

				Step("Restoring the backed up application to namespace of same name on destination cluster", func() {
					log.InfoD("Restoring the backed up application to namespace of same name on destination cluster")

					initialRestoreName = fmt.Sprintf("%s-%s-initial-%v", restoreNamePrefix, sourceNamespace, time.Now().Unix())
					destinationNameSpace := sourceNamespace
					destinationNamespaces = append(destinationNamespaces, destinationNameSpace)
					namespaceMapping[sourceNamespace] = destinationNameSpace
					err = CreateRestoreWithoutCheck(initialRestoreName, backupNames[i], namespaceMapping, destinationClusterName, orgID, ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Initiation of restore %s", initialRestoreName))

					restoreInspectRequest := &api.RestoreInspectRequest{
						Name:  initialRestoreName,
						OrgId: orgID,
					}
					restoreInProgressCheck := func() (interface{}, bool, error) {
						resp, err := Inst().Backup.InspectRestore(ctx, restoreInspectRequest)
						restoreResponseStatus := resp.GetRestore().GetStatus()
						log.FailOnError(err, "Failed getting restore status for - %s", initialRestoreName)
						// Status should be LATER than InProgress in order for next STEP to execute
						if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_InProgress {
							log.Infof("Restore status - %s", restoreResponseStatus)
							log.InfoD("Status of %s - [%s]",
								initialRestoreName, restoreResponseStatus.GetStatus())
							log.InfoD("Condition fulfilled. Proceeding towrds Restore 2")
							return "", false, nil
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_PartialSuccess {
							log.Infof("Restore status - %s", restoreResponseStatus)
							log.InfoD("Status of %s - [%s]",
								initialRestoreName, restoreResponseStatus.GetStatus())

							err := fmt.Errorf("Status of Restore 1 (%s) is PartialSuccess. This should not happen. This MAY have happened if the destination cluster hadn't been cleaned completely (remanant cluster-level resources). Also, you may need to check more frequently for InProgress State as we cannot continue with the test EVEN if PartialSuccess was a valid state.", initialRestoreName)
							log.FailOnError(err, "Unexpected Status")
							return "", false, err
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Success {
							log.Infof("Restore status - %s", restoreResponseStatus)
							log.InfoD("Status of %s - [%s]",
								initialRestoreName, restoreResponseStatus.GetStatus())

							err := fmt.Errorf("Oops! The restore seems to have progressed to Success. We cannot proceed with this test. Maybe check more frequently for Inprogress so that you don't miss the state")
							log.FailOnError(err, "Unexpected Status")
							return "", false, err
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Aborted ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Failed ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Deleting {
							log.Infof("Restore status - %s", restoreResponseStatus)
							log.InfoD("Status of %s - [%s]",
								initialRestoreName, restoreResponseStatus.GetStatus())

							err := fmt.Errorf("Something must have gone wrong. You can try restarting the test")
							log.FailOnError(err, "Unexpected Status")
							return "", false, err
						}

						return "", true, fmt.Errorf("waiting for restore status to be InProgress; got [%s]",
							restoreResponseStatus.GetStatus())
					}
					_, err = task.DoRetryWithTimeout(restoreInProgressCheck, 10*time.Minute, 2*time.Second)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Restore %s is InProgress", initialRestoreName))
				})

				Step("Restoring the backed up application to namespace with different name on destination cluster", func() {
					log.InfoD("Restoring the backed up application to namespace with different name on destination cluster")

					laterRestoreName := fmt.Sprintf("%s-%s-later-%v", restoreNamePrefix, sourceNamespace, time.Now().Unix())
					destinationNameSpace := fmt.Sprintf("%s-%s", sourceNamespace, "later")
					destinationNamespaces = append(destinationNamespaces, destinationNameSpace)
					namespaceMapping := make(map[string]string) //using local version in order to not change mapping as the key is the same
					namespaceMapping[sourceNamespace] = destinationNameSpace
					err = CreateRestoreWithoutCheck(laterRestoreName, backupNames[i], namespaceMapping, destinationClusterName, orgID, ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Initiation of restore %s", laterRestoreName))

					restoreInspectRequest := &api.RestoreInspectRequest{
						Name:  laterRestoreName,
						OrgId: orgID,
					}
					restorePartialSuccessCheck := func() (interface{}, bool, error) {
						resp, err := Inst().Backup.InspectRestore(ctx, restoreInspectRequest)
						restoreResponseStatus := resp.GetRestore().GetStatus()
						log.FailOnError(err, "Failed getting restore status for - %s", laterRestoreName)

						if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_PartialSuccess {
							log.Infof("Restore status - %s", restoreResponseStatus)
							log.InfoD("Status of %s - [%s]",
								laterRestoreName, restoreResponseStatus.GetStatus())
							log.InfoD("Condition fulfilled. Proceeding to confirm Restore 1")
							return "", false, nil
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Success {
							log.Infof("Restore status - %s", restoreResponseStatus)
							log.InfoD("Status of %s - [%s]",
								laterRestoreName, restoreResponseStatus.GetStatus())

							err := fmt.Errorf("This is not supposed to happen. This indicates that Restore 2 didn't face conflicts with 'Cluster-level resources' from Restore 1")
							log.FailOnError(err, "Unexpected Status")
							return "", false, err
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Aborted ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Failed ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Deleting {
							log.Infof("Restore status - %s", restoreResponseStatus)
							log.InfoD("Status of %s - [%s]",
								laterRestoreName, restoreResponseStatus.GetStatus())

							err := fmt.Errorf("Something must have gone wrong. You can try restarting the test")
							log.FailOnError(err, "Unexpected Status")
							return "", false, err
						}

						return "", true, fmt.Errorf("waiting for restore status to be PartialSuccess; got [%s]",
							restoreResponseStatus.GetStatus())
					}
					_, err = task.DoRetryWithTimeout(restorePartialSuccessCheck, 10*time.Minute, 30*time.Second)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Restore 2 (%s) is PartialSuccess", laterRestoreName))

					// Validation of Restore
					restoreInspectRequest = &api.RestoreInspectRequest{
						Name:  laterRestoreName,
						OrgId: orgID,
					}

					backupDriver := Inst().Backup
					resp, err := backupDriver.InspectRestore(ctx, restoreInspectRequest)
					log.FailOnError(err, "Failed to inspect restore: %s", laterRestoreName)

					log.InfoD("Switching to destination context")
					err = SetDestinationKubeConfig()
					log.FailOnError(err, "Failed to switch to context to destination cluster")

					rstCtx, err := ValidateRestore(resp, backupContexts[i], namespaceMapping)

					log.InfoD("Switching to default context")
					err1 := SetClusterContext("")
					log.FailOnError(err1, "Failed to SetClusterContext to default cluster")

					dash.VerifyFatal(err, nil, fmt.Sprintf("Restore 2 (%s) Validation", laterRestoreName))

					restoreLaterContexts = append(restoreLaterContexts, rstCtx)
				})

				Step("Verifying status of Initial Restore", func() {
					log.InfoD("Step: Verifying status of Initial Restore")

					restoreInspectRequest := &api.RestoreInspectRequest{
						Name:  initialRestoreName,
						OrgId: orgID,
					}
					restoreSuccessCheck := func() (interface{}, bool, error) {
						resp, err := Inst().Backup.InspectRestore(ctx, restoreInspectRequest)
						restoreResponseStatus := resp.GetRestore().GetStatus()
						log.FailOnError(err, "Failed getting restore status for - %s", initialRestoreName)

						if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Success {
							log.Infof("Restore status - %s", restoreResponseStatus)
							log.InfoD("Status of %s - [%s]",
								initialRestoreName, restoreResponseStatus.GetStatus())
							log.InfoD("Condition fulfilled. Proceeding to end test")
							return "", false, nil
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_PartialSuccess {
							log.Infof("Restore status - %s", restoreResponseStatus)
							log.InfoD("Status of %s - [%s]",
								initialRestoreName, restoreResponseStatus.GetStatus())

							err := fmt.Errorf("Status of Restore 1 (%s) is PartialSuccess. This should not happen. This MAY have happened if the destination cluster hadn't been cleaned completely (remanant cluster-level resources). The OTHER situation which may have caused this is explained in the comment of JIRA ticket PA-614", initialRestoreName)
							log.FailOnError(err, "Unexpected Status")
							return "", false, err
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Aborted ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Failed ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Deleting {
							log.Infof("Restore status - %s", restoreResponseStatus)
							log.InfoD("Status of %s - [%s]",
								initialRestoreName, restoreResponseStatus.GetStatus())

							err := fmt.Errorf("Something must have gone wrong. You can try restarting the test")
							log.FailOnError(err, "Unexpected Status")
							return "", false, err
						}

						return "", true, fmt.Errorf("waiting for restore status to be Success; got [%s]",
							restoreResponseStatus.GetStatus())
					}
					_, err = task.DoRetryWithTimeout(restoreSuccessCheck, 10*time.Minute, 30*time.Second)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Status of Initial Restore (%s) is Success", initialRestoreName))

					// Validation of Restore

					backupDriver := Inst().Backup
					resp, err := backupDriver.InspectRestore(ctx, restoreInspectRequest)
					log.FailOnError(err, "Failed to inspect restore: %s", initialRestoreName)

					log.InfoD("Switching to destination context")
					err = SetDestinationKubeConfig()
					log.FailOnError(err, "Failed to switch to context to destination cluster")

					rstCtx, err := ValidateRestore(resp, backupContexts[i], namespaceMapping)

					log.InfoD("Switching to default context")
					err1 := SetClusterContext("")
					log.FailOnError(err1, "Failed to SetClusterContext to default cluster")

					dash.VerifyFatal(err, nil, fmt.Sprintf("Restore 1 (%s) Validation", initialRestoreName))

					restoreContexts = append(restoreContexts, rstCtx)
				})

			}
		})

	})

	JustAfterEach(func() {

		defer EndTorpedoTest()

		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")

		// TODO: move this to AfterSuite
		if len(preRuleNameList) > 0 {
			for _, ruleName := range preRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(orgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup pre rules %s", ruleName))
			}
		}

		// TODO: move this to AfterSuite
		if len(postRuleNameList) > 0 {
			for _, ruleName := range postRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(orgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup post rules %s", ruleName))
			}
		}

		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = false
		log.InfoD("Deleting deployed applications for source and destintion clusters")

		log.InfoD("Switching to source context")
		err = SetSourceKubeConfig()
		log.FailOnError(err, "Failed to switch to context to source cluster")

		log.Info("Deleting deployed applications on source clusters")
		ValidateAndDestroy(sourceClusterAppsContexts, opts)

		log.Warn("Waiting for 1 minute, hoping for any Resources created by Operator of Custom Resources to finish being destroyed.\nThis logic will have to be changed to polling in the future, as just randomly waiting, without verification can easily lead to errors. Make sure that any errors you're seeing is not due to this.")
		time.Sleep(time.Minute * 1)

		log.InfoD("Switching to destination context")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "Failed to switch to context to destination cluster")

		destinationClusterAppsContexts = make([]*scheduler.Context, 0)
		// only adding restoreContexts, not restoreLaterContexts
		for _, rstCtx := range restoreContexts {
			destinationClusterAppsContexts = append(destinationClusterAppsContexts, rstCtx.schedCtxs...)
		}
		log.Info("Deleting deployed applications (Restore 1) on destintion clusters")
		ValidateAndDestroy(destinationClusterAppsContexts, opts)

		//TODO: this can be added in the future
		//log.Warn("Not waiting here since we're destroying two namespaces, so we can wait 'at once'")
		//log.Info("Deleting deployed applications (Restore 2) on destintion clusters")
		//here!

		log.Warn("Waiting for 1 minute, hoping for any Resources created by Operator of Custom Resources to finish being destroyed.\nThis logic will have to be changed to polling in the future, as just randomly waiting, without verification can easily lead to errors. Make sure that any errors you're seeing is not due to this.")
		time.Sleep(time.Minute * 1)

		log.InfoD("Switching to default context")
		err = SetClusterContext("")
		log.FailOnError(err, "Failed to SetClusterContext to default cluster")

		// TODO: move this to AfterSuite
		DeleteCluster(SourceClusterName, orgID, ctx)
		DeleteCluster(destinationClusterName, orgID, ctx)

		// Backups and Restores are deleted by AfterSuite

	})
})

// This testcase verifies basic backup rule,backup location, cloud setting
var _ = Describe("{BasicBackupCreation}", func() {
	var (
		appList           = Inst().AppList
		backupName        string
		contexts          []*scheduler.Context
		preRuleNameList   []string
		postRuleNameList  []string
		appContexts       []*scheduler.Context
		bkpNamespaces     []string
		clusterUid        string
		clusterStatus     api.ClusterInfo_StatusInfo_Status
		restoreName       string
		cloudCredName     string
		cloudCredUID      string
		backupLocationUID string
		bkpLocationName   string
		backupLocationMap map[string]string
		labelSelectors    map[string]string
		namespaceMapping  map[string]string
		providers         []string
		intervalName      string
		dailyName         string
		weeklyName        string
		monthlyName       string
		backupNames       []string
		restoreNames      []string
	)

	JustBeforeEach(func() {
		backupLocationMap = make(map[string]string)
		labelSelectors = make(map[string]string)
		bkpNamespaces = make([]string, 0)
		namespaceMapping = make(map[string]string)
		providers = getProviders()
		intervalName = fmt.Sprintf("%s-%v", "interval", time.Now().Unix())
		dailyName = fmt.Sprintf("%s-%v", "daily", time.Now().Unix())
		weeklyName = fmt.Sprintf("%s-%v", "weekly", time.Now().Unix())
		monthlyName = fmt.Sprintf("%s-%v", "monthly", time.Now().Unix())
		StartTorpedoTest("Backup: BasicBackupCreation", "Deploying backup", nil, 0)
		log.InfoD("Verifying if the pre/post rules for the required apps are present in the AppParameters or not")
		for i := 0; i < len(appList); i++ {
			if Contains(preRuleApp, appList[i]) {
				if _, ok := portworx.AppParameters[appList[i]]["pre"]; ok {
					dash.VerifyFatal(ok, true, fmt.Sprintf("Pre Rule details mentioned for the app [%s]", appList[i]))
				}
			}
			if Contains(postRuleApp, appList[i]) {
				if _, ok := portworx.AppParameters[appList[i]]["post"]; ok {
					dash.VerifyFatal(ok, true, fmt.Sprintf("Post Rule details mentioned for the app [%s]", appList[i]))
				}
			}
		}
		log.InfoD("Deploy applications")
		contexts = make([]*scheduler.Context, 0)
		for i := 0; i < Inst().GlobalScaleFactor; i++ {
			taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
			appContexts = ScheduleApplications(taskName)
			contexts = append(contexts, appContexts...)
			for _, ctx := range appContexts {
				ctx.ReadinessTimeout = appReadinessTimeout
				namespace := GetAppNamespace(ctx, taskName)
				bkpNamespaces = append(bkpNamespaces, namespace)
			}
		}
	})
	It("Basic Backup Creation", func() {
		Step("Validating applications", func() {
			log.InfoD("Validating applications")
			ValidateApplications(contexts)
		})
		Step("Creating rules for backup", func() {
			log.InfoD("Creating rules for backup")
			log.InfoD("Creating pre rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				preRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], orgID, "pre")
				log.FailOnError(err, "Creating pre rule for deployed app [%s] failed", appList[i])
				dash.VerifyFatal(preRuleStatus, true, "Verifying pre rule for backup")
				if ruleName != "" {
					preRuleNameList = append(preRuleNameList, ruleName)
				}
			}
			log.InfoD("Creating post rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				postRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], orgID, "post")
				log.FailOnError(err, "Creating post rule for deployed app [%s] failed", appList[i])
				dash.VerifyFatal(postRuleStatus, true, "Verifying Post rule for backup")
				if ruleName != "" {
					postRuleNameList = append(postRuleNameList, ruleName)
				}
			}
		})
		Step("Creating backup location and cloud setting", func() {
			log.InfoD("Creating backup location and cloud setting")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				bkpLocationName = fmt.Sprintf("%s-%s-bl", provider, getGlobalBucketName(provider))
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = bkpLocationName
				CreateCloudCredential(provider, cloudCredName, cloudCredUID, orgID)
				err := CreateBackupLocation(provider, bkpLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), orgID, "")
				dash.VerifyFatal(err, nil, "Creating backup location")
			}
		})
		Step("Creating backup schedule policies", func() {
			log.InfoD("Creating backup schedule policies")
			log.InfoD("Creating backup interval schedule policy")
			intervalSchedulePolicyInfo := Inst().Backup.CreateIntervalSchedulePolicy(5, 15, 2)
			intervalPolicyStatus := Inst().Backup.BackupSchedulePolicy(intervalName, uuid.New(), orgID, intervalSchedulePolicyInfo)
			dash.VerifyFatal(intervalPolicyStatus, nil, "Creating interval schedule policy")

			log.InfoD("Creating backup daily schedule policy")
			dailySchedulePolicyInfo := Inst().Backup.CreateDailySchedulePolicy(1, "9:00AM", 2)
			dailyPolicyStatus := Inst().Backup.BackupSchedulePolicy(dailyName, uuid.New(), orgID, dailySchedulePolicyInfo)
			dash.VerifyFatal(dailyPolicyStatus, nil, "Creating daily schedule policy")

			log.InfoD("Creating backup weekly schedule policy")
			weeklySchedulePolicyInfo := Inst().Backup.CreateWeeklySchedulePolicy(1, backup.Friday, "9:10AM", 2)
			weeklyPolicyStatus := Inst().Backup.BackupSchedulePolicy(weeklyName, uuid.New(), orgID, weeklySchedulePolicyInfo)
			dash.VerifyFatal(weeklyPolicyStatus, nil, "Creating weekly schedule policy")

			log.InfoD("Creating backup monthly schedule policy")
			monthlySchedulePolicyInfo := Inst().Backup.CreateMonthlySchedulePolicy(1, 29, "9:20AM", 2)
			monthlyPolicyStatus := Inst().Backup.BackupSchedulePolicy(monthlyName, uuid.New(), orgID, monthlySchedulePolicyInfo)
			dash.VerifyFatal(monthlyPolicyStatus, nil, "Creating monthly schedule policy")
		})
		Step("Registering cluster for backup", func() {
			log.InfoD("Registering cluster for backup")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			err = CreateSourceAndDestClusters(orgID, "", "", ctx)
			dash.VerifyFatal(err, nil, "Creating source and destination cluster")
			clusterStatus, clusterUid = Inst().Backup.RegisterBackupCluster(orgID, SourceClusterName, "")
			dash.VerifyFatal(clusterStatus, api.ClusterInfo_StatusInfo_Online, fmt.Sprintf("Verifying backup cluster with uid [%s]", clusterUid))
		})
		Step("Taking backup of all namespaces", func() {
			log.InfoD("Taking backup of all namespaces")
			ctx, err := backup.GetAdminCtxFromSecret()
			dash.VerifyFatal(err, nil, "Getting context")
			for _, namespace := range bkpNamespaces {
				backupName = fmt.Sprintf("%s-%s-%s", BackupNamePrefix, namespace, RandomString(4))
				for strings.Contains(strings.Join(backupNames, ","), backupName) {
					backupName = fmt.Sprintf("%s-%s-%s", BackupNamePrefix, namespace, RandomString(4))
				}
				backupNames = append(backupNames, backupName)
				err = CreateBackup(backupName, SourceClusterName, bkpLocationName, backupLocationUID, []string{namespace},
					labelSelectors, orgID, clusterUid, "", "", "", "", ctx)
				dash.VerifyFatal(err, nil, fmt.Sprintf("Verifying [%s] backup creation", backupName))
			}
		})
		Step("Restoring the backed up namespaces", func() {
			log.InfoD("Restoring the backed up namespaces")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")
			for index, namespace := range bkpNamespaces {
				restoreName = fmt.Sprintf("%s-%s-%s", "test-restore", namespace, RandomString(4))
				for strings.Contains(strings.Join(restoreNames, ","), restoreName) {
					restoreName = fmt.Sprintf("%s-%s-%s", "test-restore", namespace, RandomString(4))
				}
				restoreNames = append(restoreNames, restoreName)
				log.InfoD("Restoring [%s] namespace from the [%s] backup", namespace, backupNames[index])
				err = CreateRestore(restoreName, backupNames[index], namespaceMapping, destinationClusterName, orgID, ctx, make(map[string]string))
				dash.VerifyFatal(err, nil, fmt.Sprintf("Creating restore [%s]", restoreName))
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		policyList := []string{intervalName, dailyName, weeklyName, monthlyName}
		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "Fetching px-central-admin ctx")
		if len(preRuleNameList) > 0 {
			for _, ruleName := range preRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(orgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup pre rules [%s]", ruleName))
			}
		}
		if len(postRuleNameList) > 0 {
			for _, ruleName := range postRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(orgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("Deleting backup post rules [%s]", ruleName))
			}
		}
		err = Inst().Backup.DeleteBackupSchedulePolicy(orgID, policyList)
		dash.VerifySafely(err, nil, "Deleting backup schedule policies")
		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = true
		log.Info("Deleting deployed namespaces")
		ValidateAndDestroy(contexts, opts)
		backupDriver := Inst().Backup
		log.Info("Deleting backed up namespaces")
		for _, backupName := range backupNames {
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, orgID)
			log.FailOnError(err, "Failed while trying to get backup UID for - %s", backupName)
			backupDeleteResponse, err := DeleteBackup(backupName, backupUID, orgID, ctx)
			log.FailOnError(err, "Backup [%s] could not be deleted", backupName)
			dash.VerifyFatal(backupDeleteResponse.String(), "", fmt.Sprintf("Verifying [%s] backup deletion is successful", backupName))
		}
		log.Info("Deleting restored namespaces")
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, orgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore [%s]", restoreName))
		}
		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
	})
})
