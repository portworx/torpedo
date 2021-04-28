package openshift

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/blang/semver"
	openshiftv1 "github.com/openshift/api/config/v1"
	k8s "github.com/portworx/sched-ops/k8s/core"
	opnshift "github.com/portworx/sched-ops/k8s/openshift"
	"github.com/portworx/sched-ops/task"
	"github.com/portworx/torpedo/drivers/node"
	"github.com/portworx/torpedo/drivers/scheduler"
	kube "github.com/portworx/torpedo/drivers/scheduler/k8s"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

const (
	// SchedName is the name of the kubernetes scheduler driver implementation
	SchedName = "openshift"
	// SystemdSchedServiceName is the name of the system service responsible for scheduling
	SystemdSchedServiceName = "atomic-openshift-node"
	// OpenshiftMirror is the mirror we use do download ocp client
	OpenshiftMirror = "https://mirror.openshift.com/pub/openshift-v4/clients/ocp"
)

var (
	k8sOpenshift = opnshift.Instance()
	k8sCore      = k8s.Instance()
	versionReg   = regexp.MustCompile(`^(stable|candidate|fast)(-\d\.\d)?$`)
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

		specObjects, err := k.CreateSpecObjects(app, appNamespace, options)
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

func (k *openshift) SaveSchedulerLogsToFile(n node.Node, location string) error {
	driver, _ := node.Get(k.K8s.NodeDriverName)
	cmd := fmt.Sprintf("journalctl -lu %s* > %s/kubelet.log", SystemdSchedServiceName, location)
	_, err := driver.RunCommand(n, cmd, node.ConnectionOpts{
		Timeout:         kube.DefaultTimeout,
		TimeBeforeRetry: kube.DefaultRetryInterval,
		Sudo:            true,
	})
	return err
}

func (k *openshift) updateSecurityContextConstraints(namespace string) error {
	// Get privileged context
	context, err := k8sOpenshift.GetSecurityContextConstraints("privileged")
	if err != nil {
		return err
	}

	// Add user and namespace to context
	context.Users = append(context.Users, "system:serviceaccount:"+namespace+":default")

	// Update context
	_, err = k8sOpenshift.UpdateSecurityContextConstraints(context)
	if err != nil {
		return err
	}

	return nil
}

func (k *openshift) UpgradeScheduler(version string) error {
	var err error

	if err = downloadOCP4Client(version); err != nil {
		return err
	}

	clientVersion := ""
	if clientVersion, err = getClientVersion(); err != nil {
		return err
	}

	upgradeVersion := version
	if versionReg.MatchString(version) {
		upgradeVersion = clientVersion
	}

	if err := selectChannel(version); err != nil {
		return err
	}

	if err := startUpgrade(clientVersion, upgradeVersion); err != nil {
		return nil
	}

	if err := waitUpgradeCompletion(clientVersion); err != nil {
		return err
	}

	logrus.Info("Waiting for all the nodes to become ready...")
	if err := waitNodesToBeReady(); err != nil {
		return err
	}

	logrus.Infof("Cluster is now %s", upgradeVersion)
	return nil
}

func getClientVersion() (string, error) {
	var err error
	var output interface{}

	t := func() (interface{}, bool, error) {
		var output []byte
		cmd := "oc version --client -o json|jq -r .releaseClientVersion"
		if output, err = exec.Command("sh", "-c", cmd).CombinedOutput(); err != nil {
			return "", true, fmt.Errorf("failed to get client version. cause: %v", err)
		}
		clientVersion := strings.TrimSpace(string(output))
		clientVersion = strings.Trim(clientVersion, "\"")
		clientVersion = strings.Trim(clientVersion, "'")
		return clientVersion, false, nil
	}
	if output, err = task.DoRetryWithTimeout(t, 1*time.Minute, 5*time.Second); err != nil {
		return "", err
	}
	return output.(string), nil
}

func selectChannel(version string) error {
	var output []byte
	var err error

	channel := ""
	if channel, err = getChannel(version); err != nil {
		return err
	}
	logrus.Infof("Selected channel: %s", channel)

	patch := `
spec:
  channel: %s
`
	t := func() (interface{}, bool, error) {
		args := []string{"patch", "clusterversion", "version", "--type=merge", "--patch", fmt.Sprintf(patch, channel)}
		if output, err = exec.Command("oc", args...).CombinedOutput(); err != nil {
			return nil, true, fmt.Errorf("failed to select channel due to %s. cause: %v", string(output), err)
		}
		logrus.Info(string(output))
		return nil, false, nil
	}
	_, err = task.DoRetryWithTimeout(t, 1*time.Minute, 5*time.Second)
	return err
}

func startUpgrade(clientVersion, upgradeVersion string) error {
	var output []byte
	var err error

	args := []string{"adm", "upgrade", fmt.Sprintf("--to=%s", upgradeVersion)}
	if output, err = exec.Command("oc", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start upgrade due to %s. cause: %v", string(output), err)
	}
	logrus.Infof("Upgrade started: %s", output)

	t := func() (interface{}, bool, error) {
		clusterVersion, err := k8sOpenshift.GetClusterVersion("version")
		if err != nil {
			return nil, true, fmt.Errorf("failed to get cluster version. cause: %v", err)
		}

		desiredVersion := clusterVersion.Status.Desired.Version
		if desiredVersion != clientVersion {
			return nil, true, fmt.Errorf("version mismatch. expected: %s but got %s", upgradeVersion, desiredVersion)
		}
		return nil, false, nil
	}

	_, err = task.DoRetryWithTimeout(t, 5*time.Minute, 15*time.Second)
	return err
}

func waitUpgradeCompletion(upgradeVersion string) error {
	var err error

	t := func() (interface{}, bool, error) {
		clusterVersion, err := k8sOpenshift.GetClusterVersion("version")
		if err != nil {
			return nil, true, fmt.Errorf("failed to get cluster version. cause: %v", err)
		}

		for _, status := range clusterVersion.Status.Conditions {
			if status.Type == openshiftv1.OperatorProgressing && status.Status == openshiftv1.ConditionTrue {
				return nil, true, fmt.Errorf("cluster not upgraded yet. cause: %s", status.Message)
			} else if status.Type == openshiftv1.OperatorProgressing && status.Status == openshiftv1.ConditionFalse {
				break
			}
		}

		for _, history := range clusterVersion.Status.History {
			if history.Version == upgradeVersion && history.State != openshiftv1.CompletedUpdate {
				return nil, true, fmt.Errorf("cluster not upgraded yet. expected: %v got: %v", openshiftv1.CompletedUpdate, history.State)
			} else if history.Version == upgradeVersion && history.State == openshiftv1.CompletedUpdate {
				break
			}
		}
		return nil, false, nil
	}

	_, err = task.DoRetryWithTimeout(t, 2*time.Hour, 15*time.Second)
	return err
}

// waitNodesToBeReady waits for all nodes to become Ready and using the same k8s version
func waitNodesToBeReady() error {
	var err error

	t := func() (interface{}, bool, error) {
		var count int
		var k8sVersions = make(map[string]string)
		var versionSet = make(map[string]bool)

		nodeList, err := k8sCore.GetNodes()
		if err != nil {
			return nil, true, fmt.Errorf("failed to get nodes. cause: %v", err)
		}

		for _, k8sNode := range nodeList.Items {
			for _, status := range k8sNode.Status.Conditions {
				if status.Type == corev1.NodeReady && status.Status == corev1.ConditionTrue {
					count++
					kubeletVersion := k8sNode.Status.NodeInfo.KubeletVersion
					k8sVersions[k8sNode.Name] = kubeletVersion
					versionSet[kubeletVersion] = true
					break
				}
			}
		}

		totalNodes := len(nodeList.Items)
		if count < totalNodes {
			return nil, true, fmt.Errorf("nodes not ready. expected %d but got %d", totalNodes, count)
		}

		if len(versionSet) > 1 {
			return nil, true, fmt.Errorf("nodes are not in the same version.\n%v", k8sVersions)
		}
		return nil, false, nil
	}

	_, err = task.DoRetryWithTimeout(t, 30*time.Minute, 15*time.Second)
	return err
}

func getChannel(version string) (string, error) {
	if versionReg.MatchString(version) {
		return version, nil
	}

	ver, err := semver.Make(version)
	if err != nil {
		return "", fmt.Errorf("failed to parse version: %s. cause: %v", version, err)
	}
	channel := fmt.Sprintf("candidate-%d.%d", ver.Major, ver.Minor)
	url := fmt.Sprintf("curl -sL %s/%s | grep %s", OpenshiftMirror, channel, version)
	if output, _ := exec.Command("sh", "-c", url).CombinedOutput(); len(output) > 0 {
		return channel, nil
	}
	return fmt.Sprintf("stable-%d.%d", ver.Major, ver.Minor), nil
}

func downloadOCP4Client(ocpVersion string) error {
	var clientName = ""
	var downloadURL = ""
	var output []byte

	if ocpVersion == "" {
		ocpVersion = "latest"
	}

	logrus.Info("Downloading OCP 4.X client. May take some time...")
	if versionReg.MatchString(ocpVersion) {
		downloadURL = fmt.Sprintf("%s/%s/openshift-client-linux.tar.gz", OpenshiftMirror,
			ocpVersion)
		clientName = "openshift-client-linux.tar.gz"
	} else {
		downloadURL = fmt.Sprintf("%s/%s/openshift-client-linux-%s.tar.gz", OpenshiftMirror,
			ocpVersion, ocpVersion)
		clientName = fmt.Sprintf("openshift-client-linux-%s.tar.gz", ocpVersion)
	}

	stdout, err := exec.Command("curl", "-o", clientName, downloadURL).CombinedOutput()
	if err != nil {
		logrus.Errorf("Error while downloading OpenShift 4.X client from %s, error %v", downloadURL, err)
		logrus.Error(string(stdout))
		return err
	}

	logrus.Infof("Openshift client %s downloaded successfully.", clientName)

	stdout, err = exec.Command("tar", "-xvf", clientName).CombinedOutput()
	if err != nil {
		logrus.Errorf("Error extracting %s, error %v", clientName, err)
		logrus.Error(string(stdout))
		return err
	}

	logrus.Infof("Extracted %s successfully.", clientName)

	stdout, err = exec.Command("cp", "./oc", "/usr/local/bin").CombinedOutput()
	if err != nil {
		logrus.Errorf("Error copying %s, error %v", clientName, err)
		logrus.Error(string(stdout))
		return err
	}

	if output, err = exec.Command("oc", "version").CombinedOutput(); err != nil {
		logrus.Errorf("Error getting oc version, error %v", err)
		logrus.Error(string(stdout))
		return err
	}
	logrus.Info(string(output))
	return nil
}

func init() {
	k := &openshift{}
	scheduler.Register(SchedName, k)
}
