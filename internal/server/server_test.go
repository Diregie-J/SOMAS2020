package server

import (
	"fmt"
	"testing"

	"github.com/SOMAS2020/SOMAS2020/internal/common"
	"github.com/SOMAS2020/SOMAS2020/pkg/testutils"
)

type mockClient struct {
	common.Client
	id   common.ClientID
	echo string
}

func (c *mockClient) GetID() common.ClientID {
	return c.id
}

func (c *mockClient) Echo(s string) string {
	return c.echo
}

// TestGetEcho also exercises getClientInfoFromRegisteredClients
func TestGetEcho(t *testing.T) {
	cases := []struct {
		name  string
		input string
		reply string
		want  error
	}{
		{
			name:  "basic ok",
			input: "42",
			reply: "42",
			want:  nil,
		},
		{
			name:  "wrong reply",
			input: "42",
			reply: "43",
			want:  fmt.Errorf("Echo error: want '42' got '43' from client 1"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mClient := &mockClient{
				id:   1,
				echo: tc.reply,
			}
			clients := map[common.ClientID]common.Client{
				common.Team1: mClient,
				common.Team2: mClient,
				common.Team3: mClient,
				common.Team4: mClient,
				common.Team5: mClient,
				common.Team6: mClient,
			}
			server := &SOMASServer{
				gameState: common.GameState{
					ClientInfos: getClientInfoFromRegisteredClients(clients),
				},
			}

			got := server.GetEcho(tc.input)
			testutils.CompareTestErrors(tc.want, got, t)
		})
	}
}
