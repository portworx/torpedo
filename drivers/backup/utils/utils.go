package utils

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

type TestMaintainer string

const (
	Sagrawal TestMaintainer = "sagrawal-px"
)

const (
	// GlobalTorpedoWorkDirectory is where the Torpedo is located
	GlobalTorpedoWorkDirectory = "/go/src/github.com/portworx/"
)

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
	processedError := fmt.Errorf("%s\n  at %s <-> %s", err.Error(), callerInfo, debugInfo)
	return processedError
}

// StructToString returns the string representation of the given struct
func StructToString(s interface{}) string {
	v := reflect.ValueOf(s)
	if stringer, ok := s.(fmt.Stringer); ok {
		return stringer.String()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Sprintf("%v", s)
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
				case reflect.Ptr:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else if fieldVal.Type().Elem().Kind() == reflect.Struct {
						fieldString = fmt.Sprintf("%s: %s", field.Name, StructToString(fieldVal.Elem().Interface()))
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Elem())
					}
				case reflect.Slice:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
					}
				case reflect.Map:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
					}
				case reflect.Struct:
					fieldString = fmt.Sprintf("%s: %s", field.Name, StructToString(fieldVal.Interface()))
				case reflect.String:
					if fieldVal.Len() == 0 {
						fieldString = fmt.Sprintf("%s: \"\"", field.Name)
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
					}
				default:
					fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
				}
			}
			fields = append(fields, fieldString)
		}
	}
	return fmt.Sprintf("%s: {%s}", t.Name(), strings.Join(fields, ", "))
}
