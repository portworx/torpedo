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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// HTTPMethod represents an HTTP request method
type HTTPMethod string

const (
	// GET represents the HTTP GET method
	GET HTTPMethod = "GET"
	// POST represents the HTTP POST method
	POST HTTPMethod = "POST"
	// DELETE represents the HTTP DELETE method
	DELETE HTTPMethod = "DELETE"
)

// String returns the string representation of the HTTPMethod
func (m HTTPMethod) String() string {
	return string(m)
}

// CredentialType represents the type of user credentials in Keycloak
type CredentialType string

const (
	// Password represents the password type of user credentials in Keycloak
	Password CredentialType = "password"
)

// String returns the string representation of the CredentialType
func (t CredentialType) String() string {
	return string(t)
}

const (
	// GlobalPxCentralAdminUsername is the username for px-central-admin user
	GlobalPxCentralAdminUsername = "px-central-admin"
	// GlobalPxCentralAdminSecretName is the name of the Kubernetes secret that stores the credentials for the px-central-admin user
	GlobalPxCentralAdminSecretName = "px-central-admin"
	// GlobalPxBackupAuthTokenType defines the type of authentication token used by Px-Backup
	GlobalPxBackupAuthTokenType = "bearer"
	// GlobalPxBackupServiceName is the name of the Kubernetes service associated with Px-Backup
	GlobalPxBackupServiceName = "px-backup"
	// GlobalPxBackupOrgToken is the key for the organization-specific token within a
	// Kubernetes secret named by GlobalPxBackupAdminTokenSecretName for Px-Backup
	GlobalPxBackupOrgToken = "PX_BACKUP_ORG_TOKEN"
	// GlobalPxBackupAdminTokenSecretName is the name of the Kubernetes secret that stores the admin token for Px-Backup
	GlobalPxBackupAdminTokenSecretName = "px-backup-admin-secret"
	// GlobalPxBackupAuthHeader is the HTTP header key used for authentication in Px-Backup requests
	GlobalPxBackupAuthHeader = "authorization"
	// GlobalPxBackupKeycloakServiceName is the name of the Kubernetes service that
	// facilitates user authentication through Keycloak in Px-Backup
	GlobalPxBackupKeycloakServiceName = "pxcentral-keycloak-http"
)

var (
	// GlobalPxCentralAdminPassword is the password for px-central-admin user
	GlobalPxCentralAdminPassword string
	// GlobalHTTPClient is an HTTP client with a predefined timeout
	GlobalHTTPClient = &http.Client{
		Timeout: 1 * time.Minute,
	}
)

const (
	// DefaultOIDCSecretName is the default name of the Kubernetes secret that stores OIDC credentials for Px-Backup
	DefaultOIDCSecretName = "pxc-backup-secret"
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

// CredentialRepresentation defines the scheme for representing user credential in Keycloak
type CredentialRepresentation struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}

// UserRepresentation defines the scheme for representing a user in Keycloak
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

// NewTestUserRepresentation initializes a UserRepresentation for a test user with the given credentials
func NewTestUserRepresentation(username string, password string) *UserRepresentation {
	return &UserRepresentation{
		ID:            "",
		Username:      username,
		FirstName:     "first-" + username,
		LastName:      username + "last",
		Email:         username + "@cnbu.com",
		EmailVerified: true,
		Enabled:       true,
		Credentials: []CredentialRepresentation{
			{
				Type:      Password.String(),
				Temporary: false,
				Value:     password,
			},
		},
	}
}

// TokenRepresentation defines the scheme for representing the Keycloak access token
type TokenRepresentation struct {
	AccessToken string `json:"access_token"`
}

// ProcessHTTPRequest sends an HTTP request with the given method, URL, body, and headers, and returns the response
func ProcessHTTPRequest(ctx context.Context, method HTTPMethod, url string, body io.Reader, headers http.Header) (*http.Response, error) {
	httpRequest, err := http.NewRequestWithContext(ctx, method.String(), url, body)
	if err != nil {
		return nil, ProcessError(err)
	}
	httpRequest.Header = headers
	httpResponse, err := GlobalHTTPClient.Do(httpRequest)
	if err != nil {
		return nil, ProcessError(err)
	}
	return httpResponse, nil
}

// ProcessHTTPResponse processes the given HTTP response and returns its body as a byte slice
func ProcessHTTPResponse(response *http.Response) ([]byte, error) {
	if response == nil {
		err := fmt.Errorf("response is nil")
		return nil, ProcessError(err)
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Errorf("failed to close response body. Err: [%v]", ProcessError(err))
		}
	}()
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, ProcessError(err)
	}
	statusCode, requestURL := response.StatusCode, response.Request.URL
	switch {
	case statusCode >= 200 && statusCode < 300:
		return respBody, nil
	case statusCode >= 400 && statusCode < 500:
		err = fmt.Errorf("client-side error for URL [%s]. Status code: [%d], Response Body: [%s]", requestURL, statusCode, respBody)
		return nil, ProcessError(err)
	case statusCode >= 500:
		err = fmt.Errorf("server-side error for URL [%s]. Status code: [%d], Response Body: [%s]", requestURL, statusCode, respBody)
		return nil, ProcessError(err)
	default:
		err = fmt.Errorf("unexpected status code %d for URL %s. Response Body: %s", statusCode, requestURL, respBody)
		return nil, ProcessError(err)
	}
}

// GetOIDCSecretName retrieves the name of the OIDC secret from the environment or returns the default name
func GetOIDCSecretName() string {
	oidcSecretName := os.Getenv(PxBackupOIDCSecretName)
	if oidcSecretName == "" {
		oidcSecretName = DefaultOIDCSecretName
	}
	return oidcSecretName
}

// GetToken fetches the authentication token for the given username and password
func GetToken(ctx context.Context, username string, password string) (string, error) {
	values := make(url.Values)
	values.Set("client_id", "pxcentral")
	values.Set("username", username)
	values.Set("password", password)
	values.Set("grant_type", "password")
	values.Set("token-duration", "365d")
	keycloakEndPoint, err := GetKeycloakEndPoint(false)
	if err != nil {
		return "", ProcessError(err)
	}
	// This token endpoint is used to retrieve tokens, as detailed in: https://www.keycloak.org/docs/latest/securing_apps/#token-endpoint
	requestURL := fmt.Sprintf("%s/protocol/openid-connect/token", keycloakEndPoint)
	headers := make(http.Header)
	headers.Add("Content-Type", "application/x-www-form-urlencoded")
	httpResponse, err := ProcessHTTPRequest(ctx, POST, requestURL, strings.NewReader(values.Encode()), headers)
	if err != nil {
		return "", ProcessError(err)
	}
	body, err := ProcessHTTPResponse(httpResponse)
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

// GetCommonHTTPHeaders generates common HTTP headers including the authentication token for the provided credentials
func GetCommonHTTPHeaders(ctx context.Context, username string, password string) (http.Header, error) {
	token, err := GetToken(ctx, username, password)
	if err != nil {
		return nil, ProcessError(err)
	}
	headers := make(http.Header)
	headers.Add("Authorization", fmt.Sprintf("Bearer %v", token))
	headers.Add("Content-Type", "application/json")
	return headers, nil
}

// GetPxBackupNamespace retrieves the namespace where Px-Backup service is running
func GetPxBackupNamespace() (string, error) {
	allServices, err := core.Instance().ListServices("", metav1.ListOptions{})
	if err != nil {
		return "", ProcessError(err)
	}
	for _, svc := range allServices.Items {
		if svc.Name == GlobalPxBackupServiceName {
			return svc.Namespace, nil
		}
	}
	err = fmt.Errorf("cannot find Px-Backup service [%s] from the list of services", GlobalPxBackupServiceName)
	return "", ProcessError(err)
}

// GetKeycloakEndPoint returns the Keycloak endpoint URL based on the provided admin flag
func GetKeycloakEndPoint(admin bool) (string, error) {
	pxCentralUIURL := os.Getenv(PxCentralUIURL)
	// This condition is added to handle scenarios where Torpedo is not running as a pod in the cluster.
	// In such cases, gRPC calls to pxcentral-keycloak-http:80 would fail when executed from a VM or local machine using the Ginkgo CLI.
	// The condition checks whether an env var is set.
	if pxCentralUIURL != " " && len(pxCentralUIURL) > 0 {
		if admin {
			// adminURL: http://pxcentral-keycloak-http:80/auth/admin/realms/master
			adminURL := fmt.Sprintf("%s/auth/admin/realms/master", pxCentralUIURL)
			return adminURL, nil
		} else {
			// nonAdminURL: http://pxcentral-keycloak-http:80/auth/realms/master
			nonAdminURL := fmt.Sprintf("%s/auth/realms/master", pxCentralUIURL)
			return nonAdminURL, nil
		}
	}
	oidcSecretName := GetOIDCSecretName()
	pxbNamespace, err := GetPxBackupNamespace()
	if err != nil {
		return "", err
	}
	oidcSecret, err := core.Instance().GetSecret(oidcSecretName, pxbNamespace)
	if err != nil {
		return "", err
	}
	oidcEndpoint := string(oidcSecret.Data[PxBackupOIDCEndpoint])
	// Expand the service name for K8S DNS resolution, for Keycloak requests from different ns
	replacement := fmt.Sprintf("%s.%s.svc.cluster.local", GlobalPxBackupKeycloakServiceName, pxbNamespace)
	newURL := strings.Replace(oidcEndpoint, GlobalPxBackupKeycloakServiceName, replacement, 1)
	if admin {
		split := strings.Split(newURL, "auth")
		// adminURL: http://pxcentral-keycloak-http.px-backup.svc.cluster.local/auth/admin/realms/master
		adminURL := fmt.Sprintf("%sauth/admin%s", split[0], split[1])
		return adminURL, nil
	} else {
		// nonAdminURL: http://pxcentral-keycloak-http.px-backup.svc.cluster.local/auth/realms/master
		nonAdminURL := newURL
		return nonAdminURL, nil
	}
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
	err = json.Unmarshal(respBody, enumerateResp)
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

func ProcessKeycloakRequest(ctx context.Context, method HTTPMethod, path string, body io.Reader) ([]byte, error) {
	headers, err := GetCommonHTTPHeaders(ctx, GlobalPxCentralAdminUsername, GlobalPxCentralAdminPassword)
	if err != nil {
		return nil, ProcessError(err)
	}
	keycloakEndPoint, err := GetKeycloakEndPoint(true)
	if err != nil {
		return nil, ProcessError(err)
	}
	requestURL := keycloakEndPoint
	if path != "" {
		if strings.HasPrefix(path, "/") {
			requestURL += path
		} else {
			requestURL += fmt.Sprintf("/%s", path)
		}
	}
	httpResponse, err := ProcessHTTPRequest(ctx, method, requestURL, body, headers)
	if err != nil {
		return nil, ProcessError(err)
	}
	respBody, err := ProcessHTTPResponse(httpResponse)
	if err != nil {
		return nil, ProcessError(err)
	}
	return respBody, nil
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
