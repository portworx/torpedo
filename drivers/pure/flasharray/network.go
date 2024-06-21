package flasharray

// NetworkInterface struct for object returned by array
type NetworkServices struct {
	client *Client
}

func (n *NetworkService) EnableNetworkInterface(iface string) (*NetworkInterface, error) {

	data := map[string]bool{"enabled": true}
	m, err := n.SetNetworkInterface(iface, data)
	if err != nil {
		return nil, err
	}

	return m, err
}

// ListNetworkInterfaces list the attributes of the network interfaces
func (n *NetworkService) ListNetworkInterfaces() ([]NetworkInterface, error) {

	req, _ := n.client.NewRequest("GET", "network-interfaces", nil, nil)
	m := []NetworkInterface{}
	_, err := n.client.Do(req, &m, false)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// DisableNetworkInterface disables a network interface.
// param: iface: Name of network interface to be disabled.
// Returns an object describing the interface.
func (n *NetworkService) DisableNetworkInterface(iface string) (*NetworkInterface, error) {

	data := map[string]bool{"enabled": false}
	m, err := n.SetNetworkInterface(iface, data)
	if err != nil {
		return nil, err
	}

	return m, err
}
