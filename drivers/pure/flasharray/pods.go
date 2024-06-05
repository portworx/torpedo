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
