package team6

import (
	"sort"

	"github.com/SOMAS2020/SOMAS2020/internal/common/rules"
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
)

// GetVoteForRule returns the client's vote in favour of or against a rule.
func (c *client) GetVoteForRule(ruleName string) bool {
	for _, val := range c.favourRules {
		if val == ruleName {
			return true
		}
	}
	return false
}

// GetVoteForElection returns the client's Borda vote for the role to be elected.
// COMPULSORY: use opinion formation to decide a rank for islands for the role
func (c *client) GetVoteForElection(roleToElect shared.Role) []shared.ClientID {
	// Done ;)
	// Get all alive islands
	aliveClients := rules.VariableMap[rules.IslandsAlive]
	// Convert to ClientID type and place into unordered map
	aliveClientIDs := map[int]shared.ClientID{}
	for i, v := range aliveClients.Values {
		aliveClientIDs[i] = shared.ClientID(int(v))
	}
	// Recombine map, in shuffled order
	var returnList []shared.ClientID
	for _, v := range aliveClientIDs {
		returnList = append(returnList, v)
	}
	return returnList
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
