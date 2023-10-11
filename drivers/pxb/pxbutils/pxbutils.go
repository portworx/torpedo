package pxbutils

import (
	"fmt"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/pkg/log"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

const (
	// GlobalTorpedoWorkDirectory represents the working directory in the Torpedo container
	GlobalTorpedoWorkDirectory = "/go/src/github.com/portworx/"
	// GlobalPxBackupServiceName is the name of the Kubernetes service associated with Px-Backup
	GlobalPxBackupServiceName = "px-backup"
)

func InspectResponse(response *http.Response) ([]byte, error) {
	if response == nil {
		err := fmt.Errorf("response is nil")
		return nil, ProcessError(err)
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Errorf("failed to close response body. Err: [%v]", ProcessError(err))
		}
	}()
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, ProcessError(err)
	}
	statusCode, requestURL := response.StatusCode, response.Request.URL
	switch {
	case statusCode >= 200 && statusCode < 300:
		return respBody, nil
	case statusCode >= 400 && statusCode < 500:
		err = fmt.Errorf("client-side error for URL [%s]. Status code: [%d], Response Body: [%s]", requestURL, statusCode, respBody)
		return nil, ProcessError(err)
	case statusCode >= 500:
		err = fmt.Errorf("server-side error for URL [%s]. Status code: [%d], Response Body: [%s]", requestURL, statusCode, respBody)
		return nil, ProcessError(err)
	default:
		err = fmt.Errorf("unexpected status code %d for URL %s. Response Body: %s", statusCode, requestURL, respBody)
		return nil, ProcessError(err)
	}
}

// ProcessError formats the error message with caller information and an optional debug message
func ProcessError(err error, debugMessage ...string) error {
	if err == nil {
		return nil
	}
	_, file, line, _ := runtime.Caller(1)
	file = strings.TrimPrefix(file, GlobalTorpedoWorkDirectory)
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

// GetPxBackupNamespace retrieves the namespace where GlobalPxBackupServiceName exists
func GetPxBackupNamespace() (string, error) {
	services, err := core.Instance().ListServices("", metav1.ListOptions{})
	if err != nil {
		return "", ProcessError(err)
	}
	for _, svc := range services.Items {
		if svc.Name == GlobalPxBackupServiceName {
			return svc.Namespace, nil
		}
	}
	err = fmt.Errorf("cannot find Px-Backup service [%s] from the list of services", GlobalPxBackupServiceName)
	return "", ProcessError(err)
}
