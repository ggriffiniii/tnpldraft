package tnpldraft

import "testing"

func TestTeamToPickNext(t *testing.T) {
	controller := NewController(5)
	team := controller.TeamToPickNext()
	if team.id != 1 {
		t.Errorf("fail")
	}
	pick := draftPick{
		player:       &player{id: 2},
		offeringTeam: team,
		winningTeam:  team,
		winningBid:   13.00,
	}
	controller.RecordPick(&pick)
	team = controller.TeamToPickNext()
	if team.id != 2 {
		t.Errorf("fail2")
	}
}
