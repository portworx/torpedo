package pxbackup

//type RuleInfo struct {
//	*api.RuleObject
//}
//
//func (p *PxBackupController) setRuleInfo(ruleName string, ruleInfo *RuleInfo) {
//	if p.organizations[p.currentOrgId].rules == nil {
//		p.organizations[p.currentOrgId].rules = make(map[string]*RuleInfo, 0)
//	}
//	p.organizations[p.currentOrgId].rules[ruleName] = ruleInfo
//}
//
//func (p *PxBackupController) GetRuleInfo(ruleName string) (*RuleInfo, bool) {
//	ruleInfo, ok := p.organizations[p.currentOrgId].rules[ruleName]
//	if !ok {
//		return nil, false
//	}
//	return ruleInfo, true
//}
//
//func (p *PxBackupController) delRuleInfo(ruleName string) {
//	delete(p.organizations[p.currentOrgId].rules, ruleName)
//}
//
//type AddRuleConfig struct {
//	ruleName   string
//	rulesInfo  *api.RulesInfo
//	ruleUid    string         // default
//	controller *PxBackupController // fixed
//}
//
//func (c *AddRuleConfig) SetRuleUid(ruleUid string) *AddRuleConfig {
//	c.ruleUid = ruleUid
//	return c
//}
//
//func (p *PxBackupController) Rule(ruleName string, rulesInfo *api.RulesInfo) *AddRuleConfig {
//	return &AddRuleConfig{
//		ruleName:   ruleName,
//		rulesInfo:  rulesInfo,
//		ruleUid:    uuid.New(),
//		controller: p,
//	}
//}
//
//func (c *AddRuleConfig) Add() error {
//	ruleCreateReq := &api.RuleCreateRequest{
//		CreateMetadata: &api.CreateMetadata{
//			Name:  c.ruleName,
//			OrgId: c.controller.currentOrgId,
//			Uid:   c.ruleUid,
//		},
//		RulesInfo: c.rulesInfo,
//	}
//
//	_, err := c.controller.processPxBackupRequest(ruleCreateReq)
//	if err != nil {
//		return err
//	}
//	ruleInspectReq := &api.RuleInspectRequest{
//		OrgId: c.controller.currentOrgId,
//		Name:  c.ruleName,
//		Uid:   c.ruleUid,
//	}
//	resp, err := c.controller.processPxBackupRequest(ruleInspectReq)
//	if err != nil {
//		return err
//	}
//	rule := resp.(*api.RuleInspectResponse).GetRule()
//	c.controller.setRuleInfo(c.ruleName, &RuleInfo{
//		RuleObject: rule,
//	})
//	return nil
//}
//
//func (p *PxBackupController) DeleteRule(ruleName string) error {
//	ruleInfo, ok := p.GetRuleInfo(ruleName)
//	if ok {
//		log.Infof("Deleting backup rule [%s] of org [%s]", ruleName, p.currentOrgId)
//		ruleDeleteReq := &api.RuleDeleteRequest{
//			Name:  ruleInfo.Name,
//			OrgId: ruleInfo.OrgId,
//			Uid:   ruleInfo.Uid,
//		}
//		if _, err := p.processPxBackupRequest(ruleDeleteReq); err != nil {
//			return err
//		}
//		p.delBackupLocationInfo(ruleName)
//	}
//	return nil
//}
