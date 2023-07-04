package torpedotest_metadata

const (
	DefaultTestID = ""
)

// TorpedoTestMetaData represents the metadata for a torpedotest_manager.TorpedoTest
type TorpedoTestMetaData struct {
	TestID string
}

// GetTestID returns the TestID associated with the TorpedoTestMetaData
func (m *TorpedoTestMetaData) GetTestID() string {
	return m.TestID
}

// SetTestID sets the TestID for the TorpedoTestMetaData
func (m *TorpedoTestMetaData) SetTestID(id string) *TorpedoTestMetaData {
	m.TestID = id
	return m
}

// GetTestUID returns the TorpedoTest UID
func (m *TorpedoTestMetaData) GetTestUID() string {
	return m.GetTestID()
}

// NewTorpedoTestMetaData creates a new instance of the TorpedoTestMetaData
func NewTorpedoTestMetaData(testID string) *TorpedoTestMetaData {
	newTorpedoTestMetaData := &TorpedoTestMetaData{}
	newTorpedoTestMetaData.SetTestID(testID)
	return newTorpedoTestMetaData
}

// NewDefaultTorpedoTestMetaData creates a new instance of the TorpedoTestMetaData with default values
func NewDefaultTorpedoTestMetaData() *TorpedoTestMetaData {
	return NewTorpedoTestMetaData(DefaultTestID)
}
