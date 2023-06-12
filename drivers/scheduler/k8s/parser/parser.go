package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/portworx/torpedo/pkg/log"

	yaml2 "gopkg.in/yaml.v2"

	snapv1 "github.com/kubernetes-incubator/external-storage/snapshot/pkg/apis/crd/v1"
	apapi "github.com/libopenstorage/autopilot-api/pkg/apis/autopilot/v1alpha1"
	storkapi "github.com/libopenstorage/stork/pkg/apis/stork/v1alpha1"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsapi "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	storageapi "k8s.io/api/storage/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/kubectl/pkg/scheme"
)

// CustomResourceObjectYAML	Used as spec object for all CRs
type CustomResourceObjectYAML struct {
	Path      string
	Namespace string // Namespace will only be assigned DURING creation
	Name      string
}

type K8sParser struct {
	CustomConfig map[string]scheduler.AppConfig
}

// ParseSpecs parses the application spec file
func (p *K8sParser) ParseSpecs(specDir, storageProvisioner string) ([]interface{}, error) {
	return ParseSpecs(specDir, storageProvisioner, p.CustomConfig)
}

// ParseSpecs parses the application spec file
func ParseSpecs(specDir, volumeDriver string, customAppConfig map[string]scheduler.AppConfig) ([]interface{}, error) {
	log.Debugf("ParseSpecs k.CustomConfig = %v", customAppConfig)
	fileList := make([]string, 0)
	if err := filepath.Walk(specDir, func(path string, f os.FileInfo, err error) error {
		if f != nil && !f.IsDir() {
			if IsValidProvider(path, volumeDriver) {
				log.Debugf("	add filepath: %s", path)
				fileList = append(fileList, path)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	log.Debugf("fileList: %v", fileList)
	var specs []interface{}

	splitPath := strings.Split(specDir, "/")
	appName := splitPath[len(splitPath)-1]

	for _, fileName := range fileList {
		isHelmChart, err := IsAppHelmChartType(fileName)
		if err != nil {
			return nil, err
		}

		splitPath := strings.Split(fileName, "/")
		if strings.HasPrefix(splitPath[len(splitPath)-1], "cr-") {
			// TODO: process with templates
			specObj := &CustomResourceObjectYAML{
				Path: fileName,
			}
			specs = append(specs, specObj)
			log.Warnf("custom res: %v", specObj) //TODO: remove
		} else if !isHelmChart {
			file, err := ioutil.ReadFile(fileName)
			if err != nil {
				return nil, err
			}

			var customConfig scheduler.AppConfig
			var ok bool

			if customConfig, ok = customAppConfig[appName]; !ok {
				customConfig = scheduler.AppConfig{}
			} else {
				log.Infof("customConfig[%v] = %v", appName, customConfig)
			}
			var funcs = template.FuncMap{
				"Iterate": func(count int) []int {
					var i int
					var Items []int
					for i = 1; i <= (count); i++ {
						Items = append(Items, i)
					}
					return Items
				},
				"array": func(arr []string) string {
					string := "[\""
					for i, val := range arr {
						if i != 0 {
							string += "\", \""
						}
						string += val
					}
					return string + "\"]"
				},
			}

			tmpl, err := template.New("customConfig").Funcs(funcs).Parse(string(file))
			if err != nil {
				return nil, err
			}
			var processedFile bytes.Buffer
			err = tmpl.Execute(&processedFile, customConfig)
			if err != nil {
				return nil, err
			}

			reader := bufio.NewReader(&processedFile)
			specReader := yaml.NewYAMLReader(reader)

			for {
				specContents, err := specReader.Read()
				if err == io.EOF {
					break
				}
				if len(bytes.TrimSpace(specContents)) > 0 {
					obj, err := DecodeSpec(specContents)
					if err != nil {
						log.Warnf("Error decoding spec from %v: %v", fileName, err)
						return nil, err
					}

					specObj, err := ValidateSpec(obj)
					if err != nil {
						log.Warnf("Error parsing spec from %v: %v", fileName, err)
						return nil, err
					}
					SubstituteImageWithInternalRegistry(specObj)
					specs = append(specs, specObj)
				}
			}
		} else {
			repoInfo, err := ParseCharts(fileName)
			if err != nil {
				return nil, err
			}
			specs = append(specs, repoInfo)
		}
	}
	return specs, nil
}

func IsValidProvider(specPath, volumeDriver string) bool {
	// reject all volume drivers except for volumeDriver
	for _, driver := range volume.GetVolumeDrivers() {
		if driver != volumeDriver && strings.Contains(specPath, "/"+driver+"/") {
			return false
		}
	}
	// Get the rest of specs
	return true
}

func SubstituteImageWithInternalRegistry(spec interface{}) {
	internalDockerRegistry := os.Getenv("INTERNAL_DOCKER_REGISTRY")
	if internalDockerRegistry != "" {
		if obj, ok := spec.(*appsapi.DaemonSet); ok {
			ModifyImageInContainers(obj.Spec.Template.Spec.InitContainers, internalDockerRegistry)
			ModifyImageInContainers(obj.Spec.Template.Spec.Containers, internalDockerRegistry)
		}
		if obj, ok := spec.(*appsapi.Deployment); ok {
			ModifyImageInContainers(obj.Spec.Template.Spec.InitContainers, internalDockerRegistry)
			ModifyImageInContainers(obj.Spec.Template.Spec.Containers, internalDockerRegistry)
		}
		if obj, ok := spec.(*appsapi.StatefulSet); ok {
			ModifyImageInContainers(obj.Spec.Template.Spec.InitContainers, internalDockerRegistry)
			ModifyImageInContainers(obj.Spec.Template.Spec.Containers, internalDockerRegistry)
		}
		if obj, ok := spec.(*batchv1.Job); ok {
			ModifyImageInContainers(obj.Spec.Template.Spec.InitContainers, internalDockerRegistry)
			ModifyImageInContainers(obj.Spec.Template.Spec.Containers, internalDockerRegistry)
		}
		if obj, ok := spec.(*corev1.Pod); ok {
			ModifyImageInContainers(obj.Spec.InitContainers, internalDockerRegistry)
			ModifyImageInContainers(obj.Spec.Containers, internalDockerRegistry)
		}
	}
}

// ParseSpecsFromYamlBuf parses the yaml buf content
func ParseSpecsFromYamlBuf(yamlBuf *bytes.Buffer) ([]interface{}, error) {

	var specs []interface{}

	reader := bufio.NewReader(yamlBuf)
	specReader := yaml.NewYAMLReader(reader)

	for {
		specContents, err := specReader.Read()
		if err == io.EOF {
			break
		}
		if len(bytes.TrimSpace(specContents)) > 0 {
			obj, err := DecodeSpec(specContents)
			if err != nil {
				log.Warnf("Error decoding spec: %v", err)
				return nil, err
			}

			specObj, err := ValidateSpec(obj)
			if err != nil {
				log.Warnf("Error validating spec: %v", err)
				return nil, err
			}

			SubstituteImageWithInternalRegistry(specObj)
			specs = append(specs, specObj)
		}
	}

	return specs, nil
}

// ParseCharts parses the application spec file having helm repo info
func ParseCharts(fileName string) (interface{}, error) {
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

func ModifyImageInContainers(containers []corev1.Container, imageName string) {
	if containers != nil {
		for idx := range containers {
			containers[idx].Image = fmt.Sprintf("%s/%s", imageName, containers[idx].Image)
		}
	}
}

// IsAppHelmChartType will return true if the specDir has only one file and it has helm repo infos
// else will return false
func IsAppHelmChartType(fileName string) (bool, error) {

	// Parse the files and check for certain keys for helmRepo info

	log.Debugf("Reading file: %s", fileName)
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		return false, err
	}

	repoInfo := scheduler.HelmRepo{}
	err = yaml2.Unmarshal(file, &repoInfo)
	if err != nil {
		// Ignoring if unmarshalling fails as some app specs (like fio) failed to unmarshall
		log.Errorf("Ignoring the yaml unmarshalling failure , err: %v", err)
		return false, nil
	}

	if repoInfo.RepoName != "" && repoInfo.ChartName != "" && repoInfo.ReleaseName != "" {
		// If the yaml file with helmRepo info for the app is found, exit here.
		log.Infof("Helm chart was found in file: [%s]", fileName)
		return true, nil
	}

	return false, nil

}

func DecodeSpec(specContents []byte) (runtime.Object, error) {
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(specContents), nil, nil)
	if err != nil {
		schemeObj := runtime.NewScheme()
		if err := snapv1.AddToScheme(schemeObj); err != nil {
			return nil, err
		}

		if err := storkapi.AddToScheme(schemeObj); err != nil {
			return nil, err
		}

		if err := apapi.AddToScheme(schemeObj); err != nil {
			return nil, err
		}

		if err := monitoringv1.AddToScheme(schemeObj); err != nil {
			return nil, err
		}

		if err := apiextensionsv1beta1.AddToScheme(schemeObj); err != nil {
			return nil, err
		}

		if err := apiextensionsv1.AddToScheme(schemeObj); err != nil {
			return nil, err
		}

		if err := storageapi.AddToScheme(schemeObj); err != nil {
			return nil, err
		}

		codecs := serializer.NewCodecFactory(schemeObj)
		obj, _, err = codecs.UniversalDeserializer().Decode([]byte(specContents), nil, nil)
		if err != nil {
			return nil, err
		}
	}
	return obj, nil
}

func ValidateSpec(in interface{}) (interface{}, error) {
	if specObj, ok := in.(*appsapi.Deployment); ok {
		return specObj, nil
	} else if specObj, ok := in.(*appsapi.StatefulSet); ok {
		return specObj, nil
	} else if specObj, ok := in.(*appsapi.DaemonSet); ok {
		return specObj, nil
	} else if specObj, ok := in.(*corev1.Service); ok {
		return specObj, nil
	} else if specObj, ok := in.(*corev1.PersistentVolumeClaim); ok {
		return specObj, nil
	} else if specObj, ok := in.(*storageapi.StorageClass); ok {
		return specObj, nil
	} else if specObj, ok := in.(*snapv1.VolumeSnapshot); ok {
		return specObj, nil
	} else if specObj, ok := in.(*storkapi.GroupVolumeSnapshot); ok {
		return specObj, nil
	} else if specObj, ok := in.(*corev1.Secret); ok {
		return specObj, nil
	} else if specObj, ok := in.(*corev1.ConfigMap); ok {
		return specObj, nil
	} else if specObj, ok := in.(*storkapi.Rule); ok {
		return specObj, nil
	} else if specObj, ok := in.(*corev1.Pod); ok {
		return specObj, nil
	} else if specObj, ok := in.(*storkapi.ClusterPair); ok {
		return specObj, nil
	} else if specObj, ok := in.(*storkapi.Migration); ok {
		return specObj, nil
	} else if specObj, ok := in.(*storkapi.MigrationSchedule); ok {
		return specObj, nil
	} else if specObj, ok := in.(*storkapi.BackupLocation); ok {
		return specObj, nil
	} else if specObj, ok := in.(*storkapi.ApplicationBackup); ok {
		return specObj, nil
	} else if specObj, ok := in.(*storkapi.SchedulePolicy); ok {
		return specObj, nil
	} else if specObj, ok := in.(*storkapi.ApplicationRestore); ok {
		return specObj, nil
	} else if specObj, ok := in.(*storkapi.ApplicationClone); ok {
		return specObj, nil
	} else if specObj, ok := in.(*storkapi.VolumeSnapshotRestore); ok {
		return specObj, nil
	} else if specObj, ok := in.(*apapi.AutopilotRule); ok {
		return specObj, nil
	} else if specObj, ok := in.(*corev1.ServiceAccount); ok {
		return specObj, nil
	} else if specObj, ok := in.(*rbacv1.ClusterRole); ok {
		return specObj, nil
	} else if specObj, ok := in.(*rbacv1.ClusterRoleBinding); ok {
		return specObj, nil
	} else if specObj, ok := in.(*rbacv1.Role); ok {
		return specObj, nil
	} else if specObj, ok := in.(*rbacv1.RoleBinding); ok {
		return specObj, nil
	} else if specObj, ok := in.(*batchv1beta1.CronJob); ok {
		return specObj, nil
	} else if specObj, ok := in.(*batchv1.Job); ok {
		return specObj, nil
	} else if specObj, ok := in.(*corev1.LimitRange); ok {
		return specObj, nil
	} else if specObj, ok := in.(*networkingv1beta1.Ingress); ok {
		return specObj, nil
	} else if specObj, ok := in.(*monitoringv1.Prometheus); ok {
		return specObj, nil
	} else if specObj, ok := in.(*monitoringv1.PrometheusRule); ok {
		return specObj, nil
	} else if specObj, ok := in.(*monitoringv1.ServiceMonitor); ok {
		return specObj, nil
	} else if specObj, ok := in.(*corev1.Namespace); ok {
		return specObj, nil
	} else if specObj, ok := in.(*apiextensionsv1beta1.CustomResourceDefinition); ok {
		return specObj, nil
	} else if specObj, ok := in.(*apiextensionsv1.CustomResourceDefinition); ok {
		return specObj, nil
	} else if specObj, ok := in.(*policyv1beta1.PodDisruptionBudget); ok {
		return specObj, nil
	} else if specObj, ok := in.(*netv1.NetworkPolicy); ok {
		return specObj, nil
	} else if specObj, ok := in.(*corev1.Endpoints); ok {
		return specObj, nil
	} else if specObj, ok := in.(*storkapi.ResourceTransformation); ok {
		return specObj, nil
	} else if specObj, ok := in.(*admissionregistrationv1.ValidatingWebhookConfiguration); ok {
		return specObj, nil
	} else if specObj, ok := in.(*admissionregistrationv1.ValidatingWebhookConfigurationList); ok {
		return specObj, nil
	}

	return nil, fmt.Errorf("unsupported object: %v", reflect.TypeOf(in))
}
