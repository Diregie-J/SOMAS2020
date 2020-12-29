package iigointernal

import (
	"fmt"

	"github.com/SOMAS2020/SOMAS2020/internal/common/baseclient"
	"github.com/SOMAS2020/SOMAS2020/internal/common/config"
	"github.com/SOMAS2020/SOMAS2020/internal/common/gamestate"
	"github.com/SOMAS2020/SOMAS2020/internal/common/roles"
	"github.com/SOMAS2020/SOMAS2020/internal/common/rules"
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
	"github.com/SOMAS2020/SOMAS2020/internal/common/voting"
	"github.com/pkg/errors"
)

type legislature struct {
	gameState     *gamestate.GameState
	SpeakerID     shared.ClientID
	judgeSalary   shared.Resources
	ruleToVote    string
	ballotBox     voting.BallotBox
	votingResult  bool
	clientSpeaker roles.Speaker
}

// loadClientSpeaker checks client pointer is good and if not panics
func (l *legislature) loadClientSpeaker(clientSpeakerPointer roles.Speaker) {
	if clientSpeakerPointer == nil {
		panic(fmt.Sprintf("Client '%v' has loaded a nil speaker pointer", l.SpeakerID))
	}
	l.clientSpeaker = clientSpeakerPointer
}

// sendJudgeSalary conduct the transaction based on amount from client implementation
func (l *legislature) sendJudgeSalary() error {
	if l.clientSpeaker != nil {
		amount, judgePaid := l.clientSpeaker.PayJudge(l.judgeSalary)
		if judgePaid {
			// Subtract from common resources pool
			amountWithdraw, withdrawSuccess := WithdrawFromCommonPool(amount, l.gameState)

			if withdrawSuccess {
				// Pay into the client private resources pool
				depositIntoClientPrivatePool(amountWithdraw, JudgeIDGlobal, l.gameState)
			}
		}
	}
	return errors.Errorf("Cannot perform sendJudgeSalary")
}

// Receive a rule to call a vote on
func (l *legislature) setRuleToVote(r string) error {

	if !l.incurServiceCharge("setRuleToVote") {
		return errors.Errorf("Insufficient Budget in common Pool: setRuleToVote")
	}

	ruleToBeVoted, ruleSet := l.clientSpeaker.DecideAgenda(r)
	if ruleSet {
		l.ruleToVote = ruleToBeVoted
	}
	return nil
}

//Asks islands to vote on a rule
//Called by orchestration
func (l *legislature) setVotingResult(clientIDs []shared.ClientID) (bool, error) {

	if !l.incurServiceCharge("setVotingResult") {
		return false, errors.Errorf("Insufficient Budget in common Pool: setVotingResult")
	}

	ruleID, participatingIslands, voteDecided := l.clientSpeaker.DecideVote(l.ruleToVote, clientIDs)
	if !voteDecided {
		return false, nil
	}

	l.ballotBox = l.RunVote(ruleID, participatingIslands)

	l.votingResult = l.ballotBox.CountVotesMajority()

	return true, nil
}

//RunVote creates the voting object, returns votes by category (for, against) in BallotBox.
//Passing in empty ruleID or empty clientIDs results in no vote occurring
func (l *legislature) RunVote(ruleID string, clientIDs []shared.ClientID) voting.BallotBox {

	if ruleID == "" || len(clientIDs) == 0 {
		return voting.BallotBox{}
	}

	ruleVote := voting.RuleVote{}

	//TODO: check if rule is valid, otherwise return empty ballot, raise error?
	ruleVote.SetRule(ruleID)

	//TODO: intersection of islands alive and islands chosen to vote in case of client error
	//TODO: check if remaining slice is >0, otherwise return empty ballot, raise error?
	ruleVote.SetVotingIslands(clientIDs)

	ruleVote.GatherBallots(iigoClients)
	//TODO: log of vote occurring with ruleID, clientIDs
	//TODO: log of clientIDs vs islandsAllowedToVote
	//TODO: log of ruleID vs s.RuleToVote
	return ruleVote.GetBallotBox()
}

//Speaker declares a result of a vote (see spec to see conditions on what this means for a rule-abiding speaker)
//Called by orchestration
func (l *legislature) announceVotingResult() error {

	rule, result, announcementDecided := l.clientSpeaker.DecideAnnouncement(l.ruleToVote, l.votingResult)

	if announcementDecided {
		//Deduct action cost
		if !l.incurServiceCharge("announceVotingResult") {
			return errors.Errorf("Insufficient Budget in common Pool: announceVotingResult")
		}

		//Reset
		l.ruleToVote = ""
		l.votingResult = false

		//Perform announcement
		broadcastToAllIslands(shared.TeamIDs[l.SpeakerID], generateVotingResultMessage(rule, result))
	}
	return nil
}

func generateVotingResultMessage(ruleID string, result bool) map[shared.CommunicationFieldName]shared.CommunicationContent {
	returnMap := map[shared.CommunicationFieldName]shared.CommunicationContent{}

	returnMap[shared.RuleName] = shared.CommunicationContent{
		T:        shared.CommunicationString,
		TextData: ruleID,
	}
	returnMap[shared.RuleVoteResult] = shared.CommunicationContent{
		T:           shared.CommunicationBool,
		BooleanData: result,
	}

	return returnMap
}

//reset resets internal variables for safety
func (l *legislature) reset() {
	l.ruleToVote = ""
	l.ballotBox = voting.BallotBox{}
	l.votingResult = false
}

// updateRules updates the rules in play according to the result of a vote.
func (l *legislature) updateRules(ruleName string, ruleVotedIn bool) error {
	if !l.incurServiceCharge("updateRules") {
		return errors.Errorf("Insufficient Budget in common Pool: updateRules")
	}
	//TODO: might want to log the errors as normal messages rather than completely ignoring them? But then Speaker needs access to client's logger
	//notInRulesCache := errors.Errorf("Rule '%v' is not available in rules cache", ruleName)
	if ruleVotedIn {
		// _ = rules.PullRuleIntoPlay(ruleName)
		err := rules.PullRuleIntoPlay(ruleName)
		if ruleErr, ok := err.(*rules.RuleError); ok {
			if ruleErr.Type() == rules.RuleNotInAvailableRulesCache {
				return ruleErr
			}
		}
	} else {
		// _ = rules.PullRuleOutOfPlay(ruleName)
		err := rules.PullRuleOutOfPlay(ruleName)
		if ruleErr, ok := err.(*rules.RuleError); ok {
			if ruleErr.Type() == rules.RuleNotInAvailableRulesCache {
				return ruleErr
			}
		}

	}
	return nil

}

func (l *legislature) appointNextJudge(clientIDs []shared.ClientID) (shared.ClientID, error) {
	if !l.incurServiceCharge("appointNextJudge") {
		return l.SpeakerID, errors.Errorf("Insufficient Budget in common Pool: appointNextJudge")
	}
	var election voting.Election
	election.ProposeElection(baseclient.Judge, voting.Plurality)
	election.OpenBallot(clientIDs)
	election.Vote(iigoClients)
	return election.CloseBallot(), nil
}

func (l *legislature) incurServiceCharge(actionID string) bool {
	cost := config.GameConfig().IIGOConfig.LegislativeActionCost[actionID]
	_, ok := WithdrawFromCommonPool(cost, l.gameState)
	if ok {
		l.gameState.IIGORolesBudget["speaker"] -= cost
	}
	return ok
}
