package pod_by_name_manager

import . "github.com/portworx/torpedo/drivers/backup_controller/cluster_controller/cluster_metadata"

// PodByName represents PodByName
type PodByName struct{}

// NewPodByName creates a new instance of the PodByName
func NewPodByName() *PodByName {
	newPodByName := &PodByName{}
	return newPodByName
}

// NewDefaultPodByName creates a new instance of the PodByName with default values
func NewDefaultPodByName() *PodByName {
	return NewPodByName()
}

// PodByNameManager represents a manager for PodByName
type PodByNameManager struct {
	PodMap         map[string]*PodByName
	RemovedPodsMap map[string][]*PodByName
}

// GetPodMap returns the PodMap of the PodByNameManager
func (m *PodByNameManager) GetPodMap() map[string]*PodByName {
	return m.PodMap
}

// SetPodMap sets the PodMap of the PodByNameManager
func (m *PodByNameManager) SetPodMap(podMap map[string]*PodByName) *PodByNameManager {
	m.PodMap = podMap
	return m
}

// GetRemovedPodsMap returns the RemovedPodsMap of the PodByNameManager
func (m *PodByNameManager) GetRemovedPodsMap() map[string][]*PodByName {
	return m.RemovedPodsMap
}

// SetRemovedPodsMap sets the RemovedPodsMap of the PodByNameManager
func (m *PodByNameManager) SetRemovedPodsMap(removedPodsMap map[string][]*PodByName) *PodByNameManager {
	m.RemovedPodsMap = removedPodsMap
	return m
}

// GetPod returns the PodByName with the given PodByName UID
func (m *PodByNameManager) GetPod(podUID string) *PodByName {
	return m.PodMap[podUID]
}

// SetPod sets the PodByName with the given PodByName UID
func (m *PodByNameManager) SetPod(podUID string, pod *PodByName) *PodByNameManager {
	m.PodMap[podUID] = pod
	return m
}

// DeletePod deletes the PodByName with the given PodByName UID
func (m *PodByNameManager) DeletePod(podUID string) *PodByNameManager {
	delete(m.PodMap, podUID)
	return m
}

// RemovePod removes the PodByName with the given PodByName UID
func (m *PodByNameManager) RemovePod(podUID string) *PodByNameManager {
	if pod, isPresent := m.PodMap[podUID]; isPresent {
		m.RemovedPodsMap[podUID] = append(m.RemovedPodsMap[podUID], pod)
		delete(m.PodMap, podUID)
	}
	return m
}

// IsPodPresent checks if the PodByName with the given PodByName UID is present
func (m *PodByNameManager) IsPodPresent(podUID string) bool {
	_, isPresent := m.PodMap[podUID]
	return isPresent
}

// IsPodRemoved checks if the PodByName with the given PodByName UID is removed
func (m *PodByNameManager) IsPodRemoved(podUID string) bool {
	_, isPresent := m.RemovedPodsMap[podUID]
	return isPresent
}

// IsPodRecorded checks if the PodByName with the given PodByName UID is recorded
func (m *PodByNameManager) IsPodRecorded(podUID string) bool {
	return m.IsPodPresent(podUID) || m.IsPodRemoved(podUID)
}

// NewPodByNameManager creates a new instance of the PodByNameManager
func NewPodByNameManager(podMap map[string]*PodByName, removedPodsMap map[string][]*PodByName) *PodByNameManager {
	newPodByNameManager := &PodByNameManager{}
	newPodByNameManager.SetPodMap(podMap)
	newPodByNameManager.SetRemovedPodsMap(removedPodsMap)
	return newPodByNameManager
}

// NewDefaultPodByNameManager creates a new instance of the PodByNameManager with default values
func NewDefaultPodByNameManager() *PodByNameManager {
	return NewPodByNameManager(make(map[string]*PodByName, 0), make(map[string][]*PodByName, 0))
}

// PodByNameConfig represents the configuration for a PodByName
type PodByNameConfig struct {
	PodByNameManager  *PodByNameManager
	PodByNameMetaData *PodByNameMetaData
}

// GetPodByNameManager returns the PodByNameManager associated with the PodByNameConfig
func (c *PodByNameConfig) GetPodByNameManager() *PodByNameManager {
	return c.PodByNameManager
}

// SetPodByNameManager sets the PodByNameManager for the PodByNameConfig
func (c *PodByNameConfig) SetPodByNameManager(manager *PodByNameManager) *PodByNameConfig {
	c.PodByNameManager = manager
	return c
}

// GetPodByNameMetaData returns the PodByNameMetaData associated with the PodByNameConfig
func (c *PodByNameConfig) GetPodByNameMetaData() *PodByNameMetaData {
	return c.PodByNameMetaData
}

// SetPodByNameMetaData sets the PodByNameMetaData for the PodByNameConfig
func (c *PodByNameConfig) SetPodByNameMetaData(metaData *PodByNameMetaData) *PodByNameConfig {
	c.PodByNameMetaData = metaData
	return c
}

// NewPodByNameConfig creates a new instance of the PodByNameConfig
func NewPodByNameConfig(manager *PodByNameManager, metaData *PodByNameMetaData) *PodByNameConfig {
	newPodByNameConfig := &PodByNameConfig{}
	newPodByNameConfig.SetPodByNameManager(manager)
	newPodByNameConfig.SetPodByNameMetaData(metaData)
	return newPodByNameConfig
}

// NewDefaultPodByNameConfig creates a new instance of the PodByNameConfig with default values
func NewDefaultPodByNameConfig() *PodByNameConfig {
	return NewPodByNameConfig(NewDefaultPodByNameManager(), NewDefaultPodByNameMetaData())
}
