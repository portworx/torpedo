package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

const (
	// GlobalTorpedoWorkDirectory is where the Torpedo is located in the Kubernetes pod
	GlobalTorpedoWorkDirectory = "/go/src/github.com/portworx/"
)

type TestMaintainer string

// List of Px-Backup test maintainers
const (
	Ak             TestMaintainer = "ak-px"
	Apimpalgaonkar                = "apimpalgaonkar"
	KPhalgun                      = "kphalgun-px"
	Kshithijiyer                  = "kshithijiyer-px"
	Mkoppal                       = "mkoppal-px"
	Sagrawal                      = "sagrawal-px"
	Skonda                        = "skonda-px"
	Sn                            = "sn-px"
	Tthurlapati                   = "tthurlapati-px"
	Vpinisetti                    = "vpinisetti-px"
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
	if s == nil {
		return "nil"
	}
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

// CreateFile creates a new file at the specified file path
func CreateFile(filePath string) (err error) {
	dirPath := filepath.Dir(filePath)
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		debugStruct := struct {
			DirPath string
		}{
			DirPath: dirPath,
		}
		return ProcessError(err, StructToString(debugStruct))
	}
	file, err := os.Create(filePath)
	if err != nil {
		debugStruct := struct {
			FilePath string
		}{
			FilePath: filePath,
		}
		return ProcessError(err, StructToString(debugStruct))
	}
	defer func() {
		if file != nil {
			err = file.Close()
			err = ProcessError(err)
		}
	}()
	return nil
}
