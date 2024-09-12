package pds

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/applications/databases"
	"golang.org/x/sync/errgroup"

	"github.com/portworx/sched-ops/k8s/apps"
	"github.com/portworx/torpedo/drivers/node"

	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows"

	"time"

	"github.com/portworx/torpedo/drivers/pds/parameters"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	dslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/stworkflows/platform"
	utils "github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var (
	k8sApps = apps.Instance()
	ctx     = context.Background()
)

var (
	mutex sync.Mutex
)

type WorkflowDataService struct {
	Namespace                 *platform.WorkflowNamespace
	PDSTemplates              WorkflowPDSTemplates
	DataServiceDeployment     map[string]*dslibs.DataServiceDetails
	SkipValidatation          map[string]bool
	Dash                      *aetosutil.Dashboard
	PDSParams                 *parameters.NewPDSParams
	ValidateStorageIncrease   dslibs.ValidateStorageIncrease
	UpdateDeploymentTemplates bool
	WorkloadGenParams         *dslibs.LoadGenParams
	CreateNewDatabase         bool
}

const (
	ValidatePdsDeployment      = "VALIDATE_PDS_DEPLOYMENT"
	ValidatePdsWorkloads       = "VALIDATE_PDS_WORKLOADS"
	PlatformNamespace          = "px-system"
	ValidateDeploymentDeletion = "VALIDATE_DELETE_DEPLOYMENT"
	PDS_DEPLOYMENT_AVAILABLE   = "AVAILABLE"
	postgresql                 = "PostgreSQL"
)

func (wfDataService *WorkflowDataService) DeployDataService(ds dslibs.PDSDataService, image, version string, namespace string) (*automationModels.PDSDeploymentResponse, error) {
	mutex.Lock()
	defer mutex.Unlock()

	namespaceId := wfDataService.Namespace.Namespaces[namespace]
	namespaceName := namespace
	projectId := wfDataService.Namespace.TargetCluster.Project.ProjectId
	targetClusterId := wfDataService.Namespace.TargetCluster.ClusterUID
	appConfigId := wfDataService.PDSTemplates.ServiceConfigTemplateIds[ds.Name]
	resConfigId := wfDataService.PDSTemplates.ResourceTemplateId
	stConfigId := wfDataService.PDSTemplates.StorageTemplateId
	log.Infof("targetClusterId [%s]", targetClusterId)

	log.InfoD("Deploying DataService [%s]", ds.Name)

	imageId, err := dslibs.GetDataServiceImageId(ds.Name, image, version)
	if err != nil {
		return nil, err
	}

	log.Debugf("DS Image id-[%s]", imageId)
	deployment, err := dslibs.DeployDataService(ds, namespaceId, projectId, targetClusterId, imageId, appConfigId, resConfigId, stConfigId)
	if err != nil {
		return nil, err
	}

	channel := make(chan string)

	wfDataService.DataServiceDeployment[*deployment.Create.Meta.Uid] = &dslibs.DataServiceDetails{
		Deployment:        deployment.Create,
		Namespace:         namespaceName,
		NamespaceId:       namespaceId,
		SourceMd5Checksum: "",
		DSParams:          ds,
		// Adding default database name
		// TODO: This needs to be fetched from the service configuration
		DatabaseName: "pds",
		// Adding default database user
		DatabaseUser:                  "pds",
		DatabaseControlChannel:        &channel,
		DatabaseErrorGroup:            &errgroup.Group{},
		DatabaseDriver:                nil,
		IsContinousDataSupportEnabled: false,
	}

	// Changing value of database name as per engine provided
	if value, ok := defaultDatabaseNames[ds.Name]; ok {
		wfDataService.DataServiceDeployment[*deployment.Create.Meta.Uid].DatabaseName = value
	}

	if value, ok := wfDataService.SkipValidatation[ValidatePdsDeployment]; ok {
		if value == true {
			log.Infof("Skipping DataService Deployment Validation")
		}
	} else {
		err = wfDataService.ValidatePdsDataServiceDeployments(*deployment.Create.Meta.Uid, ds, ds.Replicas, resConfigId, stConfigId, namespaceName, version, image)
		if err != nil {
			return nil, err
		}
	}

	if value, ok := wfDataService.SkipValidatation[ValidatePdsWorkloads]; ok {
		if value == true {
			log.Infof("Data validation is skipped for this")
		}
	} else {

		// TODO: This needs to be removed once below bugs are fixed:
		// https://purestorage.atlassian.net/issues/DS-9591
		// https://purestorage.atlassian.net/issues/DS-9546
		// https://purestorage.atlassian.net/issues/DS-9305
		log.Infof("Sleeping for 1 minutes to make sure deployment gets healthy")

		_, err := wfDataService.RunDataServiceWorkloads(*deployment.Create.Meta.Uid)
		if err != nil {
			return deployment, fmt.Errorf("unable to run workfload on the data service. Error - [%s]", err.Error())
		}

		// Creating the database driver
		databaseDriver, err := wfDataService.CreateDatabaseDriver(*deployment.Create.Meta.Uid, namespace, ds.Name)
		if err != nil {
			return nil, fmt.Errorf("unable to create Database driver - %s", err.Error())
		}

		// Setting the database driver for the deployment
		wfDataService.DataServiceDeployment[*deployment.Create.Meta.Uid].DatabaseDriver = databaseDriver

		if ds.RunContinuousData {
			// Starting data injection for the data service and adding to the error groups
			wfDataService.DataServiceDeployment[*deployment.Create.Meta.Uid].DatabaseErrorGroup.Go(func() error {
				err := databaseDriver.StartData(*wfDataService.DataServiceDeployment[*deployment.Create.Meta.Uid].DatabaseControlChannel, ctx)
				return err
			})

			// Setting IsContinoutDataSupportEnabled to true
			wfDataService.DataServiceDeployment[*deployment.Create.Meta.Uid].IsContinousDataSupportEnabled = true
		}

		if strings.ToLower(ds.Name) == neo4j {
			results, err := wfDataService.ImportDataToNeo4jDB(*deployment.Create.Meta.Uid, namespace, stworkflows.CrediCardDataSetSrcPath,
				stworkflows.CrediCardDataSetDestPath, stworkflows.CreditCardQuerySrcPath, stworkflows.CreditCardQueryDestPath, "CONDUCTED_AT")
			if err != nil {
				return nil, fmt.Errorf("error while importing data to db: %v", err)
			}
			wfDataService.Dash.VerifyFatal(results[0], "14446", "validating the relationship count after importing the dataset")
		}
	}

	return deployment, nil
}

// CreateDatabaseDriver creates database driver for the given deployment
func (wfDataService *WorkflowDataService) CreateDatabaseDriver(deploymentId string, namespace string, dsName string) (databases.DatabaseDriver, error) {

	var databaseDriver databases.DatabaseDriver

	// Checking if data injection support is enabled for the data service
	if slices.Contains(DATASERVICEWITHQUERYSUPPORT, strings.ToLower(dsName)) {
		// Fetching deployment details for connection info
		dataServiceDetails, err := wfDataService.GetDeployment(deploymentId)
		if err != nil {
			return nil, fmt.Errorf("Unable to get deployment details - %s", err.Error())
		}

		wfDataService.DataServiceDeployment[deploymentId].Deployment = dataServiceDetails

		// Setting the selector for sts
		idForSelector := strings.Split(deploymentId, ":")[1]

		// Selector for the nodeport service
		selectors := map[string]string{
			"pds/deployment-id": idForSelector,
		}

		log.Infof("Selector - [%v]", selectors)

		// Fetching the deployment port
		dataServicePort, err := utils.GetDeploymentPort(dataServiceDetails.Status.ConnectionInfo["connectionDetails"], dsName)
		if err != nil {
			return nil, fmt.Errorf("Unable to get port details - %s", err.Error())
		}

		// Converting dataservice port to int
		dataServicePortToInt, err := strconv.Atoi(dataServicePort)
		if err != nil {
			return nil, fmt.Errorf("Unable to convert port details to int - %s", err.Error())
		}

		// Deploying Node port service for the deployment
		// TODO: This needs to be converted to DNS once the load balancer is added to aetos
		log.Infof("Deploying node port service for the deployment")
		svcDetails, err := utils.CreateNodePortServiceForDeployment(*dataServiceDetails.Status.CustomResourceName, namespace, selectors, dataServicePort)
		if err != nil {
			return nil, fmt.Errorf("Unable to get node port details - %s", err.Error())
		}

		// Fetching the data service password
		dataServicePassword, err := wfDataService.GetDeploymentPassword(*dataServiceDetails.Meta.Uid)
		if err != nil {
			return nil, fmt.Errorf("Unable to get deployment password - %s", err.Error())
		}

		// Creating hostname for the data service
		hostName, err := utils.CreateHostNameForApp(*dataServiceDetails.Status.CustomResourceName, svcDetails.Spec.Ports[0].NodePort, namespace)
		if err != nil {
			return nil, fmt.Errorf("Unable to get deployment hostname - %s", err.Error())
		}

		log.Debugf("dataServicePassword [%s]", dataServicePassword)
		log.Debugf("hostName [%s]", hostName)

		// Creating database driver
		databaseDriver, err = databases.GetDatabaseDriver(
			strings.ToLower(dsName),
			hostName,
			wfDataService.DataServiceDeployment[deploymentId].DatabaseUser,
			dataServicePassword,
			dataServicePortToInt,
			wfDataService.DataServiceDeployment[deploymentId].DatabaseName,
			int(svcDetails.Spec.Ports[0].NodePort),
			namespace,
			"",
			nil,
			*wfDataService.DataServiceDeployment[deploymentId].Deployment.Status.CustomResourceName)

		if err != nil {
			return nil, err
		}

	}

	return databaseDriver, nil

}

func (wfDataService *WorkflowDataService) UpdateDataService(ds dslibs.PDSDataService, deploymentId, image, version string) (*automationModels.PDSDeploymentResponse, error) {
	var (
		deployment     *automationModels.PDSDeploymentResponse
		tmpResConfigId string
	)

	namespaceId := wfDataService.DataServiceDeployment[deploymentId].NamespaceId
	namespaceName := wfDataService.DataServiceDeployment[deploymentId].Namespace
	projectId := wfDataService.Namespace.TargetCluster.Project.ProjectId
	targetClusterId := wfDataService.Namespace.TargetCluster.ClusterUID
	appConfigId := wfDataService.PDSTemplates.ServiceConfigTemplateIds[ds.Name]
	resConfigId := wfDataService.PDSTemplates.ResourceTemplateId
	stConfigId := wfDataService.PDSTemplates.StorageTemplateId
	newResConfigId := wfDataService.PDSTemplates.UpdateResourceTemplateId
	log.Infof("targetClusterId [%s]", targetClusterId)

	imageId, err := dslibs.GetDataServiceImageId(ds.Name, image, version)
	if err != nil {
		return nil, err
	}

	if wfDataService.UpdateDeploymentTemplates {
		log.Debugf("newResConfigId [%s]", newResConfigId)
		tmpResConfigId = newResConfigId
		deployment, err = dslibs.UpdateDataService(ds, deploymentId, namespaceId, projectId, imageId, appConfigId, newResConfigId, stConfigId)
	} else {
		tmpResConfigId = resConfigId
		deployment, err = dslibs.UpdateDataService(ds, deploymentId, namespaceId, projectId, imageId, appConfigId, resConfigId, stConfigId)
	}
	if err != nil {
		return nil, err
	}

	log.Infof("Sleeping for 1 minutes to make sure deployment gets updated")
	time.Sleep(1 * time.Minute)

	log.Debugf("Validate Deployment CR is updated successfully")
	//Validate the deploymentConfig update status
	err = dslibs.ValidateDeploymentConfigUpdate(*deployment.Update.Meta.Uid, "COMPLETED")
	if err != nil {
		return nil, err
	}

	log.Debugf("Updated Deployment [%+v]", deployment)
	if value, ok := wfDataService.SkipValidatation[ValidatePdsDeployment]; ok {
		if value == true {
			log.Infof("Skipping Validation")
		}
	} else {
		err = wfDataService.ValidatePdsDataServiceDeployments(*deployment.Update.Config.DataServiceDeploymentMeta.Uid, ds, ds.ScaleReplicas, tmpResConfigId, stConfigId, namespaceName, version, image)
		if err != nil {
			return nil, err
		}
	}

	return deployment, nil
}

// ValidateDataServiceDeploymentHealth validates deployment health for PDS
func (wfDataService *WorkflowDataService) WaitForDeploymentToBeAvailable(deploymentId string) error {
	// Validate the sts object and health of the pds deployment
	err := dslibs.ValidateDataServiceDeploymentHealth(deploymentId, PDS_DEPLOYMENT_AVAILABLE)
	return err
}

// ValidatePdsDataServiceDeployments validates the pds deployments resource, storage, deployment configurations and endpoints
func (wfDataService *WorkflowDataService) ValidatePdsDataServiceDeployments(deploymentId string, ds dslibs.PDSDataService, replicas int, resConfigId, stConfigId, namespace, version, image string) error {

	// Validate the sts object and health of the pds deployment
	err := dslibs.ValidateDataServiceDeploymentHealth(deploymentId, PDS_DEPLOYMENT_AVAILABLE)
	if err != nil {
		return err
	}

	//Validate Statefulset health
	err = dslibs.ValidateStatefulSetHealth(*wfDataService.DataServiceDeployment[deploymentId].Deployment.Status.CustomResourceName, wfDataService.DataServiceDeployment[deploymentId].Namespace)
	if err != nil {
		return err
	}

	if !ds.EnableTLS {
		// Validate if the dns endpoint is reachable
		//TODO: Validate the dns endpoint with cert

		err = wfDataService.ValidateDNSEndpoint(deploymentId)
		if err != nil {
			return err
		}
	} else {
		resp, err := dslibs.GetDeployment(deploymentId)
		if err != nil {
			return err
		}
		log.Infof("Is TLS enabled for DataService Deployment: [%v]", resp.Get.Config.TlsConfig.Enabled)
		wfDataService.Dash.VerifyFatal(resp.Get.Config.TlsConfig.Enabled, ds.EnableTLS, "Validating if TLS is enabled on the dataService")
	}

	// Get data service deployment resources
	resourceTemplateOps, storageOps, DeploymentConfigs, err := wfDataService.GetDsDeploymentResources(deploymentId, resConfigId)
	if err != nil {
		return err
	}

	// Validate deployment resources
	dataServiceVersionBuild := version + "-" + image
	wfDataService.ValidateDeploymentResources(resourceTemplateOps, storageOps, DeploymentConfigs, replicas, dataServiceVersionBuild)

	return nil
}

func (wfDataService *WorkflowDataService) ExportDataFromRelationalDB(identifier, deploymentId, namespace string) error {
	var dataCommands = make(map[databases.DatabaseDriver]map[string][]string)
	//var tableName string
	filePath := "/srv/pds/test_table_export.csv"
	dbDriver := wfDataService.DataServiceDeployment[deploymentId].DatabaseDriver

	// Updating SQL commmands with current identifier
	dbDriver.UpdateDataCommands(20, identifier)

	//Inserting data to db
	dataCommands[dbDriver] = dbDriver.GetRandomDataCommands(10)
	err := dbDriver.InsertBackupData(ctx, identifier, nil)
	if err != nil {
		return err
	}

	err = dbDriver.CopyBackupData(ctx, identifier)
	if err != nil {
		return err
	}

	//copying the csv file to the local file system from db pod
	_, podToExportData, err := dslibs.GetDeploymentAndPodDetails(deploymentId)
	if err != nil {
		return err
	}
	podToExportData = podToExportData + "-0"
	log.Debugf("podToImportData [%s]", podToExportData)
	err = utils.CopyFileFromPod(namespace, podToExportData, filePath, stworkflows.TestDataSetSrcPath)
	if err != nil {
		return err
	}

	return nil
}

func (wfDataService *WorkflowDataService) ImportDataToNeo4jDB(deploymentId, namespace, DataSetSrcPath, DataSetDestPath, CqlFileSrcPath, CqlFileDestPath, relationship string) ([]string, error) {
	var cmd string
	var results []string

	_, podToImportData, err := dslibs.GetDeploymentAndPodDetails(deploymentId)
	if err != nil {
		return nil, err
	}

	podToImportData = podToImportData + "-0"
	log.Debugf("podToImportData [%s]", podToImportData)

	password, err := wfDataService.GetDeploymentPassword(deploymentId)
	if err != nil {
		return nil, fmt.Errorf("error while executing the password: %v", err)
	}

	//copy dataset to the pod
	err = utils.CopyFileToPod(namespace, podToImportData, DataSetSrcPath, DataSetDestPath)
	if err != nil {
		return nil, fmt.Errorf("error while copying the file to the pod %v", err)
	}

	//copy the .cql files to the pod
	err = utils.CopyFileToPod(namespace, podToImportData, CqlFileSrcPath, CqlFileDestPath)
	if err != nil {
		return nil, fmt.Errorf("error while copying the file to the pod %v", err)
	}

	cmd = fmt.Sprintf("cypher-shell -u pds -p %s -f %s", password, CqlFileDestPath)
	_, err = utils.ExecCommandInPod(podToImportData, namespace, cmd, wfDataService.Namespace.TargetCluster.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error while executing the command: %v", err)
	}

	//sleep time to make sure the imported data gets reflected in the db
	time.Sleep(20 * time.Second)

	cmd = fmt.Sprintf("MATCH ()-[r:%s]->() RETURN COUNT(r) AS RelationshipCount;", relationship)
	results, err = wfDataService.DataServiceDeployment[deploymentId].DatabaseDriver.ExecuteCommand([]string{cmd}, ctx)
	if err != nil {
		return nil, fmt.Errorf("error while executing the query: %v", err)
	}

	for _, result := range results {
		log.Debugf("values:[%s]", result)
	}

	return results, nil
}

func (wfDataService *WorkflowDataService) GetDsDeploymentResources(deploymentId, resConfigId string) (dslibs.ResourceSettingTemplate, dslibs.StorageOps, dslibs.DeploymentConfig, error) {
	var (
		resourceTemp dslibs.ResourceSettingTemplate
		storageOp    dslibs.StorageOps
		dbConfig     dslibs.DeploymentConfig
		err          error
	)

	time.Sleep(30 * time.Second)

	deployment, podName, err := dslibs.GetDeploymentAndPodDetails(deploymentId)
	if err != nil {
		return resourceTemp, storageOp, dbConfig, err
	}

	log.Debugf("Crd Name: [%s]", dslibs.CrdMap[strings.ToLower(wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)])
	dbConfig, err = dslibs.GetDeploymentConfigurations(wfDataService.DataServiceDeployment[deploymentId].Namespace, dslibs.CrdMap[strings.ToLower(wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)], podName)
	if err != nil {
		return resourceTemp, storageOp, dbConfig, err
	}

	log.Debugf("Resource Template Id After Update [%s]", resConfigId)
	resourceTemp, err = dslibs.GetResourceTemplateConfigs(resConfigId)
	if err != nil {
		return resourceTemp, storageOp, dbConfig, err
	}

	storageOp, err = dslibs.GetStorageTemplateConfigs(*deployment.Get.Config.DataServiceDeploymentTopologies[0].StorageOptions.Id)
	if err != nil {
		return resourceTemp, storageOp, dbConfig, err
	}

	return resourceTemp, storageOp, dbConfig, err

}

func (wfDataService *WorkflowDataService) DeleteDeployment(deploymentId string) error {
	err := dslibs.DeleteDeployment(deploymentId)
	if err != nil {
		return err
	}
	if value, ok := wfDataService.SkipValidatation[ValidateDeploymentDeletion]; ok {
		if value == true {
			log.Infof("Skipping validation of dataservice deletion")
		}
	} else {
		err = dslibs.ValidateDeploymentIsDeleted(deploymentId)
		if err != nil {
			return err
		}
	}

	// Removing the data service entry from map
	delete(wfDataService.DataServiceDeployment, deploymentId)

	return nil
}

func (wfDataService *WorkflowDataService) ValidateDNSEndpoint(deploymentId string) error {
	deployment, err := dslibs.GetDeployment(deploymentId)
	if err != nil {
		return err
	}
	log.Infof("Deployment Response [+%v]", *deployment)
	log.Infof("ConnectionInfo Response [+%v]", deployment.Get.Status.ConnectionInfo["connectionDetails"])

	connectionDetails := deployment.Get.Status.ConnectionInfo["connectionDetails"]
	dnsEndPoint, err := utils.ParseInterfaceAndGetDetails(connectionDetails, wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)
	if err != nil {
		return err
	}

	err = dslibs.ValidateDNSEndPoint(dnsEndPoint)
	if err != nil {
		return err
	}

	return nil
}

// GetDNSEndpoint returns the dns endpoint for the given deployment
func (wfDataService *WorkflowDataService) GetDNSEndpoint(deploymentId string) (string, error) {
	deployment, err := dslibs.GetDeployment(deploymentId)
	if err != nil {
		return "", err
	}
	log.Infof("Deployment Response [+%v]", *deployment)
	log.Infof("ConnectionInfo Response [+%v]", deployment.Get.Status.ConnectionInfo["connectionDetails"])

	connectionDetails := deployment.Get.Status.ConnectionInfo["connectionDetails"]
	dnsEndPoint, err := utils.ParseInterfaceAndGetDetails(connectionDetails, wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)
	if err != nil {
		return "", err
	}

	return dnsEndPoint, nil
}

func (wfDataService *WorkflowDataService) RunDataServiceWorkloads(deploymentId string) (map[string]string, error) {
	chkSumMap := make(map[string]string)

	//Initializing the tableName and clusterMode
	wfDataService.WorkloadGenParams.TableName = "wltesting" + utils.RandomString(3)
	wfDataService.WorkloadGenParams.Namespace = wfDataService.DataServiceDeployment[deploymentId].Namespace
	dsInstances := *wfDataService.DataServiceDeployment[deploymentId].Deployment.Config.DataServiceDeploymentTopologies[0].Instances

	log.Debugf("dsInstances [%s]", dsInstances)
	log.Debugf("dsName [%s]", wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)

	if wfDataService.DataServiceDeployment[deploymentId].DSParams.Name == "Redis" && dsInstances == "1" {
		wfDataService.WorkloadGenParams.ClusterMode = "false"
	} else {
		log.Infof("Enabling cluster mode to read/write data for redis data service")
		wfDataService.WorkloadGenParams.ClusterMode = "true"
	}

	if slices.Contains(stworkflows.SKIPDATASERVICEFROMWORKLOAD, strings.ToLower(wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)) {
		log.Warnf("Workload is not enabled for this - [%s] - data service", wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)
		return nil, nil
	}

	chkSum, wlDep, err := dslibs.InsertDataAndReturnChecksum(*wfDataService.DataServiceDeployment[deploymentId], *wfDataService.WorkloadGenParams)
	if err != nil {
		return nil, err
	}

	// Updating the hash for the tableName
	chkSumMap[wfDataService.WorkloadGenParams.TableName] = chkSum

	// Updating the data hash for the deployment
	wfDataService.DataServiceDeployment[deploymentId].SourceMd5Checksum = chkSum

	return chkSumMap, dslibs.DeleteWorkloadDeployments(wlDep)
}

// Reads and update the md5 hash for the data service
func (wfDataService *WorkflowDataService) ReadAndUpdateDataServiceDataHash(deploymentId string) error {
	if slices.Contains(stworkflows.SKIPDATASERVICEFROMWORKLOAD, strings.ToLower(wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)) {
		log.Warnf("Workload is not enabled for this - [%s] - data service", wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)
		return nil
	}

	chkSum, _, err := dslibs.ReadDataAndReturnChecksum(
		*wfDataService.DataServiceDeployment[deploymentId],
		wfDataService.DataServiceDeployment[deploymentId].DSParams.Name,
		dslibs.CrdMap[strings.ToLower(wfDataService.DataServiceDeployment[deploymentId].DSParams.Name)],
		*wfDataService.WorkloadGenParams,
	)

	if err != nil {
		return err
	}

	wfDataService.DataServiceDeployment[deploymentId].SourceMd5Checksum = chkSum

	return nil
}

// TODO: Commenting this methods out, this needs to be refatcored as per current design
//func (wfDataService *WorkflowDataService) ValidateDataServiceWorkloads(params *parameters.NewPDSParams, restoredDeployment *automationModels.PDSRestoreResponse) error {
//	//Initializing the parameters required for workload generation
//	wkloadParams := dslibs.LoadGenParams{
//		LoadGenDepName: params.LoadGen.LoadGenDepName,
//		Namespace:      params.InfraToTest.Namespace,
//		NumOfRows:      params.LoadGen.NumOfRows,
//		Timeout:        params.LoadGen.Timeout,
//		Replicas:       params.LoadGen.Replicas,
//		TableName:      params.LoadGen.TableName,
//		Iterations:     params.LoadGen.Iterations,
//		FailOnError:    params.LoadGen.FailOnError,
//	}
//
//	deployment := make(map[string]string)
//	deployment[*restoredDeployment.Create.Meta.Name] = *restoredDeployment.Create.Meta.Uid
//	// chkSum, wlDep, err := dslibs.ReadDataAndReturnChecksum(deployment, wkloadParams)
//	_, wlDep, err := dslibs.ReadDataAndReturnChecksum(deployment, wkloadParams)
//	if err != nil {
//		return err
//	}
//
//	// deploymentName, _ := GetDeploymentNameAndId(deployment)
//	_, _ = GetDeploymentNameAndId(deployment)
//
//	// TODO: Commenting this out for now, this needs to be refactored as per current design
//	//wfDataService.RestoredDeploymentMd5Hash[deploymentName] = chkSum
//	//
//	//result := dslibs.ValidateDataMd5Hash(wfDataService.SourceDeploymentMd5Hash, wfDataService.RestoredDeploymentMd5Hash)
//	//wfDataService.Dash.VerifyFatal(result, true, "Validate md5 hash after restore")
//
//	return dslibs.DeleteWorkloadDeployments(wlDep)
//}

func (wfDataService *WorkflowDataService) GetDeployment(deploymentId string) (automationModels.V1Deployment, error) {
	dep, err := dslibs.GetDeployment(deploymentId)
	if err != nil {
		return automationModels.V1Deployment{}, err
	}
	return dep.Get, nil
}

func (wfDataService *WorkflowDataService) ValidateDeploymentResources(resourceTemp dslibs.ResourceSettingTemplate, storageOp dslibs.StorageOps, config dslibs.DeploymentConfig, replicas int, dataServiceVersionBuild string) {
	log.Debugf("filesystem used %v ", config.Spec.Topologies[0].StorageOptions.Filesystem)
	log.Debugf("storage replicas used %v ", config.Spec.Topologies[0].StorageOptions.Replicas)
	log.Debugf("cpu requests used %v ", config.Spec.Topologies[0].Resources.Requests.CPU)
	log.Debugf("memory requests used %v ", config.Spec.Topologies[0].Resources.Requests.Memory)
	log.Debugf("storage requests used %v ", config.Spec.Topologies[0].Resources.Requests.Storage)
	log.Debugf("No of nodes requested %v ", config.Spec.Topologies[0].Nodes)
	log.Debugf("volume group %v ", storageOp.VolumeGroup)
	log.Debugf("resource template values cpu req [%s]", resourceTemp.Resources.Requests.CPU)

	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].Resources.Requests.CPU, resourceTemp.Resources.Requests.CPU, "Validating CPU Request")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].Resources.Requests.Memory, resourceTemp.Resources.Requests.Memory, "Validating Memory Request")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].Resources.Requests.Storage, resourceTemp.Resources.Requests.Storage, "Validating storage Request")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].Resources.Limits.CPU, resourceTemp.Resources.Limits.CPU, "Validating CPU Limits")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].Resources.Limits.Memory, resourceTemp.Resources.Limits.Memory, "Validating Memory Limits")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].StorageOptions.Replicas, storageOp.Replicas, "Validating storage replicas")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].StorageOptions.Filesystem, storageOp.Filesystem, "Validating filesystems")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].StorageOptions.Secure, storageOp.Secure, "Validating Secure Storage Option")
	wfDataService.Dash.VerifyFatal(config.Spec.Topologies[0].Nodes, replicas, "Validating ds node replicas")
	wfDataService.Dash.VerifyFatal(config.Spec.Version, dataServiceVersionBuild, "Validating ds version")
}

func (wfDataService *WorkflowDataService) KillDBMasterNodeToValidateHA(deploymentId string) error {

	dbMaster, isNativelyDistributed := utils.GetDbMasterNode(wfDataService.DataServiceDeployment[deploymentId].Namespace,
		wfDataService.DataServiceDeployment[deploymentId].DSParams.Name,
		*wfDataService.DataServiceDeployment[deploymentId].Deployment.Status.CustomResourceName,
		wfDataService.Namespace.TargetCluster.KubeConfig)

	if isNativelyDistributed {
		err := utils.DeleteK8sPods(dbMaster, wfDataService.DataServiceDeployment[deploymentId].Namespace, wfDataService.Namespace.TargetCluster.KubeConfig)
		if err != nil {
			return err
		}
	} else {
		podName, err := utils.GetAnyPodName(*wfDataService.DataServiceDeployment[deploymentId].Deployment.Status.CustomResourceName, wfDataService.DataServiceDeployment[deploymentId].Namespace)
		if err != nil {
			return fmt.Errorf("failed while fetching pod for stateful set %v ", *wfDataService.DataServiceDeployment[deploymentId].Deployment.Meta.Name)
		}
		err = utils.KillPodsInNamespace(wfDataService.DataServiceDeployment[deploymentId].Namespace, podName)
		if err != nil {
			return fmt.Errorf("failed while deleting pod %v ", *wfDataService.DataServiceDeployment[deploymentId].Deployment.Meta.Name)
		}
	}
	return nil
}

func (wfDataService *WorkflowDataService) GetDbMasterNode(deploymentId string) (string, bool) {
	dbMaster, isNativelyDistributed := utils.GetDbMasterNode(
		wfDataService.DataServiceDeployment[deploymentId].Namespace,
		wfDataService.DataServiceDeployment[deploymentId].DSParams.Name,
		*wfDataService.DataServiceDeployment[deploymentId].Deployment.Status.CustomResourceName,
		wfDataService.Namespace.TargetCluster.KubeConfig)
	return dbMaster, isNativelyDistributed
}

func (wfDataService *WorkflowDataService) DeletePDSPods(podNames []string, namespace string) error {
	pdsPods := make([]corev1.Pod, 0)

	podList, err := utils.GetPods(namespace)
	if err != nil {
		return fmt.Errorf("Error while getting pods: %v", err)
	}
	log.Infof("PDS System Pods")
	for _, pod := range podList.Items {
		for _, name := range podNames {
			if strings.Contains(strings.ToLower(pod.Name), strings.ToLower(name)) {
				log.Infof("%v", pod.Name)
				pdsPods = append(pdsPods, pod)
				break // Once found, break out of the inner loop
			}
		}
	}

	log.InfoD("Deleting PDS System Pods")
	err = utils.DeletePods(pdsPods)
	if err != nil {
		return fmt.Errorf("Error while deleting pods: %\v ", err)
	}
	time.Sleep(90 * time.Second)
	log.InfoD("Validating PDS System Pods")
	err = utils.ValidatePods(namespace, "")
	if err != nil {
		return fmt.Errorf("Error while validating pods: %v", err)
	}
	return nil
}

func (wfDataService *WorkflowDataService) GetPodAgeForDeployment(deploymentId string) (float64, error) {
	age, err := utils.GetPodAge(
		*wfDataService.DataServiceDeployment[deploymentId].Deployment.Status.CustomResourceName,
		wfDataService.DataServiceDeployment[deploymentId].Namespace,
	)
	if err != nil {
		return 0, err
	}
	return age, nil
}

// ValidateReplicas validates replica for AAG
func (wfDataService *WorkflowDataService) ValidateSQLAAGReplicas(deploymentId string) (databases.ReplicaDetails, error) {
	// Fetching deployment details
	deploymentDetails := wfDataService.DataServiceDeployment[deploymentId]

	var replicaDetails databases.ReplicaDetails
	var err error

	validateReplicaState := func() (interface{}, bool, error) {
		replicaDetails, err = deploymentDetails.DatabaseDriver.GetReplicaDetails(ctx)
		if err != nil {
			return nil, false, err
		}

		log.Infof("Validating [%s], Current Primary Replica - [%s]", replicaDetails.AGName, replicaDetails.PrimaryReplica)

		for _, replica := range replicaDetails.Replicas {
			if replica.ConnectionState != "CONNECTED" {
				return nil, true, fmt.Errorf("replica [%s] is not in running state", replica.Name)
			}

			if replica.Health != "HEALTHY" {
				return nil, true, fmt.Errorf("replica [%s] is not in healthy state", replica.Name)
			}
		}

		return nil, false, nil
	}

	_, err = task.DoRetryWithTimeout(validateReplicaState, replicaHealthTimeout, replicaHealthInterval)

	return replicaDetails, err
}

// Purge will delete all dataservice and associated PVCs from the cluster
func (wfDataService *WorkflowDataService) Purge(ignoreError bool) error {

	var errors []string
	var namespace []string

	errors = make([]string, 0)

	for dsId, dsDetails := range wfDataService.DataServiceDeployment {

		// If continous pipeline is running, stopping and checking continous data sanity
		if dsDetails.IsContinousDataSupportEnabled == true {
			log.Infof("Stopping the continous pipeline")
			*dsDetails.DatabaseControlChannel <- DataStop

			// Closing the control channel
			close(*dsDetails.DatabaseControlChannel)

			// Fetching errors from the error group
			if err := dsDetails.DatabaseErrorGroup.Wait(); err != nil {
				errors = append(errors, fmt.Sprintf("errror occurred while fetching data from %s, error - [%s]", *dsDetails.Deployment.Meta.Name, err.Error()))
			}
		}

		if !slices.Contains(namespace, dsDetails.Namespace) {
			namespace = append(namespace, dsDetails.Namespace)
		}

		dsName := *wfDataService.DataServiceDeployment[dsId].Deployment.Meta.Name
		log.Infof("Deleting [%s] with id [%s] from [%s]-[%s] namespace ", dsName, dsId, dsDetails.Namespace, dsDetails.NamespaceId)

		deploymentDetails, err := dslibs.GetDeployment(dsId)
		if err != nil {
			log.Warnf("Unable to fetch details for [%s]. Error - [%s]", dsName, err.Error())
			if !ignoreError {
				errors = append(errors, err.Error())
			}
			continue
		}

		err = wfDataService.DeleteDeployment(*deploymentDetails.Get.Meta.Uid)
		if err != nil {
			log.Warnf("Unable to delete [%s]. Error - [%s]", dsName, err.Error())
			if !ignoreError {
				errors = append(errors, err.Error())
			}
			continue
		} else {
			log.Infof("[%s] deleted successfully", dsName)
		}

		err = utils.DeletePvandPVCs(*deploymentDetails.Get.Status.CustomResourceName, false)

		if err != nil {
			log.Warnf("Unable to delete PVs for [%s]. Error - [%s]", dsName, err.Error())
			if !ignoreError {
				errors = append(errors, err.Error())
			}
			continue
		} else {
			log.Infof("All PVs associated with [%s] deleted successfully", dsName)
		}

	}

	for _, ns := range namespace {
		err := utils.RemoveFinalizersFromAllResources(ns)
		if err != nil {
			log.Warnf("Unable to remove finalizers. Error - [%s]", err.Error())
			if !ignoreError {
				errors = append(errors, err.Error())
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("Below errors occurred during deployment cleanup.\n\n [%s]", strings.Join(errors, "\n"))
	}

	return nil
}

func (wfDataService *WorkflowDataService) RunStress(deploymentID, deploymentName, dsName, namespace string) (*corev1.Pod, *appsv1.Deployment, error) {
	var (
		pod *corev1.Pod
		dep *appsv1.Deployment
		err error
	)
	params := &dslibs.WorkloadGenerationParams{
		DataServiceName: dsName,
		DeploymentID:    deploymentID,
		DeploymentName:  deploymentName,
		Namespace:       namespace,
	}

	//params.DataServiceName = wfDataService.DataServiceDeployment[deploymentID].DSParams.Name

	log.Infof("Dataservice Name : %s", params.DataServiceName)

	if params.DataServiceName == postgresql {
		params.DeploymentName = "pgload"
		params.ScaleFactor = "100"
		params.Iterations = "1"

		log.Infof("Running Workloads on DataService %v ", params.DataServiceName)
		pod, dep, err = wfDataService.CreateDataServiceWorkloads(params)
	}
	return pod, dep, err
}

// CreateDataServiceWorkloads generates workloads for the given data services
func (wfDataService *WorkflowDataService) CreateDataServiceWorkloads(params *dslibs.WorkloadGenerationParams) (*corev1.Pod, *appsv1.Deployment, error) {
	var dep *appsv1.Deployment
	var pod *corev1.Pod

	dnsEndPoint := ""
	deployment, err := dslibs.GetDeployment(params.DeploymentID)
	if err != nil {
		return nil, nil, err
	}

	connectionDetails := deployment.Get.Status.ConnectionInfo["connectionDetails"]
	dnsEndPoint, err = utils.ParseInterfaceAndGetDetails(connectionDetails, params.DataServiceName)
	if err != nil {
		return nil, nil, err
	}

	err = dslibs.ValidateDNSEndPoint(dnsEndPoint)
	if err != nil {
		return nil, nil, fmt.Errorf("error occured while validating connection info, Err: %v", err)
	}
	log.Infof("DNS endpoints are reachable...")

	pdsPassword, err := dslibs.GetDeploymentCredentials(params.DeploymentID)
	if err != nil {
		return nil, nil, fmt.Errorf("error occured while getting credentials info, Err: %v", err)
	}

	switch params.DataServiceName {
	case postgresql:
		endPoint := strings.Split(dnsEndPoint, ":")[0]
		dep, err = utils.CreatePostgresqlWorkload(endPoint, pdsPassword, params.ScaleFactor, params.Iterations, params.DeploymentName, params.Namespace)
		if err != nil {
			return nil, nil, fmt.Errorf("error occured while creating postgresql workload, Err: %v", err)
		}

	}
	return pod, dep, nil
}

// Get Deployment credential for the data base
func (wfDataService *WorkflowDataService) GetDeploymentPassword(deploymentId string) (string, error) {
	cred, err := dslibs.GetDeploymentCredentials(deploymentId)
	if err != nil {
		return "", err
	}
	return cred, nil
}

// GetAllDataServiceHostedNodes returns all the nodes where the data service is hosted
func (wfDataService *WorkflowDataService) GetAllDataServiceHostedNodes(deploymentId string, namespace string) ([]node.Node, error) {
	var dsNodes []string
	var dsNodeList []node.Node
	ss, err := k8sApps.GetStatefulSet(*wfDataService.DataServiceDeployment[deploymentId].Deployment.Status.CustomResourceName, namespace)
	if err != nil {
		return nil, err
	}
	pods, err := k8sApps.GetStatefulSetPods(ss)
	if err != nil {
		return nil, err
	}
	for _, pod := range pods {
		nodeName := pod.Spec.NodeName
		dsNodes = append(dsNodes, nodeName)
	}
	for _, currNode := range node.GetWorkerNodes() {
		for _, dsNode := range dsNodes {
			if currNode.Name == dsNode {
				dsNodeList = append(dsNodeList, currNode)
			}
		}
	}

	return dsNodeList, nil
}
