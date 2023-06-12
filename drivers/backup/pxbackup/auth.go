package pxbackup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/pkg/log"
	"google.golang.org/grpc/metadata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// PxCentralAdminUser px central admin
	PxCentralAdminUser = "px-central-admin"
	// PxCentralAdminSecretName secret for PxCentralAdminUser
	PxCentralAdminSecretName = "px-central-admin"
	// PxCentralAdminSecretNamespace namespace of PxCentralAdminSecretName
	PxCentralAdminSecretNamespace = "px-backup"
	/// httpTimeout timeout for http request
	httpTimeout = 1 * time.Minute
	// DefaultsecretName - Default secret name for px-backup
	DefaultsecretName = "pxc-backup-secret"
	// Issuer - OIDC issuer
	Issuer = "OIDC_ENDPOINT"
)

const (
	// AuthTokenType of incoming auth token
	AuthTokenType = "bearer"
	// AuthHeader incoming auth request
	AuthHeader = "authorization"
	// OrgToken key
	OrgToken = "PX_BACKUP_ORG_TOKEN"
	// AdminTokenSecretName which has admin user jwt token information
	AdminTokenSecretName = "px-backup-admin-secret"
	// AdminTokenSecretNamespace which has admin user jwt token information
	AdminTokenSecretNamespace = "px-backup"

	defaultWaitTimeout  time.Duration = 30 * time.Second
	defaultWaitInterval time.Duration = 5 * time.Second
	// OidcSecretName where secrets for OIDC auth cred info resides
	oidcSecretName = "SECRET_NAME"
	// PxCentralUI URL Eg: http://<IP>:<Port>
	PxCentralUIURL = "PX_CENTRAL_UI_URL"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

// Doc ref: https://www.keycloak.org/docs-api/5.0/rest-api/index.html#_rolerepresentation
// Not all the fields are used and below is sample response obtained from keycloak
// {
//        "id": "12bfd2ee-bd3d-4260-809b-c288669ed5b1",
//        "name": "px-backup-app.user",
//        "description": "Portworx px-backup-app.user user role",
//        "composite": false,
//        "clientRole": false,
//        "containerId": "master"
//    },

// KeycloakRoleRepresentation role repsetaton struct
type KeycloakRoleRepresentation struct {
	ID          string                       `json:"id"`
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Composite   bool                         `json:"composite"`
	ClientRole  bool                         `json:"clientRole"`
	ContainerID string                       `json:"containerId"`
	Attributes  map[string]string            `json:"attributes"`
	Composites  RoleRespresentationComposite `json:"composites"`
}

// RoleRespresentationComposite composite role rep
type RoleRespresentationComposite struct {
	Client map[string]string `json:"client"`
	Realm  []string          `json:"realm"`
}

// KeycloakUserRepresentation user representation
type KeycloakUserRepresentation struct {
	ID            string                    `json:"id"`
	Name          string                    `json:"username"`
	FirstName     string                    `json:"firstName"`
	LastName      string                    `json:"lastName"`
	EmailVerified bool                      `json:"emailVerified"`
	Enabled       bool                      `json:"enabled"`
	Email         string                    `json:"email"`
	Credentials   []KeycloakUserCredentials `json:"credentials"`
}

// KeycloakGroupMemberRepresentation user representation
type KeycloakGroupMemberRepresentation struct {
	ID                         string                                 `json:"id"`
	CreatedTimestamp           int64                                  `json:"createdTimestamp"`
	Name                       string                                 `json:"username"`
	Totp                       bool                                   `json:"totp"`
	FirstName                  string                                 `json:"firstName"`
	LastName                   string                                 `json:"lastName"`
	EmailVerified              bool                                   `json:"emailVerified"`
	Enabled                    bool                                   `json:"enabled"`
	Email                      string                                 `json:"email"`
	Credentials                []KeycloakUserCredentials              `json:"credentials"`
	DisableableCredentialTypes []string                               `json:"disableableCredentialTypes"`
	RequiredActions            []RequiredActionProviderRepresentation `json:"requiredActions"`
	NotBefore                  int32                                  `json:"notBefore"`
}

// RequiredActionProviderRepresentation
type RequiredActionProviderRepresentation struct {
	Alias         string            `json:"alias"`
	Config        map[string]string `json:"config"`
	DefaultAction bool              `json:"defaultAction"`
	Enabled       bool              `json:"enabled"`
	Name          string            `json:"name"`
	Priority      int32             `json:"priority"`
	ProviderId    string            `json:"providerId"`
}

// KeycloakUserCredentials user credentials
type KeycloakUserCredentials struct {
	// Type is "password"
	Type string `json:"type"`
	// Temporary is the password temporary
	Temporary bool `json:"temporary"`
	// Value password for the user
	Value string `json:"value"`
}

// KeycloakGroupRepresentation group representation
type KeycloakGroupRepresentation struct {
	Name      string   `json:"name"`
	ID        string   `json:"id"`
	Path      string   `json:"path"`
	SubGroups []string `json:"subGroups"`
}

// KeycloakGroupAdd adding group
type KeycloakGroupAdd struct {
	Name string `json:"name"`
}

// KeycloakGroupToUser representation of adding group to user
type KeycloakGroupToUser struct {
	UserID  string `json:"userId"`
	GroupID string `json:"groupId"`
	Realm   string `json:"realm"`
}

type PXBAuth struct {
	// pxCentralAdminPwd is password of PxCentralAdminUser
	pxCentralAdminPwd   string
	pxCentralAdminToken string

	k8sCore core.Ops
}

// getOidcSecretName returns OIDC secret name
func getOidcSecretName() string {
	name := os.Getenv(oidcSecretName)
	if name == "" {
		name = DefaultsecretName
	}
	return name
}

func (pa *PXBAuth) getKeycloakEndPoint(admin bool) (string, error) {
	keycloakEndpoint := os.Getenv(PxCentralUIURL)
	// This condition is added for cases when torpedo is not running as a pod in the cluster
	// Since gRPC calls to pxcentral-keycloak-http:80 would fail while running from a VM or local machine using ginkgo CLI
	// This condition will check if there is an Env variable set
	if keycloakEndpoint != " " && len(keycloakEndpoint) > 0 {
		if admin {
			// admin url: http://pxcentral-keycloak-http:80/auth/realms/master
			// non-adming url: http://pxcentral-keycloak-http:80/auth/admin/realms/master
			newURL := fmt.Sprintf("%s/auth/admin/realms/master", keycloakEndpoint)
			return newURL, nil
		} else {
			newURL := fmt.Sprintf("%s/auth/realms/master", keycloakEndpoint)
			return newURL, nil
		}
	}
	name := getOidcSecretName()
	ns, err := pa.GetPxBackupNamespace()
	if err != nil {
		return "", err
	}
	// check and validate oidc details
	secret, err := pa.k8sCore.GetSecret(name, ns)
	if err != nil {
		return "", err
	}
	url := string(secret.Data[Issuer])
	// Expand the service name for K8S DNS resolution, for keycloak requests from different ns
	splitURL := strings.Split(url, ":")
	splitURL[1] = fmt.Sprintf("%s.%s.svc.cluster.local", splitURL[1], ns)
	url = strings.Join(splitURL, ":")
	// url: http://pxcentral-keycloak-http.px-backup.svc.cluster.local:80/auth/realms/master
	if admin {
		// admin url: http://pxcentral-keycloak-http.px-backup.svc.cluster.local:80/auth/realms/master
		// non-adming url: http://pxcentral-keycloak-http.px-backup.svc.cluster.local:80/auth/admin/realms/master
		split := strings.Split(url, "auth")
		newURL := fmt.Sprintf("%sauth/admin%s", split[0], split[1])
		return newURL, nil
	}
	return string(url), nil

}

// GetPxBackupNamespace returns namespace of px-backup deployment.
func (pa *PXBAuth) GetPxBackupNamespace() (string, error) {
	allServices, err := pa.k8sCore.ListServices("", metav1.ListOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get list of services. Err: %v", err)
	}
	for _, svc := range allServices.Items {
		if svc.Name == pxbServiceName {
			return svc.Namespace, nil
		}
	}
	return "", fmt.Errorf("can't find PxBackup service [%s] from list of services", pxbServiceName)
}

// GetNewToken fetches a new JWT token for given user credentials. Previous tokens aren't invalidated
func (pa *PXBAuth) GetNewToken(userName, password string) (string, error) {
	values := make(url.Values)
	values.Set("client_id", "pxcentral")
	values.Set("username", userName)
	values.Set("password", password)
	values.Set("grant_type", "password")
	values.Set("token-duration", "365d")
	keycloakEndPoint, err := pa.getKeycloakEndPoint(false)
	if err != nil {
		return "", err
	}
	reqURL := fmt.Sprintf("%s/protocol/openid-connect/token", keycloakEndPoint)
	method := "POST"
	headers := make(http.Header)
	headers.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := processHTTPRequest(method, reqURL, headers, strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}

	token := &tokenResponse{}
	err = json.Unmarshal(response, &token)
	if err != nil {
		return "", err
	}

	accessToken := token.AccessToken
	if accessToken == "" {
		return "", fmt.Errorf("AccessToken field was empty in JSON of unmarshalled response")
	}

	return accessToken, nil
}

// GetCommonHTTPHeaders populates http header
func (pa *PXBAuth) GetCommonHTTPHeaders(userName, password string) (http.Header, error) {
	fn := "GetCommonHTTPHeaders"
	token, err := pa.GetNewToken(userName, password)
	log.Debugf("new token obtained for user [%s] is \"%v\"", userName, token)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return nil, err
	}
	headers := make(http.Header)
	headers.Add("Authorization", fmt.Sprintf("Bearer %v", token))
	headers.Add("Content-Type", "application/json")

	return headers, nil
}

// SetPxCentralAdminPwd fetches password from PxCentralAdminUser from secret
func (pa *PXBAuth) SetPxCentralAdminPwd() error {
	pxbNamespace, err := pa.GetPxBackupNamespace()
	if err != nil {
		return err
	}
	secret, err := pa.k8sCore.GetSecret(PxCentralAdminSecretName, pxbNamespace)
	if err != nil {
		return err
	}

	pxCentralAdminPwd := string(secret.Data["credential"])
	if pxCentralAdminPwd == "" {
		return fmt.Errorf("px-central-admin secret is empty")
	}
	pa.pxCentralAdminPwd = pxCentralAdminPwd
	return nil
}

// SetPxCentralAdminPwd fetches password from PxCentralAdminUser from secret
func (pa *PXBAuth) GetPxCentralAdminPwd() (string, error) {
	if pa.pxCentralAdminPwd == "" {
		return "", fmt.Errorf("password is empty")
	}
	return pa.pxCentralAdminPwd, nil
}

// GetAllRoles lists all the available role in keycloak
func (pa *PXBAuth) GetAllRoles() ([]KeycloakRoleRepresentation, error) {
	fn := "GetAllRoles"
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return nil, err
	}
	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return nil, err
	}
	reqURL := fmt.Sprintf("%s/roles", keycloakEndPoint)
	method := "GET"
	response, err := processHTTPRequest(method, reqURL, headers, nil)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return nil, err
	}
	var roles []KeycloakRoleRepresentation
	err = json.Unmarshal(response, &roles)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return nil, err
	}

	return roles, nil
}

// GetRolesForUser lists all the available roles in keycloak for the provided username
func (pa *PXBAuth) GetRolesForUser(userName string) ([]KeycloakRoleRepresentation, error) {
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		return nil, err
	}
	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return nil, err
	}
	userID, err := pa.FetchIDOfUser(userName)
	if err != nil {
		return nil, err
	}
	reqURL := fmt.Sprintf("%s/users/%s/role-mappings/realm/composite", keycloakEndPoint, userID)
	method := "GET"
	response, err := processHTTPRequest(method, reqURL, headers, nil)
	if err != nil {
		return nil, err
	}
	var roles []KeycloakRoleRepresentation
	err = json.Unmarshal(response, &roles)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

type PxBackupRole string

const (
	ApplicationOwner    PxBackupRole = "px-backup-app.admin"
	ApplicationUser                  = "px-backup-app.user"
	InfrastructureOwner              = "px-backup-infra.admin"
	DefaultRoles                     = "default-roles-master"
)

// GetRoleID gets role ID for a given role
func (pa *PXBAuth) GetRoleID(role PxBackupRole) (string, error) {
	fn := "GetRoleID"
	// Fetch all roles
	roles, err := pa.GetAllRoles()
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return "", err
	}
	// Now fetch the current role ID
	var clientID string
	for _, r := range roles {
		if r.Name == string(role) {
			clientID = r.ID
			break
		}
	}

	return clientID, nil
}

// GetUserRole fetches roles for a given user
func (pa *PXBAuth) GetUserRole(userName string) error {
	fn := "GetUserRole"
	// First fetch all users to get the client id for the client
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}

	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return err
	}
	reqURL := fmt.Sprintf("%s/users", keycloakEndPoint)
	method := "GET"
	response, err := processHTTPRequest(method, reqURL, headers, nil)
	if err != nil {
		return err
	}
	var users []KeycloakUserRepresentation
	err = json.Unmarshal(response, &users)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}

	var clientID string
	for _, user := range users {
		if user.Name == userName {
			clientID = user.ID
			break
		}
	}
	// Now fetch all the roles for the fetched client ID
	keycloakEndPoint, err = pa.getKeycloakEndPoint(true)
	if err != nil {
		return err
	}
	reqURL = fmt.Sprintf("%s/users/%s/role-mappings/realm", keycloakEndPoint, clientID)
	method = "GET"
	response, err = processHTTPRequest(method, reqURL, headers, nil)
	if err != nil {
		return err
	}
	var r []KeycloakRoleRepresentation
	err = json.Unmarshal(response, &r)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}

	return nil
}

// FetchIDOfUser fetches ID for a given user
func (pa *PXBAuth) FetchIDOfUser(userName string) (string, error) {
	fn := "FetchIDOfUser"
	// First fetch all users to get the client id for the client
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return "", err
	}
	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return "", err
	}
	// TODO Need to increase the limit
	reqURL := fmt.Sprintf("%s/users", keycloakEndPoint)
	method := "GET"
	response, err := processHTTPRequest(method, reqURL, headers, nil)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return "", err
	}
	var users []KeycloakUserRepresentation
	err = json.Unmarshal(response, &users)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return "", err
	}

	var clientID string
	for _, user := range users {
		if user.Name == userName {
			clientID = user.ID
			break
		}
	}
	log.Infof("Fetching ID of user %s - %s", userName, clientID)
	return clientID, nil
}

// AddRoleToUser assigning a given role to an existing user
func (pa *PXBAuth) AddRoleToUser(userName string, role PxBackupRole, description string) error {
	fn := "AddRoleToUser"
	// First fetch the client ID of the user
	clientID, err := pa.FetchIDOfUser(userName)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}
	// Fetch the role ID
	roleID, err := pa.GetRoleID(role)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}

	// Frame the role struct to be assigned
	var kRoles []KeycloakRoleRepresentation
	kRole := KeycloakRoleRepresentation{
		ID:          roleID,
		ClientRole:  false,
		Composite:   false,
		ContainerID: "master",
		Description: description,
		Name:        string(role),
	}
	kRoles = append(kRoles, kRole)
	roleBytes, err := json.Marshal(&kRoles)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}
	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return err
	}
	reqURL := fmt.Sprintf("%s/users/%s/role-mappings/realm", keycloakEndPoint, clientID)
	method := "POST"
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}
	_, err = processHTTPRequest(method, reqURL, headers, strings.NewReader(string(roleBytes)))
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}

	return nil
}

// AddRoleToGroup assigning a given role to an existing group
func (pa *PXBAuth) AddRoleToGroup(groupName string, role PxBackupRole, description string) error {
	// First fetch the client ID of the user
	groupID, err := pa.FetchIDOfGroup(groupName)
	if err != nil {
		return err
	}
	// Fetch the role ID
	roleID, err := pa.GetRoleID(role)
	if err != nil {
		return err
	}

	// Frame the role struct to be assigned
	var kRoles []KeycloakRoleRepresentation
	kRole := KeycloakRoleRepresentation{
		ID:          roleID,
		ClientRole:  false,
		Composite:   false,
		ContainerID: "master",
		Description: description,
		Name:        string(role),
	}
	kRoles = append(kRoles, kRole)
	roleBytes, err := json.Marshal(&kRoles)
	if err != nil {
		return err
	}
	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return err
	}
	reqURL := fmt.Sprintf("%s/groups/%s/role-mappings/realm", keycloakEndPoint, groupID)
	method := "POST"
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		return err
	}
	_, err = processHTTPRequest(method, reqURL, headers, strings.NewReader(string(roleBytes)))
	if err != nil {
		return err
	}

	return nil
}

// DeleteRoleFromUser deleting role from a user
func (pa *PXBAuth) DeleteRoleFromUser(userName string, role PxBackupRole, description string) error {
	fn := "DeleteRoleFromUser"
	// First fetch the user ID of the user
	clientID, err := pa.FetchIDOfUser(userName)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}
	// Fetch the role ID
	roleID, err := pa.GetRoleID(role)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}

	// Frame the role struct to be assigned
	var kRoles []KeycloakRoleRepresentation
	kRole := KeycloakRoleRepresentation{
		ID:          roleID,
		ClientRole:  false,
		Composite:   false,
		ContainerID: "master",
		Description: description,
		Name:        string(role),
	}
	kRoles = append(kRoles, kRole)
	roleBytes, err := json.Marshal(&kRoles)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}
	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return err
	}
	reqURL := fmt.Sprintf("%s/users/%s/role-mappings/realm", keycloakEndPoint, clientID)
	method := "DELETE"
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		return err
	}
	_, err = processHTTPRequest(method, reqURL, headers, strings.NewReader(string(roleBytes)))
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}

	return nil
}

// DeleteRoleFromGroup deleting role from a group
func (pa *PXBAuth) DeleteRoleFromGroup(groupName string, role PxBackupRole, description string) error {
	// First fetch the user ID of the user
	groupID, err := pa.FetchIDOfGroup(groupName)
	if err != nil {
		return err
	}
	// Fetch the role ID
	roleID, err := pa.GetRoleID(role)
	if err != nil {
		return err
	}

	// Frame the role struct to be assigned
	var kRoles []KeycloakRoleRepresentation
	kRole := KeycloakRoleRepresentation{
		ID:          roleID,
		ClientRole:  false,
		Composite:   false,
		ContainerID: "master",
		Description: description,
		Name:        string(role),
	}
	kRoles = append(kRoles, kRole)
	roleBytes, err := json.Marshal(&kRoles)
	if err != nil {
		return err
	}
	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return err
	}
	reqURL := fmt.Sprintf("%s/groups/%s/role-mappings/realm", keycloakEndPoint, groupID)
	method := "DELETE"
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		return err
	}
	_, err = processHTTPRequest(method, reqURL, headers, strings.NewReader(string(roleBytes)))
	if err != nil {
		return err
	}
	return nil
}

// AddUser adds a new user
func (pa *PXBAuth) AddUser(userName, firstName, lastName, email, password string) error {
	fn := "AddUser"
	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return err
	}
	reqURL := fmt.Sprintf("%s/users", keycloakEndPoint)
	method := "POST"
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}
	userRep := KeycloakUserRepresentation{
		Name:      userName,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Enabled:   true,
		Credentials: []KeycloakUserCredentials{
			{
				Type:      "password",
				Temporary: false,
				Value:     password,
			},
		},
	}
	userBytes, err := json.Marshal(&userRep)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}
	response, err := processHTTPRequest(method, reqURL, headers, strings.NewReader(string(userBytes)))
	log.Infof("Response for user [%s] creation - %s", userName, string(response))
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}
	log.Infof("User [%s] created", userName)
	return nil
}

// DeleteUser deletes a user with the provided userName
func (pa *PXBAuth) DeleteUser(userName string) error {
	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return err
	}
	userID, err := pa.FetchIDOfUser(userName)
	if err != nil {
		return err
	}
	reqURL := fmt.Sprintf("%s/users/%s", keycloakEndPoint, userID)
	method := "DELETE"
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		return err
	}

	response, err := processHTTPRequest(method, reqURL, headers, nil)
	log.Infof("Response for user [%s] deletion - %s", userName, string(response))
	if err != nil {
		return err
	}
	log.Infof("Deleted User - %s", userName)
	return nil
}

// GetPxCentralAdminToken gets token for "px-central-admin"
func (pa *PXBAuth) GetPxCentralAdminToken(newToken bool) (string, error) {
	if newToken || pa.pxCentralAdminToken == "" {
		token, err := pa.GetNewToken(PxCentralAdminUser, pa.pxCentralAdminPwd)
		log.Debugf("new token obtained for admin user [%s] is \"%v\"", PxCentralAdminUser, token)
		if err != nil {
			return "", err
		}
		pa.pxCentralAdminToken = token
		return token, nil
	} else {
		return pa.pxCentralAdminToken, nil
	}
}

// GetCtxWithToken getx ctx with passed token
func GetCtxWithToken(token string) context.Context {
	ctx := context.Background()
	md := metadata.New(map[string]string{
		AuthHeader: AuthTokenType + " " + token,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	return ctx
}

// GetPxCentralAdminCtx fetch px-central-admin context
func (pa *PXBAuth) GetPxCentralAdminCtx() (context.Context, error) {
	token, err := pa.GetPxCentralAdminToken(false)
	if err != nil {
		return nil, err
	}
	ctx := GetCtxWithToken(token)
	return ctx, nil
}

// UpdatePxCentralAdminTokenInSecret updating "px-backup-admin-secret" token with
// "px-central-admin" token. forceRefresh should be set to true if a new token needs to be generated
func (pa *PXBAuth) UpdatePxCentralAdminTokenInSecret(forceRefresh bool) error {
	pxCentralAdminToken, err := pa.GetPxCentralAdminToken(forceRefresh)
	if err != nil {
		return err
	}

	pxbNamespace, err := pa.GetPxBackupNamespace()
	if err != nil {
		return err
	}
	secret, err := pa.k8sCore.GetSecret(AdminTokenSecretName, pxbNamespace)
	if err != nil {
		return err
	}
	// Now update the token into "AdminTokenSecretName"
	secret.Data[OrgToken] = ([]byte(pxCentralAdminToken))
	secret, err = pa.k8sCore.UpdateSecret(secret)
	if err != nil {
		return err
	}

	return nil
}

// GetPxCentralAdminCtxFromSecret with provided name and namespace
func (pa *PXBAuth) GetPxCentralAdminCtxFromSecret() (context.Context, error) {
	token, err := pa.GetPxCentralAdminTokenFromSecret()
	if err != nil {
		if err.Error() == "admin token is empty" {
			log.Debugf("Token in AdminTokenSecret was empty. Setting new token")
			err := pa.UpdatePxCentralAdminTokenInSecret(false)
			if err != nil {
				return nil, err
			}
			// try getting it once again
			token, err := pa.GetPxCentralAdminTokenFromSecret()
			if err != nil {
				return nil, err
			}
			ctx := GetCtxWithToken(token)
			return ctx, nil
		} else {
			return nil, err
		}
	}
	ctx := GetCtxWithToken(token)
	return ctx, nil
}

// GetPxCentralAdminTokenFromSecret with provided name and namespace
func (pa *PXBAuth) GetPxCentralAdminTokenFromSecret() (string, error) {
	pxbNamespace, err := pa.GetPxBackupNamespace()
	if err != nil {
		return "", err
	}
	secret, err := pa.k8sCore.GetSecret(AdminTokenSecretName, pxbNamespace)
	if err != nil {
		return "", err
	}

	token := string(secret.Data[OrgToken])
	if token == "" {
		return "", fmt.Errorf("admin token is empty")
	}
	return token, nil
}

// GetAllGroups fetches all available groups
func (pa *PXBAuth) GetAllGroups() ([]KeycloakGroupRepresentation, error) {
	fn := "GetAllGroups"
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return nil, err
	}
	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return nil, err
	}
	reqURL := fmt.Sprintf("%s/groups", keycloakEndPoint)
	method := "GET"
	response, err := processHTTPRequest(method, reqURL, headers, nil)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return nil, err
	}
	var groups []KeycloakGroupRepresentation
	err = json.Unmarshal(response, &groups)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return nil, err
	}
	return groups, nil
}

func (pa *PXBAuth) GetAllUsers() ([]KeycloakUserRepresentation, error) {
	fn := "GetAllGroups"
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return nil, err
	}
	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return nil, err
	}
	reqURL := fmt.Sprintf("%s/users", keycloakEndPoint)
	method := "GET"
	response, err := processHTTPRequest(method, reqURL, headers, nil)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return nil, err
	}
	var users []KeycloakUserRepresentation
	err = json.Unmarshal(response, &users)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return nil, err
	}
	return users, nil
}

// GetMembersOfGroup fetches all available members of the group
func (pa *PXBAuth) GetMembersOfGroup(group string) ([]string, error) {
	fn := "GetMembersOfGroup"
	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return nil, err
	}
	groupID, err := pa.FetchIDOfGroup(group)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/groups/%s/members", keycloakEndPoint, groupID)
	method := "GET"
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return nil, err
	}

	response, err := processHTTPRequest(method, reqURL, headers, nil)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return nil, err
	}
	var members []string
	var users []KeycloakGroupMemberRepresentation
	err = json.Unmarshal(response, &users)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return nil, err
	}
	for _, userName := range users {
		members = append(members, userName.Name)
	}
	log.Debugf("list of members : %v", members)
	return members, nil
}

// AddGroup adds a new group
func (pa *PXBAuth) AddGroup(group string) error {
	fn := "AddGroup"
	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return err
	}
	reqURL := fmt.Sprintf("%s/groups", keycloakEndPoint)
	method := "POST"
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}
	groups := KeycloakGroupAdd{
		Name: group,
	}

	userBytes, err := json.Marshal(&groups)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}
	response, err := processHTTPRequest(method, reqURL, headers, strings.NewReader(string(userBytes)))
	log.Infof("Creating group response - %s", string(response))
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}
	log.Infof("Group [%s] created", group)
	return nil
}

// DeleteGroup adds a new group
func (pa *PXBAuth) DeleteGroup(group string) error {
	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return err
	}
	groupID, err := pa.FetchIDOfGroup(group)
	if err != nil {
		return err
	}
	reqURL := fmt.Sprintf("%s/groups/%s", keycloakEndPoint, groupID)
	method := "DELETE"
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		return err
	}
	_, err = processHTTPRequest(method, reqURL, headers, nil)
	if err != nil {
		return err
	}
	log.Infof("Deleted Group - %s", group)
	return nil
}

// Deletes Multiple groups
func (pa *PXBAuth) DeleteMultipleGroups(groups []string) error {

	var wg sync.WaitGroup
	for _, group := range groups {
		wg.Add(1)
		go func(group string) {
			defer wg.Done()
			err := pa.DeleteGroup(group)
			log.FailOnError(err, "Failed to create group - %v", group)

		}(group)
		log.Infof("Deleted Group - %s", group)
	}
	wg.Wait()

	return nil
}

// Deletes Multiple users
func (pa *PXBAuth) DeleteMultipleUsers(users []string) error {

	var wg sync.WaitGroup
	for _, user := range users {
		wg.Add(1)
		go func(user string) {
			defer wg.Done()
			err := pa.DeleteUser(user)
			log.FailOnError(err, "Failed to create group - %v", user)

		}(user)
		log.Infof("Deleted User - %s", user)
	}
	wg.Wait()

	return nil
}

// AddGroupToUser add group to a user
func (pa *PXBAuth) AddGroupToUser(user, group string) error {
	fn := "AddGroupToUser"
	groupID, err := pa.FetchIDOfGroup(group)
	if err != nil {
		return err
	}

	userID, err := pa.FetchIDOfUser(user)
	if err != nil {
		return err
	}

	keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
	if err != nil {
		return err
	}
	reqURL := fmt.Sprintf("%s/users/%s/groups/%s", keycloakEndPoint, userID, groupID)
	method := "PUT"
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}
	data := KeycloakGroupToUser{
		UserID:  userID,
		GroupID: groupID,
		Realm:   "master",
	}

	dataBytes, err := json.Marshal(&data)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}
	response, err := processHTTPRequest(method, reqURL, headers, strings.NewReader(string(dataBytes)))
	log.Infof("Adding user to group response - %s", string(response))
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return err
	}
	log.Infof("Added User [%s] to Group [%s]", user, group)
	return nil
}

// FetchIDOfGroup fetched ID of a group
func (pa *PXBAuth) FetchIDOfGroup(name string) (string, error) {
	groups, err := pa.GetAllGroups()
	if err != nil {
		return "", nil
	}

	var groupID string
	for _, group := range groups {
		if group.Name == name {
			groupID = group.ID
			break
		}
	}
	log.Infof("Fetching ID of group %s - %s", name, groupID)
	return groupID, nil
}

// FetchUserDetailsFromID fetches user name and email ID
func (pa *PXBAuth) FetchUserDetailsFromID(userID string) (string, string, error) {
	fn := "FetchUserDetailsFromID"

	// First fetch all users to get the client id for the client
	headers, err := pa.GetCommonHTTPHeaders(PxCentralAdminUser, pa.pxCentralAdminPwd)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return "", "", err
	}
	var userName string
	var email string
	f := func() (interface{}, bool, error) {
		keycloakEndPoint, err := pa.getKeycloakEndPoint(true)
		if err != nil {
			return nil, true, err
		}
		reqURL := fmt.Sprintf("%s/users", keycloakEndPoint)
		method := "GET"
		response, err := processHTTPRequest(method, reqURL, headers, nil)
		if err != nil {
			log.Errorf("%s: %v", fn, err)
			return nil, true, err
		}
		var users []KeycloakUserRepresentation
		err = json.Unmarshal(response, &users)
		if err != nil {
			log.Errorf("%s: %v", fn, err)
			return nil, true, err
		}

		for _, user := range users {
			if user.ID == userID {
				userName = user.Name
				email = user.Email
				break
			}
		}
		if userName == "" || email == "" {
			// In some case there might be no error but we got empty user name/email, retry again
			return nil, true, fmt.Errorf("got emptry user/email, retrying again")
		}

		return nil, false, nil
	}

	_, err = task.DoRetryWithTimeout(f, defaultWaitTimeout, defaultWaitInterval)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch user name/email: [%v]", err)
	}

	return userName, email, nil
}

func processHTTPRequest(
	method string,
	url string,
	headers http.Header,
	body io.Reader,
) ([]byte, error) {
	httpRequest, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	httpRequest.Header = headers
	client := &http.Client{
		Timeout: httpTimeout,
	}
	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := httpResponse.Body.Close()
		if err != nil {
			log.Debugf("Error closing http response body: %v", err)
		}
	}()

	respBodyBytes, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, err
	}
	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Response - Status code: [%d], Error message: [%s]", httpResponse.StatusCode, string(respBodyBytes))
	}
	return respBodyBytes, nil
}

func (pa *PXBAuth) GetPxCentralNonAdminCtx(username, password string) (context.Context, error) {
	token, err := pa.GetNewToken(username, password)
	log.Debugf("new token obtained for user [%s] is \"%v\"", username, token)
	if err != nil {
		return nil, err
	}
	ctx := GetCtxWithToken(token)
	return ctx, nil
}

func (pa *PXBAuth) GetRandomUserFromGroup(groupName string) (string, error) {
	fn := "GetRandomUserFromGroup"
	users, err := pa.GetMembersOfGroup(groupName)
	if err != nil {
		log.Errorf("%s: %v", fn, err)
		return "", err
	}
	rand.Seed(time.Now().Unix())
	userName := users[rand.Intn(len(users))]
	return userName, nil
}
