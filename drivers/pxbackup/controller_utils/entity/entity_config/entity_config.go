package entity_config

import (
	. "github.com/portworx/torpedo/drivers/pxbackup/controller_utils/entity"
	. "github.com/portworx/torpedo/drivers/pxbackup/controller_utils/entity/entity_config/entity_manager"
	. "github.com/portworx/torpedo/drivers/pxbackup/controller_utils/entity/entity_config/entity_spec"
)

// EntityConfig represents the configuration for the given entity_spec.EntitySpec and entity.Entity
type EntityConfig[S EntitySpec, E Entity] struct {
	Spec    S
	Manager *EntityManager[E]
}

// GetSpec gets the Spec associated with the EntityConfig
func (c *EntityConfig[S, E]) GetSpec() S {
	return c.Spec
}

// SetSpec sets the Spec for the EntityConfig
func (c *EntityConfig[S, E]) SetSpec(spec S) *EntityConfig[S, E] {
	c.Spec = spec
	return c
}

// GetManager gets the Manager associated with the EntityConfig
func (c *EntityConfig[S, E]) GetManager() *EntityManager[E] {
	return c.Manager
}

// SetManager sets the Manager for the EntityConfig
func (c *EntityConfig[S, E]) SetManager(manager *EntityManager[E]) *EntityConfig[S, E] {
	c.Manager = manager
	return c
}

// NewEntityConfig creates a new instance of the EntityConfig
func NewEntityConfig[S EntitySpec, E Entity](spec S, manager *EntityManager[E]) *EntityConfig[S, E] {
	entityConfig := &EntityConfig[S, E]{}
	entityConfig.SetSpec(spec)
	entityConfig.SetManager(manager)
	return entityConfig
}

// NewDefaultEntityConfig creates a new instance of the EntityConfig with default values
func NewDefaultEntityConfig[S EntitySpec, E Entity](spec S) *EntityConfig[S, E] {
	return NewEntityConfig[S, E](spec, NewDefaultEntityManager[E]())
}
