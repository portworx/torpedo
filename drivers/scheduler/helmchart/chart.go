package helmchart

import (
	"fmt"

	"github.com/portworx/torpedo/drivers/scheduler"
)

// Parser provides operations for parsing application charts
type Parser interface {
	ParseCharts(chartDir string) (*scheduler.HelmRepo, error)
}

// AppChart defines a helm repo specification for an app
type AppChart struct {
	// Key is used by applications to register to the factory
	Key string
	// List of charts
	ChartList []*scheduler.HelmRepo
	// Enabled indicates if the application is enabled in the factory
	Enabled bool
}

// GetID returns the unique ID for the app charts
func (in *AppChart) GetID(instanceID string) string {
	return fmt.Sprintf("%s-%s", in.Key, instanceID)
}

// DeepCopy Creates a copy of the AppChart
func (in *AppChart) DeepCopy() *AppChart {
	if in == nil {
		return nil
	}
	out := new(AppChart)
	out.Key = in.Key
	out.Enabled = in.Enabled
	out.ChartList = make([]*scheduler.HelmRepo, 0)
	for _, chart := range in.ChartList {
		out.ChartList = append(out.ChartList, chart)
	}
	return out
}
