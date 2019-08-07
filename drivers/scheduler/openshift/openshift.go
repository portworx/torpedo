package openshift

import (
	k8s_ops "github.com/portworx/sched-ops/k8s"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	kube "github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
)

const (
	// SchedName is the name of the kubernetes scheduler driver implementation
	SchedName = "openshift"
	// SystemdSchedServiceName is the name of the system service resposible for scheduling
	SystemdSchedServiceName = "atomic-openshift-node"
)

type openshift struct {
	kube.K8s
}

func (k *openshift) StopSchedOnNode(n node.Node) error {
	driver, _ := node.Get(k.K8s.NodeDriverName)
	systemOpts := node.SystemctlOpts{
		ConnectionOpts: node.ConnectionOpts{
			Timeout:         kube.FindFilesOnWorkerTimeout,
			TimeBeforeRetry: kube.DefaultRetryInterval,
		},
		Action: "stop",
	}
	err := driver.Systemctl(n, SystemdSchedServiceName, systemOpts)
	if err != nil {
		return &scheduler.ErrFailedToStopSchedOnNode{
			Node:          n,
			SystemService: SystemdSchedServiceName,
			Cause:         err.Error(),
		}
	}
	return nil
}

func (k *openshift) StartSchedOnNode(n node.Node) error {
	driver, _ := node.Get(k.K8s.NodeDriverName)
	systemOpts := node.SystemctlOpts{
		ConnectionOpts: node.ConnectionOpts{
			Timeout:         kube.DefaultTimeout,
			TimeBeforeRetry: kube.DefaultRetryInterval,
		},
		Action: "start",
	}
	err := driver.Systemctl(n, SystemdSchedServiceName, systemOpts)
	if err != nil {
		return &scheduler.ErrFailedToStartSchedOnNode{
			Node:          n,
			SystemService: SystemdSchedServiceName,
			Cause:         err.Error(),
		}
	}
	return nil
}

func (k *openshift) Schedule(instanceID string, options scheduler.ScheduleOptions) ([]*scheduler.Context, error) {
	var apps []*spec.AppSpec
	if len(options.AppKeys) > 0 {
		for _, key := range options.AppKeys {
			spec, err := k.SpecFactory.Get(key)
			if err != nil {
				return nil, err
			}
			apps = append(apps, spec)
		}
	} else {
		apps = k.SpecFactory.GetAll()
	}

	var contexts []*scheduler.Context
	for _, app := range apps {

		appNamespace := app.GetID(instanceID)

		// Update security context for namespace and user
		if err := k.updateSecurityContextConstraints(appNamespace); err != nil {
			return nil, err
		}

		specObjects, err := k.CreateSpecObjects(app, appNamespace, options.StorageProvisioner)
		if err != nil {
			return nil, err
		}

		ctx := &scheduler.Context{
			UID: instanceID,
			App: &spec.AppSpec{
				Key:      app.Key,
				SpecList: specObjects,
				Enabled:  app.Enabled,
			},
		}

		contexts = append(contexts, ctx)
	}

	return contexts, nil
}

func (k *openshift) updateSecurityContextConstraints(namespace string) error {
	// Get privileged context
	context, err := k8s_ops.Instance().GetSecurityContextConstraints("privileged")
	if err != nil {
		return err
	}

	// Add user and namespace to context
	context.Users = append(context.Users, "system:serviceaccount:"+namespace+":default")

	// Update context
	_, err = k8s_ops.Instance().UpdateSecurityContextConstraints(context)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	k := &openshift{}
	scheduler.Register(SchedName, k)
}
