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

type environment struct {
	ControlPlaneURL, ControlPlaneKubeconfig, TargetKubeconfig string
}

func MustHaveEnvVariables() environment {
	return environment{
		ControlPlaneURL:        mustGetEnvVariable(envControlPlaneURL),
		ControlPlaneKubeconfig: mustGetEnvVariable(envControlPlaneKubeconfig),
		TargetKubeconfig:       mustGetEnvVariable(envTargetKubeconfig),
	}
}

func mustGetEnvVariable(key string) string {
	value, isExist := os.LookupEnv(key)
	if !isExist {
		logrus.Errorf("Key: %v doesn't exist", key)
	}
	return value
}
