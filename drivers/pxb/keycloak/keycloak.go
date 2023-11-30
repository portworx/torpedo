package keycloak

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/backup"
	"github.com/portworx/torpedo/drivers/pxb/pxbutils"
	"os"
	"strings"
)

var instance *Keycloak

type UserCredential struct {
	Username string
	Password string
}

type Keycloak struct {
	BaseURL        string
	UserCredential UserCredential
}

func (k *Keycloak) SignIn() error {
	return nil
}

func Instance() *Keycloak {
	return instance
}

const (
	realmPath string = "auth/realms"
	adminPath string = "admin"
	realmName string = "master"
)

const (
	// PxBackupOIDCSecret is the Kubernetes secret storing OIDC (OpenID Connect) credentials
	PxBackupOIDCSecret = "pxc-backup-secret"
	// PxBackupOIDCEndpointKey is the key in PxBackupOIDCSecret for the OIDC endpoint
	PxBackupOIDCEndpointKey = "OIDC_ENDPOINT"
	// PxBackupKeycloakService is the Kubernetes service for Keycloak-based user authentication
	PxBackupKeycloakService = "pxcentral-keycloak-http"
)

const (
	// EnvPxCentralUIURL is the environment variable key for the Px-Central UI URL
	// Example: http://pxcentral-keycloak-http:80
	EnvPxCentralUIURL = "PX_CENTRAL_UI_URL"
)

func (k *Keycloak) GetAdminAndNonAdminURL(pxbNamespace string) error {
	baseURL := ""
	pxCentralUIURL := os.Getenv(EnvPxCentralUIURL)
	// The condition checks whether pxCentralUIURL is set. This condition is added to
	// handle scenarios where Torpedo is not running as a pod in the cluster. In such
	// cases, gRPC calls pxcentral-keycloak-http:80 would fail when made from a VM or
	// local machine using the Ginkgo CLI.
	if pxCentralUIURL != "" && len(pxCentralUIURL) > 0 {
		baseURL = pxCentralUIURL
	} else {
		oidcSecret, err := core.Instance().GetSecret(PxBackupOIDCSecret, pxbNamespace)
		if err != nil {
			return pxbutils.ProcessError(err)
		}
		oidcEndpoint := string(oidcSecret.Data[PxBackupOIDCEndpointKey])
		// Construct the fully qualified domain name (FQDN) for the Keycloak service to
		// ensure DNS resolution within Kubernetes, especially for requests originating
		// from different pxbNamespace
		keycloakFQDN := fmt.Sprintf("%s.%s.svc.cluster.local", PxBackupKeycloakService, pxbNamespace)
		baseURL = strings.Replace(oidcEndpoint, PxBackupKeycloakService, keycloakFQDN, 1)
	}
	k.AdminURL = fmt.Sprintf("%s/%s/%s/%s", baseURL, realmPath, adminPath, realmName)
	k.NonAdminURL = fmt.Sprintf("%s/%s/%s", baseURL, realmPath, realmName)
	return nil
}

func Init() error {
	pxbNamespace, err := pxbutils.GetPxBackupNamespace()
	if err != nil {
		return pxbutils.ProcessError(err)
	}
	AdminURL, NonAdminURL, err := GetAdminAndNonAdminURL(pxbNamespace)
	if err != nil {
		debugMap := pxbutils.DebugMap{}
		debugMap.Add("pxbNamespace", pxbNamespace)
		return pxbutils.ProcessError(err, debugMap.String())
	}
	PxCentralAdminPassword, err := backup.GetPxCentralAdminPwd()
	if err != nil {
		return pxbutils.ProcessError(err)
	}
	return nil
}
