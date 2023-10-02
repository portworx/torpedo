package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	k8s "github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/pxb/pxbutils"
	"github.com/portworx/torpedo/pkg/log"
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
	GlobalPxCentralAdminUsername      = "px-central-admin"
	GlobalPxCentralAdminSecretName    = "px-central-admin"
	GlobalPxBackupServiceName         = "px-backup"
	GlobalPxBackupKeycloakServiceName = "pxcentral-keycloak-http"
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
	// PxBackupKeycloakServiceName is the env var for the Keycloak service name within Px-Backup namespace, defaulting to GlobalPxBackupKeycloakServiceName
	PxBackupKeycloakServiceName = "KEYCLOAK_SERVICE_NAME"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type CredentialRepresentation struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}

func NewPasswordCredential(value string, temporary bool) *CredentialRepresentation {
	return &CredentialRepresentation{
		Type:      "password",
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
		return nil, pxbutils.ProcessError(err, pxbutils.StructToString(httpRequest))
	}
	httpRequest.Header = headers
	client := &http.Client{}
	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return nil, pxbutils.ProcessError(err, pxbutils.StructToString(httpRequest))
	}
	defer func() {
		err := httpResponse.Body.Close()
		if err != nil {
			log.Errorf("error closing HTTP response body: %v", pxbutils.ProcessError(err, pxbutils.StructToString(httpResponse)))
		}
	}()
	responseBody, err = io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, pxbutils.ProcessError(err, pxbutils.StructToString(httpResponse))
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
		return "", err
	}
	reqURL := fmt.Sprintf("%s/protocol/openid-connect/token", keycloakEndPoint)
	headers := make(http.Header)
	headers.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := ProcessHTTPRequest(ctx, POST, reqURL, headers, strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}
	token := &TokenResponse{}
	err = json.Unmarshal(response, &token)
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}

func GetCommonHTTPHeaders(ctx context.Context, username string, password string) (http.Header, error) {
	token, err := GetToken(ctx, username, password)
	if err != nil {
		debugStruct := struct {
			username string
			password string
		}{
			username: username,
			password: "", // password left blank on purpose
		}
		return nil, pxbutils.ProcessError(err, pxbutils.StructToString(debugStruct))
	}
	headers := make(http.Header)
	headers.Add("Authorization", fmt.Sprintf("Bearer %v", token))
	headers.Add("Content-Type", "application/json")
	return headers, nil
}

func AddUserByPassword(ctx context.Context, username string, firstName string, lastName string, email string, enabled bool, password string, temporary bool) error {
	keycloakEndPoint, err := GetKeycloakEndPoint(true)
	if err != nil {
		return pxbutils.ProcessError(err)
	}
	requestURL := fmt.Sprintf("%s/users", keycloakEndPoint)
	headers, err := GetCommonHTTPHeaders(ctx, GlobalPxCentralAdminUsername, GlobalPxCentralAdminPassword)
	if err != nil {
		return pxbutils.ProcessError(err)
	}
	userRepresentation := &UserRepresentation{
		Name:      username,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Enabled:   enabled,
		Credentials: []CredentialRepresentation{
			*NewPasswordCredential(password, temporary),
		},
	}
	userBytes, err := json.Marshal(userRepresentation)
	if err != nil {
		return pxbutils.ProcessError(err)
	}
	response, err := ProcessHTTPRequest(ctx, POST, requestURL, headers, strings.NewReader(string(userBytes)))
	log.Infof("response %v", response)
	if err != nil {
		return err
	}
	log.Infof("User [%s] created", username)
	return nil
}

// GetPxBackupNamespace returns namespace of px-backup deployment.
func GetPxBackupNamespace() (string, error) {
	allServices, err := core.Instance().ListServices("", metav1.ListOptions{})
	if err != nil {
		return "", pxbutils.ProcessError(err)
	}
	for _, svc := range allServices.Items {
		if svc.Name == GlobalPxBackupServiceName {
			return svc.Namespace, nil
		}
	}
	err = fmt.Errorf("cannot find Px-Backup service [%s] from list of services", GlobalPxBackupServiceName)
	return "", pxbutils.ProcessError(err)
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
	secret, err := k8s.Instance().GetSecret(name, ns)
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
		return "", pxbutils.ProcessError(err)
	}
	secret, err := k8s.Instance().GetSecret(GlobalPxCentralAdminSecretName, pxbNamespace)
	if err != nil {
		debugStruct := struct {
			PxbNamespace string
		}{
			PxbNamespace: pxbNamespace,
		}
		return "", pxbutils.ProcessError(err, pxbutils.StructToString(debugStruct))
	}
	PxCentralAdminPwd := string(secret.Data["credential"])
	if PxCentralAdminPwd == "" {
		err = fmt.Errorf("%s secret is empty", GlobalPxCentralAdminSecretName)
		return "", pxbutils.ProcessError(err)
	}
	return PxCentralAdminPwd, nil
}

func init() {
	str, err := GetPxCentralAdminPassword()
	if err != nil {
		log.Errorf("Error fetching password from secret: %v", err)
	}
	GlobalPxCentralAdminPassword = str
	log.Infof("GlobalPxCentralAdminPassword %s", GlobalPxCentralAdminPassword)
}
