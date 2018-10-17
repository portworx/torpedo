package tests

import (
	"fmt"
	"math"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	. "github.com/portworx/torpedo/tests"
	"github.com/portworx/torpedo/drivers/node"
)

func TestVolOps(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Torpedo : VolOps")
}

var _ = BeforeSuite(func() {
	InitInstance()
})


// Volume replication change
var _ = Describe("{VolumeReplicationChange}", func() {
	It("has to schedule apps and increase/decrease replication factor on all volumes of the apps", func() {
		var err error
		var contexts []*scheduler.Context
		expReplMap := make(map[*volume.Volume]int64)
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("volreplup-%d", i))...)
		}

		Step("get volumes for all apps in test and increase/decrease replication factor", func() {
			for _, ctx := range contexts {
				var appVolumes []*volume.Volume
				Step(fmt.Sprintf("get volumes for %s app", ctx.App.Key), func() {
					appVolumes, err = Inst().S.GetVolumes(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(appVolumes).NotTo(BeEmpty())
				})
				MaxRF := Inst().V.GetMaxReplicationFactor()
				MinRF := Inst().V.GetMinReplicationFactor()
				for _, v := range appVolumes {
					Step(
						fmt.Sprintf("repl decrease volume driver %s on app %s's volume: %v",
							Inst().V.String(), ctx.App.Key, v),
						func() {
							errExpected := false
							currRep, err := Inst().V.GetReplicationFactor(v)
							Expect(err).NotTo(HaveOccurred())
							if currRep == MinRF {
								errExpected = true
							}
							expReplMap[v] = int64(math.Max(float64(MinRF), float64(currRep)-1))
							err = Inst().V.SetReplicationFactor(v, currRep-1)
							if !errExpected {
								Expect(err).NotTo(HaveOccurred())
							} else {
								Expect(err).To(HaveOccurred())
							}

						})
					Step(
						fmt.Sprintf("validate successful repl decrease on app %s's volume: %v",
							ctx.App.Key, v),
						func() {
							newRepl, err := Inst().V.GetReplicationFactor(v)
							Expect(err).NotTo(HaveOccurred())
							Expect(newRepl).To(Equal(expReplMap[v]))
						})
					Step(
						fmt.Sprintf("repl increase volume driver %s on app %s's volume: %v",
							Inst().V.String(), ctx.App.Key, v),
						func() {
							errExpected := false
							currRep, err := Inst().V.GetReplicationFactor(v)
							Expect(err).NotTo(HaveOccurred())
							currAggr, err := Inst().V.GetAggregationLevel(v)
							Expect(err).NotTo(HaveOccurred())
							if currAggr > 1 {
								MaxRF = int64(len(node.GetWorkerNodes())) / currAggr
							}
							if currRep == MaxRF {
								errExpected = true
							}
							expReplMap[v] = int64(math.Min(float64(MaxRF), float64(currRep)+1))
							err = Inst().V.SetReplicationFactor(v, currRep+1)
							if !errExpected {
								Expect(err).NotTo(HaveOccurred())
							} else {
								Expect(err).To(HaveOccurred())
							}
						})
					Step(
						fmt.Sprintf("validate successful repl increase on app %s's volume: %v",
							ctx.App.Key, v),
						func() {
							newRepl, err := Inst().V.GetReplicationFactor(v)
							Expect(err).NotTo(HaveOccurred())
							Expect(newRepl).To(Equal(expReplMap[v]))
						})
				}
			}
		})

		Step("destroy apps", func() {
			opts := make(map[string]bool)
			opts[scheduler.OptionsWaitForResourceLeakCleanup] = true
			for _, ctx := range contexts {
				TearDownContext(ctx, opts)
			}
		})

	})
})

var _ = AfterSuite(func() {
	PerformSystemCheck()
	CollectSupport()
	ValidateCleanup()
})

func init() {
	ParseFlags()
}
