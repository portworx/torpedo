package app_spec

// TeardownParams represents the parameters for tearing down an App
type TeardownParams struct {
	WaitForDestroy             bool
	WaitForResourceLeakCleanup bool
	SkipClusterScopedObjects   bool
}
