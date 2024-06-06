package flasharray

func (p *PodServices) CreatePod(params map[string]string, data interface{}) (*[]PodResponse, error) {
	req, _ := p.client.NewRequest("POST", "pods", params, data)
	m := &[]PodResponse{}
	_, err := p.client.Do(req, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
func (p *PodServices) DeletePod(Patchparams, deleteParams map[string]string, data interface{}) error {
	req, _ := p.client.NewRequest("PATCH", "pods", Patchparams, data)
	m := &[]PodResponse{}
	_, err := p.client.Do(req, m)
	if err != nil {
		return err
	}
	req, _ = p.client.NewRequest("DELETE", "pods", deleteParams, nil)
	m = &[]PodResponse{}
	_, err = p.client.Do(req, m)
	if err != nil {
		return err
	}

	return nil
}
