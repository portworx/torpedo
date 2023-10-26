package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	pdsdriver "github.com/portworx/torpedo/drivers/pds"
	"github.com/portworx/torpedo/drivers/pds/controlplane"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	pdsbkp "github.com/portworx/torpedo/drivers/pds/pdsbackup"
	restoreBkp "github.com/portworx/torpedo/drivers/pds/pdsrestore"
	tc "github.com/portworx/torpedo/drivers/pds/targetcluster"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	v1 "k8s.io/api/apps/v1"
	"math/rand"
	"strconv"
	"time"
)

var _ = Describe("{ResizeStorageAndPerformRestoreXFSRepl2}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ResizeStorageAndPerformRestoreXFSRepl2", "Perform PVC Resize and validate the updated vol in the storage config also perform restore of the ds", pdsLabels, 0)
		credName := targetName + pdsbkp.RandString(8)
		bkpClient, err = pdsbkp.InitializePdsBackup()
		log.FailOnError(err, "Failed to initialize backup for pds.")
		bkpTarget, err = bkpClient.CreateAwsS3BackupCredsAndTarget(tenantID, fmt.Sprintf("%v-aws", credName), deploymentTargetID)
		log.FailOnError(err, "Failed to create S3 backup target.")
		log.InfoD("AWS S3 target - %v created successfully", bkpTarget.GetName())
		awsBkpTargets = append(awsBkpTargets, bkpTarget)
		//Initializing the parameters required for workload generation
		wkloadParams = pdsdriver.LoadGenParams{
			LoadGenDepName: params.LoadGen.LoadGenDepName,
			Namespace:      params.InfraToTest.Namespace,
			NumOfRows:      params.LoadGen.NumOfRows,
			Timeout:        params.LoadGen.Timeout,
			Replicas:       params.LoadGen.Replicas,
			TableName:      params.LoadGen.TableName,
			Iterations:     params.LoadGen.Iterations,
			FailOnError:    params.LoadGen.FailOnError,
		}
	})

	It("Perform PVC Resize and validate the updated vol in the storage config", func() {

		var (
			updatedDeployment        *pds.ModelsDeployment
			restoredDeployments      []*pds.ModelsDeployment
			wlDeploymentsToBeCleaned []*v1.Deployment
			updatedDepList           []*pds.ModelsDeployment
			depList                  []*pds.ModelsDeployment
			resConfigModelUpdated    *pds.ModelsResourceSettingsTemplate
			stConfigModelUpdated     *pds.ModelsStorageOptionsTemplate
			newResourceTemplateID    string
			newStorageTemplateID     string
			updatedPvcSize           uint64
		)
		restoredDeploymentsmd5Hash := make(map[string]string)
		stepLog := "Create Custom Templates , Deploy ds and Trigger Workload"
		Step(stepLog, func() {
			backupSupportedDataServiceNameIDMap, err = bkpClient.GetAllBackupSupportedDataServices()
			log.FailOnError(err, "Error while fetching the backup supported ds.")
			for _, ds := range params.DataServiceToTest {
				log.InfoD(stepLog)

				CleanMapEntries(restoredDeploymentsmd5Hash)

				_, supported := backupSupportedDataServiceNameIDMap[ds.Name]
				if !supported {
					log.InfoD("Data service: %v doesn't support backup, skipping...", ds.Name)
					continue
				}
				deployment, initialCapacity, resConfigModel, stConfigModel, appConfigID, workloadDep, pdsdeploymentsmd5Hash, err := DeployDSWithCustomTemplatesRunWorkloads(ds, tenantID, controlplane.Templates{
					CpuLimit:       "2",
					CpuRequest:     "1",
					MemoryLimit:    "4G",
					MemoryRequest:  "2G",
					StorageRequest: "1G",
					FsType:         "xfs",
					ReplFactor:     2,
					Provisioner:    "pxd.portworx.com",
					Secure:         false,
					VolGroups:      false,
				})
				depList = append(depList, deployment)
				dataserviceID, _ := dsTest.GetDataServiceID(ds.Name)
				stepLog = "Update the resource/storage template with increased storage size"
				Step(stepLog, func() {
					newTemplateName := "autoTemp-" + strconv.Itoa(rand.Int())
					updatedTemplateConfig := controlplane.Templates{
						CpuLimit:       *resConfigModel.CpuLimit,
						CpuRequest:     *resConfigModel.CpuRequest,
						DataServiceID:  dataserviceID,
						MemoryLimit:    *resConfigModel.MemoryLimit,
						MemoryRequest:  *resConfigModel.MemoryRequest,
						Name:           newTemplateName,
						StorageRequest: "500G",
						FsType:         *stConfigModel.Fs,
						ReplFactor:     *stConfigModel.Repl,
						Provisioner:    *stConfigModel.Provisioner,
						Secure:         false,
						VolGroups:      false,
					}
					stConfigModelUpdated, resConfigModelUpdated, err = controlPlane.CreateCustomResourceTemplate(tenantID, updatedTemplateConfig)
					log.FailOnError(err, "Unable to update template")
					log.InfoD("Successfully updated the template with ID- %v", resConfigModelUpdated.GetId())
					newResourceTemplateID = resConfigModelUpdated.GetId()
					newStorageTemplateID = stConfigModelUpdated.GetId()
				})
				stepLog = "Apply updated template to the dataservice deployment"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					if appConfigID == "" {
						appConfigID, err = controlPlane.GetAppConfTemplate(tenantID, ds.Name)
						log.FailOnError(err, "Error while fetching AppConfigID")
					}
					updatedDeployment, err = dsTest.UpdateDataServices(deployment.GetId(),
						appConfigID, deployment.GetImageId(),
						int32(ds.Replicas), newResourceTemplateID, params.InfraToTest.Namespace)
					log.FailOnError(err, "Error while updating dataservices")
					Step("Validate Deployments after template update", func() {
						err = dsTest.ValidateDataServiceDeployment(updatedDeployment, namespace)
						log.FailOnError(err, "Error while validating dataservices")
						log.InfoD("Data-service: %v is up and healthy", ds.Name)
						updatedDepList = append(updatedDepList, updatedDeployment)
						updatedPvcSize, err = GetVolumeCapacityInGB(namespace, updatedDeployment)
						log.InfoD("Updated Storage Size is- %v", updatedPvcSize)
					})
					stepLog = "Validate Workload is running after storage resize"
					Step(stepLog, func() {
						err = k8sApps.ValidateDeployment(workloadDep, timeOut, 10*time.Second)
						log.FailOnError(err, "Workload is not running after Storage Size Increase")
					})
					stepLog = "Verify storage size before and after storage resize - Verify at STS, PV,PVC level"
					Step(stepLog, func() {

						_, _, config, err := pdslib.ValidateDataServiceVolumes(updatedDeployment, ds.Name, newResourceTemplateID, newStorageTemplateID, params.InfraToTest.Namespace)
						log.FailOnError(err, "error on ValidateDataServiceVolumes method")
						log.InfoD("resConfigModel.StorageRequest val is- %v and updated config val is- %v", *resConfigModelUpdated.StorageRequest, config.Spec.Resources.Requests.Storage)
						dash.VerifyFatal(config.Spec.Resources.Requests.Storage, *resConfigModelUpdated.StorageRequest, "Validating the storage size is updated in the config post resize (STS-LEVEL)")
						dash.VerifyFatal(config.Spec.StorageOptions.Filesystem, *stConfigModel.Fs, "Validating the File System Type post storage resize (FileSystem-LEVEL)")
						stringRelFactor := strconv.Itoa(int(*stConfigModel.Repl))
						dash.VerifyFatal(config.Spec.StorageOptions.Replicas, stringRelFactor, "Validating the Replication Factor count post storage resize (RepelFactor-LEVEL)")
						if updatedPvcSize > initialCapacity {
							flag := true
							dash.VerifyFatal(flag, true, "Validating the storage size is updated in the config post resize (PV/PVC-LEVEL)")
							log.InfoD("Initial PVC Capacity is- %v and Updated PVC Capacity is- %v", initialCapacity, updatedPvcSize)
						} else {
							log.FailOnError(err, "Failed to verify Storage Resize at PV/PVC level")
						}
					})
				})
				stepLog = "Perform backup after PVC Resize"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					log.Infof("Deployment ID: %v, backup target ID: %v", updatedDeployment.GetId(), bkpTarget.GetId())
					err = bkpClient.TriggerAndValidateAdhocBackup(updatedDeployment.GetId(), bkpTarget.GetId(), "s3")
					log.FailOnError(err, "Failed while performing adhoc backup.")
				})
				stepLog = "Perform Restore after PVC Resize"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					ctx, err := GetSourceClusterConfigPath()
					log.FailOnError(err, "failed while getting src cluster path")
					restoreTarget := tc.NewTargetCluster(ctx)
					restoreClient := restoreBkp.RestoreClient{
						TenantId:             tenantID,
						ProjectId:            projectID,
						Components:           components,
						Deployment:           updatedDeployment,
						RestoreTargetCluster: restoreTarget,
					}
					backupJobs, err := restoreClient.Components.BackupJob.ListBackupJobsBelongToDeployment(projectID, updatedDeployment.GetId())
					log.FailOnError(err, "Error while fetching the backup jobs for the deployment: %v", updatedDeployment.GetClusterResourceName())
					for _, backupJob := range backupJobs {
						log.InfoD("[Restoring] Details Backup job name- %v, Id- %v", backupJob.GetName(), backupJob.GetId())
						restoredModel, err := restoreClient.TriggerAndValidateRestore(backupJob.GetId(), params.InfraToTest.Namespace, dsEntity, true, true)
						log.FailOnError(err, "Failed during restore.")
						restoredDeployment, err = restoreClient.Components.DataServiceDeployment.GetDeployment(restoredModel.GetDeploymentId())
						log.FailOnError(err, fmt.Sprintf("Failed while fetching the restore data service instance: %v", restoredModel.GetClusterResourceName()))
						restoredDeployments = append(restoredDeployments, restoredDeployment)
						deploymentsToBeCleaned = append(deploymentsToBeCleaned, restoredDeployment)
						log.InfoD("Restored successfully. Deployment- %v", restoredModel.GetClusterResourceName())
					}
				})

				stepLog = "Validate md5hash for the restored deployments"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					for _, pdsDeployment := range restoredDeployments {
						err := dsTest.ValidateDataServiceDeployment(pdsDeployment, params.InfraToTest.Namespace)
						log.FailOnError(err, "Error while validating deployment before validating checksum")
						ckSum, wlDep, err := dsTest.ReadDataAndReturnChecksum(pdsDeployment, wkloadParams)
						wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDep)
						log.FailOnError(err, "Error while Running workloads")
						log.Debugf("Checksum for the deployment %s is %s", *pdsDeployment.ClusterResourceName, ckSum)
						restoredDeploymentsmd5Hash[*pdsDeployment.ClusterResourceName] = ckSum
					}

					dash.VerifyFatal(dsTest.ValidateDataMd5Hash(pdsdeploymentsmd5Hash, restoredDeploymentsmd5Hash),
						true, "Validate md5 hash after restore")
				})

				Step("Clean up workload deployments", func() {
					for _, wlDep := range wlDeploymentsToBeCleaned {
						err := k8sApps.DeleteDeployment(wlDep.Name, wlDep.Namespace)
						log.FailOnError(err, "Failed while deleting the workload deployment")
					}
				})

				Step("Delete Deployments", func() {
					CleanupDeployments(deploymentsToBeCleaned)
					controlPlane.CleanupCustomTemplates(stConfigModel.GetId(), resConfigModel.GetId())
					controlPlane.CleanupCustomTemplates(newStorageTemplateID, newResourceTemplateID)
				})
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		err := bkpClient.AWSStorageClient.DeleteBucket()
		log.FailOnError(err, "Failed while deleting the bucket")
	})
})

var _ = Describe("{ResizeStorageAndPerformRestoreXFSRepl3}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ResizeStorageAndPerformRestoreXFSRepl3", "Perform PVC Resize and validate the updated vol in the storage config also perform restore of the ds", pdsLabels, 0)
		credName := targetName + pdsbkp.RandString(8)
		bkpClient, err = pdsbkp.InitializePdsBackup()
		log.FailOnError(err, "Failed to initialize backup for pds.")
		bkpTarget, err = bkpClient.CreateAwsS3BackupCredsAndTarget(tenantID, fmt.Sprintf("%v-aws", credName), deploymentTargetID)
		log.FailOnError(err, "Failed to create S3 backup target.")
		log.InfoD("AWS S3 target - %v created successfully", bkpTarget.GetName())
		awsBkpTargets = append(awsBkpTargets, bkpTarget)
		//Initializing the parameters required for workload generation
		wkloadParams = pdsdriver.LoadGenParams{
			LoadGenDepName: params.LoadGen.LoadGenDepName,
			Namespace:      params.InfraToTest.Namespace,
			NumOfRows:      params.LoadGen.NumOfRows,
			Timeout:        params.LoadGen.Timeout,
			Replicas:       params.LoadGen.Replicas,
			TableName:      params.LoadGen.TableName,
			Iterations:     params.LoadGen.Iterations,
			FailOnError:    params.LoadGen.FailOnError,
		}
	})

	It("Perform PVC Resize and validate the updated vol in the storage config", func() {

		var (
			updatedDeployment        *pds.ModelsDeployment
			restoredDeployments      []*pds.ModelsDeployment
			wlDeploymentsToBeCleaned []*v1.Deployment
			updatedDepList           []*pds.ModelsDeployment
			depList                  []*pds.ModelsDeployment
			resConfigModelUpdated    *pds.ModelsResourceSettingsTemplate
			stConfigModelUpdated     *pds.ModelsStorageOptionsTemplate
			newResourceTemplateID    string
			newStorageTemplateID     string
			updatedPvcSize           uint64
		)
		restoredDeploymentsmd5Hash := make(map[string]string)
		stepLog := "Create Custom Templates , Deploy ds and Trigger Workload"
		Step(stepLog, func() {
			backupSupportedDataServiceNameIDMap, err = bkpClient.GetAllBackupSupportedDataServices()
			log.FailOnError(err, "Error while fetching the backup supported ds.")
			for _, ds := range params.DataServiceToTest {
				log.InfoD(stepLog)

				CleanMapEntries(restoredDeploymentsmd5Hash)

				_, supported := backupSupportedDataServiceNameIDMap[ds.Name]
				if !supported {
					log.InfoD("Data service: %v doesn't support backup, skipping...", ds.Name)
					continue
				}
				deployment, initialCapacity, resConfigModel, stConfigModel, appConfigID, workloadDep, pdsdeploymentsmd5Hash, err := DeployDSWithCustomTemplatesRunWorkloads(ds, tenantID, controlplane.Templates{
					CpuLimit:       "2",
					CpuRequest:     "1",
					MemoryLimit:    "4G",
					MemoryRequest:  "2G",
					StorageRequest: "1G",
					FsType:         "xfs",
					ReplFactor:     3,
					Provisioner:    "pxd.portworx.com",
					Secure:         false,
					VolGroups:      false,
				})
				depList = append(depList, deployment)
				dataserviceID, _ := dsTest.GetDataServiceID(ds.Name)
				stepLog = "Update the resource/storage template with increased storage size"
				Step(stepLog, func() {
					newTemplateName := "autoTemp-" + strconv.Itoa(rand.Int())
					updatedTemplateConfig := controlplane.Templates{
						CpuLimit:       *resConfigModel.CpuLimit,
						CpuRequest:     *resConfigModel.CpuRequest,
						DataServiceID:  dataserviceID,
						MemoryLimit:    *resConfigModel.MemoryLimit,
						MemoryRequest:  *resConfigModel.MemoryRequest,
						Name:           newTemplateName,
						StorageRequest: "500G",
						FsType:         *stConfigModel.Fs,
						ReplFactor:     *stConfigModel.Repl,
						Provisioner:    *stConfigModel.Provisioner,
						Secure:         false,
						VolGroups:      false,
					}
					stConfigModelUpdated, resConfigModelUpdated, err = controlPlane.CreateCustomResourceTemplate(tenantID, updatedTemplateConfig)
					log.FailOnError(err, "Unable to update template")
					log.InfoD("Successfully updated the template with ID- %v", resConfigModelUpdated.GetId())
					newResourceTemplateID = resConfigModelUpdated.GetId()
					newStorageTemplateID = stConfigModelUpdated.GetId()
				})
				stepLog = "Apply updated template to the dataservice deployment"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					if appConfigID == "" {
						appConfigID, err = controlPlane.GetAppConfTemplate(tenantID, ds.Name)
						log.FailOnError(err, "Error while fetching AppConfigID")
					}
					updatedDeployment, err = dsTest.UpdateDataServices(deployment.GetId(),
						appConfigID, deployment.GetImageId(),
						int32(ds.Replicas), newResourceTemplateID, params.InfraToTest.Namespace)
					log.FailOnError(err, "Error while updating dataservices")
					Step("Validate Deployments after template update", func() {
						err = dsTest.ValidateDataServiceDeployment(updatedDeployment, namespace)
						log.FailOnError(err, "Error while validating dataservices")
						log.InfoD("Data-service: %v is up and healthy", ds.Name)
						updatedDepList = append(updatedDepList, updatedDeployment)
						updatedPvcSize, err = GetVolumeCapacityInGB(namespace, updatedDeployment)
						log.InfoD("Updated Storage Size is- %v", updatedPvcSize)
					})
					stepLog = "Validate Workload is running after storage resize"
					Step(stepLog, func() {
						err = k8sApps.ValidateDeployment(workloadDep, timeOut, 10*time.Second)
						log.FailOnError(err, "Workload is not running after Storage Size Increase")
					})
					stepLog = "Verify storage size before and after storage resize - Verify at STS, PV,PVC level"
					Step(stepLog, func() {

						_, _, config, err := pdslib.ValidateDataServiceVolumes(updatedDeployment, ds.Name, newResourceTemplateID, newStorageTemplateID, params.InfraToTest.Namespace)
						log.FailOnError(err, "error on ValidateDataServiceVolumes method")
						log.InfoD("resConfigModel.StorageRequest val is- %v and updated config val is- %v", *resConfigModelUpdated.StorageRequest, config.Spec.Resources.Requests.Storage)
						dash.VerifyFatal(config.Spec.Resources.Requests.Storage, *resConfigModelUpdated.StorageRequest, "Validating the storage size is updated in the config post resize (STS-LEVEL)")
						dash.VerifyFatal(config.Spec.StorageOptions.Filesystem, *stConfigModel.Fs, "Validating the File System Type post storage resize (FileSystem-LEVEL)")
						stringRelFactor := strconv.Itoa(int(*stConfigModel.Repl))
						dash.VerifyFatal(config.Spec.StorageOptions.Replicas, stringRelFactor, "Validating the Replication Factor count post storage resize (RepelFactor-LEVEL)")
						if updatedPvcSize > initialCapacity {
							flag := true
							dash.VerifyFatal(flag, true, "Validating the storage size is updated in the config post resize (PV/PVC-LEVEL)")
							log.InfoD("Initial PVC Capacity is- %v and Updated PVC Capacity is- %v", initialCapacity, updatedPvcSize)
						} else {
							log.FailOnError(err, "Failed to verify Storage Resize at PV/PVC level")
						}
					})
				})
				stepLog = "Perform backup after PVC Resize"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					log.Infof("Deployment ID: %v, backup target ID: %v", updatedDeployment.GetId(), bkpTarget.GetId())
					err = bkpClient.TriggerAndValidateAdhocBackup(updatedDeployment.GetId(), bkpTarget.GetId(), "s3")
					log.FailOnError(err, "Failed while performing adhoc backup.")
				})
				stepLog = "Perform Restore after PVC Resize"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					ctx, err := GetSourceClusterConfigPath()
					log.FailOnError(err, "failed while getting src cluster path")
					restoreTarget := tc.NewTargetCluster(ctx)
					restoreClient := restoreBkp.RestoreClient{
						TenantId:             tenantID,
						ProjectId:            projectID,
						Components:           components,
						Deployment:           updatedDeployment,
						RestoreTargetCluster: restoreTarget,
					}
					backupJobs, err := restoreClient.Components.BackupJob.ListBackupJobsBelongToDeployment(projectID, updatedDeployment.GetId())
					log.FailOnError(err, "Error while fetching the backup jobs for the deployment: %v", updatedDeployment.GetClusterResourceName())
					for _, backupJob := range backupJobs {
						log.InfoD("[Restoring] Details Backup job name- %v, Id- %v", backupJob.GetName(), backupJob.GetId())
						restoredModel, err := restoreClient.TriggerAndValidateRestore(backupJob.GetId(), params.InfraToTest.Namespace, dsEntity, true, true)
						log.FailOnError(err, "Failed during restore.")
						restoredDeployment, err = restoreClient.Components.DataServiceDeployment.GetDeployment(restoredModel.GetDeploymentId())
						log.FailOnError(err, fmt.Sprintf("Failed while fetching the restore data service instance: %v", restoredModel.GetClusterResourceName()))
						restoredDeployments = append(restoredDeployments, restoredDeployment)
						deploymentsToBeCleaned = append(deploymentsToBeCleaned, restoredDeployment)
						log.InfoD("Restored successfully. Deployment- %v", restoredModel.GetClusterResourceName())
					}
				})

				stepLog = "Validate md5hash for the restored deployments"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					for _, pdsDeployment := range restoredDeployments {
						err := dsTest.ValidateDataServiceDeployment(pdsDeployment, params.InfraToTest.Namespace)
						log.FailOnError(err, "Error while validating deployment before validating checksum")
						ckSum, wlDep, err := dsTest.ReadDataAndReturnChecksum(pdsDeployment, wkloadParams)
						wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDep)
						log.FailOnError(err, "Error while Running workloads")
						log.Debugf("Checksum for the deployment %s is %s", *pdsDeployment.ClusterResourceName, ckSum)
						restoredDeploymentsmd5Hash[*pdsDeployment.ClusterResourceName] = ckSum
					}

					dash.VerifyFatal(dsTest.ValidateDataMd5Hash(pdsdeploymentsmd5Hash, restoredDeploymentsmd5Hash),
						true, "Validate md5 hash after restore")
				})

				Step("Clean up workload deployments", func() {
					for _, wlDep := range wlDeploymentsToBeCleaned {
						err := k8sApps.DeleteDeployment(wlDep.Name, wlDep.Namespace)
						log.FailOnError(err, "Failed while deleting the workload deployment")
					}
				})

				Step("Delete Deployments", func() {
					CleanupDeployments(deploymentsToBeCleaned)
					controlPlane.CleanupCustomTemplates(stConfigModel.GetId(), resConfigModel.GetId())
					controlPlane.CleanupCustomTemplates(newStorageTemplateID, newResourceTemplateID)
				})
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		err := bkpClient.AWSStorageClient.DeleteBucket()
		log.FailOnError(err, "Failed while deleting the bucket")
	})
})

var _ = Describe("{ResizeStorageAndPerformRestoreExt4Repl2}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ResizeStorageAndPerformRestoreExt4Repl2", "Perform PVC Resize and validate the updated vol in the storage config also perform restore of the ds", pdsLabels, 0)
		credName := targetName + pdsbkp.RandString(8)
		bkpClient, err = pdsbkp.InitializePdsBackup()
		log.FailOnError(err, "Failed to initialize backup for pds.")
		bkpTarget, err = bkpClient.CreateAwsS3BackupCredsAndTarget(tenantID, fmt.Sprintf("%v-aws", credName), deploymentTargetID)
		log.FailOnError(err, "Failed to create S3 backup target.")
		log.InfoD("AWS S3 target - %v created successfully", bkpTarget.GetName())
		awsBkpTargets = append(awsBkpTargets, bkpTarget)
		//Initializing the parameters required for workload generation
		wkloadParams = pdsdriver.LoadGenParams{
			LoadGenDepName: params.LoadGen.LoadGenDepName,
			Namespace:      params.InfraToTest.Namespace,
			NumOfRows:      params.LoadGen.NumOfRows,
			Timeout:        params.LoadGen.Timeout,
			Replicas:       params.LoadGen.Replicas,
			TableName:      params.LoadGen.TableName,
			Iterations:     params.LoadGen.Iterations,
			FailOnError:    params.LoadGen.FailOnError,
		}
	})

	It("Perform PVC Resize and validate the updated vol in the storage config", func() {

		var (
			updatedDeployment        *pds.ModelsDeployment
			restoredDeployments      []*pds.ModelsDeployment
			wlDeploymentsToBeCleaned []*v1.Deployment
			updatedDepList           []*pds.ModelsDeployment
			depList                  []*pds.ModelsDeployment
			resConfigModelUpdated    *pds.ModelsResourceSettingsTemplate
			stConfigModelUpdated     *pds.ModelsStorageOptionsTemplate
			newResourceTemplateID    string
			newStorageTemplateID     string
			updatedPvcSize           uint64
		)
		restoredDeploymentsmd5Hash := make(map[string]string)
		stepLog := "Create Custom Templates , Deploy ds and Trigger Workload"
		Step(stepLog, func() {
			backupSupportedDataServiceNameIDMap, err = bkpClient.GetAllBackupSupportedDataServices()
			log.FailOnError(err, "Error while fetching the backup supported ds.")
			for _, ds := range params.DataServiceToTest {
				log.InfoD(stepLog)

				CleanMapEntries(restoredDeploymentsmd5Hash)

				_, supported := backupSupportedDataServiceNameIDMap[ds.Name]
				if !supported {
					log.InfoD("Data service: %v doesn't support backup, skipping...", ds.Name)
					continue
				}
				deployment, initialCapacity, resConfigModel, stConfigModel, appConfigID, workloadDep, pdsdeploymentsmd5Hash, err := DeployDSWithCustomTemplatesRunWorkloads(ds, tenantID, controlplane.Templates{
					CpuLimit:       "2",
					CpuRequest:     "1",
					MemoryLimit:    "4G",
					MemoryRequest:  "2G",
					StorageRequest: "1G",
					FsType:         "ext4",
					ReplFactor:     2,
					Provisioner:    "pxd.portworx.com",
					Secure:         false,
					VolGroups:      false,
				})
				depList = append(depList, deployment)
				dataserviceID, _ := dsTest.GetDataServiceID(ds.Name)
				stepLog = "Update the resource/storage template with increased storage size"
				Step(stepLog, func() {
					newTemplateName := "autoTemp-" + strconv.Itoa(rand.Int())
					updatedTemplateConfig := controlplane.Templates{
						CpuLimit:       *resConfigModel.CpuLimit,
						CpuRequest:     *resConfigModel.CpuRequest,
						DataServiceID:  dataserviceID,
						MemoryLimit:    *resConfigModel.MemoryLimit,
						MemoryRequest:  *resConfigModel.MemoryRequest,
						Name:           newTemplateName,
						StorageRequest: "500G",
						FsType:         *stConfigModel.Fs,
						ReplFactor:     *stConfigModel.Repl,
						Provisioner:    *stConfigModel.Provisioner,
						Secure:         false,
						VolGroups:      false,
					}
					stConfigModelUpdated, resConfigModelUpdated, err = controlPlane.CreateCustomResourceTemplate(tenantID, updatedTemplateConfig)
					log.FailOnError(err, "Unable to update template")
					log.InfoD("Successfully updated the template with ID- %v", resConfigModelUpdated.GetId())
					newResourceTemplateID = resConfigModelUpdated.GetId()
					newStorageTemplateID = stConfigModelUpdated.GetId()
				})
				stepLog = "Apply updated template to the dataservice deployment"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					if appConfigID == "" {
						appConfigID, err = controlPlane.GetAppConfTemplate(tenantID, ds.Name)
						log.FailOnError(err, "Error while fetching AppConfigID")
					}
					updatedDeployment, err = dsTest.UpdateDataServices(deployment.GetId(),
						appConfigID, deployment.GetImageId(),
						int32(ds.Replicas), newResourceTemplateID, params.InfraToTest.Namespace)
					log.FailOnError(err, "Error while updating dataservices")
					Step("Validate Deployments after template update", func() {
						err = dsTest.ValidateDataServiceDeployment(updatedDeployment, namespace)
						log.FailOnError(err, "Error while validating dataservices")
						log.InfoD("Data-service: %v is up and healthy", ds.Name)
						updatedDepList = append(updatedDepList, updatedDeployment)
						updatedPvcSize, err = GetVolumeCapacityInGB(namespace, updatedDeployment)
						log.InfoD("Updated Storage Size is- %v", updatedPvcSize)
					})
					stepLog = "Validate Workload is running after storage resize"
					Step(stepLog, func() {
						err = k8sApps.ValidateDeployment(workloadDep, timeOut, 10*time.Second)
						log.FailOnError(err, "Workload is not running after Storage Size Increase")
					})
					stepLog = "Verify storage size before and after storage resize - Verify at STS, PV,PVC level"
					Step(stepLog, func() {

						_, _, config, err := pdslib.ValidateDataServiceVolumes(updatedDeployment, ds.Name, newResourceTemplateID, newStorageTemplateID, params.InfraToTest.Namespace)
						log.FailOnError(err, "error on ValidateDataServiceVolumes method")
						log.InfoD("resConfigModel.StorageRequest val is- %v and updated config val is- %v", *resConfigModelUpdated.StorageRequest, config.Spec.Resources.Requests.Storage)
						dash.VerifyFatal(config.Spec.Resources.Requests.Storage, *resConfigModelUpdated.StorageRequest, "Validating the storage size is updated in the config post resize (STS-LEVEL)")
						dash.VerifyFatal(config.Spec.StorageOptions.Filesystem, *stConfigModel.Fs, "Validating the File System Type post storage resize (FileSystem-LEVEL)")
						stringRelFactor := strconv.Itoa(int(*stConfigModel.Repl))
						dash.VerifyFatal(config.Spec.StorageOptions.Replicas, stringRelFactor, "Validating the Replication Factor count post storage resize (RepelFactor-LEVEL)")
						if updatedPvcSize > initialCapacity {
							flag := true
							dash.VerifyFatal(flag, true, "Validating the storage size is updated in the config post resize (PV/PVC-LEVEL)")
							log.InfoD("Initial PVC Capacity is- %v and Updated PVC Capacity is- %v", initialCapacity, updatedPvcSize)
						} else {
							log.FailOnError(err, "Failed to verify Storage Resize at PV/PVC level")
						}
					})
				})
				stepLog = "Perform backup after PVC Resize"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					log.Infof("Deployment ID: %v, backup target ID: %v", updatedDeployment.GetId(), bkpTarget.GetId())
					err = bkpClient.TriggerAndValidateAdhocBackup(updatedDeployment.GetId(), bkpTarget.GetId(), "s3")
					log.FailOnError(err, "Failed while performing adhoc backup.")
				})
				stepLog = "Perform Restore after PVC Resize"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					ctx, err := GetSourceClusterConfigPath()
					log.FailOnError(err, "failed while getting src cluster path")
					restoreTarget := tc.NewTargetCluster(ctx)
					restoreClient := restoreBkp.RestoreClient{
						TenantId:             tenantID,
						ProjectId:            projectID,
						Components:           components,
						Deployment:           updatedDeployment,
						RestoreTargetCluster: restoreTarget,
					}
					backupJobs, err := restoreClient.Components.BackupJob.ListBackupJobsBelongToDeployment(projectID, updatedDeployment.GetId())
					log.FailOnError(err, "Error while fetching the backup jobs for the deployment: %v", updatedDeployment.GetClusterResourceName())
					for _, backupJob := range backupJobs {
						log.InfoD("[Restoring] Details Backup job name- %v, Id- %v", backupJob.GetName(), backupJob.GetId())
						restoredModel, err := restoreClient.TriggerAndValidateRestore(backupJob.GetId(), params.InfraToTest.Namespace, dsEntity, true, true)
						log.FailOnError(err, "Failed during restore.")
						restoredDeployment, err = restoreClient.Components.DataServiceDeployment.GetDeployment(restoredModel.GetDeploymentId())
						log.FailOnError(err, fmt.Sprintf("Failed while fetching the restore data service instance: %v", restoredModel.GetClusterResourceName()))
						restoredDeployments = append(restoredDeployments, restoredDeployment)
						deploymentsToBeCleaned = append(deploymentsToBeCleaned, restoredDeployment)
						log.InfoD("Restored successfully. Deployment- %v", restoredModel.GetClusterResourceName())
					}
				})

				stepLog = "Validate md5hash for the restored deployments"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					for _, pdsDeployment := range restoredDeployments {
						err := dsTest.ValidateDataServiceDeployment(pdsDeployment, params.InfraToTest.Namespace)
						log.FailOnError(err, "Error while validating deployment before validating checksum")
						ckSum, wlDep, err := dsTest.ReadDataAndReturnChecksum(pdsDeployment, wkloadParams)
						wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDep)
						log.FailOnError(err, "Error while Running workloads")
						log.Debugf("Checksum for the deployment %s is %s", *pdsDeployment.ClusterResourceName, ckSum)
						restoredDeploymentsmd5Hash[*pdsDeployment.ClusterResourceName] = ckSum
					}

					dash.VerifyFatal(dsTest.ValidateDataMd5Hash(pdsdeploymentsmd5Hash, restoredDeploymentsmd5Hash),
						true, "Validate md5 hash after restore")
				})

				Step("Clean up workload deployments", func() {
					for _, wlDep := range wlDeploymentsToBeCleaned {
						err := k8sApps.DeleteDeployment(wlDep.Name, wlDep.Namespace)
						log.FailOnError(err, "Failed while deleting the workload deployment")
					}
				})

				Step("Delete Deployments", func() {
					CleanupDeployments(deploymentsToBeCleaned)
					controlPlane.CleanupCustomTemplates(stConfigModel.GetId(), resConfigModel.GetId())
					controlPlane.CleanupCustomTemplates(newStorageTemplateID, newResourceTemplateID)
				})
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		err := bkpClient.AWSStorageClient.DeleteBucket()
		log.FailOnError(err, "Failed while deleting the bucket")
	})
})

var _ = Describe("{ResizeStorageAndPerformRestoreExt4Repl3}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ResizeStorageAndPerformRestoreExt4Repl3", "Perform PVC Resize and validate the updated vol in the storage config also perform restore of the ds", pdsLabels, 0)
		credName := targetName + pdsbkp.RandString(8)
		bkpClient, err = pdsbkp.InitializePdsBackup()
		log.FailOnError(err, "Failed to initialize backup for pds.")
		bkpTarget, err = bkpClient.CreateAwsS3BackupCredsAndTarget(tenantID, fmt.Sprintf("%v-aws", credName), deploymentTargetID)
		log.FailOnError(err, "Failed to create S3 backup target.")
		log.InfoD("AWS S3 target - %v created successfully", bkpTarget.GetName())
		awsBkpTargets = append(awsBkpTargets, bkpTarget)
		//Initializing the parameters required for workload generation
		wkloadParams = pdsdriver.LoadGenParams{
			LoadGenDepName: params.LoadGen.LoadGenDepName,
			Namespace:      params.InfraToTest.Namespace,
			NumOfRows:      params.LoadGen.NumOfRows,
			Timeout:        params.LoadGen.Timeout,
			Replicas:       params.LoadGen.Replicas,
			TableName:      params.LoadGen.TableName,
			Iterations:     params.LoadGen.Iterations,
			FailOnError:    params.LoadGen.FailOnError,
		}
	})

	It("Perform PVC Resize and validate the updated vol in the storage config", func() {

		var (
			updatedDeployment        *pds.ModelsDeployment
			restoredDeployments      []*pds.ModelsDeployment
			wlDeploymentsToBeCleaned []*v1.Deployment
			updatedDepList           []*pds.ModelsDeployment
			depList                  []*pds.ModelsDeployment
			resConfigModelUpdated    *pds.ModelsResourceSettingsTemplate
			stConfigModelUpdated     *pds.ModelsStorageOptionsTemplate
			newResourceTemplateID    string
			newStorageTemplateID     string
			updatedPvcSize           uint64
		)
		restoredDeploymentsmd5Hash := make(map[string]string)
		stepLog := "Create Custom Templates , Deploy ds and Trigger Workload"
		Step(stepLog, func() {
			backupSupportedDataServiceNameIDMap, err = bkpClient.GetAllBackupSupportedDataServices()
			log.FailOnError(err, "Error while fetching the backup supported ds.")
			for _, ds := range params.DataServiceToTest {
				log.InfoD(stepLog)

				CleanMapEntries(restoredDeploymentsmd5Hash)

				_, supported := backupSupportedDataServiceNameIDMap[ds.Name]
				if !supported {
					log.InfoD("Data service: %v doesn't support backup, skipping...", ds.Name)
					continue
				}
				deployment, initialCapacity, resConfigModel, stConfigModel, appConfigID, workloadDep, pdsdeploymentsmd5Hash, err := DeployDSWithCustomTemplatesRunWorkloads(ds, tenantID, controlplane.Templates{
					CpuLimit:       "2",
					CpuRequest:     "1",
					MemoryLimit:    "4G",
					MemoryRequest:  "2G",
					StorageRequest: "1G",
					FsType:         "ext4",
					ReplFactor:     3,
					Provisioner:    "pxd.portworx.com",
					Secure:         false,
					VolGroups:      false,
				})
				depList = append(depList, deployment)
				dataserviceID, _ := dsTest.GetDataServiceID(ds.Name)
				stepLog = "Update the resource/storage template with increased storage size"
				Step(stepLog, func() {
					newTemplateName := "autoTemp-" + strconv.Itoa(rand.Int())
					updatedTemplateConfig := controlplane.Templates{
						CpuLimit:       *resConfigModel.CpuLimit,
						CpuRequest:     *resConfigModel.CpuRequest,
						DataServiceID:  dataserviceID,
						MemoryLimit:    *resConfigModel.MemoryLimit,
						MemoryRequest:  *resConfigModel.MemoryRequest,
						Name:           newTemplateName,
						StorageRequest: "500G",
						FsType:         *stConfigModel.Fs,
						ReplFactor:     *stConfigModel.Repl,
						Provisioner:    *stConfigModel.Provisioner,
						Secure:         false,
						VolGroups:      false,
					}
					stConfigModelUpdated, resConfigModelUpdated, err = controlPlane.CreateCustomResourceTemplate(tenantID, updatedTemplateConfig)
					log.FailOnError(err, "Unable to update template")
					log.InfoD("Successfully updated the template with ID- %v", resConfigModelUpdated.GetId())
					newResourceTemplateID = resConfigModelUpdated.GetId()
					newStorageTemplateID = stConfigModelUpdated.GetId()
				})
				stepLog = "Apply updated template to the dataservice deployment"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					if appConfigID == "" {
						appConfigID, err = controlPlane.GetAppConfTemplate(tenantID, ds.Name)
						log.FailOnError(err, "Error while fetching AppConfigID")
					}
					updatedDeployment, err = dsTest.UpdateDataServices(deployment.GetId(),
						appConfigID, deployment.GetImageId(),
						int32(ds.Replicas), newResourceTemplateID, params.InfraToTest.Namespace)
					log.FailOnError(err, "Error while updating dataservices")
					Step("Validate Deployments after template update", func() {
						err = dsTest.ValidateDataServiceDeployment(updatedDeployment, namespace)
						log.FailOnError(err, "Error while validating dataservices")
						log.InfoD("Data-service: %v is up and healthy", ds.Name)
						updatedDepList = append(updatedDepList, updatedDeployment)
						updatedPvcSize, err = GetVolumeCapacityInGB(namespace, updatedDeployment)
						log.InfoD("Updated Storage Size is- %v", updatedPvcSize)
					})
					stepLog = "Validate Workload is running after storage resize"
					Step(stepLog, func() {
						err = k8sApps.ValidateDeployment(workloadDep, timeOut, 10*time.Second)
						log.FailOnError(err, "Workload is not running after Storage Size Increase")
					})
					stepLog = "Verify storage size before and after storage resize - Verify at STS, PV,PVC level"
					Step(stepLog, func() {

						_, _, config, err := pdslib.ValidateDataServiceVolumes(updatedDeployment, ds.Name, newResourceTemplateID, newStorageTemplateID, params.InfraToTest.Namespace)
						log.FailOnError(err, "error on ValidateDataServiceVolumes method")
						log.InfoD("resConfigModel.StorageRequest val is- %v and updated config val is- %v", *resConfigModelUpdated.StorageRequest, config.Spec.Resources.Requests.Storage)
						dash.VerifyFatal(config.Spec.Resources.Requests.Storage, *resConfigModelUpdated.StorageRequest, "Validating the storage size is updated in the config post resize (STS-LEVEL)")
						dash.VerifyFatal(config.Spec.StorageOptions.Filesystem, *stConfigModel.Fs, "Validating the File System Type post storage resize (FileSystem-LEVEL)")
						stringRelFactor := strconv.Itoa(int(*stConfigModel.Repl))
						dash.VerifyFatal(config.Spec.StorageOptions.Replicas, stringRelFactor, "Validating the Replication Factor count post storage resize (RepelFactor-LEVEL)")
						if updatedPvcSize > initialCapacity {
							flag := true
							dash.VerifyFatal(flag, true, "Validating the storage size is updated in the config post resize (PV/PVC-LEVEL)")
							log.InfoD("Initial PVC Capacity is- %v and Updated PVC Capacity is- %v", initialCapacity, updatedPvcSize)
						} else {
							log.FailOnError(err, "Failed to verify Storage Resize at PV/PVC level")
						}
					})
				})
				stepLog = "Perform backup after PVC Resize"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					log.Infof("Deployment ID: %v, backup target ID: %v", updatedDeployment.GetId(), bkpTarget.GetId())
					err = bkpClient.TriggerAndValidateAdhocBackup(updatedDeployment.GetId(), bkpTarget.GetId(), "s3")
					log.FailOnError(err, "Failed while performing adhoc backup.")
				})
				stepLog = "Perform Restore after PVC Resize"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					ctx, err := GetSourceClusterConfigPath()
					log.FailOnError(err, "failed while getting src cluster path")
					restoreTarget := tc.NewTargetCluster(ctx)
					restoreClient := restoreBkp.RestoreClient{
						TenantId:             tenantID,
						ProjectId:            projectID,
						Components:           components,
						Deployment:           updatedDeployment,
						RestoreTargetCluster: restoreTarget,
					}
					backupJobs, err := restoreClient.Components.BackupJob.ListBackupJobsBelongToDeployment(projectID, updatedDeployment.GetId())
					log.FailOnError(err, "Error while fetching the backup jobs for the deployment: %v", updatedDeployment.GetClusterResourceName())
					for _, backupJob := range backupJobs {
						log.InfoD("[Restoring] Details Backup job name- %v, Id- %v", backupJob.GetName(), backupJob.GetId())
						restoredModel, err := restoreClient.TriggerAndValidateRestore(backupJob.GetId(), params.InfraToTest.Namespace, dsEntity, true, true)
						log.FailOnError(err, "Failed during restore.")
						restoredDeployment, err = restoreClient.Components.DataServiceDeployment.GetDeployment(restoredModel.GetDeploymentId())
						log.FailOnError(err, fmt.Sprintf("Failed while fetching the restore data service instance: %v", restoredModel.GetClusterResourceName()))
						restoredDeployments = append(restoredDeployments, restoredDeployment)
						deploymentsToBeCleaned = append(deploymentsToBeCleaned, restoredDeployment)
						log.InfoD("Restored successfully. Deployment- %v", restoredModel.GetClusterResourceName())
					}
				})

				stepLog = "Validate md5hash for the restored deployments"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					for _, pdsDeployment := range restoredDeployments {
						err := dsTest.ValidateDataServiceDeployment(pdsDeployment, params.InfraToTest.Namespace)
						log.FailOnError(err, "Error while validating deployment before validating checksum")
						ckSum, wlDep, err := dsTest.ReadDataAndReturnChecksum(pdsDeployment, wkloadParams)
						wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDep)
						log.FailOnError(err, "Error while Running workloads")
						log.Debugf("Checksum for the deployment %s is %s", *pdsDeployment.ClusterResourceName, ckSum)
						restoredDeploymentsmd5Hash[*pdsDeployment.ClusterResourceName] = ckSum
					}

					dash.VerifyFatal(dsTest.ValidateDataMd5Hash(pdsdeploymentsmd5Hash, restoredDeploymentsmd5Hash),
						true, "Validate md5 hash after restore")
				})

				Step("Clean up workload deployments", func() {
					for _, wlDep := range wlDeploymentsToBeCleaned {
						err := k8sApps.DeleteDeployment(wlDep.Name, wlDep.Namespace)
						log.FailOnError(err, "Failed while deleting the workload deployment")
					}
				})

				Step("Delete Deployments", func() {
					CleanupDeployments(deploymentsToBeCleaned)
					controlPlane.CleanupCustomTemplates(stConfigModel.GetId(), resConfigModel.GetId())
					controlPlane.CleanupCustomTemplates(newStorageTemplateID, newResourceTemplateID)
				})
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		err := bkpClient.AWSStorageClient.DeleteBucket()
		log.FailOnError(err, "Failed while deleting the bucket")
	})
})

var _ = Describe("{ScaleUpDsPostStorageSizeIncreaseRepl2}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ScaleUpDsPostStorageSizeIncreaseRepl2", "Scale up the DS and Perform PVC Resize, validate the updated vol in the storage config.", pdsLabels, 0)
		//Initializing the parameters required for workload generation
		wkloadParams = pdsdriver.LoadGenParams{
			LoadGenDepName: params.LoadGen.LoadGenDepName,
			Namespace:      params.InfraToTest.Namespace,
			NumOfRows:      params.LoadGen.NumOfRows,
			Timeout:        params.LoadGen.Timeout,
			Replicas:       params.LoadGen.Replicas,
			TableName:      params.LoadGen.TableName,
			Iterations:     params.LoadGen.Iterations,
			FailOnError:    params.LoadGen.FailOnError,
		}
	})

	It("Perform PVC Resize and validate the updated vol in the storage config", func() {

		var (
			updatedDeployment        *pds.ModelsDeployment
			updatedDeployment1       *pds.ModelsDeployment
			wlDeploymentsToBeCleaned []*v1.Deployment
			updatedDepList           []*pds.ModelsDeployment
			depList                  []*pds.ModelsDeployment
			updatedDepList1          []*pds.ModelsDeployment
			resConfigModelUpdated1   *pds.ModelsResourceSettingsTemplate
			stConfigModelUpdated1    *pds.ModelsStorageOptionsTemplate
			newResourceTemplateID1   string
			newStorageTemplateID1    string
			resConfigModelUpdated2   *pds.ModelsResourceSettingsTemplate
			stConfigModelUpdated2    *pds.ModelsStorageOptionsTemplate
			newResourceTemplateID2   string
			newStorageTemplateID2    string
			updatedPvcSize           uint64
			updatedPvcSize1          uint64
		)

		stepLog := "Create Custom Templates , Deploy ds and Trigger Workload"
		Step(stepLog, func() {
			for _, ds := range params.DataServiceToTest {
				log.InfoD(stepLog)
				deployment, initialCapacity, resConfigModel, stConfigModel, appConfigID, workloadDep, _, err := DeployDSWithCustomTemplatesRunWorkloads(ds, tenantID, controlplane.Templates{
					CpuLimit:       "2",
					CpuRequest:     "1",
					MemoryLimit:    "4G",
					MemoryRequest:  "2G",
					StorageRequest: "1G",
					FsType:         "xfs",
					ReplFactor:     2,
					Provisioner:    "pxd.portworx.com",
					Secure:         false,
					VolGroups:      false,
				})
				depList = append(depList, deployment)
				dataserviceID, _ := dsTest.GetDataServiceID(ds.Name)
				stepLog = "Check PVC for full condition based upto 90% full"
				stepLog = "Scale up the DS with increased storage size and Repl factor as 2 "
				Step(stepLog, func() {
					newTemplateName1 := "autoTemp-" + strconv.Itoa(rand.Int())
					updatedTemplateConfig1 := controlplane.Templates{
						CpuLimit:       *resConfigModel.CpuLimit,
						CpuRequest:     *resConfigModel.CpuRequest,
						DataServiceID:  dataserviceID,
						MemoryLimit:    *resConfigModel.MemoryLimit,
						MemoryRequest:  *resConfigModel.MemoryRequest,
						Name:           newTemplateName1,
						StorageRequest: "500G",
						FsType:         *stConfigModel.Fs,
						ReplFactor:     *stConfigModel.Repl,
						Provisioner:    *stConfigModel.Provisioner,
						Secure:         false,
						VolGroups:      false,
					}
					stConfigModelUpdated1, resConfigModelUpdated1, err = controlPlane.CreateCustomResourceTemplate(tenantID, updatedTemplateConfig1)
					log.FailOnError(err, "Unable to update template")
					log.InfoD("Successfully updated the template with ID- %v and ReplicationFactor- %v", resConfigModelUpdated1.GetId(), updatedTemplateConfig1.ReplFactor)
					newResourceTemplateID1 = resConfigModelUpdated1.GetId()
					newStorageTemplateID1 = stConfigModelUpdated1.GetId()
				})
				stepLog = "Apply updated template to the dataservice deployment"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					if appConfigID == "" {
						appConfigID, err = controlPlane.GetAppConfTemplate(tenantID, ds.Name)
						log.FailOnError(err, "Error while fetching AppConfigID")
					}
					updatedDeployment, err = dsTest.UpdateDataServices(deployment.GetId(),
						appConfigID, deployment.GetImageId(),
						int32(ds.ScaleReplicas), newResourceTemplateID1, params.InfraToTest.Namespace)
					log.FailOnError(err, "Error while updating dataservices")
					Step("Validate Deployments after template update", func() {
						err = dsTest.ValidateDataServiceDeployment(updatedDeployment, namespace)
						log.FailOnError(err, "Error while validating dataservices")
						log.InfoD("Data-service: %v is up and healthy", ds.Name)
						updatedDepList = append(updatedDepList, updatedDeployment)
						updatedPvcSize, err = GetVolumeCapacityInGB(namespace, updatedDeployment)
						log.InfoD("Updated Storage Size is- %v", updatedPvcSize)
					})
					stepLog = "Validate Workload is running after storage resize"
					Step(stepLog, func() {
						err = k8sApps.ValidateDeployment(workloadDep, timeOut, 10*time.Second)
						log.FailOnError(err, "Workload is not running after Storage Size Increase")
					})
					stepLog = "Verify storage size before and after storage resize - Verify at STS, PV,PVC level"
					Step(stepLog, func() {

						_, _, config, err := pdslib.ValidateDataServiceVolumes(updatedDeployment, ds.Name, newResourceTemplateID1, newStorageTemplateID1, params.InfraToTest.Namespace)
						log.FailOnError(err, "error on ValidateDataServiceVolumes method")
						log.InfoD("resConfigModel.StorageRequest val is- %v and updated config val is- %v", *resConfigModelUpdated1.StorageRequest, config.Spec.Resources.Requests.Storage)
						dash.VerifyFatal(config.Spec.Resources.Requests.Storage, *resConfigModelUpdated1.StorageRequest, "Validating the storage size is updated in the config post resize (STS-LEVEL)")
						dash.VerifyFatal(config.Spec.Nodes, int32(ds.ScaleReplicas), "Verify node scale up is successful with storage resize")
						stringRelFactor := strconv.Itoa(int(*stConfigModelUpdated1.Repl))
						dash.VerifyFatal(config.Spec.StorageOptions.Replicas, stringRelFactor, "Validating the Replication Factor count post storage resize (RepelFactor-LEVEL)")

						if updatedPvcSize > initialCapacity {
							flag := true
							dash.VerifyFatal(flag, true, "Validating the storage size is updated in the config post resize (PV/PVC-LEVEL)")
							log.InfoD("Initial PVC Capacity is- %v and Updated PVC Capacity is- %v", initialCapacity, updatedPvcSize)
						} else {
							log.FailOnError(err, "Failed to verify Storage Resize at PV/PVC level")
						}
					})
				})
				stepLog = "Increase the storage size again after Scale-UP"
				Step(stepLog, func() {
					newTemplateName2 := "autoTemp-" + strconv.Itoa(rand.Int())
					updatedTemplateConfig2 := controlplane.Templates{
						CpuLimit:       *resConfigModelUpdated1.CpuLimit,
						CpuRequest:     *resConfigModelUpdated1.CpuRequest,
						DataServiceID:  dataserviceID,
						MemoryLimit:    *resConfigModelUpdated1.MemoryLimit,
						MemoryRequest:  *resConfigModelUpdated1.MemoryRequest,
						Name:           newTemplateName2,
						StorageRequest: "1000G",
						FsType:         *stConfigModel.Fs,
						ReplFactor:     *stConfigModel.Repl,
						Provisioner:    *stConfigModelUpdated1.Provisioner,
						Secure:         false,
						VolGroups:      false,
					}
					stConfigModelUpdated2, resConfigModelUpdated2, err = controlPlane.CreateCustomResourceTemplate(tenantID, updatedTemplateConfig2)
					log.FailOnError(err, "Unable to update template")
					log.InfoD("Successfully updated the template with ID- %v and ReplicationFactor- %v", resConfigModelUpdated1.GetId(), updatedTemplateConfig2.ReplFactor)
					newResourceTemplateID2 = resConfigModelUpdated2.GetId()
					newStorageTemplateID2 = stConfigModelUpdated2.GetId()
				})
				stepLog = "Apply updated template to the dataservice deployment"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					if appConfigID == "" {
						appConfigID, err = controlPlane.GetAppConfTemplate(tenantID, ds.Name)
						log.FailOnError(err, "Error while fetching AppConfigID")
					}
					updatedDeployment1, err = dsTest.UpdateDataServices(deployment.GetId(),
						appConfigID, deployment.GetImageId(),
						int32(ds.ScaleReplicas), newResourceTemplateID2, params.InfraToTest.Namespace)
					log.FailOnError(err, "Error while updating dataservices")
					Step("Validate Deployments after template update", func() {
						err = dsTest.ValidateDataServiceDeployment(updatedDeployment1, namespace)
						log.FailOnError(err, "Error while validating dataservices")
						log.InfoD("Data-service: %v is up and healthy", ds.Name)
						updatedDepList1 = append(updatedDepList1, updatedDeployment1)
						updatedPvcSize1, err = GetVolumeCapacityInGB(namespace, updatedDeployment)
						log.InfoD("Updated Storage Size is- %v", updatedPvcSize1)
					})
					stepLog = "Validate Workload is running after storage resize"
					Step(stepLog, func() {
						err = k8sApps.ValidateDeployment(workloadDep, timeOut, 10*time.Second)
						log.FailOnError(err, "Workload is not running after Storage Size Increase")
					})
					stepLog = "Verify storage size before and after storage resize - Verify at STS, PV,PVC level"
					Step(stepLog, func() {

						_, _, config, err := pdslib.ValidateDataServiceVolumes(updatedDeployment1, ds.Name, newResourceTemplateID2, newStorageTemplateID2, params.InfraToTest.Namespace)
						log.FailOnError(err, "error on ValidateDataServiceVolumes method")
						log.InfoD("resConfigModel.StorageRequest val is- %v and updated config val is- %v", *resConfigModelUpdated2.StorageRequest, config.Spec.Resources.Requests.Storage)
						dash.VerifyFatal(config.Spec.Resources.Requests.Storage, *resConfigModelUpdated2.StorageRequest, "Validating the storage size is updated in the config post resize (STS-LEVEL)")
						dash.VerifyFatal(config.Spec.Nodes, int32(ds.ScaleReplicas), "Verify node scale up is successful with storage resize")
						stringRelFactor := strconv.Itoa(int(*stConfigModelUpdated2.Repl))
						dash.VerifyFatal(config.Spec.StorageOptions.Replicas, stringRelFactor, "Validating the Replication Factor count post storage resize (RepelFactor-LEVEL)")

						if updatedPvcSize1 > updatedPvcSize {
							flag := true
							dash.VerifyFatal(flag, true, "Validating the storage size is updated in the config post resize (PV/PVC-LEVEL)")
							log.InfoD("Initial PVC Capacity is- %v and Updated PVC Capacity is- %v", updatedPvcSize, updatedPvcSize1)
						} else {
							log.FailOnError(err, "Failed to verify Storage Resize at PV/PVC level")
						}
					})
				})
				Step("Clean up workload deployments", func() {
					for _, wlDep := range wlDeploymentsToBeCleaned {
						err := k8sApps.DeleteDeployment(wlDep.Name, wlDep.Namespace)
						log.FailOnError(err, "Failed while deleting the workload deployment")
					}
				})

				Step("Delete Deployments", func() {
					CleanupDeployments(deploymentsToBeCleaned)
					controlPlane.CleanupCustomTemplates(stConfigModel.GetId(), resConfigModel.GetId())
					controlPlane.CleanupCustomTemplates(newStorageTemplateID1, newResourceTemplateID1)
					controlPlane.CleanupCustomTemplates(newStorageTemplateID2, newResourceTemplateID2)
				})
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{ScaleUpDsPostStorageSizeIncreaseRepl3}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ScaleUpDsPostStorageSizeIncreaseRepl3", "Scale up the DS and Perform PVC Resize, validate the updated vol in the storage config.", pdsLabels, 0)
		//Initializing the parameters required for workload generation
		wkloadParams = pdsdriver.LoadGenParams{
			LoadGenDepName: params.LoadGen.LoadGenDepName,
			Namespace:      params.InfraToTest.Namespace,
			NumOfRows:      params.LoadGen.NumOfRows,
			Timeout:        params.LoadGen.Timeout,
			Replicas:       params.LoadGen.Replicas,
			TableName:      params.LoadGen.TableName,
			Iterations:     params.LoadGen.Iterations,
			FailOnError:    params.LoadGen.FailOnError,
		}
	})

	It("Perform PVC Resize and validate the updated vol in the storage config", func() {

		var (
			updatedDeployment        *pds.ModelsDeployment
			updatedDeployment1       *pds.ModelsDeployment
			wlDeploymentsToBeCleaned []*v1.Deployment
			updatedDepList           []*pds.ModelsDeployment
			depList                  []*pds.ModelsDeployment
			updatedDepList1          []*pds.ModelsDeployment
			resConfigModelUpdated1   *pds.ModelsResourceSettingsTemplate
			stConfigModelUpdated1    *pds.ModelsStorageOptionsTemplate
			newResourceTemplateID1   string
			newStorageTemplateID1    string
			resConfigModelUpdated2   *pds.ModelsResourceSettingsTemplate
			stConfigModelUpdated2    *pds.ModelsStorageOptionsTemplate
			newResourceTemplateID2   string
			newStorageTemplateID2    string
			updatedPvcSize           uint64
			updatedPvcSize1          uint64
		)

		stepLog := "Create Custom Templates , Deploy ds and Trigger Workload"
		Step(stepLog, func() {
			for _, ds := range params.DataServiceToTest {
				log.InfoD(stepLog)
				deployment, initialCapacity, resConfigModel, stConfigModel, appConfigID, workloadDep, _, err := DeployDSWithCustomTemplatesRunWorkloads(ds, tenantID, controlplane.Templates{
					CpuLimit:       "2",
					CpuRequest:     "1",
					MemoryLimit:    "4G",
					MemoryRequest:  "2G",
					StorageRequest: "1G",
					FsType:         "xfs",
					ReplFactor:     3,
					Provisioner:    "pxd.portworx.com",
					Secure:         false,
					VolGroups:      false,
				})
				depList = append(depList, deployment)
				dataserviceID, _ := dsTest.GetDataServiceID(ds.Name)
				stepLog = "Check PVC for full condition based upto 90% full"
				stepLog = "Scale up the DS with increased storage size and Repl factor as 2 "
				Step(stepLog, func() {
					newTemplateName1 := "autoTemp-" + strconv.Itoa(rand.Int())
					updatedTemplateConfig1 := controlplane.Templates{
						CpuLimit:       *resConfigModel.CpuLimit,
						CpuRequest:     *resConfigModel.CpuRequest,
						DataServiceID:  dataserviceID,
						MemoryLimit:    *resConfigModel.MemoryLimit,
						MemoryRequest:  *resConfigModel.MemoryRequest,
						Name:           newTemplateName1,
						StorageRequest: "500G",
						FsType:         *stConfigModel.Fs,
						ReplFactor:     *stConfigModel.Repl,
						Provisioner:    *stConfigModel.Provisioner,
						Secure:         false,
						VolGroups:      false,
					}
					stConfigModelUpdated1, resConfigModelUpdated1, err = controlPlane.CreateCustomResourceTemplate(tenantID, updatedTemplateConfig1)
					log.FailOnError(err, "Unable to update template")
					log.InfoD("Successfully updated the template with ID- %v and ReplicationFactor- %v", resConfigModelUpdated1.GetId(), updatedTemplateConfig1.ReplFactor)
					newResourceTemplateID1 = resConfigModelUpdated1.GetId()
					newStorageTemplateID1 = stConfigModelUpdated1.GetId()
				})
				stepLog = "Apply updated template to the dataservice deployment"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					if appConfigID == "" {
						appConfigID, err = controlPlane.GetAppConfTemplate(tenantID, ds.Name)
						log.FailOnError(err, "Error while fetching AppConfigID")
					}
					updatedDeployment, err = dsTest.UpdateDataServices(deployment.GetId(),
						appConfigID, deployment.GetImageId(),
						int32(ds.ScaleReplicas), newResourceTemplateID1, params.InfraToTest.Namespace)
					log.FailOnError(err, "Error while updating dataservices")
					Step("Validate Deployments after template update", func() {
						err = dsTest.ValidateDataServiceDeployment(updatedDeployment, namespace)
						log.FailOnError(err, "Error while validating dataservices")
						log.InfoD("Data-service: %v is up and healthy", ds.Name)
						updatedDepList = append(updatedDepList, updatedDeployment)
						updatedPvcSize, err = GetVolumeCapacityInGB(namespace, updatedDeployment)
						log.InfoD("Updated Storage Size is- %v", updatedPvcSize)
					})
					stepLog = "Validate Workload is running after storage resize"
					Step(stepLog, func() {
						err = k8sApps.ValidateDeployment(workloadDep, timeOut, 10*time.Second)
						log.FailOnError(err, "Workload is not running after Storage Size Increase")
					})
					stepLog = "Verify storage size before and after storage resize - Verify at STS, PV,PVC level"
					Step(stepLog, func() {

						_, _, config, err := pdslib.ValidateDataServiceVolumes(updatedDeployment, ds.Name, newResourceTemplateID1, newStorageTemplateID1, params.InfraToTest.Namespace)
						log.FailOnError(err, "error on ValidateDataServiceVolumes method")
						log.InfoD("resConfigModel.StorageRequest val is- %v and updated config val is- %v", *resConfigModelUpdated1.StorageRequest, config.Spec.Resources.Requests.Storage)
						dash.VerifyFatal(config.Spec.Resources.Requests.Storage, *resConfigModelUpdated1.StorageRequest, "Validating the storage size is updated in the config post resize (STS-LEVEL)")
						dash.VerifyFatal(config.Spec.Nodes, int32(ds.ScaleReplicas), "Verify node scale up is successful with storage resize")
						stringRelFactor := strconv.Itoa(int(*stConfigModelUpdated1.Repl))
						dash.VerifyFatal(config.Spec.StorageOptions.Replicas, stringRelFactor, "Validating the Replication Factor count post storage resize (RepelFactor-LEVEL)")

						if updatedPvcSize > initialCapacity {
							flag := true
							dash.VerifyFatal(flag, true, "Validating the storage size is updated in the config post resize (PV/PVC-LEVEL)")
							log.InfoD("Initial PVC Capacity is- %v and Updated PVC Capacity is- %v", initialCapacity, updatedPvcSize)
						} else {
							log.FailOnError(err, "Failed to verify Storage Resize at PV/PVC level")
						}
					})
				})
				stepLog = "Increase the storage size again after Scale-UP"
				Step(stepLog, func() {
					newTemplateName2 := "autoTemp-" + strconv.Itoa(rand.Int())
					updatedTemplateConfig2 := controlplane.Templates{
						CpuLimit:       *resConfigModelUpdated1.CpuLimit,
						CpuRequest:     *resConfigModelUpdated1.CpuRequest,
						DataServiceID:  dataserviceID,
						MemoryLimit:    *resConfigModelUpdated1.MemoryLimit,
						MemoryRequest:  *resConfigModelUpdated1.MemoryRequest,
						Name:           newTemplateName2,
						StorageRequest: "1000G",
						FsType:         *stConfigModel.Fs,
						ReplFactor:     *stConfigModel.Repl,
						Provisioner:    *stConfigModelUpdated1.Provisioner,
						Secure:         false,
						VolGroups:      false,
					}
					stConfigModelUpdated2, resConfigModelUpdated2, err = controlPlane.CreateCustomResourceTemplate(tenantID, updatedTemplateConfig2)
					log.FailOnError(err, "Unable to update template")
					log.InfoD("Successfully updated the template with ID- %v and ReplicationFactor- %v", resConfigModelUpdated1.GetId(), updatedTemplateConfig2.ReplFactor)
					newResourceTemplateID2 = resConfigModelUpdated2.GetId()
					newStorageTemplateID2 = stConfigModelUpdated2.GetId()
				})
				stepLog = "Apply updated template to the dataservice deployment"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					if appConfigID == "" {
						appConfigID, err = controlPlane.GetAppConfTemplate(tenantID, ds.Name)
						log.FailOnError(err, "Error while fetching AppConfigID")
					}
					updatedDeployment1, err = dsTest.UpdateDataServices(deployment.GetId(),
						appConfigID, deployment.GetImageId(),
						int32(ds.ScaleReplicas), newResourceTemplateID2, params.InfraToTest.Namespace)
					log.FailOnError(err, "Error while updating dataservices")
					Step("Validate Deployments after template update", func() {
						err = dsTest.ValidateDataServiceDeployment(updatedDeployment1, namespace)
						log.FailOnError(err, "Error while validating dataservices")
						log.InfoD("Data-service: %v is up and healthy", ds.Name)
						updatedDepList1 = append(updatedDepList1, updatedDeployment1)
						updatedPvcSize1, err = GetVolumeCapacityInGB(namespace, updatedDeployment)
						log.InfoD("Updated Storage Size is- %v", updatedPvcSize1)
					})
					stepLog = "Validate Workload is running after storage resize"
					Step(stepLog, func() {
						err = k8sApps.ValidateDeployment(workloadDep, timeOut, 10*time.Second)
						log.FailOnError(err, "Workload is not running after Storage Size Increase")
					})
					stepLog = "Verify storage size before and after storage resize - Verify at STS, PV,PVC level"
					Step(stepLog, func() {

						_, _, config, err := pdslib.ValidateDataServiceVolumes(updatedDeployment1, ds.Name, newResourceTemplateID2, newStorageTemplateID2, params.InfraToTest.Namespace)
						log.FailOnError(err, "error on ValidateDataServiceVolumes method")
						log.InfoD("resConfigModel.StorageRequest val is- %v and updated config val is- %v", *resConfigModelUpdated2.StorageRequest, config.Spec.Resources.Requests.Storage)
						dash.VerifyFatal(config.Spec.Resources.Requests.Storage, *resConfigModelUpdated2.StorageRequest, "Validating the storage size is updated in the config post resize (STS-LEVEL)")
						dash.VerifyFatal(config.Spec.Nodes, int32(ds.ScaleReplicas), "Verify node scale up is successful with storage resize")
						stringRelFactor := strconv.Itoa(int(*stConfigModelUpdated2.Repl))
						dash.VerifyFatal(config.Spec.StorageOptions.Replicas, stringRelFactor, "Validating the Replication Factor count post storage resize (RepelFactor-LEVEL)")

						if updatedPvcSize1 > updatedPvcSize {
							flag := true
							dash.VerifyFatal(flag, true, "Validating the storage size is updated in the config post resize (PV/PVC-LEVEL)")
							log.InfoD("Initial PVC Capacity is- %v and Updated PVC Capacity is- %v", updatedPvcSize, updatedPvcSize1)
						} else {
							log.FailOnError(err, "Failed to verify Storage Resize at PV/PVC level")
						}
					})
				})
				Step("Clean up workload deployments", func() {
					for _, wlDep := range wlDeploymentsToBeCleaned {
						err := k8sApps.DeleteDeployment(wlDep.Name, wlDep.Namespace)
						log.FailOnError(err, "Failed while deleting the workload deployment")
					}
				})

				Step("Delete Deployments", func() {
					CleanupDeployments(deploymentsToBeCleaned)
					controlPlane.CleanupCustomTemplates(stConfigModel.GetId(), resConfigModel.GetId())
					controlPlane.CleanupCustomTemplates(newStorageTemplateID1, newResourceTemplateID1)
					controlPlane.CleanupCustomTemplates(newStorageTemplateID2, newResourceTemplateID2)
				})
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{PerformStorageResizeBy1Gb100Times}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("PerformStorageResizeBy1Gb100Times", "Perform PVC Resize by 1GB for 100 times in a loop and validate the updated vol in the storage config.", pdsLabels, 0)
		//Initializing the parameters required for workload generation
		wkloadParams = pdsdriver.LoadGenParams{
			LoadGenDepName: params.LoadGen.LoadGenDepName,
			Namespace:      params.InfraToTest.Namespace,
			NumOfRows:      params.LoadGen.NumOfRows,
			Timeout:        params.LoadGen.Timeout,
			Replicas:       params.LoadGen.Replicas,
			TableName:      params.LoadGen.TableName,
			Iterations:     params.LoadGen.Iterations,
			FailOnError:    params.LoadGen.FailOnError,
		}
	})

	It("Perform PVC Resize and validate the updated vol in the storage config", func() {

		var (
			updatedDeployment     *pds.ModelsDeployment
			updatedDepList        []*pds.ModelsDeployment
			resConfigModelUpdated *pds.ModelsResourceSettingsTemplate
			stConfigModelUpdated  *pds.ModelsStorageOptionsTemplate
			newResourceTemplateID string
			newStorageTemplateID  string
			updatedPvcSize        uint64
		)
		stepLog := "Create Custom Templates , Deploy ds and Trigger Workload"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for _, ds := range params.DataServiceToTest {
				deployment, initialCapacity, resConfigModel, stConfigModel, appConfigID, _, _, err := DeployDSWithCustomTemplatesRunWorkloads(ds, tenantID, controlplane.Templates{
					CpuLimit:       "2",
					CpuRequest:     "1",
					MemoryLimit:    "4G",
					MemoryRequest:  "2G",
					StorageRequest: "1G",
					FsType:         "xfs",
					ReplFactor:     2,
					Provisioner:    "pxd.portworx.com",
					Secure:         false,
					VolGroups:      false,
				})

				dataserviceID, _ := dsTest.GetDataServiceID(ds.Name)
				stepLog = "Check PVC for full condition based upto 90% full"
				storageSizeCounter := 0
				for i := 2; i <= 10; i++ {
					log.InfoD("Iam exuting for the %v time", i)
					storageSizeCounter = i
					storageSize := fmt.Sprint(storageSizeCounter, "G")
					log.InfoD("StorageSize calculated is %v", storageSize)
					stepLog = "Update the resource/storage template with increased storage size"
					Step(stepLog, func() {
						newTemplateName := "autoTemp-" + strconv.Itoa(rand.Int())
						updatedTemplateConfig := controlplane.Templates{
							CpuLimit:       *resConfigModel.CpuLimit,
							CpuRequest:     *resConfigModel.CpuRequest,
							DataServiceID:  dataserviceID,
							MemoryLimit:    *resConfigModel.MemoryLimit,
							MemoryRequest:  *resConfigModel.MemoryRequest,
							Name:           newTemplateName,
							StorageRequest: storageSize,
							FsType:         *stConfigModel.Fs,
							ReplFactor:     *stConfigModel.Repl,
							Provisioner:    *stConfigModel.Provisioner,
							Secure:         false,
							VolGroups:      false,
						}
						stConfigModelUpdated, resConfigModelUpdated, err = controlPlane.CreateCustomResourceTemplate(tenantID, updatedTemplateConfig)
						log.FailOnError(err, "Unable to update template")
						log.InfoD("Successfully updated the template with ID- %v", resConfigModelUpdated.GetId())
						newResourceTemplateID = resConfigModelUpdated.GetId()
						newStorageTemplateID = stConfigModelUpdated.GetId()
					})
					stepLog = "Apply updated template to the dataservice deployment"
					Step(stepLog, func() {

						if appConfigID == "" {
							appConfigID, err = controlPlane.GetAppConfTemplate(tenantID, ds.Name)
							log.FailOnError(err, "Error while fetching AppConfigID")
						}

						updatedDeployment, err = dsTest.UpdateDataServices(deployment.GetId(),
							appConfigID, deployment.GetImageId(),
							int32(ds.Replicas), newResourceTemplateID, params.InfraToTest.Namespace)
						log.FailOnError(err, "Error while updating dataservices")
						Step("Validate Deployments after template update", func() {
							err = dsTest.ValidateDataServiceDeployment(updatedDeployment, namespace)
							log.FailOnError(err, "Error while validating dataservices")
							log.InfoD("Data-service: %v is up and healthy", ds.Name)
							updatedDepList = append(updatedDepList, updatedDeployment)
							updatedPvcSize, err = GetVolumeCapacityInGB(namespace, updatedDeployment)
							log.InfoD("Updated Storage Size is- %v", updatedPvcSize)
						})
						stepLog = "Verify storage size before and after storage resize - Verify at STS, PV,PVC level"
						Step(stepLog, func() {

							_, _, config, err := pdslib.ValidateDataServiceVolumes(updatedDeployment, ds.Name, newResourceTemplateID, newStorageTemplateID, params.InfraToTest.Namespace)
							log.FailOnError(err, "error on ValidateDataServiceVolumes method")
							log.InfoD("resConfigModel.StorageRequest val is- %v and updated config val is- %v", *resConfigModelUpdated.StorageRequest, config.Spec.Resources.Requests.Storage)
							dash.VerifyFatal(config.Spec.Resources.Requests.Storage, *resConfigModelUpdated.StorageRequest, "Validating the storage size is updated in the config post resize (STS-LEVEL)")

							if updatedPvcSize > initialCapacity {
								flag := true
								dash.VerifyFatal(flag, true, "Validating the storage size is updated in the config post resize (PV/PVC-LEVEL)")
								log.InfoD("Initial PVC Capacity is- %v and Updated PVC Capacity is- %v", initialCapacity, updatedPvcSize)
							} else {
								log.FailOnError(err, "Failed to verify Storage Resize at PV/PVC level")
							}
						})
						stepLog = "Delete created templates"
						Step(stepLog, func() {
							controlPlane.CleanupCustomTemplates(newStorageTemplateID, newResourceTemplateID)
						})
						Step("Delete Deployments", func() {
							CleanupDeployments(deploymentsToBeCleaned)
							controlPlane.CleanupCustomTemplates(stConfigModel.GetId(), resConfigModel.GetId())
						})

					})

				}

			}

		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		err := bkpClient.AWSStorageClient.DeleteBucket()
		log.FailOnError(err, "Failed while deleting the bucket")
	})
})

var _ = Describe("{PerformStorageResizeOnAllDS}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("PerformStorageResizeOnAllDS", "Perform Storage Resize on all supported Dataservice one by one", pdsLabels, 0)
		wkloadParams = pdsdriver.LoadGenParams{
			LoadGenDepName: params.LoadGen.LoadGenDepName,
			Namespace:      params.InfraToTest.Namespace,
			NumOfRows:      params.LoadGen.NumOfRows,
			Timeout:        params.LoadGen.Timeout,
			Replicas:       params.LoadGen.Replicas,
			TableName:      params.LoadGen.TableName,
			Iterations:     params.LoadGen.Iterations,
			FailOnError:    params.LoadGen.FailOnError,
		}
		Step("Get All Supported Dataservices and Versions", func() {
			supportedDataServicesNameIDMap = pdslib.GetAllSupportedDataServices()
			for dsName := range supportedDataServicesNameIDMap {
				supportedDataServices = append(supportedDataServices, dsName)
			}
			for index := range supportedDataServices {
				log.Infof("supported data service %v ", supportedDataServices[index])
			}
			Step("Get the resource and app config template for supported dataservice", func() {
				dataServiceDefaultResourceTemplateIDMap, dataServiceNameIDMap, err = pdslib.GetAllDataserviceResourceTemplate(tenantID, supportedDataServices)
				Expect(err).NotTo(HaveOccurred())
				Expect(dataServiceDefaultResourceTemplateIDMap).NotTo(BeEmpty())
				Expect(dataServiceNameIDMap).NotTo(BeEmpty())

				dataServiceNameDefaultAppConfigMap, err = pdslib.GetAllDataServiceAppConfTemplate(tenantID, dataServiceNameIDMap)
				Expect(err).NotTo(HaveOccurred())
				Expect(dataServiceNameDefaultAppConfigMap).NotTo(BeEmpty())
			})
		})
	})

	It("Deploy All SupportedDataServices", func() {

		Step("Deploy All Supported Data Services", func() {
			var (
				updatedDeployment     *pds.ModelsDeployment
				updatedDepList        []*pds.ModelsDeployment
				resConfigModel        *pds.ModelsResourceSettingsTemplate
				stConfigModel         *pds.ModelsStorageOptionsTemplate
				newResourceTemplateID string
				newStorageTemplateID  string
				appConfigID           string
				updatedPvcSize        string
			)
			var generateWorkloads = make(map[string]string)
			replicas = 3
			log.InfoD("Deploying All Supported DataService")
			deployments, _, _, err := pdslib.DeployAllDataServices(supportedDataServicesNameIDMap, projectID,
				deploymentTargetID,
				dnsZone,
				deploymentName,
				namespaceID,
				dataServiceNameDefaultAppConfigMap,
				replicas,
				serviceType,
				dataServiceDefaultResourceTemplateIDMap,
				storageTemplateID,
				namespace,
			)
			Expect(err).NotTo(HaveOccurred())
			Step("Validate Storage Configurations", func() {
				for ds, deployment := range deployments {
					for index := range deployment {
						deploymentsToBeCleaned := []*pds.ModelsDeployment{}
						wlDeploymentsToBeCleaned := []*v1.Deployment{}
						log.Infof("data service deployed %v ", ds)
						resourceTemp, storageOp, config, err := pdslib.ValidateAllDataServiceVolumes(deployment[index], ds, dataServiceDefaultResourceTemplateIDMap, storageTemplateID)
						Expect(err).NotTo(HaveOccurred())
						log.Infof("filesystem used %v ", config.Spec.StorageOptions.Filesystem)
						log.Infof("storage replicas used %v ", config.Spec.StorageOptions.Replicas)
						log.Infof("cpu requests used %v ", config.Spec.Resources.Requests.CPU)
						log.Infof("memory requests used %v ", config.Spec.Resources.Requests.Memory)
						log.Infof("storage requests used %v ", config.Spec.Resources.Requests.Storage)
						log.Infof("No of nodes requested %v ", config.Spec.Nodes)
						log.Infof("volume group %v ", storageOp.VolumeGroup)

						Expect(resourceTemp.Resources.Requests.CPU).Should(Equal(config.Spec.Resources.Requests.CPU))
						Expect(resourceTemp.Resources.Requests.Memory).Should(Equal(config.Spec.Resources.Requests.Memory))
						Expect(resourceTemp.Resources.Requests.Storage).Should(Equal(config.Spec.Resources.Requests.Storage))
						Expect(resourceTemp.Resources.Limits.CPU).Should(Equal(config.Spec.Resources.Limits.CPU))
						Expect(resourceTemp.Resources.Limits.Memory).Should(Equal(config.Spec.Resources.Limits.Memory))
						repl, err := strconv.Atoi(config.Spec.StorageOptions.Replicas)
						Expect(err).NotTo(HaveOccurred())
						Expect(storageOp.Replicas).Should(Equal(int32(repl)))
						Expect(storageOp.Filesystem).Should(Equal(config.Spec.StorageOptions.Filesystem))
						Expect(config.Spec.Nodes).Should(Equal(replicas))

						initialCapacity := resourceTemp.Resources.Requests.Storage
						deploymentsToBeCleaned = append(deploymentsToBeCleaned, deployment[index])
						Step("Trigger Workloads on the dataservice", func() {

							log.InfoD("Running Workloads on DataService %v ", *deployment[index].Name)
							_, wlDep, err := dsTest.InsertDataAndReturnChecksum(deployment[index], wkloadParams)
							log.FailOnError(err, "Error while genearating workloads")
							generateWorkloads[*deployment[index].Name] = wlDep.Name
							for dsName, workloadContainer := range generateWorkloads {
								log.Debugf("dsName %s, workloadContainer %s", dsName, workloadContainer)
							}
							wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDep)
						})
						defer func() {
							for dsName, workloadContainer := range generateWorkloads {
								Step("Delete the workload generating deployments", func() {
									if Contains(dataServiceDeploymentWorkloads, dsName) {
										log.InfoD("Deleting Workload Generating deployment %v ", workloadContainer)
										err = pdslib.DeleteK8sDeployments(workloadContainer, namespace)
									} else if Contains(dataServicePodWorkloads, dsName) {
										log.InfoD("Deleting Workload Generating pod %v ", workloadContainer)
										err = pdslib.DeleteK8sPods(workloadContainer, namespace)
									}
									log.FailOnError(err, "error deleting workload generating pods")
								})
							}
						}()
						Step("Update the resource/storage template with increased storage size", func() {
							dataserviceID, _ := dsTest.GetDataServiceID(deployment[index].GetName())
							newTemplateName := "autoTemp-" + strconv.Itoa(rand.Int())
							updatedTemplateConfig := controlplane.Templates{
								CpuLimit:       resourceTemp.Resources.Limits.CPU,
								CpuRequest:     resourceTemp.Resources.Requests.CPU,
								DataServiceID:  dataserviceID,
								MemoryLimit:    resourceTemp.Resources.Limits.Memory,
								MemoryRequest:  resourceTemp.Resources.Requests.Memory,
								Name:           newTemplateName,
								StorageRequest: "100G",
								FsType:         config.Spec.StorageOptions.Filesystem,
								ReplFactor:     replicas,
								Provisioner:    config.Spec.StorageClass.Provisioner,
								Secure:         false,
								VolGroups:      false,
							}
							stConfigModel, resConfigModel, err = controlPlane.CreateCustomResourceTemplate(tenantID, updatedTemplateConfig)
							log.FailOnError(err, "Unable to update template")
							log.InfoD("Successfully updated the template with ID- %v", resConfigModel.GetId())
							newResourceTemplateID = resConfigModel.GetId()
							newStorageTemplateID = stConfigModel.GetId()
						})
						Step("Apply updated template to the dataservice deployment", func() {
							log.InfoD("Apply updated template to the dataservice deployment")
							if appConfigID == "" {
								appConfigID, err = controlPlane.GetAppConfTemplate(tenantID, deployment[index].GetName())
								log.FailOnError(err, "Error while fetching AppConfigID")
							}
							updatedDeployment, err = dsTest.UpdateDataServices(deployment[index].GetId(),
								appConfigID, deployment[index].GetImageId(),
								int32(replicas), newResourceTemplateID, params.InfraToTest.Namespace)
							log.FailOnError(err, "Error while updating dataservices")
							Step("Validate Deployments after template update", func() {
								err = dsTest.ValidateDataServiceDeployment(updatedDeployment, namespace)
								log.FailOnError(err, "Error while validating dataservices")
								log.InfoD("Data-service: %v is up and healthy", deployment[index].GetName())
								updatedDepList = append(updatedDepList, updatedDeployment)
								updatedPvc, err := GetVolumeCapacityInGB(namespace, updatedDeployment)
								log.FailOnError(err, "Error while validating getting updated PVC Size")
								updatedPvcSize = strconv.Itoa(int(updatedPvc))
								log.InfoD("Updated Storage Size is- %v", updatedPvcSize)

							})
							deploymentsToBeCleaned = append(deploymentsToBeCleaned, updatedDeployment)
							Step("Verify storage size before and after storage resize - Verify at STS, PV,PVC level", func() {
								log.InfoD("Verify storage size before and after storage resize - Verify at STS, PV,PVC level")
								_, _, config, err := pdslib.ValidateDataServiceVolumes(updatedDeployment, updatedDeployment.GetName(), newResourceTemplateID, newStorageTemplateID, params.InfraToTest.Namespace)
								log.FailOnError(err, "error on ValidateDataServiceVolumes method")
								log.InfoD("resConfigModel.StorageRequest val is- %v and updated config val is- %v", *resConfigModel.StorageRequest, config.Spec.Resources.Requests.Storage)
								dash.VerifyFatal(config.Spec.Resources.Requests.Storage, *resConfigModel.StorageRequest, "Validating the storage size is updated in the config post resize (STS-LEVEL)")
								dash.VerifyFatal(config.Spec.StorageOptions.Filesystem, *stConfigModel.Fs, "Validating the File System Type post storage resize (FileSystem-LEVEL)")
								stringRelFactor := strconv.Itoa(int(*stConfigModel.Repl))
								dash.VerifyFatal(config.Spec.StorageOptions.Replicas, stringRelFactor, "Validating the Replication Factor count post storage resize (RepelFactor-LEVEL)")
								if updatedPvcSize > initialCapacity {
									flag := true
									dash.VerifyFatal(flag, true, "Validating the storage size is updated in the config post resize (PV/PVC-LEVEL)")
									log.InfoD("Initial PVC Capacity is- %v and Updated PVC Capacity is- %v", initialCapacity, updatedPvcSize)
								} else {
									log.FailOnError(err, "Failed to verify Storage Resize at PV/PVC level")
								}
							})

						})
						Step("Delete created templates", func() {
							controlPlane.CleanupCustomTemplates(newStorageTemplateID, newResourceTemplateID)
						})
						Step("Delete Deployments", func() {
							CleanupDeployments(deploymentsToBeCleaned)
							controlPlane.CleanupCustomTemplates(stConfigModel.GetId(), resConfigModel.GetId())
						})
					}
				}
			})
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

func DeployDSWithCustomTemplatesRunWorkloads(ds PDSDataService, tenantId string, templates controlplane.Templates) (*pds.ModelsDeployment, uint64, *pds.ModelsResourceSettingsTemplate, *pds.ModelsStorageOptionsTemplate, string, *v1.Deployment, map[string]string, error) {
	var (
		dsVersions             = make(map[string]map[string][]string)
		depList                []*pds.ModelsDeployment
		initialCapacity        uint64
		dataServiceAppConfigID string
		workloadDep            *v1.Deployment
		stConfigModel          *pds.ModelsStorageOptionsTemplate
		resConfigModel         *pds.ModelsResourceSettingsTemplate
	)

	pdsdeploymentsmd5Hash := make(map[string]string)
	cusTempName := "autoTemp-" + strconv.Itoa(rand.Int())

	dataserviceID, _ := dsTest.GetDataServiceID(ds.Name)
	stConfigModel, resConfigModel, err = controlPlane.CreateCustomResourceTemplate(tenantId, controlplane.Templates{
		CpuLimit:       templates.CpuLimit,
		CpuRequest:     templates.CpuRequest,
		DataServiceID:  dataserviceID,
		MemoryLimit:    templates.MemoryLimit,
		MemoryRequest:  templates.MemoryRequest,
		Name:           cusTempName,
		StorageRequest: templates.StorageRequest,
		FsType:         templates.FsType,
		ReplFactor:     templates.ReplFactor,
		Provisioner:    templates.Provisioner,
		Secure:         templates.Secure,
		VolGroups:      templates.VolGroups})
	log.FailOnError(err, "Unable to create custom templates")
	customStorageTemplateID := stConfigModel.GetId()
	log.InfoD("created storageTemplateName is- %v and ID fetched is- %v ", *stConfigModel.Name, customStorageTemplateID)
	customResourceTemplateID := resConfigModel.GetId()
	log.InfoD("created resourceTemplateName is- %v and ID fetched is- %v ", *resConfigModel.Name, customResourceTemplateID)

	dataServiceAppConfigID, err = controlPlane.GetAppConfTemplate(tenantID, ds.Name)
	log.FailOnError(err, "error while getting app configuration template")
	log.InfoD("ds App config ID is- %v ", dataServiceAppConfigID)

	deploymentsToBeCleaned = []*pds.ModelsDeployment{}
	isDeploymentsDeleted = false
	log.InfoD("Starting to deploy DataService- %v", ds.Name)
	deployment, _, dataServiceVersionBuildMap, err = dsTest.DeployDS(ds.Name, projectID,
		deploymentTargetID,
		dnsZone,
		deploymentName,
		namespaceID,
		dataServiceAppConfigID,
		int32(ds.Replicas),
		serviceType,
		customResourceTemplateID,
		customStorageTemplateID,
		ds.Version,
		ds.Image,
		namespace)
	log.FailOnError(err, "Error while deploying data services")
	err = dsTest.ValidateDataServiceDeployment(deployment, namespace)
	log.FailOnError(err, "Error while validating dataservices")
	log.InfoD("Data-service: %v is up and healthy", ds.Name)
	depList = append(depList, deployment)
	dsVersions[ds.Name] = dataServiceVersionBuildMap
	log.FailOnError(err, "Error while deploying data services")
	initialCapacity, err = GetVolumeCapacityInGB(namespace, deployment)
	log.FailOnError(err, "Error while fetching pvc size for the ds")
	log.InfoD("Initial volume storage size is : %v", initialCapacity)
	dsEntity = restoreBkp.DSEntity{
		Deployment: deployment,
	}
	CleanMapEntries(pdsdeploymentsmd5Hash)
	ckSum, wlDep, err := dsTest.InsertDataAndReturnChecksum(deployment, wkloadParams)
	workloadDep = wlDep
	log.FailOnError(err, "Error while Running workloads")
	log.Debugf("Checksum for the deployment %s is %s", *deployment.ClusterResourceName, ckSum)
	pdsdeploymentsmd5Hash[*deployment.ClusterResourceName] = ckSum

	return deployment, initialCapacity, resConfigModel, stConfigModel, dataServiceAppConfigID, workloadDep, pdsdeploymentsmd5Hash, nil
}
