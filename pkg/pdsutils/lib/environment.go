package lib

import (
	"os"

	"github.com/sirupsen/logrus"
)

const (
	envControlPlaneURL        = "CONTROL_PLANE_URL"
	envControlPlaneKubeconfig = "CONTROL_PLANE_KUBECONFIG"
	envTargetKubeconfig       = "TARGET_KUBECONFIG"
)

// Environment lhasha
type Environment struct {
	ControlPlaneURL        string
	ControlPlaneKubeconfig string
	TargetKubeconfig       string
}

// MustHaveEnvVariables ljsas
func MustHaveEnvVariables() Environment {
	return Environment{
		ControlPlaneURL:        mustGetEnvVariable(envControlPlaneURL),
		ControlPlaneKubeconfig: mustGetEnvVariable(envControlPlaneKubeconfig),
		TargetKubeconfig:       mustGetEnvVariable(envTargetKubeconfig),
	}
}

// mustGetEnvVariable jasljla
func mustGetEnvVariable(key string) string {
	value, isExist := os.LookupEnv(key)
	if !isExist {
		logrus.Errorf("Key: %v doesn't exist", key)
	}
	return value
}
