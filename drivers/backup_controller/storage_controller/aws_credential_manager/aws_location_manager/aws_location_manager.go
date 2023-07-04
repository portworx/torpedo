package aws_location_manager

import (
	. "github.com/portworx/torpedo/drivers/backup_controller/storage_controller/storage_location_metadata"
)

const (
	DefaultObjectLock     = false
	DefaultRetainCount    = int64(0)
	DefaultObjectLockMode = ""
)

// AWSLocationSpec represents the spec for AWSLocation
type AWSLocationSpec struct {
	ObjectLock     bool
	RetainCount    int64
	ObjectLockMode string
}

// GetObjectLock returns the ObjectLock associated with the AWSLocationSpec
func (s *AWSLocationSpec) GetObjectLock() bool {
	return s.ObjectLock
}

// SetObjectLock sets the ObjectLock for the AWSLocationSpec
func (s *AWSLocationSpec) SetObjectLock(objectLock bool) *AWSLocationSpec {
	s.ObjectLock = objectLock
	return s
}

// GetRetainCount returns the RetainCount associated with the AWSLocationSpec
func (s *AWSLocationSpec) GetRetainCount() int64 {
	return s.RetainCount
}

// SetRetainCount sets the RetainCount for the AWSLocationSpec
func (s *AWSLocationSpec) SetRetainCount(retainCount int64) *AWSLocationSpec {
	s.RetainCount = retainCount
	return s
}

// GetObjectLockMode returns the ObjectLockMode associated with the AWSLocationSpec
func (s *AWSLocationSpec) GetObjectLockMode() string {
	return s.ObjectLockMode
}

// SetObjectLockMode sets the ObjectLockMode for the AWSLocationSpec
func (s *AWSLocationSpec) SetObjectLockMode(objectLockMode string) *AWSLocationSpec {
	s.ObjectLockMode = objectLockMode
	return s
}

// NewAWSLocationSpec creates a new instance of the AWSLocationSpec
func NewAWSLocationSpec(awsCredential *AWSCredential, objectLock bool, retainCount int64, objectLockMode string) *AWSLocationSpec {
	awsLocationSpec := &AWSLocationSpec{}
	awsLocationSpec.SetAWSCredential(awsCredential)
	awsLocationSpec.SetObjectLock(objectLock)
	awsLocationSpec.SetRetainCount(retainCount)
	awsLocationSpec.SetObjectLockMode(objectLockMode)
	return awsLocationSpec
}

// NewDefaultAWSLocationSpec creates a new instance of the AWSLocationSpec with default values
func NewDefaultAWSLocationSpec() *AWSLocationSpec {
	return NewAWSLocationSpec(NewDefaultAWSCredential(), DefaultObjectLock, DefaultRetainCount, DefaultObjectLockMode)
}

// AWSLocation represents AWSLocation
type AWSLocation struct {
	AWSLocationSpec *AWSLocationSpec
}

// GetAWSLocationSpec returns the AWSLocationSpec associated with the AWSLocation
func (l *AWSLocation) GetAWSLocationSpec() *AWSLocationSpec {
	return l.AWSLocationSpec
}

// SetAWSLocationSpec sets the AWSLocationSpec for the AWSLocation
func (l *AWSLocation) SetAWSLocationSpec(spec *AWSLocationSpec) *AWSLocation {
	l.AWSLocationSpec = spec
	return l
}

// NewAWSLocation creates a new instance of the AWSLocation
func NewAWSLocation(awsLocationSpec *AWSLocationSpec) *AWSLocation {
	newAWSLocation := &AWSLocation{}
	newAWSLocation.SetAWSLocationSpec(awsLocationSpec)
	return newAWSLocation
}

// NewDefaultAWSLocation creates a new instance of the AWSLocation with default values
func NewDefaultAWSLocation() *AWSLocation {
	return NewAWSLocation(NewDefaultAWSLocationSpec())
}

// AWSLocationManager represents a manager for AWSLocation
type AWSLocationManager struct {
	AWSLocationMap         map[string]*AWSLocation
	RemovedAWSLocationsMap map[string][]*AWSLocation
}

// GetAWSLocationMap returns the AWSLocationMap of the AWSLocationManager
func (m *AWSLocationManager) GetAWSLocationMap() map[string]*AWSLocation {
	return m.AWSLocationMap
}

// SetAWSLocationMap sets the AWSLocationMap of the AWSLocationManager
func (m *AWSLocationManager) SetAWSLocationMap(awsLocationMap map[string]*AWSLocation) *AWSLocationManager {
	m.AWSLocationMap = awsLocationMap
	return m
}

// GetRemovedAWSLocationsMap returns the RemovedAWSLocationsMap of the AWSLocationManager
func (m *AWSLocationManager) GetRemovedAWSLocationsMap() map[string][]*AWSLocation {
	return m.RemovedAWSLocationsMap
}

// SetRemovedAWSLocationsMap sets the RemovedAWSLocationsMap of the AWSLocationManager
func (m *AWSLocationManager) SetRemovedAWSLocationsMap(removedAWSLocationsMap map[string][]*AWSLocation) *AWSLocationManager {
	m.RemovedAWSLocationsMap = removedAWSLocationsMap
	return m
}

// GetAWSLocation returns the AWSLocation with the given AWSLocation UID
func (m *AWSLocationManager) GetAWSLocation(awsLocationUID string) *AWSLocation {
	return m.AWSLocationMap[awsLocationUID]
}

// SetAWSLocation sets the AWSLocation with the given AWSLocation UID
func (m *AWSLocationManager) SetAWSLocation(awsLocationUID string, awsLocation *AWSLocation) *AWSLocationManager {
	m.AWSLocationMap[awsLocationUID] = awsLocation
	return m
}

// DeleteAWSLocation deletes the AWSLocation with the given AWSLocation UID
func (m *AWSLocationManager) DeleteAWSLocation(awsLocationUID string) *AWSLocationManager {
	delete(m.AWSLocationMap, awsLocationUID)
	return m
}

// RemoveAWSLocation removes the AWSLocation with the given AWSLocation UID
func (m *AWSLocationManager) RemoveAWSLocation(awsLocationUID string) *AWSLocationManager {
	if awsLocation, isPresent := m.AWSLocationMap[awsLocationUID]; isPresent {
		m.RemovedAWSLocationsMap[awsLocationUID] = append(m.RemovedAWSLocationsMap[awsLocationUID], awsLocation)
		delete(m.AWSLocationMap, awsLocationUID)
	}
	return m
}

// IsAWSLocationPresent checks if the AWSLocation with the given AWSLocation UID is present
func (m *AWSLocationManager) IsAWSLocationPresent(awsLocationUID string) bool {
	_, isPresent := m.AWSLocationMap[awsLocationUID]
	return isPresent
}

// IsAWSLocationRemoved checks if the AWSLocation with the given AWSLocation UID is removed
func (m *AWSLocationManager) IsAWSLocationRemoved(awsLocationUID string) bool {
	_, isPresent := m.RemovedAWSLocationsMap[awsLocationUID]
	return isPresent
}

// IsAWSLocationRecorded checks if the AWSLocation with the given AWSLocation UID is recorded
func (m *AWSLocationManager) IsAWSLocationRecorded(awsLocationUID string) bool {
	return m.IsAWSLocationPresent(awsLocationUID) || m.IsAWSLocationRemoved(awsLocationUID)
}

// NewAWSLocationManager creates a new instance of the AWSLocationManager
func NewAWSLocationManager(awsLocationMap map[string]*AWSLocation, removedAWSLocationsMap map[string][]*AWSLocation) *AWSLocationManager {
	newAWSLocationManager := &AWSLocationManager{}
	newAWSLocationManager.SetAWSLocationMap(awsLocationMap)
	newAWSLocationManager.SetRemovedAWSLocationsMap(removedAWSLocationsMap)
	return newAWSLocationManager
}

// NewDefaultAWSLocationManager creates a new instance of the AWSLocationManager with default values
func NewDefaultAWSLocationManager() *AWSLocationManager {
	return NewAWSLocationManager(make(map[string]*AWSLocation, 0), make(map[string][]*AWSLocation, 0))
}

// AWSLocationConfig represents the configuration for a AWSLocation
type AWSLocationConfig struct {
	AWSLocationManager  *AWSLocationManager
	AWSLocationMetaData *AWSLocationMetaData
}

// GetAWSLocationManager returns the AWSLocationManager associated with the AWSLocationConfig
func (c *AWSLocationConfig) GetAWSLocationManager() *AWSLocationManager {
	return c.AWSLocationManager
}

// SetAWSLocationManager sets the AWSLocationManager for the AWSLocationConfig
func (c *AWSLocationConfig) SetAWSLocationManager(manager *AWSLocationManager) *AWSLocationConfig {
	c.AWSLocationManager = manager
	return c
}

// GetAWSLocationMetaData returns the AWSLocationMetaData associated with the AWSLocationConfig
func (c *AWSLocationConfig) GetAWSLocationMetaData() *AWSLocationMetaData {
	return c.AWSLocationMetaData
}

// SetAWSLocationMetaData sets the AWSLocationMetaData for the AWSLocationConfig
func (c *AWSLocationConfig) SetAWSLocationMetaData(metaData *AWSLocationMetaData) *AWSLocationConfig {
	c.AWSLocationMetaData = metaData
	return c
}

// NewAWSLocationConfig creates a new instance of the AWSLocationConfig
func NewAWSLocationConfig(manager *AWSLocationManager, metaData *AWSLocationMetaData) *AWSLocationConfig {
	newAWSLocationConfig := &AWSLocationConfig{}
	newAWSLocationConfig.SetAWSLocationManager(manager)
	newAWSLocationConfig.SetAWSLocationMetaData(metaData)
	return newAWSLocationConfig
}

// NewDefaultAWSLocationConfig creates a new instance of the AWSLocationConfig with default values
func NewDefaultAWSLocationConfig() *AWSLocationConfig {
	return NewAWSLocationConfig(NewDefaultAWSLocationManager(), NewDefaultAWSLocationMetaData())
}
