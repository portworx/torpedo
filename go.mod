module github.com/portworx/torpedo

go 1.12

require (
	github.com/Azure/azure-storage-blob-go v0.9.0
	github.com/LINBIT/golinstor v0.27.0
	github.com/aws/aws-sdk-go v1.35.37
	github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/fatih/color v1.9.0
	github.com/gambol99/go-marathon v0.7.1
	github.com/gofrs/flock v0.8.0
	github.com/golang/protobuf v1.4.3
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/vault/api v1.0.5-0.20200317185738-82f498082f02
	github.com/kubernetes-incubator/external-storage v0.20.4-openstorage-rc2
	github.com/libopenstorage/autopilot-api v1.3.0
	github.com/libopenstorage/cloudops v0.0.0-20210223183702-b9c6b74cbf1d
	github.com/libopenstorage/openstorage v8.0.1-0.20200914191223-6fb8d163a67c+incompatible
	github.com/libopenstorage/stork v1.3.0-beta1.0.20210331175814-30fe03cc1e33
	github.com/onsi/ginkgo v1.15.0
	github.com/onsi/gomega v1.10.5
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/portworx/px-backup-api v1.0.1-0.20200915150042-274508e876ef
	github.com/portworx/sched-ops v1.20.4-rc1.0.20210401031424-429c07e70276
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.46.0
	github.com/sendgrid/sendgrid-go v3.6.0+incompatible
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/vmware/govmomi v0.22.2
	gocloud.dev v0.20.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	google.golang.org/grpc v1.33.2
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.5.2
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v12.0.0+incompatible
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.2.0+incompatible
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20190717161051-705d9623b7c1+incompatible
	github.com/kubernetes-incubator/external-storage => github.com/libopenstorage/external-storage v0.20.4-openstorage-rc3
	github.com/kubernetes-incubator/external-storage v0.0.0-00010101000000-000000000000 => github.com/libopenstorage/external-storage v5.3.0-alpha.1.0.20200130041458-d2b33d4448ea+incompatible
	github.com/libopenstorage/autopilot-api => github.com/libopenstorage/autopilot-api v0.6.1-0.20210301232050-ca2633c6e114
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v1.8.2-0.20190424153033-d3245f150225
	helm.sh/helm/v3 => helm.sh/helm/v3 v3.5.2
	k8s.io/api => k8s.io/api v0.20.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.4
	k8s.io/apiserver => k8s.io/apiserver v0.20.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.4
	k8s.io/client-go => k8s.io/client-go v0.20.4
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.4
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.4
	k8s.io/code-generator => k8s.io/code-generator v0.20.4
	k8s.io/component-base => k8s.io/component-base v0.20.4
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.4
	k8s.io/controller-manager => k8s.io/controller-manager v0.20.4
	k8s.io/cri-api => k8s.io/cri-api v0.20.4
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.4
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.4
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.4
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.4
	k8s.io/kubectl => k8s.io/kubectl v0.20.4
	k8s.io/kubelet => k8s.io/kubelet v0.20.4
	k8s.io/kubernetes => k8s.io/kubernetes v1.20.4
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.4
	k8s.io/metrics => k8s.io/metrics v0.20.4
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.4
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.4

)
