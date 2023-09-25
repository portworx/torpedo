package kubevirt

import (
	"time"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

// VirtualMachineOps is an interface to perform kubevirt virtualMachine operations
type VirtualMachineOps interface {
	CreateVirtualMachine(*kubevirtv1.VirtualMachine) (*kubevirtv1.VirtualMachine, error)
	ListVirtualMachines(namespace string) (*kubevirtv1.VirtualMachineList, error)
	IsVirtualMachineRunning(string, string, time.Duration, time.Duration) error
	DeleteVirtualMachine(string, string) error
	GetVirtualMachine(string, string) (*kubevirtv1.VirtualMachine, error)
}

// GetStorageClasses returns all storageClasses that match given optional label selector
func (c *Client) ListVirtualMachines(namespace string) (*kubevirtv1.VirtualMachineList, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubevirt.VirtualMachine(namespace).List(&k8smetav1.ListOptions{})
}

func (c *Client) CreateVirtualMachine(vm *kubevirtv1.VirtualMachine) (*kubevirtv1.VirtualMachine, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubevirt.VirtualMachine(vm.GetNamespace()).Create(vm)
}

func (c *Client) GetVirtualMachine(name string, namespace string) (*kubevirtv1.VirtualMachine, error) {
	if err := c.initClient(); err != nil {
		return nil, err
	}

	return c.kubevirt.VirtualMachine(namespace).Get(name, &k8smetav1.GetOptions{})
}

func (c *Client) DeleteVirtualMachine(name, namespace string) error {
	if err := c.initClient(); err != nil {
		return err
	}

	return c.kubevirt.VirtualMachine(namespace).Delete(name, &k8smetav1.DeleteOptions{})
}

func (c *Client) IsVirtualMachineRunning(name, namespace string, timeout, retryInterval time.Duration) error {
	if err := c.initClient(); err != nil {
		return err
	}

	if vm, err := c.GetVirtualMachine(name, namespace); err == nil {
		if *vm.Spec.Running {
			return nil
		}

	}
	return nil

}
