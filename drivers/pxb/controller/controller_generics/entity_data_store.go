package controller_generics

// Entity represents an Entity
type Entity interface {
	Remove() error
}

// EntityDataStore represents a data store for an entity.Entity
type EntityDataStore[E Entity] struct {
	PresentMap map[string]E
	RemovedMap map[string][]E
}

// GetPresentMap gets the PresentMap associated with the EntityDataStore
func (s *EntityDataStore[E]) GetPresentMap() map[string]E {
	return s.PresentMap
}

// SetPresentMap sets the PresentMap for the EntityDataStore
func (s *EntityDataStore[E]) SetPresentMap(presentMap map[string]E) *EntityDataStore[E] {
	s.PresentMap = presentMap
	return s
}

// GetRemovedMap gets the RemovedMap associated with the EntityDataStore
func (s *EntityDataStore[E]) GetRemovedMap() map[string][]E {
	return s.RemovedMap
}

// SetRemovedMap sets the RemovedMap for the EntityDataStore
func (s *EntityDataStore[E]) SetRemovedMap(removedMap map[string][]E) *EntityDataStore[E] {
	s.RemovedMap = removedMap
	return s
}

// Get gets the entity.Entity with the given entity.Entity UID
func (s *EntityDataStore[E]) Get(entityUID string) E {
	return s.PresentMap[entityUID]
}

// Set sets the entity.Entity with the given entity.Entity UID
func (s *EntityDataStore[E]) Set(entityUID string, entity E) *EntityDataStore[E] {
	s.PresentMap[entityUID] = entity
	return s
}

// Delete deletes the entity.Entity with the given entity.Entity UID
func (s *EntityDataStore[E]) Delete(entityUID string) *EntityDataStore[E] {
	delete(s.PresentMap, entityUID)
	return s
}

// Remove removes the entity.Entity with the given entity.Entity UID
func (s *EntityDataStore[E]) Remove(entityUID string) *EntityDataStore[E] {
	if entity, isPresent := s.PresentMap[entityUID]; isPresent {
		s.RemovedMap[entityUID] = append(s.RemovedMap[entityUID], entity)
		delete(s.PresentMap, entityUID)
	}
	return s
}

// IsPresent checks if the entity.Entity with the given entity.Entity UID is present
func (s *EntityDataStore[E]) IsPresent(entityUID string) bool {
	_, isPresent := s.PresentMap[entityUID]
	return isPresent
}

// IsRemoved checks if the entity.Entity with the given entity.Entity UID is removed
func (s *EntityDataStore[E]) IsRemoved(entityUID string) bool {
	_, isRemoved := s.RemovedMap[entityUID]
	return isRemoved
}

// IsRecorded checks if the entity.Entity with the given entity.Entity UID is recorded
func (s *EntityDataStore[E]) IsRecorded(entityUID string) bool {
	_, isPresent := s.PresentMap[entityUID]
	_, isRemoved := s.RemovedMap[entityUID]
	return isPresent || isRemoved
}

// NewEntityDataStore creates a new instance of the EntityDataStore
func NewEntityDataStore[E Entity](
	presentMap map[string]E,
	removedMap map[string][]E,
) *EntityDataStore[E] {
	dataStore := &EntityDataStore[E]{}
	dataStore.SetPresentMap(presentMap)
	dataStore.SetRemovedMap(removedMap)
	return dataStore
}

// NewDefaultEntityDataStore creates a new instance of the EntityDataStore with default values
func NewDefaultEntityDataStore[E Entity]() *EntityDataStore[E] {
	return NewEntityDataStore[E](
		make(map[string]E, 0),
		make(map[string][]E, 0),
	)
}
