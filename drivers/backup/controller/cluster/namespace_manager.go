package cluster

type NamespaceMetaData struct {
	Namespace string
}

func (m *NamespaceMetaData) GetName() string {
	return m.Namespace
}

func NewNamespaceMetaData(namespace string) *NamespaceMetaData {
	return &NamespaceMetaData{
		Namespace: namespace,
	}
}

type Namespace struct {
	AppManager *AppManager
}

type NamespaceManager struct {
	Namespaces        map[string]*Namespace
	RemovedNamespaces map[string][]*Namespace
}

func NewNamespace() *Namespace {
	return &Namespace{
		AppManager: &AppManager{
			Apps:        make(map[string]*App, 0),
			RemovedApps: make(map[string][]*App, 0),
		},
	}
}

func (m *NamespaceManager) GetNamespace(nsMetaData *NamespaceMetaData) *Namespace {
	return m.Namespaces[nsMetaData.GetName()]
}

func (m *NamespaceManager) AddNamespace(nsMetaData *NamespaceMetaData, namespace *Namespace) {
	m.Namespaces[nsMetaData.GetName()] = namespace
}

func (m *NamespaceManager) DeleteNamespace(nsMetaData *NamespaceMetaData) {
	delete(m.Namespaces, nsMetaData.GetName())
}

func (m *NamespaceManager) RemoveNamespace(nsMetaData *NamespaceMetaData) {
	m.RemovedNamespaces[nsMetaData.GetName()] = append(m.RemovedNamespaces[nsMetaData.GetName()], m.GetNamespace(nsMetaData))
	m.DeleteNamespace(nsMetaData)
}

func (m *NamespaceManager) IsNamespacePresent(nsMetaData *NamespaceMetaData) bool {
	_, ok := m.Namespaces[nsMetaData.GetName()]
	return ok
}
