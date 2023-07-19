package controller_generics

import (
	"reflect"
	"strings"
)

// EntitySpec represents the specification for an entity.Entity
type EntitySpec = any

// EntitySpecArray represents an array of entity_spec.EntitySpec
type EntitySpecArray struct {
	SpecArray []EntitySpec
}

// GetSpecArray gets the SpecArray associated with the EntitySpecArray
func (a *EntitySpecArray) GetSpecArray() []EntitySpec {
	return a.SpecArray
}

// SetSpecArray sets the SpecArray for the EntitySpecArray
func (a *EntitySpecArray) SetSpecArray(array []EntitySpec) *EntitySpecArray {
	a.SpecArray = array
	return a
}

// Append appends the given entity_spec.EntitySpec to the EntitySpecArray
func (a *EntitySpecArray) Append(spec EntitySpec) *EntitySpecArray {
	a.SpecArray = append(a.SpecArray, spec)
	return a
}

// TypeString returns the string representation of the types in the SpecArray
func (a *EntitySpecArray) TypeString() string {
	var types []string
	for _, spec := range a.SpecArray {
		specType := reflect.TypeOf(spec)
		if specType.Kind() == reflect.Ptr {
			specType = specType.Elem()
		}
		types = append(types, specType.Name())
	}
	return strings.Join(types, "/")
}

// NewEntitySpecArray creates a new instance of the EntitySpecArray
func NewEntitySpecArray(specArray []EntitySpec) *EntitySpecArray {
	array := &EntitySpecArray{}
	array.SetSpecArray(specArray)
	return array
}

// NewDefaultEntitySpecArray creates a new instance of the EntitySpecArray with default values
func NewDefaultEntitySpecArray() *EntitySpecArray {
	return NewEntitySpecArray(
		make([]EntitySpec, 0),
	)
}
