package stats

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRebootStats(t *testing.T) {
	exportableReboot, err := PushRebootStats(time.Now().TimeNow().Format("2006-01-02 15:04:05"), "testNode")
	require.NoError(t, err, "failed to create exportable stats")

	fmt.Printf("Exprtable Stats")
}
