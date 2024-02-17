package dataservice

import (
	"context"
	"github.com/jinzhu/copier"
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
func DeployDataservice(depName, namespace, accountID string) (*apiStructs.WorkFlowResponse, error) {
	log.Info("Data service will be deployed as per the config map passed..")

	//TODO: Take the json input param and populate the depCreate request, write the required support lib func

	var depInputs *apiStructs.WorkFlowRequest //make this global var
	var namespaceId string

	//TODO:  Form the request body once the backend's are ready for create deployment
	var depCreateRequest pdsv2.ApiDeploymentServiceCreateDeploymentRequest
	depCreateRequest = depCreateRequest.V1Deployment(pdsv2.V1Deployment{
		Meta: &pdsv2.V1Meta{
			Name: &depName,
		},
		Config: &pdsv2.V1Config1{
			References:           nil,
			TlsEnabled:           nil,
			DeploymentTopologies: nil,
		},
		Status: nil,
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
