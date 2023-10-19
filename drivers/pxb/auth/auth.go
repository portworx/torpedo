package auth

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
	"os"
	"strings"
)

const (
	// GlobalPxCentralAdminUsername is the username for px-central-admin user
	GlobalPxCentralAdminUsername = "px-central-admin"
	// GlobalPxBackupOrgToken is the organization token key within GlobalPxBackupAdminSecretName
	GlobalPxBackupOrgToken = "PX_BACKUP_ORG_TOKEN"
	// GlobalPxBackupAuthHeader is the HTTP header used for authentication in Px-Backup requests
	GlobalPxBackupAuthHeader = "authorization"
	// GlobalPxBackupAuthTokenType is the type of authentication token in Px-Backup requests
	GlobalPxBackupAuthTokenType = "bearer"
	// GlobalKeycloakServiceName is the Kubernetes service that facilitates
	// user authentication through Keycloak in Px-Backup
	GlobalKeycloakServiceName = "pxcentral-keycloak-http"
	// GlobalPxCentralAdminSecretName is the Kubernetes secret that stores px-central-admin credentials
	GlobalPxCentralAdminSecretName = "px-central-admin"
	// GlobalPxBackupAdminSecretName is the Kubernetes secret that stores Px-Backup admin token
	GlobalPxBackupAdminSecretName = "px-backup-admin-secret"
)

// DefaultOIDCSecretName is the fallback Kubernetes secret in case PxBackupOIDCSecretName is not set
const DefaultOIDCSecretName = "pxc-backup-secret"

const (
	// PxBackupOIDCEndpoint is the env var for the OIDC endpoint
	PxBackupOIDCEndpoint = "OIDC_ENDPOINT"
	// PxBackupOIDCSecretName is the env var for the OIDC secret within px-backup namespace
	PxBackupOIDCSecretName = "SECRET_NAME"
	// PxCentralUIURL is the env var for the px-central UI URL. Example: http://pxcentral-keycloak-http:80
	PxCentralUIURL = "PX_CENTRAL_UI_URL"
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

// NewTestUserRepresentation initializes the UserRepresentation for a test user with the given credentials
func NewTestUserRepresentation(username string, password string) *UserRepresentation {
	return &UserRepresentation{
		ID:            "",
		Username:      username,
		FirstName:     "first-" + username,
		LastName:      username + "-last",
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

type Keycloak struct {
	*http.Client
	Namespace string
}

func (k *Keycloak) BuildURL(admin bool, route string) (string, error) {
	baseURL := ""
	oidcSecretName := k.GetOIDCSecretName()
	pxCentralUIURL := os.Getenv(PxCentralUIURL)
	// The condition checks whether pxCentralUIURL is set. This condition is added to
	// handle scenarios where Torpedo is not running as a pod in the cluster. In such
	// cases, gRPC calls pxcentral-keycloak-http:80 would fail when made from a VM or
	// local machine using the Ginkgo CLI.
	if pxCentralUIURL != " " && len(pxCentralUIURL) > 0 {
		if admin {
			baseURL = fmt.Sprint(pxCentralUIURL, "/auth/admin/realms/master")
		} else {
			baseURL = fmt.Sprint(pxCentralUIURL, "/auth/realms/master")
		}
	} else {
		oidcSecret, err := core.Instance().GetSecret(oidcSecretName, k.Namespace)
		if err != nil {
			debugMap := DebugMap{}
			debugMap.Add("ODICSecretName", oidcSecretName)
			return "", ProcessError(err, debugMap.String())
		}
		oidcEndpoint := string(oidcSecret.Data[PxBackupOIDCEndpoint])
		// Construct the fully qualified domain name (FQDN) for the Keycloak service to
		// ensure DNS resolution within Kubernetes, especially for requests originating
		// from different namespace
		replacement := fmt.Sprintf("%s.%s.svc.cluster.local", GlobalKeycloakServiceName, k.Namespace)
		newURL := strings.Replace(oidcEndpoint, GlobalKeycloakServiceName, replacement, 1)
		if admin {
			split := strings.Split(newURL, "auth")
			baseURL = fmt.Sprint(split[0], "auth/admin", split[1])
		} else {
			baseURL = newURL
		}
	}
	if route != "" {
		if !strings.HasPrefix(route, "/") {
			baseURL += "/"
		}
		baseURL += route
	}
	return baseURL, nil
}

func (k *Keycloak) GetCommonHeaderMap(ctx context.Context) (map[string]string, error) {
	pxCentralAdminToken, err := k.GetPxCentralAdminToken(ctx)
	if err != nil {
		return nil, ProcessError(err)
	}
	headerMap := make(map[string]string)
	headerMap["Content-Type"] = "application/json"
	headerMap["Authorization"] = "Bearer " + pxCentralAdminToken
	return headerMap, nil
}

func (k *Keycloak) MakeRequest(ctx context.Context, method string, admin bool, route string, body interface{}, headerMap map[string]string) (*http.Request, error) {
	reqURL, err := k.BuildURL(admin, route)
	if err != nil {
		return nil, ProcessError(err)
	}
	reqBody, err := func() ([]byte, error) {
		switch c := body.(type) {
		case nil:
			return nil, nil
		case []byte:
			return c, nil
		case string:
			return []byte(c), nil
		default:
			bodyBytes, err := json.Marshal(c)
			if err != nil {
				debugMap := DebugMap{}
				debugMap.Add("content", c)
				return nil, ProcessError(err, debugMap.String())
			}
			return bodyBytes, nil
		}
	}()
	if err != nil {
		return nil, ProcessError(err)
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bytes.NewReader(reqBody))
	if err != nil {
		debugMap := DebugMap{}
		debugMap.Add("ReqURL", reqURL)
		debugMap.Add("ReqBody", reqBody)
		return nil, ProcessError(err, debugMap.String())
	}
	for key, value := range headerMap {
		req.Header.Set(key, value)
	}
	return req, nil
}

func (k *Keycloak) GetResponse(req *http.Request) (*http.Response, error) {
	resp, err := k.Client.Do(req)
	if err != nil {
		return nil, ProcessError(err)
	}
	return resp, nil
}

func (k *Keycloak) Execute(ctx context.Context, method string, admin bool, route string, body interface{}, headerMap map[string]string) ([]byte, error) {
	req, err := k.MakeRequest(ctx, method, admin, route, body, headerMap)
	if err != nil {
		return nil, ProcessError(err)
	}
	resp, err := k.GetResponse(req)
	if err != nil {
		debugMap := DebugMap{}
		debugMap.Add("Request", req)
		return nil, ProcessError(err, debugMap.String())
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Errorf("failed to close response body. Err: [%v]", ProcessError(err))
		}
	}()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		debugMap := DebugMap{}
		debugMap.Add("ResponseBody", resp.Body)
		return nil, ProcessError(err, debugMap.String())
	}
	statusCode := resp.StatusCode
	switch {
	case statusCode >= 200 && statusCode < 300:
		return respBody, nil
	default:
		reqURL, statusText := resp.Request.URL, http.StatusText(statusCode)
		err = fmt.Errorf("[%s] [%s] returned status [%d]: [%s]", method, reqURL, statusCode, statusText)
		debugMap := DebugMap{}
		debugMap.Add("Response", resp)
		return nil, ProcessError(err, debugMap.String())
	}
}

func (k *Keycloak) ExecuteWithAdminToken(ctx context.Context, method string, route string, body interface{}) ([]byte, error) {
	headerMap, err := k.GetCommonHeaderMap(ctx)
	if err != nil {
		return nil, ProcessError(err)
	}
	return k.Execute(ctx, method, true, route, body, headerMap)
}

func (k *Keycloak) GetToken(ctx context.Context, username, password string) (string, error) {
	route := "/protocol/openid-connect/token"
	values := make(url.Values)
	values.Set("username", username)
	values.Set("password", password)
	values.Set("grant_type", "password")
	values.Set("client_id", "pxcentral")
	values.Set("token-duration", "365d")
	headerMap := make(map[string]string)
	headerMap["Content-Type"] = "application/x-www-form-urlencoded"
	body, err := k.Execute(ctx, "POST", false, route, values.Encode(), headerMap)
	if err != nil {
		return "", ProcessError(err)
	}
	token := &TokenRepresentation{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		debugMap := DebugMap{}
		debugMap.Add("Body", body)
		return "", ProcessError(err, debugMap.String())
	}
	return token.AccessToken, nil
}

func (k *Keycloak) GetPxCentralAdminToken(ctx context.Context) (string, error) {
	pxCentralAdminPassword, err := k.GetPxCentralAdminPassword()
	if err != nil {
		return "", ProcessError(err)
	}
	pxCentralAdminToken, err := k.GetToken(ctx, GlobalPxCentralAdminUsername, pxCentralAdminPassword)
	if err != nil {
		return "", ProcessError(err)
	}
	return pxCentralAdminToken, nil
}

func (k *Keycloak) GetPxCentralAdminPassword() (string, error) {
	pxCentralAdminSecret, err := core.Instance().GetSecret(GlobalPxCentralAdminSecretName, k.Namespace)
	if err != nil {
		return "", ProcessError(err)
	}
	pxCentralAdminPassword := string(pxCentralAdminSecret.Data["credential"])
	if pxCentralAdminPassword == "" {
		err = fmt.Errorf("invalid secret [%s]", GlobalPxCentralAdminSecretName)
		return "", ProcessError(err)
	}
	return pxCentralAdminPassword, nil
}

func (k *Keycloak) UpdatePxBackupAdminSecret(token string) error {
	pxBackupAdminSecret, err := core.Instance().GetSecret(GlobalPxBackupAdminSecretName, k.Namespace)
	if err != nil {
		return ProcessError(err)
	}
	pxBackupAdminSecret.Data[GlobalPxBackupOrgToken] = []byte(token)
	_, err = core.Instance().UpdateSecret(pxBackupAdminSecret)
	if err != nil {
		return ProcessError(err)
	}
	return nil
}

func (k *Keycloak) GetAdminCtxFromSecret(ctx context.Context, update bool) (context.Context, error) {
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

func (k *Keycloak) GetOIDCSecretName() string {
	oidcSecretName, ok := os.LookupEnv(PxBackupOIDCSecretName)
	if !ok || oidcSecretName == "" {
		oidcSecretName = DefaultOIDCSecretName
	}
	return oidcSecretName
}

func (k *Keycloak) GetCtxWithToken(ctx context.Context, token string) context.Context {
	authMetadata := metadata.New(
		map[string]string{
			GlobalPxBackupAuthHeader: fmt.Sprintf("%s %s", GlobalPxBackupAuthTokenType, token),
		},
	)
	return metadata.NewOutgoingContext(ctx, authMetadata)
}

type AddUserRequest struct {
	UserRepresentation *UserRepresentation `json:"userRepresentation"`
}

type AddUserResponse struct{}

func (k *Keycloak) AddUser(ctx context.Context, req *AddUserRequest) (*AddUserResponse, error) {
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

func (k *Keycloak) EnumerateUser(ctx context.Context, _ *EnumerateUserRequest) (*EnumerateUserResponse, error) {
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

func (k *Keycloak) DeleteUser(ctx context.Context, req *DeleteUserRequest) (*DeleteUserResponse, error) {
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

func (k *Keycloak) GetUserID(ctx context.Context, username string) (string, error) {
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
