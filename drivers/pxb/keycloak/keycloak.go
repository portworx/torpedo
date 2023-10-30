package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	. "github.com/portworx/torpedo/drivers/pxb/pxbutils"
	"github.com/portworx/torpedo/pkg/log"
	"google.golang.org/grpc/metadata"
	"io"
	"net/http"
	"net/url"
)

type CredentialType string

const (
	Password CredentialType = "password"
)

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

// NewTestUserRepresentation initializes UserRepresentation for a test user with the given credentials
func NewTestUserRepresentation(username string, password string) *UserRepresentation {
	return &UserRepresentation{
		Username:      username,
		Email:         username + "@cnbu.com",
		EmailVerified: true,
		Enabled:       true,
		Credentials: []CredentialRepresentation{
			{
				Type:  string(Password),
				Value: password,
			},
		},
	}
}

// TokenRepresentation defines the scheme for representing the Keycloak access token
type TokenRepresentation struct {
	AccessToken string `json:"access_token"`
}

type Keycloak struct {
	URL struct {
		Admin    string
		NonAdmin string
	}
	Login struct {
		Username string
		Password string
	}
}

func (k *Keycloak) Invoke(ctx context.Context, method string, url string, body interface{}, headerMap map[string]string) ([]byte, error) {
	reqBody, err := ToByteArray(body)
	if err != nil {
		return nil, ProcessError(err)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, ProcessError(err)
	}
	for key, val := range headerMap {
		req.Header.Set(key, val)
	}
	resp, err := HTTPClient.Do(req)
	if err != nil {
		return nil, ProcessError(err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Errorf("failed to close response body. Err: [%v]", ProcessError(err))
		}
	}()
	statusCode := resp.StatusCode
	statusText := http.StatusText(statusCode)
	switch {
	case statusCode >= 200 && statusCode < 300:
		return io.ReadAll(resp.Body)
	default:
		err = fmt.Errorf("[%s] [%s] returned status [%d]: [%s]", method, url, statusCode, statusText)
		return nil, ProcessError(err)
	}
}

func (k *Keycloak) GetAccessToken(ctx context.Context, username string, password string) (string, error) {
	values := make(url.Values)
	values.Set("username", username)
	values.Set("password", password)
	values.Set("grant_type", "password")
	values.Set("client_id", "pxcentral")
	values.Set("token-duration", "365d")
	headerMap := make(map[string]string)
	headerMap["Content-Type"] = "application/x-www-form-urlencoded"
	reqURL := k.URL.NonAdmin + "/protocol/openid-connect/token"
	respBody, err := k.Invoke(ctx, "POST", reqURL, values.Encode(), headerMap)
	if err != nil {
		return "", ProcessError(err)
	}
	tokenRep := &TokenRepresentation{}
	err = json.Unmarshal(respBody, &tokenRep)
	if err != nil {
		return "", ProcessError(err)
	}
	return tokenRep.AccessToken, nil
}

func GetCommonHeaderMap(ctx context.Context) (map[string]string, error) {
	accessToken, err := GetPxCentralAdminToken(ctx)
	if err != nil {
		return nil, ProcessError(err)
	}
	headerMap := make(map[string]string)
	headerMap["Content-Type"] = "application/json"
	headerMap["Authorization"] = "Bearer " + accessToken
	return headerMap, nil
}

func ProcessWithCommonHeaderMap(ctx context.Context, method string, route string, body interface{}) ([]byte, error) {
	headerMap, err := GetCommonHeaderMap(ctx)
	if err != nil {
		return nil, ProcessError(err)
	}
	return Process(ctx, method, true, route, namespace, body, headerMap)
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
	Username           string
	Password           string
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
