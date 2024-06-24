package flasharray

// NetworkInterface struct for object returned by array
type NetworkServices struct {
	client *Client
}

// SetNetworkInterface modifies network interface attributes
func (n *NetworkServices) SetNetworkInterface(params map[string]string, data interface{}) ([]NetworkInterface, error) {

	req, _ := n.client.NewRequest("PATCH", "network-interfaces", params, data)
	m := []NetworkInterface{}
	_, err := n.client.Do(req, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
func (n *NetworkServices) EnableNetworkInterface(iface string) ([]NetworkInterface, error) {

	params := make(map[string]string)
	params["names"] = iface
	data := map[string]bool{"enabled": true}
	m, err := n.SetNetworkInterface(params, data)
	if err != nil {
		return nil, err
	}

	return m, err
}

// ListNetworkInterfaces list the attributes of the network interfaces
func (n *NetworkServices) ListNetworkInterfaces() ([]NetworkInterfaceResponse, error) {

	req, _ := n.client.NewRequest("GET", "network-interfaces", nil, nil)
	m := []NetworkInterfaceResponse{}
	_, err := n.client.Do(req, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// DisableNetworkInterface disables a network interface.
// param: iface: Name of network interface to be disabled.
// Returns an object describing the interface.
func (n *NetworkServices) DisableNetworkInterface(iface string) ([]NetworkInterface, error) {

	params := make(map[string]string)
	params["names"] = iface
	data := map[string]bool{"enabled": false}
	m, err := n.SetNetworkInterface(params, data)
	if err != nil {
		return nil, err
	}

	return m, err
}
