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

	semver "github.com/blang/semver"
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

var _ = Describe("{BackupRestoreCRsOnDifferentK8sVersions}", func() {

	var (
		backupNames                    []string                // backups in px-backup
		restoreNames                   []string                // restores in px-backup
		restoreLaterNames              []string                // restore-laters in px-backup
		sourceClusterAppsContexts      []*scheduler.Context    // Each Context is for one Namespace which corresponds to one App
		destinationClusterAppsContexts []*scheduler.Context    // Each Context is for one Namespace which corresponds to one App
		backupContexts                 []*BackupRestoreContext // Each Context is for one backup in px-backup
		restoreContexts                []*BackupRestoreContext // Each Context is for one restore in px-backup
		restoreLaterContexts           []*BackupRestoreContext // Each Context is for one restore-later in px-backup
		preRuleNameList                []string
		postRuleNameList               []string
		clusterUid                     string
		cloudCredName                  string
		cloudCredUID                   string
		backupLocationUID              string
		backupLocationName             string
	)

	var (
		appList               = Inst().AppList
		sourceNamespaces      = make([]string, 0)
		destinationNamespaces = make([]string, 0)
		namespaceMapping      = make(map[string]string)
		backupLocationMap     = make(map[string]string)
		labelSelectors        = make(map[string]string)
	)

	providers := getProviders()

	JustBeforeEach(func() {

		StartTorpedoTest("BackupRestoreCRsOnDifferentK8sVersions", "Deploy CRs (CRD + webhook); Backup; two simulatanous Restores with one Success and other PartialSuccess. (Backup and Restore on different K8s version)", nil, 83716)

		log.InfoD("verifying if the pre/post rules for the required apps are present in the AppParameters or not")
		for i := 0; i < len(appList); i++ {
			if Contains(postRuleApp, appList[i]) {
				if _, ok := portworx.AppParameters[appList[i]]["post"]; ok {
					dash.VerifyFatal(ok, true, "post rule details mentioned for the apps")
				}
			}
			if Contains(preRuleApp, appList[i]) {
				if _, ok := portworx.AppParameters[appList[i]]["pre"]; ok {
					dash.VerifyFatal(ok, true, "pre rule details mentioned for the apps")
				}
			}
		}

	})

	It("Deploy CRs (CRD + webhook); Backup; two simulatanous Restores with one Success and other PartialSuccess. (Backup and Restore on different K8s version)", func() {

		defer func() {
			log.InfoD("switching to default context")
			err1 := SetClusterContext("")
			log.FailOnError(err1, "failed to SetClusterContext to default cluster")
		}()

		Step("Verify if app used to execute test is a valid/allowed spec (apps) for *this* test", func() {
			log.InfoD("specs (apps) allowed in execution of test: %v", appsWithCRDsAndWebhooks)
			for i := 0; i < len(appList); i++ {
				contains := Contains(appsWithCRDsAndWebhooks, appList[i])
				dash.VerifyFatal(contains, true,
					fmt.Sprintf("app [%s] allowed in execution of this test", appList[i]))
			}
		})

		Step("verify kubernetes version of source and destination cluster", func() {
			var srcVer, destVer semver.Version
			log.InfoD("begin verification kubernetes version of source and destination cluster")

			Step("register cluster for backup", func() {
				log.InfoD("register cluster for backup")
				ctx, err := backup.GetAdminCtxFromSecret()
				log.FailOnError(err, "fetching px-central-admin ctx")
				err = CreateSourceAndDestClusters(orgID, "", "", ctx)
				dash.VerifyFatal(err, nil, "creating source and destination cluster")
				_, clusterUid = Inst().Backup.RegisterBackupCluster(orgID, SourceClusterName, "")
			})

			Step("Get kubernetes source cluster version", func() {
				log.InfoD("switched context to source")

				sourceClusterConfigPath, err := GetSourceClusterConfigPath()
				log.FailOnError(err, "failed to get kubeconfig path for source cluster. Error: [%v]", err)

				err = Inst().S.SetConfig(sourceClusterConfigPath)
				log.FailOnError(err, "failed to switch to context to source cluster [%v]", sourceClusterConfigPath)

				ver, err := k8s.ClusterVersion()
				log.FailOnError(err, "failed to get source cluster version")
				srcVer, err = semver.Make(ver)
				log.FailOnError(err, "failed to get source cluster version")
			})

			Step("Get kubernetes destination cluster version", func() {
				log.InfoD("switched context to destination")

				destinationClusterConfigPath, err := GetDestinationClusterConfigPath()
				log.FailOnError(err, "failed to get kubeconfig path for destination cluster. Error: [%v]", err)

				err = Inst().S.SetConfig(destinationClusterConfigPath)
				log.FailOnError(err, "failed to switch to context to destination cluster [%v]", destinationClusterConfigPath)

				ver, err := k8s.ClusterVersion()
				log.FailOnError(err, "failed to get destination cluster version")
				destVer, err = semver.Make(ver)
				log.FailOnError(err, "failed to get destination cluster version")
			})

			Step("Compare Source and Destination cluster version numbers", func() {
				log.InfoD("source cluster version: %s ; destination cluster version: %s", srcVer.String(), destVer.String())
				isValid := srcVer.LT(destVer)
				dash.VerifyFatal(isValid, true,
					"source cluster kubernetes version must be lesser than the destination cluster kubernetes version.")
			})

			log.InfoD("switching to default context")
			err := SetClusterContext("")
			log.FailOnError(err, "failed to SetClusterContext to default cluster")
		})

		Step("deploy the applications on Src cluster", func() {
			log.InfoD("deploy the applications on Src cluster")

			Step("deploy applications", func() {
				log.InfoD("deploy applications")

				log.InfoD("switching to source context")
				err := SetSourceKubeConfig()
				log.FailOnError(err, "failed to switch to context to source cluster")

				log.InfoD("ccheduling applications")
				sourceClusterAppsContexts = make([]*scheduler.Context, 0)
				for i := 0; i < Inst().GlobalScaleFactor; i++ {
					taskName := fmt.Sprintf("%s-%d", taskNamePrefix, i)
					appContexts := ScheduleApplications(taskName)
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

				log.InfoD("switching to default context")
				err := SetClusterContext("")
				log.FailOnError(err, "failed to SetClusterContext to default cluster")
			})

			log.InfoD("waiting (for 2 minutes) for any CRs to finish starting up.")
			time.Sleep(time.Minute * 2)
			log.Warnf("no verification is done; it might lead to undetectable errors.")
		})

		Step("Creating rules for backup", func() {
			log.InfoD("creating pre rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				preRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], orgID, "pre")
				log.FailOnError(err, "creating pre rule for deployed apps failed")
				dash.VerifyFatal(preRuleStatus, true, "verifying pre rule for backup")

				if ruleName != "" {
					preRuleNameList = append(preRuleNameList, ruleName)
				}
			}
			log.InfoD("Creating post rule for deployed apps")
			for i := 0; i < len(appList); i++ {
				postRuleStatus, ruleName, err := Inst().Backup.CreateRuleForBackup(appList[i], orgID, "post")
				log.FailOnError(err, "creating post rule for deployed apps failed")
				dash.VerifyFatal(postRuleStatus, true, "verifying Post rule for backup")
				if ruleName != "" {
					postRuleNameList = append(postRuleNameList, ruleName)
				}
			}
		})

		Step("Creating bucket, backup location and cloud credentials", func() {
			log.InfoD("Creating bucket, backup location and cloud credentials")
			for _, provider := range providers {
				cloudCredName = fmt.Sprintf("%s-%s-%v", "cred", provider, time.Now().Unix())
				backupLocationName = fmt.Sprintf("%s-%s-bl", provider, getGlobalBucketName(provider))
				cloudCredUID = uuid.New()
				backupLocationUID = uuid.New()
				backupLocationMap[backupLocationUID] = backupLocationName
				CreateCloudCredential(provider, cloudCredName, cloudCredUID, orgID)
				log.InfoD("creating backup location [%s] with cloud cred [%s]", backupLocationName, cloudCredName)
				err := CreateBackupLocation(provider, backupLocationName, backupLocationUID, cloudCredName, cloudCredUID, getGlobalBucketName(provider), orgID, "")
				dash.VerifyFatal(err, nil, "creating backup location")
			}
		})

		Step("Taking backup of application from source cluster", func() {
			log.InfoD("taking backup of applications")
			ctx, err := backup.GetAdminCtxFromSecret()
			dash.VerifyFatal(err, nil, "getting context")
			backupNames = make([]string, len(sourceNamespaces))
			backupContexts = make([]*BackupRestoreContext, len(sourceNamespaces))
			for i, namespace := range sourceNamespaces {
				backupName := fmt.Sprintf("%s-%s-%v", BackupNamePrefix, namespace, time.Now().Unix())
				log.InfoD("creating backup [%s] in source cluster [%s] (%s), organization [%s], of namespace [%s], in backup location [%s]", backupName, SourceClusterName, clusterUid, orgID, namespace, backupLocationName)
				backupCtx, err := CreateBackupAndGetBackupCtx(backupName, SourceClusterName, backupLocationName, backupLocationUID, []string{namespace}, labelSelectors, orgID, clusterUid, "", "", "", "", ctx, []*scheduler.Context{sourceClusterAppsContexts[i]})

				dash.VerifyFatal(err, nil, "verifying backup creation")
				backupNames[i] = backupName
				backupContexts[i] = backupCtx
			}
		})

		Step("Restoring the backed up applications on destination cluster", func() {

			log.InfoD("Restoring the backed up applications on destination cluster")
			ctx, err := backup.GetAdminCtxFromSecret()
			log.FailOnError(err, "Fetching px-central-admin ctx")

			for i, sourceNamespace := range sourceNamespaces {
				var initialRestoreName, laterRestoreName string

				Step("Restoring the backed up application to namespace of same name on destination cluster", func() {
					log.InfoD("restoring the backed up application to namespace of same name on destination cluster")

					initialRestoreName = fmt.Sprintf("%s-%s-initial-%v", restoreNamePrefix, sourceNamespace, time.Now().Unix())
					restoreNames = append(restoreNames, initialRestoreName)
					destinationNameSpace := sourceNamespace
					destinationNamespaces = append(destinationNamespaces, destinationNameSpace)
					namespaceMapping[sourceNamespace] = destinationNameSpace

					log.InfoD("creating initial-restore [%s] in destination cluster [%s], organization [%s], in namespace [%s]", initialRestoreName, destinationClusterName, orgID, destinationNameSpace)
					err = CreateRestoreWithoutCheck(initialRestoreName, backupNames[i], namespaceMapping, destinationClusterName, orgID, ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Initiation of restore %s", initialRestoreName))

					restoreInspectRequest := &api.RestoreInspectRequest{
						Name:  initialRestoreName,
						OrgId: orgID,
					}
					restoreInProgressCheck := func() (interface{}, bool, error) {
						resp, err := Inst().Backup.InspectRestore(ctx, restoreInspectRequest)
						restoreResponseStatus := resp.GetRestore().GetStatus()
						if err != nil {
							err := fmt.Errorf("failed getting restore status for - %s; Err: %s", initialRestoreName, err)
							return "", false, err
						}

						// Status should be LATER than InProgress in order for next STEP to execute
						if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_InProgress {
							log.InfoD("sestore status of [%s] is [%s]; expected [InProgress].\ncondition fulfilled.", initialRestoreName, restoreResponseStatus.GetStatus())
							return "", false, nil
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_PartialSuccess {
							err := fmt.Errorf("restore status of [%s] is [%s]; expected [InProgress].\nhelp: check for remnant cluster-level resources on destination cluster.", initialRestoreName, restoreResponseStatus.GetStatus())
							return "", false, err
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Success {
							err := fmt.Errorf("restore status of [%s] is [%s]; expected [InProgress].\nhelp: check for status frequently", initialRestoreName, restoreResponseStatus.GetStatus())
							return "", false, err
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Aborted ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Failed ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Deleting {
							err := fmt.Errorf("restore status of [%s] is [%s]; expected [InProgress].", initialRestoreName, restoreResponseStatus.GetStatus())
							return "", false, err
						}

						err = fmt.Errorf("restore status of [%s] is [%s]; waiting for [InProgress]...", initialRestoreName, restoreResponseStatus.GetStatus())
						return "", true, err
					}
					_, err = task.DoRetryWithTimeout(restoreInProgressCheck, 10*time.Minute, 5*time.Second)
					dash.VerifyFatal(err, nil, fmt.Sprintf("restore %s is [InProgress]", initialRestoreName))
				})

				var restoreLaterStatuserr error

				Step("Restoring the backed up application to namespace with different name on destination cluster", func() {
					log.InfoD("Restoring the backed up application to namespace with different name on destination cluster")

					laterRestoreName = fmt.Sprintf("%s-%s-later-%v", restoreNamePrefix, sourceNamespace, time.Now().Unix())
					restoreLaterNames = append(restoreLaterNames, laterRestoreName)
					destinationNameSpace := fmt.Sprintf("%s-%s", sourceNamespace, "later")
					destinationNamespaces = append(destinationNamespaces, destinationNameSpace)
					namespaceMapping := make(map[string]string) //using local version in order to not change mapping as the key is the same
					namespaceMapping[sourceNamespace] = destinationNameSpace

					log.InfoD("creating later-restore [%s] in destination cluster [%s], organization [%s], in namespace [%s]", laterRestoreName, destinationClusterName, orgID, destinationNameSpace)
					err = CreateRestoreWithoutCheck(laterRestoreName, backupNames[i], namespaceMapping, destinationClusterName, orgID, ctx)
					dash.VerifyFatal(err, nil, fmt.Sprintf("Initiation of restore %s", laterRestoreName))

					restoreInspectRequest := &api.RestoreInspectRequest{
						Name:  laterRestoreName,
						OrgId: orgID,
					}
					restorePartialSuccessCheck := func() (interface{}, bool, error) {
						resp, err := Inst().Backup.InspectRestore(ctx, restoreInspectRequest)
						restoreResponseStatus := resp.GetRestore().GetStatus()
						if err != nil {
							err := fmt.Errorf("failed getting restore status for - %s; Err: %s", laterRestoreName, err)
							return "", false, err
						}

						if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_PartialSuccess {
							log.InfoD("restore status of [%s] is [%s]; expected [PartialSuccess].\ncondition fulfilled. proceeding to confirm restore [%s].", laterRestoreName, restoreResponseStatus.GetStatus(), initialRestoreName)
							return "", false, nil
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Success {
							err := fmt.Errorf("restore status of [%s] is [%s]; expected [PartialSuccess].\nhelp: 1. app must have cluster-level resources.\n2. issue with restore [%s]: was not [Success]\n3. refer jira ticket PA-614", laterRestoreName, restoreResponseStatus.GetStatus(), initialRestoreName)
							return "", false, err
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Aborted ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Failed ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Deleting {
							err := fmt.Errorf("restore status of [%s] is [%s]; expected [PartialSuccess].", laterRestoreName, restoreResponseStatus.GetStatus())
							return "", false, err
						}

						err = fmt.Errorf("restore status of [%s] is [%s]; waiting for [PartialSuccess]...", initialRestoreName, restoreResponseStatus.GetStatus())
						return "", true, err
					}
					_, restoreLaterStatuserr = task.DoRetryWithTimeout(restorePartialSuccessCheck, 10*time.Minute, 30*time.Second)
					// we don't end the test if there is an error here, as we also want to ensure that we look into the status of the following `step`, so that we have the full details of what went wrong.
					dash.VerifySafely(restoreLaterStatuserr, nil, fmt.Sprintf("restore [%s] is [PartialSuccess]", laterRestoreName))
					if restoreLaterStatuserr != nil {
						log.Warnf("due to error in restore status check, skipping validation of restore. proceeding to next step, after which the test will be failed")
						return
					}

					// Validation of Restore
					destinationClusterConfigPath, err := GetDestinationClusterConfigPath()
					log.FailOnError(err, "failed to get kubeconfig path for destination cluster. Error: [%v]", err)

					restoreLaterCtx, err := ValidateRestore(laterRestoreName, destinationClusterConfigPath, orgID, ctx, backupContexts[i], namespaceMapping)
					dash.VerifyFatal(err, nil, fmt.Sprintf("restore (%s) validation", laterRestoreName))
					restoreLaterContexts = append(restoreLaterContexts, restoreLaterCtx)
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
						if err != nil {
							err := fmt.Errorf("failed getting restore status for - %s; Err: %s", initialRestoreName, err)
							return "", false, err
						}

						if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Success {
							log.InfoD("restore status of [%s] is [%s]; expected [InProgress].\ncondition fulfilled.", initialRestoreName, restoreResponseStatus.GetStatus())
							return "", false, nil
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_PartialSuccess {
							err := fmt.Errorf("restore status of [%s] is [%s]; expected [Success].\nhelp: 1. check for remnant cluster-level resources on destination cluster.\n2. refer jira ticket PA-614", initialRestoreName, restoreResponseStatus.GetStatus())
							return "", false, err
						} else if restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Aborted ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Failed ||
							restoreResponseStatus.GetStatus() == api.RestoreInfo_StatusInfo_Deleting {
							err := fmt.Errorf("restore status of [%s] is [%s]; expected [InProgress].", initialRestoreName, restoreResponseStatus.GetStatus())
							return "", false, err
						}

						err = fmt.Errorf("sestore status of [%s] is [%s]; waiting for [Success]...", initialRestoreName, restoreResponseStatus.GetStatus())
						return "", true, err
					}
					_, err = task.DoRetryWithTimeout(restoreSuccessCheck, 10*time.Minute, 30*time.Second)
					dash.VerifyFatal(err, nil, fmt.Sprintf("status of initial restore [%s] is success", initialRestoreName))

					// Validation of Restore
					destinationClusterConfigPath, err := GetDestinationClusterConfigPath()
					log.FailOnError(err, "failed to get kubeconfig path for destination cluster. Error: [%v]", err)

					restoreCtx, err := ValidateRestore(initialRestoreName, destinationClusterConfigPath, orgID, ctx, backupContexts[i], namespaceMapping)
					dash.VerifyFatal(err, nil, fmt.Sprintf("restore (%s) validation", initialRestoreName))
					restoreContexts = append(restoreContexts, restoreCtx)

					// If this was an error before, we have to fail the test at this point, having processed the other stage
					dash.VerifyFatal(restoreLaterStatuserr, nil, fmt.Sprintf("restore [%s] is [PartialSuccess]", laterRestoreName))
				})

			}
		})

	})

	JustAfterEach(func() {

		defer EndTorpedoTest()

		ctx, err := backup.GetAdminCtxFromSecret()
		log.FailOnError(err, "fetching px-central-admin ctx")

		// TODO: move this to AfterSuite
		if len(preRuleNameList) > 0 {
			for _, ruleName := range preRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(orgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("deleting backup pre rules %s", ruleName))
			}
		}

		// TODO: move this to AfterSuite
		if len(postRuleNameList) > 0 {
			for _, ruleName := range postRuleNameList {
				err := Inst().Backup.DeleteRuleForBackup(orgID, ruleName)
				dash.VerifySafely(err, nil, fmt.Sprintf("deleting backup post rules %s", ruleName))
			}
		}

		opts := make(map[string]bool)
		opts[SkipClusterScopedObjects] = false
		log.InfoD("deleting deployed applications for source and destination clusters")

		log.InfoD("switching to source context")
		err = SetSourceKubeConfig()
		log.FailOnError(err, "failed to switch to context to source cluster")

		log.InfoD("deleting deployed applications on source clusters")
		ValidateAndDestroy(sourceClusterAppsContexts, opts)

		log.InfoD("waiting (for 1 minute) for any Resources created by Operator of Custom Resources to finish being destroyed.")
		time.Sleep(time.Minute * 1)
		log.Warn("no verification of destruction is done; it might lead to undetectable errors.")

		log.InfoD("switching to destination context")
		err = SetDestinationKubeConfig()
		log.FailOnError(err, "failed to switch to context to destination cluster")

		destinationClusterAppsContexts = make([]*scheduler.Context, 0)
		// only adding restoreContexts, not restoreLaterContexts
		for _, restoreCtx := range restoreContexts {
			destinationClusterAppsContexts = append(destinationClusterAppsContexts, restoreCtx.schedulerCtxs...)
		}
		log.InfoD("deleting deployed applications (initial restore) on destination clusters")
		ValidateAndDestroy(destinationClusterAppsContexts, opts)

		//TODO: delete restore-later apps
		log.Warn("not deleting deployed applications (restore-later) on destination clusters")

		log.InfoD("waiting (for 1 minute) for any Resources created by Operator of Custom Resources to finish being destroyed.")
		time.Sleep(time.Minute * 1)
		log.Warn("no verification of destruction is done; it might lead to undetectable errors.")

		log.InfoD("switching to default context")
		err = SetClusterContext("")
		log.FailOnError(err, "failed to SetClusterContext to default cluster")

		backupDriver := Inst().Backup

		log.InfoD("deleting backed up namespaces")
		for _, backupName := range backupNames {
			backupUID, err := backupDriver.GetBackupUID(ctx, backupName, orgID)
			log.FailOnError(err, "failed while trying to get backup UID for - %s", backupName)
			backupDeleteResponse, err := DeleteBackup(backupName, backupUID, orgID, ctx)
			log.FailOnError(err, "backup [%s] could not be deleted", backupName)
			dash.VerifyFatal(backupDeleteResponse.String(), "", fmt.Sprintf("verifying [%s] backup deletion is successful", backupName))
		}

		log.InfoD("deleting restores")
		for _, restoreName := range restoreNames {
			err = DeleteRestore(restoreName, orgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("Deleting Restore [%s]", restoreName))
		}

		log.InfoD("deleting restore-laters")
		for _, restoreLaterName := range restoreLaterNames {
			err = DeleteRestore(restoreLaterName, orgID, ctx)
			dash.VerifyFatal(err, nil, fmt.Sprintf("deleting Restore [%s]", restoreLaterName))
		}

		CleanupCloudSettingsAndClusters(backupLocationMap, cloudCredName, cloudCredUID, ctx)
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
