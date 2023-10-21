package pxb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers/pxb/generics"
	"github.com/portworx/torpedo/drivers/pxb/keycloak"
	. "github.com/portworx/torpedo/drivers/pxb/pxbutils"
	"github.com/portworx/torpedo/pkg/log"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Organization struct {
	Spec                     *api.OrganizationObject
	BackupDataStore          *generics.DataStore[*api.BackupObject]
	BackupLocationDataStore  *generics.DataStore[*api.BackupLocationObject]
	BackupScheduleDataStore  *generics.DataStore[*api.BackupScheduleObject]
	SchedulePolicyDataStore  *generics.DataStore[*api.SchedulePolicyObject]
	RoleDataStore            *generics.DataStore[*api.RoleObject]
	RuleDataStore            *generics.DataStore[*api.RuleObject]
	ClusterDataStore         *generics.DataStore[*api.ClusterObject]
	RestoreDataStore         *generics.DataStore[*api.RestoreObject]
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

func (b *PxBackup) BuildKeycloakURL(admin bool, route string) (string, error) {
	reqURL := ""
	oidcSecretName := GetOIDCSecretName()
	pxCentralUIURL := os.Getenv(EnvPxCentralUIURL)
	// The condition checks whether pxCentralUIURL is set. This condition is added to
	// handle scenarios where Torpedo is not running as a pod in the cluster. In such
	// cases, gRPC calls pxcentral-keycloak-http:80 would fail when made from a VM or
	// local machine using the Ginkgo CLI.
	if len(pxCentralUIURL) > 0 {
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

func (b *PxBackup) ProcessKeycloakRequest(ctx context.Context, method string, admin bool, route string, body interface{}, headerMap map[string]string) ([]byte, error) {
	reqURL, err := b.BuildKeycloakURL(admin, route)
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
	statusCode := resp.StatusCode
	switch {
	case statusCode >= 200 && statusCode < 300:
		return io.ReadAll(resp.Body)
	default:
		reqURL, statusText := resp.Request.URL, http.StatusText(statusCode)
		err = fmt.Errorf("[%s] [%s] returned status [%d]: [%s]", method, reqURL, statusCode, statusText)
		return nil, ProcessError(err)
	}
}

func (b *PxBackup) GetKeycloakAccessToken(ctx context.Context, username, password string) (string, error) {
	route := "/protocol/openid-connect/token"
	values := make(url.Values)
	values.Set("username", username)
	values.Set("password", password)
	values.Set("grant_type", "password")
	values.Set("client_id", "pxcentral")
	values.Set("token-duration", "365d")
	headerMap := make(map[string]string)
	headerMap["Content-Type"] = "application/x-www-form-urlencoded"
	body, err := b.ProcessKeycloakRequest(ctx, "POST", false, route, values.Encode(), headerMap)
	if err != nil {
		return "", ProcessError(err)
	}
	token := &keycloak.TokenRepresentation{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		debugMap := DebugMap{}
		debugMap.Add("Body", body)
		return "", ProcessError(err, debugMap.String())
	}
	return token.AccessToken, nil
}
