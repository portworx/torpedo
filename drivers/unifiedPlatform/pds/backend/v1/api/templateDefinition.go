package api

import (
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/portworx/torpedo/drivers/utilities"
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
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateDefinitionServiceGetRevision`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(tempRevisions, &tempDefResponse.ListRevision)
	if err != nil {
		return nil, err
	}
	return &tempDefResponse, nil
}

func (tempDef *PDS_API_V1) GetTemplateRevisions(revisionUid string) (*TemplateDefinitionResponse, error) {
	ctx, client, err := tempDef.getTemplateDefinitionClient()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	templateResponse := TemplateDefinitionResponse{}
	var temGetRevisionReq tempDefv1.ApiTemplateDefinitionServiceGetRevisionRequest
	temGetRevisionReq = temGetRevisionReq.ApiService.TemplateDefinitionServiceGetRevision(ctx)
	temGetRevisionReq = temGetRevisionReq.Uid(revisionUid)
	tempRevisions, res, err := client.TemplateDefinitionServiceGetRevisionExecute(temGetRevisionReq)
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `TemplateDefinitionServiceGetRevision`: %v\n.Full HTTP response: %v", err, res)
	}
	err = utilities.CopyStruct(tempRevisions, &templateResponse.GetRevision)
	if err != nil {
		return nil, err
	}
	return &templateResponse, nil
}
