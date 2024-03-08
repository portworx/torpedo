package api

import (
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/apiStructs"
	"github.com/portworx/torpedo/pkg/log"
)

// ListTemplates return service identities models for a project.
func (template *PLATFORM_API_V1) ListTemplates(listReq *WorkFlowRequest) ([]WorkFlowResponse, error) {
	log.Infof("Value of Template - [%v]", listReq)
	return nil, nil
}

// CreateTemplates returns newly create template RoleBinding object
func (template *PLATFORM_API_V1) CreateTemplates(createReq *WorkFlowRequest) (*WorkFlowResponse, error) {

	log.Infof("Value of Template - [%v]", createReq)
	return nil, nil
}

func (template *PLATFORM_API_V1) UpdateTemplates(updateReq *WorkFlowRequest) (*WorkFlowResponse, error) {

	log.Infof("Value of Template - [%v]", updateReq)
	return nil, nil
}

// GetTemplateByID return template model.
func (template *PLATFORM_API_V1) GetTemplateByID(templateId *WorkFlowRequest) (*WorkFlowResponse, error) {

	log.Infof("Value of Template - [%v]", templateId)
	return nil, nil
}

// DeleteTemplate delete template and return status.
func (template *PLATFORM_API_V1) DeleteTemplate(templateId *WorkFlowRequest) error {
	log.Infof("Value of Template after copy - [%v]", templateId)
	return nil
}
