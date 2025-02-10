package schedops

// This is a subclass of k8sSchedOps
// This is needed to differentiate k8s and ROSA scheduler
type rosaSchedOps struct {
	k8sSchedOps
}

func init() {
	i := &rosaSchedOps{}
	Register("rosa", i)
}
