package platformLibs

import (
	"context"
	"fmt"
	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	"github.com/jetstack/cert-manager/pkg/client/clientset/versioned"
	pdsdriver "github.com/pure-px/torpedo/drivers/pds"
	"github.com/pure-px/torpedo/drivers/unifiedPlatform/automationModels"
	"github.com/pure-px/torpedo/pkg/log"
	"github.com/pure-px/torpedo/pkg/osutils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

const (
	targetClusterHealthOK = "CONNECTED"
	CertManager           = "jetstack/cert-manager"
	CertManagerVersion    = "v1.11.0"
	TLSFeatureGates       = "AdditionalCertificateOutputFormats=true"
)

// GetManifest Get the manifest for the account and tenant-id that can be used to install the platform agent
func GetManifest(tenantId string, clusterName string) (*automationModels.V1TargetClusterRegistrationManifest, error) {
	manifestInputs := automationModels.PlatformTargetClusterRequest{
		GetManifest: automationModels.PlatformGetTargetClusterManifest{
			ClusterName: clusterName,
			TenantId:    tenantId,
		},
	}

	// TODO: Proxy and Registry configs need to be added to this call

	if clusterName == "" {
		clusterName = fmt.Sprintf("Cluster_%v", time.Now().Unix())
	}

	log.Infof("cluster name [%s]", manifestInputs.GetManifest.ClusterName)

	// Get Manifest from API
	manifest, err := v2Components.Platform.GetTargetClusterRegistrationManifest(&manifestInputs)
	if err != nil {
		return nil, err
	}

	return &manifest.GetManifest, nil
}

func ListTargetClusters(tenantId string) (*automationModels.V1ListTargetClustersResponse, error) {
	wfRequest := automationModels.PlatformTargetClusterRequest{
		ListTargetClusters: automationModels.PlatformListTargetCluster{
			TenantId: tenantId,
		},
	}

	tcList, err := v2Components.Platform.ListTargetClusters(&wfRequest)
	if err != nil {
		return nil, err
	}

	totalRecords := *tcList.ListTargetClusters.Pagination.TotalRecords
	log.Infof("Total target clusters under [%s] are [%s]", tenantId, totalRecords)

	wfRequest = automationModels.PlatformTargetClusterRequest{
		ListTargetClusters: automationModels.PlatformListTargetCluster{
			TenantId:             tenantId,
			PaginationPageSize:   totalRecords,
			PaginationPageNumber: DEFAULT_PAGE_NUMBER,
			SortSortOrder:        DEFAULT_SORT_ORDER,
			SortSortBy:           DEFAULT_SORT_BY,
		},
	}

	tcList, err = v2Components.Platform.ListTargetClusters(&wfRequest)
	if err != nil {
		return &tcList.ListTargetClusters, err
	}

	return &tcList.ListTargetClusters, nil
}

func GetTargetCluster(clusterId string) (*automationModels.V1TargetCluster, error) {
	wfRequest := automationModels.PlatformTargetClusterRequest{
		GetTargetCluster: automationModels.PlatformGetTargetCluster{
			Id: clusterId,
		},
	}
	tc, err := v2Components.Platform.GetTargetCluster(&wfRequest)
	if err != nil {
		return nil, err
	}
	return &tc.GetTargetCluster, nil
}

// InstallCertManager installs cert-manager using helm
func InstallCertManager(ClusterIssuerNamespace string) error {

	cmd1 := fmt.Sprintf("helm repo add jetstack https://charts.jetstack.io && helm repo update")
	output, _, err := osutils.ExecShell(cmd1)
	if err != nil {
		return fmt.Errorf("cert manager installation failed : %v", err)
	}

	cmd2 := fmt.Sprintf("helm upgrade --install cert-manager %s --namespace %s --create-namespace "+
		"--version %s --set installCRDs=%s --set featureGates=%s --set webhook.extraArgs={--feature-gates=%s}",
		CertManager, ClusterIssuerNamespace, CertManagerVersion, "true", TLSFeatureGates, TLSFeatureGates)
	output, _, err = osutils.ExecShell(cmd2)
	if err != nil {
		return fmt.Errorf("cert manager installation failed : %v", err)
	}
	log.Infof("Terminal output: %v", output)
	return nil
}

// CreateSelfSignedClusterIssuer Creates SelfSigned ClusterIssuer
func CreateSelfSignedClusterIssuer(ClusterIssuerName string) error {
	_, config, err := pdsdriver.GetK8sContext()
	if err != nil {
		return err
	}
	ClusterIssuerSelfSignedSpec := &cm.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name: ClusterIssuerName,
		},
		Spec: cm.IssuerSpec{
			IssuerConfig: cm.IssuerConfig{
				SelfSigned: new(cm.SelfSignedIssuer),
			},
		},
	}
	certManagerClient, err := versioned.NewForConfig(config)
	if err != nil {
		return err
	}
	_, err = certManagerClient.CertmanagerV1().ClusterIssuers().Create(context.TODO(), ClusterIssuerSelfSignedSpec, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		log.Infof("Cluster issuer alredy exists with the name %s", ClusterIssuerName)
	} else {
		return err
	}
	return nil
}

func DeleteTargetCluster(targetClusterId string, force bool) error {
	wfRequest := automationModels.PlatformTargetClusterRequest{
		DeleteTargetCluster: automationModels.PlatformDeleteTargetCluster{
			Id: targetClusterId,
		},
	}
	err := v2Components.Platform.DeleteTargetCluster(&wfRequest, force)
	return err
}
