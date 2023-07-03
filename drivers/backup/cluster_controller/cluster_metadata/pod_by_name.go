package cluster_metadata

import . "github.com/portworx/torpedo/drivers/backup/cluster_controller/cluster_utils"

// GetNamespaceMetaData returns the NamespaceMetaData associated with the PodByNameMetaData
func (m *PodByNameMetaData) GetNamespaceMetaData() *NamespaceMetaData {
	return m.NamespaceMetaData
}

// SetNamespaceMetaData sets the NamespaceMetaData for the PodByNameMetaData
func (m *PodByNameMetaData) SetNamespaceMetaData(metaData *NamespaceMetaData) *PodByNameMetaData {
	m.NamespaceMetaData = metaData
	return m
}

// GetPodName returns the PodName associated with the PodByNameMetaData
func (m *PodByNameMetaData) GetPodName() string {
	return m.PodName
}

// SetPodName sets the PodName for the PodByNameMetaData
func (m *PodByNameMetaData) SetPodName(podName string) *PodByNameMetaData {
	m.PodName = podName
	return m
}

// GetPodByNameUID returns the PodByName UID
func (m *PodByNameMetaData) GetPodByNameUID() string {
	return m.GetPodName()
}

// NewPodByNameMetaData creates a new instance of the PodByNameMetaData
func NewPodByNameMetaData(metaData *NamespaceMetaData, podName string) *PodByNameMetaData {
	newPodByNameMetaData := &PodByNameMetaData{}
	newPodByNameMetaData.SetNamespaceMetaData(metaData)
	newPodByNameMetaData.SetPodName(podName)
	return newPodByNameMetaData
}

// NewDefaultPodByNameMetaData creates a new instance of the PodByNameMetaData with default values
func NewDefaultPodByNameMetaData() *PodByNameMetaData {
	return NewPodByNameMetaData(NewDefaultNamespaceMetaData(), DefaultPodName)
}
