package vpsutil

import (
	"github.com/portworx/talisman/pkg/apis/portworx/v1beta1"
	talisman_v1beta2 "github.com/portworx/talisman/pkg/apis/portworx/v1beta2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func VolumeAntiAffinityByMatchExpression(name string, matchExpression []*v1beta1.LabelSelectorRequirement) talisman_v1beta2.VolumePlacementStrategy {
	return talisman_v1beta2.VolumePlacementStrategy{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: talisman_v1beta2.VolumePlacementSpec{
			VolumeAntiAffinity: []*talisman_v1beta2.CommonPlacementSpec{
				{
					Enforcement:      v1beta1.EnforcementRequired,
					MatchExpressions: matchExpression,
				},
			},
		},
	}
}

func VolumeAffinityByMatchExpression(name string, matchExpression []*v1beta1.LabelSelectorRequirement) talisman_v1beta2.VolumePlacementStrategy {
	return talisman_v1beta2.VolumePlacementStrategy{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: talisman_v1beta2.VolumePlacementSpec{
			VolumeAffinity: []*talisman_v1beta2.CommonPlacementSpec{
				{
					Enforcement:      v1beta1.EnforcementRequired,
					MatchExpressions: matchExpression,
				},
			},
		},
	}
}
