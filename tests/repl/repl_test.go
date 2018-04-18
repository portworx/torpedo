package tests

import (
	"fmt"
	"math"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/portworx/torpedo/drivers/scheduler"
	"github.com/portworx/torpedo/drivers/volume"
	. "github.com/portworx/torpedo/tests"
)

func TestRepl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Torpedo : Repl")
}

var _ = BeforeSuite(func() {
	InitInstance()
})

// This test performs basic test of starting an application and destroying it (along with storage)
var _ = Describe("SetupTeardown", func() {
	It("has to setup, validate and teardown apps", func() {
		var contexts []*scheduler.Context
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("setupteardown-%d", i))...)
		}

		opts := make(map[string]bool)
		opts[scheduler.OptionsWaitForResourceLeakCleanup] = true

		for _, ctx := range contexts {
			TearDownContext(ctx, opts)
		}
	})
})

// Volume Replication Decrease
var _ = Describe("VolumeReplicationDecrease", func() {
	It("has to schedule apps and decrease replication factor on all volumes of the apps", func() {
		var err error
		var contexts []*scheduler.Context
		expReplMap := make(map[*volume.Volume]int64)
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("volrepldown-%d", i))...)
		}

		Step("get volumes for all apps in test and decrease replication factor", func() {
			for _, ctx := range contexts {
				var appVolumes []*volume.Volume
				Step(fmt.Sprintf("get volumes for %s app", ctx.App.Key), func() {
					appVolumes, err = Inst().S.GetVolumes(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(appVolumes).NotTo(BeEmpty())
				})

				Step(
					fmt.Sprintf("repl decrease volume driver %s on app %s's volumes: %v",
						Inst().V.String(), ctx.App.Key, appVolumes),
					func() {
						for _, v := range appVolumes {
							errExpected := false
							currRep, err := Inst().V.GetReplicationFactor(v)
							Expect(err).NotTo(HaveOccurred())
							if currRep == 1 {
								errExpected = true
							}
							expReplMap[v] = int64(math.Min(3, float64(currRep)-1))
							err = Inst().V.SetReplicationFactor(v, currRep-1)
							if !errExpected {
								Expect(err).NotTo(HaveOccurred())
							}

						}
					})
				Step(
					fmt.Sprintf("validate successful repl decrease"),
					func() {
						time.Sleep(1 * time.Minute)
						for _, v := range appVolumes {
							newRepl, err := Inst().V.GetReplicationFactor(v)
							Expect(err).NotTo(HaveOccurred())
							Expect(newRepl).To(Equal(expReplMap[v]))
						}
					})

				ValidateContext(ctx)

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

// Volume Driver Plugin is down, unavailable - and the client container should not be impacted.
var _ = Describe("VolumeReplicationIncrease", func() {
	It("has to schedule apps and increase replication factor on all volumes of the apps", func() {
		var err error
		var contexts []*scheduler.Context
		expReplMap := make(map[*volume.Volume]int64)
		for i := 0; i < Inst().ScaleFactor; i++ {
			contexts = append(contexts, ScheduleAndValidate(fmt.Sprintf("volreplup-%d", i))...)
		}

		Step("get volumes for all apps in test and increase replication factor", func() {
			for _, ctx := range contexts {
				var appVolumes []*volume.Volume
				Step(fmt.Sprintf("get volumes for %s app", ctx.App.Key), func() {
					appVolumes, err = Inst().S.GetVolumes(ctx)
					Expect(err).NotTo(HaveOccurred())
					Expect(appVolumes).NotTo(BeEmpty())
				})

				Step(
					fmt.Sprintf("repl increase volume driver %s on app %s's volumes: %v",
						Inst().V.String(), ctx.App.Key, appVolumes),
					func() {
						//IncreaseVolumeReplication(appVolumes)
						for _, v := range appVolumes {
							errExpected := false
							currRep, err := Inst().V.GetReplicationFactor(v)
							Expect(err).NotTo(HaveOccurred())
							if currRep == 3 {
								errExpected = true
							}
							expReplMap[v] = int64(math.Min(3, float64(currRep)+1))
							err = Inst().V.SetReplicationFactor(v, currRep+1)
							if !errExpected {
								Expect(err).NotTo(HaveOccurred())
							}

						}
					})
				Step(
					fmt.Sprintf("validate successful repl increase"),
					func() {
						time.Sleep(1 * time.Minute)
						for _, v := range appVolumes {
							newRepl, err := Inst().V.GetReplicationFactor(v)
							Expect(err).NotTo(HaveOccurred())
							Expect(newRepl).To(Equal(expReplMap[v]))
						}
					})

				ValidateContext(ctx)

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
	CollectSupport()
	ValidateCleanup()
})

func init() {
	ParseFlags()
}
