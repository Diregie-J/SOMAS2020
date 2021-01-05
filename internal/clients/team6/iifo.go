package team6

import (
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
)

// ------ TODO: COMPULSORY ------
func (c *client) MakeDisasterPrediction() shared.DisasterPredictionInfo {
	return c.BaseClient.MakeDisasterPrediction()
}

// ------ TODO: COMPULSORY ------
func (c *client) ReceiveDisasterPredictions(receivedPredictions shared.ReceivedDisasterPredictionsDict) {
	c.BaseClient.ReceiveDisasterPredictions(receivedPredictions)
}

// ------ TODO: OPTIONAL ------
func (c *client) MakeForageInfo() shared.ForageShareInfo {
	return c.BaseClient.MakeForageInfo()
}

func (c *client) ReceiveForageInfo(forageInfo []shared.ForageShareInfo) {
	for _, val := range forageInfo {
		c.forageHistory[val.DecisionMade.Type] =
			append(
				c.forageHistory[val.DecisionMade.Type],
				ForageResults{
					forageIn:     val.DecisionMade.Contribution,
					forageReturn: val.ResourceObtained,
				},
			)
	}
}
