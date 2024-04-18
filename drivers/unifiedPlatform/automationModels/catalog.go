package automationModels

type CatalogResponse struct {
	DataServiceList        []V1DataService
	DataServiceVersionList []V1DataService
	DataServiceImageList   []V1Image
}

type V1DataService struct {
	Meta *V1Meta `copier:"must,nopanic"`
}

// V1Image Resource representing the data service image.
type V1Image struct {
	Meta *V1Meta      `copier:"must,nopanic"`
	Info *V1InfoImage `copier:"must,nopanic"`
}

type V1InfoImage struct {
	References *V1ImageReferences `copier:"must,nopanic"`
	// Image registry where the image is stored.
	Registry *string `copier:"must,nopanic"`
	// Image registry namespace where the image is stored.
	Namespace *string `copier:"must,nopanic"`
	// Tag associated with the image.
	Tag *string `copier:"must,nopanic"`
	// Build version of the image.
	Build *string `copier:"must,nopanic"`
	// Flag indicating if TLS is supported for a data service using this image.
	TlsSupport *bool `copier:"must,nopanic"`
	// Capabilities associated with this image.
	Capabilities *map[string]string `copier:"must,nopanic"`
	// Additional images associated with this data service image.
	AdditionalImages *map[string]string `copier:"must,nopanic"`
}

// V1References References to other resources.
type V1ImageReferences struct {
	// UID of the Data service.
	DataServiceId *string `copier:"must,nopanic"`
	// UID of the Data service version.
	DataServiceVersionId *string `copier:"must,nopanic"`
}
