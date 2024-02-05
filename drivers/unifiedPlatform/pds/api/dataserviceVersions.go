package api

import (
	pdsv2 "github.com/portworx/pds-api-go-client/unifiedcp/v1alpha1"
)

// DataServiceVersionsV2 struct
type DataServiceVersionsV2 struct {
	ApiClientV2 *pdsv2.APIClient
	AccountID   string
}
