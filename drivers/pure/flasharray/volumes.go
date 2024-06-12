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

func (vols *VolumeServices) ListAllAvailableVolumes(params map[string]string) ([]VolResponse, error) {
	req, err := vols.client.NewRequest("GET", "volumes", params, nil)
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

func (realms *RealmsServices) ListAllAvailableRealms(params map[string]string) ([]RealmResponse, error) {
	req, err := realms.client.NewRequest("GET", "realms", params, nil)
	if err != nil {
		return nil, err
	}
	m := []RealmResponse{}
	_, err = realms.client.Do(req, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (vols *PodServices) ListAllAvailablePods(params map[string]string) ([]PodResponse, error) {
	req, err := vols.client.NewRequest("GET", "pods", params, nil)
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
