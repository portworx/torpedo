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
	// GlobalPxBackupAdminSecretName is the Kubernetes secret that stores the token for Px-Backup admin
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
	oidcSecretName := GetOIDCSecretName()
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
			return "", ProcessError(err, oidcSecretName)
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
		if strings.HasPrefix(route, "/") {
			baseURL += route
		} else {
			baseURL += fmt.Sprintf("/%s", route)
		}
	}
	return baseURL, nil
}

func (k *Keycloak) GetCommonHeaderMap(token string) map[string]string {
	headerMap := make(map[string]string)
	headerMap["Content-Type"] = "application/json"
	headerMap["Authorization"] = fmt.Sprint("Bearer ", token)
	return headerMap
}

func (k *Keycloak) MakeRequest(ctx context.Context, method string, admin bool, route string, body interface{}, headerMap map[string]string) (*http.Request, error) {
	reqURL, err := k.BuildURL(admin, route)
	if err != nil {
		return nil, ProcessError(err)
	}
	reqBody, err := func() (*bytes.Reader, error) {
		switch v := body.(type) {
		case nil:
			return nil, nil
		case string:
			return bytes.NewReader([]byte(v)), nil
		case []byte:
			return bytes.NewReader(v), nil
		default:
			bodyBytes, err := json.Marshal(v)
			if err != nil {
				return nil, ProcessError(err)
			}
			return bytes.NewReader(bodyBytes), nil
		}
	}()
	if err != nil {
		return nil, ProcessError(err)
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, reqBody)
	if err != nil {
		return nil, ProcessError(err, reqURL)
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
	httpRequest, err := k.MakeRequest(ctx, method, admin, route, body, headerMap)
	if err != nil {
		return nil, ProcessError(err)
	}
	httpResponse, err := k.GetResponse(httpRequest)
	if err != nil {
		return nil, ProcessError(err, ToString(httpRequest))
	}
	defer func() {
		err := httpResponse.Body.Close()
		if err != nil {
			log.Errorf("failed to close response body. Err: [%v]", ProcessError(err))
		}
	}()
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, ProcessError(err, ToString(httpResponse.Body))
	}
	statusCode := httpResponse.StatusCode
	switch {
	case statusCode >= 200 && statusCode < 300:
		return responseBody, nil
	default:
		requestURL, statusText := httpResponse.Request.URL, http.StatusText(statusCode)
		err = fmt.Errorf("%s %s returned status %d: %s", method, requestURL, statusCode, statusText)
		return nil, ProcessError(err, ToString(httpResponse))
	}
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
		return "", ProcessError(err, ToString(body))
	}
	return token.AccessToken, nil
}

func GetOIDCSecretName() string {
	oidcSecretName, ok := os.LookupEnv(PxBackupOIDCSecretName)
	if !ok || oidcSecretName == "" {
		oidcSecretName = DefaultOIDCSecretName
	}
	return oidcSecretName
}

func GetPxCentralAdminPassword(pxbNamespace string) (string, error) {
	pxCentralAdminSecret, err := core.Instance().GetSecret(GlobalPxCentralAdminSecretName, pxbNamespace)
	if err != nil {
		return "", ProcessError(err)
	}
	pxCentralAdminPwd := string(pxCentralAdminSecret.Data["credential"])
	if pxCentralAdminPwd == "" {
		err = fmt.Errorf("%s secret is empty", GlobalPxCentralAdminSecretName)
		return "", ProcessError(err)
	}
	return pxCentralAdminPwd, nil
}

func (k *Keycloak) LoginAsAdmin() error {

}

func (k *Keycloak) UpdatePxBackupAdminSecret(ctx context.Context, pxbNamespace string) error {
	pxCentralAdminToken, err := k.GetToken(ctx)
	if err != nil {
		return ProcessError(err)
	}
	secret, err := core.Instance().GetSecret(GlobalPxBackupAdminSecretName, pxbNamespace)
	if err != nil {
		return ProcessError(err)
	}
	secret.Data[GlobalPxBackupOrgToken] = []byte(pxCentralAdminToken)
	_, err = core.Instance().UpdateSecret(secret)
	if err != nil {
		return ProcessError(err)
	}
	return nil
}

func GetCtxWithToken(ctx context.Context, token string) context.Context {
	return metadata.NewOutgoingContext(
		ctx,
		metadata.New(
			map[string]string{
				GlobalPxBackupAuthHeader: GlobalPxBackupAuthTokenType + " " + token,
			},
		),
	)
}

func GetAdminCtxFromSecret(ctx context.Context) (context.Context, error) {
	pxbNamespace, err := GetPxBackupNamespace()
	if err != nil {
		return nil, ProcessError(err)
	}
	err = UpdatePxBackupAdminSecret(ctx)
	if err != nil {
		return nil, ProcessError(err)
	}
	secret, err := core.Instance().GetSecret(GlobalPxBackupAdminSecretName, pxbNamespace)
	if err != nil {
		return nil, ProcessError(err)
	}
	token := string(secret.Data[GlobalPxBackupOrgToken])
	if token == "" {
		err = fmt.Errorf("[%s] token in secret [%s] is empty", GlobalPxBackupAdminSecretName, GlobalPxBackupOrgToken)
		return nil, ProcessError(err)
	}
	return GetCtxWithToken(ctx, token), nil
}

type AddUserRequest struct {
	UserRepresentation *UserRepresentation `json:"userRepresentation"`
}

type AddUserResponse struct{}

func AddUser(ctx context.Context, req *AddUserRequest) (*AddUserResponse, error) {
	path := "users"
	userBytes, err := json.Marshal(req.UserRepresentation)
	if err != nil {
		return nil, ProcessError(err)
	}
	_, err = ProcessKeycloakRequest(ctx, POST, path, strings.NewReader(string(userBytes)))
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
	path := "users"
	respBody, err := ProcessKeycloakRequest(ctx, GET, path, nil)
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
	userID, err := GetUserID(ctx, req.Username)
	if err != nil {
		return nil, ProcessError(err)
	}
	path := fmt.Sprintf("users/%s", userID)
	_, err = ProcessKeycloakRequest(ctx, DELETE, path, nil)
	if err != nil {
		return nil, err
	}
	return &DeleteUserResponse{}, nil
}

func GetUserID(ctx context.Context, username string) (string, error) {
	enumerateUserReq := &EnumerateUserRequest{}
	enumerateUserResp, err := EnumerateUser(ctx, enumerateUserReq)
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

func init() {
	str, err := GetPxCentralAdminPassword()
	if err != nil {
		err = fmt.Errorf("error fetching user [%s] password from secret: [%v]", GlobalPxCentralAdminUsername, err)
		log.Errorf(ProcessError(err).Error())
	} else {
		GlobalPxCentralAdminPassword = str
	}
}
