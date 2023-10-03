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
)

type HTTPMethod string

const (
	GET    HTTPMethod = "GET"
	POST              = "POST"
	PUT               = "PUT"
	DELETE            = "DELETE"
)

const (
	GlobalPxCentralAdminUsername       = "px-central-admin"
	GlobalPxCentralAdminSecretName     = "px-central-admin"
	GlobalPxBackupAuthTokenType        = "bearer"
	GlobalPxBackupServiceName          = "px-backup"
	GlobalPxBackupOrgToken             = "PX_BACKUP_ORG_TOKEN"
	GlobalPxBackupAdminTokenSecretName = "px-backup-admin-secret"
	GlobalPxBackupAuthHeader           = "authorization"
	GlobalPxBackupKeycloakServiceName  = "pxcentral-keycloak-http"
)

var (
	GlobalPxCentralAdminPassword string
)

const (
	// DefaultOIDCSecretName is the default OIDC secret name
	DefaultOIDCSecretName = "pxc-backup-secret"
)

const (
	// PxCentralUIURL is the env var for the PX-Central UI URL. Example: http://<IP>:<Port>
	PxCentralUIURL = "PX_CENTRAL_UI_URL"
	// PxBackupOIDCEndpoint is the env var for the OIDC endpoint
	PxBackupOIDCEndpoint = "OIDC_ENDPOINT"
	// PxBackupOIDCSecretName is the env var for the OIDC secret name within Px-Backup namespace, defaulting to DefaultOIDCSecretName
	PxBackupOIDCSecretName = "SECRET_NAME"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type CredentialRepresentation struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}

func NewCredentialRepresentation(credType string, value string, temporary bool) *CredentialRepresentation {
	return &CredentialRepresentation{
		Type:      credType,
		Value:     value,
		Temporary: temporary,
	}
}

type UserRepresentation struct {
	ID            string                     `json:"id"`
	Name          string                     `json:"username"`
	FirstName     string                     `json:"firstName"`
	LastName      string                     `json:"lastName"`
	EmailVerified bool                       `json:"emailVerified"`
	Enabled       bool                       `json:"enabled"`
	Email         string                     `json:"email"`
	Credentials   []CredentialRepresentation `json:"credentials"`
}

func ProcessHTTPRequest(ctx context.Context, method HTTPMethod, url string, headers http.Header, body io.Reader) (responseBody []byte, err error) {
	httpRequest, err := http.NewRequestWithContext(ctx, string(method), url, body)
	if err != nil {
		return nil, ProcessError(err, StructToString(httpRequest))
	}
	httpRequest.Header = headers
	client := &http.Client{}
	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return nil, ProcessError(err, StructToString(httpRequest))
	}
	defer func() {
		err := httpResponse.Body.Close()
		if err != nil {
			log.Errorf("error closing HTTP response body: %v", ProcessError(err, StructToString(httpResponse)))
		}
	}()
	responseBody, err = io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, ProcessError(err, StructToString(httpResponse))
	}
	return responseBody, nil
}

func GetOIDCSecretName() string {
	oidcSecretName := os.Getenv(PxBackupOIDCSecretName)
	if oidcSecretName == "" {
		oidcSecretName = DefaultOIDCSecretName
	}
	return oidcSecretName
}

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
	reqURL := fmt.Sprintf("%s/protocol/openid-connect/token", keycloakEndPoint)
	headers := make(http.Header)
	headers.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := ProcessHTTPRequest(ctx, POST, reqURL, headers, strings.NewReader(values.Encode()))
	if err != nil {
		return "", ProcessError(err)
	}
	token := &TokenResponse{}
	err = json.Unmarshal(response, &token)
	if err != nil {
		return "", ProcessError(err)
	}
	return token.AccessToken, nil
}

func GetUserAuthHeaders(ctx context.Context, username string, password string) (http.Header, error) {
	token, err := GetToken(ctx, username, password)
	if err != nil {
		debugStruct := struct {
			username string
			password string
		}{
			username: username,
			password: "", // password left blank on purpose
		}
		return nil, ProcessError(err, StructToString(debugStruct))
	}
	headers := make(http.Header)
	headers.Add("Authorization", fmt.Sprintf("Bearer %v", token))
	headers.Add("Content-Type", "application/json")
	return headers, nil
}

func AddUser(ctx context.Context, username string, firstName string, lastName string, email string, enabled bool, password string, temporary bool) error {
	keycloakEndPoint, err := GetKeycloakEndPoint(true)
	if err != nil {
		return ProcessError(err)
	}
	requestURL := fmt.Sprintf("%s/users", keycloakEndPoint)
	headers, err := GetUserAuthHeaders(ctx, GlobalPxCentralAdminUsername, GlobalPxCentralAdminPassword)
	if err != nil {
		return ProcessError(err)
	}
	userRepresentation := &UserRepresentation{
		Name:      username,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Enabled:   enabled,
		Credentials: []CredentialRepresentation{
			*NewCredentialRepresentation("password", password, temporary),
		},
	}
	userBytes, err := json.Marshal(userRepresentation)
	if err != nil {
		return ProcessError(err)
	}
	response, err := ProcessHTTPRequest(ctx, POST, requestURL, headers, strings.NewReader(string(userBytes)))
	log.Infof("response %v", response)
	if err != nil {
		return err
	}
	log.Infof("User [%s] created", username)
	return nil
}

func GetUserID(ctx context.Context, username string) (string, error) {
	headers, err := GetUserAuthHeaders(ctx, GlobalPxCentralAdminUsername, GlobalPxCentralAdminPassword)
	if err != nil {
		return "", ProcessError(err)
	}
	keycloakEndPoint, err := GetKeycloakEndPoint(true)
	if err != nil {
		return "", ProcessError(err)
	}
	// TODO Need to increase the limit
	reqURL := fmt.Sprintf("%s/users", keycloakEndPoint)
	response, err := ProcessHTTPRequest(ctx, GET, reqURL, headers, nil)
	if err != nil {
		return "", ProcessError(err)
	}
	var users []UserRepresentation
	err = json.Unmarshal(response, &users)
	if err != nil {
		return "", ProcessError(err)
	}
	var clientID string
	for _, user := range users {
		if user.Name == username {
			clientID = user.ID
			break
		}
	}
	log.Infof("Fetching ID of user %s - %s", username, clientID)
	return clientID, nil
}

func DeleteUser(ctx context.Context, username string) error {
	keycloakEndPoint, err := GetKeycloakEndPoint(true)
	if err != nil {
		return err
	}
	userID, err := GetUserID(ctx, username)
	if err != nil {
		return err
	}
	reqURL := fmt.Sprintf("%s/users/%s", keycloakEndPoint, userID)
	headers, err := GetUserAuthHeaders(context.Background(), GlobalPxCentralAdminUsername, GlobalPxCentralAdminPassword)
	if err != nil {
		return err
	}

	response, err := ProcessHTTPRequest(context.Background(), DELETE, reqURL, headers, nil)
	log.Infof("Response for user [%s] deletion - %s", username, string(response))
	if err != nil {
		return err
	}
	log.Infof("Deleted User - %s", username)
	return nil
}

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

func GetKeycloakEndPoint(admin bool) (string, error) {
	pxCentralUIURL := os.Getenv(PxCentralUIURL)
	// This condition is added to handle scenarios where Torpedo is not running as a pod in the cluster.
	// In such cases, gRPC calls to pxcentral-keycloak-http:80 would fail when executed from a VM or local machine using the Ginkgo CLI.
	// The condition checks whether an env var is set.
	if pxCentralUIURL != " " && len(pxCentralUIURL) > 0 {
		if admin {
			// http://pxcentral-keycloak-http:80/auth/admin/realms/master
			adminURL := fmt.Sprintf("%s/auth/admin/realms/master", pxCentralUIURL)
			return adminURL, nil
		} else {
			// http://pxcentral-keycloak-http:80/auth/realms/master
			nonAdminURL := fmt.Sprintf("%s/auth/realms/master", pxCentralUIURL)
			return nonAdminURL, nil
		}
	}
	name := GetOIDCSecretName()
	ns, err := GetPxBackupNamespace()
	if err != nil {
		return "", err
	}
	secret, err := core.Instance().GetSecret(name, ns)
	if err != nil {
		return "", err
	}
	url := string(secret.Data[PxBackupOIDCEndpoint])
	// Expand the service name for K8S DNS resolution, for keycloak requests from different ns
	replacement := fmt.Sprintf("%s.%s.svc.cluster.local", GlobalPxBackupKeycloakServiceName, ns)
	newURL := strings.Replace(url, GlobalPxBackupKeycloakServiceName, replacement, 1)
	// url: http://pxcentral-keycloak-http.px-backup.svc.cluster.local/auth/realms/master
	if admin {
		// admin url: http://pxcentral-keycloak-http.px-backup.svc.cluster.local/auth/realms/master
		// non-admin url: http://pxcentral-keycloak-http.px-backup.svc.cluster.local/auth/admin/realms/master
		split := strings.Split(newURL, "auth")
		newURL = fmt.Sprintf("%sauth/admin%s", split[0], split[1])
		return newURL, nil
	}
	return string(newURL), nil
}

func GetPxCentralAdminPassword() (string, error) {
	pxbNamespace, err := GetPxBackupNamespace()
	if err != nil {
		return "", ProcessError(err)
	}
	secret, err := core.Instance().GetSecret(GlobalPxCentralAdminSecretName, pxbNamespace)
	if err != nil {
		debugStruct := struct {
			PxbNamespace string
		}{
			PxbNamespace: pxbNamespace,
		}
		return "", ProcessError(err, StructToString(debugStruct))
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
	pxCentralAdminToken, err := GetPxCentralAdminToken(ctx)
	if err != nil {
		return ProcessError(err)
	}
	pxbNamespace, err := GetPxBackupNamespace()
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

func GetCtxWithToken(token string) context.Context {
	return metadata.NewOutgoingContext(
		context.Background(),
		metadata.New(
			map[string]string{
				GlobalPxBackupAuthHeader: GlobalPxBackupAuthTokenType + " " + token,
			},
		),
	)
}

func GetAdminCtxFromSecret(ctx context.Context) (context.Context, error) {
	err := UpdatePxBackupAdminSecret(ctx)
	if err != nil {
		return nil, ProcessError(err)
	}
	pxbNamespace, err := GetPxBackupNamespace()
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
	return GetCtxWithToken(token), nil
}

func init() {
	str, err := GetPxCentralAdminPassword()
	if err != nil {
		err = fmt.Errorf("error fetching [%s] password from secret: [%v]", err)
		log.Errorf(ProcessError(err).Error())
	}
	GlobalPxCentralAdminPassword = str
}
