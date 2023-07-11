package cloud_credential_spec

import . "github.com/portworx/torpedo/drivers/torpedo_controller/pxb_controller/user_manager/user/organization_manager/organization/cloud_credential_spec/aws_credential_spec"

type CloudCredentialSpec interface {
	*AWSCredentialSpec
}
