package tests

//var _ = Describe("{ScaleUpDsPostStorageSizeIncreaseVariousRepl}", func() {
//	JustBeforeEach(func() {
//		StartTorpedoTest("ScaleUpDsPostStorageSizeIncreaseVariousRepl", "Scale up the DS and Perform PVC Resize, validate the updated vol in the storage config.", nil, 0)
//	})
//	var (
//		workflowDataservice    pds.WorkflowDataService
//		workFlowTemplates      pds.WorkflowPDSTemplates
//		deployment             *automationModels.PDSDeploymentResponse
//		updateDeployment       *automationModels.PDSDeploymentResponse
//		updateDeploymentScaled *automationModels.PDSDeploymentResponse
//		templates              []string
//		initialCapacity        uint64
//		increasedPvcSize       uint64
//		beforeResizePodAge     float64
//		err                    error
//	)
//	It("Perform PVC Resize and validate the updated vol in the storage config", func() {
//		Step("Create a PDS Namespace", func() {
//			Namespace = strings.ToLower("pds-test-ns-" + utilities.RandString(5))
//			WorkflowNamespace.TargetCluster = &WorkflowTargetCluster
//			workFlowTemplates.Platform = WorkflowPlatform
//			WorkflowNamespace.Namespaces = make(map[string]string)
//			workflowNamespace, err := WorkflowNamespace.CreateNamespaces(Namespace)
//			log.FailOnError(err, "Unable to create namespace")
//			log.Infof("Namespaces created - [%s]", workflowNamespace.Namespaces)
//			log.Infof("Namespace id - [%s]", workflowNamespace.Namespaces[Namespace])
//		})
//
//		for _, ds := range NewPdsParams.DataServiceToTest {
//			for _, repl := range NewPdsParams.StorageConfigurationsSSIE.ReplFactor {
//				workflowDataservice.Namespace = &WorkflowNamespace
//				NewPdsParams.StorageConfiguration.Repl = repl
//				deployment, err = workflowDataservice.DeployDataService(ds, ds.Image, ds.Version, PDS_DEFAULT_NAMESPACE)
//				log.FailOnError(err, "Error while deploying ds")
//				log.Debugf("Source Deployment Id: [%s]", *deployment.Create.Meta.Uid)
//
//				initialCapacity, _ = workflowDataservice.GetVolumeCapacityInGBForDeployment(workflowDataservice.DataServiceDeployment[*deployment.Create.Meta.Uid].Namespace, *deployment.Create.Status.CustomResourceName)
//				log.FailOnError(err, "Error while fetching pvc size for the ds")
//				log.InfoD("Initial volume storage size is : %v", initialCapacity)
//
//				beforeResizePodAge, _ = workflowDataservice.GetPodAgeForDeployment(*deployment.Create.Status.CustomResourceName)
//				//log.FailOnError(err, "unable to get pods AGE before Storage resize")
//				log.InfoD("Pods Age before storage resize is- [%v]Min", beforeResizePodAge)
//
//				defer func() {
//					Step("Delete DataServiceDeployment and Templates", func() {
//						err := workFlowTemplates.DeleteCreatedCustomPdsTemplates(templates)
//						log.FailOnError(err, "Unable to delete Custom Templates for PDS")
//
//						log.InfoD("Cleaning Up dataservice...")
//						err = workflowDataservice.DeleteDeployment(*deployment.Create.Meta.Uid)
//						log.FailOnError(err, "Error while deleting dataservice")
//					})
//				}()
//
//				// Run Workloads
//
//				log.InfoD("Increase the storage size by 1 gb through template")
//				resConfigIdUpdated, err := workFlowTemplates.CreateResourceTemplateWithCustomValue(NewPdsParams)
//
//				log.FailOnError(err, "Unable to create Custom Templates for PDS")
//				templates = append(templates, resConfigIdUpdated)
//				log.InfoD("Updated Resource Template ID- [updated- %v]", resConfigIdUpdated)
//				workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigIdUpdated
//
//				log.InfoD("Scale up the DataService with increased Storage Size")
//				updateDeployment, err = workflowDataservice.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
//				log.FailOnError(err, "Error while updating ds")
//				log.Debugf("Updated Deployment Id: [%s]", *deployment.Create.Meta.Uid)
//
//				//Verify storage size before and after storage resize - Verify at STS, PV,PVC level
//				increasedPvcSize, err = workflowDataservice.GetVolumeCapacityInGBForDeployment(workflowDataservice.DataServiceDeployment[*deployment.Create.Meta.Uid].Namespace, *deployment.Create.Status.CustomResourceName)
//				log.InfoD("Increased Storage Size is- %v", increasedPvcSize)
//
//				log.InfoD("Verify storage size before and after storage resize - Verify at STS, PV,PVC level")
//				stIncrease := workflowDataservice.ValidateStorageIncrease
//				stIncrease.UpdatedDeployment = updateDeployment
//				stIncrease.ResConfigIdUpdated = resConfigIdUpdated
//				stIncrease.InitialCapacity = initialCapacity
//				stIncrease.IncreasedStorageSize = increasedPvcSize
//				stIncrease.BeforeResizePodAge = beforeResizePodAge
//				err = workflowDataservice.ValidateDepConfigPostStorageIncrease(*deployment.Create.Meta.Uid, &stIncrease)
//				log.FailOnError(err, "Failed to validate DS Volume configuration Post Storage resize")
//
//				beforeResizePodAge2, err := workflowDataservice.GetPodAgeForDeployment(*deployment.Create.Status.CustomResourceName)
//
//				log.InfoD("Increase the storage size again after Scale-UP")
//				resConfigIdUpdatedScaled, err := workFlowTemplates.CreateResourceTemplateWithCustomValue(NewPdsParams)
//				log.FailOnError(err, "Unable to create Custom Templates for PDS")
//
//				log.InfoD("Updated Resource Template ID- [updated- %v]", resConfigIdUpdatedScaled)
//				workflowDataservice.PDSTemplates.ResourceTemplateId = resConfigIdUpdatedScaled
//
//				log.InfoD("Apply the updated template after scale up")
//				updateDeploymentScaled, err = workflowDataservice.UpdateDataService(ds, *deployment.Create.Meta.Uid, ds.Image, ds.Version)
//				log.FailOnError(err, "Error while updating ds")
//				log.Debugf("Updated Deployment Id: [%s]", *deployment.Create.Meta.Uid)
//
//				increasedPvcSizeScaleUp, err := workflowDataservice.GetVolumeCapacityInGBForDeployment(workflowDataservice.DataServiceDeployment[*deployment.Create.Meta.Uid].Namespace, *deployment.Create.Status.CustomResourceName)
//				log.InfoD("Increased Storage Size is- %v", increasedPvcSizeScaleUp)
//
//				//Verify storage size before and after storage resize - Verify at STS, PV,PVC level
//				log.InfoD("Verify storage size before and after storage resize - Verify at STS, PV,PVC level")
//				stIncreaseScaleup := workflowDataservice.ValidateStorageIncrease
//				stIncrease.UpdatedDeployment = updateDeploymentScaled
//				stIncrease.ResConfigIdUpdated = resConfigIdUpdatedScaled
//				stIncrease.InitialCapacity = increasedPvcSize
//				stIncrease.IncreasedStorageSize = increasedPvcSizeScaleUp
//				stIncrease.BeforeResizePodAge = beforeResizePodAge2
//				err = workflowDataservice.ValidateDepConfigPostStorageIncrease(*deployment.Create.Meta.Uid, &stIncreaseScaleup)
//				log.FailOnError(err, "Failed to validate DS Volume configuration Post Storage resize")
//
//			}
//		}
//	})
//	JustAfterEach(func() {
//		defer EndTorpedoTest()
//	})
//})
