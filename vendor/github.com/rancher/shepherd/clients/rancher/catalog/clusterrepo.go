package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"image/png"

	"github.com/rancher/shepherd/pkg/api/steve/catalog/types"
	scheme "github.com/rancher/shepherd/pkg/generated/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ClusterRepoSteveResourceType = "catalog.cattle.io.clusterrepo"

	action           = "action"
	chartsURL        = "v1/catalog.cattle.io.clusterrepos/"
	link             = "link"
	index            = "index"
	info             = "info"
	install          = "install"
	RancherChartRepo = "rancher-charts"
	rancherAppsURL   = "v1/catalog.cattle.io.apps/"
	upgrade          = "upgrade"
	uninstall        = "uninstall"
	chartName        = "chartName"
	icon             = "icon"
	version          = "version"
	values           = "values"
)

// GetChartValues - fetches the chart values from the given repo, chart and chartVersion and validates the result.
func (c *Client) GetChartValues(repoName, chart, chartVersion string) (map[string]interface{}, error) {
	result, err := c.RESTClient().Get().
		AbsPath(chartsURL+repoName).Param(link, info).Param(chartName, chart).Param(version, chartVersion).
		Do(context.Background()).Raw()

	if err != nil {
		return nil, err
	}

	var mapResponse map[string]interface{}
	if err = json.Unmarshal(result, &mapResponse); err != nil {
		return nil, err
	}

	values, ok := mapResponse[values].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to convert values to map[string]interface{}")
	}

	return values, nil
}

// FetchChartIcon - fetches the chart icon from the given repo and chart and validates the result.
func (c *Client) FetchChartIcon(repo, chart string) (int, error) {
	resp := c.RESTClient().Get().
		AbsPath(chartsURL+repo).
		Param(chartName, chart).Param(link, icon).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(context.Background())

	var contentType string

	resp.ContentType(&contentType)

	result, err := resp.Raw()
	if err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil // No icon
	}

	if contentType == "image/svg+xml" {
		var xmlData interface{}

		err = xml.Unmarshal(result, &xmlData)
		if err != nil {
			// If XML parsing fails, this is not a valid svg
			return 0, err
		}

		return len(result), nil // Valid SVG
	}

	// Try to decode as PNG
	_, err = png.Decode(bytes.NewReader(result))
	if err != nil {
		return 0, err // Not an valid image
	}

	// valid PNG image
	return len(result), nil
}

// GetChartsFromClusterRepo will return all the installable charts from a given cluster repo name
func (c *Client) GetChartsFromClusterRepo(repoName string) (map[string]interface{}, error) {
	result, err := c.RESTClient().Get().
		AbsPath(chartsURL+repoName).Param(link, index).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(context.Background()).Raw()

	if err != nil {
		return map[string]interface{}{}, err
	}

	var mapResponse map[string]interface{}
	if err = json.Unmarshal(result, &mapResponse); err != nil {
		return map[string]interface{}{}, err
	}

	charts, ok := mapResponse["entries"].(map[string]interface{})
	if !ok {
		return map[string]interface{}{}, fmt.Errorf("failed to convert charts to map[string]interface{}")
	}

	return charts, nil

}

// GetListChartVersions is used to get the list of versions of `chartName` from a given `repoName`
func (c *Client) GetListChartVersions(chartName, repoName string) ([]string, error) {
	result, err := c.RESTClient().Get().
		AbsPath(chartsURL+repoName).Param(link, index).
		VersionedParams(&metav1.GetOptions{}, scheme.ParameterCodec).
		Do(context.TODO()).Raw()

	if err != nil {
		return nil, err
	}

	var mapResponse map[string]interface{}
	if err = json.Unmarshal(result, &mapResponse); err != nil {
		return nil, err
	}

	entries := mapResponse["entries"]
	specifiedChartEntries := entries.(map[string]interface{})[chartName].([]interface{})
	if len(specifiedChartEntries) < 1 {
		return nil, fmt.Errorf("failed to find chart %s from the chart repo", chartName)
	}

	versionsList := []string{}
	for _, entry := range specifiedChartEntries {
		entryMap := entry.(map[string]interface{})
		versionsList = append(versionsList, entryMap[version].(string))
	}

	return versionsList, nil
}

// GetLatestChartVersion is used to get the lastest version of `chartName` from a given `repoName`
func (c *Client) GetLatestChartVersion(chartName string, repoName string) (string, error) {
	versionsList, err := c.GetListChartVersions(chartName, repoName)
	if err != nil {
		return "", err
	}
	lastestVersion := versionsList[0]

	return lastestVersion, nil
}

// InstallChart installs the chart according to the parameter `chart` from a given repoName
func (c *Client) InstallChart(chart *types.ChartInstallAction, repoName string) error {
	bodyContent, err := json.Marshal(chart)
	if err != nil {
		return err
	}

	result := c.RESTClient().Post().
		AbsPath(chartsURL+repoName).Param(action, install).
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(bodyContent).
		Do(context.TODO())
	return result.Error()
}

// UpgradeChart upgrades the chart according to the parameter `chart`
func (c *Client) UpgradeChart(chart *types.ChartUpgradeAction, repoName string) error {
	bodyContent, err := json.Marshal(chart)
	if err != nil {
		return err
	}

	result := c.RESTClient().Post().
		AbsPath(chartsURL+repoName).Param(action, upgrade).
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Body(bodyContent).
		Do(context.TODO())

	return result.Error()
}

// UninstallChart uninstalls the chart according to `chartNamespace`, `chartName`, and `uninstallAction`
func (c *Client) UninstallChart(chartName, chartNamespace string, uninstallAction *types.ChartUninstallAction) error {
	bodyContent, err := json.Marshal(uninstallAction)
	if err != nil {
		return err
	}

	url := rancherAppsURL + chartNamespace
	result := c.RESTClient().Post().
		Name(chartName).
		AbsPath(url).Param(action, uninstall).
		Body(bodyContent).
		VersionedParams(&metav1.CreateOptions{}, scheme.ParameterCodec).
		Do(context.TODO())

	return result.Error()
}
