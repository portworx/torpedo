package app_spec

import "time"

const (
	// DefaultWaitForRunningTimeout indicates the duration to wait for an app to reach the running state
	DefaultWaitForRunningTimeout = 10 * time.Minute
	// DefaultWaitForRunningRetryInterval indicates the interval between retries when waiting for an app to reach the running state
	DefaultWaitForRunningRetryInterval = 10 * time.Second
	// DefaultValidateVolumeTimeout indicates the duration to wait for volume validation of an app
	DefaultValidateVolumeTimeout = 10 * time.Minute
	// DefaultValidateVolumeRetryInterval indicates the interval between retries when performing volume validation of an app
	DefaultValidateVolumeRetryInterval = 10 * time.Second
)

// ValidateParams represents the parameters for validating an App
type ValidateParams struct {
	WaitForRunningTimeout       time.Duration
	WaitForRunningRetryInterval time.Duration
	ValidateVolumeTimeout       time.Duration
	ValidateVolumeRetryInterval time.Duration
}

// GetWaitForRunningTimeout returns the WaitForRunningTimeout associated with the ValidateParams
func (p *ValidateParams) GetWaitForRunningTimeout() time.Duration {
	return p.WaitForRunningTimeout
}

// SetWaitForRunningTimeout sets the WaitForRunningTimeout for the ValidateParams
func (p *ValidateParams) SetWaitForRunningTimeout(timeout time.Duration) *ValidateParams {
	p.WaitForRunningTimeout = timeout
	return p
}

// GetWaitForRunningRetryInterval returns the WaitForRunningRetryInterval associated with the ValidateParams
func (p *ValidateParams) GetWaitForRunningRetryInterval() time.Duration {
	return p.WaitForRunningRetryInterval
}

// SetWaitForRunningRetryInterval sets the WaitForRunningRetryInterval for the ValidateParams
func (p *ValidateParams) SetWaitForRunningRetryInterval(interval time.Duration) *ValidateParams {
	p.WaitForRunningRetryInterval = interval
	return p
}

// GetValidateVolumeTimeout returns the ValidateVolumeTimeout associated with the ValidateParams
func (p *ValidateParams) GetValidateVolumeTimeout() time.Duration {
	return p.ValidateVolumeTimeout
}

// SetValidateVolumeTimeout sets the ValidateVolumeTimeout for the ValidateParams
func (p *ValidateParams) SetValidateVolumeTimeout(timeout time.Duration) *ValidateParams {
	p.ValidateVolumeTimeout = timeout
	return p
}

// GetValidateVolumeRetryInterval returns the ValidateVolumeRetryInterval associated with the ValidateParams
func (p *ValidateParams) GetValidateVolumeRetryInterval() time.Duration {
	return p.ValidateVolumeRetryInterval
}

// SetValidateVolumeRetryInterval sets the ValidateVolumeRetryInterval for the ValidateParams
func (p *ValidateParams) SetValidateVolumeRetryInterval(interval time.Duration) *ValidateParams {
	p.ValidateVolumeRetryInterval = interval
	return p
}

// NewValidateParams creates a new instance of the ValidateParams
func NewValidateParams(waitForRunningTimeout time.Duration, waitForRunningRetryInterval time.Duration, validateVolumeTimeout time.Duration, validateVolumeRetryInterval time.Duration) *ValidateParams {
	validateParams := &ValidateParams{}
	validateParams.SetWaitForRunningTimeout(waitForRunningTimeout)
	validateParams.SetWaitForRunningRetryInterval(waitForRunningRetryInterval)
	validateParams.SetValidateVolumeTimeout(validateVolumeTimeout)
	validateParams.SetValidateVolumeRetryInterval(validateVolumeRetryInterval)
	return validateParams
}

// NewDefaultValidateParams creates a new instance of the ValidateParams with default values
func NewDefaultValidateParams() *ValidateParams {
	return NewValidateParams(DefaultWaitForRunningTimeout, DefaultWaitForRunningRetryInterval, DefaultValidateVolumeTimeout, DefaultValidateVolumeRetryInterval)
}
