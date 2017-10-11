package schedops

import (
	"fmt"
	"strings"
)

// ErrFailedToCleanupVolume error type for orphan pods or unclean vol directories
type ErrFailedToCleanupVolume struct {
	// OrphanPods is a map of node to list of pod UIDs whose portworx volume dir is not deleted
	OrphanPods map[string][]string
	// DirtyVolPods is a map of node to list of pod UIDs which still has data written
	// under the volume mount point
	DirtyVolPods map[string][]string
}

func (e *ErrFailedToCleanupVolume) Error() string {
	var cause []string
	for node, pods := range e.OrphanPods {
		cause = append(cause, fmt.Sprintf("Failed to remove orphan volume dir on "+
			"node: %v for pods: %v", node, pods))
	}
	for node, pods := range e.DirtyVolPods {
		cause = append(cause, fmt.Sprintf("Failed to cleanup data under volume directory on "+
			"node: %v for pods: %v", node, pods))
	}
	return strings.Join(cause, ", ")
}

// ErrLabelsMissingOnNode error type for missing volume labels on node
type ErrLabelsMissingOnNode struct {
	// Nodes is a list of node names which have missing labels for certain PVCs
	Nodes []string
}

func (e *ErrLabelsMissingOnNode) Error() string {
	return fmt.Sprintf("Labels missing on nodes %v", e.Nodes)
}
