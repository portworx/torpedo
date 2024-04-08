package api

import (
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
	"github.com/portworx/torpedo/pkg/log"
	tempDefv1 "github.com/pure-px/platform-api-go-client/pds/v1/catalog"
	status "net/http"
)

// ListTemplateKinds will list all tempDef kinds available for PDS
func (tempDef *PDS_API_V1) ListTemplateKinds() (*TemplateDefinitionResponse, error) {
	ctx, client, err := tempDef.getTemplateDefinitionClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	tempDefResponse := TemplateDefinitionResponse{
		ListKinds: ListTemplateKindsResponse{},
	}
	var listRequest tempDefv1.ApiTemplateDefinitionServiceListTemplateKindsRequest
	listRequest = listRequest.ApiService.TemplateDefinitionServiceListTemplateKinds(ctx)
	templatesList, res, err := client.TemplateDefinitionServiceListTemplateKindsExecute(listRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateDefinitionServiceListTemplateKinds`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(templatesList, tempDefResponse.ListKinds)
	if err != nil {
		return nil, err
	}
	return &tempDefResponse, nil
}

func (tempDef *PDS_API_V1) ListTemplateTypes() (*TemplateDefinitionResponse, error) {
	ctx, client, err := tempDef.getTemplateDefinitionClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	tempDefResponse := TemplateDefinitionResponse{
		ListTypes: ListTemplateTypesResponse{},
	}
	var listRequest tempDefv1.ApiTemplateDefinitionServiceListTemplateTypesRequest
	listRequest = listRequest.ApiService.TemplateDefinitionServiceListTemplateTypes(ctx)
	templatesList, res, err := client.TemplateDefinitionServiceListTemplateTypesExecute(listRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateDefinitionServiceListTemplateTypes`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(templatesList, tempDefResponse.ListTypes)
	if err != nil {
		return nil, err
	}
	return &tempDefResponse, nil
}

func (tempDef *PDS_API_V1) ListTemplateSamples() (*TemplateDefinitionResponse, error) {
	ctx, client, err := tempDef.getTemplateDefinitionClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	tempDefResponse := TemplateDefinitionResponse{
		ListSamples: ListTemplateSamplesResponse{},
	}
	var listRequest tempDefv1.ApiTemplateDefinitionServiceListTemplateSamplesRequest
	listRequest = listRequest.ApiService.TemplateDefinitionServiceListTemplateSamples(ctx)
	templatesList, res, err := client.TemplateDefinitionServiceListTemplateSamplesExecute(listRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateDefinitionServiceListTemplateSamples`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(templatesList, tempDefResponse.ListSamples)
	if err != nil {
		return nil, err
	}
	return &tempDefResponse, nil
}

func (tempDef *PDS_API_V1) ListTemplateRevisions() (*TemplateDefinitionResponse, error) {
	ctx, client, err := tempDef.getTemplateDefinitionClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	tempDefResponse := TemplateDefinitionResponse{
		ListRevision: ListRevisionResponse{},
	}
	var tempRevisionReq tempDefv1.ApiTemplateDefinitionServiceListRevisionsRequest
	tempRevisionReq = tempRevisionReq.ApiService.TemplateDefinitionServiceListRevisions(ctx)
	tempRevisions, res, err := client.TemplateDefinitionServiceListRevisionsExecute(tempRevisionReq)
	log.InfoD("rEVISON IS- [%v]", tempRevisions)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateDefinitionServiceGetRevision`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(tempRevisions, &tempDefResponse.ListRevision)
	if err != nil {
		return nil, err
	}
	return &tempDefResponse, nil
}

func (tempDef *PDS_API_V1) GetTemplateRevisions() (*TemplateDefinitionResponse, error) {
	ctx, client, err := tempDef.getTemplateDefinitionClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	templateResponse := TemplateDefinitionResponse{}
	templateModel, res, err := client.TemplateDefinitionServiceGetRevision(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateDefinitionServiceGetRevision`: %v\n.Full HTTP response: %v", err, res)
	}
	log.InfoD("Successfully fetched the template Roles")
	log.Infof("Value of template - [%v]", templateModel)
	err = utilities.CopyStruct(templateModel, templateResponse)
	log.Infof("Value of template after copy - [%v]", templateResponse)
	return &templateResponse, nil
}

func (tempDef *PDS_API_V1) GetTemplateTypes(tempDefinitionReq *TemplateDefinitionRequest) (*TemplateDefinitionResponse, error) {
	ctx, client, err := tempDef.getTemplateDefinitionClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	templateResponse := TemplateDefinitionResponse{}
	templateGetRequest := tempDefv1.ApiTemplateDefinitionServiceGetTemplateTypeRequest{}
	templateGetRequest = templateGetRequest.ApiService.TemplateDefinitionServiceGetTemplateType(ctx, tempDefinitionReq.GetType.Id)
	templateModel, res, err := client.TemplateDefinitionServiceGetTemplateTypeExecute(templateGetRequest)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateDefinitionServiceGetTemplateType`: %v\n.Full HTTP response: %v", err, res)
	}
	log.InfoD("Successfully fetched the template Roles")
	log.Infof("Value of template - [%v]", templateModel)
	err = utilities.CopyStruct(templateModel, templateResponse)
	log.Infof("Value of template after copy - [%v]", templateResponse)
	return &templateResponse, nil
}
