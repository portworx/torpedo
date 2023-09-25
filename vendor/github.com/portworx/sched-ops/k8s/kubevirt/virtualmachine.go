package kubevirt

import (
	"fmt"
	"time"

	"github.com/portworx/sched-ops/task"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

// VirtualMachineOps is an interface to perform kubevirt virtualMachine operations
type VirtualMachineOps interface {
	CreateVirtualMachine(*kubevirtv1.VirtualMachine) (*kubevirtv1.VirtualMachine, error)
	ListVirtualMachines(namespace string) (*kubevirtv1.VirtualMachineList, error)
	ValidateVirtualMachineRunning(string, string, time.Duration, time.Duration) error
	DeleteVirtualMachine(string, string) error
	GetVirtualMachine(string, string) (*kubevirtv1.VirtualMachine, error)
	StartVirtualMachine(*kubevirtv1.VirtualMachine) error
	StopVirtualMachine(*kubevirtv1.VirtualMachine) error
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

func (c *Client) ValidateVirtualMachineRunning(name, namespace string, timeout, retryInterval time.Duration) error {
	if err := c.initClient(); err != nil {
		return err
	}
	vm, err := c.GetVirtualMachine(name, namespace)
	if err != nil {
		return fmt.Errorf("failed to get Virtual Machine")
	}

	if !*vm.Spec.Running {
		if err = instance.StartVirtualMachine(vm); err != nil {
			return fmt.Errorf("Failed to start VirtualMachine %v", err)
		}
	}

	t := func() (interface{}, bool, error) {

		vm, err = c.GetVirtualMachine(name, namespace)
		if err != nil {
			return "", false, fmt.Errorf("failed to get Virtual Machine")
		}

		if vm.Status.PrintableStatus == kubevirtv1.VirtualMachineStatusRunning {
			return "", false, nil
		}
		return "", true, fmt.Errorf("Virtual Machine not in running state: %v", vm.Status.PrintableStatus)

	}
	if _, err := task.DoRetryWithTimeout(t, timeout, retryInterval); err != nil {
		return err
	}
	return nil

}

func (c *Client) StartVirtualMachine(vm *kubevirtv1.VirtualMachine) error {

	if err := c.initClient(); err != nil {
		return err
	}

	return c.kubevirt.VirtualMachine(vm.GetNamespace()).Start(vm.GetName(), &kubevirtv1.StartOptions{})
}

func (c *Client) StopVirtualMachine(vm *kubevirtv1.VirtualMachine) error {
	if err := c.initClient(); err != nil {
		return err
	}

	return c.kubevirt.VirtualMachine(vm.GetNamespace()).Stop(vm.GetName(), &kubevirtv1.StopOptions{})
}
