package tests

import (
	"encoding/json"
	"fmt"
	"github.com/pborman/uuid"
	"github.com/portworx/torpedo/drivers/backup/backup_utils"
	"github.com/portworx/torpedo/pkg/log"
	"net/http"
)

type BackupAPIError struct {
	Error     error
	ErrorCode string
}

func (e *BackupAPIError) GetError() error {
	return e.Error
}

func (e *BackupAPIError) SetError(err error) *BackupAPIError {
	e.Error = err
	return e
}

func (e *BackupAPIError) GetErrorCode() string {
	return e.ErrorCode
}

func (e *BackupAPIError) SetErrorCode(errCode string) *BackupAPIError {
	e.ErrorCode = errCode
	return e
}

func NewBackupAPIError(err error, errCode string) *BackupAPIError {
	backupAPIError := &BackupAPIError{}
	backupAPIError.SetError(err)
	backupAPIError.SetErrorCode(errCode)
	return backupAPIError
}

func NewDefaultBackupAPIError() *BackupAPIError {
	return NewBackupAPIError(fmt.Errorf("internal server error"), uuid.New())
}

func (e *BackupAPIError) AsMap() map[string]string {
	return map[string]string{
		"error":      e.GetError().Error(),
		"error_code": e.GetErrorCode(),
	}
}

func RespondWithError(w http.ResponseWriter, code int, backupAPIError *BackupAPIError) {
	RespondWithJSON(w, code, backupAPIError.AsMap())
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("an error encountered while marshaling JSON: %s", backup_utils.ProcessError(err).Error())
		response, err = json.Marshal(NewDefaultBackupAPIError().AsMap())
		if err != nil {
			log.Errorf("an error encountered while writing response: %s", backup_utils.ProcessError(err).Error())
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write(response)
		if err != nil {
			log.Errorf("an error encountered while writing response: %s", backup_utils.ProcessError(err).Error())
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		log.Errorf("an error encountered while writing response: %s", backup_utils.ProcessError(err).Error())
	}
}
