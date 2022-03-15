// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// AccountRoleBinding struct comprise of Context , PDS api client.
type AccountRoleBinding struct {
	Context   context.Context
	apiClient *pds.APIClient
}

// ListAccountsRoleBindings function return list of Account and assciated roles.
func (accountRoleBinding *AccountRoleBinding) ListAccountsRoleBindings(accountID string) ([]pds.ModelsAccountRoleBinding, error) {
	log.Info("List Account Role Bindings.")
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	accountRoleBindings, res, err := client.ApiAccountsIdRoleBindingsGet(accountRoleBinding.Context, accountID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsIdRoleBindingsGet``: %v", err)
		log.Debugf("Full HTTP response: %v", res)
		return nil, err
	}
	return accountRoleBindings.GetData(), nil
}

// ListAccountRoleBindingsOfUser function return list of account to which the given user belongs to.
func (accountRoleBinding *AccountRoleBinding) ListAccountRoleBindingsOfUser(userID string) ([]pds.ModelsAccountRoleBinding, error) {
	log.Infof("List Account Role Bindings of the user having ID: %v.", userID)
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	accRoleModels, res, err := client.ApiUsersIdAccountRoleBindingsGet(accountRoleBinding.Context, userID).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiUsersIdAccountRoleBindingsGet``: %v\n", err)
		log.Debugf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return accRoleModels.GetData(), nil
}

// UpdateAccountRoleBinding function used to update the user roles for given user.
func (accountRoleBinding *AccountRoleBinding) UpdateAccountRoleBinding(accountID string, actorID string, actorType string, roleName string) (*pds.ModelsAccountRoleBinding, error) {
	log.Infof("Update user with ID: %v", actorID)
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	updateReq := pds.ControllersUpsertAccountRoleBindingRequest{ActorId: &actorID, ActorType: &actorType, RoleName: &roleName}
	accRoleBinding, res, err := client.ApiAccountsIdRoleBindingsPut(accountRoleBinding.Context, accountID).Body(updateReq).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsIdRoleBindingsPut``: %v", err)
		log.Debugf("Full HTTP response: %v", res)
		return nil, err
	}
	return accRoleBinding, nil
}

// AddUser function add the user which is already registered to portworx.
func (accountRoleBinding *AccountRoleBinding) AddUser(accountID string, email string, isAdmin bool) error {
	log.Info("Adding new user.")
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	rBinding := "account-reader"
	if isAdmin {
		rBinding = "account-admin"
	}
	invitationRequest := pds.ControllersInvitationRequest{Email: &email, RoleName: &rBinding}
	res, err := client.ApiAccountsIdInvitationsPost(accountRoleBinding.Context, accountID).Body(invitationRequest).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsIdInvitationsPost``: %v", err)
		log.Debugf("Full HTTP response: %v", res)
		return err
	}
	return nil
}
