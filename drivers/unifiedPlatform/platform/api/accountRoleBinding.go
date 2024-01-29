// Package api comprises of all the components and associated CRUD functionality
package api

import (
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v2alpha1"
)

// AccountRoleBindingV2 struct
type AccountRoleBindingV2 struct {
	apiClientv2 *pdsv2.APIClient
}
