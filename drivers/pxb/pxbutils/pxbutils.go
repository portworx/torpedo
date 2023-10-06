package pxbutils

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
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
	return fmt.Errorf("%s\n  at %s <-> %s", err.Error(), callerInfo, debugInfo)
}

// ToString provides a string representation of the given value.
// If the value is empty, it returns an empty string (""); for nil, it returns "nil"
func ToString(value interface{}) string {
	v := reflect.ValueOf(value)
	if stringer, ok := value.(fmt.Stringer); ok {
		return stringer.String()
	}
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
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
				case reflect.Ptr:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else if fieldVal.Type().Elem().Kind() == reflect.Struct {
						fieldString = fmt.Sprintf("%s: %s", field.Name, ToString(fieldVal.Elem().Interface()))
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Elem())
					}
				case reflect.Slice:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else {
						fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
					}
				case reflect.Array:
					fieldString = fmt.Sprintf("%s: %v", field.Name, fieldVal.Interface())
				case reflect.Map:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else {
						mapKeys := fieldVal.MapKeys()
						var mapStrings []string
						for _, key := range mapKeys {
							value := fieldVal.MapIndex(key)
							mapStrings = append(mapStrings, fmt.Sprintf("%s: %s", ToString(key.Interface()), ToString(value.Interface())))
						}
						fieldString = fmt.Sprintf("%s: {%s}", field.Name, strings.Join(mapStrings, ", "))
					}
				case reflect.Interface:
					if fieldVal.IsNil() {
						fieldString = fmt.Sprintf("%s: nil", field.Name)
					} else {
						fieldString = fmt.Sprintf("%s: %s", field.Name, ToString(fieldVal.Elem().Interface()))
					}
				case reflect.Struct:
					fieldString = fmt.Sprintf("%s: %s", field.Name, ToString(fieldVal.Interface()))
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
