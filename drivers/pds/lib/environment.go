package lib

import (
	"os"
	"strconv"

	"github.com/onsi/gomega"
)

const (
	envPDSTestAccountName = "TEST_ACCOUNT_NAME"
	envTargetKubeconfig   = "TARGET_KUBECONFIG"
	envUsername           = "PDS_USERNAME"
	envPassword           = "PDS_PASSWORD"
	envPDSClientSecret    = "PDS_CLIENT_SECRET"
	envPDSClientID        = "PDS_CLIENT_ID"
	envPDSISSUERURL       = "PDS_ISSUER_URL"
)

// GetAndExpectStringEnvVar parses a string from env variable.
func GetAndExpectStringEnvVar(varName string) string {
	varValue := os.Getenv(varName)
	return varValue
}

// GetAndExpectIntEnvVar parses an int from env variable.
func GetAndExpectIntEnvVar(varName string) (int, error) {
	varValue := GetAndExpectStringEnvVar(varName)
	varIntValue, err := strconv.Atoi(varValue)
	return varIntValue, err
}

// GetAndExpectBoolEnvVar parses a boolean from env variable.
func GetAndExpectBoolEnvVar(varName string) bool {
	varValue := GetAndExpectStringEnvVar(varName)
	varBoolValue, err := strconv.ParseBool(varValue)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Error Parsing "+varName)
	return varBoolValue
}
