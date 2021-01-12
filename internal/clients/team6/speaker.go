package team6

import (
	"github.com/SOMAS2020/SOMAS2020/internal/common/baseclient"
	"github.com/SOMAS2020/SOMAS2020/internal/common/shared"
)

type speaker struct {
	*baseclient.BaseSpeaker
	*client
}

func (s *speaker) DecideNextJudge(winner shared.ClientID) shared.ClientID {
	if s.client.friendship[winner] <= s.client.clientConfig.maxFriendship/2 {
		return s.client.GetID()
	}

	return winner
}
