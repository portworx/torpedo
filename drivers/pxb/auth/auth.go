package auth

import (
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
	// GlobalPxBackupAuthTokenType defines the type of authentication token used by Px-Backup
	GlobalPxBackupAuthTokenType = "bearer"
	// GlobalPxBackupAuthHeader is the HTTP header key used for authentication in Px-Backup requests
	GlobalPxBackupAuthHeader = "authorization"
	// GlobalPxBackupOrgToken is the key for the organization-specific token within a
	// Kubernetes secret named by GlobalPxBackupAdminTokenSecretName for Px-Backup
	GlobalPxBackupOrgToken = "PX_BACKUP_ORG_TOKEN"
	// GlobalPxBackupKeycloakServiceName is the Kubernetes service that facilitates
	// user authentication through Keycloak in Px-Backup
	GlobalPxBackupKeycloakServiceName = "pxcentral-keycloak-http"
	// GlobalPxCentralAdminSecretName is the Kubernetes secret that stores px-central-admin credentials
	GlobalPxCentralAdminSecretName = "px-central-admin"
	// GlobalPxBackupAdminTokenSecretName is the Kubernetes secret that stores the token for Px-Backup admin
	GlobalPxBackupAdminTokenSecretName = "px-backup-admin-secret"
)

var (
	// GlobalPxCentralAdminUsername is the username for px-central-admin user
	GlobalPxCentralAdminUsername = "px-central-admin"
	// GlobalPxCentralAdminPassword is the password for px-central-admin user
	GlobalPxCentralAdminPassword string
)

const (
	// PxCentralUIURL is the env var for the Px-Central UI URL. Example: http://<IP>:<Port>
	PxCentralUIURL = "PX_CENTRAL_UI_URL"
	// PxBackupOIDCEndpoint is the env var for the OIDC endpoint
	PxBackupOIDCEndpoint = "OIDC_ENDPOINT"
	// PxBackupOIDCSecretName is the env var for the OIDC secret name within
	// Px-Backup namespace, defaulting to DefaultOIDCSecretName
	PxBackupOIDCSecretName = "SECRET_NAME"
)

// DefaultOIDCSecretName is the fallback Kubernetes secret in case PxBackupOIDCSecretName is not set
const DefaultOIDCSecretName = "pxc-backup-secret"

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
	Namespace     string
	AdminUsername string
	AdminPassword string
}

func (k *Keycloak) GetEndpoint(admin bool, route string) (string, error) {
	baseURL := ""
	oidcSecretName := GetOIDCSecretName()
	pxCentralUIURL := os.Getenv(PxCentralUIURL)
	// The condition checks whether pxCentralUIURL is set. This condition is added to
	// handle scenarios where Torpedo is not running as a pod in the cluster. In such
	// cases, gRPC calls to pxcentral-keycloak-http:80 would fail when run on a VM or
	// local machine using the Ginkgo CLI.
	if pxCentralUIURL != " " && len(pxCentralUIURL) > 0 {
		if admin {
			// Example: http://pxcentral-keycloak-http:80/auth/admin/realms/master
			baseURL = fmt.Sprintf("%s/auth/admin/realms/master", pxCentralUIURL)
		} else {
			// Example: http://pxcentral-keycloak-http:80/auth/realms/master
			baseURL = fmt.Sprintf("%s/auth/realms/master", pxCentralUIURL)
		}
	}
	oidcSecret, err := core.Instance().GetSecret(oidcSecretName, k.Namespace)
	if err != nil {
		return "", ProcessError(err, oidcSecretName)
	}
	oidcEndpoint := string(oidcSecret.Data[PxBackupOIDCEndpoint])
	// Construct the fully qualified domain name (FQDN) for the Keycloak service to
	// ensure DNS resolution within Kubernetes, especially for requests originating
	// from different namespace
	keycloakServiceName := GlobalPxBackupKeycloakServiceName
	replacement := fmt.Sprintf("%s.%s.svc.cluster.local", keycloakServiceName, k.Namespace)
	newURL := strings.Replace(oidcEndpoint, keycloakServiceName, replacement, 1)
	if admin {
		split := strings.Split(newURL, "auth")
		// Example: http://pxcentral-keycloak-http.px-backup.svc.cluster.local/auth/admin/realms/master
		baseURL = fmt.Sprintf("%sauth/admin%s", split[0], split[1])
	} else {
		// Example: http://pxcentral-keycloak-http.px-backup.svc.cluster.local/auth/realms/master
		baseURL = newURL
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

func (k *Keycloak) GetCommonHeaders(token string) http.Header {
	headers := make(http.Header)
	headers.Add("Content-Type", "application/json")
	headers.Add("Authorization", fmt.Sprintf("Bearer %v", token))
	return headers
}

func (k *Keycloak) MakeRequest(ctx context.Context, method string, admin bool, route string, body io.Reader, header http.Header) (*http.Request, error) {
	keycloakEndpoint, err := k.GetEndpoint(admin, route)
	if err != nil {
		return nil, ProcessError(err)
	}
	reqURL := keycloakEndpoint
	if route != "" {
		if strings.HasPrefix(route, "/") {
			reqURL += route
		} else {
			reqURL += fmt.Sprintf("/%s", route)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, ProcessError(err)
	}
	req.Header = header
	return req, nil
}

func (k *Keycloak) GetResponse(req *http.Request) (*http.Response, error) {
	resp, err := k.Client.Do(req)
	if err != nil {
		return nil, ProcessError(err)
	}
	return resp, nil
}

func (k *Keycloak) GetToken(ctx context.Context) (string, error) {
	values := make(url.Values)
	values.Set("client_id", "pxcentral")
	values.Set("username", k.AdminUsername)
	values.Set("password", k.AdminPassword)
	values.Set("grant_type", "password")
	values.Set("token-duration", "365d")
	keycloakEndpoint, err := k.GetEndpoint(false, "/protocol/openid-connect/token")
	if err != nil {
		return "", ProcessError(err)
	}
	// This token endpoint is used to retrieve tokens, as detailed in: https://www.keycloak.org/docs/latest/securing_apps/#token-endpoint
	reqURL := fmt.Sprintf("%s", keycloakEndpoint)

	header := make(http.Header)
	header.Add("Content-Type", "application/x-www-form-urlencoded")
	httpRequest, err := k.MakeRequest(ctx, "POST", reqURL, strings.NewReader(values.Encode()), header)
	if err != nil {
		return "", ProcessError(err)
	}
	httpResponse, err := k.GetResponse(httpRequest)
	if err != nil {
		return "", ProcessError(err)
	}
	body, err := InspectResponse(httpResponse)
	if err != nil {
		return "", ProcessError(err)
	}
	token := &TokenRepresentation{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		return "", ProcessError(err)
	}
	return token.AccessToken, nil
}

// GetOIDCSecretName retrieves the name of the OIDC secret from the environment or returns the default name
func GetOIDCSecretName() string {
	oidcSecretName := os.Getenv(PxBackupOIDCSecretName)
	if oidcSecretName == "" {
		oidcSecretName = DefaultOIDCSecretName
	}
	return oidcSecretName
}

func GetPxCentralAdminPassword() (string, error) {
	pxbNamespace, err := GetPxBackupNamespace()
	if err != nil {
		return "", ProcessError(err)
	}
	secret, err := core.Instance().GetSecret(GlobalPxCentralAdminSecretName, pxbNamespace)
	if err != nil {
		return "", ProcessError(err)
	}
	PxCentralAdminPwd := string(secret.Data["credential"])
	if PxCentralAdminPwd == "" {
		err = fmt.Errorf("%s secret is empty", GlobalPxCentralAdminSecretName)
		return "", ProcessError(err)
	}
	return PxCentralAdminPwd, nil
}

func GetPxCentralAdminToken(ctx context.Context) (string, error) {
	token, err := GetToken(ctx, GlobalPxCentralAdminUsername, GlobalPxCentralAdminPassword)
	if err != nil {
		return "", err
	}
	return token, nil
}

func UpdatePxBackupAdminSecret(ctx context.Context) error {
	pxbNamespace, err := GetPxBackupNamespace()
	if err != nil {
		return ProcessError(err)
	}
	pxCentralAdminToken, err := GetPxCentralAdminToken(ctx)
	if err != nil {
		return ProcessError(err)
	}
	secret, err := core.Instance().GetSecret(GlobalPxBackupAdminTokenSecretName, pxbNamespace)
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
	secret, err := core.Instance().GetSecret(GlobalPxBackupAdminTokenSecretName, pxbNamespace)
	if err != nil {
		return nil, ProcessError(err)
	}
	token := string(secret.Data[GlobalPxBackupOrgToken])
	if token == "" {
		err = fmt.Errorf("[%s] token in secret [%s] is empty", GlobalPxBackupAdminTokenSecretName, GlobalPxBackupOrgToken)
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
