package apiStructs

type PDSRestore struct {
	Create   PDSCreateRestore
	ReCreate PDSReCreateRestore
	Get      PDSGetRestore
	Delete   PDSDeleteRestore
	List     PDSListRestores
}

type PDSCreateRestore struct {
	NamespaceId string   `copier:"must,nopanic"`
	ProjectId   string   `copier:"must,nopanic"`
	Restore     *Restore `copier:"must,nopanic"`
	// SourceReferences for the restore.
	SourceReferences *SourceReferences `copier:"must,nopanic"`
	// Destination references for the restore.
	DestinationReferences *DestinationReferences `copier:"must,nopanic"`
	// K8s resource name for restore, built from ["restore-" + name + short-id].
	CustomResourceName string `copier:"must,nopanic"`
}

type Restore struct {
	Meta   *Meta   `copier:"must,nopanic"`
	Config *Config `copier:"must,nopanic"`
	Status *Status `copier:"must,nopanic"`
}

type PDSReCreateRestore struct {
	Id              string `copier:"must,nopanic"`
	TargetClusterId string `copier:"must,nopanic"`
	Name            string `copier:"must,nopanic"`
	NamespaceId     string `copier:"must,nopanic"`
}

type PDSGetRestore struct {
	Id string `copier:"must,nopanic"`
}

type PDSDeleteRestore struct {
	Id string `copier:"must,nopanic"`
}

type PDSListRestores struct {
	Sort       *Sort                       `copier:"must,nopanic"`
	Pagination *PageBasedPaginationRequest `copier:"must,nopanic"`
}
