package api

import (
	"fmt"
	. "github.com/portworx/torpedo/drivers/unifiedPlatform/utils"
	"github.com/portworx/torpedo/pkg/log"
	platformV2 "github.com/pure-px/platform-api-go-client/v1alpha1"
	status "net/http"
)

// InvitationV2 struct
type InvitationV2 struct {
	ApiClientV2 *platformV2.APIClient
}

// ListAllInvitations lists all invitations
func (invite *InvitationV2) ListAllInvitations() ([]platformV2.V1Invitation, error) {
	client := invite.ApiClientV2.InvitationServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	inviteList, res, err := client.InvitationServiceListInvitations(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `InvitationServiceListInvitations`: %v\n.Full HTTP response: %v", err, res)
	}
	return inviteList.Invitations, nil
}

// AcceptInvitation accepts and received invitation and returns its response
func (invite *InvitationV2) AcceptInvitation() (*platformV2.V1AcceptInvitationResponse, error) {
	client := invite.ApiClientV2.InvitationServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	inviteModel, res, err := client.InvitationServiceAcceptInvitation(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `InvitationServiceAcceptInvitation`: %v\n.Full HTTP response: %v", err, res)
	}
	return inviteModel, nil
}

// CreateInvitation creates a new invitation and returns its model
func (invite *InvitationV2) CreateInvitation() (*platformV2.V1Invitation, error) {
	client := invite.ApiClientV2.InvitationServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	inviteModel, res, err := client.InvitationServiceCreateInvitation(ctx).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `InvitationServiceCreateInvitation`: %v\n.Full HTTP response: %v", err, res)
	}
	return inviteModel, nil
}

// GetInvitation get invitation details by its ID
func (invite *InvitationV2) GetInvitation(inviteUuId string) (*platformV2.V1Invitation, error) {
	client := invite.ApiClientV2.InvitationServiceAPI
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	inviteModel, res, err := client.InvitationServiceGetInvitation(ctx, inviteUuId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return nil, fmt.Errorf("Error when calling `InvitationServiceGetInvitation`: %v\n.Full HTTP response: %v", err, res)
	}
	return inviteModel, nil
}

// DeleteInvite deletes an invite based on its ID
func (invite *InvitationV2) DeleteInvite(inviteUuId string) (*status.Response, error) {
	client := invite.ApiClientV2.InvitationServiceAPI
	log.Info("Delete tenant.")
	ctx, err := GetContext()
	if err != nil {
		return nil, fmt.Errorf("Error in getting context for api call: %v\n", err)
	}
	_, res, _ := client.InvitationServiceDeleteInvitation(ctx, inviteUuId).Execute()
	if err != nil && res.StatusCode != status.StatusOK {
		return res, fmt.Errorf("Error when calling `InvitationServiceDeleteInvitation`: %v\n.Full HTTP response: %v", err, res)
	}
	return res, nil
}
