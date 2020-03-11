module github.com/portworx/torpedo

go 1.12

require (
	github.com/Azure/azure-sdk-for-go v38.2.0+incompatible // indirect
	github.com/Azure/go-autorest v13.3.1+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.9.4 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.2 // indirect
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/aws/aws-sdk-go v1.25.11
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.4.2-0.20170731201938-4f3616fb1c11
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/donovanhide/eventsource v0.0.0-20171031113327-3ed64d21fb0b // indirect
	github.com/fatih/color v1.7.0
	github.com/gambol99/go-marathon v0.7.1
	github.com/golang/groupcache v0.0.0-20191002201903-404acd9df4cc // indirect
	github.com/golang/protobuf v1.3.2
	github.com/kubernetes-incubator/external-storage v0.0.0-00010101000000-000000000000
	github.com/libopenstorage/autopilot-api v0.6.1-0.20200115200747-7383c6007283
	github.com/libopenstorage/cloudops v0.0.0-20200114171448-10fa10d97720
	github.com/libopenstorage/gossip v0.0.0-20190507031959-c26073a01952 // indirect
	github.com/libopenstorage/openstorage v8.0.1-0.20200129213931-ba5ed2ee0f18+incompatible
	github.com/libopenstorage/stork v1.3.0-beta1.0.20191009210244-6a3497c42b2a
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/onsi/ginkgo v1.10.2
	github.com/onsi/gomega v1.7.0
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pborman/uuid v0.0.0-20180906182336-adf5a7427709
	github.com/portworx/kvdb v0.0.0-20191223203141-f42097b1fcd8 // indirect
	github.com/portworx/px-backup-api v0.0.0-20200205061835-5dc42f2a6d0f
	github.com/portworx/sched-ops v0.0.0-20200221004841-9205a51a1fee
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.4.0
	go.opencensus.io v0.22.1 // indirect
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413
	golang.org/x/net v0.0.0-20191009170851-d66e71096ffb
	google.golang.org/grpc v1.24.0
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/apiserver v0.0.0-20190820063401-c43cd040845a // indirect
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cloud-provider v0.17.2 // indirect
	k8s.io/csi-translation-lib v0.0.0-20190913091657-9745ba0e69cf // indirect
	k8s.io/kubernetes v1.14.6 // indirect
)

replace (
	github.com/kubernetes-incubator/external-storage v0.0.0-00010101000000-000000000000 => github.com/libopenstorage/external-storage v5.1.1-0.20190919185747-9394ee8dd536+incompatible
	k8s.io/api => k8s.io/api v0.0.0-20190816222004-e3a6b8045b0b
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190918224502-6154570c2037
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190816221834-a9f1d8a9c101
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190820063401-c43cd040845a
	k8s.io/client-go v2.0.0-alpha.0.0.20181121191925-a47917edff34+incompatible => k8s.io/client-go v2.0.0+incompatible
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190913091657-9745ba0e69cf
	k8s.io/kubernetes => k8s.io/kubernetes v1.14.6
)
