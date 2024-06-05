package flasharray

type VolumeServices struct {
	client *Client
}
type RealmsServices struct {
	client *Client
}
type PodServices struct {
	client *Client
}

func (vols *VolumeServices) ListAllAvailableVolumes(params map[string]string, data interface{}) ([]VolResponse, error) {
	req, err := vols.client.NewRequest("GET", "volumes", params, data)
	if err != nil {
		return nil, err
	}
	m := []VolResponse{}
	_, err = vols.client.Do(req, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (realm *RealmsServices) ListAllAvailableRealms(params map[string]string, data interface{}) ([]RealmResponse, error) {
	req, err := realm.client.NewRequest("GET", "realms", params, data)
	if err != nil {
		return nil, err
	}
	m := []RealmResponse{}
	_, err = realm.client.Do(req, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (vols *PodServices) ListAllAvailablePods(params map[string]string, data interface{}) ([]PodResponse, error) {
	req, err := vols.client.NewRequest("GET", "pods", params, data)
	if err != nil {
		return nil, err
	}
	m := []PodResponse{}
	_, err = vols.client.Do(req, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
