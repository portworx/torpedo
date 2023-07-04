package storage_location_metadata

const (
	// DefaultAWSLocationName is the default name for the aws_location_manager.AWSLocation
	DefaultAWSLocationName = ""
)

// AWSLocationMetaData represents the metadata for aws_location_manager.AWSLocation
type AWSLocationMetaData struct {
	AWSLocationName string
}

// GetAWSLocationName returns the AWSLocationName associated with the AWSLocationMetaData
func (m *AWSLocationMetaData) GetAWSLocationName() string {
	return m.AWSLocationName
}

// SetAWSLocationName sets the AWSLocationName for the AWSLocationMetaData
func (m *AWSLocationMetaData) SetAWSLocationName(awsLocationName string) *AWSLocationMetaData {
	m.AWSLocationName = awsLocationName
	return m
}

// GetAWSLocationUID returns the aws_location_manager.AWSLocation UID
func (m *AWSLocationMetaData) GetAWSLocationUID() string {
	return m.GetAWSLocationName()
}

// NewAWSLocationMetaData creates a new instance of the AWSLocationMetaData
func NewAWSLocationMetaData(awsLocationName string) *AWSLocationMetaData {
	newAWSLocationMetaData := &AWSLocationMetaData{}
	newAWSLocationMetaData.SetAWSLocationName(awsLocationName)
	return newAWSLocationMetaData
}

// NewDefaultAWSLocationMetaData creates a new instance of the AWSLocationMetaData with default values
func NewDefaultAWSLocationMetaData() *AWSLocationMetaData {
	return NewAWSLocationMetaData(DefaultAWSLocationName)
}
