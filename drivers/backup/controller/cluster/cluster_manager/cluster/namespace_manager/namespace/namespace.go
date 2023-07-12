package namespace

import (
	. "github.com/portworx/torpedo/drivers/backup/controller/cluster/cluster_manager/cluster/namespace_spec"
)

// Namespace represents Namespace
type Namespace struct {
	NamespaceSpec *NamespaceSpec
}

// GetNamespaceSpec returns the NamespaceSpec associated with the Namespace
func (n *Namespace) GetNamespaceSpec() *NamespaceSpec {
	return n.NamespaceSpec
}

// SetNamespaceSpec sets the NamespaceSpec for the Namespace
func (n *Namespace) SetNamespaceSpec(spec *NamespaceSpec) *Namespace {
	n.NamespaceSpec = spec
	return n
}

// NewNamespace creates a new instance of the Namespace
func NewNamespace(namespaceSpec *NamespaceSpec) *Namespace {
	namespace := &Namespace{}
	namespace.SetNamespaceSpec(namespaceSpec)
	return namespace
}

// NewDefaultNamespace creates a new instance of the Namespace with default values
func NewDefaultNamespace(namespaceSpec *NamespaceSpec) *Namespace {
	return NewNamespace(namespaceSpec)
}
