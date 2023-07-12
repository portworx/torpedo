package namespace_spec

// NamespaceSpec represents the specification for a Namespace
type NamespaceSpec struct {
	Namespace string
}

// GetNamespace returns the Namespace associated with the NamespaceSpec
func (s *NamespaceSpec) GetNamespace() string {
	return s.Namespace
}

// SetNamespace sets the Namespace for the NamespaceSpec
func (s *NamespaceSpec) SetNamespace(namespace string) *NamespaceSpec {
	s.Namespace = namespace
	return s
}

// NewNamespaceSpec creates a new instance of the NamespaceSpec
func NewNamespaceSpec(namespace string) *NamespaceSpec {
	namespaceSpec := &NamespaceSpec{}
	namespaceSpec.SetNamespace(namespace)
	return namespaceSpec
}

// NewDefaultNamespaceSpec creates a new instance of the NamespaceSpec with default values
func NewDefaultNamespaceSpec(namespace string) *NamespaceSpec {
	return NewNamespaceSpec(namespace)
}
