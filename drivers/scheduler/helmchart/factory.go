package helmchart

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Factory is an application chart factory
type Factory struct {
	chartDir    string
	chartParser Parser
}

var appChartFactory = make(map[string]*AppChart)

// register registers a new chart with the factory
func (f *Factory) register(id string, app *AppChart) {
	if _, ok := appChartFactory[id]; !ok {
		logrus.Infof("Registering new app: %v", id)
	} else {
		logrus.Infof("Substitute with new app: %v", id)
	}
	// NOTE: In case of chart rescan we need to substitute old app with another one
	appChartFactory[id] = app
}

// Get returns a registered application
func (f *Factory) Get(id string) (*AppChart, error) {
	if d, ok := appChartFactory[id]; ok && d.Enabled {
		if copy := d.DeepCopy(); copy != nil {
			return copy, nil
		}
		return nil, fmt.Errorf("error creating copy of app: %v", d)
	}

	return nil, &errors.ErrNotFound{
		ID:   id,
		Type: "AppChart",
	}
}

// GetAll returns all registered enabled applications
func (f *Factory) GetAll() []*AppChart {
	var charts []*AppChart
	for _, val := range appChartFactory {
		if val.Enabled {
			valCopy := val.DeepCopy()
			if valCopy != nil {
				charts = append(charts, valCopy)
			}
		}
	}

	return charts
}

// NewFactory creates a new chart factory
func NewFactory(chartDir, storageProvisioner string, parser Parser) (*Factory, error) {
	f := &Factory{
		chartDir:    chartDir,
		chartParser: parser,
	}

	appDirList, err := ioutil.ReadDir(f.chartDir)
	if err != nil {
		return nil, err
	}

	for _, file := range appDirList {
		if file.IsDir() {
			chartID := file.Name()
			chartToParse := path.Join(f.chartDir, chartID)
			logrus.Infof("Parsing: %v...", path.Join(f.chartDir, chartID))
			chart, err := f.chartParser.ParseCharts(chartToParse)
			if err != nil {
				return nil, err
			}

			// Register the chart
			f.register(chartID, &AppChart{
				Key:       chartID,
				ChartList: []*scheduler.HelmRepo{chart},
				Enabled:   true,
			})
		}
	}

	if apps := f.GetAll(); len(apps) == 0 {
		return nil, fmt.Errorf("found 0 supported applications in given chartDir: %v", chartDir)
	}

	return f, nil
}
