package dataservice

import (
	"context"
	"github.com/jinzhu/copier"
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
	"strconv"
)

var (
	v2Components     *unifiedPlatform.UnifiedPlatformComponents
	depInputs        *apiStructs.WorkFlowRequest
	depCreateRequest pdsv2.ApiDeploymentServiceCreateDeploymentRequest
	namespaceId      string
	err              error
)

type PDSDataService struct {
	DeploymentName        string "json:\"DeploymentName\""
	Name                  string "json:\"Name\""
	Version               string "json:\"Version\""
	Image                 string "json:\"Image\""
	Replicas              int    "json:\"Replicas\""
	ScaleReplicas         int    "json:\"ScaleReplicas\""
	OldVersion            string "json:\"OldVersion\""
	OldImage              string "json:\"OldImage\""
	DataServiceEnabledTLS bool   "json:\"DataServiceEnabledTLS\""
	ServiceType           string "json:\"ServiceType\""
}

// InitUnifiedApiComponents
func InitUnifiedApiComponents(controlPlaneURL, accountID string) error {
	v2Components, err = unifiedPlatform.NewUnifiedPlatformComponents(controlPlaneURL, accountID)
	if err != nil {
		return err
	}
	return nil
}

// DeployDataService should be called from workflows
func DeployDataService(ds PDSDataService) (*apiStructs.WorkFlowResponse, error) {
	log.Info("Data service will be deployed as per the config map passed..")

	// TODO call the below methods and fill up the structs
	// Get TargetClusterID
	// Get ImageID
	// Get ProjectID
	// Get App, Resource and storage Template Ids

	depCreateRequest = depCreateRequest.V1Deployment(pdsv2.V1Deployment{
		Meta: &pdsv2.V1Meta{
			Name: &ds.DeploymentName,
		},
		Config: &pdsv2.V1Config1{
			References: &pdsv2.V1References2{
				TargetClusterId: nil,
				ImageId:         nil,
				ProjectId:       nil,
				RestoreId:       nil,
			},
			TlsEnabled: &ds.DataServiceEnabledTLS,
			DeploymentTopologies: []pdsv2.V1DeploymentTopology{
				{
					Name:                     &ds.Name,
					Description:              nil,
					Replicas:                 intToPointerString(ds.Replicas),
					ServiceType:              &ds.ServiceType,
					ServiceName:              nil,
					LoadBalancerSourceRanges: nil,
					ResourceTemplate: &pdsv2.V1Template{
						Id:              nil,
						ResourceVersion: nil,
						Values:          nil,
					},
					ApplicationTemplate: &pdsv2.V1Template{
						Id:              nil,
						ResourceVersion: nil,
						Values:          nil,
					},
					StorageTemplate: &pdsv2.V1Template{
						Id:              nil,
						ResourceVersion: nil,
						Values:          nil,
					},
				},
			},
		},
	})

	//TODO: Get the namespaceID, write method to get the namespaceID from the give namespace

	depCreateRequest = depCreateRequest.ApiService.DeploymentServiceCreateDeployment(context.Background(), namespaceId)

	copier.Copy(&depInputs, depCreateRequest)

	deployment, err := v2Components.PDS.CreateDeployment(depInputs)
	if err != nil {
		return nil, err
	}
	return deployment, err

}

func ValidateDataServiceDeployment() {
	log.Info("Data service will be validated in this method")
}

func intToPointerString(n int) *string {
	// Convert the integer to a string
	str := strconv.Itoa(n)
	// Create a pointer to the string
	ptr := &str
	// Return the pointer to the string
	return ptr
}
