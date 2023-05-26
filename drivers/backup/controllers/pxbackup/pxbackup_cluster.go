package pxbackup

//type ClusterInfo struct {
//	*api.ClusterObject
//}
//
//func (p *PxBackupController) setClusterInfo(clusterName string, clusterInfo *ClusterInfo) {
//	if p.organizations[p.currentOrgId].clusters == nil {
//		p.organizations[p.currentOrgId].clusters = make(map[string]*ClusterInfo, 0)
//	}
//	p.organizations[p.currentOrgId].clusters[clusterName] = clusterInfo
//}
//
//func (p *PxBackupController) GetClusterInfo(clusterName string) (*ClusterInfo, bool) {
//	clusterInfo, ok := p.organizations[p.currentOrgId].clusters[clusterName]
//	if !ok {
//		return nil, false
//	}
//	return clusterInfo, true
//}
//
//func (p *PxBackupController) delClusterInfo(clusterName string) {
//	delete(p.organizations[p.currentOrgId].clusters, clusterName)
//}
//
//type AddClusterConfig struct {
//	clusterName      string
//	kubeconfigPath   string
//	cloudAccountName string         // default
//	clusterUid       string         // default
//	controller       *PxBackupController // fixed
//}
//
//func (c *AddClusterConfig) SetCloudAccountName(cloudAccountName string) *AddClusterConfig {
//	c.cloudAccountName = cloudAccountName
//	return c
//}
//
//func (c *AddClusterConfig) SetClusterUid(clusterUid string) *AddClusterConfig {
//	c.clusterUid = clusterUid
//	return c
//}
//
//func (p *PxBackupController) Cluster(clusterName string, kubeconfigPath string) *AddClusterConfig {
//	return &AddClusterConfig{
//		clusterName:      clusterName,
//		kubeconfigPath:   kubeconfigPath,
//		cloudAccountName: "",
//		clusterUid:       uuid.New(),
//		controller:       p,
//	}
//}
//
//func (c *AddClusterConfig) Add() error {
//	log.Infof("Adding cluster [%s] for org [%s]", c.clusterName, c.controller.currentOrgId)
//	kubeconfigRaw, err := os.ReadFile(c.kubeconfigPath)
//	if err != nil {
//		return err
//	}
//	var clusterCreateReq *api.ClusterCreateRequest
//	if c.cloudAccountName != "" {
//		cloudAccountInfo, ok := c.controller.GetCloudAccountInfo(c.cloudAccountName)
//		if !ok {
//			return fmt.Errorf("cloud account [%s] not found in cache", c.cloudAccountName)
//		}
//		clusterCreateReq = &api.ClusterCreateRequest{
//			CreateMetadata: &api.CreateMetadata{
//				Name:  c.clusterName,
//				OrgId: c.controller.currentOrgId,
//				Uid:   c.clusterUid,
//			},
//			Kubeconfig: base64.StdEncoding.EncodeToString(kubeconfigRaw),
//			CloudCredentialRef: &api.ObjectRef{
//				Name: c.cloudAccountName,
//				Uid:  cloudAccountInfo.GetUid(),
//			},
//		}
//	} else {
//		clusterCreateReq = &api.ClusterCreateRequest{
//			CreateMetadata: &api.CreateMetadata{
//				Name:  c.clusterName,
//				OrgId: c.controller.currentOrgId,
//			},
//			Kubeconfig: base64.StdEncoding.EncodeToString(kubeconfigRaw),
//		}
//	}
//	_, err = c.controller.processPxBackupRequest(clusterCreateReq)
//	if err != nil {
//		return err
//	}
//	clusterInspectInspectReq := &api.ClusterInspectRequest{
//		OrgId:          c.controller.currentOrgId,
//		Name:           c.clusterName,
//		IncludeSecrets: false,
//	}
//	resp, err := c.controller.processPxBackupRequest(clusterInspectInspectReq)
//	if err != nil {
//		return err
//	}
//	cluster := resp.(*api.ClusterInspectResponse).GetCluster()
//	c.controller.setClusterInfo(c.clusterName, &ClusterInfo{
//		ClusterObject: cluster,
//	})
//	return nil
//}
//
//func (p *PxBackupController) DeleteCluster(clusterName string) error {
//	clusterInfo, ok := p.GetClusterInfo(clusterName)
//	if ok {
//		log.Infof("Deleting cluster [%s] of org [%s]", clusterInfo.Name, clusterInfo.OrgId)
//		clusterDeleteReq := &api.ClusterDeleteRequest{
//			Name:  clusterInfo.Name,
//			OrgId: clusterInfo.OrgId,
//		}
//		if _, err := p.processPxBackupRequest(clusterDeleteReq); err != nil {
//			return err
//		}
//		p.delClusterInfo(clusterName)
//	}
//	return nil
//}
//
//type clusterStatusInfo struct {
//	Status api.ClusterInfo_StatusInfo_Status
//	Reason string
//	Error  error
//}
//
//func (p *PxBackupController) getClusterStatus(clusterName string) (clusterStatusInfo, error) {
//	clusterInfo, ok := p.GetClusterInfo(clusterName)
//	if !ok {
//		return clusterStatusInfo{}, fmt.Errorf("cluster [%s] not found in cache", clusterName)
//	}
//	clusterInspectRequest := &api.ClusterInspectRequest{
//		OrgId: clusterInfo.OrgId,
//		Name:  clusterInfo.Name,
//	}
//	resp, err := p.processPxBackupRequest(clusterInspectRequest)
//	if err != nil {
//		return clusterStatusInfo{
//			Status: api.ClusterInfo_StatusInfo_Invalid,
//			Reason: "",
//			Error:  err,
//		}, nil
//	}
//	cluster := resp.(*api.ClusterInspectResponse).GetCluster()
//	status := cluster.GetStatus().Status
//	reason := cluster.GetStatus().Reason
//	return clusterStatusInfo{
//		Status: status,
//		Reason: reason,
//		Error:  nil,
//	}, nil
//}
//
//func (p *PxBackupController) WaitForClusterCompletion(clusterName string) (api.ClusterInfo_StatusInfo_Status, error) {
//	getClusterStatus := func() interface{} {
//		status, err := p.getClusterStatus(clusterName)
//		log.Infof("backup status for [%s] is [%s]", clusterName, status)
//		if err != nil {
//			return clusterStatusInfo{
//				Status: api.ClusterInfo_StatusInfo_Invalid,
//				Error:  err,
//			}
//		}
//		return status
//	}
//	shouldRetry := func(result interface{}) bool {
//		res, ok := result.(clusterStatusInfo)
//		if !ok || res.Error != nil {
//			return false
//		}
//		finalStates := [...]api.ClusterInfo_StatusInfo_Status{
//			api.ClusterInfo_StatusInfo_Invalid,
//			api.ClusterInfo_StatusInfo_Online,
//			api.ClusterInfo_StatusInfo_Offline,
//			api.ClusterInfo_StatusInfo_Failed,
//			api.ClusterInfo_StatusInfo_Success,
//		}
//		log.Infof("cluster status for [%s] is [%s] bc [%s]", clusterName, res.Status, res.Reason)
//		for _, status := range finalStates {
//			if res.Status == status {
//				return false
//			}
//		}
//		return true
//	}
//	res, err := utils.DoRetryWithTimeout(getClusterStatus, utils.DefaultClusterAdditionTimeout, utils.DefaultClusterAdditionRetryTime, shouldRetry)
//	if err != nil {
//		return res.(clusterStatusInfo).Status, err
//	}
//	return res.(clusterStatusInfo).Status, nil
//}
