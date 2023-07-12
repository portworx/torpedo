package entity_manager

import . "github.com/portworx/torpedo/drivers/pxbackup/controller_utils/entity"

// EntityManager represents the manager for an entity.Entity
type EntityManager[E Entity] struct {
	PresentMap map[string]E
	RemovedMap map[string][]E
}

// GetPresentMap gets the PresentMap associated with the EntityManager
func (m *EntityManager[E]) GetPresentMap() map[string]E {
	return m.PresentMap
}

// SetPresentMap sets the PresentMap for the EntityManager
func (m *EntityManager[E]) SetPresentMap(presentMap map[string]E) *EntityManager[E] {
	m.PresentMap = presentMap
	return m
}

// GetRemovedMap gets the RemovedMap associated with the EntityManager
func (m *EntityManager[E]) GetRemovedMap() map[string][]E {
	return m.RemovedMap
}

// SetRemovedMap sets the RemovedMap for the EntityManager
func (m *EntityManager[E]) SetRemovedMap(removedMap map[string][]E) *EntityManager[E] {
	m.RemovedMap = removedMap
	return m
}

// Get gets the entity.Entity with the given entity.Entity UID
func (m *EntityManager[E]) Get(entityUID string) E {
	return m.PresentMap[entityUID]
}

// Set sets the entity.Entity with the given entity.Entity UID
func (m *EntityManager[E]) Set(entityUID string, entity E) *EntityManager[E] {
	m.PresentMap[entityUID] = entity
	return m
}

// Delete deletes the entity.Entity with the given entity.Entity UID
func (m *EntityManager[E]) Delete(entityUID string) *EntityManager[E] {
	delete(m.PresentMap, entityUID)
	return m
}

// Remove removes the entity.Entity with the given entity.Entity UID
func (m *EntityManager[E]) Remove(entityUID string) *EntityManager[E] {
	if entity, isPresent := m.PresentMap[entityUID]; isPresent {
		m.RemovedMap[entityUID] = append(m.RemovedMap[entityUID], entity)
		delete(m.PresentMap, entityUID)
	}
	return m
}

// IsPresent checks if the entity.Entity with the given entity.Entity UID IsPresent
func (m *EntityManager[E]) IsPresent(entityUID string) bool {
	_, isPresent := m.PresentMap[entityUID]
	return isPresent
}

// IsRemoved checks if the entity.Entity with the given entity.Entity UID IsRemoved
func (m *EntityManager[E]) IsRemoved(entityUID string) bool {
	_, isRemoved := m.RemovedMap[entityUID]
	return isRemoved
}

// IsRecorded checks if the entity.Entity with the given entity.Entity UID IsRecorded
func (m *EntityManager[E]) IsRecorded(entityUID string) bool {
	_, isPresent := m.PresentMap[entityUID]
	_, isRemoved := m.RemovedMap[entityUID]
	return isPresent || isRemoved
}

// NewEntityManager creates a new instance of the EntityManager
func NewEntityManager[E Entity](presentMap map[string]E, removedMap map[string][]E) *EntityManager[E] {
	entityManager := &EntityManager[E]{}
	entityManager.SetPresentMap(presentMap)
	entityManager.SetRemovedMap(removedMap)
	return entityManager
}

// NewDefaultEntityManager creates a new instance of the EntityManager with default values
func NewDefaultEntityManager[E Entity]() *EntityManager[E] {
	return NewEntityManager[E](make(map[string]E, 0), make(map[string][]E, 0))
}
