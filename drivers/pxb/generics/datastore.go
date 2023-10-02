package generics

import "sync"

// DataStore represents a generic thread-safe DataStore
type DataStore[E interface{}] struct {
	sync.RWMutex
	PresentMap map[string]E
	RemovedMap map[string][]E
}

// GetPresentMap gets the PresentMap associated with the DataStore
func (d *DataStore[E]) GetPresentMap() map[string]E {
	d.RLock()
	defer d.RUnlock()
	return d.PresentMap
}

// SetPresentMap sets the PresentMap for the DataStore
func (d *DataStore[E]) SetPresentMap(presentMap map[string]E) *DataStore[E] {
	d.Lock()
	defer d.Unlock()
	d.PresentMap = presentMap
	return d
}

// GetRemovedMap gets the RemovedMap associated with the DataStore
func (d *DataStore[E]) GetRemovedMap() map[string][]E {
	d.RLock()
	defer d.RUnlock()
	return d.RemovedMap
}

// SetRemovedMap sets the RemovedMap for the DataStore
func (d *DataStore[E]) SetRemovedMap(removedMap map[string][]E) *DataStore[E] {
	d.Lock()
	defer d.Unlock()
	d.RemovedMap = removedMap
	return d
}

// Get gets the entity with the given UID
func (d *DataStore[E]) Get(entityUID string) E {
	d.RLock()
	defer d.RUnlock()
	return d.PresentMap[entityUID]
}

// Set sets the entity with the given UID
func (d *DataStore[E]) Set(entityUID string, entity E) *DataStore[E] {
	d.Lock()
	defer d.Unlock()
	d.PresentMap[entityUID] = entity
	return d
}

// Delete deletes the entity with the given UID
func (d *DataStore[E]) Delete(entityUID string) *DataStore[E] {
	d.Lock()
	defer d.Unlock()
	delete(d.PresentMap, entityUID)
	return d
}

// Remove removes the entity with the given UID
func (d *DataStore[E]) Remove(entityUID string) *DataStore[E] {
	d.Lock()
	defer d.Unlock()
	if entity, isPresent := d.PresentMap[entityUID]; isPresent {
		d.RemovedMap[entityUID] = append(d.RemovedMap[entityUID], entity)
		delete(d.PresentMap, entityUID)
	}
	return d
}

// IsPresent checks if the entity with the given UID is present
func (d *DataStore[E]) IsPresent(entityUID string) bool {
	d.RLock()
	defer d.RUnlock()
	_, isPresent := d.PresentMap[entityUID]
	return isPresent
}

// IsRemoved checks if the entity with the given UID is removed
func (d *DataStore[E]) IsRemoved(entityUID string) bool {
	d.RLock()
	defer d.RUnlock()
	_, isRemoved := d.RemovedMap[entityUID]
	return isRemoved
}

// IsRecorded checks if the entity with the given UID is recorded
func (d *DataStore[E]) IsRecorded(entityUID string) bool {
	d.RLock()
	defer d.RUnlock()
	_, isPresent := d.PresentMap[entityUID]
	_, isRemoved := d.RemovedMap[entityUID]
	return isPresent || isRemoved
}

// NewDataStore creates and initializes a new instance of the DataStore
func NewDataStore[E interface{}]() *DataStore[E] {
	dataStore := &DataStore[E]{}
	dataStore.SetPresentMap(make(map[string]E))
	dataStore.SetRemovedMap(make(map[string][]E))
	return dataStore
}
