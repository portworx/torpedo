package dataservice

import (
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
	"github.com/portworx/torpedo/drivers/unifiedPlatform"
	"github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
)

var (
	v2Components *unifiedPlatform.UnifiedPlatformComponents
	err          error
)

// InitUnifiedApiComponents
func InitUnifiedApiComponents(controlPlaneURL, accountID string) error {
	v2Components, err = unifiedPlatform.NewUnifiedPlatformComponents(controlPlaneURL, accountID)
	if err != nil {
		return err
	}
	return nil
}

// DeployDataservice should be called from workflows
func DeployDataservice(depName, nameSpace string) (*apiStructs.ApiResponse, error) {
	log.Info("Data service will be deployed as per the config map passed..")
	var depRequest pdsv2.ApiDeploymentServiceCreateDeploymentRequest

	//TODO:  Form the request body once the api's are ready for create deployment
	//var depCreateRequest pdsv2.ApiDeploymentServiceCreateDeploymentRequest
	//depCreateRequest = depCreateRequest.V1Deployment(pdsv2.V1Deployment{
	//	Meta: &pdsv2.V1Meta{
	//		Name: &deploymentName,
	//	},
	//	Config: &pdsv2.V1Config1{
	//		References:           nil,
	//		TlsEnabled:           nil,
	//		DeploymentTopologies: nil,
	//	},
	//	Status: nil,
	//})
	//
	//depCreateRequest = depCreateRequest.ApiService.DeploymentServiceCreateDeployment(ctx, namespaceId)

	deployment, err := v2Components.PDS.CreateDeployment(depRequest)
	if err != nil {
		return nil, err
	}
	return deployment, err

}

func ValidateDataServiceDeployment() {
	log.Info("Data service will be validated in this method")
}
