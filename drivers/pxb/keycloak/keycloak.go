package keycloak

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	. "github.com/portworx/torpedo/drivers/pxb/pxbutils"
	"os"
	"strings"
)

const (
	// EnvPxCentralUIURL is the environment variable key for the px-central UI URL.
	// Example: http://pxcentral-keycloak-http:80
	EnvPxCentralUIURL = "PX_CENTRAL_UI_URL"
)

const (
	// PxBackupOIDCSecret is the Kubernetes secret that contains OIDC (OpenID Connect) credentials
	PxBackupOIDCSecret = "pxc-backup-secret"
	// PxBackupOIDCEndpointKey is the key in PxBackupOIDCSecret for the OIDC endpoint
	PxBackupOIDCEndpointKey = "OIDC_ENDPOINT"
	// PxBackupKeycloakService is the Kubernetes service for Keycloak-based user authentication
	PxBackupKeycloakService = "pxcentral-keycloak-http"
)

type SignIn struct {
	Username string
	Password string
}

func NewSignIn(username string, password string) *SignIn {
	return &SignIn{
		Username: username,
		Password: password,
	}
}

func GetAdminAndNonAdminURL(namespace string) (string, string, error) {
	realmPath, adminPath, realmName := "auth/realms", "admin", "master"
	adminURL, nonAdminURL := "", ""
	pxCentralUIURL := os.Getenv(EnvPxCentralUIURL)
	// The condition checks whether pxCentralUIURL is set. This condition is added to
	// handle scenarios where Torpedo is not running as a pod in the cluster. In such
	// cases, gRPC calls pxcentral-keycloak-http:80 would fail when made from a VM or
	// local machine using the Ginkgo CLI.
	if pxCentralUIURL != "" && len(pxCentralUIURL) > 0 {
		adminURL = fmt.Sprintf("%s/%s/%s/%s", pxCentralUIURL, realmPath, adminPath, realmName)
		nonAdminURL = fmt.Sprintf("%s/%s/%s", pxCentralUIURL, realmPath, realmName)
	} else {
		oidcSecret, err := core.Instance().GetSecret(PxBackupOIDCSecret, namespace)
		if err != nil {
			return "", "", ProcessError(err)
		}
		oidcEndpoint := string(oidcSecret.Data[PxBackupOIDCEndpointKey])
		// Construct the fully qualified domain name (FQDN) for the Keycloak service to
		// ensure DNS resolution within Kubernetes, especially for requests originating
		// from different namespace
		keycloakFQDN := fmt.Sprintf("%s.%s.svc.cluster.local", PxBackupKeycloakService, namespace)
		newOIDCEndpoint := strings.Replace(oidcEndpoint, PxBackupKeycloakService, keycloakFQDN, 1)
		adminURL = fmt.Sprintf("%s/%s/%s/%s", newOIDCEndpoint, realmPath, adminPath, realmName)
		nonAdminURL = fmt.Sprintf("%s/%s/%s", newOIDCEndpoint, realmPath, realmName)
	}
	return adminURL, nonAdminURL, nil
}

func Init() error {
	return nil
}
