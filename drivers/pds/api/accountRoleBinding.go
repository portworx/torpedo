// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// AccountRoleBinding struct
type AccountRoleBinding struct {
	context   context.Context
	apiClient *pds.APIClient
}

// ListAccountsRoleBindings func
func (accountRoleBinding *AccountRoleBinding) ListAccountsRoleBindings(accountID string) ([]pds.ModelsAccountRoleBinding, error) {
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	log.Info("List Account Role Bindings.")
	accountRoleBindings, res, err := client.ApiAccountsIdRoleBindingsGet(accountRoleBinding.context, accountID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsIdRoleBindingsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return accountRoleBindings.GetData(), nil
}

// ListAccountRoleBindingsOfUser func
func (accountRoleBinding *AccountRoleBinding) ListAccountRoleBindingsOfUser(userID string) ([]pds.ModelsAccountRoleBinding, error) {
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	accRoleModels, res, err := client.ApiUsersIdAccountRoleBindingsGet(accountRoleBinding.context, userID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiUsersIdAccountRoleBindingsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return accRoleModels.GetData(), nil
}

// UpdateAccountRoleBinding func
func (accountRoleBinding *AccountRoleBinding) UpdateAccountRoleBinding(accountID string, actorID string, actorType string, roleName string) (*pds.ModelsAccountRoleBinding, error) {
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	updateReq := pds.ControllersUpsertAccountRoleBindingRequest{ActorId: &actorID, ActorType: &actorType, RoleName: &roleName}
	log.Info("Get list of Accounts.")
	accRoleBinding, res, err := client.ApiAccountsIdRoleBindingsPut(accountRoleBinding.context, accountID).Body(updateReq).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsIdRoleBindingsPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return accRoleBinding, nil
}

// AddUser func
func (accountRoleBinding *AccountRoleBinding) AddUser(accountID string, email string, isAdmin bool) error {
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	rBinding := "account-reader"
	if isAdmin {
		rBinding = "account-admin"
	}
	invitationRequest := pds.ControllersInvitationRequest{Email: &email, RoleName: &rBinding}
	log.Info("Get list of Accounts.")
	res, err := client.ApiAccountsIdInvitationsPost(accountRoleBinding.context, accountID).Body(invitationRequest).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsIdInvitationsPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return err
	}
	return nil
}
