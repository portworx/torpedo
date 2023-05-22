package spec

import (
	"fmt"
	snapv1 "github.com/kubernetes-incubator/external-storage/snapshot/pkg/apis/crd/v1"
	apapi "github.com/libopenstorage/autopilot-api/pkg/apis/autopilot/v1alpha1"
	storkapi "github.com/libopenstorage/stork/pkg/apis/stork/v1alpha1"
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
)

// Parser provides operations for parsing application specs
type Parser interface {
	ParseSpecs(specDir, storageProvisioner string) ([]interface{}, error)
}

// AppSpec defines a k8s application specification
type AppSpec struct {
	// Key is used by applications to register to the factory
	Key string
	// List of k8s spec objects
	SpecList []interface{}
	// Enabled indicates if the application is enabled in the factory
	Enabled bool
}

// GetID returns the unique ID for the app specs
func (in *AppSpec) GetID(instanceID string) string {
	return fmt.Sprintf("%s-%s", in.Key, instanceID)
}

// DeepCopy Creates a copy of the AppSpec
func (in *AppSpec) DeepCopy() *AppSpec {
	if in == nil {
		return nil
	}
	out := new(AppSpec)
	out.Key = in.Key
	out.Enabled = in.Enabled
	out.SpecList = make([]interface{}, len(in.SpecList))
	for i, spec := range in.SpecList {
		switch v := spec.(type) {
		case *appsapi.Deployment:
			out.SpecList[i] = v.DeepCopy()
		case *appsapi.StatefulSet:
			out.SpecList[i] = v.DeepCopy()
		case *appsapi.DaemonSet:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.Service:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.PersistentVolumeClaim:
			out.SpecList[i] = v.DeepCopy()
		case *storageapi.StorageClass:
			out.SpecList[i] = v.DeepCopy()
		case *snapv1.VolumeSnapshot:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.GroupVolumeSnapshot:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.Secret:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.ConfigMap:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.Rule:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.Pod:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.ClusterPair:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.Migration:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.MigrationSchedule:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.BackupLocation:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.ApplicationBackup:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.SchedulePolicy:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.ApplicationRestore:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.ApplicationClone:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.VolumeSnapshotRestore:
			out.SpecList[i] = v.DeepCopy()
		case *apapi.AutopilotRule:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.ServiceAccount:
			out.SpecList[i] = v.DeepCopy()
		case *rbacv1.ClusterRole:
			out.SpecList[i] = v.DeepCopy()
		case *rbacv1.ClusterRoleBinding:
			out.SpecList[i] = v.DeepCopy()
		case *rbacv1.Role:
			out.SpecList[i] = v.DeepCopy()
		case *rbacv1.RoleBinding:
			out.SpecList[i] = v.DeepCopy()
		case *batchv1beta1.CronJob:
			out.SpecList[i] = v.DeepCopy()
		case *batchv1.Job:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.LimitRange:
			out.SpecList[i] = v.DeepCopy()
		case *networkingv1beta1.Ingress:
			out.SpecList[i] = v.DeepCopy()
		case *monitoringv1.Prometheus:
			out.SpecList[i] = v.DeepCopy()
		case *monitoringv1.PrometheusRule:
			out.SpecList[i] = v.DeepCopy()
		case *monitoringv1.ServiceMonitor:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.Namespace:
			out.SpecList[i] = v.DeepCopy()
		case *apiextensionsv1beta1.CustomResourceDefinition:
			out.SpecList[i] = v.DeepCopy()
		case *apiextensionsv1.CustomResourceDefinition:
			out.SpecList[i] = v.DeepCopy()
		case *policyv1beta1.PodDisruptionBudget:
			out.SpecList[i] = v.DeepCopy()
		case *netv1.NetworkPolicy:
			out.SpecList[i] = v.DeepCopy()
		case *corev1.Endpoints:
			out.SpecList[i] = v.DeepCopy()
		case *storkapi.ResourceTransformation:
			out.SpecList[i] = v.DeepCopy()
		case *admissionregistrationv1.ValidatingWebhookConfiguration:
			out.SpecList[i] = v.DeepCopy()
		case *admissionregistrationv1.ValidatingWebhookConfigurationList:
			out.SpecList[i] = v.DeepCopy()
		default:
			out.SpecList[i] = v
		}
	}
	return out
}
