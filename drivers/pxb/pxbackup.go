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
	User                     *User
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

type Keycloak struct {
	HTTPClient *http.Client
}

func (k *Keycloak) GetURL(admin bool, route string) (string, error) {
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
		oidcSecret, err := core.Instance().GetSecret(oidcSecretName, k.PxBackup.Spec.Namespace)
		if err != nil {
			return "", ProcessError(err)
		}
		oidcEndpoint := string(oidcSecret.Data[EnvPxBackupOIDCEndpoint])
		// Construct the fully qualified domain name (FQDN) for the Keycloak service to
		// ensure DNS resolution within Kubernetes, especially for requests originating
		// from different namespace
		replacement := fmt.Sprintf("%s.%s.svc.cluster.local", PxBackupKeycloakServiceName, k.PxBackup.Spec.Namespace)
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

func (k *Keycloak) Invoke(ctx context.Context, method string, admin bool, route string, body interface{}, headerMap map[string]string) ([]byte, error) {
	reqURL, err := k.GetURL(admin, route)
	if err != nil {
		return nil, ProcessError(err)
	}
	reqBody, err := ToByteArray(body)
	if err != nil {
		return nil, ProcessError(err)
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bytes.NewReader(reqBody))
	if err != nil {
		debugMap := DebugMap{}
		debugMap.Add("ReqURL", reqURL)
		return nil, ProcessError(err, debugMap.String())
	}
	for key, val := range headerMap {
		req.Header.Set(key, val)
	}
	resp, err := k.HTTPClient.Do(req)
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

func (k *Keycloak) GetAccessToken(ctx context.Context, username, password string) (string, error) {
	route := "/protocol/openid-connect/token"
	values := make(url.Values)
	values.Set("username", username)
	values.Set("password", password)
	values.Set("grant_type", "password")
	values.Set("client_id", "pxcentral")
	values.Set("token-duration", "365d")
	headerMap := make(map[string]string)
	headerMap["Content-Type"] = "application/x-www-form-urlencoded"
	respBody, err := k.Invoke(ctx, "POST", false, route, values.Encode(), headerMap)
	if err != nil {
		return "", ProcessError(err)
	}
	token := &TokenRepresentation{}
	err = json.Unmarshal(respBody, &token)
	if err != nil {
		return "", ProcessError(err)
	}
	return token.AccessToken, nil
}

func (k *Keycloak) GetPxCentralAdminAccessToken(ctx context.Context) (string, error) {
	pxCentralAdminPassword, err := k.GetPxCentralAdminPassword()
	if err != nil {
		return "", ProcessError(err)
	}
	pxCentralAdminToken, err := k.GetAccessToken(ctx, PxCentralAdminUsername, pxCentralAdminPassword)
	if err != nil {
		return "", ProcessError(err)
	}
	return pxCentralAdminToken, nil
}

func UpdatePxBackupAdminSecret(token string) error {
	pxBackupAdminSecret, err := core.Instance().GetSecret(PxBackupAdminSecretName, k.Namespace)
	if err != nil {
		return ProcessError(err)
	}
	pxBackupAdminSecret.Data[PxBackupOrgToken] = []byte(token)
	_, err = core.Instance().UpdateSecret(pxBackupAdminSecret)
	if err != nil {
		return ProcessError(err)
	}
	return nil
}

func GetPxCentralAdminCtxFromSecret(ctx context.Context, update bool) (context.Context, error) {
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

type User struct {
	PxBackup              *PxBackup
	Spec                  *keycloak.UserRepresentation
	OrganizationDataStore *generics.DataStore[*Organization]
}

type PxBackupSpec struct {
	Namespace    string
	OIDCEndpoint string
}

type PxBackup struct {
	Spec          *PxBackupSpec
	UserDataStore *generics.DataStore[*User]
}
