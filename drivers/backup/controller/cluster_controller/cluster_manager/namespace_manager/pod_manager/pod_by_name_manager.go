package pod_manager

type PodByNameMetaData struct {
	PodName string
}

func (m *PodByNameMetaData) GetPodName() string {
	return m.PodName
}

func (m *PodByNameMetaData) SetPodName(podName string) {
	m.PodName = podName
}

func (m *PodByNameMetaData) GetPodUid() string {
	return m.GetPodName()
}

func NewPodMetaData() *PodByNameMetaData {
	newPodMetaData := &PodByNameMetaData{}
	return newPodMetaData
}

type PodConfig struct {
	PodMetaData *PodByNameMetaData
}

type PodManager struct{}

func NewPodManager() *PodManager {
	newPodManager := &PodManager{}
	return newPodManager
}
