package pdslibs

import (
	"fmt"
	"github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	"strings"
)

const DEPLOYMENT_TOPOLOGY = "pds-qa-test-topology"

type DataServiceDetails struct {
	Deployment        automationModels.V1Deployment
	Namespace         string
	NamespaceId       string
	SourceMd5Checksum string
	DSParams          PDSDataService
}

// InitUnifiedApiComponents
func InitUnifiedApiComponents(controlPlaneURL, accountID string) error {
	v2Components, err = unifiedPlatform.NewUnifiedPlatformComponents(controlPlaneURL, accountID)
	if err != nil {
		return err
	}
	return nil
}

func GetDeploymentConfig(deploymentConfigId string) (*automationModels.PDSDeploymentResponse, error) {
	depInputs := &automationModels.PDSDeploymentRequest{
		Update: automationModels.PDSDeploymentUpdate{
			DeploymentConfigId: deploymentConfigId,
		},
	}
	deployment, err := v2Components.PDS.GetDeploymentConfig(depInputs)
	if err != nil {
		return nil, err
	}
	return deployment, err
}

func UpdateDataService(ds PDSDataService, deploymentId, namespaceId, projectId, imageId, appConfigId, resConfigId, stConfigId string) (*automationModels.PDSDeploymentResponse, error) {
	log.Info("Update Data service will be performed")
	depInputs := &automationModels.PDSDeploymentRequest{
		Update: automationModels.PDSDeploymentUpdate{
			NamespaceID:  namespaceId,
			ProjectID:    projectId,
			DeploymentID: deploymentId,
			V1Deployment: automationModels.V1DeploymentUpdate{
				Meta: automationModels.Meta{
					Name: &ds.DeploymentName,
				},
				Config: automationModels.DeploymentUpdateConfig{
					DeploymentMeta: automationModels.Meta{
						Description: StringPtr("pds-qa-tests"),
					},
					DeploymentConfig: automationModels.V1Config1{
						References: automationModels.Reference{
							ImageId: &imageId,
						},
						DeploymentTopologies: []automationModels.DeploymentTopology{
							{
								Name:     StringPtr(DEPLOYMENT_TOPOLOGY),
								Replicas: intToPointerString(ds.ScaleReplicas),
								ResourceSettings: &automationModels.PdsTemplates{
									Id: &resConfigId,
								},
								ServiceConfigurations: &automationModels.PdsTemplates{
									Id: &appConfigId,
								},
								StorageOptions: &automationModels.PdsTemplates{
									Id: &stConfigId,
								},
							},
						},
					},
				},
			},
		},
	}
	deployment, err := v2Components.PDS.UpdateDeployment(depInputs)
	if err != nil {
		return nil, err
	}
	return deployment, err
}

// DeleteDeployment Deletes the given deployment
func DeleteDeployment(deploymentId string) error {
	return v2Components.PDS.DeleteDeployment(deploymentId)
}

func GetDeployment(deploymentId string) (*automationModels.PDSDeploymentResponse, string, error) {
	deployment, err := v2Components.PDS.GetDeployment(deploymentId)
	if err != nil {
		return nil, "", err
	}
	log.Debugf("deployment [%+v]", deployment)
	pod := deployment.Get.Status.DeploymentTopologyStatus[0].ConnectionInfo.Pods[0].Name
	log.Debugf("pods [%+v]", *pod)
	podName := utilities.GetBasePodName(*pod)
	return deployment, podName, err
}

// DeployDataService Deploys the dataservices based on the given params
func DeployDataService(ds PDSDataService, namespaceId, projectId, targetClusterId, imageId, appConfigId, resConfigId, stConfigId string) (*automationModels.PDSDeploymentResponse, error) {
	log.Info("Data service will be deployed as per the config map passed..")
	depInputs := &automationModels.PDSDeploymentRequest{
		Create: automationModels.PDSDeployment{
			NamespaceID: namespaceId,
			ProjectID:   projectId,
			V1Deployment: automationModels.V1Deployment{
				Meta: automationModels.Meta{
					Name: &ds.DeploymentName,
				},
				Config: automationModels.V1Config1{
					References: automationModels.Reference{
						ImageId: &imageId,
					},
					TlsEnabled: nil,
					DeploymentTopologies: []automationModels.DeploymentTopology{
						{
							Name:        StringPtr(DEPLOYMENT_TOPOLOGY),
							Replicas:    intToPointerString(ds.Replicas),
							ServiceType: StringPtr(ds.ServiceType),
							ResourceSettings: &automationModels.PdsTemplates{
								Id: &resConfigId,
							},
							ServiceConfigurations: &automationModels.PdsTemplates{
								Id: &appConfigId,
							},
							StorageOptions: &automationModels.PdsTemplates{
								Id: &stConfigId,
							},
						},
					},
				},
			},
		},
	}

	log.Infof("deployment name  [%s]", *depInputs.Create.V1Deployment.Meta.Name)
	log.Infof("app template ids [%s]", *depInputs.Create.V1Deployment.Config.DeploymentTopologies[0].ServiceConfigurations.Id)
	log.Infof("resource template ids [%s]", *depInputs.Create.V1Deployment.Config.DeploymentTopologies[0].ResourceSettings.Id)
	log.Infof("storage template ids [%s]", *depInputs.Create.V1Deployment.Config.DeploymentTopologies[0].StorageOptions.Id)

	log.Infof("depInputs [+%v]", depInputs.Create)
	deployment, err := v2Components.PDS.CreateDeployment(depInputs)
	if err != nil {
		return nil, err
	}
	return deployment, err
}

// GetDataServiceId gets the DataService's ID
func GetDataServiceId(dsName string) (string, error) {
	ds, err := v2Components.PDS.ListDataServices()
	if err != nil {
		return "", fmt.Errorf("Failed to list DataServices: %v", err)
	}
	for _, dataService := range ds.DataServiceList {
		log.Debugf("Dataservice name: [%s]", *dataService.Meta.Name)
		if strings.Contains(strings.ToLower(strings.ReplaceAll(*dataService.Meta.Name, " ", "")), strings.ToLower(dsName)) {
			return *dataService.Meta.Uid, nil
		}
	}
	return "", fmt.Errorf("Failed to find DataService with name %s", dsName)
}

func ListDataServiceVersions(dsId string) (*automationModels.CatalogResponse, error) {
	input := automationModels.WorkFlowRequest{
		DataServiceId: dsId,
	}
	ds, err := v2Components.PDS.ListDataServiceVersions(&input)
	return ds, err
}

func ListDataServiceImages(dsId, dsVersionId string) (*automationModels.CatalogResponse, error) {
	input := automationModels.WorkFlowRequest{
		DataServiceId:        dsId,
		DataServiceVersionId: dsVersionId,
	}
	ds, err := v2Components.PDS.ListDataServiceImages(&input)
	return ds, err
}

func DeleteAllDeployments(projectId string) error {
	var numberOfDeploymentsDeleted int
	deployments, err := v2Components.PDS.ListDeployment(projectId)
	if err != nil {
		return err
	}

	if len(deployments.List) <= 0 {
		return fmt.Errorf("Deployments List is empty, No deployments to delete.\n")
	}

	for _, dep := range deployments.List {
		log.Infof("Deleting Deployment [%d]", *dep.Meta.Uid)
		err := v2Components.PDS.DeleteDeployment(*dep.Meta.Uid)
		if err != nil {
			//TODO: Check for associated backup's and delete it
			log.Infof("Error occured while deleting deployments, skipping for now: [%s]", err)
			numberOfDeploymentsDeleted -= 1
		}
		numberOfDeploymentsDeleted += 1
	}

	log.Infof("Total number of deployments Deleted [%d]", numberOfDeploymentsDeleted)
	return nil
}

// ParseInterfaceAndGetDetails takes interface as input and checks for the particular type and extracts the host and port information
// Returns the host and port as dnsEndpoints
func ParseInterfaceAndGetDetails(connectionDetails interface{}, dataServiceName string) (string, error) {
	var (
		defaultPort string
		dsNode      string
	)

	connDetailsMap, ok := connectionDetails.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Error: connectionDetails is not of type map[string]interface{}")
	}

	nodesInterface, ok := connDetailsMap["nodes"]
	if !ok {
		return "", fmt.Errorf("Error: nodes not found in connectionDetails")
	}
	nodes, ok := nodesInterface.([]interface{})
	if !ok {
		return "", fmt.Errorf("Error: nodes is not of type []interface{}")
	}

	log.Debugf("Available nodes")
	for _, nodeInterface := range nodes {
		node, err := utilities.ConvertInterfacetoString(nodeInterface)
		if err != nil {
			return "", err
		}
		log.Debugf("[%s]", node)
		if strings.Contains(node, "vip") {
			dsNode = node
		}
	}

	// Extract ports from the map
	portsInterface, ok := connDetailsMap["ports"]
	if !ok {
		return "", fmt.Errorf("Error: ports not found in connectionDetails")
	}
	ports, ok := portsInterface.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Error: ports is not of type map[string]interface{}")
	}

	log.Debugf("Available ports")
	for portName, portInterface := range ports {
		port, err := utilities.ConvertInterfacetoString(portInterface)
		if err != nil {
			return "", err
		}
		log.Debugf("[%s]:[%s]", portName, port)
		switch strings.ToLower(dataServiceName) {
		case "postgresql":
			if portName == "postgresql" {
				defaultPort = port
			}
		case "cassandra":
			if portName == "cql" {
				defaultPort = port
			}
		case "couchbase":
			if portName == "Rest" {
				defaultPort = port
			}
		case "redis":
			if portName == "client" {
				defaultPort = port
			}
		case "rabbitmq":
			if portName == "amqp" {
				defaultPort = port
			}
		case "kafka":
			if portName == "client" {
				defaultPort = port
			}
		case "elasticsearch":
			if portName == "Rest" {
				defaultPort = port
			}
		case "mongodb":
			if portName == "Mongos" {
				defaultPort = port
			}
		case "consul":
			if portName == "Http" {
				defaultPort = port
			}
		case "mysql":
			if portName == "Mysql-Router" {
				defaultPort = port
			}
		case "sqlserver":
			if portName == "Client" {
				defaultPort = port
			}
		}
	}

	dnsEndPoint := dsNode + ":" + defaultPort
	log.Debugf("DNS Endpoint [%s]", dnsEndPoint)

	if dsNode == "" || string(defaultPort) == "" {
		return "", fmt.Errorf("Node or Port value is empty..\n")
	}

	return dnsEndPoint, nil
}
