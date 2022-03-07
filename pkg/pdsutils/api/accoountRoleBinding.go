// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

/*
	AccountRoleBinding struct consists the context and apiClient.
	It can be used to utilize CRUD functionality with respect to Account Role Binding API.
*/
type AccountRoleBinding struct {
	Context   context.Context
	apiClient *pds.APIClient
}

/*
	Get the List of Account Role bindings
	@param accountId string - Account UUID.
	@return []pds.ModelsAccountRoleBinding, error
*/
func (accountRoleBinding *AccountRoleBinding) ListAccountsRoleBindings(accountId string) ([]pds.ModelsAccountRoleBinding, error) {
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	log.Info("List Account Role Bindings.")
	accountRoleBindings, res, err := client.ApiAccountsIdRoleBindingsGet(accountRoleBinding.Context, accountId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsGet``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
		return nil, err
	}
	return accountRoleBindings.GetData(), nil
}

/*
	Get the List of Account Role bindings for given user.
	@param userId string - User actor ID.
	@return []pds.ModelsAccountRoleBinding, error
*/
func (accountRoleBinding *AccountRoleBinding) ListAccountRoleBindingsOfUser(userId string) ([]pds.ModelsAccountRoleBinding, error) {
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	accRoleModels, res, err := client.ApiUsersIdAccountRoleBindingsGet(accountRoleBinding.Context, userId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsGet``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
		return nil, err
	}
	return accRoleModels.GetData(), nil
}

/*
	Update Account Role binding for given user.
	@param accountId string - Account UUID.
	@param actorId string -  User id.
	@param actorType string - User type. i.e user/service
	@return *pds.ModelsAccountRoleBinding, error
*/
func (accountRoleBinding *AccountRoleBinding) UpdateAccountRoleBinding(accountId string, actorId string, actorType string, roleName string) (*pds.ModelsAccountRoleBinding, error) {
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	updateReq := pds.ControllersUpsertAccountRoleBindingRequest{&actorId, &actorType, &roleName}
	log.Info("Get list of Accounts.")
	accRoleBinding, res, err := client.ApiAccountsIdRoleBindingsPut(accountRoleBinding.Context, accountId).Body(updateReq).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsGet``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
		return nil, err
	}
	return accRoleBinding, nil
}

/*
	Add new user if its already part of PX-Central.
	@param accountId string - Account UUID.
	@param email string - Email address.
	@param isAdmin bool - If the new should be added as admin-user.
	@return error
*/
func (accountRoleBinding *AccountRoleBinding) AddUser(accountId string, email string, isAdmin bool) error {
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	rBinding := "account-reader"
	if isAdmin {
		rBinding = "account-admin"
	}
	invitationRequest := pds.ControllersInvitationRequest{&email, &rBinding}
	log.Info("Get list of Accounts.")
	res, err := client.ApiAccountsIdInvitationsPost(accountRoleBinding.Context, accountId).Body(invitationRequest).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsGet``: %v\n", err)
		log.Error("Full HTTP response: %v\n", res)
		return err
	}
	return nil
}
