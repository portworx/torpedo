package pds

import (
	"context"
	"time"
)

// ControlPlane wraps a PDS control plane.
type ControlPlane struct {
	*cluster
}

// NewControlPlane creates a ControlPlane instance using the specified kubeconfig path.
// Fails if a kubernetes go-client cannot be configured based on the kubeconfig.
func NewControlPlane(kubeconfig string) (*ControlPlane, error) {
	cluster, err := newCluster(kubeconfig)
	if err != nil {
		return nil, err
	}
	return &ControlPlane{cluster}, nil
}

// ComponentLogsSince extracts the logs of all relevant PDS components, beginning at the specified time.
func (cp *ControlPlane) ComponentLogsSince(ctx context.Context, since time.Time) (string, error) {
	components := []namespacedName{
		{pdsSystemNamespace, "api-server"},
		{pdsSystemNamespace, "api-worker"},
		{pdsSystemNamespace, "faktory"},
	}
	return cp.getLogsForComponents(ctx, components, since)
}
