package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	. "github.com/portworx/torpedo/drivers/pxb/pxbutils"
	"google.golang.org/grpc/metadata"
	"net/http"
	"time"
)

// HTTPClient is an HTTP client with a predefined timeout
var HTTPClient = &http.Client{
	Timeout: 1 * time.Minute,
}

// CredentialRepresentation defines the scheme for representing the user credential in Keycloak
type CredentialRepresentation struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}

// UserRepresentation defines the scheme for representing the user in Keycloak
type UserRepresentation struct {
	ID            string                     `json:"id"`
	Username      string                     `json:"username"`
	FirstName     string                     `json:"firstName"`
	LastName      string                     `json:"lastName"`
	Email         string                     `json:"email"`
	EmailVerified bool                       `json:"emailVerified"`
	Enabled       bool                       `json:"enabled"`
	Credentials   []CredentialRepresentation `json:"credentials"`
}

// NewTestUserRepresentation initializes the UserRepresentation for a test user with the given credentials
func NewTestUserRepresentation(username string, password string) *UserRepresentation {
	return &UserRepresentation{
		ID:            "",
		Username:      username,
		Email:         username + "@cnbu.com",
		EmailVerified: true,
		Enabled:       true,
		Credentials: []CredentialRepresentation{
			{Type: "password", Value: password, Temporary: false},
		},
	}
}

// TokenRepresentation defines the scheme for representing the Keycloak access token
type TokenRepresentation struct {
	AccessToken string `json:"access_token"`
}

func GetCommonHeaderMap(ctx context.Context) (map[string]string, error) {
	pxCentralAdminToken, err := GetPxCentralAdminToken(ctx)
	if err != nil {
		return nil, ProcessError(err)
	}
	headerMap := make(map[string]string)
	headerMap["Content-Type"] = "application/json"
	headerMap["Authorization"] = "Bearer " + pxCentralAdminToken
	return headerMap, nil
}

func ProcessWithCommonHeaderMap(ctx context.Context, method string, route string, body interface{}) ([]byte, error) {
	headerMap, err := GetCommonHeaderMap(ctx)
	if err != nil {
		return nil, ProcessError(err)
	}
	return Process(ctx, method, true, route, namespace, body, headerMap)
}

func GetPxCentralAdminToken(ctx context.Context) (string, error) {
	pxCentralAdminPassword, err := k.GetPxCentralAdminPassword()
	if err != nil {
		return "", ProcessError(err)
	}
	pxCentralAdminToken, err := k.GetToken(ctx, PxCentralAdminUsername, pxCentralAdminPassword)
	if err != nil {
		return "", ProcessError(err)
	}
	return pxCentralAdminToken, nil
}

func GetPxCentralAdminPassword() (string, error) {
	pxCentralAdminSecret, err := core.Instance().GetSecret(PxCentralAdminSecretName, k.Namespace)
	if err != nil {
		return "", ProcessError(err)
	}
	pxCentralAdminPassword := string(pxCentralAdminSecret.Data["credential"])
	if pxCentralAdminPassword == "" {
		err = fmt.Errorf("invalid secret [%s]", PxCentralAdminSecretName)
		return "", ProcessError(err)
	}
	return pxCentralAdminPassword, nil
}

func UpdatePxBackupAdminSecret(token string) error {
	pxBackupAdminSecret, err := core.Instance().GetSecret(PxBackupAdminSecretName, k.Namespace)
	if err != nil {
		return ProcessError(err)
	}
	pxBackupAdminSecret.Data[PxBackupOrgToken] = []byte(token)
	_, err = core.Instance().UpdateSecret(pxBackupAdminSecret)
	if err != nil {
		return ProcessError(err)
	}
	return nil
}

func GetAdminCtxFromSecret(ctx context.Context, update bool) (context.Context, error) {
	pxCentralAdminToken, err := k.GetPxCentralAdminToken(ctx)
	if err != nil {
		return nil, ProcessError(err)
	}
	if update {
		err = k.UpdatePxBackupAdminSecret(pxCentralAdminToken)
		if err != nil {
			return nil, ProcessError(err)
		}
	}
	adminCtx := k.GetCtxWithToken(ctx, pxCentralAdminToken)
	return adminCtx, nil
}

func GetCtxWithToken(ctx context.Context, token string) context.Context {
	authMetadata := metadata.New(
		map[string]string{
			PxBackupAuthHeader: fmt.Sprintf("%s %s", PxBackupAuthTokenType, token),
		},
	)
	return metadata.NewOutgoingContext(ctx, authMetadata)
}

type AddUserRequest struct {
	UserRepresentation *UserRepresentation `json:"userRepresentation"`
}

type AddUserResponse struct{}

func AddUser(ctx context.Context, req *AddUserRequest) (*AddUserResponse, error) {
	route := "users"
	_, err := k.ExecuteWithAdminToken(ctx, "POST", route, req.UserRepresentation)
	if err != nil {
		return nil, err
	}
	return &AddUserResponse{}, nil
}

type EnumerateUserRequest struct{}

type EnumerateUserResponse struct {
	Users []*UserRepresentation
}

func EnumerateUser(ctx context.Context, _ *EnumerateUserRequest) (*EnumerateUserResponse, error) {
	route := "users"
	headerMap, err := k.GetCommonHeaderMap(ctx)
	if err != nil {
		return nil, ProcessError(err)
	}
	respBody, err := k.Execute(ctx, "GET", true, route, nil, headerMap)
	if err != nil {
		return nil, err
	}
	enumerateResp := &EnumerateUserResponse{}
	err = json.Unmarshal(respBody, &enumerateResp.Users)
	if err != nil {
		return nil, ProcessError(err)
	}
	return enumerateResp, nil
}

type DeleteUserRequest struct {
	Username string `json:"username"`
}

type DeleteUserResponse struct{}

func DeleteUser(ctx context.Context, req *DeleteUserRequest) (*DeleteUserResponse, error) {
	userID, err := k.GetUserID(ctx, req.Username)
	if err != nil {
		return nil, ProcessError(err)
	}
	route := fmt.Sprintf("users/%s", userID)
	_, err = k.ExecuteWithAdminToken(ctx, "DELETE", route, nil)
	if err != nil {
		return nil, err
	}
	return &DeleteUserResponse{}, nil
}

func GetUserID(ctx context.Context, username string) (string, error) {
	enumerateUserReq := &EnumerateUserRequest{}
	enumerateUserResp, err := k.EnumerateUser(ctx, enumerateUserReq)
	if err != nil {
		return "", ProcessError(err)
	}
	var userID string
	for _, user := range enumerateUserResp.Users {
		if user.Username == username {
			userID = user.ID
			break
		}
	}
	if userID == "" {
		err = fmt.Errorf("no user found with the username: [%s]", username)
		return "", ProcessError(err)
	}
	return userID, nil
}
