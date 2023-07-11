package app_spec

// AppSpec represents the specification for an App
type AppSpec struct {
	AppKey         string
	ValidateParams *ValidateParams
	TearDownParams *TeardownParams
}

// GetAppKey returns the AppKey associated with the AppSpec
func (s *AppSpec) GetAppKey() string {
	return s.AppKey
}

// SetAppKey sets the AppKey for the AppSpec
func (s *AppSpec) SetAppKey(appKey string) *AppSpec {
	s.AppKey = appKey
	return s
}

// GetValidateParams returns the ValidateParams associated with the AppSpec
func (s *AppSpec) GetValidateParams() *ValidateParams {
	return s.ValidateParams
}

// SetValidateParams sets the ValidateParams for the AppSpec
func (s *AppSpec) SetValidateParams(validateParams *ValidateParams) *AppSpec {
	s.ValidateParams = validateParams
	return s
}

// GetTearDownParams returns the TeardownParams associated with the AppSpec
func (s *AppSpec) GetTearDownParams() *TeardownParams {
	return s.TearDownParams
}

// SetTearDownParams sets the TeardownParams for the AppSpec
func (s *AppSpec) SetTearDownParams(tearDownParams *TeardownParams) *AppSpec {
	s.TearDownParams = tearDownParams
	return s
}

// NewAppSpec creates a new instance of the AppSpec
func NewAppSpec(appKey string, validateParams *ValidateParams, teardownParams *TeardownParams) *AppSpec {
	appSpec := &AppSpec{}
	appSpec.SetAppKey(appKey)
	appSpec.SetValidateParams(validateParams)
	appSpec.SetTearDownParams(teardownParams)
	return appSpec
}
