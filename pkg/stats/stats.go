package stats

import (
	"encoding/json"
	"fmt"

	"github.com/portworx/torpedo/pkg/aetosutil"
	"github.com/portworx/torpedo/pkg/log"
)

var dash *aetosutil.Dashboard

const (
	rebootEventName = "Node Reboot"
)

// Add more fields here if required
type NodeRebootStatsType struct {
	RebootTime string
	Node       string
	PxVersion  string
}

func getRebootStats(rebootTime, nodeID, pxVersion string) (map[string]string, error) {
	rebootStats := &NodeRebootStatsType{
		RebootTime: rebootTime,
		Node:       nodeID,
		PxVersion:  pxVersion,
	}

	data, _ := json.Marshal(rebootStats)
	rebootExportable := make(map[string]string)
	json.Unmarshal(data, &rebootExportable)
	log.InfoD("Reboot Stats are: %v", rebootExportable)
	return rebootExportable, nil
}

func PushStats(eventType interface{}) error {
	dash = aetosutil.Get()
	var exportableData map[string]string
	var err error
	// TODO: implement this for all eventTypes not just reboots
	if obj, ok := eventType.(*NodeRebootStatsType); ok {
		//  TODO: Here exportableData.PxVersion may be replaced by the current release for which this is being run
		pxVersion := obj.PxVersion
		exportableData, err = getRebootStats(obj.RebootTime, obj.Node, pxVersion)
		if err != nil {
			return err
		}
		dash.IsEnabled = true
		fmt.Printf("Pushing stats: %v", dash.IsEnabled)
		dash.UpdateStats("longevity", "SSIE", "reboot", pxVersion, exportableData)
	} else {
		fmt.Printf("Object not identified")
	}
	return nil
}
