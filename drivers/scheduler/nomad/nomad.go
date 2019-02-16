package nomad

import (
	"fmt"
	//"io/ioutil"
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
	// NomadMaster is a nodemad api concact (subject to change)
	NomadMaster = "http://70.0.57.114:4646"
)

type nomad struct {
	dockerClient  *docker.Client
	nomadClient   *api.Client
	specFactory   *spec.Factory
	volDriverName string
}

func (nm *nomad) Init(specDir, volDriverName, nodeDriverName string) error {
	var err error
	fmt.Println("KOKADBG: Init(): START")

	// Initialize nomad client
	nm.nomadClient, err = api.NewClient(&api.Config{Address: NomadMaster})
	if err != nil {
		return err
	}

	// TODO Make sure list is not empty
	nomadNodeList, err := nm.getNomadNodes()
	if err != nil {
		return err
	}

	for _, nodeInfo := range nomadNodeList {
		newNode := nm.parseNomadNode(nodeInfo)
		if err := nm.IsNodeReady(newNode); err != nil {
			return err
		}
		if err := node.AddNode(newNode); err != nil {
			return err
		}
	}

	// TODO REMOVE LATER ON: DEBUG: PRINT NODE FROM LIST
	nodelist := node.GetNodes()
	for _, nd := range nodelist {
		fmt.Printf("KOKADBG: NODE: %+v\n", nd)
	}
	// END

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

	// Driver name
	nm.volDriverName = volDriverName

	fmt.Println("KOKADBG: Init(): END")
	return nil
}

func (nm *nomad) getNomadNodes() ([]*api.Node, error) {
	fmt.Println("KOKADBG: getNomadNodes(): START")
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

	fmt.Println("KOKADBG: getNomadNodes(): END")
	return nomadNodeList, nil
}

// This function parses nomad node into node.Node struct format
func (nm *nomad) parseNomadNode(n *api.Node) node.Node {
	fmt.Println("KOKADBG: parseNomadNode(): START")
	var nodeType node.Type

	// All nodes are workers in nomad cluster, all have px installed
	nodeType = node.TypeWorker

	fmt.Println("KOKADBG: parseNomadNode(): END")
	return node.Node{
		Name:      n.ID,
		Addresses: []string{strings.TrimSuffix(n.HTTPAddr, ":4646")},
		Type:      nodeType,
	}
}

// This function gets nomad cluster leader IP
func (nm *nomad) isLeader() string {
	fmt.Println("KOKADBG: isLeader(): START")
	// Get leader
	leader, err := nm.nomadClient.Status().Leader()
	if err != nil {
		return ""
	}

	leader = strings.TrimSuffix(leader, ":4647")
	fmt.Println("KOKADBG: isLeader(): END")
	return leader
}

// String returns the string name of this driver.
func (nm *nomad) String() string {
	return SchedName
}

func (nm *nomad) IsNodeReady(n node.Node) error {
	// TODO: Implement this method
	return nil
}

func (nm *nomad) ParseSpecs(specDir string) ([]interface{}, error) {
	fmt.Println("KOKADBG: ParseSpecs(): START")
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

	fmt.Println("KOKADBG: ParseSpecs(): END")
	return specs, nil
}

func (nm *nomad) Schedule(instanceID string, options scheduler.ScheduleOptions) ([]*scheduler.Context, error) {
	fmt.Println("KOKADBG: Schedule(): START")
	var apps []*spec.AppSpec
	var contexts []*scheduler.Context

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
		var specObjects []interface{}
		for _, spec := range app.SpecList {
			//fmt.Printf("KOKADBG: LOOP app.SpecList: spec: %v\n", spec)
			if application, ok := spec.(*api.Job); ok {
				//fmt.Printf("\nKOKADBG: APPLICATION NAME: %+v\n", *application.Name)
				if err := nm.randomizeVolumeNames(application); err != nil {
					return nil, &scheduler.ErrFailedToScheduleApp{
						App:   app,
						Cause: err.Error(),
					}
				}
				//fmt.Printf("\nKOKADBG: GONNA createJob()\n")
				// Create new job
				obj, err := nm.createJob(application)
				if err != nil {
					return nil, &scheduler.ErrFailedToScheduleApp{
						App:   app,
						Cause: err.Error(),
					}
				}
				//fmt.Printf("\nKOKADBG: createJob() DONE!\n")
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

	fmt.Println("KOKADBG: Schedule(): END")
	return contexts, nil
}

func (nm *nomad) createJob(jobSpec *api.Job) (*api.Job, error) {
	fmt.Println("KOKADBG: createJob(): START")

	// Register job
	_, _, err := nm.nomadClient.Jobs().Register(jobSpec, &api.WriteOptions{})
	if err != nil {
		return nil, err
	}

	//fmt.Printf("KOKADBG: JOB RESPONSE ID: %s\n", jobResponse.EvalID)
	fmt.Println("KOKADBG: createJob(): END")
	return jobSpec, nil
}

func (nm *nomad) randomizeVolumeNames(application *api.Job) error {
	fmt.Println("KOKADBG: randomizeVolumeNames(): START")
	volDriver, err := volume.Get(nm.volDriverName)
	if err != nil {
		return fmt.Errorf("KOKADBG: FAILED TO GET VOL DRIVER: %v", err)
	}

	//fmt.Printf("KOKADBG: randomizeVolumeNames: After volume.Get(): volDriver: %s\n", volDriver)
	var newVolumes []interface{}
loop:
	for _, tskG := range application.TaskGroups {
		for _, tsk := range tskG.Tasks {
			for key, value := range tsk.Config {
				//fmt.Printf("KOKADBG: CFG RANGE KEY: %+v\n", key)
				//fmt.Printf("KOKADBG: CFG RANGE VALUE: %+v\n", value)

				// Find volumes
				if key == "volumes" {
					//fmt.Printf("KOKADBG: Key: %v, Value: %v\n", key, value)

					// Get volume List
					volList, ok := value.([]interface{})
					if !ok {
						fmt.Printf("KOKADBG: randomizeVolumeNames(): ERROR: volList Not OK! volList Content: %+v\n", volList)
					}

					// Get each volume from volume list
					for _, v := range volList {
						vol, ok := v.(string)
						if !ok {
							fmt.Printf("KOKADBG: randomizeVolumeNames(): ERROR: vol Not OK! vol Content: %+v\n", vol)
						}
						vol = volDriver.RandomizeVolumeName(vol)
						//fmt.Printf("KOKADBG: VOL: %v\n", vol)
						newVolumes = append(newVolumes, vol)
						//fmt.Printf("KOKADBG: NEWVOLUMES INSIDE LOOP: %v\n", newVolumes)
					}
					tsk.Config[key] = newVolumes
					break loop
				}
			}
		}
	}

	fmt.Println("KOKADBG: randomizeVolumeNames(): END")
	return nil
}

func (nm *nomad) GetNodesForApp(ctx *scheduler.Context) ([]node.Node, error) {
	// TODO: Implement this method
	fmt.Println("KOKADBG: GetNodesForApp(): START")
	nodeList := node.GetNodes()
	for _, nd := range nodeList {
		fmt.Printf("KOKADBG: GetNodesForApp(): Node: %+v\n", nd)
	}

	// Get allocations
	alloc, _, err := nm.nomadClient.Jobs().Allocations("nginx", true, &api.QueryOptions{})
	if err != nil {
		return nil, err
	}

	// Clean garbage collector
	if err := nm.cleanGC(); err != nil {
		return nil, err
	}

	fmt.Println("KOKADBG: GetNodesForApp(): Sleep 10 seconds\n")
	time.Sleep(10 * time.Second)
	// Get allocations
	alloc, _, err = nm.nomadClient.Jobs().Allocations("nginx", true, &api.QueryOptions{})
	if err != nil {
		return nil, err
	}

	for _, nd := range alloc {
		fmt.Printf("KOKADBG: GetNodesForApp(): Nodes Allocated: Nomad NodeID: %v\n", nd.NodeID)
	}

	fmt.Println("KOKADBG: GetNodesForApp(): END")
	return nodeList, nil
}

func (nm *nomad) WaitForRunning(ctx *scheduler.Context, timeout, retryInterval time.Duration) error {
	fmt.Println("KOKADBG: WaitForRunning(): START")
	var status string

	for _, spec := range ctx.App.SpecList {
		if obj, ok := spec.(*api.Job); ok {

			// Get job Info
			for i := 0; i < 10; i++ {
				jobInfo, err := nm.getJobInfo(obj)
				status = *jobInfo.Status
				if err != nil {
					fmt.Printf("KOKADBG: ERROR CONTINUE in 10: %v\n", err)
					time.Sleep(10 * time.Second)
					continue
				}
				// Check if status is running
				if status == "running" {
					fmt.Printf("KOKADBG: WaitForRunning: SUCCESSFULL JOB STATUS: %v, OK!\n", status)
					fmt.Println("KOKADBG: WaitForRunning(): END")
					return nil
				}

				fmt.Printf("KOKADBG: WaitForRunning: JOB STATUS: %v\n", status)
				time.Sleep(10 * time.Second)
			}
		} else {
			logrus.Warnf("Invalid spec received for app %v in WaitForRunning", ctx.App.Key)
		}
	}
	fmt.Println("KOKADBG: WaitForRunning(): FAIL: END")
	return fmt.Errorf("KOKADBG: Failed to get running status in time, Status: %s\n", status)
}

// This function gets list of existing jobs and returns it
func (nm *nomad) getJobList() ([]*api.JobListStub, error) {
	var jobList []*api.JobListStub
	fmt.Println("KOKADBG: getJobList(): START")

	// Get job list
	jobList, _, err := nm.nomadClient.Jobs().List(&api.QueryOptions{})
	if err != nil {
		return nil, fmt.Errorf("KOKADBG: getJobList(): ERROR JOB LIST: %v\n", err)
	}

	fmt.Println("KOKADBG: getJobList(): END")
	return jobList, nil
}

// This functing check if job exists
func (nm *nomad) jobExist(myJob *api.Job) bool {
	fmt.Println("KOKADBG: jobExist(): START")
	jobList, err := nm.getJobList()
	if err != nil {
		return false
	}

	// Check if list is not empty
	if len(jobList) == 0 {
		logrus.Warn("No existing jobs found")
		return false
	}

	// Find requested job in list
	for _, job := range jobList {
		fmt.Printf("KOKADBG: jobExist(): job.ID: %v, job.Name: %v\n", job.ID, job.Name)
		fmt.Printf("KOKADBG: jobExist(): myJob.Name: %v, myJob.ID: %v\n", myJob.Name, myJob.ID)
		if job.ID == *myJob.Name {
			fmt.Printf("KOKADBG: jobExist(): FOUND: job.ID: %v, job.Name: %v\n", job.ID, job.Name)
			fmt.Printf("KOKADBG: jobExist(): FOUND: myJob.Name: %v, myJob.ID: %v\n", myJob.Name, myJob.ID)
			return true
		}
	}
	fmt.Println("KOKADBG: jobExist(): END")
	return false
}

// This function gets info about specific job and returns it
func (nm *nomad) getJobInfo(jobSpec *api.Job) (*api.Job, error) {
	var jobInfo *api.Job
	fmt.Println("KOKADBG: getJobInfo(): START")

	// Get job Info
	jobInfo, _, err := nm.nomadClient.Jobs().Info(*jobSpec.ID, &api.QueryOptions{})
	if err != nil {
		return nil, fmt.Errorf("KOKADBG: getJobInfo(): ERROR INFO JOB: %v\n", err)
	}

	fmt.Println("KOKADBG: getJobInfo(): END")
	return jobInfo, nil
}

func (nm *nomad) AddTasks(ctx *scheduler.Context, options scheduler.ScheduleOptions) error {
	// TODO: Implement this method
	return nil
}

func (nm *nomad) Destroy(ctx *scheduler.Context, opts map[string]bool) error {
	// TODO: Implement this method
	fmt.Println("KOKADBG: Destroy(): START")
	for _, spec := range ctx.App.SpecList {
		if obj, ok := spec.(*api.Job); ok {
			if err := nm.deleteJob(obj); err != nil {
				return &scheduler.ErrFailedToDestroyApp{
					App:   ctx.App,
					Cause: fmt.Sprintf("Failed to destroy Application: %v. Err: %v", obj.ID, err),
				}
			}
			logrus.Infof("[%v] Destroyed application: %v", ctx.App.Key, obj.ID)
		} else {
			logrus.Warnf("Invalid spec received for app %v in Destroy", ctx.App.Key)
		}
	}
	fmt.Println("KOKADBG: Destroy(): END")
	return nil
}

func (nm *nomad) deleteJob(jobSpec *api.Job) error {
	// TODO delete job
	fmt.Println("KOKADBG: deleteJob(): START")

	// Delete job
	_, _, err := nm.nomadClient.Jobs().Deregister(*jobSpec.ID, true, &api.WriteOptions{})
	if err != nil {
		return err
	}

	// TODO TEST SLEEP
	fmt.Println("KOKADBG: deleteJob(): Sleeping for 20 seconds to make sure volumes got unmounted")
	time.Sleep(20 * time.Second)

	fmt.Println("KOKADBG: deleteJob(): END")
	return nil
}

func (nm *nomad) cleanGC() error {
	fmt.Println("KOKADBG: cleanGC(): START")

	// Clean garbage collector
	if err := nm.nomadClient.System().GarbageCollect(); err != nil {
		return fmt.Errorf("Failed to clean garbage collector, Err: %v", err)
	}

	fmt.Println("KOKADBG: cleanGC(): END")
	return nil
}

func (nm *nomad) WaitForDestroy(ctx *scheduler.Context) error {
	// TODO: Implement this method
	fmt.Println("KOKADBG: WaitForDestroy(): START")
	var status string
	var jobExist bool

	for _, spec := range ctx.App.SpecList {
		if obj, ok := spec.(*api.Job); ok {
			// TODO Check if job exists, what to do if multiple jobs? For now just works with nginx
			jobExist = nm.jobExist(obj)
			if jobExist == false {
				fmt.Printf("KOKADBG: WaitForDestroy(): Job doesn't exist\n")
				return nil
			}

			// Get job Info
			for i := 0; i < 10; i++ {
				jobInfo, err := nm.getJobInfo(obj)
				status = *jobInfo.Status
				if err != nil {
					fmt.Printf("KOKADBG: WaitForDestroy: ERROR CONTINUE in 10: %v\n", err)
					time.Sleep(10 * time.Second)
					continue
				}
				// Check if status is running
				if status == "dead" {
					fmt.Printf("KOKADBG: WaitForDestroy: SUCCESSFULL (dead) JOB STATUS: %v, OK!\n", status)
					fmt.Println("KOKADBG: WaitForDestroy(): END")
					return nil
				}

				fmt.Printf("KOKADBG: WaitForDestroy: JOB STATUS: %v\n", status)
				time.Sleep(10 * time.Second)
			}
		} else {
			logrus.Warnf("Invalid spec received for app %v in WaitForDestroy", ctx.App.Key)
		}
	}
	fmt.Println("KOKADBG: WaitForDestroy(): FAIL: END")
	return fmt.Errorf("KOKADBG: WaitForDestroy: Failed to get dead status in time, Status: %s\n", status)
}

func (nm *nomad) DeleteTasks(ctx *scheduler.Context) error {
	// TODO: Implement this method
	return nil
}

func (nm *nomad) GetVolumeParameters(ctx *scheduler.Context) (map[string]map[string]string, error) {
	// TODO: Implement this method
	fmt.Println("KOKADBG: GetVolumeParameters(): START")
	result := make(map[string]map[string]string)
	populateParamsFunc := func(volName string, volParams map[string]string) error {
		result[volName] = volParams
		return nil
	}

	if err := nm.volumeOperation(ctx, populateParamsFunc); err != nil {
		return nil, err
	}
	fmt.Println("KOKADBG: GetVolumeParameters(): END")
	return result, nil
}

func (nm *nomad) InspectVolumes(ctx *scheduler.Context, timeout, retryInterval time.Duration) error {
	fmt.Println("KOKADBG: InspectVolumes(): START")
	inspectDockerVolumeFunc := func(volName string, _ map[string]string) error {
		t := func() (interface{}, bool, error) {
			out, err := nm.dockerClient.VolumeInspect(context.Background(), volName)
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

	fmt.Println("KOKADBG: InspectVolumes(): END")
	return nm.volumeOperation(ctx, inspectDockerVolumeFunc)
}

func (nm *nomad) volumeOperation(ctx *scheduler.Context, f func(string, map[string]string) error) error {
	// Nomad does not have volume objects like Kubernetes. We get the volume information from
	// the app spec and get the options parsed from the respective volume driver
	fmt.Println("KOKADBG: volumeOperation(): START")
	volDriver, err := volume.Get(nm.volDriverName)
	if err != nil {
		return err
	}
loop:
	//fmt.Printf("KOKADBG: volumeOperation: after volume.Get(): volDriver: %s\n", volDriver)
	for _, spec := range ctx.App.SpecList {
		//fmt.Printf("KOKADBG: volumeOperation(): LOOP ctx.App.SpecList: spec: %v\n", spec)
		if obj, ok := spec.(*api.Job); ok {
			//fmt.Printf("KOKADBG: volumeOperation(): spec.(*api.Job): obj.Name: %v\n", obj.Name)
			for _, tskG := range obj.TaskGroups {
				//fmt.Printf("KOKADBG: volumeOperation(): LOOP obj.TaskGroups: tskG: %v\n", tskG)
				for _, tsk := range tskG.Tasks {
					//fmt.Printf("KOKADBG: volumeOperation(): LOOP tskG.Tasks: tsk: %v\n", tsk)
					for key, value := range tsk.Config {
						//fmt.Printf("KOKADBG: volumeOperation(): CFG RANGE KEY: %+v\n", key)
						//fmt.Printf("KOKADBG: volumeOperation(): CFG RANGE VALUE: %+v\n", value)

						// Find volumes
						if key == "volumes" {
							//fmt.Printf("KOKADBG: volumeOperation(): Key: %v, Value: %v\n", key, value)

							// Get volume List
							volList, ok := value.([]interface{})
							if !ok {
								fmt.Printf("KOKADBG: volumeOperation(): ERROR: volList Not OK! volList Content: %+v\n", volList)
							}
							//fmt.Printf("KOKADBG: volumeOperation: volList: %v\n", volList)

							// Get each volume from volume list
							for _, v := range volList {
								vol, ok := v.(string)
								if !ok {
									fmt.Printf("KOKADBG: volumeOperation(): ERROR: vol Not OK! vol Content: %+v\n", v)
								}
								//fmt.Printf("KOKADBG: volumeOperation(): VOL INSIDE LOOP: %v\n", vol)
								volName, volParams, err := volDriver.ExtractVolumeInfo(vol)
								if err != nil {
									return &scheduler.ErrFailedToGetVolumeParameters{
										App:   ctx.App,
										Cause: fmt.Sprintf("Failed to extract volume info: %v. Err: %v", vol, err),
									}
								}

								fmt.Printf("KOKADBG: volumeOperation(): voName: %v, volParams: %v\n", volName, volParams)
								if err := f(volName, volParams); err != nil {
									return err
								}
								//fmt.Printf("KOKADBG: VOL NAME: %v, VOL PARAM: %v\n", volName, volParams)
							}
							break loop
						}
					}
				}
			}
		} else {
			logrus.Warnf("Invalid spec received for app %v", ctx.App.Key)
		}
	}

	fmt.Println("KOKADBG: volumeOperation(): END")
	return nil
}

func (nm *nomad) DeleteVolumes(ctx *scheduler.Context) ([]*volume.Volume, error) {
	// TODO: Implement this method
	fmt.Println("KOKADBG: DeleteVolumes(): START")
	var vols []*volume.Volume

	deleteDockerVolumeFunc := func(volName string, _ map[string]string) error {
		vols = append(vols, &volume.Volume{Name: volName})
		t := func() (interface{}, bool, error) {
			return nil, true, nm.dockerClient.VolumeRemove(context.Background(), volName, false)
		}

		if _, err := task.DoRetryWithTimeout(t, 2*time.Minute, 10*time.Second); err != nil {
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
	fmt.Println("KOKADBG: DeleteVolumes(): END")
	return vols, nil
}

func (nm *nomad) GetVolumes(ctx *scheduler.Context) ([]*volume.Volume, error) {
	// TODO: Implement this method
	return []*volume.Volume{}, nil
}

func (nm *nomad) ResizeVolume(ctx *scheduler.Context) ([]*volume.Volume, error) {
	// TODO: Implement this method
	return []*volume.Volume{}, nil
}

func (nm *nomad) GetSnapshots(ctx *scheduler.Context) ([]*volume.Snapshot, error) {
	// TODO: Implement this method
	return []*volume.Snapshot{}, nil
}

func (nm *nomad) Describe(ctx *scheduler.Context) (string, error) {
	// TODO: Implement this method
	return "", nil
}

func (nm *nomad) ScaleApplication(ctx *scheduler.Context, scaleFactorMap map[string]int32) error {
	// TODO: Implement this method
	return nil
}

func (nm *nomad) GetScaledFactorMap(ctx *scheduler.Context) (map[string]int32, error) {
	// TODO: Implement this method
	var result map[string]int32
	return result, nil
}

func (nm *nomad) GetScaleFactorMap(ctx *scheduler.Context) (map[string]int32, error) {
	// TODO: Implement this method
	var result map[string]int32
	return result, nil
}

func (nm *nomad) StopSchedOnNode(n node.Node) error {
	// TODO: Implement this method
	return nil
}

func (nm *nomad) StartSchedOnNode(n node.Node) error {
	// TODO: Implement this method
	return nil
}

func (nm *nomad) RescanSpecs(specDir string) error {
	// TODO: Implement this method
	return nil
}

func init() {
	nm := &nomad{}
	scheduler.Register(SchedName, nm)
}
