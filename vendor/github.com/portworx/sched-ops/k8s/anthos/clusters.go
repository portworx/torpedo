package anthos

import (
	"context"

	v1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/deprecated/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

// ClusterOps is an interface to perform k8s cluster operations
type ClusterOps interface {
	// ListCluster lists all kubernetes clusters
	ListCluster(ctx context.Context, options metav1.ListOptions) (*v1alpha1.ClusterList, error)
	// GetCluster returns a cluster for the given name
	GetCluster(ctx context.Context, name string, options metav1.GetOptions) (*v1alpha1.Cluster, error)
	// GetClusterStatus return the given cluster status
	GetClusterStatus(ctx context.Context, name string, options metav1.GetOptions) (*v1alpha1.ClusterStatus, error)
}

// ListCluster lists all kubernetes clusters
func (c *Client) ListCluster(ctx context.Context, options metav1.ListOptions) (*v1alpha1.ClusterList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	clusterList := &v1alpha1.ClusterList{}
	err := c.RESTClient().Get().
		Resource("clusters").
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(clusterList)

	return clusterList, err
}

// GetCluster returns a cluster for the given name
func (c *Client) GetCluster(ctx context.Context, name string, options metav1.GetOptions) (*v1alpha1.Cluster, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	cluster := &v1alpha1.Cluster{}
	err := c.RESTClient().Get().
		Resource("clusters").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).Do(ctx).Into(cluster)

	return cluster, err
}

// GetClusterStatus return the given cluster status
func (c *Client) GetClusterStatus(ctx context.Context, name string, options metav1.GetOptions) (*v1alpha1.ClusterStatus, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}
	cluster := &v1alpha1.Cluster{}
	err := c.RESTClient().Get().
		Resource("clusters").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).Do(ctx).Into(cluster)

	return &cluster.Status, err
}
