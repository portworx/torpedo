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
	"time"
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
	// GlobalPxBackupKeycloakServiceName is the Kubernetes service that facilitates user authentication
	// through Keycloak in Px-Backup
	GlobalPxBackupKeycloakServiceName = "pxcentral-keycloak-http"
	// GlobalPxCentralAdminSecretName is the Kubernetes secret that stores px-central-admin credentials
	GlobalPxCentralAdminSecretName = "px-central-admin"
	// GlobalPxBackupAdminSecretName is the Kubernetes secret that stores Px-Backup admin token
	GlobalPxBackupAdminSecretName = "px-backup-admin-secret"
)

// DefaultOIDCSecretName is the fallback Kubernetes secret in case PxBackupOIDCSecretName is not set
const DefaultOIDCSecretName = "pxc-backup-secret"

// GlobalHTTPClient is an HTTP client with a predefined timeout
var GlobalHTTPClient = &http.Client{
	Timeout: 1 * time.Minute,
}

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

func MakeFQDN(serviceName string, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace)
}

func BuildURL(admin bool, route string, namespace string) (string, error) {
	reqURL := ""
	oidcSecretName := GetOIDCSecretName()
	pxCentralUIURL := os.Getenv(PxCentralUIURL)
	// The condition checks whether pxCentralUIURL is set. This condition is added to
	// handle scenarios where Torpedo is not running as a pod in the cluster. In such
	// cases, gRPC calls pxcentral-keycloak-http:80 would fail when made from a VM or
	// local machine using the Ginkgo CLI.
	if pxCentralUIURL != " " && len(pxCentralUIURL) > 0 {
		if admin {
			reqURL = fmt.Sprint(pxCentralUIURL, "/auth/admin/realms/master")
		} else {
			reqURL = fmt.Sprint(pxCentralUIURL, "/auth/realms/master")
		}
	} else {
		oidcSecret, err := core.Instance().GetSecret(oidcSecretName, namespace)
		if err != nil {
			return "", ProcessError(err)
		}
		oidcEndpoint := string(oidcSecret.Data[PxBackupOIDCEndpoint])
		// Construct the fully qualified domain name (FQDN) for the Keycloak service to
		// ensure DNS resolution within Kubernetes, especially for requests originating
		// from different namespace
		replacement := MakeFQDN(GlobalPxBackupKeycloakServiceName, namespace)
		newURL := strings.Replace(oidcEndpoint, GlobalPxBackupKeycloakServiceName, replacement, 1)
		if admin {
			split := strings.Split(newURL, "auth")
			reqURL = fmt.Sprint(split[0], "auth/admin", split[1])
		} else {
			reqURL = newURL
		}
	}
	if route != "" {
		if !strings.HasPrefix(route, "/") {
			reqURL += "/"
		}
		reqURL += route
	}
	return reqURL, nil
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

func ToByteArray(body interface{}) ([]byte, error) {
	if b, ok := body.([]byte); ok {
		return b, nil
	}
	if s, ok := body.(string); ok {
		return []byte(s), nil
	}
	return json.Marshal(body)
}

func MakeRequest(ctx context.Context, method string, reqURL string, body interface{}, headerMap map[string]string) (*http.Request, error) {
	reqBody, err := ToByteArray(body)
	if err != nil {
		return nil, ProcessError(err)
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, ProcessError(err)
	}
	for key, val := range headerMap {
		req.Header.Set(key, val)
	}
	return req, nil
}

func GetResponse(req *http.Request) (*http.Response, error) {
	resp, err := GlobalHTTPClient.Do(req)
	if err != nil {
		return nil, ProcessError(err)
	}
	return resp, nil
}

func Process(ctx context.Context, method string, admin bool, route string, namespace string, body interface{}, headerMap map[string]string) ([]byte, error) {
	reqURL, err := BuildURL(admin, route, namespace)
	if err != nil {
		return nil, ProcessError(err)
	}
	req, err := MakeRequest(ctx, method, reqURL, body, headerMap)
	if err != nil {
		return nil, ProcessError(err)
	}
	resp, err := GetResponse(req)
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

func ProcessWithCommonHeaderMap(ctx context.Context, method string, route string, body interface{}) ([]byte, error) {
	headerMap, err := GetCommonHeaderMap(ctx)
	if err != nil {
		return nil, ProcessError(err)
	}
	return Process(ctx, method, true, route, namespace, body, headerMap)
}

func GetToken(ctx context.Context, username, password string) (string, error) {
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

func GetPxCentralAdminToken(ctx context.Context) (string, error) {
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

func GetPxCentralAdminPassword() (string, error) {
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

func UpdatePxBackupAdminSecret(token string) error {
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

func GetOIDCSecretName() string {
	oidcSecretName, ok := os.LookupEnv(PxBackupOIDCSecretName)
	if !ok || oidcSecretName == "" {
		oidcSecretName = DefaultOIDCSecretName
	}
	return oidcSecretName
}

func GetCtxWithToken(ctx context.Context, token string) context.Context {
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
