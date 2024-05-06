package automationModels

//TODO: This needs to be moved to workflow level

type DataServiceDetails struct {
	Deployment        V1Deployment
	Namespace         string
	NamespaceId       string
	SourceMd5Checksum string
}
