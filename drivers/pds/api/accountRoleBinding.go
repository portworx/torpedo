// Package api comprises of all the components and associated CRUD functionality
package api

import (
	"context"
	status "net/http"

	pds "github.com/portworx/pds-api-go-client/pds/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type AccountRoleBinding struct {
	context   context.Context
	apiClient *pds.APIClient
}

func (accountRoleBinding *AccountRoleBinding) ListAccountsRoleBindings(accountId string) ([]pds.ModelsAccountRoleBinding, error) {
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	log.Info("List Account Role Bindings.")
	accountRoleBindings, res, err := client.ApiAccountsIdRoleBindingsGet(accountRoleBinding.context, accountId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsIdRoleBindingsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return accountRoleBindings.GetData(), nil
}

func (accountRoleBinding *AccountRoleBinding) ListAccountRoleBindingsOfUser(userId string) ([]pds.ModelsAccountRoleBinding, error) {
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	accRoleModels, res, err := client.ApiUsersIdAccountRoleBindingsGet(accountRoleBinding.context, userId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiUsersIdAccountRoleBindingsGet``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return accRoleModels.GetData(), nil
}

func (accountRoleBinding *AccountRoleBinding) UpdateAccountRoleBinding(accountId string, actorId string, actorType string, roleName string) (*pds.ModelsAccountRoleBinding, error) {
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	updateReq := pds.ControllersUpsertAccountRoleBindingRequest{ActorId: &actorId, ActorType: &actorType, RoleName: &roleName}
	log.Info("Get list of Accounts.")
	accRoleBinding, res, err := client.ApiAccountsIdRoleBindingsPut(accountRoleBinding.context, accountId).Body(updateReq).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsIdRoleBindingsPut``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return nil, err
	}
	return accRoleBinding, nil
}

func (accountRoleBinding *AccountRoleBinding) AddUser(accountId string, email string, isAdmin bool) error {
	client := accountRoleBinding.apiClient.AccountRoleBindingsApi
	rBinding := "account-reader"
	if isAdmin {
		rBinding = "account-admin"
	}
	invitationRequest := pds.ControllersInvitationRequest{Email: &email, RoleName: &rBinding}
	log.Info("Get list of Accounts.")
	res, err := client.ApiAccountsIdInvitationsPost(accountRoleBinding.context, accountId).Body(invitationRequest).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		log.Errorf("Error when calling `ApiAccountsIdInvitationsPost``: %v\n", err)
		log.Errorf("Full HTTP response: %v\n", res)
		return err
	}
	return nil
}
