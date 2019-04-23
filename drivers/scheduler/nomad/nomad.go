package nomad

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	docker "github.com/docker/docker/client"
	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/portworx/torpedo/drivers/volume"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	// SchedName is the name of the nomad scheduler driver implementation
	SchedName = "nomad"

	// DefaultWaitTime is just a gloal wait time
	DefaultWaitTime = 10 * time.Second

	// DefaultTimeout is just a global timeout
	DefaultTimeout = 2 * time.Minute

	// DefaultRetry is just a global retry
	DefaultRetry = 20
)

type nomad struct {
	dockerClient   *docker.Client
	nomadClient    *api.Client
	specFactory    *spec.Factory
	nodeDriverName string
	volDriverName  string
}

func (nm *nomad) Init(specDir, volDriverName, nodeDriverName string) error {
	var err error

	nomadAPI := os.Getenv("NOMAD_API_ADDR")
	if nomadAPI == "" {
		nomadAPI, err = os.Hostname()
		if err != nil {
			return fmt.Errorf("Failed to get nomad API endpoint, because of Err: %v", err)
		}
	}

	NomadMaster := fmt.Sprintf("http://%s:4646", nomadAPI)

	// Initialize nomad client
	nm.nomadClient, err = api.NewClient(&api.Config{Address: NomadMaster})
	if err != nil {
		return err
	}

	// Get nomad node list
	nomadNodeList, err := nm.getNomadNodes()
	if err != nil {
		return err
	}

	if len(nomadNodeList) < 0 {
		return fmt.Errorf("Unable to proceed with %d nomad nodes", len(nomadNodeList))
	}

	for _, nodeInfo := range nomadNodeList {
		newNode := nm.parseNomadNode(nodeInfo)
		/*
			if err := nm.IsNodeReady(newNode); err != nil {
				return err
			}
		*/
		if err := node.AddNode(newNode); err != nil {
			return err
		}
	}

	// Get all apps
	nm.specFactory, err = spec.NewFactory(specDir, nm)
	if err != nil {
		return err
	}

	// Initialize docker client
	nm.dockerClient, err = docker.NewEnvClient()
	if err != nil {
		return err
	}

	// Node driver
	nm.nodeDriverName = nodeDriverName

	// Volume driver
	nm.volDriverName = volDriverName

	return nil
}

func (nm *nomad) getNomadNodes() ([]*api.Node, error) {
	var nomadNodeList []*api.Node

	// Get list of nodes
	nodeList, _, err := nm.nomadClient.Nodes().List(&api.QueryOptions{})
	if err != nil {
		return nil, err
	}

	// Get info about each node from the list and add it to node collection
	for _, nomadNode := range nodeList {
		nodeInfo, _, err := nm.nomadClient.Nodes().Info(nomadNode.ID, &api.QueryOptions{})
		if err != nil {
			return nil, err
		}
		nomadNodeList = append(nomadNodeList, nodeInfo)
	}

	return nomadNodeList, nil
}

// This function parses nomad node into node.Node struct format
func (nm *nomad) parseNomadNode(n *api.Node) node.Node {
	var nodeType node.Type

	// All nodes are workers in nomad cluster, all have px installed
	nodeType = node.TypeWorker

	return node.Node{
		Name:      n.ID,
		Addresses: []string{strings.TrimSuffix(n.HTTPAddr, ":4646")},
		Type:      nodeType,
	}
}

// This function gets nomad cluster leader IP
func (nm *nomad) isLeader() string {
	// Get leader
	leader, err := nm.nomadClient.Status().Leader()
	if err != nil {
		return ""
	}

	leader = strings.TrimSuffix(leader, ":4647")

	return leader
}

// String returns the string name of this driver.
func (nm *nomad) String() string {
	return SchedName
}

func (nm *nomad) IsNodeReady(n node.Node) error {
	// TODO: Implement this method
	return fmt.Errorf("IsNodeReady() is not implemented yet")
}

func (nm *nomad) ParseSpecs(specDir string) ([]interface{}, error) {
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
		jobF, err := jobspec.ParseFile(fileName)
		if err != nil {
			return nil, err
		}
		specs = append(specs, jobF)
	}

	return specs, nil
}

func (nm *nomad) Schedule(instanceID string, options scheduler.ScheduleOptions) ([]*scheduler.Context, error) {
	var apps []*spec.AppSpec
	var contexts []*scheduler.Context
	var ctx *scheduler.Context

	// Clean GC
	nm.cleanGC()

	if len(options.AppKeys) > 0 {
		for _, key := range options.AppKeys {
			spec, err := nm.specFactory.Get(key)
			if err != nil {
				return nil, err
			}
			apps = append(apps, spec)
		}
	} else {
		apps = nm.specFactory.GetAll()
	}

	for _, app := range apps {
		// Set app name
		appName := app.GetID(instanceID)

		// Create spec objects
		specObjects, err := nm.createSpecObjects(app, appName)
		if err != nil {
			return nil, err
		}

		ctx = &scheduler.Context{
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

func (nm *nomad) createSpecObjects(spec *spec.AppSpec, appName string) ([]interface{}, error) {
	var jobObjects []interface{}

	for _, app := range spec.SpecList {
		if application, ok := app.(*api.Job); ok {
			// Set job unique Name and ID
			*application.ID = appName
			*application.Name = appName

			// Set jib unique volumes
			if err := nm.randomizeVolumeNames(application); err != nil {
				return nil, &scheduler.ErrFailedToScheduleApp{
					App:   spec,
					Cause: err.Error(),
				}
			}

			// Create new job
			job, err := nm.registerJob(application)
			if err != nil {
				return nil, &scheduler.ErrFailedToScheduleApp{
					App:   spec,
					Cause: err.Error(),
				}
			}

			// Get job info
			obj, err := nm.getJobInfo(job)
			if err != nil {
				return nil, &scheduler.ErrFailedToScheduleApp{
					App:   spec,
					Cause: err.Error(),
				}
			}

			if obj != nil {
				jobObjects = append(jobObjects, obj)
			}
		} else {
			return nil, fmt.Errorf("Invalid spec received for app: %v", spec.Key)
		}

	} // End of app.SpecList forloop

	return jobObjects, nil
}

func (nm *nomad) registerJob(jobSpec *api.Job) (*api.Job, error) {
	// Register job
	_, _, err := nm.nomadClient.Jobs().Register(jobSpec, &api.WriteOptions{})
	if err != nil {
		return nil, err
	}

	return jobSpec, nil
}

func (nm *nomad) randomizeVolumeNames(application *api.Job) error {
	// Get volume driver
	volDriver, err := volume.Get(nm.volDriverName)
	if err != nil {
		return err
	}

	var newVolumes []interface{}
	for _, tskG := range application.TaskGroups {
		for _, tsk := range tskG.Tasks {
			for key, value := range tsk.Config {
				// Find volumes
				if key == "volumes" {
					// Get volume List
					volList, ok := value.([]interface{})
					if !ok {
						return fmt.Errorf("VolList is not ok, %+v", volList)
					}
					// Get each volume from volume list
					for _, v := range volList {
						vol, ok := v.(string)
						if !ok {
							return fmt.Errorf("Volume is not ok, %+v", vol)
						}
						vol = volDriver.RandomizeVolumeName(vol)
						newVolumes = append(newVolumes, vol)
					}
					tsk.Config[key] = newVolumes
				}
			}
		}
	}

	return nil
}

func (nm *nomad) GetNodesForApp(ctx *scheduler.Context) ([]node.Node, error) {
	var appNodeList []node.Node

	// Clean garbage collector
	if err := nm.cleanGC(); err != nil {
		return nil, err
	}

	for _, spec := range ctx.App.SpecList {
		if application, ok := spec.(*api.Job); ok {
			// Get nodes that used by app
			for _, nodeL := range node.GetNodes() {
				var isNode = nodeL
				for _, nodeA := range nm.getAlloc(application) {
					if isNode.Name == nodeA.NodeID {
						appNodeList = append(appNodeList, isNode)
					} else {
						continue
					}
				}
			}
		} else {
			logrus.Warnf("Invalid spec received for app: %v", ctx.App.Key)
		}
	}

	if len(appNodeList) < 0 {
		return nil, fmt.Errorf("Didn't find any nodes for app: %v", ctx.App.Key)
	}

	return appNodeList, nil
}

func (nm *nomad) getAlloc(application *api.Job) []*api.AllocationListStub {
	// clean GC
	nm.cleanGC()

	// Get nodes that used by app
	allocNodeList, _, err := nm.nomadClient.Jobs().Allocations(*application.Name, true, &api.QueryOptions{})
	if err != nil {
		logrus.Errorf("Failed to get allocations for app: %s, because of Err: %v", *application.Name, err)
		return nil
	}

	return allocNodeList
}

func (nm *nomad) WaitForRunning(ctx *scheduler.Context, timeout, retryInterval time.Duration) error {
	var status string
	var allocNodeList []*api.AllocationListStub
	var maxCount int

	for _, spec := range ctx.App.SpecList {
		if application, ok := spec.(*api.Job); ok {
			for _, tskGrp := range application.TaskGroups {
				maxCount = *tskGrp.Count
			}

			// Get job status
			for i := 0; i < DefaultRetry; i++ {
				jobInfo, err := nm.getJobInfo(application)
				status = *jobInfo.Status
				if err != nil {
					time.Sleep(DefaultWaitTime)
					continue
				}

				// Check if status is running
				if status == "running" {
					for i := 0; i < DefaultRetry; i++ {
						// Get Allocation list
						allocNodeList = nm.getAlloc(application)

						readyAlloc := 0
						for _, alloc := range allocNodeList {
							if alloc.ClientStatus == "running" {
								logrus.Infof("Current allocation state: %s. Expected allocation state: running", alloc.ClientStatus)
								readyAlloc = readyAlloc + 1
								continue
							}
							logrus.Infof("Current allocation state: %s. Expected allocation state: running", alloc.ClientStatus)
							continue
						}

						if readyAlloc != maxCount {
							logrus.Infof("Current allocations running: %d. Expected allocations running: %d. Next retry in: %v", readyAlloc, maxCount, DefaultWaitTime)
							time.Sleep(DefaultWaitTime)
							continue
						} else {
							logrus.Infof("All expected allocations are running: %d/%d", readyAlloc, maxCount)
							return nil
						}
					}
				}
				logrus.Infof("Current app %v status: %s. Expected app %v status: running. Next retry in: %v", *application.Name, status, *application.Name, DefaultWaitTime)
				time.Sleep(DefaultWaitTime)
			}
		} else {
			logrus.Warnf("Invalid spec received for app: %v", ctx.App.Key)
		}
	}

	return fmt.Errorf("Failed to get running status in time, status: %s", status)
}

// This function gets list of existing jobs and returns it
func (nm *nomad) getJobList() ([]*api.JobListStub, error) {
	var jobList []*api.JobListStub

	// Get job list
	jobList, _, err := nm.nomadClient.Jobs().List(&api.QueryOptions{})
	if err != nil {
		return nil, fmt.Errorf("Failed to get list of jobs, because of Err: %v", err)
	}

	return jobList, nil
}

// This functing check if job exists
func (nm *nomad) jobExist(myJob *api.Job) bool {
	// Get list of all jobs
	jobList, err := nm.getJobList()
	if err != nil {
		return false
	}

	if len(jobList) == 0 {
		logrus.Warn("No existing jobs found")
		return false
	}

	// Find requested job in list
	for _, job := range jobList {
		if job.ID == *myJob.Name {
			return true
		}
	}

	return false
}

// This function gets info about specific job and returns it
func (nm *nomad) getJobInfo(jobSpec *api.Job) (*api.Job, error) {
	var jobInfo *api.Job

	// Get job Info
	jobInfo, _, err := nm.nomadClient.Jobs().Info(*jobSpec.ID, &api.QueryOptions{})
	if err != nil {
		return nil, fmt.Errorf("Failed to get job info, because of Err: %v", err)
	}

	return jobInfo, nil
}

func (nm *nomad) getTasks(jobSpec *api.Job) []*api.Task {
	var taskList []*api.Task

	for _, tskGrp := range jobSpec.TaskGroups {
		for _, tsk := range tskGrp.Tasks {
			taskList = append(taskList, tsk)
		}
	}

	return taskList
}

func (nm *nomad) AddTasks(ctx *scheduler.Context, options scheduler.ScheduleOptions) error {
	// TODO: Implement this method
	return fmt.Errorf("AddTasks() is not implemented yet")
}

func (nm *nomad) Destroy(ctx *scheduler.Context, opts map[string]bool) error {
	var err error
	var evalID interface{}

	for _, spec := range ctx.App.SpecList {
		t := func() (interface{}, bool, error) {
			if evalID, err = nm.destroySpecObjects(spec, opts, ctx.App); err != nil {
				return nil, true, err
			}
			return evalID, false, nil
		}
		_, err := task.DoRetryWithTimeout(t, DefaultTimeout, DefaultWaitTime)
		if err != nil {
			return err
		}
	}

	return nil
}

func (nm *nomad) destroySpecObjects(spec interface{}, opts map[string]bool, app *spec.AppSpec) (interface{}, error) {
	var evalID string
	purge := true

	if application, ok := spec.(*api.Job); ok {
		evalID, err := nm.deleteJob(application, purge)
		if err != nil {
			return nil, &scheduler.ErrFailedToDestroyApp{
				App:   app,
				Cause: fmt.Sprintf("Failed to destroy app: %v. Err: %v", *application.Name, err),
			}
		}
		logrus.Infof("Successfully destroyed app: %v with id: %s", *application.Name, evalID)
	} else {
		return nil, fmt.Errorf("Invalid spec received for app: %v", app.Key)
	}
	return evalID, nil
}

func (nm *nomad) deleteJob(jobSpec *api.Job, purge bool) (string, error) {
	// Delete job
	evalID, _, err := nm.nomadClient.Jobs().Deregister(*jobSpec.ID, purge, &api.WriteOptions{})
	if err != nil {
		return evalID, err
	}

	// Let deleted jobs to stabilize
	logrus.Infof("Wait for %v to make sure volumes got unmounted", DefaultWaitTime)
	time.Sleep(DefaultWaitTime)

	return evalID, nil
}

func (nm *nomad) cleanGC() error {
	// Clean garbage collector
	if err := nm.nomadClient.System().GarbageCollect(); err != nil {
		return fmt.Errorf("Failed to clean garbage collector, Err: %v", err)
	}

	return nil
}

func (nm *nomad) WaitForDestroy(ctx *scheduler.Context) error {
	var status string
	var jobExist bool

	for _, spec := range ctx.App.SpecList {
		if application, ok := spec.(*api.Job); ok {
			jobExist = nm.jobExist(application)
			if jobExist == false {
				logrus.Warnf("Application %v doesn't exist", *application.Name)

				// Clean GC
				nm.cleanGC()

				return nil
			}

			// Get job Info
			for i := 0; i < DefaultRetry; i++ {
				jobInfo, err := nm.getJobInfo(application)
				status = *jobInfo.Status
				if err != nil {
					logrus.Warnf("Failed to get job status, because of Err: %v. Next retry in: %v", err, DefaultWaitTime)
					time.Sleep(DefaultWaitTime)
					continue
				}
				// Check if status is running
				if status == "dead" {
					logrus.Infof("Job %v is %s", *application.Name, status)

					// Clean GC
					nm.cleanGC()

					return nil
				}

				logrus.Warnf("Current job %v status: %s. Expected job %v status: dead. Next retry in: %v", *application.Name, status, *application.Name, DefaultWaitTime)
				time.Sleep(DefaultWaitTime)
			}
		} else {
			logrus.Warnf("Invalid spec received for app: %v", ctx.App.Key)
		}
	}

	return fmt.Errorf("Failed to get dead status in time, status: %s", status)
}

func (nm *nomad) DeleteTasks(ctx *scheduler.Context) error {
	// Get node driver
	driver, _ := node.Get(nm.nodeDriverName)
	connOpts := node.ConnectionOpts{
		TimeBeforeRetry: DefaultWaitTime,
		Timeout:         DefaultTimeout,
		IgnoreError:     true,
	}

	for _, spec := range ctx.App.SpecList {
		if application, ok := spec.(*api.Job); ok {
			// Filter out node that ae used by app from all nodes
			for _, nodeL := range node.GetNodes() {
				var isNode = nodeL
				for _, nodeA := range nm.getAlloc(application) {
					if isNode.Name == nodeA.NodeID {
						// Delete container
						cmd := fmt.Sprintf("docker rm -f %v-%v", *application.Name, nodeA.ID)
						_, err := driver.RunCommand(isNode, cmd, connOpts)
						if err != nil {
							return fmt.Errorf("Failed to execute: %s on node with IP: %s", cmd, isNode.Addresses)
						}
					} else {
						continue
					}
				}
			}
			logrus.Infof("Deleted docker containers for all allocations for app: %v", *application.Name)
		} else {
			logrus.Warnf("Invalid spec received for app: %v", ctx.App.Key)
		}
	}

	return nil
}

func (nm *nomad) GetVolumeParameters(ctx *scheduler.Context) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	populateParamsFunc := func(volName string, volParams map[string]string) error {
		result[volName] = volParams
		return nil
	}

	if err := nm.volumeOperation(ctx, populateParamsFunc); err != nil {
		return nil, err
	}

	return result, nil
}

func (nm *nomad) InspectVolumes(ctx *scheduler.Context, timeout, retryInterval time.Duration) error {
	inspectDockerVolumeFunc := func(volName string, _ map[string]string) error {
		t := func() (interface{}, bool, error) {
			out, err := nm.dockerClient.VolumeInspect(context.Background(), volName)
			return out, true, err
		}

		if _, err := task.DoRetryWithTimeout(t, timeout, retryInterval); err != nil {
			return &scheduler.ErrFailedToValidateStorage{
				App:   ctx.App,
				Cause: fmt.Sprintf("Failed to inspect docker volume: %v. Err: %v", volName, err),
			}
		}
		return nil
	}

	if err := nm.volumeOperation(ctx, inspectDockerVolumeFunc); err != nil {
		return err
	}

	return nil
}

func (nm *nomad) volumeOperation(ctx *scheduler.Context, f func(string, map[string]string) error) error {
	// Nomad does not have volume objects like Kubernetes. We get the volume information from
	// the app spec and get the options parsed from the respective volume driver
	volDriver, err := volume.Get(nm.volDriverName)
	if err != nil {
		return err
	}

	for _, spec := range ctx.App.SpecList {
		if application, ok := spec.(*api.Job); ok {
			for _, tskG := range application.TaskGroups {
				for _, tsk := range tskG.Tasks {
					for key, value := range tsk.Config {
						// Find volumes
						if key == "volumes" {
							// Get volume List
							volList, ok := value.([]interface{})
							if !ok {
								return fmt.Errorf("Volume list is not ok! Content: %+v", volList)
							}
							// Get each volume from volume list
							for _, v := range volList {
								vol, ok := v.(string)
								if !ok {
									return fmt.Errorf("Volume is not ok! Content: %+v", v)
								}
								volName, volParams, err := volDriver.ExtractVolumeInfo(vol)
								if err != nil {
									return &scheduler.ErrFailedToGetVolumeParameters{
										App:   nil,
										Cause: fmt.Sprintf("Failed to extract volume info: %v. Err: %v", vol, err),
									}
								}

								if err := f(volName, volParams); err != nil {
									return err
								}
							}
						}
					}
				}
			}
		} else {
			logrus.Warnf("Invalid spec received for app: %v", ctx.App.Key)
		}
	}

	return nil
}

func (nm *nomad) DeleteVolumes(ctx *scheduler.Context) ([]*volume.Volume, error) {
	var vols []*volume.Volume

	deleteDockerVolumeFunc := func(volName string, _ map[string]string) error {
		vols = append(vols, &volume.Volume{Name: volName})
		t := func() (interface{}, bool, error) {
			return nil, true, nm.dockerClient.VolumeRemove(context.Background(), volName, false)
		}

		if _, err := task.DoRetryWithTimeout(t, DefaultTimeout, DefaultWaitTime); err != nil {
			return &scheduler.ErrFailedToDestroyStorage{
				App:   ctx.App,
				Cause: fmt.Sprintf("Failed to remove docker volume: %v. Err: %v", volName, err),
			}
		}
		return nil
	}

	if err := nm.volumeOperation(ctx, deleteDockerVolumeFunc); err != nil {
		return nil, err
	}

	return vols, nil
}

func (nm *nomad) GetVolumes(ctx *scheduler.Context) ([]*volume.Volume, error) {
	var vols []*volume.Volume

	inspectDockerVolumeFunc := func(volName string, _ map[string]string) error {
		vols = append(vols, &volume.Volume{Name: volName})
		t := func() (interface{}, bool, error) {
			out, err := nm.dockerClient.VolumeInspect(context.Background(), volName)
			return out, true, err
		}

		if _, err := task.DoRetryWithTimeout(t, DefaultTimeout, DefaultWaitTime); err != nil {
			return &scheduler.ErrFailedToValidateStorage{
				App:   ctx.App,
				Cause: fmt.Sprintf("Failed to inspect docker volume: %v. Err: %v", volName, err),
			}
		}
		return nil
	}

	if err := nm.volumeOperation(ctx, inspectDockerVolumeFunc); err != nil {
		return nil, err
	}

	return vols, nil
}

func (nm *nomad) ResizeVolume(ctx *scheduler.Context) ([]*volume.Volume, error) {
	// TODO: Implement this method
	return []*volume.Volume{}, fmt.Errorf("ResizeVolume() is not implemented yet")
}

func (nm *nomad) GetSnapshots(ctx *scheduler.Context) ([]*volume.Snapshot, error) {
	// TODO: Implement this method
	return []*volume.Snapshot{}, fmt.Errorf("GetSnapshots() is not implemented yet")
}

func (nm *nomad) Describe(ctx *scheduler.Context) (string, error) {
	// TODO: Implement this method
	return "", fmt.Errorf("Describe() is not implemented yet")
}

func (nm *nomad) ScaleApplication(ctx *scheduler.Context, scaleFactorMap map[string]int32) error {
	for _, spec := range ctx.App.SpecList {
		// Check if scalable application
		if !nm.IsScalable(spec) {
			continue
		}

		if application, ok := spec.(*api.Job); ok {
			for _, tskGrp := range application.TaskGroups {
				// Update scale factor
				newScaleFactor := scaleFactorMap[*application.Name]
				*tskGrp.Count = int(newScaleFactor)
			}

			// Update job
			_, err := nm.registerJob(application)
			if err != nil {
				return &scheduler.ErrFailedToUpdateApp{
					App:   ctx.App,
					Cause: err.Error(),
				}
			}
		}
	}

	return nil
}

func (nm *nomad) GetScaleFactorMap(ctx *scheduler.Context) (map[string]int32, error) {
	result := make(map[string]int32, len(ctx.App.SpecList))

	// Clean garbage collector
	nm.cleanGC()

	// Get current scale factor
	for _, spec := range ctx.App.SpecList {
		if application, ok := spec.(*api.Job); ok {
			for _, tskG := range application.TaskGroups {
				appName := string(*application.Name)
				scaleCount := int32(*tskG.Count)
				result[appName] = scaleCount
			}
		}
	}

	return result, nil
}

func (nm *nomad) StopSchedOnNode(n node.Node) error {
	// TODO: Implement this method
	return fmt.Errorf("StopSchedOnNode() is not implemented yet")
}

func (nm *nomad) StartSchedOnNode(n node.Node) error {
	// TODO: Implement this method
	return fmt.Errorf("StartSchedOnNode() is not implemented yet")
}

func (nm *nomad) RescanSpecs(specDir string) error {
	var err error

	logrus.Infof("Rescanning specs for %v", specDir)
	nm.specFactory, err = spec.NewFactory(specDir, nm)
	if err != nil {
		return err
	}
	return nil
}

func (nm *nomad) IsScalable(spec interface{}) bool {
	if application, ok := spec.(*api.Job); ok {
		for _, tskG := range application.TaskGroups {
			for _, tsk := range tskG.Tasks {
				for key, value := range tsk.Config {
					// Find volumes
					if key == "volumes" {
						// Get volume List
						volList, ok := value.([]interface{})
						if !ok {
							logrus.Errorf("Volume list is not ok! Content: %+v", volList)
							return false
						}
						// Get each volume from volume list
						for _, v := range volList {
							vol, ok := v.(string)
							if !ok {
								logrus.Errorf("Volume is not ok! Content: %+v", v)
								return false
							}

							// Do not scale app if volume is not shared
							if !strings.Contains(vol, "shared") {
								return false
							}
						}
					}
				}
			}
		}
	} else {
		logrus.Errorf("Invalid spec received for app")
		return false
	}

	return true
}

func init() {
	nm := &nomad{}
	scheduler.Register(SchedName, nm)
}
