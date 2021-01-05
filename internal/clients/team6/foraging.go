package team6

import (
	"math/rand"
	"sort"

	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
)

// ForageHistory stores our forage history
// type ForageHistory map[shared.ForageType][]ForageResults

// type ForageResults struct {
// 	forageIn     shared.Resources
// 	forageReturn shared.Resources
// }
func (result ForageResults) calcROI() float64 {
	if result.forageIn == 0 {
		return 0
	} else {
		return float64(result.forageReturn/result.forageIn) - 1
	}
}

func (c client) changeMultiplier() float64 {
	for _, results := range forageHistory {
		var lastRoi float64 = 0
		var lastlastRoi float64 = 0
		for _, result := range results {
			if result.turn == c.ServerReadHandle.GetGameState().Turn-2 {
				lastlastRoi = result.calcROI()
			}
			if result.turn == c.ServerReadHandle.GetGameState().Turn-1 {
				lastRoi = result.calcROI()
			}
			if lastRoi != 0 && lastlastRoi != 0 {
				//it means last round and the round before last round are using the same forage type
				if lastlastRoi < lastRoi {
					c.config.multiplier += 0.1
				} else {
					c.config.multiplier += 0.1
				}
			}
		}
	}
	return c.config.multiplier
}
func (c *client) changeForageType() shared.ForageType {
	//fishing is a safer choice if we contributed a lot
	if c.config.multiplier > 0.5 {
		return shared.FishForageType
	}
	return shared.DeerForageType
}

func (c *client) decideContribution() shared.Resources {

	var safetyBuffer shared.Resources = 10
	ourResources := c.ServerReadHandle.GetGameState().ClientInfo.Resources
	return shared.Resources(c.changeMultiplier()) * (ourResources - safetyBuffer)
}

func (c *client) randomForage() shared.ForageDecision {
	var resources shared.Resources
	var forageType shared.ForageType

	if c.ServerReadHandle.GetGameState().Turn == 2 {
		forageType = shared.FishForageType
	} else {
		forageType = shared.DeerForageType
	}
	tmp := rand.Float64()
	if tmp > 0.2 { //up to 20% resources
		resources = 0.2 * c.ServerReadHandle.GetGameState().ClientInfo.Resources
	} else {
		resources = shared.Resources(tmp) * c.ServerReadHandle.GetGameState().ClientInfo.Resources
	}

	return shared.ForageDecision{
		Type:         shared.ForageType(forageType),
		Contribution: shared.Resources(resources),
	}
}

func (c *client) noramlForage() shared.ForageDecision {
	ft := c.changeForageType()
	amt := c.decideContribution()
	return shared.ForageDecision{
		Type:         shared.ForageType(ft),
		Contribution: shared.Resources(amt),
	}
}

func (c *client) DecideForage() (shared.ForageDecision, error) {
	if c.ServerReadHandle.GetGameState().Turn < 3 { //the agent will randomly forage at the start
		return c.randomForage(), nil
	} else {
		return c.noramlForage(), nil
	}

}

func (c *client) ForageUpdate(forageDecision shared.ForageDecision, outcome shared.Resources) {
	currTurn := c.ServerReadHandle.GetGameState().Turn

	c.forageHistory[forageDecision.Type] =
		append(
			c.forageHistory[forageDecision.Type],
			ForageResults{
				forageIn:     forageDecision.Contribution,
				forageReturn: outcome,
				turn:         currTurn,
			},
		)

	c.Logf(
		"Forage History Updated: Type %v ,Conribution: %v ,Return: %v",
		forageDecision.Type,
		forageDecision.Contribution,
		outcome,
	)
}

// Pair is used for sorting the friendship
type Pair struct {
	Key   shared.ClientID
	Value uint
}

// PairList is a slice of Pairs that implements sort.Interface to sort by Value.
type PairList []Pair

func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }

// A function to turn a map into a PairList, then sort and return it.
func sortMapByValue(m map[shared.ClientID]uint) PairList {
	p := make(PairList, len(m))
	i := 0
	for k, v := range m {
		p[i] = Pair{k, v}
	}
	sort.Sort(p)
	return p
}
