package lib

import (
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	envControlPlaneURL    = "PDS_CONTROL_PLANE_URL"
	envPDSTestAccountName = "PDS_TEST_ACCOUNT_NAME"
	envTargetKubeconfig   = "PDS_TARGET_KUBECONFIG"
	envUsername           = "PDS_USERNAME"
	envPassword           = "PDS_PASSWORD"
	envPDSClientSecret    = "PDS_CLIENT_SECRET"
	envPDSClientId        = "PDS_CLIENT_ID"
	envPDSISSUERURL       = "PDS_ISSUER_URL"
	pdsSystemNamespace    = "pds-system"
	envClusterType        = "PDS_TARGET_CLUSTER_TYPE"
)

// Environment lhasha
type Environment struct {
	PDS_CONTROL_PLANE_URL   string
	PDS_TEST_ACCOUNT_NAME   string
	PDS_TARGET_KUBECONFIG   string
	PDS_USERNAME            string
	PDS_PASSWORD            string
	PDS_ISSUER_URL          string
	PDS_CLIENT_ID           string
	PDS_CLIENT_SECRET       string
	PDS_TARGET_CLUSTER_TYPE string
}

// MustHaveEnvVariables ljsas
func MustHaveEnvVariables() Environment {
	return Environment{
		PDS_CONTROL_PLANE_URL:   mustGetEnvVariable(envControlPlaneURL),
		PDS_TEST_ACCOUNT_NAME:   mustGetEnvVariable(envPDSTestAccountName),
		PDS_TARGET_KUBECONFIG:   mustGetEnvVariable(envTargetKubeconfig),
		PDS_USERNAME:            mustGetEnvVariable(envUsername),
		PDS_PASSWORD:            mustGetEnvVariable(envPassword),
		PDS_ISSUER_URL:          mustGetEnvVariable(envPDSISSUERURL),
		PDS_CLIENT_ID:           mustGetEnvVariable(envPDSClientId),
		PDS_CLIENT_SECRET:       mustGetEnvVariable(envPDSClientSecret),
		PDS_TARGET_CLUSTER_TYPE: mustGetEnvVariable(envClusterType),
	}
}

// mustGetEnvVariable jasljla
func mustGetEnvVariable(key string) string {
	value, isExist := os.LookupEnv(key)
	if !isExist {
		log.Panicf("Key: %v doesn't exist, Kindly visit -  https://github.com/portworx/pds-functional-test/blob/main/README.md#setting-up-the-environment-variable ", key)

	}
	return value
}
