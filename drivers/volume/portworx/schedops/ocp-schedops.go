package schedops

import "github.com/sirupsen/logrus"

// This is a subclass of k8sSchedOps
// This is needed to differentiate k8s and OCP scheduler
type ocpSchedOps struct {
	k8sSchedOps
	tpLog *logrus.Logger
}

func init() {
	k := &ocpSchedOps{}
	Register("openshift", k)
}

func (o *ocpSchedOps) Init(tpLog *logrus.Logger) {
	o.tpLog = tpLog
}
