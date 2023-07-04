package aws_credential_manager

const (
	// DefaultAWSCredentialSpecID specifies the default id for AWSCredentialSpec
	DefaultAWSCredentialSpecID = ""
	// DefaultAWSCredentialSpecSecret specifies the default secret for AWSCredentialSpec
	DefaultAWSCredentialSpecSecret = ""
	// DefaultAWSCredentialSpecEndpoint specifies the default endpoint for AWSCredentialSpec
	DefaultAWSCredentialSpecEndpoint = ""
	// DefaultAWSCredentialSpecS3Region specifies the default s3-region for AWSCredentialSpec
	DefaultAWSCredentialSpecS3Region = ""
	// DefaultAWSCredentialSpecDisableSSL specifies the default flag indicating whether SSL should be disabled for AWSCredentialSpec
	DefaultAWSCredentialSpecDisableSSL = false
)

// AWSCredentialSpec represents AWSCredentialSpec
type AWSCredentialSpec struct {
	ID         string
	Secret     string
	Endpoint   string
	S3Region   string
	DisableSSL bool
}

// GetID returns the ID associated with the AWSCredentialSpec
func (c *AWSCredentialSpec) GetID() string {
	return c.ID
}

// SetID sets the ID for the AWSCredentialSpec
func (c *AWSCredentialSpec) SetID(id string) *AWSCredentialSpec {
	c.ID = id
	return c
}

// GetSecret returns the Secret associated with the AWSCredentialSpec
func (c *AWSCredentialSpec) GetSecret() string {
	return c.Secret
}

// SetSecret sets the Secret for the AWSCredentialSpec
func (c *AWSCredentialSpec) SetSecret(secret string) *AWSCredentialSpec {
	c.Secret = secret
	return c
}

// GetEndpoint returns the Endpoint associated with the AWSCredentialSpec
func (c *AWSCredentialSpec) GetEndpoint() string {
	return c.Endpoint
}

// SetEndpoint sets the Endpoint for the AWSCredentialSpec
func (c *AWSCredentialSpec) SetEndpoint(endpoint string) *AWSCredentialSpec {
	c.Endpoint = endpoint
	return c
}

// GetS3Region returns the S3Region associated with the AWSCredentialSpec
func (c *AWSCredentialSpec) GetS3Region() string {
	return c.S3Region
}

// SetS3Region sets the S3Region for the AWSCredentialSpec
func (c *AWSCredentialSpec) SetS3Region(s3Region string) *AWSCredentialSpec {
	c.S3Region = s3Region
	return c
}

// GetDisableSSL returns the flag indicating whether SSL is disabled
func (c *AWSCredentialSpec) GetDisableSSL() bool {
	return c.DisableSSL
}

// SetDisableSSL sets the flag indicating whether SSL should be disabled
func (c *AWSCredentialSpec) SetDisableSSL(disableSSL bool) *AWSCredentialSpec {
	c.DisableSSL = disableSSL
	return c
}

// NewAWSCredentialSpec creates a new instance of the AWSCredentialSpec
func NewAWSCredentialSpec(id string, secret string, endpoint string, s3Region string, disableSSL bool) *AWSCredentialSpec {
	awsCredentialSpec := &AWSCredentialSpec{}
	awsCredentialSpec.SetID(id)
	awsCredentialSpec.SetSecret(secret)
	awsCredentialSpec.SetEndpoint(endpoint)
	awsCredentialSpec.SetS3Region(s3Region)
	awsCredentialSpec.SetDisableSSL(disableSSL)
	return awsCredentialSpec
}

// NewDefaultAWSCredentialSpec creates a new instance of the AWSCredentialSpec with default values
func NewDefaultAWSCredentialSpec() *AWSCredentialSpec {
	return NewAWSCredentialSpec(DefaultAWSCredentialSpecID, DefaultAWSCredentialSpecSecret, DefaultAWSCredentialSpecEndpoint, DefaultAWSCredentialSpecS3Region, DefaultAWSCredentialSpecDisableSSL)
}

type AWSCredential struct {
	AWSCredentialSpec *AWSCredentialSpec
	//AWSLocationManager *aws_location_manager.AWSLocationManager
}

type AWSCredentialManager struct {
	AWSCredentialMap        map[string]*AWSCredential
	RemovedAWSCredentialMap map[string][]*AWSCredential
}
