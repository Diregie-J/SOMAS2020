package team5

import (
	"math"
	"math/rand"

	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
	"gonum.org/v1/gonum/stat/distuv"
)

/*
================================================================
	FORAGING
================================================================
	Foraging Functions:
		DecideForage() (shared.ForageDecision, error)
		ForageUpdate(shared.ForageDecision, shared.Resources)

	IIFO Foraging Functions:
		MakeForageInfo() shared.ForageShareInfo
		ReceiveForageInfo([]shared.ForageShareInfo)
================================================================
	Work in progress
	- Find out cost of living
================================================================
*/

// DecideForage helps us pick the foraging method
func (c *client) DecideForage() (shared.ForageDecision, error) {
	if c.forageHistorySize() < c.config.InitialForageTurns { // Start with initial foraging turns (semi - randomized)
		return c.InitialForage(), nil
	} else if c.wealth() == dying { // If dying go to last hope
		return c.lastHopeForage(), nil
	}
	return c.normalForage(), nil // else forage normally
}

//================================================================
/*	Foraging types
	Each of the types of foraging are below */
//=================================================================
/* InitialForage() (Risk for wealth if we have wealth or conserve if we dont)
Born in the middle class
Roll randomly
If we suceed we are JB and will risk more to gain more
If we lose again then we are in the Imperial class and we fish to try to get back to middle class
*/
func (c *client) InitialForage() shared.ForageDecision {
	// Figure out how much the cost of livin per turn is
	// Instead of looking at history we can just store the first foraging amount
	//=============================================================================
	if c.gameState().Turn == 2 { // On turn 2
		var turncost shared.Resources
		for _, outcomes := range c.forageHistory { // For each foraging type find the outcome
			for _, returns := range outcomes { // For each outcome find the returns
				turncost += returns.input // cumlative sum of the return on investment (should only be 1 return from foraging)
				c.Logf("[DEBUG] %v - %v + %v", c.resourceHistory[1], c.resourceHistory[2], returns.input)
			}
		}
		costOfTurn1 := c.resourceHistory[1] - c.resourceHistory[2] + turncost
		c.Logf("[Debug] Cost per turn to live %v = %v - %v + %v", costOfTurn1, c.resourceHistory[1], c.resourceHistory[2], turncost)
	}
	// =============================================================================

	var forageType shared.ForageType
	// Default contribution amount is a random amount between 1% -> 5% of out wealth
	forageContribution := shared.Resources(0.01+rand.Float64()*(0.05-0.01)) * c.gameState().ClientInfo.Resources
	switch {
	case c.wealth() == jeffBezos: // JB then we have so much might as well gamble 5%->10% of it
		forageContribution = shared.Resources(0.05+rand.Float64()*(0.10-0.05)) * c.gameState().ClientInfo.Resources
		forageType = shared.DeerForageType
	case c.wealth() == imperialStudent: // Imperial student (Need to save money so dont spent a lot)
		forageType = shared.FishForageType
	case c.wealth() == dying: // dying (Its all or nothing now )
		c.lastHopeForage()
	default: // Midle class (lets see where the coin takes us)
		if rand.Float64() < 0.50 { // Coin
			forageType = shared.DeerForageType
		} else {
			forageType = shared.FishForageType
		}
	}

	c.Logf("[Debug] - [Initial Forage]:[%v][%v]", forageType, forageContribution)

	return shared.ForageDecision{
		Type:         forageType,
		Contribution: forageContribution,
	}
}

// bestForagingType indicates the best foraging method best of RoI (Return on Investment Output/Input - 1)
func (c *client) bestHistoryForaging(forageHistory forageHistory) shared.ForageType {
	bestForagingMethod := shared.ForageType(-1) // Default is that there is no good method
	bestReturn := 0.0

	for forageType, outcomes := range forageHistory { // For each foraging type
		returnOI := 0.0
		for _, returns := range outcomes {
			returnOI += float64((returns.output / returns.input) - 1) // Cumlative sum of the return on investment
		}
		returnOI = returnOI / float64(len(outcomes)) // Average RoI for the type

		if returnOI > bestReturn && returnOI > 0 { // Compares the type to the previous type and 0
			bestReturn = returnOI // If its greater than 0 then it has some return
			bestForagingMethod = forageType
		}
	}

	// Looking at our previous foraging history - Work in progress
	//=============================================================================
	deerHunters := int(0) // Number of hunters
	fishHunters := int(0)
	probDeerHunting := float64(0.1) // Base probaility to add some randomness
	probFishHunting := float64(0.1) // So we dont always go for the best type

	for forageType, FOutcome := range forageHistory { // For the whole foraging history
		for _, returns := range FOutcome {
			if forageType == shared.DeerForageType && returns.turn == c.gameState().Turn-1 { // If someone foraged for deer last turn
				deerHunters++          // Increment number of hunters
				probDeerHunting += 0.1 // Incremenet the probabilty we hunt
			} else if forageType == shared.FishForageType && returns.turn == c.gameState().Turn-1 {
				fishHunters++
				probFishHunting += 0.1
			}
		}
	}
	c.Logf("[Debug] Number of Deer Hunters from pervious turn %v", deerHunters)
	c.Logf("[Debug] Number of Fish Hunters %v", fishHunters)
	if bestForagingMethod == shared.FishForageType { // Fishing is best but 3 Deer hunters last turn
		bDeer := distuv.Bernoulli{P: 1 - probDeerHunting}     // P(1)[Fishing]=0.6 (1-0.1+0.3*3) if 3 deer hunter
		bestForagingMethod *= shared.ForageType(bDeer.Rand()) // Multipy the 0 in if Deer Hunting was picked in randomness
	} else if bestForagingMethod == shared.DeerForageType { // Deer hunting is the best choice but 3 Fishers
		bFish := distuv.Bernoulli{P: probFishHunting}         // P(1)[Fishing]= 0.1 + 0.1*3 = 0.4
		bestForagingMethod += shared.ForageType(bFish.Rand()) // +1 [makes it fishing] if Fishing was picked in randomness
	} // If best foraging was none of the 2 above then return shared.ForageType(-1)
	// ================================================================

	return bestForagingMethod
}

/* normalForage() (Past the initial, based on the history of our foraging and some randomness) */
func (c *client) normalForage() shared.ForageDecision {
	bestForagingMethod := c.bestHistoryForaging(c.forageHistory) // Find the best foragine type in based on history

	// No good returns all our history had RoI < 0
	//=============================================================================
	if bestForagingMethod == shared.ForageType(-1) && c.config.SkipForage > 0 {
		c.Logf("[Debug] - Skipping Foraging %v", c.config.SkipForage)
		c.config.SkipForage = c.config.SkipForage - 1 // Count down the number of turns to skip
		return shared.ForageDecision{                 // Dont go foraging
			Type:         shared.FishForageType,
			Contribution: 0,
		}
	} else if bestForagingMethod == shared.ForageType(-1) && c.config.SkipForage == 0 { // Force Foraging
		c.Logf("[Debug] - Force Foraging %v", c.config.SkipForage)
		c.config.SkipForage = 1 // Reassign the number of skips for next time we have RoI < 0

		// Randomly pick type and invest 1->3%
		var forageMethod shared.ForageType
		forageContribution := shared.Resources(0.01+rand.Float64()*(0.03-0.01)) * c.gameState().ClientInfo.Resources
		if rand.Float64() < 0.50 {
			forageMethod = shared.DeerForageType
		} else {
			forageMethod = shared.FishForageType
		}

		return shared.ForageDecision{
			Type:         forageMethod,
			Contribution: forageContribution,
		}
	}
	//=============================================================================

	// Foraging with previous history thats not -1
	pastOutcomes := c.forageHistory[bestForagingMethod]
	bestInput := shared.Resources(0)
	bestRoI := shared.Resources(0)

	// For all returns find the best return on investment ((output/input) -1 )
	for _, returns := range pastOutcomes { // Look at the returns of the previous
		if returns.input != 0 { // If returns are not 0
			RoI := (returns.output / returns.input) - 1 // Find the input that gave the best RoI
			if RoI > bestRoI {                          // RoI better than previous
				bestInput = returns.input // best amount to invest
				bestRoI = RoI             // best RoI so far
			}
		}
	}

	// Pick the minimum value between the best value and 10% of our resources
	bestInput = shared.Resources(math.Min(
		float64(bestInput),
		float64(0.10*c.gameState().ClientInfo.Resources)),
	)
	// Add a random amount to the bestInput (max 5%)
	bestInput += shared.Resources(math.Min(
		rand.Float64(),
		float64(0.05*c.gameState().ClientInfo.Resources)),
	)
	// Now return the foraging decision
	forageDecision := shared.ForageDecision{
		Type:         bestForagingMethod,
		Contribution: bestInput,
	}

	c.Logf(
		"[Debug] - [Forage][Normal]:Method: %v | Input: %v | Expected RoI: %v",
		bestForagingMethod, bestInput, bestRoI)

	return forageDecision
}

/*  dying MODE, RISK IT ALL, put everything in foraging for Deer */
func (c *client) lastHopeForage() shared.ForageDecision {
	forageDecision := shared.ForageDecision{
		Type:         shared.DeerForageType,
		Contribution: 0.95 * c.gameState().ClientInfo.Resources, // Almost everything we still want to be > 0 in case 0 means insta death
	}
	c.Logf("[Debug] - [Forage][LastHopeForage]: Decision %v | Amount %v",
		forageDecision, forageDecision.Contribution)
	return forageDecision
}

//================================================================
/*	Foraging History Functions */
//=================================================================

//ForageUpdate Updates the foraging history
func (c *client) ForageUpdate(forageDecision shared.ForageDecision, output shared.Resources) {
	c.forageHistory[forageDecision.Type] = append(c.forageHistory[forageDecision.Type], forageOutcome{ // Append new data
		turn:   c.gameState().Turn,
		input:  forageDecision.Contribution,
		output: output,
	})

	c.Logf(
		"[Debug] - [Update Forage History]: Type %v | Input %v | Profit %v | Real RoI %v",
		forageDecision.Type,
		forageDecision.Contribution,
		output-forageDecision.Contribution,
		(output/forageDecision.Contribution)-1,
	)
}

// forageHistorySize gets the size of our history to tell us how many rounds we have foraged for
func (c *client) forageHistorySize() uint {
	length := uint(0)
	for _, lst := range c.forageHistory {
		length += uint(len(lst))
	}
	return length // Return how many turns of foraging we have been on depending on the History
}

//======================= Part of IIFO ====================================

//RecieveForageInfo get info from other teams
func (c *client) ReceiveForageInfo(forageInfos []shared.ForageShareInfo) {
	for _, forageInfo := range forageInfos { // for all foraging information from all islands (ignore the islands)
		c.forageHistory[forageInfo.DecisionMade.Type] = // all their information (based on method of foraging)
			append( // add to our history
				c.forageHistory[forageInfo.DecisionMade.Type], // Type of foraging
				forageOutcome{ // Outcome of their foraging
					turn:   c.gameState().Turn,
					input:  forageInfo.DecisionMade.Contribution,
					output: forageInfo.ResourceObtained,
				},
			)
	}
}

//MakeForageInfo
func (c *client) MakeForageInfo() shared.ForageShareInfo {
	var shareTo []shared.ClientID

	for id, status := range c.gameState().ClientLifeStatuses { // Check the clients that are alive
		if status != shared.Dead { // if they are not dead then append the shareTo,id
			shareTo = append(shareTo, id)
		}
	}

	lastTurn := c.gameState().Turn - 1 // value of the last turn
	if lastTurn < 0 {                  // No previous foraging
		shareTo = []shared.ClientID{}
	}

	var contribution shared.ForageDecision
	var output shared.Resources
	for forageType, outcomes := range c.forageHistory { //For each type look at the outcome
		for _, outcome := range outcomes {
			if uint(outcome.turn) == lastTurn { // If the turn is the same as the last turn then return the result
				output = outcome.output               // output of the outcome
				contribution = shared.ForageDecision{ // Foraging Decision
					Type:         forageType,
					Contribution: outcome.input,
				}
			}
		}
	}

	forageInfo := shared.ForageShareInfo{ // Build the struct
		DecisionMade:     contribution, // contribution and Resources obtained
		ResourceObtained: output,       // How much we got back
		ShareTo:          shareTo,      // []shared.ClientIDs
	}

	c.Logf("Sharing forage info: %v", forageInfo)
	return forageInfo
}
