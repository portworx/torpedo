package stworkflows

import pdslibs "github.com/portworx/torpedo/drivers/unifiedPlatform/pdsLibs"

type WorkflowK8s struct {
	FunctionName        string
	WorkflowDataService WorkflowDataService
}

func (k8sOp *WorkflowK8s) ExecuteK8sOperations(functionName string, namespace string, deployment map[string]string) error {
	deployment = k8sOp.WorkflowDataService.DataServiceDeployment
	err := pdslibs.ExecuteK8sOperations(functionName, namespace, deployment)
	if err != nil {
		return err
	}
	return nil
}
