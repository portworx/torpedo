package cloud_credential

import (
	. "github.com/portworx/px-backup-api/pkg/apis/v1"
	. "github.com/portworx/torpedo/drivers/backup/controller/pxb/user_manager/user/organization_manager/organization/cloud_credential_spec"
	. "github.com/portworx/torpedo/drivers/backup/controller/pxb/user_manager/user/organization_manager/organization/cloud_credential_spec/aws_credential_spec"
)

type (
	AWSCredential   = CloudCredential[*AWSCredentialSpec]
	InspectResponse = CloudCredentialInspectResponse
)

type CloudCredential[S CloudCredentialSpec] struct {
	Spec            S
	InspectResponse *InspectResponse
}

func (c *CloudCredential[S]) GetSpec() S {
	return c.Spec
}

func (c *CloudCredential[S]) SetSpec(spec S) *CloudCredential[S] {
	c.Spec = spec
	return c
}

func (c *CloudCredential[S]) GetInspectResponse() *InspectResponse {
	return c.InspectResponse
}

func (c *CloudCredential[S]) SetInspectResponse(response *InspectResponse) *CloudCredential[S] {
	c.InspectResponse = response
	return c
}
