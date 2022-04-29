package pds

import (
	"context"
	"time"
)

type ControlPlane struct {
	*cluster
}

func NewControlPlane(kubeconfig string) (*ControlPlane, error) {
	cluster, err := newCluster(kubeconfig)
	if err != nil {
		return nil, err
	}
	return &ControlPlane{cluster}, nil
}

func (cp *ControlPlane) ComponentLogsSince(ctx context.Context, since time.Time) (string, error) {
	components := []namespacedName{
		{pdsSystemNamespace, "api-server"},
		{pdsSystemNamespace, "api-worker"},
		{pdsSystemNamespace, "faktory"},
	}
	return cp.getLogsForComponents(ctx, components, since)
}
