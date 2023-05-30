package dcos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	docker "github.com/docker/docker/client"
	marathon "github.com/gambol99/go-marathon"
	v1beta1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1beta1"
	snapv1 "github.com/kubernetes-incubator/external-storage/snapshot/pkg/apis/crd/v1"
	apapi "github.com/libopenstorage/autopilot-api/pkg/apis/autopilot/v1alpha1"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/portworx/torpedo/pkg/errors"
	"github.com/portworx/torpedo/pkg/log"
	"golang.org/x/net/context"

	corev1 "k8s.io/api/core/v1"
	storageapi "k8s.io/api/storage/v1"
)

const (
	// SchedName is the name of the dcos scheduler driver implementation
	SchedName      = "dcos"
	defaultTimeout = 5 * time.Minute
)

type Dcos struct {
	dockerClient   *docker.Client
	specFactory    *spec.Factory
	marathonClient *marathonOps
	NodeRegistry   *node.NodeRegistry

	// volume driver must be set after it has been initialized. It is not set during Init
	VolumeDriver volume.Driver
}

// DeepCopy deep copies the driver instance
func (d *Dcos) DeepCopy() scheduler.Driver {
	if d == nil {
		return nil
	}
	out := *d

	if d.NodeRegistry != nil {
		out.NodeRegistry = &node.NodeRegistry{
			Nodes: make(map[string]node.Node),
		}
		for _, node := range d.NodeRegistry.GetNodes() {
			out.NodeRegistry.AddNode(node)
		}
	}

	if d.specFactory != nil {
		specFactory := *d.specFactory
		out.specFactory = &specFactory
	}

	if d.dockerClient != nil {
		dockerClient := *d.dockerClient
		out.dockerClient = &dockerClient
	}

	if d.marathonClient != nil {
		marathonClient := *d.marathonClient
		out.marathonClient = &marathonClient
	}
	// ISSUE: this is not useful as client is built from env vars, which are always the same
	// ISSUE: mesos client alos has above problem, hence not deepcopied

	return &out
}

func (d *Dcos) Init(schedOpts scheduler.InitOptions) error {
	privateAgents, err := MesosClient().GetPrivateAgentNodes()
	if err != nil {
		return err
	}

	for _, n := range privateAgents {
		newNode := d.parseMesosNode(n)
		if err := d.IsNodeReady(newNode); err != nil {
			return err
		}
		if err := d.NodeRegistry.AddNode(newNode); err != nil {
			return err
		}
	}

	d.specFactory, err = spec.NewFactory(schedOpts.SpecDir, schedOpts.VolumeDriverName, d)
	if err != nil {
		return err
	}

	d.dockerClient, err = docker.NewEnvClient()
	if err != nil {
		return err
	}

	return nil
}

func (d *Dcos) parseMesosNode(n AgentNode) node.Node {
	return node.Node{
		Name:      n.ID,
		Addresses: []string{n.Hostname},
		Type:      node.TypeWorker,
	}
}

func (d *Dcos) String() string {
	return SchedName
}

// GetEvents dumps events from event storage
func (d *Dcos) GetEvents() map[string][]scheduler.Event {
	return nil
}

// ValidateAutopilotEvents validates events for PVCs injected by autopilot
func (d *Dcos) ValidateAutopilotEvents(ctx *scheduler.Context) error {
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ValidateAutopilotEvents()",
	}
}

// ValidateAutopilotRuleObjects validates autopilot rule objects
func (d *Dcos) ValidateAutopilotRuleObjects() error {
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ValidateAutopilotRuleObjects()",
	}
}

// GetSnapShotData retruns given snapshots
func (d *Dcos) GetSnapShotData(ctx *scheduler.Context, snapshotName, snapshotNameSpace string) (*snapv1.VolumeSnapshotData, error) {
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ValidateSnapShot()",
	}
}

// DeleteSnapshots  delete the snapshots
func (d *Dcos) DeleteSnapShot(ctx *scheduler.Context, snapshotName, snapshotNameSpace string) error {
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "DeleteSnapShot()",
	}
}

// GetSnapshotsInNameSpace get the snapshots list for the namespace
func (d *Dcos) GetSnapshotsInNameSpace(ctx *scheduler.Context, snapshotNameSpace string) (*snapv1.VolumeSnapshotList, error) {

	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "DeleteSnapShot()",
	}
}

func (d *Dcos) ParseSpecs(specDir, storageProvisioner string) ([]interface{}, error) {
	fileList := []string{}
	if err := filepath.Walk(specDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var specs []interface{}
	for _, fileName := range fileList {
		raw, err := ioutil.ReadFile(fileName)
		if err != nil {
			return nil, err
		}

		app := new(marathon.Application)
		if err := json.Unmarshal(raw, app); err != nil {
			return nil, err
		}

		specs = append(specs, app)
	}

	return specs, nil
}

func (d *Dcos) IsNodeReady(n node.Node) error {
	// TODO: Implement this method
	return nil
}

func (d *Dcos) GetNodesForApp(ctx *scheduler.Context) ([]node.Node, error) {
	var tasks []marathon.Task
	for _, spec := range ctx.App.SpecList {
		if obj, ok := spec.(*marathon.Application); ok {
			appTasks, err := d.marathonClient.GetApplicationTasks(obj.ID)
			if err != nil {
				return nil, &scheduler.ErrFailedToGetNodesForApp{
					App:   ctx.App,
					Cause: fmt.Sprintf("Failed to get tasks for application %v. %v", obj.ID, err),
				}
			}
			tasks = append(tasks, appTasks...)
		} else {
			log.Warnf("Invalid spec received for app %v in GetNodesForApp", ctx.App.Key)
		}
	}

	var result []node.Node
	nodeMap := d.NodeRegistry.GetNodesByName()

	for _, task := range tasks {
		n, ok := nodeMap[task.SlaveID]
		if !ok {
			return nil, &scheduler.ErrFailedToGetNodesForApp{
				App:   ctx.App,
				Cause: fmt.Sprintf("node [%v] not present in node map", task.SlaveID),
			}
		}

		if d.NodeRegistry.Contains(result, n) {
			continue
		}
		result = append(result, n)
	}

	return result, nil
}

func (d *Dcos) Schedule(instanceID string, options scheduler.ScheduleOptions) ([]*scheduler.Context, error) {
	var apps []*spec.AppSpec
	if len(options.AppKeys) > 0 {
		for _, key := range options.AppKeys {
			spec, err := d.specFactory.Get(key)
			if err != nil {
				return nil, err
			}
			apps = append(apps, spec)
		}
	} else {
		apps = d.specFactory.GetAll()
	}

	var contexts []*scheduler.Context
	for _, app := range apps {
		var specObjects []interface{}
		for _, spec := range app.SpecList {
			if application, ok := spec.(*marathon.Application); ok {
				if err := d.randomizeVolumeNames(application); err != nil {
					return nil, &scheduler.ErrFailedToScheduleApp{
						App:   app,
						Cause: err.Error(),
					}
				}
				obj, err := d.marathonClient.CreateApplication(application)
				if err != nil {
					return nil, &scheduler.ErrFailedToScheduleApp{
						App:   app,
						Cause: err.Error(),
					}
				}

				specObjects = append(specObjects, obj)
			} else {
				return nil, fmt.Errorf("Unsupported object received in app %v while scheduling", app.Key)
			}
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

// ScheduleWithCustomAppSpecs Schedules the application with custom app specs
func (d *Dcos) ScheduleWithCustomAppSpecs(apps []*spec.AppSpec, instanceID string, options scheduler.ScheduleOptions) ([]*scheduler.Context, error) {
	var contexts []*scheduler.Context
	for _, app := range apps {
		var specObjects []interface{}
		for _, spec := range app.SpecList {
			if application, ok := spec.(*marathon.Application); ok {
				if err := d.randomizeVolumeNames(application); err != nil {
					return nil, &scheduler.ErrFailedToScheduleApp{
						App:   app,
						Cause: err.Error(),
					}
				}
				obj, err := d.marathonClient.CreateApplication(application)
				if err != nil {
					return nil, &scheduler.ErrFailedToScheduleApp{
						App:   app,
						Cause: err.Error(),
					}
				}

				specObjects = append(specObjects, obj)
			} else {
				return nil, fmt.Errorf("Unsupported object received in app %v while scheduling", app.Key)
			}
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

// AddTasks adds tasks to an existing context
func (d *Dcos) AddTasks(ctx *scheduler.Context, options scheduler.ScheduleOptions) error {
	if ctx == nil {
		return fmt.Errorf("Context to add tasks to cannot be nil")
	}
	if len(options.AppKeys) == 0 {
		return fmt.Errorf("Need to specify list of applications to add to context")
	}

	var apps []*spec.AppSpec
	for _, key := range options.AppKeys {
		spec, err := d.specFactory.Get(key)
		if err != nil {
			return err
		}
		apps = append(apps, spec)
	}

	specObjects := ctx.App.SpecList
	for _, app := range apps {
		for _, spec := range app.SpecList {
			if application, ok := spec.(*marathon.Application); ok {
				if err := d.randomizeVolumeNames(application); err != nil {
					return &scheduler.ErrFailedToScheduleApp{
						App:   app,
						Cause: err.Error(),
					}
				}
				obj, err := d.marathonClient.CreateApplication(application)
				if err != nil {
					return &scheduler.ErrFailedToScheduleApp{
						App:   app,
						Cause: err.Error(),
					}
				}
				specObjects = append(specObjects, obj)
			} else {
				return fmt.Errorf("Unsupported object received in app %v while scheduling", app.Key)
			}

		}
	}
	ctx.App.SpecList = specObjects
	return nil
}

// ScheduleUninstall uninstalls tasks from an existing context
func (d *Dcos) ScheduleUninstall(ctx *scheduler.Context, options scheduler.ScheduleOptions) error {
	// TODO implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ScheduleUninstall()",
	}
}

// RemoveAppSpecsByName removes certain specs from list to avoid validation
func (d *Dcos) RemoveAppSpecsByName(ctx *scheduler.Context, removeSpecs []interface{}) error {
	// TODO implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "RemoveAppSpecsByName()",
	}
}

func (d *Dcos) UpdateTasksID(ctx *scheduler.Context, id string) error {
	// TODO: Add implementation
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "UpdateTasksID()",
	}
}

func (d *Dcos) randomizeVolumeNames(application *marathon.Application) error {
	params := *application.Container.Docker.Parameters
	for i := range params {
		p := &params[i]
		if p.Key == "volume" {
			p.Value = d.VolumeDriver.RandomizeVolumeName(p.Value)
		}
	}
	return nil
}

func (d *Dcos) WaitForRunning(ctx *scheduler.Context, timeout, retryInterval time.Duration) error {
	for _, spec := range ctx.App.SpecList {
		if obj, ok := spec.(*marathon.Application); ok {
			if err := d.marathonClient.WaitForApplicationStart(obj.ID); err != nil {
				return &scheduler.ErrFailedToValidateApp{
					App:   ctx.App,
					Cause: fmt.Sprintf("Failed to validate Application: %v. Err: %v", obj.ID, err),
				}
			}
			log.Infof("[%v] Validated application: %v", ctx.App.Key, obj.ID)
		} else {
			log.Warnf("Invalid spec received for app %v in WaitForRunning", ctx.App.Key)
		}
	}
	return nil
}

func (d *Dcos) Destroy(ctx *scheduler.Context, opts map[string]bool) error {
	for _, spec := range ctx.App.SpecList {
		if obj, ok := spec.(*marathon.Application); ok {
			if err := d.marathonClient.DeleteApplication(obj.ID); err != nil {
				return &scheduler.ErrFailedToDestroyApp{
					App:   ctx.App,
					Cause: fmt.Sprintf("Failed to destroy Application: %v. Err: %v", obj.ID, err),
				}
			}
			log.Infof("[%v] Destroyed application: %v", ctx.App.Key, obj.ID)
		} else {
			log.Warnf("Invalid spec received for app %v in Destroy", ctx.App.Key)
		}
	}

	if value, ok := opts[scheduler.OptionsWaitForResourceLeakCleanup]; ok && value {
		// TODO: wait until all the resources have been cleaned up properly
		if err := d.WaitForDestroy(ctx, defaultTimeout); err != nil {
			return err
		}
	} else if value, ok := opts[scheduler.OptionsWaitForDestroy]; ok && value {
		if err := d.WaitForDestroy(ctx, defaultTimeout); err != nil {
			return err
		}
	}

	return nil
}

func (d *Dcos) WaitForDestroy(ctx *scheduler.Context, timeout time.Duration) error {
	for _, spec := range ctx.App.SpecList {
		if obj, ok := spec.(*marathon.Application); ok {
			if err := d.marathonClient.WaitForApplicationTermination(obj.ID); err != nil {
				return &scheduler.ErrFailedToValidateAppDestroy{
					App:   ctx.App,
					Cause: fmt.Sprintf("Failed to destroy Application: %v. Err: %v", obj.ID, err),
				}
			}
			log.Infof("[%v] Validated destroy of Application: %v", ctx.App.Key, obj.ID)
		} else {
			log.Warnf("Invalid spec received for app %v in WaitForDestroy", ctx.App.Key)
		}
	}

	return nil
}

// SelectiveWaitForTermination waits for application pods to be terminated except on the nodes
// provided in the exclude list
func (d *Dcos) SelectiveWaitForTermination(ctx *scheduler.Context, timeout time.Duration, excludeList []node.Node) error {
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "SelectiveWaitForTermination",
	}
}

func (d *Dcos) DeleteTasks(ctx *scheduler.Context, opts *scheduler.DeleteTasksOptions) error {
	if opts != nil {
		log.Warnf("DCOS driver doesn't yet support delete task options")
	}

	for _, spec := range ctx.App.SpecList {
		if obj, ok := spec.(*marathon.Application); ok {
			if err := d.marathonClient.KillApplicationTasks(obj.ID); err != nil {
				return &scheduler.ErrFailedToDeleteTasks{
					App:   ctx.App,
					Cause: fmt.Sprintf("failed to delete tasks for application: %v. %v", obj.ID, err),
				}
			}
		} else {
			log.Warnf("Invalid spec received for app %v in DeleteTasks", ctx.App.Key)
		}
	}
	return nil
}

func (d *Dcos) GetVolumeParameters(ctx *scheduler.Context) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	populateParamsFunc := func(volName string, volParams map[string]string) error {
		result[volName] = volParams
		return nil
	}

	if err := d.volumeOperation(ctx, populateParamsFunc); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *Dcos) ValidateVolumes(ctx *scheduler.Context, timeout, retryInterval time.Duration,
	options *scheduler.VolumeOptions) error {
	inspectDockerVolumeFunc := func(volName string, _ map[string]string) error {
		t := func() (interface{}, bool, error) {
			out, err := d.dockerClient.VolumeInspect(context.Background(), volName)
			return out, true, err
		}

		if _, err := task.DoRetryWithTimeout(t, 2*time.Minute, 10*time.Second); err != nil {
			return &scheduler.ErrFailedToValidateStorage{
				App:   ctx.App,
				Cause: fmt.Sprintf("Failed to inspect docker volume: %v. Err: %v", volName, err),
			}
		}
		return nil
	}

	return d.volumeOperation(ctx, inspectDockerVolumeFunc)
}

func (d *Dcos) DeleteVolumes(ctx *scheduler.Context, options *scheduler.VolumeOptions) ([]*volume.Volume, error) {
	var vols []*volume.Volume

	deleteDockerVolumeFunc := func(volName string, _ map[string]string) error {
		vols = append(vols, &volume.Volume{Name: volName})
		t := func() (interface{}, bool, error) {
			return nil, true, d.dockerClient.VolumeRemove(context.Background(), volName, false)
		}

		if _, err := task.DoRetryWithTimeout(t, 2*time.Minute, 10*time.Second); err != nil {
			return &scheduler.ErrFailedToDestroyStorage{
				App:   ctx.App,
				Cause: fmt.Sprintf("Failed to remove docker volume: %v. Err: %v", volName, err),
			}
		}
		return nil
	}

	if err := d.volumeOperation(ctx, deleteDockerVolumeFunc); err != nil {
		return nil, err
	}
	return vols, nil
}

func (d *Dcos) GetVolumeDriverVolumeName(name string, namespace string) (string, error) {
	// TODO: Add implementation
	return "", &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetVolumeDriverVolumeName()",
	}
}

func (d *Dcos) GetVolumes(ctx *scheduler.Context) ([]*volume.Volume, error) {
	// TODO: Add implementation
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetVolumes()",
	}
}

func (d *Dcos) GetPureVolumes(ctx *scheduler.Context, pureVolType string) ([]*volume.Volume, error) {
	// TODO: Add implementation
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetPureVolumes()",
	}
}

func (d *Dcos) GetPodsForPVC(pvcname, namespace string) ([]corev1.Pod, error) {
	// TODO: Add implementation
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetPodsForPVC()",
	}
}

// GetPodLog returns logs for all the pods in the specified context
func (d *Dcos) GetPodLog(ctx *scheduler.Context, sinceSeconds int64, containerName string) (map[string]string, error) {
	// TODO: Add implementation
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetPodLog()",
	}
}

func (d *Dcos) ResizeVolume(cxt *scheduler.Context, configMap string) ([]*volume.Volume, error) {
	// TODO implement this method
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ResizeVolume()",
	}
}

func (d *Dcos) GetSnapshots(ctx *scheduler.Context) ([]*volume.Snapshot, error) {
	// TODO: Add implementation
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetSnapshots()",
	}
}

func (d *Dcos) volumeOperation(ctx *scheduler.Context, f func(string, map[string]string) error) error {
	// DC/OS does not have volume objects like Kubernetes. We get the volume information from
	// the app spec and get the options parsed from the respective volume driver

	for _, spec := range ctx.App.SpecList {
		if obj, ok := spec.(*marathon.Application); ok {
			// TODO: This handles only docker volumes. Implement for UCR/mesos containers
			params := *obj.Container.Docker.Parameters
			for _, p := range params {
				if p.Key == "volume" {
					volName, volParams, err := d.VolumeDriver.ExtractVolumeInfo(p.Value)
					if err != nil {
						return &scheduler.ErrFailedToGetVolumeParameters{
							App:   ctx.App,
							Cause: fmt.Sprintf("Failed to extract volume info: %v. Err: %v", p.Value, err),
						}
					}
					if err := f(volName, volParams); err != nil {
						return err
					}
				}
			}
		} else {
			log.Warnf("Invalid spec received for app %v", ctx.App.Key)
		}
	}

	return nil
}

func (d *Dcos) SetConfig(configPath string) error {
	// TODO: Implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "SetConfig()",
	}
}

func (d *Dcos) Describe(ctx *scheduler.Context) (string, error) {
	// TODO: Implement this method
	return "", &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "Describe()",
	}
}

func (d *Dcos) ScaleApplication(ctx *scheduler.Context, scaleFactorMap map[string]int32) error {
	// TODO implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ScaleApplication()",
	}
}

func (d *Dcos) GetScaleFactorMap(ctx *scheduler.Context) (map[string]int32, error) {
	// TODO implement this method
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetScaleFactorMap()",
	}
}

func (d *Dcos) StopSchedOnNode(node node.Node) error {
	// TODO implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "StopSchedOnNode()",
	}
}

func (d *Dcos) StartSchedOnNode(node node.Node) error {
	// TODO implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "StartSchedOnNode()",
	}
}
func (d *Dcos) RescanSpecs(specDir, storageDriver string) error {
	// TODO implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "RescanSpecs()",
	}
}

func (d *Dcos) PrepareNodeToDecommission(n node.Node, provisioner string) error {
	// TODO implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "PrepareNodeToDecommission()",
	}
}

func (d *Dcos) EnableSchedulingOnNode(n node.Node) error {
	// TODO implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "EnableSchedulingOnNode()",
	}
}

func (d *Dcos) DisableSchedulingOnNode(n node.Node) error {
	// TODO implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "DisableSchedulingOnNode()",
	}
}

func (d *Dcos) RefreshNodeRegistry() error {
	// TODO implement this method
	return nil
}

func (d *Dcos) IsScalable(spec interface{}) bool {
	// TODO implement this method
	return false
}

func (d *Dcos) ValidateVolumeSnapshotRestore(ctx *scheduler.Context, timeStart time.Time) error {
	return fmt.Errorf("not implemenented")
}

func (d *Dcos) GetTokenFromConfigMap(string) (string, error) {
	// TODO implement this method
	return "", &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetTokenFromConfigMap()",
	}
}

func (d *Dcos) AddLabelOnNode(n node.Node, lKey string, lValue string) error {
	// TODO implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "AddLabelOnNode()",
	}
}

func (d *Dcos) RemoveLabelOnNode(n node.Node, lKey string) error {
	// TODO implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "RemoveLabelOnNode()",
	}
}

func (d *Dcos) IsAutopilotEnabledForVolume(*volume.Volume) bool {
	// TODO implement this method
	return false
}

func (d *Dcos) GetSpecAppEnvVar(ctx *scheduler.Context, key string) string {
	// TODO implement this method
	return ""
}

func (d *Dcos) SaveSchedulerLogsToFile(n node.Node, location string) error {
	// TODO implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "SaveSchedulerLogsToFile()",
	}
}

// GetWorkloadSizeFromAppSpec gets workload size from an application spec
func (d *Dcos) GetWorkloadSizeFromAppSpec(ctx *scheduler.Context) (uint64, error) {
	// TODO: not implemented
	return 0, nil
}

func (d *Dcos) GetAutopilotNamespace() (string, error) {
	// TODO implement this method
	return "", &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetAutopilotNamespace()",
	}
}

// GetIOBandwidth returns the IO bandwidth for the given pod name and namespace
func (d *Dcos) GetIOBandwidth(string, string) (int, error) {
	// TODO implement this method
	return 0, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetIOBandwidth()",
	}
}

func (d *Dcos) CreateAutopilotRule(apRule apapi.AutopilotRule) (*apapi.AutopilotRule, error) {
	// TODO implement this method
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "CreateAutopilotRule()",
	}
}

func (d *Dcos) GetAutopilotRule(name string) (*apapi.AutopilotRule, error) {
	// TODO implement this method
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetAutopilotRule()",
	}
}

func (d *Dcos) UpdateAutopilotRule(*apapi.AutopilotRule) (*apapi.AutopilotRule, error) {
	// TODO implement this method
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "UpdateAutopilotRule()",
	}
}

func (d *Dcos) ListAutopilotRules() (*apapi.AutopilotRuleList, error) {
	// TODO implement this method
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ListAutopilotRules()",
	}
}

func (d *Dcos) DeleteAutopilotRule(name string) error {
	// TODO implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "DeleteAutopilotRule()",
	}
}

func (d *Dcos) GetActionApproval(namespace, name string) (*apapi.ActionApproval, error) {
	// TODO implement this method
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetActionApproval()",
	}
}

func (d *Dcos) UpdateActionApproval(namespace string, actionApproval *apapi.ActionApproval) (*apapi.ActionApproval, error) {
	// TODO implement this method
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "UpdateActionApproval()",
	}
}

func (d *Dcos) DeleteActionApproval(namespace, name string) error {
	// TODO implement this method
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "DeleteActionApproval()",
	}
}

func (d *Dcos) ListActionApprovals(namespace string) (*apapi.ActionApprovalList, error) {
	// TODO implement this method
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ListActionApprovals()",
	}
}

func (d *Dcos) UpgradeScheduler(version string) error {
	// TODO: Add implementation
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "UpgradeScheduler()",
	}
}

func (d *Dcos) CreateSecret(namespace, name, dataField, secretDataString string) error {
	// TODO: Add implementation
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "CreateSecret()",
	}
}

func (d *Dcos) GetSecretData(namespace, name, dataField string) (string, error) {
	// TODO: Add implementation
	return "", &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetSecret()",
	}
}

func (d *Dcos) DeleteSecret(namespace, name string) error {
	// TODO: Add implementation
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "DeleteSecret()",
	}
}

func (d *Dcos) ParseCharts(chartDir string) (*scheduler.HelmRepo, error) {
	// TODO implement this method
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ParseCharts()",
	}
}

func (d *Dcos) RecycleNode(n node.Node) error {
	//Recycle is not supported
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "RecycleNode()",
	}
}

func (d *Dcos) ValidateTopologyLabel(ctx *scheduler.Context) error {
	//ValidateTopologyLabel is not supported
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ValidateTopologyLabel()",
	}
}

func (d *Dcos) CreateCsiSnapshotClass(snapClassName string, deleionPolicy string) (*v1beta1.VolumeSnapshotClass, error) {
	//CreateCsiSnapshotClass is not supported
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "CreateCsiSnapshotClass()",
	}
}

func (d *Dcos) CreateCsiSnapshot(name string, namespace string, class string, pvc string) (*v1beta1.VolumeSnapshot, error) {
	//CreateCsiSanpshot is not supported
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "CreateCsiSanpshot()",
	}
}

func (d *Dcos) CreateCsiSnapsForVolumes(ctx *scheduler.Context, snapClass string) (map[string]*v1beta1.VolumeSnapshot, error) {
	//CreateCsiSnapsForVolumes is not supported
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "CreateCsiSnapsForVolumes()",
	}
}

func (d *Dcos) CSICloneTest(ctx *scheduler.Context, request scheduler.CSICloneRequest) error {
	//CSICloneTest is not supported for DCOS
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "CSICloneTest()",
	}
}

func (d *Dcos) CSISnapshotTest(ctx *scheduler.Context, request scheduler.CSISnapshotRequest) error {
	//CSISnapshotTest is not supported for DCOS
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "CSISnapshotTest()",
	}
}

func (d *Dcos) CSISnapshotAndRestoreMany(ctx *scheduler.Context, request scheduler.CSISnapshotRequest) error {
	//CSISnapshotAndRestoreMany is not supported for DCOS
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "CSISnapshotAndRestoreMany()",
	}
}

func (d *Dcos) GetCsiSnapshots(namespace string, pvcName string) ([]*v1beta1.VolumeSnapshot, error) {
	// GetCsiSnapshots is not supported
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetCsiSnapshots()",
	}
}

func (d *Dcos) ValidateCsiSnapshots(ctx *scheduler.Context, volSnapMa map[string]*v1beta1.VolumeSnapshot) error {
	// ValidateCsiSnapshots is not supported
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "ValidateCsiSnapshots()",
	}
}

func (d *Dcos) RestoreCsiSnapAndValidate(ctx *scheduler.Context, scList map[string]*storageapi.StorageClass) (map[string]corev1.PersistentVolumeClaim, error) {
	// RestoreCsiSnapAndValidate is not supported
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "RestoreCsiSnapAndValidate()",
	}

}

func (d *Dcos) DeleteCsiSnapsForVolumes(ctx *scheduler.Context, retainCount int) error {
	// DeleteCsiSnapsForVolumes is not supported
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "DeleteCsiSnapsForVolumes()",
	}

}

func (d *Dcos) DeleteCsiSnapshot(ctx *scheduler.Context, snapshotName string, snapshotNameSpace string) error {
	// DeleteCsiSnapshot is not supported
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "DeleteCsiSnapshot()",
	}

}

// GetAllSnapshotClasses returns the list of all volume snapshot classes present in the cluster
func (d *Dcos) GetAllSnapshotClasses() (*v1beta1.VolumeSnapshotClassList, error) {
	// GetAllSnapshotClasses is not supported
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetAllSnapshotClasses()",
	}
}

func (d *Dcos) GetPodsRestartCount(namespace string, label map[string]string) (map[*corev1.Pod]int32, error) {
	// GetPodsRestartCount is not supported
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetPodsRestartCoun()",
	}
}

func (d *Dcos) AddNamespaceLabel(namespace string, labelMap map[string]string) error {
	// AddNamespaceLabel is not supported
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "AddNamespaceLabel()",
	}
}

func (d *Dcos) RemoveNamespaceLabel(namespace string, labelMap map[string]string) error {
	// RemoveNamespaceLabel is not supported
	return &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "RemoveNamespaceLabel()",
	}
}

func (d *Dcos) GetNamespaceLabel(namespace string) (map[string]string, error) {
	// GetNamespaceLabel is not supported
	return nil, &errors.ErrNotSupported{
		Type:      "Function",
		Operation: "GetNamespaceLabel()",
	}
}

func init() {
	d := &Dcos{}
	d.marathonClient = &marathonOps{}

	scheduler.Register(SchedName, d)
}
