package tests

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	"github.com/portworx/torpedo/drivers/node"
	pdsdriver "github.com/portworx/torpedo/drivers/pds"
	"github.com/portworx/torpedo/drivers/pds/controlplane"
	dss "github.com/portworx/torpedo/drivers/pds/dataservice"
	pdslib "github.com/portworx/torpedo/drivers/pds/lib"
	restoreBkp "github.com/portworx/torpedo/drivers/pds/pdsrestore"
	tc "github.com/portworx/torpedo/drivers/pds/targetcluster"
	"github.com/portworx/torpedo/pkg/log"
	. "github.com/portworx/torpedo/tests"
	v1 "k8s.io/api/apps/v1"
	"math/rand"
	"net/http"
	"strconv"
)

var _ = Describe("{RestartPXDuringAppScaleUp}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("RestartPXDuringAppScaleUp", "Restart PX on a node during application is scaled up", pdsLabels, 0)
		pdslib.MarkResiliencyTC(true)
	})

	It("Deploy Dataservices and Restart PX During App scaleup", func() {
		var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
		var generateWorkloads = make(map[string]string)

		Step("Deploy Data Services", func() {
			for _, ds := range params.DataServiceToTest {
				Step("Deploy and validate data service", func() {
					isDeploymentsDeleted = false
					deployment, _, _, err = DeployandValidateDataServices(ds, params.InfraToTest.Namespace, tenantID, projectID)
					log.FailOnError(err, "Error while deploying data services")
					deployments[ds] = deployment
				})
			}
		})

		defer func() {
			for _, newDeployment := range deployments {
				Step("Delete created deployments")
				resp, err := pdslib.DeleteDeployment(newDeployment.GetId())
				log.FailOnError(err, "Error while deleting data services")
				dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
				err = pdslib.DeletePvandPVCs(*newDeployment.ClusterResourceName, false)
				log.FailOnError(err, "Error while deleting PV and PVCs")
			}
		}()

		Step("Scale Up Data Services and Restart Portworx", func() {
			for ds, deployment := range deployments {

				failuretype := pdslib.TypeOfFailure{
					Type: RestartPxDuringDSScaleUp,
					Method: func() error {
						return pdslib.RestartPXDuringDSScaleUp(params.InfraToTest.Namespace, deployment)
					},
				}
				pdslib.DefineFailureType(failuretype)

				log.InfoD("Scaling up DataService %v ", ds.Name)
				dataServiceDefaultAppConfigID, err = controlPlane.GetAppConfTemplate(tenantID, ds.Name)
				log.FailOnError(err, "Error while getting app configuration template")
				dash.VerifyFatal(dataServiceDefaultAppConfigID != "", true, "Validating dataServiceDefaultAppConfigID")

				dataServiceDefaultResourceTemplateID, err = controlPlane.GetResourceTemplate(tenantID, ds.Name)
				log.FailOnError(err, "Error while getting resource setting template")
				dash.VerifyFatal(dataServiceDefaultResourceTemplateID != "", true, "Validating dataServiceDefaultResourceTemplateID")

				updatedDeployment, err := dsTest.UpdateDataServices(deployment.GetId(),
					dataServiceDefaultAppConfigID, deployment.GetImageId(),
					int32(ds.ScaleReplicas), dataServiceDefaultResourceTemplateID, namespace)
				log.FailOnError(err, "Error while updating dataservices")

				//wait for the scaled up data service and restart px
				err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, int32(ds.ScaleReplicas))
				log.FailOnError(err, fmt.Sprintf("Error happened while restarting px for data service %v", *deployment.ClusterResourceName))

				id := pdslib.GetDataServiceID(ds.Name)
				dash.VerifyFatal(id != "", true, "Validating dataservice id")
				log.Infof("Getting versionID  for Data service version %s and buildID for %s ", ds.Version, ds.Image)

				_, _, dsVersionBuildMap, err := pdslib.GetVersionsImage(ds.Version, ds.Image, id)
				log.FailOnError(err, "Error while fetching versions/image information")

				//TODO: Rename the method ValidateDataServiceVolumes
				resourceTemp, storageOp, config, err := pdslib.ValidateDataServiceVolumes(updatedDeployment, ds.Name, dataServiceDefaultResourceTemplateID, storageTemplateID, namespace)
				log.FailOnError(err, "error on ValidateDataServiceVolumes method")
				ValidateDeployments(resourceTemp, storageOp, config, int(ds.ScaleReplicas), dsVersionBuildMap)
				dash.VerifyFatal(int32(ds.ScaleReplicas), config.Replicas, "Validating replicas after scaling up of dataservice")

			}
		})
		Step("Running Workloads", func() {
			for ds, deployment := range deployments {
				if Contains(dataServicePodWorkloads, ds.Name) || Contains(dataServiceDeploymentWorkloads, ds.Name) {
					log.InfoD("Running Workloads on DataService %v ", ds.Name)
					var params pdslib.WorkloadGenerationParams
					pod, dep, err = RunWorkloads(params, ds, deployment, namespace)
					log.FailOnError(err, fmt.Sprintf("Error while genearating workloads for dataservice [%s]", ds.Name))
					if dep == nil {
						generateWorkloads[ds.Name] = pod.Name
					} else {
						generateWorkloads[ds.Name] = dep.Name
					}
					for dsName, workloadContainer := range generateWorkloads {
						log.Debugf("dsName %s, workloadContainer %s", dsName, workloadContainer)
					}
				} else {
					log.InfoD("Workload script not available for ds %v", ds.Name)
				}
			}
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
	})
	JustAfterEach(func() {
		EndTorpedoTest()

	})
})

var _ = Describe("{RebootNodeDuringAppResourceUpdate}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("RebootNodeDuringAppResourceUpdate", "Reboot active node during application resource Update, example increase the CPU/Mem limits ", pdsLabels, 0)
		pdslib.MarkResiliencyTC(true)
		//Initializing the parameters required for workload generation
		wkloadParams = pdsdriver.LoadGenParams{
			LoadGenDepName: params.LoadGen.LoadGenDepName,
			Namespace:      params.InfraToTest.Namespace,
			NumOfRows:      params.LoadGen.NumOfRows,
			Timeout:        params.LoadGen.Timeout,
			Replicas:       params.LoadGen.Replicas,
			TableName:      "wltestingnew",
			Iterations:     params.LoadGen.Iterations,
			FailOnError:    params.LoadGen.FailOnError,
		}
	})

	It("Deploy Dataservices and Restart PX During App scaleup", func() {
		var (
			scaledUpdatedDeployment      *pds.ModelsDeployment
			wlDeploymentsToBeCleaned     []*v1.Deployment
			deploymentsToBeCleaned       []*pds.ModelsDeployment
			depList                      []*pds.ModelsDeployment
			scaledUpdatedDepList         []*pds.ModelsDeployment
			stIds                        []string
			resIds                       []string
			resConfigModelUpdated        *pds.ModelsResourceSettingsTemplate
			stConfigModelUpdated         *pds.ModelsStorageOptionsTemplate
			newResourceTemplateID        string
			newStorageTemplateID         string
			increasedPvcSizeAfterScaleUp uint64
			beforeResizePodAge           float64
		)
		stepLog := "Create Custom Templates , Deploy ds and Trigger Workload"
		Step(stepLog, func() {
			pdsdeploymentsmd5HashAfterResize := make(map[string]string)
			for _, ds := range params.DataServiceToTest {
				log.InfoD(stepLog)
				CleanMapEntries(pdsdeploymentsmd5HashAfterResize)
				stIds, resIds = nil, nil
				depList, _, scaledUpdatedDepList, deploymentsToBeCleaned, restoredDeployments = []*pds.ModelsDeployment{}, []*pds.ModelsDeployment{}, []*pds.ModelsDeployment{}, []*pds.ModelsDeployment{}, []*pds.ModelsDeployment{}
				wlDeploymentsToBeCleaned = []*v1.Deployment{}
				deployment, _, resConfigModel, stConfigModel, appConfigID, _, _, _, err := DeployDSWithCustomTemplatesRunWorkloads(ds, tenantID, controlplane.Templates{
					CpuLimit:       params.StorageConfigurations.CpuLimit,
					CpuRequest:     params.StorageConfigurations.CpuRequest,
					MemoryLimit:    params.StorageConfigurations.MemoryLimit,
					MemoryRequest:  params.StorageConfigurations.MemoryRequest,
					StorageRequest: params.StorageConfigurations.StorageRequest,
					FsType:         "xfs",
					ReplFactor:     2,
					Provisioner:    "pxd.portworx.com",
					Secure:         false,
					VolGroups:      false,
				})
				stIds = append(stIds, stConfigModel.GetId())
				resIds = append(resIds, resConfigModel.GetId())
				depList = append(depList, deployment)
				deploymentsToBeCleaned = append(deploymentsToBeCleaned, deployment)
				dataserviceID, _ := dsTest.GetDataServiceID(ds.Name)
				beforeResizePodAge, err = GetPodAge(deployment, params.InfraToTest.Namespace)
				log.FailOnError(err, "unable to get pods AGE before Storage resize")
				log.InfoD("Pods Age before storage resize is- [%v]Min", beforeResizePodAge)

				stepLog = "Scale up the DS with increased storage size and Repl factor as 2 "
				Step(stepLog, func() {
					failuretype := pdslib.TypeOfFailure{
						Type: RebootNodeDuringAppResourceUpdate,
						Method: func() error {
							return pdslib.RebootActiveNodeDuringDeployment(params.InfraToTest.Namespace, deployment, 1)
						},
					}
					pdslib.DefineFailureType(failuretype)
					newTemplateName1 := "autoTemp-" + strconv.Itoa(rand.Int())
					newCpuLimit, _ := strconv.Atoi(*resConfigModel.CpuLimit)
					newCpuReq, _ := strconv.Atoi(*resConfigModel.CpuRequest)
					currMemoryLimit := *resConfigModel.StorageRequest
					newMemoryLimit, _ := strconv.Atoi(currMemoryLimit[:len(currMemoryLimit)-1])
					updatedMemoryLimit := strconv.Itoa(newMemoryLimit+2) + "G"
					currMemoryReq := *resConfigModel.StorageRequest
					newMemoryReq, _ := strconv.Atoi(currMemoryReq[:len(currMemoryReq)-1])
					updatedMemoryReq := strconv.Itoa(newMemoryReq+2) + "G"
					updatedTemplateConfig1 := controlplane.Templates{
						CpuLimit:       fmt.Sprint(newCpuLimit + 2),
						CpuRequest:     fmt.Sprint(newCpuReq + 2),
						DataServiceID:  dataserviceID,
						MemoryLimit:    updatedMemoryLimit,
						MemoryRequest:  updatedMemoryReq,
						Name:           newTemplateName1,
						StorageRequest: params.StorageConfigurations.NewStorageSize,
						FsType:         *stConfigModel.Fs,
						ReplFactor:     *stConfigModel.Repl,
						Provisioner:    *stConfigModel.Provisioner,
						Secure:         false,
						VolGroups:      false,
					}
					stConfigModelUpdated, resConfigModelUpdated, err = controlPlane.CreateCustomResourceTemplate(tenantID, updatedTemplateConfig1)
					log.FailOnError(err, "Unable to update template")
					log.InfoD("Successfully updated the template with ID- %v and ReplicationFactor- %v", resConfigModelUpdated.GetId(), updatedTemplateConfig1.ReplFactor)
					newResourceTemplateID = resConfigModelUpdated.GetId()
					newStorageTemplateID = stConfigModelUpdated.GetId()
					stIds = append(stIds, newStorageTemplateID)
					resIds = append(resIds, newResourceTemplateID)
				})
				stepLog = "Apply updated template to the dataservice deployment and Induce failure"
				Step(stepLog, func() {
					log.InfoD(stepLog)
					if appConfigID == "" {
						appConfigID, err = controlPlane.GetAppConfTemplate(tenantID, ds.Name)
						log.FailOnError(err, "Error while fetching AppConfigID")
					}
					scaledUpdatedDeployment, err = dsTest.UpdateDataServices(deployment.GetId(),
						appConfigID, deployment.GetImageId(),
						int32(ds.Replicas), newResourceTemplateID, params.InfraToTest.Namespace)
					log.FailOnError(err, "Error while updating dataservices")
					//wait for the scaled up data service and restart px
					err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, int32(ds.Replicas))
					log.FailOnError(err, fmt.Sprintf("Error happened while restarting px for data service %v", *deployment.ClusterResourceName))
				})
				Step("Validate Deployments after template update", func() {
					err = dsTest.ValidateDataServiceDeployment(scaledUpdatedDeployment, namespace)
					log.FailOnError(err, "Error while validating dataservices")
					log.InfoD("Data-service: %v is up and healthy", ds.Name)
					scaledUpdatedDepList = append(scaledUpdatedDepList, scaledUpdatedDeployment)
					increasedPvcSizeAfterScaleUp, err = GetVolumeCapacityInGB(namespace, scaledUpdatedDeployment)
					log.InfoD("Increased Storage Size after scale-up is- %v", increasedPvcSizeAfterScaleUp)
				})
				Step("Validate resource update is successful after active node reboots during resource update", func() {
					_, _, config, err := pdslib.ValidateDataServiceVolumes(scaledUpdatedDeployment, ds.Name, newResourceTemplateID, newStorageTemplateID, params.InfraToTest.Namespace)
					log.FailOnError(err, "error on ValidateDataServiceVolumes method")
					dash.VerifyFatal(config.Resources.Requests.CPU, *resConfigModelUpdated.CpuRequest, "Verifying increase in CPU request from original deployment")
					dash.VerifyFatal(config.Resources.Limits.CPU, *resConfigModelUpdated.CpuLimit, "Verifying increase in CPU Limits from original deployment")
					dash.VerifyFatal(config.Resources.Requests.Memory, *resConfigModelUpdated.MemoryRequest, "Verifying increase in MEM request from original deployment")
					dash.VerifyFatal(config.Resources.Limits.Memory, *resConfigModelUpdated.MemoryLimit, "Verifying increase in MEM limits from original deployment")
				})
				Step("Clean up workload deployments", func() {
					for _, wlDep := range wlDeploymentsToBeCleaned {
						err := k8sApps.DeleteDeployment(wlDep.Name, wlDep.Namespace)
						log.FailOnError(err, "Failed while deleting the workload deployment")
					}
				})

				Step("Delete Deployments", func() {
					CleanupDeployments(deploymentsToBeCleaned)
					err := controlPlane.CleanupCustomTemplates(stIds, resIds)
					log.FailOnError(err, "Failed to delete custom templates")
				})
			}
		})
	})
	JustAfterEach(func() {
		EndTorpedoTest()
	})
})

var _ = Describe("{KillPdsAgentPodDuringAppScaleUp}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("KillPdsAgentPodDuringAppScaleUp", "Kill PDS-Agent Pod during application is scaled up", pdsLabels, 0)
		pdslib.MarkResiliencyTC(true)
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

	It("Deploy Dataservices and Kill PDS Pod Agent while App scaleup", func() {
		var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
		var wlDeploymentsToBeCleaned []*v1.Deployment

		Step("Deploy Data Services", func() {
			for _, ds := range params.DataServiceToTest {
				Step("Deploy and validate data service", func() {
					isDeploymentsDeleted = false
					deployment, _, _, err = DeployandValidateDataServices(ds, params.InfraToTest.Namespace, tenantID, projectID)
					log.FailOnError(err, "Error while deploying data services")
					deployments[ds] = deployment
				})
			}
		})

		defer func() {
			for _, newDeployment := range deployments {
				Step("Delete created deployments")
				resp, err := pdslib.DeleteDeployment(newDeployment.GetId())
				log.FailOnError(err, "Error while deleting data services")
				dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
				err = pdslib.DeletePvandPVCs(*newDeployment.ClusterResourceName, false)
				log.FailOnError(err, "Error while deleting PV and PVCs")
			}
		}()

		Step("Scale Up Data Services and Restart Portworx", func() {
			for ds, deployment := range deployments {
				failuretype := pdslib.TypeOfFailure{
					Type: KillPdsAgentPodDuringAppScaleUp,
					Method: func() error {
						return pdslib.KillPodsInNamespace(params.InfraToTest.PDSNamespace, pdslib.PdsAgentPod)
					},
				}
				pdslib.DefineFailureType(failuretype)

				log.InfoD("Scaling up DataService %v ", ds.Name)
				dataServiceDefaultAppConfigID, err = controlPlane.GetAppConfTemplate(tenantID, ds.Name)
				log.FailOnError(err, "Error while getting app configuration template")
				dash.VerifyFatal(dataServiceDefaultAppConfigID != "", true, "Validating dataServiceDefaultAppConfigID")

				dataServiceDefaultResourceTemplateID, err = controlPlane.GetResourceTemplate(tenantID, ds.Name)
				log.FailOnError(err, "Error while getting resource setting template")
				dash.VerifyFatal(dataServiceDefaultResourceTemplateID != "", true, "Validating dataServiceDefaultResourceTemplateID")

				updatedDeployment, err := dsTest.UpdateDataServices(deployment.GetId(),
					dataServiceDefaultAppConfigID, deployment.GetImageId(),
					int32(ds.ScaleReplicas), dataServiceDefaultResourceTemplateID, namespace)
				log.FailOnError(err, "Error while updating dataservices")

				//wait for the scaled up data service and restart px
				err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, int32(ds.ScaleReplicas))
				log.FailOnError(err, fmt.Sprintf("Error happened while restarting px for data service %v", *deployment.ClusterResourceName))

				id := pdslib.GetDataServiceID(ds.Name)
				dash.VerifyFatal(id != "", true, "Validating dataservice id")
				log.Infof("Getting versionID  for Data service version %s and buildID for %s ", ds.Version, ds.Image)

				_, _, dsVersionBuildMap, err := pdslib.GetVersionsImage(ds.Version, ds.Image, id)
				log.FailOnError(err, "Error while fetching versions/image information")

				//TODO: Rename the method ValidateDataServiceVolumes
				resourceTemp, storageOp, config, err := pdslib.ValidateDataServiceVolumes(updatedDeployment, ds.Name, dataServiceDefaultResourceTemplateID, storageTemplateID, namespace)
				log.FailOnError(err, "error on ValidateDataServiceVolumes method")
				ValidateDeployments(resourceTemp, storageOp, config, ds.ScaleReplicas, dsVersionBuildMap)
				dash.VerifyFatal(ds.ScaleReplicas, config.Replicas, "Validating replicas after scaling up of dataservice")
			}
		})
		Step("Running Workloads", func() {
			for _, deployment := range deployments {
				ckSum2, wlDep, err := dsTest.InsertDataAndReturnChecksum(deployment, wkloadParams)
				log.FailOnError(err, "Error while Running workloads-%v", wlDep)
				log.Debugf("Checksum for the deployment %s is %s", *deployment.ClusterResourceName, ckSum2)
				wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDep)
			}
		})
		Step("Clean up workload deployments", func() {
			for _, wlDep := range wlDeploymentsToBeCleaned {
				err := k8sApps.DeleteDeployment(wlDep.Name, wlDep.Namespace)
				log.FailOnError(err, "Failed while deleting the workload deployment")
			}
		})
	})
	JustAfterEach(func() {
		EndTorpedoTest()
	})
})

var _ = Describe("{RebootActiveNodeDuringDeployment}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("RebootActiveNodeDuringDeployment", "Reboots a Node onto which a pod is coming up", pdsLabels, 0)
	})

	It("deploy Dataservices", func() {
		Step("Deploy Data Services", func() {
			for _, ds := range params.DataServiceToTest {
				var dsVersionBuildMap = make(map[string][]string)
				var num_reboots int
				num_reboots = 1
				Step("Start deployment, Reboot a node on which deployment is coming up and validate data service", func() {
					isDeploymentsDeleted = false
					// Global Resiliency TC marker
					pdslib.MarkResiliencyTC(true)

					// Deploy and Validate this Data service after injecting the type of failure we want to catch
					deployment, _, dsVersionBuildMap, err = dsTest.TriggerDeployDataService(ds, params.InfraToTest.Namespace, tenantID, projectID, false,
						dss.TestParams{NamespaceId: namespaceID, StorageTemplateId: storageTemplateID, DeploymentTargetId: deploymentTargetID, DnsZone: dnsZone, ServiceType: serviceType})
					log.FailOnError(err, "Error while deploying data services")

					// Type of failure that this TC needs to cover
					failuretype := pdslib.TypeOfFailure{
						Type: ActiveNodeRebootDuringDeployment,
						Method: func() error {
							return pdslib.RebootActiveNodeDuringDeployment(params.InfraToTest.Namespace, deployment, num_reboots)
						},
					}
					pdslib.DefineFailureType(failuretype)

					err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, params.ResiliencyTest.CheckTillReplica)
					log.FailOnError(err, fmt.Sprintf("Error happened while executing Reboot test for data service %v", *deployment.ClusterResourceName))

					dataServiceDefaultResourceTemplateID, err = controlPlane.GetResourceTemplate(tenantID, ds.Name)
					log.FailOnError(err, "Error while getting resource setting template")
					dash.VerifyFatal(dataServiceDefaultResourceTemplateID != "", true, "Validating dataServiceDefaultResourceTemplateID")

					resourceTemp, storageOp, config, err := pdslib.ValidateDataServiceVolumes(deployment, ds.Name, dataServiceDefaultResourceTemplateID, storageTemplateID, namespace)
					log.FailOnError(err, "error on ValidateDataServiceVolumes method")
					ValidateDeployments(resourceTemp, storageOp, config, int(ds.Replicas), dsVersionBuildMap)
				})
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()

		if !isDeploymentsDeleted {
			Step("Delete created deployments")
			resp, err := pdslib.DeleteDeployment(deployment.GetId())
			log.FailOnError(err, "Error while deleting data services")
			dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
		}
	})
})

var _ = Describe("{RebootNodeDuringAppVersionUpdate}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("RebootNodeDuringAppVersionUpdate", "Reboot node while app version update is going on", pdsLabels, 0)
		// Global Resiliency TC marker
		pdslib.MarkResiliencyTC(true)
	})

	It("Reboot Node While App Version update is going on", func() {
		var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
		var generateWorkloads = make(map[string]string)
		var dsVersionBuildMap = make(map[string][]string)

		Step("Deploy and Validate Data services", func() {
			for _, ds := range params.DataServiceToTest {
				deployment, _, dsVersionBuildMap, err = dsTest.TriggerDeployDataService(ds, params.InfraToTest.Namespace, tenantID, projectID, true,
					dss.TestParams{NamespaceId: namespaceID, StorageTemplateId: storageTemplateID, DeploymentTargetId: deploymentTargetID, DnsZone: dnsZone, ServiceType: serviceType})
				log.FailOnError(err, "Error while deploying data services")

				err = dsTest.ValidateDataServiceDeployment(deployment, params.InfraToTest.Namespace)
				log.FailOnError(err, "Error while validating data service deployment")
				deployments[ds] = deployment

				dataServiceDefaultResourceTemplateID, err = controlPlane.GetResourceTemplate(tenantID, ds.Name)
				log.FailOnError(err, "Error while getting resource template")
				log.InfoD("dataServiceDefaultResourceTemplateID %v ", dataServiceDefaultResourceTemplateID)

				resourceTemp, storageOp, config, err := pdslib.ValidateDataServiceVolumes(deployment, ds.Name, dataServiceDefaultResourceTemplateID, storageTemplateID, namespace)
				log.FailOnError(err, "error on ValidateDataServiceVolumes method")
				ValidateDeployments(resourceTemp, storageOp, config, ds.Replicas, dsVersionBuildMap)
			}
		})

		defer func() {
			for _, newDeployment := range deployments {
				Step("Delete created deployments")
				resp, err := pdslib.DeleteDeployment(newDeployment.GetId())
				log.FailOnError(err, "Error while deleting data services")
				dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
				err = pdslib.DeletePvandPVCs(*newDeployment.ClusterResourceName, false)
				log.FailOnError(err, "Error while deleting PV and PVCs")
			}
		}()

		Step("Update Data Service Version and reboot node", func() {
			for ds, deployment := range deployments {

				log.Infof("Version/Build: %v %v", ds.Version, ds.Image)
				updatedDeployment, err := pdslib.UpdateDataServiceVerison(deployment.GetDataServiceId(), deployment.GetId(),
					dataServiceDefaultAppConfigID,
					int32(ds.Replicas), dataServiceDefaultResourceTemplateID, ds.Image, ds.Version)
				log.FailOnError(err, "Error while updating data services")
				log.InfoD("data service %v update triggered", ds.Name)

				// Type of failure that this TC needs to cover
				failuretype := pdslib.TypeOfFailure{
					Type: RebootNodeDuringAppVersionUpdate,
					Method: func() error {
						return pdslib.NodeRebootDurinAppVersionUpdate(params.InfraToTest.Namespace, deployment)
					},
				}
				pdslib.DefineFailureType(failuretype)

				err = pdslib.InduceFailureAfterWaitingForCondition(updatedDeployment, namespace, params.ResiliencyTest.CheckTillReplica)
				log.FailOnError(err, fmt.Sprintf("Error happened while executing Reboot test for data service %v", *deployment.ClusterResourceName))

				dataServiceDefaultResourceTemplateID, err = controlPlane.GetResourceTemplate(tenantID, ds.Name)
				log.FailOnError(err, "Error while getting resource template")
				log.InfoD("dataServiceDefaultResourceTemplateID %v ", dataServiceDefaultResourceTemplateID)

				resourceTemp, storageOp, config, err := pdslib.ValidateDataServiceVolumes(updatedDeployment, ds.Name, dataServiceDefaultResourceTemplateID, storageTemplateID, namespace)
				log.FailOnError(err, "error on ValidateDataServiceVolumes method")

				id := pdslib.GetDataServiceID(ds.Name)
				dash.VerifyFatal(id != "", true, "Validating dataservice id")
				log.Infof("Getting versionID  for Data service version %s and buildID for %s ", ds.Version, ds.Image)
				for version := range dsVersionBuildMap {
					delete(dsVersionBuildMap, version)
				}
				_, _, dsVersionBuildMap, err = pdslib.GetVersionsImage(ds.Version, ds.Image, id)
				log.FailOnError(err, "Error while fetching versions/image information")

				ValidateDeployments(resourceTemp, storageOp, config, int(ds.Replicas), dsVersionBuildMap)
				dash.VerifyFatal(config.Version, ds.Version+"-"+ds.Image, "validating ds build and version")
			}

		})

		Step("Running Workloads", func() {
			for ds, deployment := range deployments {
				if Contains(dataServicePodWorkloads, ds.Name) || Contains(dataServiceDeploymentWorkloads, ds.Name) {
					log.InfoD("Running Workloads on DataService %v ", ds.Name)
					var params pdslib.WorkloadGenerationParams
					pod, dep, err = RunWorkloads(params, ds, deployment, namespace)
					log.FailOnError(err, fmt.Sprintf("Error while genearating workloads for dataservice [%s]", ds.Name))
					if dep == nil {
						generateWorkloads[ds.Name] = pod.Name
					} else {
						generateWorkloads[ds.Name] = dep.Name
					}
					for dsName, workloadContainer := range generateWorkloads {
						log.Debugf("dsName %s, workloadContainer %s", dsName, workloadContainer)
					}
				} else {
					log.InfoD("Workload script not available for ds %v", ds.Name)
				}
			}
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
	})
	JustAfterEach(func() {
		EndTorpedoTest()
	})

})

var _ = Describe("{KillDeploymentControllerDuringDeployment}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("KillDeploymentControllerDuringDeployment", "Kill Deployment Controller Pod when a DS Deployment is happening", pdsLabels, 0)
	})

	It("Deploy Data Services", func() {
		Step("Deploy Data Services", func() {
			var dsVersionBuildMap = make(map[string][]string)
			for _, ds := range params.DataServiceToTest {
				Step("Start deployment, Kill Deployment Controller Pod while deployment is ongoing and validate data service", func() {
					isDeploymentsDeleted = false
					// Global Resiliency TC marker
					pdslib.MarkResiliencyTC(true)
					// Type of failure that this TC needs to cover
					failuretype := pdslib.TypeOfFailure{
						Type: KillDeploymentControllerPod,
						Method: func() error {
							return pdslib.KillPodsInNamespace(params.InfraToTest.PDSNamespace, pdslib.PdsDeploymentControllerManagerPod)
						},
					}
					pdslib.DefineFailureType(failuretype)
					// Deploy and Validate this Data service after injecting the type of failure we want to catch
					deployment, _, dsVersionBuildMap, err = dsTest.TriggerDeployDataService(ds, params.InfraToTest.Namespace, tenantID, projectID, false,
						dss.TestParams{NamespaceId: namespaceID, StorageTemplateId: storageTemplateID, DeploymentTargetId: deploymentTargetID, DnsZone: dnsZone, ServiceType: serviceType})
					log.FailOnError(err, "Error while deploying data services")

					err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, params.ResiliencyTest.CheckTillReplica)
					log.FailOnError(err, fmt.Sprintf("Error happened while executing Kill Deployment Controller test for data service %v", *deployment.ClusterResourceName))

					dataServiceDefaultResourceTemplateID, err = controlPlane.GetResourceTemplate(tenantID, ds.Name)
					log.FailOnError(err, "Error while getting resource template")
					log.InfoD("dataServiceDefaultResourceTemplateID %v ", dataServiceDefaultResourceTemplateID)

					resourceTemp, storageOp, config, err := pdslib.ValidateDataServiceVolumes(deployment, ds.Name, dataServiceDefaultResourceTemplateID, storageTemplateID, namespace)
					log.FailOnError(err, "error on ValidateDataServiceVolumes method")
					ValidateDeployments(resourceTemp, storageOp, config, int(ds.Replicas), dsVersionBuildMap)
				})
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()

		if !isDeploymentsDeleted {
			Step("Delete created deployments")
			resp, err := pdslib.DeleteDeployment(deployment.GetId())
			log.FailOnError(err, "Error while deleting data services")
			dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
		}
	})
})

var _ = Describe("{RebootAllWorkerNodesDuringDeployment}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("RebootAllWorkerNodesDuringDeployment", "Reboots all worker nodes while a data service pod is coming up", pdsLabels, 0)
	})

	It("deploy Dataservices", func() {
		Step("Deploy Data Services", func() {
			var dsVersionBuildMap = make(map[string][]string)
			for _, ds := range params.DataServiceToTest {
				Step("Start deployment, Reboot multiple nodes on which deployment is coming up and validate data service", func() {
					isDeploymentsDeleted = false
					// Global Resiliency TC marker
					pdslib.MarkResiliencyTC(true)

					// Deploy and Validate this Data service after injecting the type of failure we want to catch
					deployment, _, dsVersionBuildMap, err = dsTest.TriggerDeployDataService(ds, params.InfraToTest.Namespace, tenantID, projectID, false,
						dss.TestParams{NamespaceId: namespaceID, StorageTemplateId: storageTemplateID, DeploymentTargetId: deploymentTargetID, DnsZone: dnsZone, ServiceType: serviceType})
					log.FailOnError(err, "Error while deploying data services")

					// Type of failure that this TC needs to cover
					failuretype := pdslib.TypeOfFailure{
						Type: RebootNodesDuringDeployment,
						Method: func() error {
							return pdslib.RebootWorkerNodesDuringDeployment(params.InfraToTest.Namespace, deployment, "all")
						},
					}
					pdslib.DefineFailureType(failuretype)

					err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, params.ResiliencyTest.CheckTillReplica)
					log.FailOnError(err, fmt.Sprintf("Error happened while executing Reboot all worker nodes test for data service %v", *deployment.ClusterResourceName))

					dataServiceDefaultResourceTemplateID, err = controlPlane.GetResourceTemplate(tenantID, ds.Name)
					log.FailOnError(err, "Error while getting resource setting template")
					dash.VerifyFatal(dataServiceDefaultResourceTemplateID != "", true, "Validating dataServiceDefaultResourceTemplateID")

					resourceTemp, storageOp, config, err := pdslib.ValidateDataServiceVolumes(deployment, ds.Name, dataServiceDefaultResourceTemplateID, storageTemplateID, namespace)
					log.FailOnError(err, "error on ValidateDataServiceVolumes method")
					ValidateDeployments(resourceTemp, storageOp, config, int(ds.Replicas), dsVersionBuildMap)
				})
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()

		if !isDeploymentsDeleted {
			Step("Delete created deployments")
			resp, err := pdslib.DeleteDeployment(deployment.GetId())
			log.FailOnError(err, "Error while deleting data services")
			dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
		}
	})
})

var _ = Describe("{KillAgentDuringDeployment}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("KillAgentDuringDeployment", "Kill Agent Pod when a DS Deployment is happening", pdsLabels, 0)
	})

	It("Deploy Dataservices", func() {
		Step("Deploy Data Services", func() {
			var dsVersionBuildMap = make(map[string][]string)
			for _, ds := range params.DataServiceToTest {
				Step("Start deployment, Kill Agent Pod while deployment is ongoing and validate data service", func() {
					isDeploymentsDeleted = false
					// Global Resiliency TC marker
					pdslib.MarkResiliencyTC(true)
					// Type of failure that this TC needs to cover
					failuretype := pdslib.TypeOfFailure{
						Type: KillAgentPodDuringDeployment,
						Method: func() error {
							return pdslib.KillPodsInNamespace(params.InfraToTest.PDSNamespace, pdslib.PdsAgentPod)
						},
					}
					pdslib.DefineFailureType(failuretype)
					// Deploy and Validate this Data service after injecting the type of failure we want to catch
					deployment, _, dsVersionBuildMap, err = dsTest.TriggerDeployDataService(ds, params.InfraToTest.Namespace, tenantID, projectID, false,
						dss.TestParams{NamespaceId: namespaceID, StorageTemplateId: storageTemplateID, DeploymentTargetId: deploymentTargetID, DnsZone: dnsZone, ServiceType: serviceType})
					log.FailOnError(err, "Error while deploying data services")

					err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, params.ResiliencyTest.CheckTillReplica)
					log.FailOnError(err, fmt.Sprintf("Error happened while executing Kill Agent Pod test for data service %v", *deployment.ClusterResourceName))

					dataServiceDefaultResourceTemplateID, err = controlPlane.GetResourceTemplate(tenantID, ds.Name)
					log.FailOnError(err, "Error while getting resource template")
					log.InfoD("dataServiceDefaultResourceTemplateID %v ", dataServiceDefaultResourceTemplateID)

					resourceTemp, storageOp, config, err := pdslib.ValidateDataServiceVolumes(deployment, ds.Name, dataServiceDefaultResourceTemplateID, storageTemplateID, namespace)
					log.FailOnError(err, "error on ValidateDataServiceVolumes method")

					ValidateDeployments(resourceTemp, storageOp, config, int(ds.Replicas), dsVersionBuildMap)
				})
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()

		if !isDeploymentsDeleted {
			Step("Delete created deployments")
			resp, err := pdslib.DeleteDeployment(deployment.GetId())
			log.FailOnError(err, "Error while deleting data services")
			dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
		}
	})
})

var _ = Describe("{RestartAppDuringResourceUpdate}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("RestartAppDuringResourceUpdate", "Restart application pod during resource update", pdsLabels, 0)
		pdslib.MarkResiliencyTC(true)
	})

	It("Deploy Data Services", func() {
		var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
		Step("Deploy Data Services", func() {
			for _, ds := range params.DataServiceToTest {
				Step("Deploy and validate data service", func() {
					deployment, _, _, err = DeployandValidateDataServices(ds, params.InfraToTest.Namespace, tenantID, projectID)
					log.FailOnError(err, "Error while deploying data services")
					deployments[ds] = deployment
				})
			}
		})

		defer func() {
			for _, newDeployment := range deployments {
				Step("Delete created deployments")
				resp, err := pdslib.DeleteDeployment(newDeployment.GetId())
				log.FailOnError(err, "Error while deleting data services")
				dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
				err = pdslib.DeletePvandPVCs(*newDeployment.ClusterResourceName, false)
				log.FailOnError(err, "Error while deleting PV and PVCs")
			}
		}()

		Step("Update the resource and Restart application pods", func() {
			for _, deployment := range deployments {
				failureType := pdslib.TypeOfFailure{
					Type: RestartAppDuringResourceUpdate,
					Method: func() error {
						return pdslib.RestartApplicationDuringResourceUpdate(params.InfraToTest.Namespace, deployment)
					},
				}
				pdslib.DefineFailureType(failureType)

				err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, 0)
				log.FailOnError(err, fmt.Sprintf("Error while pod restart during Resource update %v", *deployment.ClusterResourceName))

				err = dsTest.ValidateDataServiceDeployment(deployment, namespace)
				log.FailOnError(err, "error on ValidateDataServiceDeployment")
			}
		})
	})
	JustAfterEach(func() {
		EndTorpedoTest()
	})
})

var _ = Describe("{KillTeleportDuringDeployment}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("KillTeleportDuringDeployment", "Kill Teleport Pod when a DS Deployment is happening", pdsLabels, 0)
	})

	It("Deploy Dataservices", func() {
		Step("Deploy Data Services", func() {
			var dsVersionBuildMap = make(map[string][]string)
			for _, ds := range params.DataServiceToTest {
				Step("Start deployment, Kill Agent Pod while deployment is ongoing and validate data service", func() {
					isDeploymentsDeleted = false
					// Global Resiliency TC marker
					pdslib.MarkResiliencyTC(true)
					// Type of failure that this TC needs to cover
					failuretype := pdslib.TypeOfFailure{
						Type: KillTeleportPodDuringDeployment,
						Method: func() error {
							return pdslib.KillPodsInNamespace(params.InfraToTest.PDSNamespace, pdslib.PdsTeleportPod)
						},
					}
					pdslib.DefineFailureType(failuretype)
					// Deploy and Validate this Data service after injecting the type of failure we want to catch
					deployment, _, dsVersionBuildMap, err = dsTest.TriggerDeployDataService(ds, params.InfraToTest.Namespace, tenantID, projectID, false,
						dss.TestParams{NamespaceId: namespaceID, StorageTemplateId: storageTemplateID, DeploymentTargetId: deploymentTargetID, DnsZone: dnsZone, ServiceType: serviceType})
					log.FailOnError(err, "Error while deploying data services")

					err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, params.ResiliencyTest.CheckTillReplica)
					log.FailOnError(err, fmt.Sprintf("Error happened while executing Kill Teleport Pod test for data service %v", *deployment.ClusterResourceName))

					dataServiceDefaultResourceTemplateID, err = controlPlane.GetResourceTemplate(tenantID, ds.Name)
					log.FailOnError(err, "Error while getting resource template")
					log.InfoD("dataServiceDefaultResourceTemplateID %v ", dataServiceDefaultResourceTemplateID)

					resourceTemp, storageOp, config, err := pdslib.ValidateDataServiceVolumes(deployment, ds.Name, dataServiceDefaultResourceTemplateID, storageTemplateID, namespace)
					log.FailOnError(err, "error on ValidateDataServiceVolumes method")

					ValidateDeployments(resourceTemp, storageOp, config, int(ds.Replicas), dsVersionBuildMap)
				})
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()

		if !isDeploymentsDeleted {
			Step("Delete created deployments")
			resp, err := pdslib.DeleteDeployment(deployment.GetId())
			log.FailOnError(err, "Error while deleting data services")
			dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
		}
	})
})

var _ = Describe("{RebootActiveNodeMultipleTimesDuringDeployment}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("RebootActiveNodeMultipleTimesDuringDeployment", "Reboots a Node multiple times onto which a pod is coming up", pdsLabels, 0)
	})

	It("deploy Dataservices", func() {
		Step("Deploy Data Services", func() {
			for _, ds := range params.DataServiceToTest {
				var dsVersionBuildMap = make(map[string][]string)
				var num_reboots int
				num_reboots = 3
				Step("Start deployment, Reboot a node on which deployment is coming up and validate data service", func() {
					isDeploymentsDeleted = false
					// Global Resiliency TC marker
					pdslib.MarkResiliencyTC(true)

					// Deploy and Validate this Data service after injecting the type of failure we want to catch
					deployment, _, dsVersionBuildMap, err = dsTest.TriggerDeployDataService(ds, params.InfraToTest.Namespace, tenantID, projectID, false,
						dss.TestParams{NamespaceId: namespaceID, StorageTemplateId: storageTemplateID, DeploymentTargetId: deploymentTargetID, DnsZone: dnsZone, ServiceType: serviceType})
					log.FailOnError(err, "Error while deploying data services")

					// Type of failure that this TC needs to cover
					failuretype := pdslib.TypeOfFailure{
						Type: ActiveNodeRebootDuringDeployment,
						Method: func() error {
							return pdslib.RebootActiveNodeDuringDeployment(params.InfraToTest.Namespace, deployment, num_reboots)
						},
					}
					pdslib.DefineFailureType(failuretype)

					err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, params.ResiliencyTest.CheckTillReplica)
					log.FailOnError(err, fmt.Sprintf("Error happened while executing Reboot test for data service %v", *deployment.ClusterResourceName))

					dataServiceDefaultResourceTemplateID, err = controlPlane.GetResourceTemplate(tenantID, ds.Name)
					log.FailOnError(err, "Error while getting resource setting template")
					dash.VerifyFatal(dataServiceDefaultResourceTemplateID != "", true, "Validating dataServiceDefaultResourceTemplateID")

					resourceTemp, storageOp, config, err := pdslib.ValidateDataServiceVolumes(deployment, ds.Name, dataServiceDefaultResourceTemplateID, storageTemplateID, namespace)
					log.FailOnError(err, "error on ValidateDataServiceVolumes method")
					ValidateDeployments(resourceTemp, storageOp, config, int(ds.Replicas), dsVersionBuildMap)
				})
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()

		if !isDeploymentsDeleted {
			Step("Delete created deployments")
			resp, err := pdslib.DeleteDeployment(deployment.GetId())
			log.FailOnError(err, "Error while deleting data services")
			dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
		}
	})
})

var _ = Describe("{KillPdsAgentDuringWorkloadRun}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("KillPdsAgentDuringWorkloadRun", "Kill Pds Agent Pods while Workload is running", pdsLabels, 0)
	})
	It("Deploy Dataservices and Restart PX During App scaleup", func() {
		var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
		var generateWorkloads = make(map[string]string)
		Step("Deploy Data Services", func() {
			for _, ds := range params.DataServiceToTest {
				Step("Deploy and validate data service", func() {
					isDeploymentsDeleted = false
					deployment, _, _, err = DeployandValidateDataServices(ds, params.InfraToTest.Namespace, tenantID, projectID)
					log.FailOnError(err, "Error while deploying data services")
					deployments[ds] = deployment
				})
			}
		})

		defer func() {
			for _, newDeployment := range deployments {
				Step("Delete created deployments")
				resp, err := pdslib.DeleteDeployment(newDeployment.GetId())
				log.FailOnError(err, "Error while deleting data services")
				dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
				err = pdslib.DeletePvandPVCs(*newDeployment.ClusterResourceName, false)
				log.FailOnError(err, "Error while deleting PV and PVCs")
			}
		}()
		Step("Running Workloads", func() {
			for ds, deployment := range deployments {
				if Contains(dataServicePodWorkloads, ds.Name) || Contains(dataServiceDeploymentWorkloads, ds.Name) {
					log.InfoD("Running Workloads on DataService %v ", ds.Name)
					var params pdslib.WorkloadGenerationParams
					pod, dep, err = RunWorkloads(params, ds, deployment, namespace)
					log.FailOnError(err, fmt.Sprintf("Error while generating workloads for dataservice [%s]", ds.Name))
					if dep == nil {
						generateWorkloads[ds.Name] = pod.Name
					} else {
						generateWorkloads[ds.Name] = dep.Name
					}
					for dsName, workloadContainer := range generateWorkloads {
						log.Debugf("dsName %s, workloadContainer %s", dsName, workloadContainer)
					}
				} else {
					log.InfoD("Workload script not available for ds %v", ds.Name)
				}
			}
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

		Step("Killing PDS Agent Pods", func() {
			err = pdslib.KillPodsInNamespace(params.InfraToTest.PDSNamespace, pdslib.PdsAgentPod)
			log.FailOnError(err, "Failed while deleting PDS Agent Pods")
		})

		// TODO : Once Workload Validation Module is ready, we will add that here. AI: @jyoti

	})
	JustAfterEach(func() {
		EndTorpedoTest()
	})
})

var _ = Describe("{RebootMoreThanQuorumWorkerNodesDuringDeployment}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("RebootMoreThanQuorumWorkerNodesDuringDeployment", "Reboots more worker nodes than required for Px Quorum while a data service pod is coming up", pdsLabels, 0)
	})

	It("Deploy DS, Reboot Nodes, Validate DS, Run Workload", func() {
		var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
		var generateWorkloads = make(map[string]string)
		Step("Deploy DS, Reboot Nodes, Validate DS, Run Workload", func() {
			var dsVersionBuildMap = make(map[string][]string)
			for _, ds := range params.DataServiceToTest {
				Step("Start deployment, Reboot multiple nodes on which deployment is coming up, validate data service and run workload", func() {
					isDeploymentsDeleted = false
					// Global Resiliency TC marker
					pdslib.MarkResiliencyTC(true)

					// Deploy and Validate this Data service after injecting the type of failure we want to catch
					deployment, _, dsVersionBuildMap, err = dsTest.TriggerDeployDataService(ds, params.InfraToTest.Namespace, tenantID, projectID, false,
						dss.TestParams{NamespaceId: namespaceID, StorageTemplateId: storageTemplateID, DeploymentTargetId: deploymentTargetID, DnsZone: dnsZone, ServiceType: serviceType})
					log.FailOnError(err, "Error while deploying data services")

					deployments[ds] = deployment
					// Type of failure that this TC needs to cover
					failuretype := pdslib.TypeOfFailure{
						Type: RebootNodesDuringDeployment,
						Method: func() error {
							return pdslib.RebootWorkerNodesDuringDeployment(params.InfraToTest.Namespace, deployment, "quorum")
						},
					}
					pdslib.DefineFailureType(failuretype)

					err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, params.ResiliencyTest.CheckTillReplica)
					log.FailOnError(err, fmt.Sprintf("Error happened while executing Reboot all worker nodes test for data service %v", *deployment.ClusterResourceName))

					dataServiceDefaultResourceTemplateID, err = controlPlane.GetResourceTemplate(tenantID, ds.Name)
					log.FailOnError(err, "Error while getting resource setting template")
					dash.VerifyFatal(dataServiceDefaultResourceTemplateID != "", true, "Validating dataServiceDefaultResourceTemplateID")

					resourceTemp, storageOp, config, err := pdslib.ValidateDataServiceVolumes(deployment, ds.Name, dataServiceDefaultResourceTemplateID, storageTemplateID, namespace)
					log.FailOnError(err, "error on ValidateDataServiceVolumes method")
					ValidateDeployments(resourceTemp, storageOp, config, int(ds.Replicas), dsVersionBuildMap)

					Step("Running Workloads", func() {
						for ds, deployment := range deployments {
							if Contains(dataServicePodWorkloads, ds.Name) || Contains(dataServiceDeploymentWorkloads, ds.Name) {
								log.InfoD("Running Workloads on DataService %v ", ds.Name)
								var params pdslib.WorkloadGenerationParams
								pod, dep, err = RunWorkloads(params, ds, deployment, namespace)
								log.FailOnError(err, fmt.Sprintf("Error while genearating workloads for dataservice [%s]", ds.Name))
								if dep == nil {
									generateWorkloads[ds.Name] = pod.Name
								} else {
									generateWorkloads[ds.Name] = dep.Name
								}
								for dsName, workloadContainer := range generateWorkloads {
									log.Debugf("dsName %s, workloadContainer %s", dsName, workloadContainer)
								}
							} else {
								log.InfoD("Workload script not available for ds %v", ds.Name)
							}
						}
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

				})
			}
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()

		if !isDeploymentsDeleted {
			Step("Delete created deployments")
			resp, err := pdslib.DeleteDeployment(deployment.GetId())
			log.FailOnError(err, "Error while deleting data services")
			dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
		}
	})
})

var _ = Describe("{RebootNodeForUnrelatedDS}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("RebootNodeForUnrelatedDS", "Reboot Nodes that are unrelated to the Deployment", pdsLabels, 0)
		pdslib.MarkResiliencyTC(true)
	})

	It("Deploy Dataservices and Restart PX During App scaleup", func() {
		var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
		var generateWorkloads = make(map[string]string)
		var originalDS = make(map[PDSDataService]*pds.ModelsDeployment)

		Step("Deploy Data Services", func() {
			for _, ds := range params.DataServiceToTest {
				Step("Deploy and validate data service", func() {
					isDeploymentsDeleted = false
					deployment, _, _, err = DeployandValidateDataServices(ds, params.InfraToTest.Namespace, tenantID, projectID)
					log.FailOnError(err, "Error while deploying data services")
					originalDS[ds] = deployment
				})
			}
		})

		defer func() {
			for _, newDeployment := range deployments {
				Step("Delete created deployments")
				resp, err := pdslib.DeleteDeployment(newDeployment.GetId())
				log.FailOnError(err, "Error while deleting data services")
				dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
				err = pdslib.DeletePvandPVCs(*newDeployment.ClusterResourceName, false)
				log.FailOnError(err, "Error while deleting PV and PVCs")
			}
			for _, newDeployment := range originalDS {
				Step("Delete created deployments")
				resp, err := pdslib.DeleteDeployment(newDeployment.GetId())
				log.FailOnError(err, "Error while deleting data services")
				dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
				err = pdslib.DeletePvandPVCs(*newDeployment.ClusterResourceName, false)
				log.FailOnError(err, "Error while deleting PV and PVCs")
			}
		}()
		Step("Running Workloads", func() {
			for ds, deployment := range deployments {
				if Contains(dataServicePodWorkloads, ds.Name) || Contains(dataServiceDeploymentWorkloads, ds.Name) {
					log.InfoD("Running Workloads on DataService %v ", ds.Name)
					var params pdslib.WorkloadGenerationParams
					pod, dep, err = RunWorkloads(params, ds, deployment, namespace)
					log.FailOnError(err, fmt.Sprintf("Error while genearating workloads for dataservice [%s]", ds.Name))
					if dep == nil {
						generateWorkloads[ds.Name] = pod.Name
					} else {
						generateWorkloads[ds.Name] = dep.Name
					}
					for dsName, workloadContainer := range generateWorkloads {
						log.Debugf("dsName %s, workloadContainer %s", dsName, workloadContainer)
					}
				} else {
					log.InfoD("Workload script not available for ds %v", ds.Name)
				}
			}
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
		Step("Disable Scheduling on the nodes of deployment", func() {
			for _, deployment := range deployments {
				nodes, err := pdslib.GetNodesOfSS(*deployment.ClusterResourceName, namespace)
				log.FailOnError(err, fmt.Sprintf("Cannot fetch nodes of the running Data Service %v", *deployment.ClusterResourceName))
				for _, nodeObj := range nodes {
					err = k8sCore.CordonNode(nodeObj.Name, defaultCommandTimeout, defaultCommandRetry)
					log.FailOnError(err, fmt.Sprintf("Error in disabling scheduling for node %v", nodeObj.Name))
					log.Infof("Node with Name : %v is now cordoned", nodeObj.Name)
				}
			}
		})
		Step("Deploy Data Services", func() {
			for _, ds := range params.DataServiceToTest {
				var dsVersionBuildMap = make(map[string][]string)
				var newDeployment *pds.ModelsDeployment
				var num_reboots int
				num_reboots = 1
				Step("Start deployment, Reboot a node on which deployment is coming up and validate data service", func() {
					isDeploymentsDeleted = false
					// Global Resiliency TC marker
					pdslib.MarkResiliencyTC(true)

					// Deploy and Validate this Data service after injecting the type of failure we want to catch
					newDeployment, _, dsVersionBuildMap, err = dsTest.TriggerDeployDataService(ds, params.InfraToTest.Namespace, tenantID, projectID, false,
						dss.TestParams{NamespaceId: namespaceID, StorageTemplateId: storageTemplateID, DeploymentTargetId: deploymentTargetID, DnsZone: dnsZone, ServiceType: serviceType})
					log.FailOnError(err, "Error while deploying data services")
					deployments[ds] = newDeployment
					// Type of failure that this TC needs to cover
					failuretype := pdslib.TypeOfFailure{
						Type: ActiveNodeRebootDuringDeployment,
						Method: func() error {
							return pdslib.RebootActiveNodeDuringDeployment(params.InfraToTest.Namespace, newDeployment, num_reboots)
						},
					}
					pdslib.DefineFailureType(failuretype)

					err = pdslib.InduceFailureAfterWaitingForCondition(newDeployment, namespace, params.ResiliencyTest.CheckTillReplica)
					log.FailOnError(err, fmt.Sprintf("Error happened while executing Reboot test for data service %v", *deployment.ClusterResourceName))

					dataServiceDefaultResourceTemplateID, err = controlPlane.GetResourceTemplate(tenantID, ds.Name)
					log.FailOnError(err, "Error while getting resource setting template")
					dash.VerifyFatal(dataServiceDefaultResourceTemplateID != "", true, "Validating dataServiceDefaultResourceTemplateID")

					resourceTemp, storageOp, config, err := pdslib.ValidateDataServiceVolumes(newDeployment, ds.Name, dataServiceDefaultResourceTemplateID, storageTemplateID, namespace)
					log.FailOnError(err, "error on ValidateDataServiceVolumes method")
					ValidateDeployments(resourceTemp, storageOp, config, int(ds.Replicas), dsVersionBuildMap)
				})
			}
		})

		// TODO : Once Workload Validation Module is ready, we will add that here. AI: @jyoti

		Step("Enable Scheduling on the nodes of deployment", func() {
			nodeList, err := k8sCore.GetNodes()
			log.FailOnError(err, fmt.Sprintf("Cannot fetch nodes of the running Data Service %v", *deployment.ClusterResourceName))
			for _, nodeObj := range nodeList.Items {
				err = k8sCore.UnCordonNode(nodeObj.Name, defaultCommandTimeout, defaultCommandRetry)
				log.FailOnError(err, fmt.Sprintf("Error in re-enabling scheduling for node %v", nodeObj.Name))
				log.Infof("Node with name %v successfully uncordoned", nodeObj.Name)
			}
		})
		Step("Removing Existing and Running New Workloads Again", func() {
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
			for ds, deployment := range deployments {
				if Contains(dataServicePodWorkloads, ds.Name) || Contains(dataServiceDeploymentWorkloads, ds.Name) {
					log.InfoD("Running Workloads on DataService %v ", ds.Name)
					var params pdslib.WorkloadGenerationParams
					pod, dep, err = RunWorkloads(params, ds, deployment, namespace)
					log.FailOnError(err, fmt.Sprintf("Error while genearating workloads for dataservice [%s]", ds.Name))
					if dep == nil {
						generateWorkloads[ds.Name] = pod.Name
					} else {
						generateWorkloads[ds.Name] = dep.Name
					}
					for dsName, workloadContainer := range generateWorkloads {
						log.Debugf("dsName %s, workloadContainer %s", dsName, workloadContainer)
					}
				} else {
					log.InfoD("Workload script not available for ds %v", ds.Name)
				}
			}
		})
	})
	JustAfterEach(func() {
		EndTorpedoTest()
	})
})

var _ = Describe("{KillTeleportDuringWorkloadRun}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("KillTeleportDuringWorkloadRun", "Kill Pds Agent Pods while Workload is running", pdsLabels, 0)
	})
	It("Deploy Dataservices and Restart PX During App scaleup", func() {
		var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
		var generateWorkloads = make(map[string]string)
		Step("Deploy Data Services", func() {
			for _, ds := range params.DataServiceToTest {
				Step("Deploy and validate data service", func() {
					isDeploymentsDeleted = false
					deployment, _, _, err = DeployandValidateDataServices(ds, params.InfraToTest.Namespace, tenantID, projectID)
					log.FailOnError(err, "Error while deploying data services")
					deployments[ds] = deployment
				})
			}
		})

		defer func() {
			for _, newDeployment := range deployments {
				Step("Delete created deployments")
				resp, err := pdslib.DeleteDeployment(newDeployment.GetId())
				log.FailOnError(err, "Error while deleting data services")
				dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
				err = pdslib.DeletePvandPVCs(*newDeployment.ClusterResourceName, false)
				log.FailOnError(err, "Error while deleting PV and PVCs")
			}
		}()
		Step("Running Workloads", func() {
			for ds, deployment := range deployments {
				if Contains(dataServicePodWorkloads, ds.Name) || Contains(dataServiceDeploymentWorkloads, ds.Name) {
					log.InfoD("Running Workloads on DataService %v ", ds.Name)
					var params pdslib.WorkloadGenerationParams
					pod, dep, err = RunWorkloads(params, ds, deployment, namespace)
					log.FailOnError(err, fmt.Sprintf("Error while generating workloads for dataservice [%s]", ds.Name))
					if dep == nil {
						generateWorkloads[ds.Name] = pod.Name
					} else {
						generateWorkloads[ds.Name] = dep.Name
					}
					for dsName, workloadContainer := range generateWorkloads {
						log.Debugf("dsName %s, workloadContainer %s", dsName, workloadContainer)
					}
				} else {
					log.InfoD("Workload script not available for ds %v", ds.Name)
				}
			}
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

		Step("Killing Teleport Pods", func() {
			err = pdslib.KillPodsInNamespace(params.InfraToTest.PDSNamespace, pdslib.PdsTeleportPod)
			log.FailOnError(err, "Failed while deleting PDS Agent Pods")
		})

		// TODO : Once Workload Validation Module is ready, we will add that here. AI: @jyoti

	})
	JustAfterEach(func() {
		EndTorpedoTest()
	})
})

// This testcase requires a cloud-drive setup
var _ = Describe("{RestoreDSDuringPXPoolExpansion}", func() {
	var deps []*pds.ModelsDeployment
	pdsdeploymentsmd5Hash := make(map[string]string)
	restoredDeploymentsmd5Hash := make(map[string]string)
	var deploymentsToBeCleaned []*pds.ModelsDeployment
	var wlDeploymentsToBeCleaned []*v1.Deployment
	JustBeforeEach(func() {
		StartTorpedoTest("RestoreDSDuringPXPoolExpansion", "Restore DataService during the PX Pool expansion", pdsLabels, 0)
		pdslib.MarkResiliencyTC(true)
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
	It("Deploy Dataservices and Restore during PX-Pool Expansion", func() {
		var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
		var depList []*pds.ModelsDeployment
		Step("Deploy Data Services", func() {
			for _, ds := range params.DataServiceToTest {
				Step("Deploy and validate data service", func() {
					isDeploymentsDeleted = false
					deployment, _, _, err = DeployandValidateDataServices(ds, params.InfraToTest.Namespace, tenantID, projectID)
					log.FailOnError(err, "Error while deploying data services")
					deployments[ds] = deployment
					depList = append(depList, deployment)
					deploymentsToBeCleaned = append(deploymentsToBeCleaned, deployment)
					deps = append(deps, deployment)
					dsEntity = restoreBkp.DSEntity{
						Deployment: deployment,
					}
				})
			}
		})

		Step("Running Workloads before taking backups", func() {
			for _, pdsDeployment := range deps {
				ckSum, wlDep, err := dsTest.InsertDataAndReturnChecksum(pdsDeployment, wkloadParams)
				wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDep)
				log.FailOnError(err, "Error while Running workloads")
				log.Debugf("Checksum for the deployment %s is %s", *pdsDeployment.ClusterResourceName, ckSum)
				pdsdeploymentsmd5Hash[*pdsDeployment.ClusterResourceName] = ckSum
			}
		})
		Step("Perform adhoc backup and validate them", func() {
			log.Infof("Deployment ID: %v, backup target ID: %v", deployment.GetId(), bkpTarget.GetId())
			err = bkpClient.TriggerAndValidateAdhocBackup(deployment.GetId(), bkpTarget.GetId(), "s3")
			log.FailOnError(err, "Failed while performing adhoc backup")
		})
		Step("Resize PX-POOL and trigger restore at the same time", func() {
			for _, deployment := range deployments {
				failuretype := pdslib.TypeOfFailure{
					Type: RestoreDSDuringPXPoolExpansion,
					Method: func() error {
						ctx, err := Inst().Pds.CreateSchedulerContextForPDSApps(depList)
						log.Infof("Created scheduler context", ctx)
						log.FailOnError(err, "Unable to create the scheduler context")
						return pdslib.ExpandAndValidatePxPool(ctx)
					},
				}
				pdslib.DefineFailureType(failuretype)
				err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, params.ResiliencyTest.CheckTillReplica)
				log.FailOnError(err, fmt.Sprintf("Error happened while restarting px for data service %v", *deployment.ClusterResourceName))
			}
		})
		Step("Validate Deployments after Px-pool Resize", func() {
			for ds, deployment := range deployments {
				err = dsTest.ValidateDataServiceDeployment(deployment, namespace)
				log.FailOnError(err, "Error while validating dataservices")
				log.InfoD("Data-service: %v is up and healthy", ds.Name)
			}
			dsEntity = restoreBkp.DSEntity{
				Deployment: deployment,
			}
		})
		Step("Taking adhoc backup and trigger restore again", func() {
			log.Infof("Deployment ID: %v, backup target ID: %v", deployment.GetId(), bkpTarget.GetId())
			err = bkpClient.TriggerAndValidateAdhocBackup(deployment.GetId(), bkpTarget.GetId(), "s3")
			log.FailOnError(err, "Failed while performing adhoc backup")
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
				deploymentsToBeCleaned = append(deploymentsToBeCleaned, restoredDeployment)
				log.InfoD("Restored successfully. Deployment- %v", restoredModel.GetClusterResourceName())
			}
		})
		Step("Validate md5hash for the restored deployments", func() {
			for _, pdsDeployment := range deploymentsToBeCleaned {
				ckSum, wlDep, err := dsTest.ReadDataAndReturnChecksum(pdsDeployment, wkloadParams)
				wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDep)
				log.FailOnError(err, "Error while Running workloads")
				log.Debugf("Checksum for the deployment %s is %s", *pdsDeployment.ClusterResourceName, ckSum)
				restoredDeploymentsmd5Hash[*pdsDeployment.ClusterResourceName] = ckSum
			}
			defer func() {
				for _, wlDep := range wlDeploymentsToBeCleaned {
					err := k8sApps.DeleteDeployment(wlDep.Name, wlDep.Namespace)
					log.FailOnError(err, "Failed while deleting the workload deployment")
				}
			}()
			dash.VerifyFatal(dsTest.ValidateDataMd5Hash(pdsdeploymentsmd5Hash, restoredDeploymentsmd5Hash),
				true, "Validate md5 hash after restore")
		})
		Step("Delete Deployments", func() {
			CleanupDeployments(deploymentsToBeCleaned)
		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{RestoreDuringNodesAreRebooted}", func() {
	var deps []*pds.ModelsDeployment
	pdsdeploymentsmd5Hash := make(map[string]string)
	restoredDeploymentsmd5Hash := make(map[string]string)
	var deploymentsToBeCleaned []*pds.ModelsDeployment
	var wlDeploymentsToBeCleaned []*v1.Deployment
	JustBeforeEach(func() {
		StartTorpedoTest("RestoreDuringNodesAreRebooted", "Restore DataService during nodes are rebooted", pdsLabels, 0)
		pdslib.MarkResiliencyTC(true)
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
	It("Deploy Dataservices and Restore during nodes are rebooted", func() {
		var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
		var depList []*pds.ModelsDeployment
		Step("Deploy Data Services", func() {
			for _, ds := range params.DataServiceToTest {
				Step("Deploy and validate data service", func() {
					isDeploymentsDeleted = false
					deployment, _, _, err = DeployandValidateDataServices(ds, params.InfraToTest.Namespace, tenantID, projectID)
					log.FailOnError(err, "Error while deploying data services")
					deployments[ds] = deployment
					depList = append(depList, deployment)
					deploymentsToBeCleaned = append(deploymentsToBeCleaned, deployment)
					deps = append(deps, deployment)
					log.InfoD("Number of backups to be taken are- %v", len(deps))
					dsEntity = restoreBkp.DSEntity{
						Deployment: deployment,
					}
				})
			}
		})
		Step("Running Workloads before taking backups", func() {
			for _, pdsDeployment := range deps {
				ckSum, wlDep, err := dsTest.InsertDataAndReturnChecksum(pdsDeployment, wkloadParams)
				log.FailOnError(err, "Error while Running workloads")
				wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDep)
				log.Debugf("Checksum for the deployment %s is %s", *pdsDeployment.ClusterResourceName, ckSum)
				pdsdeploymentsmd5Hash[*pdsDeployment.ClusterResourceName] = ckSum
			}
		})
		Step("Perform multiple adhoc backup and validate them", func() {
			log.Infof("Deployment ID: %v, backup target ID: %v", deployment.GetId(), bkpTarget.GetId())
			for _, pdsDeployment := range deps {
				log.Infof("Deployment ID: %v, backup target ID: %v", pdsDeployment.GetId(), bkpTarget.GetId())
				err = bkpClient.TriggerAndValidateAdhocBackup(pdsDeployment.GetId(), bkpTarget.GetId(), "s3")
				log.FailOnError(err, "Failed while performing adhoc backup")
			}
		})
		stepLog := "Trigger restore during all worker nodes are getting rebooted"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for ds, deployment := range deployments {
				failuretype := pdslib.TypeOfFailure{
					Type: RestoreDuringAllNodesReboot,
					Method: func() error {
						return pdslib.RebootWorkernodesDuringRestore(params.InfraToTest.Namespace, deployment, "all")
					},
				}
				pdslib.DefineFailureType(failuretype)
				err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, params.ResiliencyTest.CheckTillReplica)
				log.FailOnError(err, fmt.Sprintf("Error happened while restarting px for data service %v", *deployment.ClusterResourceName))

				//validate original deployment after pod reboot
				err = dsTest.ValidateDataServiceDeployment(deployment, namespace)
				log.FailOnError(err, "Error while validating dataservices")
				log.InfoD("Data-service: %v is up and healthy", ds.Name)

				newDeps := pdslib.GetRestoredDeployment()

				//validate restored deployments health
				for _, pdsDeployment := range newDeps {
					err = dsTest.ValidateDataServiceDeployment(pdsDeployment, namespace)
					log.FailOnError(err, "Error while validating dataservices")
					log.InfoD("Data-service: %v is up and healthy", ds.Name)
				}

			}
		})

		stepLog = "Trigger restore during all ds nodes are getting rebooted (ds not in quorum)"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			for ds, deployment := range deployments {
				failuretype := pdslib.TypeOfFailure{
					Type: RestoreDuringAllNodesReboot,
					Method: func() error {
						return pdslib.RebootWorkerNodesDuringDeployment(params.InfraToTest.Namespace, deployment, "quorum")
					},
				}
				pdslib.DefineFailureType(failuretype)
				err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, params.ResiliencyTest.CheckTillReplica)
				log.FailOnError(err, fmt.Sprintf("Error happened while restarting px for data service %v", *deployment.ClusterResourceName))

				//validate original deployment after pod reboot
				err = dsTest.ValidateDataServiceDeployment(deployment, namespace)
				log.FailOnError(err, "Error while validating dataservices")
				log.InfoD("Data-service: %v is up and healthy", ds.Name)

				newDeps := pdslib.GetRestoredDeployment()

				//validate restored deployments health
				for _, pdsDeployment := range newDeps {
					err = dsTest.ValidateDataServiceDeployment(pdsDeployment, namespace)
					log.FailOnError(err, "Error while validating dataservices")
					log.InfoD("Data-service: %v is up and healthy", ds.Name)
				}

				Step("Validate md5hash for the restored deployments", func() {
					for _, pdsDeployment := range newDeps {
						ckSum, wlDep, err := dsTest.ReadDataAndReturnChecksum(pdsDeployment, wkloadParams)
						wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDep)
						log.FailOnError(err, "Error while Running workloads")
						restoredDeploymentsmd5Hash[*pdsDeployment.ClusterResourceName] = ckSum
					}
					dash.VerifyFatal(dsTest.ValidateDataMd5Hash(pdsdeploymentsmd5Hash, restoredDeploymentsmd5Hash),
						true, "Validate md5 hash after restore")
				})
			}
		})

		stepLog = "Trigger restore and reboot the node on which restored ds is hosted"
		Step(stepLog, func() {
			log.InfoD(stepLog)
			num_reboots := 1
			for ds, deployment := range deployments {
				// Trigger Restore and get the restored deployment
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
					restoredModel, err := restoreClient.TriggerAndValidateRestore(backupJob.GetId(), params.InfraToTest.Namespace, dsEntity, true, false)
					log.FailOnError(err, "Failed during restore.")
					restoredDeployment, err = restoreClient.Components.DataServiceDeployment.GetDeployment(restoredModel.GetDeploymentId())
					log.FailOnError(err, fmt.Sprintf("Failed while fetching the restore data service instance: %v", restoredModel.GetClusterResourceName()))
					deploymentsToBeCleaned = append(deploymentsToBeCleaned, restoredDeployment)
					log.InfoD("Restored successfully. Deployment- %v", restoredModel.GetClusterResourceName())

					failuretype := pdslib.TypeOfFailure{
						Type: ActiveNodeRebootDuringDeployment,
						Method: func() error {
							return pdslib.RebootActiveNodeDuringDeployment(params.InfraToTest.Namespace, restoredDeployment, num_reboots)
						},
					}
					pdslib.DefineFailureType(failuretype)
					err = pdslib.InduceFailureAfterWaitingForCondition(restoredDeployment, namespace, params.ResiliencyTest.CheckTillReplica)
					log.FailOnError(err, fmt.Sprintf("Error happened while restarting px for data service %v", *restoredDeployment.ClusterResourceName))

					//validate original deployment after node reboot
					err = dsTest.ValidateDataServiceDeployment(deployment, namespace)
					log.FailOnError(err, "Error while validating dataservices")
					log.InfoD("Data-service: %v is up and healthy", ds.Name)

					//validate restored deployment after node reboot
					err = dsTest.ValidateDataServiceDeployment(restoredDeployment, namespace)
					log.FailOnError(err, "Error while validating dataservices")
					log.InfoD("Data-service: %v is up and healthy", ds.Name)
				}

				Step("Validate md5hash for the restored deployments", func() {
					ckSum, wlDep, err := dsTest.ReadDataAndReturnChecksum(restoredDeployment, wkloadParams)
					wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDep)
					log.FailOnError(err, "Error while Running workloads")
					restoredDeploymentsmd5Hash[*restoredDeployment.ClusterResourceName] = ckSum

					dash.VerifyFatal(dsTest.ValidateDataMd5Hash(pdsdeploymentsmd5Hash, restoredDeploymentsmd5Hash),
						true, "Validate md5 hash after restore")
				})
			}
		})

		Step("CleanUp Workload Deployments", func() {
			for _, wlDep := range wlDeploymentsToBeCleaned {
				log.Debugf("Deleting workload deployment [%s]", wlDep.Name)
				err := k8sApps.DeleteDeployment(wlDep.Name, wlDep.Namespace)
				log.FailOnError(err, "Failed while deleting the workload deployment")
			}
		})

		Step("Delete Deployments", func() {
			dynamicDeps := pdslib.GetDynamicDeployments()
			deploymentsToBeCleaned = append(deploymentsToBeCleaned, dynamicDeps...)
			CleanupDeployments(deploymentsToBeCleaned)

		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{RestoreDSDuringKVDBFailOver}", func() {
	var deps []*pds.ModelsDeployment
	pdsdeploymentsmd5Hash := make(map[string]string)
	restoredDeploymentsmd5Hash := make(map[string]string)
	var deploymentsToBeCleaned []*pds.ModelsDeployment
	var wlDeploymentsToBeCleaned []*v1.Deployment
	JustBeforeEach(func() {
		StartTorpedoTest("RestoreDSDuringKVDBFailOver", "Restore DataService during KVDB Pods are down", pdsLabels, 0)
		pdslib.MarkResiliencyTC(true)
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
	It("Deploy Dataservices and Restore during KVDB failover", func() {
		var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
		var depList []*pds.ModelsDeployment
		Step("Deploy Data Services", func() {
			for _, ds := range params.DataServiceToTest {
				Step("Deploy and validate data service", func() {
					isDeploymentsDeleted = false
					deployment, _, _, err = DeployandValidateDataServices(ds, params.InfraToTest.Namespace, tenantID, projectID)
					log.FailOnError(err, "Error while deploying data services")
					deployments[ds] = deployment
					depList = append(depList, deployment)
					deploymentsToBeCleaned = append(deploymentsToBeCleaned, deployment)
					deps = append(deps, deployment)
					log.InfoD("Number of backups to be taken are- %v", len(deps))
					dsEntity = restoreBkp.DSEntity{
						Deployment: deployment,
					}
				})
			}
		})
		Step("Running Workloads before taking backups", func() {
			for _, pdsDeployment := range deps {
				ckSum, wlDep, err := dsTest.InsertDataAndReturnChecksum(pdsDeployment, wkloadParams)
				log.FailOnError(err, "Error while Running workloads")
				wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDep)
				log.Debugf("Checksum for the deployment %s is %s", *pdsDeployment.ClusterResourceName, ckSum)
				pdsdeploymentsmd5Hash[*pdsDeployment.ClusterResourceName] = ckSum
			}
		})
		Step("Perform multiple adhoc backup and validate them", func() {
			log.Infof("Deployment ID: %v, backup target ID: %v", deployment.GetId(), bkpTarget.GetId())
			for _, pdsDeployment := range deps {
				log.Infof("Deployment ID: %v, backup target ID: %v", pdsDeployment.GetId(), bkpTarget.GetId())
				err = bkpClient.TriggerAndValidateAdhocBackup(pdsDeployment.GetId(), bkpTarget.GetId(), "s3")
				log.FailOnError(err, "Failed while performing adhoc backup")
			}

		})
		Step("Trigger restore during KVDB pods are down", func() {
			for _, deployment := range deployments {
				failuretype := pdslib.TypeOfFailure{
					Type: RestoreDSDuringKVDBFailOver,
					Method: func() error {
						return KillKvdbMasterNodeAndFailover()
					},
				}
				pdslib.DefineFailureType(failuretype)
				err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, params.ResiliencyTest.CheckTillReplica)
				log.FailOnError(err, fmt.Sprintf("Error happened while restarting px for data service %v", *deployment.ClusterResourceName))
			}
		})
		Step("Validate Deployments after KVDB POD restarts", func() {
			for ds, deployment := range deployments {
				err = dsTest.ValidateDataServiceDeployment(deployment, namespace)
				log.FailOnError(err, "Error while validating dataservices")
				log.InfoD("Data-service: %v is up and healthy", ds.Name)
			}
			dsEntity = restoreBkp.DSEntity{
				Deployment: deployment,
			}
		})
		Step("Taking adhoc backup and trigger restore again", func() {
			log.Infof("Deployment ID: %v, backup target ID: %v", deployment.GetId(), bkpTarget.GetId())
			err = bkpClient.TriggerAndValidateAdhocBackup(deployment.GetId(), bkpTarget.GetId(), "s3")
			log.FailOnError(err, "Failed while performing adhoc backup")
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
				deploymentsToBeCleaned = append(deploymentsToBeCleaned, restoredDeployment)
				log.InfoD("Restored successfully. Deployment- %v", restoredModel.GetClusterResourceName())
			}
		})
		Step("Validate md5hash for the restored deployments", func() {
			for _, pdsDeployment := range deploymentsToBeCleaned {
				ckSum, wlDep, err := dsTest.ReadDataAndReturnChecksum(pdsDeployment, wkloadParams)
				wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDep)
				log.FailOnError(err, "Error while Running workloads")
				log.Debugf("Checksum for the deployment %s is %s", *pdsDeployment.ClusterResourceName, ckSum)
				restoredDeploymentsmd5Hash[*pdsDeployment.ClusterResourceName] = ckSum
			}
			defer func() {
				for _, wlDep := range wlDeploymentsToBeCleaned {
					err := k8sApps.DeleteDeployment(wlDep.Name, wlDep.Namespace)
					log.FailOnError(err, "Failed while deleting the workload deployment")
				}
			}()
			dash.VerifyFatal(dsTest.ValidateDataMd5Hash(pdsdeploymentsmd5Hash, restoredDeploymentsmd5Hash),
				true, "Validate md5 hash after restore")
		})
		Step("Delete Deployments", func() {
			dynamicDeps := pdslib.GetDynamicDeployments()
			deploymentsToBeCleaned = append(deploymentsToBeCleaned, dynamicDeps...)
			CleanupDeployments(deploymentsToBeCleaned)

		})
	})
	JustAfterEach(func() {
		defer EndTorpedoTest()
	})
})

var _ = Describe("{StopPXDuringStorageResize}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("StopPXDuringStorageResize", "Stop PX on a node during application's storage is resized", pdsLabels, 0)
		pdslib.MarkResiliencyTC(true)
		wkloadParams = pdsdriver.LoadGenParams{
			LoadGenDepName: params.LoadGen.LoadGenDepName,
			Namespace:      params.InfraToTest.Namespace,
			NumOfRows:      params.LoadGen.NumOfRows,
			Timeout:        params.LoadGen.Timeout,
			Replicas:       params.LoadGen.Replicas,
			TableName:      "wltestingnew",
			Iterations:     params.LoadGen.Iterations,
			FailOnError:    params.LoadGen.FailOnError,
		}
	})

	It("Deploy Dataservices and Restart PX During App scaleup", func() {
		var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
		var wlDeploymentsToBeCleaned []*v1.Deployment
		var volNodesWithPx []node.Node

		Step("Deploy Data Services", func() {
			for _, ds := range params.DataServiceToTest {
				Step("Deploy and validate data service", func() {
					isDeploymentsDeleted = false
					deployment, _, _, err = DeployandValidateDataServices(ds, params.InfraToTest.Namespace, tenantID, projectID)
					log.FailOnError(err, "Error while deploying data services")
					deployments[ds] = deployment
				})
			}
		})

		defer func() {
			for _, newDeployment := range deployments {
				Step("Delete created deployments")
				resp, err := pdslib.DeleteDeployment(newDeployment.GetId())
				log.FailOnError(err, "Error while deleting data services")
				dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
				err = pdslib.DeletePvandPVCs(*newDeployment.ClusterResourceName, false)
				log.FailOnError(err, "Error while deleting PV and PVCs")
			}
		}()
		Step("Fetch Volume Nodes on which PX is Running", func() {
			volNodesWithPx = GetVolumeNodesOnWhichPxIsRunning()
			log.InfoD("volume nodes list calculated is- %v", volNodesWithPx)
		})
		Step("Stop Px on Ds Node and replica node while storage size increase", func() {
			for ds, deployment := range deployments {
				failuretype := pdslib.TypeOfFailure{
					Type: StopPXDuringStorageResize,
					Method: func() error {
						return StopPxOnReplicaVolumeNode(volNodesWithPx)
					},
				}
				pdslib.DefineFailureType(failuretype)
				//Stop PX during storage size Increase by updating the DS from "small" to "medium" template
				err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, int32(ds.ScaleReplicas))
				log.FailOnError(err, fmt.Sprintf("Error happened while stopping px for data service %v", *deployment.ClusterResourceName))

			}
		})
		Step("Start PX on the same node after volume resize", func() {
			StartPxOnReplicaVolumeNode(volNodesWithPx)
		})
		Step("Running Workloads", func() {
			for _, deployment := range deployments {
				ckSum2, wlDep, err := dsTest.InsertDataAndReturnChecksum(deployment, wkloadParams)
				log.FailOnError(err, "Error while Running workloads-%v", wlDep)
				log.Debugf("Checksum for the deployment %s is %s", *deployment.ClusterResourceName, ckSum2)
				wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDep)
			}
		})
		Step("Clean up workload deployments", func() {
			for _, wlDep := range wlDeploymentsToBeCleaned {
				err := k8sApps.DeleteDeployment(wlDep.Name, wlDep.Namespace)
				log.FailOnError(err, "Failed while deleting the workload deployment")
			}

		})
	})
	JustAfterEach(func() {
		EndTorpedoTest()

	})
})

var _ = Describe("{KillDbMasterNodeDuringStorageResize}", func() {
	JustBeforeEach(func() {
		StartTorpedoTest("KillDbMasterNodeDuringStorageResize", "Kill DB Master node during application's storage is resized", pdsLabels, 0)
		pdslib.MarkResiliencyTC(true)

		wkloadParams = pdsdriver.LoadGenParams{
			LoadGenDepName: params.LoadGen.LoadGenDepName,
			Namespace:      params.InfraToTest.Namespace,
			NumOfRows:      params.LoadGen.NumOfRows,
			Timeout:        params.LoadGen.Timeout,
			Replicas:       params.LoadGen.Replicas,
			TableName:      "wltestingnew",
			Iterations:     params.LoadGen.Iterations,
			FailOnError:    params.LoadGen.FailOnError,
		}
	})

	It("Deploy Dataservices and Restart PX During App scaleup", func() {
		var deployments = make(map[PDSDataService]*pds.ModelsDeployment)
		var wlDeploymentsToBeCleaned []*v1.Deployment

		Step("Deploy Data Services", func() {
			for _, ds := range params.DataServiceToTest {
				//This test is applicable only for SQL dbs
				if (ds.Name == postgresql) || (ds.Name == mysql) {
					//This test requires minimum of 3 replicas to be deployed
					ds.Replicas = 3
					ds.ScaleReplicas = 4
					Step("Deploy and validate data service", func() {
						isDeploymentsDeleted = false
						deployment, _, _, err = DeployandValidateDataServices(ds, params.InfraToTest.Namespace, tenantID, projectID)
						log.FailOnError(err, "Error while deploying data services")
						deployments[ds] = deployment
					})
				} else {
					log.InfoD("This testcase is valid only for SQL databases, Skipping this testcase as DB is- [%v]", ds.Name)
				}
			}
		})

		defer func() {
			for _, newDeployment := range deployments {
				Step("Delete created deployments")
				resp, err := pdslib.DeleteDeployment(newDeployment.GetId())
				log.FailOnError(err, "Error while deleting data services")
				dash.VerifyFatal(resp.StatusCode, http.StatusAccepted, "validating the status response")
				err = pdslib.DeletePvandPVCs(*newDeployment.ClusterResourceName, false)
				log.FailOnError(err, "Error while deleting PV and PVCs")
			}
		}()

		Step("Kill DB Master node during application's storage is resized", func() {
			for ds, deployment := range deployments {
				ctx, err := GetSourceClusterConfigPath()
				log.FailOnError(err, "failed while getting src cluster path")
				sourceTarget := tc.NewTargetCluster(ctx)
				log.InfoD("source target is - %v", sourceTarget)
				failuretype := pdslib.TypeOfFailure{
					Type: KillDbMasterNodeDuringStorageResize,
					Method: func() error {
						return KillDbMasterNodeDuringStorageIncrease(ds.Name, namespace, deployment, sourceTarget)
					},
				}
				pdslib.DefineFailureType(failuretype)
				//Kill DB Master Node during storage size Increase by updating the DS from "small" to "medium" template
				err = pdslib.InduceFailureAfterWaitingForCondition(deployment, namespace, int32(ds.ScaleReplicas))
				log.FailOnError(err, fmt.Sprintf("Error happened while stopping px for data service %v", *deployment.ClusterResourceName))
			}
		})
		Step("Running Workloads", func() {

			for _, deployment := range deployments {
				ckSum2, wlDep, err := dsTest.InsertDataAndReturnChecksum(deployment, wkloadParams)
				log.FailOnError(err, "Error while Running workloads-%v", wlDep)
				log.Debugf("Checksum for the deployment %s is %s", *deployment.ClusterResourceName, ckSum2)
				wlDeploymentsToBeCleaned = append(wlDeploymentsToBeCleaned, wlDep)
			}
		})
		Step("Clean up workload deployments", func() {
			for _, wlDep := range wlDeploymentsToBeCleaned {
				err := k8sApps.DeleteDeployment(wlDep.Name, wlDep.Namespace)
				log.FailOnError(err, "Failed while deleting the workload deployment")
			}

		})
	})
	JustAfterEach(func() {
		EndTorpedoTest()

	})
})
