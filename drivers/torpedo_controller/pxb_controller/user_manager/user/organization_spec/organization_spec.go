package organization_spec

const (
	// DefaultName is the default Name for OrganizationSpec
	DefaultName = "default"
)

// OrganizationSpec represents the specification for an Organization
type OrganizationSpec struct {
	Name string
}

// GetName returns the Name associated with the OrganizationSpec
func (s *OrganizationSpec) GetName() string {
	return s.Name
}

// SetName sets the Name for the OrganizationSpec
func (s *OrganizationSpec) SetName(name string) *OrganizationSpec {
	s.Name = name
	return s
}

// NewOrganizationSpec creates a new instance of the OrganizationSpec
func NewOrganizationSpec(name string) *OrganizationSpec {
	organizationSpec := &OrganizationSpec{}
	organizationSpec.SetName(name)
	return organizationSpec
}

// NewDefaultOrganizationSpec creates a new instance of the OrganizationSpec with default values
func NewDefaultOrganizationSpec() *OrganizationSpec {
	return NewOrganizationSpec(DefaultName)
}
