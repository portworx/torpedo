package entity_manager

// Entity represents an Entity
type Entity any

// EntityManager represents a manager for Entity
type EntityManager[E Entity] struct {
	PresentEntityMap map[string]*E
	RemovedEntityMap map[string][]*E
}

// GetPresentMap returns the PresentEntityMap of the EntityManager
func (m *EntityManager[E]) GetPresentMap() map[string]*E {
	return m.PresentEntityMap
}

// SetPresentMap sets the PresentEntityMap of the EntityManager
func (m *EntityManager[E]) SetPresentMap(presentEntityMap map[string]*E) *EntityManager[E] {
	m.PresentEntityMap = presentEntityMap
	return m
}

// GetRemovedMap returns the RemovedEntityMap of the EntityManager
func (m *EntityManager[E]) GetRemovedMap() map[string][]*E {
	return m.RemovedEntityMap
}

// SetRemovedMap sets the RemovedEntityMap of the EntityManager
func (m *EntityManager[E]) SetRemovedMap(removedEntityMap map[string][]*E) *EntityManager[E] {
	m.RemovedEntityMap = removedEntityMap
	return m
}

// Get returns the Entity with the given Entity UID
func (m *EntityManager[E]) Get(entityUID string) *E {
	return m.PresentEntityMap[entityUID]
}

// Set sets the Entity with the given Entity UID
func (m *EntityManager[E]) Set(entityUID string, entity *E) *EntityManager[E] {
	m.PresentEntityMap[entityUID] = entity
	return m
}

// Delete deletes the Entity with the given Entity UID
func (m *EntityManager[E]) Delete(entityUID string) *EntityManager[E] {
	delete(m.PresentEntityMap, entityUID)
	return m
}

// Remove removes the Entity with the given Entity UID
func (m *EntityManager[E]) Remove(entityUID string) *EntityManager[E] {
	if entity, isPresent := m.PresentEntityMap[entityUID]; isPresent {
		m.RemovedEntityMap[entityUID] = append(m.RemovedEntityMap[entityUID], entity)
		delete(m.PresentEntityMap, entityUID)
	}
	return m
}

// IsPresent checks if the Entity with the given Entity UID is present
func (m *EntityManager[E]) IsPresent(entityUID string) bool {
	_, isPresent := m.PresentEntityMap[entityUID]
	return isPresent
}

// IsRemoved checks if the Entity with the given Entity UID is removed
func (m *EntityManager[E]) IsRemoved(entityUID string) bool {
	_, isRemoved := m.RemovedEntityMap[entityUID]
	return isRemoved
}

// IsRecorded checks if the Entity with the given Entity UID is recorded
func (m *EntityManager[E]) IsRecorded(entityUID string) bool {
	_, isPresent := m.PresentEntityMap[entityUID]
	_, isRemoved := m.RemovedEntityMap[entityUID]
	return isPresent || isRemoved
}

// NewManager creates a new instance of the EntityManager
func NewManager[T any](presentEntityMap map[string]*T, removedEntityMap map[string][]*T) *EntityManager[T] {
	newManager := &EntityManager[T]{}
	newManager.SetPresentMap(presentEntityMap)
	newManager.SetRemovedMap(removedEntityMap)
	return newManager
}

// NewDefaultManager creates a new instance of the EntityManager with default values
func NewDefaultManager[T any]() *EntityManager[T] {
	return NewManager[T](make(map[string]*T, 0), make(map[string][]*T, 0))
}
