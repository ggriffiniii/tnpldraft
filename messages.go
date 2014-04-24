package tnpldraft

import (
	"encoding/json"
	"reflect"
	"time"
)

// Messages sent over the websocket will be of this format. Type
// will describe the type of Data and Data will be one of the
// substructures below.
type SocketMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func SocketMessageFrom(msg interface{}) *SocketMessage {
	msgType := reflect.TypeOf(msg)
	json, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return &SocketMessage{
		Type: msgType.Name(),
		Data: json,
	}
}

// Sent when a new connection is registered detailing the teams involved in the draft.
type DraftSummary struct {
	*DraftController
	Team TeamId `json:"team"`
}

// A new team has joined or left the draft. Included are the list
// of teams that are currently connected and disconnected from the
// draft.
type TeamJoinLeaveMessage struct {
	Connected    []TeamId `json:"connected"`
	Disconnected []TeamId `json:"disconnected"`
}

// Sent to all teams not picking the next player.
type WaitingForPick struct {
	Team TeamId `json:"team"`
}

// The player and opening bid the team has chosen. Sent to the
// draft leaders for approval.
type Pick struct {
	Player *Player `json:"player"`
	Bid    int     `json:"bid"`
}

// Sent to team in response to PickPlayerRequest.
type PickPendingApproval struct {
	Player *Player `json:"player"`
	Bid    int     `json:"bid"`
}

// Sent to the picking team if the draft leaders don't approve the
// player along with a reason.
type PlayerRejected struct {
	Player *Player `json:"player"`
	Bid    int     `json:"bid"`
	Reason string  `json:"reason"`
}

// Sent to all teams to indicate the current state of the auction.
type Auction struct {
	Player  *Player   `json:"player"`
	Team    TeamId    `json:"team"`
	Bid     int       `json:"bid"`
	EndTime time.Time `json:"end_time"`
}

// Sent by any team to bid on a player.
type Bid struct {
	Player *Player `json:"player"`
	Bid    int     `json:"bid"`
}

// Sent when a bid is rejected.
type BidRejected struct {
	Player *Player `json:"player"`
	Bid    int     `json:"bid"`
	Reason string  `json:"reason"`
}

// Sent to all teams when an auction has completed.
type AuctionComplete struct {
	Player       *OwnedPlayer `json:"player"`
	OfferingTeam TeamId       `json:"offering_team"`
	WinningTeam  TeamId       `json:"winning_team"`
}

type DraftComplete struct {
}

// Sent to get the server's time.
type TimeRequest struct {
}

// Response with the server's time.
type TimeResponse struct {
	Time time.Time `json:"time"`
}
