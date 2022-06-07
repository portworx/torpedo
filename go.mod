module github.com/portworx/torpedo

go 1.12

require (
	github.com/Azure/azure-storage-blob-go v0.9.0
	github.com/LINBIT/golinstor v0.27.0
	github.com/andygrunwald/go-jira v1.15.0
	github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
	github.com/aws/aws-sdk-go v1.40.39
	github.com/blang/semver v3.5.1+incompatible
	github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/educlos/testrail v0.0.0-20210915115134-adb5e6f62a6d
	github.com/fatih/color v1.9.0
	github.com/frankban/quicktest v1.14.2 // indirect
	github.com/gambol99/go-marathon v0.7.1
	github.com/gofrs/flock v0.8.1
	github.com/golang/protobuf v1.5.2
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/vault/api v1.0.5-0.20200902155336-f9d5ce5a171a
	github.com/heptio/velero v0.0.0-00010101000000-000000000000 // indirect
	github.com/kubernetes-incubator/external-storage v0.20.4-openstorage-rc7
	github.com/libopenstorage/autopilot-api v1.3.0
	github.com/libopenstorage/cloudops v0.0.0-20220420143942-8bdd341e5b41
	github.com/libopenstorage/openstorage v8.0.1-0.20211105030910-665c2f474186+incompatible
	github.com/libopenstorage/operator v0.0.0-20220603212904-3320db3aef81
	github.com/libopenstorage/stork v1.4.1-0.20220323180113-0ea773109d05
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.19.0
	github.com/openshift/api v0.0.0-20210105115604-44119421ec6b
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/portworx/px-backup-api v1.2.2-0.20210917042806-f2b0725444af
	github.com/portworx/sched-ops v1.20.4-rc1.0.20220327212454-cc1a88ecb579
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.46.0
	github.com/sendgrid/sendgrid-go v3.6.0+incompatible
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/vmware/govmomi v0.22.2
	gocloud.dev v0.20.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f
	google.golang.org/genproto v0.0.0-20210921142501-181ce0d877f6
	google.golang.org/grpc v1.40.0
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.0.0-00010101000000-000000000000
	k8s.io/api v0.21.4
	k8s.io/apiextensions-apiserver v0.21.4
	k8s.io/apimachinery v0.21.4
	k8s.io/client-go v12.0.0+incompatible
)

replace (
	// Misc dependencies
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.2.0+incompatible
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20190717161051-705d9623b7c1+incompatible
	github.com/go-logr/logr => github.com/go-logr/logr v0.3.0
	github.com/heptio/ark => github.com/heptio/ark v1.0.0
	github.com/heptio/velero => github.com/heptio/velero v1.0.0
	//github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible

	// PX dependencies
	github.com/kubernetes-incubator/external-storage => github.com/libopenstorage/external-storage v0.20.4-openstorage-rc7
	github.com/libopenstorage/autopilot-api => github.com/libopenstorage/autopilot-api v0.6.1-0.20210301232050-ca2633c6e114
	github.com/libopenstorage/openstorage => github.com/libopenstorage/openstorage v1.0.1-0.20220204053814-097a5af93b1e
	//github.com/libopenstorage/operator => github.com/libopenstorage/operator v0.0.0-20220121222253-3431532a94f9

	// Stork dependencies
	github.com/libopenstorage/stork => github.com/libopenstorage/stork v1.4.1-0.20220323180113-0ea773109d05
	github.com/portworx/sched-ops => github.com/portworx/sched-ops v1.20.4-rc1.0.20220327212454-cc1a88ecb579
	helm.sh/helm/v3 => helm.sh/helm/v3 v3.6.0

	// Replacing k8s.io dependencies is required if a dependency or any dependency of a dependency
	// depends on k8s.io/kubernetes. See https://github.com/kubernetes/kubernetes/issues/79384#issuecomment-505725449
	k8s.io/api => k8s.io/api v0.21.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.5-rc.0
	k8s.io/apiserver => k8s.io/apiserver v0.21.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.4
	k8s.io/client-go => k8s.io/client-go v0.21.4
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.4
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.4
	k8s.io/code-generator => k8s.io/code-generator v0.21.5-rc.0
	k8s.io/component-base => k8s.io/component-base v0.21.4
	k8s.io/component-helpers => k8s.io/component-helpers v0.21.4
	k8s.io/controller-manager => k8s.io/controller-manager v0.21.4
	k8s.io/cri-api => k8s.io/cri-api v0.21.5-rc.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.4
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.4.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.4
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.4
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.4
	k8s.io/kubectl => k8s.io/kubectl v0.21.4
	k8s.io/kubelet => k8s.io/kubelet v0.21.4
	k8s.io/kubernetes => k8s.io/kubernetes v1.20.4
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.4
	k8s.io/metrics => k8s.io/metrics v0.21.4
	k8s.io/mount-utils => k8s.io/mount-utils v0.21.5-rc.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.4
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.21.4
	k8s.io/sample-controller => k8s.io/sample-controller v0.21.4
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.9.0
	sigs.k8s.io/sig-storage-lib-external-provisioner/v6 => sigs.k8s.io/sig-storage-lib-external-provisioner/v6 v6.3.0

)
