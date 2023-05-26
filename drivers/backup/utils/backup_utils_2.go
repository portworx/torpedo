package utils

import (
	"context"
	"fmt"
	"github.com/portworx/torpedo/pkg/log"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	api "github.com/portworx/px-backup-api/pkg/apis/v1"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/portworx/torpedo/drivers"
	"github.com/portworx/torpedo/drivers/backup/portworx"
)

type VariablePair struct {
	Name  string
	Value interface{}
}

const (
	DefaultConfigMapName       = "kubeconfigs"
	DefaultKubeconfigDirectory = "/tmp"
	DefaultConfigMapNamespace  = "default"
)

var (
	DefaultCloudProviders = []string{"aws"}
)

var (
	GlobalAWSBucketName         string
	GlobalAzureBucketName       string
	GlobalGCPBucketName         string
	GlobalAWSLockedBucketName   string
	GlobalAzureLockedBucketName string
	GlobalGCPLockedBucketName   string
)

type ControllerResult struct {
	Error      error
	LogMessage string
}

const (
	DefaultBackupLocationAdditionTimeout = 5 * time.Minute
	DefaultBackupLocationDeletionTimeout = 5 * time.Minute
	DefaultClusterAdditionTimeout        = 5 * time.Minute
	DefaultBackupCreationTimeout         = 40 * time.Minute
	DefaultBackupDeletionTimeout         = 20 * time.Minute
	DefaultRestoreCreationTimeout        = 40 * time.Minute
	DefaultRestoreDeletionTimeout        = 20 * time.Minute
)

const (
	DefaultBackupLocationAdditionRetryTime = 10 * time.Second
	DefaultBackupLocationDeletionRetryTime = 10 * time.Second
	DefaultClusterAdditionRetryTime        = 10 * time.Second
	DefaultBackupCreationRetryTime         = 30 * time.Second
	DefaultBackupDeletionRetryTime         = 30 * time.Second
	DefaultRestoreCreationRetryTime        = 30 * time.Second
	DefaultRestoreDeletionRetryTime        = 30 * time.Second
)

// var (
// 	// User should keep updating preRuleApps, postRuleApps
// 	preRuleApps  = []string{"cassandra", "postgres"}
// 	postRuleApps = []string{"cassandra"}
// )

func GetKubeconfigsFromEnv() []string {
	kubeconfigs := os.Getenv("KUBECONFIGS")
	kubeconfigsList := strings.Split(kubeconfigs, ",")
	return kubeconfigsList
}

func GetClusterConfigPath(kubeconfigKey, configMapName string, configMapNamespace string) (string, error) {
	cm, err := core.Instance().GetConfigMap(configMapName, configMapNamespace)
	if err != nil {
		return "", fmt.Errorf("error reading config map: %v", err)
	}

	kubeconfig, ok := cm.Data[kubeconfigKey]
	if !ok {
		return "", fmt.Errorf("kubeconfig for %s not found in the config map %s", kubeconfigKey, configMapName)
	}

	filePath := fmt.Sprintf("%s/%s", DefaultKubeconfigDirectory, kubeconfigKey)

	err = ioutil.WriteFile(filePath, []byte(kubeconfig), 0644)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func GetGlobalBucketName(provider string) string {
	switch provider {
	case drivers.ProviderAws:
		return GlobalAWSBucketName
	case drivers.ProviderAzure:
		return GlobalAzureBucketName
	case drivers.ProviderGke:
		return GlobalGCPBucketName
	default:
		return GlobalAWSBucketName
	}
}

func GetPreRulesInfoFromAppParameters(appName string) (*api.RulesInfo, error) {
	preActionList, ok := portworx.AppParameters[appName]["pre"]["pre_action_list"]
	if !ok || len(preActionList) == 0 {
		return nil, fmt.Errorf("pre actions for app [%s] do not exist in AppParameters %v", appName, portworx.AppParameters)
	}
	preRules := portworx.AppParameters[appName]["pre"]
	preActions := preRules["pre_action_list"]
	podSelectors := preRules["pod_selector_list"]
	backgrounds := preRules["background"]
	runInSinglePods := preRules["runInSinglePod"]
	containers := preRules["container"]
	rulesInfo := &api.RulesInfo{}
	for i := 0; i < len(preActions); i++ {
		podSelectorsList := strings.Split(podSelectors[i], "=")
		podSelectorsMap := map[string]string{podSelectorsList[0]: podSelectorsList[1]}
		backgroundVal, _ := strconv.ParseBool(backgrounds[i])
		runInSinglePodVal, _ := strconv.ParseBool(runInSinglePods[i])
		ruleAction := api.RulesInfo_Action{
			Background:     backgroundVal,
			RunInSinglePod: runInSinglePodVal,
			Value:          preActions[i],
		}
		ruleItem := api.RulesInfo_RuleItem{
			PodSelector: podSelectorsMap,
			Actions:     []*api.RulesInfo_Action{&ruleAction},
			Container:   containers[i],
		}
		rulesInfo.Rules = append(rulesInfo.Rules, &ruleItem)
	}
	return rulesInfo, nil
}

func GetPostRulesInfoFromAppParameters(appName string) (*api.RulesInfo, error) {
	postActionList, ok := portworx.AppParameters[appName]["post"]["post_action_list"]
	if !ok || len(postActionList) == 0 {
		return nil, fmt.Errorf("post actions for app [%s] do not exist in AppParameters %v", appName, portworx.AppParameters)
	}
	postRules := portworx.AppParameters[appName]["post"]
	postActions := postRules["post_action_list"]
	podSelectors := postRules["pod_selector_list"]
	backgrounds := postRules["background"]
	runInSinglePods := postRules["runInSinglePod"]
	containers := postRules["container"]
	rulesInfo := &api.RulesInfo{}
	for i := 0; i < len(postActions); i++ {
		podSelectorsList := strings.Split(podSelectors[i], "=")
		podSelectorsMap := map[string]string{podSelectorsList[0]: podSelectorsList[1]}
		backgroundVal, _ := strconv.ParseBool(backgrounds[i])
		runInSinglePodVal, _ := strconv.ParseBool(runInSinglePods[i])
		ruleAction := api.RulesInfo_Action{
			Background:     backgroundVal,
			RunInSinglePod: runInSinglePodVal,
			Value:          postActions[i],
		}
		ruleItem := api.RulesInfo_RuleItem{
			PodSelector: podSelectorsMap,
			Actions:     []*api.RulesInfo_Action{&ruleAction},
			Container:   containers[i],
		}
		rulesInfo.Rules = append(rulesInfo.Rules, &ruleItem)
	}
	return rulesInfo, nil
}

func DoRetryWithTimeout(fn interface{}, retryDuration, retryInterval time.Duration, shouldRetry func(interface{}) bool) (interface{}, error) {
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		return nil, fmt.Errorf("%v is not a function", fn)
	}
	ctx, cancel := context.WithTimeout(context.Background(), retryDuration)
	defer cancel()
	resultChan := make(chan interface{})
	errChan := make(chan error)
	go func() {
		for {
			select {
			case <-ctx.Done():
				if ctx.Err() != nil {
					errChan <- ctx.Err()
				}
				return
			default:
				resultValues := fnValue.Call(nil)
				if len(resultValues) == 0 {
					errChan <- fmt.Errorf("function returned no values")
					return
				}
				result := resultValues[0].Interface()
				var err error
				if len(resultValues) > 1 {
					err, _ = resultValues[1].Interface().(error)
				}
				if err == nil && !shouldRetry(result) {
					resultChan <- result
					return
				}
				if err != nil && !shouldRetry(err) {
					errChan <- err
					return
				}
				log.Infof("Next retry in: %v", retryInterval)
				time.Sleep(retryInterval)
			}
		}
	}()
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		if err == context.DeadlineExceeded {
			return nil, err
		}
		return nil, err
	}
}

func ValidateNonEmptyStrings(pairs ...VariablePair) error {
	var fieldNames []string
	for _, pair := range pairs {
		v := reflect.ValueOf(pair.Value)
		t := v.Type()
		if v.Kind() == reflect.Ptr && v.Type().Elem().Kind() == reflect.String {
			if v.IsNil() || v.Elem().String() == "" {
				fieldNames = append(fieldNames, fmt.Sprintf("'%s' (type: %s)", pair.Name, t))
			}
		} else if v.Kind() == reflect.String {
			if v.String() == "" {
				fieldNames = append(fieldNames, fmt.Sprintf("'%s' (type: %s)", pair.Name, t))
			}
		}
	}
	if len(fieldNames) > 0 {
		return fmt.Errorf("string variables %s are empty or nil", fieldNames)
	}
	return nil
}

func GetProvidersFromEnv() []string {
	providers := os.Getenv("PROVIDERS")
	providersList := strings.Split(providers, ",")
	return providersList
}
