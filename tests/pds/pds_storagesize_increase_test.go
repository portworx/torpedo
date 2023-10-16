package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	pdsdriver "github.com/portworx/torpedo/drivers/pds"
	"github.com/portworx/torpedo/drivers/pds/controlplane"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	restoreBkp "github.com/portworx/torpedo/drivers/pds/pdsrestore"
	tc "github.com/portworx/torpedo/drivers/pds/targetcluster"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	v1 "k8s.io/api/apps/v1"
	"math/rand"
	"strconv"
	"time"
)

var _ = Describe("{ResizeStorageAndValidateTemplate}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("ResizeStorageAndValidateTemplate", "Perform PVC Resize and validate the updated vol in the storage config.", pdsLabels, 0)
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
			resConfigModelUpdated    *pds.ModelsResourceSettingsTemplate
			stConfigModelUpdated     *pds.ModelsStorageOptionsTemplate
			newResourceTemplateID    string
			newStorageTemplateID     string
			//customResourceTemplateID string
			//customStorageTemplateID  string
			updatedPvcSize uint64
		)
		restoredDeploymentsmd5Hash := make(map[string]string)
		stepLog := "Create Custom Templates , Deploy ds and Trigger Workload"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			depList, deployments, initialCapacity, resConfigModel, stConfigModel, appConfigID, workloadDep, pdsdeploymentsmd5Hash, err := DeployDSWithCustomTemplatesRunWorkloads(tenantID, controlplane.Templates{
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
			//customResourceTemplateID = resConfigModel.GetId()
			//customStorageTemplateID = stConfigModel.GetId()
			for _, ds := range params.DataServiceToTest {
				dataserviceID, _ := dsTest.GetDataServiceID(ds.Name)
				stepLog = "Check PVC for full condition based upto 90% full"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					ctx, err := Inst().Pds.CreateSchedulerContextForPDSApps(depList)
					log.FailOnError(err, "Unable to create scheduler context")
					err = CheckPVCtoFullCondition(ctx)
					log.FailOnError(err, "Failing while filling the PVC to 90 percentage of its capacity due to ...")

				})
				stepLog = "Update the resource/storage template with increased storage size"
				Step(stepLog, func() {
					newTemplateName := "autoTemp-" + strconv.Itoa(rand.Int())
					updatedTemplateConfig := controlplane.Templates{
						CpuLimit:       "4",
						CpuRequest:     "2",
						DataServiceID:  dataserviceID,
						MemoryLimit:    "6G",
						MemoryRequest:  "4G",
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
					for ds, deployment := range deployments {
						if appConfigID == "" {
							appConfigID, err = controlPlane.GetAppConfTemplate(tenantID, ds.Name)
							log.FailOnError(err, "Error while fetching AppConfigID")
						}
						log.InfoD("deployment.GetId()- %v", deployment.GetId())
						log.InfoD("dataServiceDefaultAppConfigID- %v", appConfigID)
						log.InfoD("deployment.GetImageId()- %v", deployment.GetImageId())
						log.InfoD("updateTemplateID- %v", newResourceTemplateID)
						updatedDeployment, err = dsTest.UpdateDataServices(deployment.GetId(),
							appConfigID, deployment.GetImageId(),
							int32(ds.Replicas), newResourceTemplateID, params.InfraToTest.Namespace)
						log.FailOnError(err, "Error while updating dataservices")
						Step("Validate Deployments after template update", func() {
							err = dsTest.ValidateDataServiceDeployment(updatedDeployment, namespace)
							log.FailOnError(err, "Error while validating dataservices")
							log.InfoD("Data-service: %v is up and healthy", ds.Name)
							updatedDepList = append(updatedDepList, updatedDeployment)
							ctx, err := Inst().Pds.CreateSchedulerContextForPDSApps(updatedDepList)
							log.FailOnError(err, "Unable to create scheduler context")
							updatedPvcSize, err = GetVolumeCapacityInGB(ctx)
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
							dash.VerifyFatal(*resConfigModelUpdated.StorageRequest, config.Spec.Resources.Requests.Storage, "Validating the storage size is updated in the config post resize (STS-LEVEL)")

							if updatedPvcSize > initialCapacity {
								flag := true
								dash.VerifyFatal("PV/Pvc Resize is successful", flag, "Validating the storage size is updated in the config post resize (PV/PVC-LEVEL)")
							} else {
								log.FailOnError(err, "Failed to verify Storage Resize at PV/PVC level")
							}
						})
						stepLog = "Perform backup after PVC Resize"
						Step(stepLog, func() {
							log.InfoD(stepLog)
							log.Infof("Deployment ID: %v, backup target ID: %v", deployment.GetId(), bkpTarget.GetId())
							err = bkpClient.TriggerAndValidateAdhocBackup(deployment.GetId(), bkpTarget.GetId(), "s3")
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
								Deployment:           deployment,
								RestoreTargetCluster: restoreTarget,
							}
							backupJobs, err := restoreClient.Components.BackupJob.ListBackupJobsBelongToDeployment(projectID, deployment.GetId())
							log.FailOnError(err, "Error while fetching the backup jobs for the deployment: %v", deployment.GetClusterResourceName())
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
			}

			Step("Delete Deployments", func() {
				CleanupDeployments(deploymentsToBeCleaned)
				controlPlane.CleanupCustomTemplates(stConfigModel.GetId(), resConfigModel.GetId())
				controlPlane.CleanupCustomTemplates(newStorageTemplateID, newResourceTemplateID)
			})
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
		err := bkpClient.AWSStorageClient.DeleteBucket()
		log.FailOnError(err, "Failed while deleting the bucket")
	})
})

func DeployDSWithCustomTemplatesRunWorkloads(tenantId string, templates controlplane.Templates) ([]*pds.ModelsDeployment, map[PDSDataService]*pds.ModelsDeployment, uint64, *pds.ModelsResourceSettingsTemplate, *pds.ModelsStorageOptionsTemplate, string, *v1.Deployment, map[string]string, error) {
	var (
		wlDeploymentsToBeCleaned []*v1.Deployment
		deployments              = make(map[PDSDataService]*pds.ModelsDeployment)
		dsVersions               = make(map[string]map[string][]string)
		depList                  []*pds.ModelsDeployment
		initialCapacity          uint64
		dataServiceAppConfigID   string
		workloadDep              *v1.Deployment
		stConfigModel            *pds.ModelsStorageOptionsTemplate
		resConfigModel           *pds.ModelsResourceSettingsTemplate
	)
	pdsdeploymentsmd5Hash := make(map[string]string)
	cusTempName := "autoTemp-" + strconv.Itoa(rand.Int())
	for _, ds := range params.DataServiceToTest {
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
		deployments[ds] = deployment
		depList = append(depList, deployment)
		dsVersions[ds.Name] = dataServiceVersionBuildMap
		ctx, err := Inst().Pds.CreateSchedulerContextForPDSApps(depList)
		log.FailOnError(err, "Error while deploying data services")
		initialCapacity, err = GetVolumeCapacityInGB(ctx)
		log.FailOnError(err, "Error while fetching pvc size for the ds")
		log.InfoD("Initial volume storage size is : %v", initialCapacity)
		dsEntity = restoreBkp.DSEntity{
			Deployment: deployment,
		}
		ckSum, wlDep, err := dsTest.InsertDataAndReturnChecksum(deployment, wkloadParams)
		wlDeploymentsToBeCleaned := append(wlDeploymentsToBeCleaned, wlDep)
		workloadDep = wlDep
		log.FailOnError(err, "Error while Running workloads")
		log.Debugf("Checksum for the deployment %s is %s", *deployment.ClusterResourceName, ckSum)
		pdsdeploymentsmd5Hash[*deployment.ClusterResourceName] = ckSum
		wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDeploymentsToBeCleaned...)
	}
	return depList, deployments, initialCapacity, resConfigModel, stConfigModel, dataServiceAppConfigID, workloadDep, pdsdeploymentsmd5Hash, nil
}
