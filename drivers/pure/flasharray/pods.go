package flasharray

import (
	"github.com/portworx/torpedo/pkg/log"
)

func (p *PodServices) CreatePod(params map[string]string, data interface{}) (*[]PodResponse, error) {
	req, _ := p.client.NewRequest("POST", "pods", params, data)
	m := &[]PodResponse{}
	_, err := p.client.Do(req, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
func (p *PodServices) PatchPod(params map[string]string, data interface{}) (*[]PodResponse, error) {
	req, _ := p.client.NewRequest("PATCH", "pods", params, data)
	m := &[]PodResponse{}
	_, err := p.client.Do(req, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
func (p *PodServices) DeletePod(Patchparams, deleteParams map[string]string, data interface{}) error {
	podinfo, err := p.PatchPod(Patchparams, data)
	if err != nil {
		return err

	}
	for _, poditems := range *podinfo {
		for _, pod := range poditems.Items {
			log.InfoD("Pod [%v] patched successfully", pod.Name)
		}
	}
	req, _ := p.client.NewRequest("DELETE", "pods", deleteParams, data)
	m := &[]PodResponse{}
	_, err = p.client.Do(req, m)
	if err != nil {
		return err
	}

	return nil
}
