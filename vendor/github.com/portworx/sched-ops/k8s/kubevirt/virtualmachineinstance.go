package kubevirt

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

// VirtualMachineInstanceOps is an interface to perform kubevirt operations on virtual machine instances
type VirtualMachineInstanceOps interface {
	// GetVirtualMachineInstance gets updated Virtual Machine Instance from client matching name and namespace
	GetVirtualMachineInstance(context.Context, string, string) (*kubevirtv1.VirtualMachineInstance, error)
	// UpdateVirtualMachineInstance updates existing Virtual Machine Instance
	UpdateVirtualMachineInstance(context.Context, *kubevirtv1.VirtualMachineInstance) (*kubevirtv1.VirtualMachineInstance, error)
}

// GetVirtualMachineInstance gets updated Virtual Machine Instance from client matching name and namespace
func (c *Client) GetVirtualMachineInstance(ctx context.Context, name string, namespace string) (*kubevirtv1.VirtualMachineInstance, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubevirt.VirtualMachineInstance(namespace).Get(ctx, name, &metav1.GetOptions{})
}

// UpdateVirtualMachineInstance updates existing Virtual Machine Instance
func (c *Client) UpdateVirtualMachineInstance(ctx context.Context, vmInstance *kubevirtv1.VirtualMachineInstance) (*kubevirtv1.VirtualMachineInstance, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	ns := vmInstance.Namespace
	if len(ns) == 0 {
		ns = corev1.NamespaceDefault
	}

	return c.kubevirt.VirtualMachineInstance(ns).Update(ctx, vmInstance)
}
