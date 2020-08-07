package k8s

import (
	"bytes"
	"context"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	"github.com/sirupsen/logrus"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/strvals"
)

var settings *cli.EnvSettings

// HelmSchedule will install the application with helm
func (k *K8s) HelmSchedule(instanceID string, options scheduler.ScheduleOptions) ([]*scheduler.Context, *scheduler.HelmRepo, error) {
	var err error

	// For supporting multiple values to override, will read from a configmap
	// TODO: Need to change the following hardcoded value for sc too.
	args := map[string]string{
		// comma seperated values to set
		"set": "persistentStorage.storageClassName=central-sc,persistentStorage.enabled=true",
	}

	// Currently the installation supports for one app in one shot
	app, err := k.ChartFactory.Get(options.AppKeys[0])
	if err != nil {
		return nil, nil, err
	}

	var contexts []*scheduler.Context
	var yamlBuf bytes.Buffer
	repoInfo := app.ChartList[0]

	settings = cli.New()

	// Add helm repo
	err = k.RepoAdd(repoInfo)
	if err != nil {
		return nil, nil, err
	}
	// Update charts from the helm repo
	err = k.RepoUpdate()
	if err != nil {
		return nil, nil, err
	}

	// Install charts
	appNamespace := app.GetID(instanceID)
	// tempApp as type AppSpec for creating the namespace
	tempApp := &spec.AppSpec{
		Key:      app.Key,
		SpecList: nil,
		Enabled:  app.Enabled,
	}
	_, err = k.createNamespace(tempApp, appNamespace, options)
	if err != nil {
		return nil, nil, err
	}

	// Install the chart through helm
	repoInfo.Namespace = appNamespace
	manifest, err := k.InstallChart(repoInfo, args)
	if err != nil {
		return nil, nil, err
	}

	// Parse the manifest which is a yaml to get the k8s spec objects
	yamlBuf.WriteString(manifest)
	specObjects, err := k.ParseSpecsFromYamlBuf(&yamlBuf)
	if err != nil {
		return nil, nil, err
	}

	ctx := &scheduler.Context{
		UID: instanceID,
		App: &spec.AppSpec{
			Key:      app.Key,
			SpecList: specObjects,
			Enabled:  app.Enabled,
		},
		ScheduleOptions: options,
	}

	// Set the namespace for the specObjects
	err = k.UpdateTasksID(ctx, repoInfo.Namespace)
	if err != nil {
		return nil, nil, err
	}
	contexts = append(contexts, ctx)

	return contexts, repoInfo, nil
}

// ParseCharts parses the application spec file having helm repo info
func (k *K8s) ParseCharts(chartDir string) (*scheduler.HelmRepo, error) {
	fileList := make([]string, 0)
	if err := filepath.Walk(chartDir, func(path string, f os.FileInfo, err error) error {
		if f != nil && !f.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// There should be one custom file describing the helm repo info
	fileName := fileList[0]

	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	repoInfo := scheduler.HelmRepo{}
	err = yaml.Unmarshal(file, &repoInfo)
	if err != nil {
		return nil, err
	}

	return &repoInfo, nil
}

// RepoAdd adds repo with given name and url
func (k *K8s) RepoAdd(repoInfo *scheduler.HelmRepo) error {
	name := repoInfo.RepoName
	url := repoInfo.URL
	repoFile := settings.RepositoryConfig

	//Ensure the file directory exists as it is required for file locking
	err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Acquire a file lock for process synchronization
	fileLock := flock.New(strings.Replace(repoFile, filepath.Ext(repoFile), ".lock", 1))
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		defer fileLock.Unlock()
	}
	if err != nil {
		return err
	}

	b, err := ioutil.ReadFile(repoFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return err
	}

	if f.Has(name) {
		logrus.Errorf("repository name (%s) already exists\n", name)
		return nil
	}

	c := repo.Entry{
		Name: name,
		URL:  url,
	}

	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		return err
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		err := errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", url)
		return err
	}

	f.Update(&c)

	if err := f.WriteFile(repoFile, 0644); err != nil {
		return err
	}
	logrus.Printf("%q has been added to your repositories\n", name)
	return nil
}

// RepoUpdate updates charts for all helm repos
func (k *K8s) RepoUpdate() error {
	repoFile := settings.RepositoryConfig

	f, err := repo.LoadFile(repoFile)
	if os.IsNotExist(errors.Cause(err)) || len(f.Repositories) == 0 {
		return fmt.Errorf("No repositories found, need to add one before updating, err: %v", err)
	}
	var repos []*repo.ChartRepository
	for _, cfg := range f.Repositories {
		r, err := repo.NewChartRepository(cfg, getter.All(settings))
		if err != nil {
			return err
		}
		repos = append(repos, r)
	}

	logrus.Printf("Getting the latest from the chart repositories\n")
	var wg sync.WaitGroup
	for _, re := range repos {
		wg.Add(1)
		go func(re *repo.ChartRepository) {
			defer wg.Done()
			if _, err := re.DownloadIndexFile(); err != nil {
				logrus.Printf("Unable to get an update from the %q chart repository (%s):\n\t%s\n", re.Config.Name, re.Config.URL, err)
			} else {
				logrus.Printf("Successfully got an update from the %q chart repository\n", re.Config.Name)
			}
		}(re)
	}
	wg.Wait()
	logrus.Printf("RepoUpdate Completed successfully.\n")
	return nil
}

// InstallChart will install the helm chart
func (k *K8s) InstallChart(repoInfo *scheduler.HelmRepo, args map[string]string) (string, error) {
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), repoInfo.Namespace, os.Getenv("HELM_DRIVER"), debug); err != nil {
		return "", err
	}
	client := action.NewInstall(actionConfig)

	if client.Version == "" && client.Devel {
		client.Version = ">0.0.0-0"
	}
	client.ReleaseName = repoInfo.ReleaseName
	cp, err := client.ChartPathOptions.LocateChart(fmt.Sprintf("%s/%s", repoInfo.RepoName, repoInfo.ChartName), settings)
	if err != nil {
		return "", err
	}

	logrus.Printf("chart path: %s\n", cp)

	p := getter.All(settings)
	valueOpts := &values.Options{}
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return "", err
	}

	// Add args
	if err := strvals.ParseInto(args["set"], vals); err != nil {
		return "", errors.Wrap(err, "failed parsing --set data")
	}

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return "", err
	}

	validInstallableChart, err := isChartInstallable(chartRequested)
	if !validInstallableChart {
		return "", err
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		// If CheckDependencies returns an error, we have unfulfilled dependencies.
		// As of Helm 2.4.0, this is treated as a stopping condition:
		// https://github.com/helm/helm/issues/2209
		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:              os.Stdout,
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
				}
				if err := man.Update(); err != nil {
					return "", err
				}
			} else {
				return "", err
			}
		}
	}

	client.Namespace = repoInfo.Namespace
	release, err := client.Run(chartRequested, vals)
	if err != nil {
		return "", err
	}
	return release.Manifest, nil
}

func isChartInstallable(ch *chart.Chart) (bool, error) {
	switch ch.Metadata.Type {
	case "", "application":
		return true, nil
	}
	return false, errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}

// UnInstallHelmChart will uninstall the release
func (k *K8s) UnInstallHelmChart(repoInfo *scheduler.HelmRepo) error {
	var err error
	actionConfig := new(action.Configuration)
	if err = actionConfig.Init(settings.RESTClientGetter(), repoInfo.Namespace, os.Getenv("HELM_DRIVER"), debug); err != nil {
		return err
	}

	client := action.NewUninstall(actionConfig)
	_, err = client.Run(repoInfo.ReleaseName)
	if err != nil {
		return err
	}

	return nil
}

func debug(format string, v ...interface{}) {
	format = fmt.Sprintf(" %s\n", format)
	logrus.Printf(fmt.Sprintf(format, v...))
}
