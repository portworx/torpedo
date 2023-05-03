package pxbackup

import (
	"github.com/pborman/uuid"
	api "github.com/portworx/px-backup-api/pkg/apis/v1"
)

type SchedulePolicyType int32

const (
	Interval = iota
	Daily
	Weekly
	Monthly
)

type Weekday string

const (
	Monday    Weekday = "Mon"
	Tuesday           = "Tue"
	Wednesday         = "Wed"
	Thursday          = "Thu"
	Friday            = "Fri"
	Saturday          = "Sat"
	Sunday            = "Sun"
)

type SchedulePolicyInfo struct {
	*api.SchedulePolicyObject
}

func (p *PxbController) setSchedulePolicyInfo(schedulePolicyName string, schedulePolicyInfo *SchedulePolicyInfo) {
	if p.organizations[p.currentOrgId].schedulePolicies == nil {
		p.organizations[p.currentOrgId].schedulePolicies = make(map[string]*SchedulePolicyInfo, 0)
	}
	p.organizations[p.currentOrgId].schedulePolicies[schedulePolicyName] = schedulePolicyInfo
}

func (p *PxbController) GetSchedulePolicyInfo(schedulePolicyName string) (*SchedulePolicyInfo, bool) {
	schedulePolicyInfo, ok := p.organizations[p.currentOrgId].schedulePolicies[schedulePolicyName]
	if !ok {
		return nil, false
	}
	return schedulePolicyInfo, true
}

func (p *PxbController) delSchedulePolicyInfo(schedulePolicyName string) {
	delete(p.organizations[p.currentOrgId].schedulePolicies, schedulePolicyName)
}

type AddSchedulePolicyConfig struct {
	schedulePolicyName string
	schedulePolicyType SchedulePolicyType
	retain             int64
	minutes            int64
	time               string
	incrCount          uint64
	date               int64
	weekDay            Weekday
	schedulePolicyUid  string         // default
	controller         *PxbController // fixed
}

func (c *AddSchedulePolicyConfig) SetSchedulePolicyUid(schedulePolicyUid string) *AddSchedulePolicyConfig {
	c.schedulePolicyUid = schedulePolicyUid
	return c
}

func (p *PxbController) IntervalSchedulePolicy(schedulePolicyName string, retain int64, minutes int64, incrCount uint64) *AddSchedulePolicyConfig {
	return &AddSchedulePolicyConfig{
		schedulePolicyName: schedulePolicyName,
		schedulePolicyType: Interval,
		retain:             retain,
		minutes:            minutes,
		incrCount:          incrCount,
		schedulePolicyUid:  uuid.New(),
		controller:         p,
	}
}

func (p *PxbController) DailySchedulePolicy(schedulePolicyName string, retain int64, time string, incrCount uint64) *AddSchedulePolicyConfig {
	return &AddSchedulePolicyConfig{
		schedulePolicyName: schedulePolicyName,
		schedulePolicyType: Daily,
		retain:             retain,
		time:               time,
		incrCount:          incrCount,
		schedulePolicyUid:  uuid.New(),
		controller:         p,
	}
}

func (p *PxbController) WeeklySchedulePolicy(schedulePolicyName string, retain int64, weekDay Weekday, time string, incrCount uint64) *AddSchedulePolicyConfig {
	return &AddSchedulePolicyConfig{
		schedulePolicyName: schedulePolicyName,
		schedulePolicyType: Weekly,
		retain:             retain,
		time:               time,
		weekDay:            weekDay,
		incrCount:          incrCount,
		schedulePolicyUid:  uuid.New(),
		controller:         p,
	}
}

func (p *PxbController) MonthlySchedulePolicy(schedulePolicyName string, retain int64, date int64, time string, incrCount uint64) *AddSchedulePolicyConfig {
	return &AddSchedulePolicyConfig{
		schedulePolicyName: schedulePolicyName,
		schedulePolicyType: Monthly,
		retain:             retain,
		time:               time,
		date:               date,
		incrCount:          incrCount,
		schedulePolicyUid:  uuid.New(),
		controller:         p,
	}
}

func (c *AddSchedulePolicyConfig) Add() error {
	var schedulePolicyInfo *api.SchedulePolicyInfo
	switch c.schedulePolicyType {
	case Interval:
		schedulePolicyInfo = &api.SchedulePolicyInfo{
			Interval: &api.SchedulePolicyInfo_IntervalPolicy{
				Retain:  c.retain,
				Minutes: c.minutes,
				IncrementalCount: &api.SchedulePolicyInfo_IncrementalCount{
					Count: c.incrCount,
				},
			},
		}
	case Daily:
		schedulePolicyInfo = &api.SchedulePolicyInfo{
			Daily: &api.SchedulePolicyInfo_DailyPolicy{
				Retain: c.retain,
				Time:   c.time,
				IncrementalCount: &api.SchedulePolicyInfo_IncrementalCount{
					Count: c.incrCount,
				},
			},
		}
	case Weekly:
		schedulePolicyInfo = &api.SchedulePolicyInfo{
			Weekly: &api.SchedulePolicyInfo_WeeklyPolicy{
				Retain: c.retain,
				Day:    string(c.weekDay),
				Time:   c.time,
				IncrementalCount: &api.SchedulePolicyInfo_IncrementalCount{
					Count: c.incrCount,
				},
			},
		}
	case Monthly:
		schedulePolicyInfo = &api.SchedulePolicyInfo{
			Monthly: &api.SchedulePolicyInfo_MonthlyPolicy{
				Retain: c.retain,
				Date:   c.date,
				Time:   c.time,
				IncrementalCount: &api.SchedulePolicyInfo_IncrementalCount{
					Count: c.incrCount,
				},
			},
		}
	}

	schedulePolicyCreateRequest := &api.SchedulePolicyCreateRequest{
		CreateMetadata: &api.CreateMetadata{
			Name:  c.schedulePolicyName,
			Uid:   c.schedulePolicyUid,
			OrgId: c.controller.currentOrgId,
		},
		SchedulePolicy: schedulePolicyInfo,
	}
	if _, err := c.controller.processPxBackupRequest(schedulePolicyCreateRequest); err != nil {
		return err
	}
	schedulePolicyInspectReq := &api.SchedulePolicyInspectRequest{
		OrgId: c.controller.currentOrgId,
		Name:  c.schedulePolicyName,
		Uid:   c.schedulePolicyUid,
	}
	resp, err := c.controller.processPxBackupRequest(schedulePolicyInspectReq)
	if err != nil {
		return err
	}
	schedulePolicy := resp.(*api.SchedulePolicyInspectResponse).GetSchedulePolicy()
	c.controller.setSchedulePolicyInfo(c.schedulePolicyName, &SchedulePolicyInfo{
		SchedulePolicyObject: schedulePolicy,
	})
	return nil
}

func (p *PxbController) DeleteSchedulePolicy(schedulePolicyName string) error {
	schedulePolicyInfo, ok := p.GetSchedulePolicyInfo(schedulePolicyName)
	if ok {
		schedulePolicyDeleteReq := &api.SchedulePolicyDeleteRequest{
			Name:  schedulePolicyName,
			OrgId: p.currentOrgId,
			Uid:   schedulePolicyInfo.GetUid(),
		}
		if _, err := p.processPxBackupRequest(schedulePolicyDeleteReq); err != nil {
			return err
		}
		p.delSchedulePolicyInfo(schedulePolicyName)
	}
	return nil
}
