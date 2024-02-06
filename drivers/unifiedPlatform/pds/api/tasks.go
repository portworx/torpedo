package api

import (
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
)

type TasksV2 struct {
	ApiClientV2 *platformV2.APIClient
	AccountID   string
}
