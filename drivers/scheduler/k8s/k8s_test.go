package k8s

import (
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/scheduler/spec"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestK8s_SubstituteNamespaceInSpec(t *testing.T) {
	namespaceBefore := "before"
	namespaceAfter := "after"

	namespaceMapping := map[string]string{
		namespaceBefore: namespaceAfter,
	}

	obj1 := coreV1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespaceBefore,
		},
	}
	obj2 := coreV1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespaceBefore,
		},
	}

	k := K8s{}
		ctx := &scheduler.Context{
			App: &spec.AppSpec{
				SpecList: []interface{}{
					&obj1,
					&obj2,
				},
			},
		}

		err := k.SubstituteNamespaceInSpec(ctx, namespaceMapping)

		if err != nil {
			t.Errorf("unexpected error  %v", err)
		}

		if obj1.Namespace != namespaceAfter {
			t.Errorf("expected namespace %s actual %s",
				namespaceAfter, obj1.Namespace)
			return
		}

		if obj2.Namespace != namespaceAfter {
			t.Errorf("expected namespace %s actual %s",
				namespaceAfter, obj2.Namespace)
		}
}
