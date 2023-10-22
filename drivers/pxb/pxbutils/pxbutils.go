package pxbutils

import (
	"encoding/json"
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"reflect"
	"runtime"
	"strings"
)

const (
	// PxCentralAdminUsername is the username for px-central-admin user
	PxCentralAdminUsername = "px-central-admin"
	// PxBackupAuthHeader is the HTTP header used for authentication in Px-Backup requests
	PxBackupAuthHeader = "authorization"
	// PxBackupAuthTokenType is the type of authentication token in Px-Backup requests
	PxBackupAuthTokenType = "bearer"
	// PxBackupKeycloakServiceName is the Kubernetes service that facilitates user authentication
	// through Keycloak in Px-Backup
	PxBackupKeycloakServiceName = "pxcentral-keycloak-http"
	// PxBackupOrgToken is the organization token key within PxBackupAdminSecretName
	PxBackupOrgToken = "PX_BACKUP_ORG_TOKEN"
	// PxBackupAdminSecretName is the Kubernetes secret that stores Px-Backup admin token
	PxBackupAdminSecretName = "px-backup-admin-secret"
	// PxBackupServiceName is the Kubernetes service within Px-Backup namespace
	PxBackupServiceName = "px-backup"
	// PxCentralAdminSecretName is the Kubernetes secret that stores px-central-admin credentials
	PxCentralAdminSecretName = "px-central-admin"
)

const (
	// TorpedoWorkDirectory is the working directory inside the Torpedo container
	TorpedoWorkDirectory = "/go/src/github.com/portworx/"
)

const (
	// DefaultOIDCSecretName is the fallback Kubernetes secret in case EnvPxBackupOIDCSecretName is empty
	DefaultOIDCSecretName = "pxc-backup-secret"
)

const (
	// EnvPxBackupOIDCEndpoint is the env var for the OIDC endpoint
	EnvPxBackupOIDCEndpoint = "OIDC_ENDPOINT"
	// EnvPxBackupOIDCSecretName is the env var for the OIDC secret within px-backup namespace
	EnvPxBackupOIDCSecretName = "SECRET_NAME"
	// EnvPxCentralUIURL is the env var for the px-central UI URL. Example: http://pxcentral-keycloak-http:80
	EnvPxCentralUIURL = "PX_CENTRAL_UI_URL"
)

type DebugMap map[string]interface{}

func (m *DebugMap) Add(key string, value interface{}) {
	(*m)[key] = value
}

func (m *DebugMap) String() string {
	return ToString(m)
}

// ProcessError formats the error message with caller information and an optional debug message
func ProcessError(err error, debugMessage ...string) error {
	if err == nil {
		return nil
	}
	_, file, line, _ := runtime.Caller(1)
	file = strings.TrimPrefix(file, TorpedoWorkDirectory)
	callerInfo := fmt.Sprintf("%s:%d", file, line)
	debugInfo := "no debug message"
	if len(debugMessage) > 0 {
		debugInfo = "debug message: " + debugMessage[0]
	}
	return fmt.Errorf("%s\n  at %s <-> %s", err.Error(), callerInfo, debugInfo)
}

// ToString provides a string representation of the given value.
// If the value is empty, it returns an empty string (""); for nil, it returns "nil"
func ToString(value interface{}) string {
	v := reflect.ValueOf(value)
	if stringer, ok := value.(fmt.Stringer); ok {
		return stringer.String()
	}
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return "nil"
		}
		return ToString(v.Elem().Interface())
	}
	if v.Kind() != reflect.Struct {
		return fmt.Sprintf("%v", value)
	}
	t := v.Type()
	var fields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			fieldVal := v.Field(i)
			var fieldString string
			if stringer, ok := fieldVal.Interface().(fmt.Stringer); ok {
				fieldString = fmt.Sprintf("%s: %s", field.Name, stringer.String())
			} else {
				switch fieldVal.Kind() {
				case reflect.Ptr, reflect.Interface:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else {
						fieldString = fmt.Sprintf("%s: %s", field.Name, ToString(fieldVal.Elem().Interface()))
					}
				case reflect.Slice:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
					}
				case reflect.Struct:
					fieldString = fmt.Sprintf("%s: %s", field.Name, ToString(fieldVal.Interface()))
				case reflect.Chan, reflect.Func:
					fieldString = fmt.Sprintf("%s: %T", field.Name, fieldVal.Interface())
				case reflect.String:
					if fieldVal.Len() == 0 {
						fieldString = fmt.Sprintf("%s: \"\"", field.Name)
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
					}
				case reflect.Map:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else {
						mapKeys := fieldVal.MapKeys()
						var mapStrings []string
						for _, key := range mapKeys {
							keyItem := ToString(key.Interface())
							valItem := ToString(fieldVal.MapIndex(key).Interface())
							mapStrings = append(mapStrings, fmt.Sprintf("%s: %s", keyItem, valItem))
						}
						fieldString = fmt.Sprintf("%s: {%s}", field.Name, strings.Join(mapStrings, ","))
					}
				default:
					fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
				}
			}
			fields = append(fields, fieldString)
		} else {
			fields = append(fields, fmt.Sprintf("%s: unexported", field.Name))
		}
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
}

func GetOIDCSecretName() string {
	oidcSecretName, ok := os.LookupEnv(EnvPxBackupOIDCSecretName)
	if !ok || oidcSecretName == "" {
		oidcSecretName = DefaultOIDCSecretName
	}
	return oidcSecretName
}

// GetPxBackupNamespace retrieves the namespace where PxBackupServiceName exists
func GetPxBackupNamespace() (string, error) {
	services, err := core.Instance().ListServices("", metav1.ListOptions{})
	if err != nil {
		return "", ProcessError(err)
	}
	for _, svc := range services.Items {
		if svc.Name == PxBackupServiceName {
			return svc.Namespace, nil
		}
	}
	err = fmt.Errorf("cannot find Px-Backup service [%s] from the list of services", PxBackupServiceName)
	return "", ProcessError(err)
}

func BuildFQDN(serviceName string, namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace)
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
