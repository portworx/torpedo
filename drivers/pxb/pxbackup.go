package pxb

import (
	"bytes"
	"context"
	"fmt"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/pxb/generics"
	"github.com/portworx/torpedo/drivers/pxb/keycloak"
	. "github.com/portworx/torpedo/drivers/pxb/pxbutils"
	"github.com/portworx/torpedo/pkg/log"
	"io"
	"net/http"
	"os"
	"strings"
)

type Organization struct {
	Spec                     *api.OrganizationObject
	BackupDataStore          *generics.DataStore[*api.BackupObject]
	BackupLocationDataStore  *generics.DataStore[*api.BackupLocationObject]
	BackupScheduleDataStore  *generics.DataStore[*api.BackupScheduleObject]
	ClusterDataStore         *generics.DataStore[*api.ClusterObject]
	SchedulePolicyDataStore  *generics.DataStore[*api.SchedulePolicyObject]
	RoleDataStore            *generics.DataStore[*api.RoleObject]
	RestoreDataStore         *generics.DataStore[*api.RestoreObject]
	RuleDataStore            *generics.DataStore[*api.RuleObject]
	CloudCredentialDataStore *generics.DataStore[*api.CloudCredentialObject]
}

type PxBackupSpec struct {
	Namespace string
}

type User struct {
	Spec                  *keycloak.UserRepresentation
	PxBackup              *PxBackup
	OrganizationDataStore *generics.DataStore[*Organization]
}

type PxBackup struct {
	Spec          *PxBackupSpec
	UserDataStore *generics.DataStore[*User]
}

func (b *PxBackup) GetKeycloakURL(admin bool, route string) (string, error) {
	reqURL := ""
	oidcSecretName := GetOIDCSecretName()
	pxCentralUIURL := os.Getenv(EnvPxCentralUIURL)
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
		oidcSecret, err := core.Instance().GetSecret(oidcSecretName, b.Spec.Namespace)
		if err != nil {
			return "", ProcessError(err)
		}
		oidcEndpoint := string(oidcSecret.Data[EnvPxBackupOIDCEndpoint])
		// Construct the fully qualified domain name (FQDN) for the Keycloak service to
		// ensure DNS resolution within Kubernetes, especially for requests originating
		// from different namespace
		replacement := BuildFQDN(PxBackupKeycloakServiceName, b.Spec.Namespace)
		newURL := strings.Replace(oidcEndpoint, PxBackupKeycloakServiceName, replacement, 1)
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

func (b *PxBackup) GetKeycloakResponse(ctx context.Context, method string, admin bool, route string, body interface{}, headerMap map[string]string) ([]byte, error) {
	reqURL, err := b.GetKeycloakURL(admin, route)
	if err != nil {
		return nil, ProcessError(err)
	}
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
	resp, err := keycloak.HTTPClient.Do(req)
	if err != nil {
		return nil, ProcessError(err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Errorf("failed to close response body. Err: [%v]", ProcessError(err))
		}
	}()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ProcessError(err)
	}
	statusCode := resp.StatusCode
	switch {
	case statusCode >= 200 && statusCode < 300:
		return respBody, nil
	default:
		reqURL, statusText := resp.Request.URL, http.StatusText(statusCode)
		err = fmt.Errorf("[%s] [%s] returned status [%d]: [%s]", method, reqURL, statusCode, statusText)
		return nil, ProcessError(err)
	}
}
