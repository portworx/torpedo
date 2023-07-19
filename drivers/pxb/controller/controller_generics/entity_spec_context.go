package controller_generics

// EntitySpecContext represents the context for an Entity
type EntitySpecContext[E Entity] struct {
	SpecArray *EntitySpecArray
	DataStore *EntityDataStore[E]
}

// GetSpecArray gets the SpecArray associated with the EntitySpecContext
func (c *EntitySpecContext[E]) GetSpecArray() *EntitySpecArray {
	return c.SpecArray
}

// SetSpecArray sets the SpecArray for the EntitySpecContext
func (c *EntitySpecContext[E]) SetSpecArray(array *EntitySpecArray) *EntitySpecContext[E] {
	c.SpecArray = array
	return c
}

// GetDataStore gets the DataStore associated with the EntitySpecContext
func (c *EntitySpecContext[E]) GetDataStore() *EntityDataStore[E] {
	return c.DataStore
}

// SetDataStore sets the DataStore for the EntitySpecContext
func (c *EntitySpecContext[E]) SetDataStore(store *EntityDataStore[E]) *EntitySpecContext[E] {
	c.DataStore = store
	return c
}

// NewEntitySpecContext creates a new instance of the EntitySpecContext
func NewEntitySpecContext[E Entity](
	specArray *EntitySpecArray,
	dataStore *EntityDataStore[E],
) *EntitySpecContext[E] {
	config := &EntitySpecContext[E]{}
	config.SetSpecArray(specArray)
	config.SetDataStore(dataStore)
	return config
}

// NewDefaultEntitySpecContext creates a new instance of the EntitySpecContext with default values
func NewDefaultEntitySpecContext[E Entity]() *EntitySpecContext[E] {
	return NewEntitySpecContext[E](
		NewDefaultEntitySpecArray(),
		NewDefaultEntityDataStore[E](),
	)
}
