package automationModels

type BackupLocationRequest struct {
	Create BackupLocation
	List   BackupLocation
	Delete DeleteBackupLocation
}

type DeleteBackupLocation struct {
	Id string
}

type BackupLocationResponse struct {
	Create BackupLocation
	Get    BackupLocation
	List   ListBackupLocation
	Update BackupLocation
}

type ListBackupLocation struct {
	BackupLocations []BackupLocation
	Pagination      *V1PageBasedPaginationResponse
}

type BackupLocation struct {
	TenantID             string
	Meta                 Meta
	Config               BackupLocationConfig
	Label                string
	SortSortBy           string
	SortSortOrder        string
	PaginationPageNumber string
	PaginationPageSize   string
}

type BackupLocationConfig struct {
	Provider           Provider
	CloudCredentialsId string
	BkpLocation        BkpLocation
}

type BkpLocation struct {
	S3Storage     S3Storage
	AzureStorage  AzureStorage
	GoogleStorage GoogleStorage
}

type S3Storage struct {
	BucketName string
	Region     string
	Endpoint   string
}

type AzureStorage struct {
	ContainerName string
}

type GoogleStorage struct {
	BucketName string
}
