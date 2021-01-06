package team6

import (
	"github.com/SOMAS2020/SOMAS2020/internal/common/roles"
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
)

func (c *client) GetClientPresidentPointer() roles.President {
	return &president{client: c}
}

func (c *client) GetClientJudgePointer() roles.Judge {
	return &judge{client: c}
}

func (c *client) GetClientSpeakerPointer() roles.Speaker {
	return &speaker{client: c}
}

func (c *client) ReceiveCommunication(sender shared.ClientID, data map[shared.CommunicationFieldName]shared.CommunicationContent) {
	for fieldName, content := range data {
		switch fieldName {
		case shared.TaxAmount:
			c.config.payingTax = shared.Resources(content.IntegerData)
		} //add sth else
	}
}

// ------ TODO: COMPULSORY -----
func (c *client) MonitorIIGORole(roleName shared.Role) bool {
	return c.BaseClient.MonitorIIGORole(roleName)
}

// ------ TODO: COMPULSORY -----
func (c *client) DecideIIGOMonitoringAnnouncement(monitoringResult bool) (resultToShare bool, announce bool) {
	return c.BaseClient.DecideIIGOMonitoringAnnouncement(monitoringResult)
}

func (c *client) CommonPoolResourceRequest() shared.Resources {
	minThreshold := c.ServerReadHandle.GetGameConfig().MinimumResourceThreshold
	ownResources := c.ServerReadHandle.GetGameState().ClientInfo.Resources
	if ownResources > minThreshold { //if current resource > threshold, our agent skip to request resource from common pool
		return 0
	}
	return minThreshold - ownResources
}

// ------ TODO: COMPULSORY -----
func (c *client) ResourceReport() shared.ResourcesReport {
	return c.BaseClient.ResourceReport()
}

// ------ TODO: COMPULSORY -----
func (c *client) RuleProposal() string {
	return c.BaseClient.RuleProposal()
}

func (c *client) GetTaxContribution() shared.Resources {
	ourPersonality := c.getPersonality()
	if ourPersonality == Selfish { //evade tax when we are selfish
		return 0
	}
	return c.config.payingTax
}

// ------ TODO: COMPULSORY -----
func (c *client) GetSanctionPayment() shared.Resources {
	return c.BaseClient.GetSanctionPayment()
}

func (c *client) RequestAllocation() shared.Resources {
	return c.BaseClient.RequestAllocation()
}
