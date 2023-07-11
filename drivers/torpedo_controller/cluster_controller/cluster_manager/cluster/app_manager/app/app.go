package app

import (
	"github.com/portworx/torpedo/drivers/scheduler"
	. "github.com/portworx/torpedo/drivers/torpedo_controller/torpedo_utils/entity_generics"
)

type App[S CloudCredentialSpec] struct {
	Spec     S
	Contexts []*scheduler.Context
}

// GetSpec returns the Spec associated with the App
func (a *App[S]) GetSpec() S {
	return a.Spec
}

// SetSpec sets the Spec for the App
func (a *App[S]) SetSpec(spec S) *App[S] {
	a.Spec = spec
	return a
}

func (a *App[S]) GetContexts() []*scheduler.Context {
	return a.Contexts
}

func (a *App[S]) SetContexts(contexts []*scheduler.Context) *App[S] {
	a.Contexts = contexts
	return a
}
