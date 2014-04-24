// Package tnpldraft is a package to run TNPL drafts.
package tnpldraft

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ggriffiniii/googleauth"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

// ConnectionId represents a unique connection to the draft.
type Connection struct {
	Ws   *websocket.Conn
	User *googleauth.Profile
}

func (conn Connection) String() string {
	return fmt.Sprintf("{%v, %p}", conn.User.Email, conn.Ws)
}

// DraftSupervisor manages draftControllers. It creates a draftController object when the first connection is registered for a particular draft, and invokes the controllers Run method in a separate goroutine. It will destroy the controller when it's Run method has returned.
type DraftSupervisor struct {
	sync.Mutex
	drafts map[int64]*DraftController
}

// Create a new DraftSupervisor.
func NewSupervisor() *DraftSupervisor {
	supervisor := DraftSupervisor{
		drafts: map[int64]*DraftController{},
	}
	return &supervisor
}

// Register a new connection with supervisor. If this is the first connection for draftId it will start a new controller for the draft and invoke it's Run method in a separate goroutine.
func (supervisor *DraftSupervisor) RegisterConnection(draftId int64, conn Connection) error {
	log.Println("Registering new connection")
	supervisor.Lock()
	ctrl, ok := supervisor.drafts[draftId]
	if !ok {
		log.Printf("First connection for draft id %v", draftId)
		ctrl = NewController(draftId)
		supervisor.drafts[draftId] = ctrl
		supervisor.Unlock()
		go supervisor.runThenRemove(draftId)
		if err := ctrl.RegisterConnection(conn); err != nil {
			log.Printf("Unable to register connection %v for draft: %v: %v", conn, draftId, err)
			return err
		}
		log.Printf("Successfully registered connection: %v for draft: %v", conn, draftId)
		return nil
	} else {
		supervisor.Unlock()
		log.Printf("Adding connection: %v to already running draft id %v", conn, draftId)
		return ctrl.RegisterConnection(conn)
	}
}

func (supervisor *DraftSupervisor) runThenRemove(draftId int64) {
	supervisor.Lock()
	draft := supervisor.drafts[draftId]
	supervisor.Unlock()
	draft.Run()
	log.Printf("Draft %v exited. Removing", draftId)
	supervisor.Lock()
	delete(supervisor.drafts, draftId)
	supervisor.Unlock()
}

// Type of messages sent to the controller when a SocketMessage is received.
type TeamMessage struct {
	Connection
	SocketMessage
}

func (conn *Connection) reader(send chan<- *TeamMessage, done chan<- Connection) {
	for {
		var message SocketMessage
		if err := conn.Ws.ReadJSON(&message); err != nil {
			log.Printf("Error reading from %v", conn)
			break
		}
		send <- &TeamMessage{
			Connection:    *conn,
			SocketMessage: message,
		}
	}
	conn.Ws.Close()
	done <- *conn
}

func (conn *Connection) writer(send <-chan *SocketMessage) {
	for message := range send {
		conn.Ws.SetWriteDeadline(time.Now().Add(30 * time.Second))
		if err := conn.Ws.WriteJSON(message); err != nil {
			break
		}
	}
	conn.Ws.Close()
}

type Player struct {
	Id        int64    `json:"id"`
	Firstname string   `json:"firstname"`
	Lastname  string   `json:"lastname"`
	Mlbteam   string   `json:"mlbteam"`
	Positions []string `json:"positions"`
}

type OwnedPlayer struct {
	*Player
	Salary int `json:"salary"`
}

type TeamId int64

type Team struct {
	Id          TeamId `json:"id"`
	Name        string `json:"name"`
	playerIds   map[int64]*OwnedPlayer
	Players     []*OwnedPlayer `json:"players"`
	connections map[Connection]chan<- *SocketMessage
	owners      []string
}

func (team *Team) SendMessage(msg *SocketMessage) {
	for _, ch := range team.connections {
		select {
		case ch <- msg:
		default:
			// Outbound buffer is full. Give up.
			close(ch)
		}
	}
}

type DraftState int

const (
	WAITING_FOR_TEAMS DraftState = iota
	WAITING_FOR_PICK
	PICK_PENDING_APPROVAL
	AUCTION_IN_PROGRESS
	DRAFT_COMPLETE
)

type AuctionInfo struct {
	player       *Player
	offeringTeam *Team
	highBidder   *Team
	bid          int
	endTime      time.Time
}

type DraftController struct {
	id                int64
	Name              string           `json:"name"`
	Teams             []*Team          `json:"teams"` // in draft order
	owners            map[string]*Team // teams indexed by owner
	leaders           []string
	started           bool
	CompletedAuctions []*AuctionComplete `json:"picks"` // in draft order
	RequiredPos       map[string]int     `json:"positions"`
	SalaryCap         int                `json:"salary_cap"`
	requiredPlayers   int
	state             DraftState
	auction           *AuctionInfo

	register   chan *registerConnectionRequest
	unregister chan Connection

	// Channel this controller should listen for messages from.
	receive chan *TeamMessage
}

func rosterHelper(playersRemaining []*Player, posRemaining map[string]int) bool {
	if len(playersRemaining) == 0 {
		return true
	}
	remaining := make(map[string]int, len(posRemaining))
	for k, v := range posRemaining {
		remaining[k] = v
	}
	player := playersRemaining[0]
	for _, pos := range player.Positions {
		if remaining[pos] > 0 {
			remaining[pos]--
			if rosterHelper(playersRemaining[1:], remaining) {
				return true
			}
			remaining[pos]++
		}
	}
	return false
}

func (c *DraftController) teamHasRoomFor(team *Team, player *Player) bool {
	players := make([]*Player, len(team.Players), len(team.Players)+1)
	for i, ownedPlayer := range team.Players {
		players[i] = ownedPlayer.Player
	}
	players = append(players, player)
	return rosterHelper(players, c.RequiredPos)
}

func (c *DraftController) teamIsFull(team *Team) bool {
	return c.requiredPlayers == len(team.Players)
}

func (c *DraftController) maxTeamCanBid(team *Team) int {
	moneyLeft := c.SalaryCap
	playersNeeded := c.requiredPlayers
	for _, player := range team.Players {
		moneyLeft -= player.Salary
		playersNeeded--
	}
	return 50 + moneyLeft - playersNeeded*50
}

func loadDraftFromDb(draftId int64) *DraftController {
	conf := DraftController{
		id:   draftId,
		Name: "Test Draft",
		Teams: []*Team{
			&Team{
				Id:        1,
				Name:      "RH Team",
				playerIds: map[int64]*OwnedPlayer{},
				Players: []*OwnedPlayer{
					&OwnedPlayer{
						Player: &Player{
							Id:        91,
							Firstname: "Ben",
							Lastname:  "Zobrist",
							Mlbteam:   "Tampa Bay Rays",
							Positions: []string{"2B", "SS", "MI", "OF", "U"},
						},
						Salary: 650,
					},
				},
				connections: make(map[Connection]chan<- *SocketMessage),
				owners: []string{
					"griffin@randomhost.net",
				},
			},
			&Team{
				Id:          2,
				Name:        "Goog Team",
				playerIds:   map[int64]*OwnedPlayer{},
				Players:     []*OwnedPlayer{},
				connections: make(map[Connection]chan<- *SocketMessage),
				owners: []string{
					"glenng@google.com",
				},
			},
		},
		owners: map[string]*Team{},
		leaders: []string{
			"griffin@randomhost.net",
		},
		CompletedAuctions: []*AuctionComplete{},
		RequiredPos: map[string]int{
			"P":  10,
			"C":  2,
			"1B": 1,
			"2B": 1,
			"3B": 1,
			"SS": 1,
			"CI": 1,
			"MI": 1,
			"OF": 5,
			"U":  2,
		},
		SalaryCap: 13000,
	}
	for _, c := range conf.RequiredPos {
		conf.requiredPlayers += c
	}
	for _, team := range conf.Teams {
		for _, owner := range team.owners {
			conf.owners[owner] = team
		}
	}

	conf.auction = conf.nextAuction(conf.auction)
	if conf.auction == nil {
		conf.state = DRAFT_COMPLETE
	} else {
		conf.state = WAITING_FOR_TEAMS
	}
	return &conf
}

type registerConnectionRequest struct {
	conn Connection
	done chan error
}

func NewController(draftId int64) *DraftController {
	log.Printf("Creating new controller for draft id %v", draftId)
	controller := loadDraftFromDb(draftId)
	controller.register = make(chan *registerConnectionRequest)
	controller.unregister = make(chan Connection)
	controller.receive = make(chan *TeamMessage, 256)
	return controller
}

func (c *DraftController) RegisterConnection(conn Connection) error {
	done := make(chan error)
	c.register <- &registerConnectionRequest{
		conn: conn,
		done: done,
	}
	return <-done
}

func (c *DraftController) Run() {
	log.Printf("Running draft %v", c.id)
	for {
		if !c.EventLoop() {
			return
		}
	}
}

func (c *DraftController) EventLoop() bool {
	select {
	case request := <-c.register:
		conn := request.conn
		team, ok := c.owners[conn.User.Email]
		if !ok {
			request.done <- errors.New(conn.User.Email + " is not an owner of a team.")
			return true
		}
		sendCh := make(chan *SocketMessage, 512)
		team.connections[conn] = sendCh
		log.Printf("Registered connection %v for team %v", conn, team.Name)
		go conn.reader(c.receive, c.unregister)
		go conn.writer(sendCh)
		summary := DraftSummary{
			DraftController: c,
			Team:            team.Id,
		}
		sendCh <- SocketMessageFrom(summary)
		request.done <- nil
		switch {
		case c.state == DRAFT_COMPLETE:
			sendCh <- SocketMessageFrom(DraftComplete{})
		case c.state == WAITING_FOR_TEAMS:
			joinMsg := c.GetJoinLeaveMessage()
			if len(joinMsg.Disconnected) > 0 {
				c.broadcast(SocketMessageFrom(joinMsg))
				return true
			}
			c.startAuction(c.auction)
		case c.state == WAITING_FOR_PICK:
			sendCh <- SocketMessageFrom(c.GetWaitingForPickMessage())
		case c.state == PICK_PENDING_APPROVAL:
			if team == c.auction.offeringTeam {
				sendCh <- SocketMessageFrom(c.GetPickPendingApprovalMessage())
			} else {
				sendCh <- SocketMessageFrom(c.GetWaitingForPickMessage())
			}
		case c.state == AUCTION_IN_PROGRESS:
			sendCh <- SocketMessageFrom(c.GetAuctionMessage())
		}
	case conn := <-c.unregister:
		team := c.owners[conn.User.Email]
		delete(team.connections, conn)
		log.Printf("Unregistering connection %v for team %v", conn, team.Name)
		if c.numConnections() == 0 {
			log.Printf("Last connection closed. Exiting.")
			return false
		}
		c.broadcast(SocketMessageFrom(c.GetJoinLeaveMessage()))
	case msg := <-c.receive:
		c.handleMessage(msg)
	case <-c.auctionExpired():
		log.Println("Auction completed")
		c.finishAuction()
	}
	return true
}

func (c *DraftController) auctionExpired() <-chan time.Time {
	if c.state != AUCTION_IN_PROGRESS {
		return make(chan time.Time)
	}
	return time.After(c.auction.endTime.Sub(time.Now()))
}

func (c *DraftController) finishAuction() {
	msg := AuctionComplete{
		Player: &OwnedPlayer{
			Player: c.auction.player,
			Salary: c.auction.bid,
		},
		OfferingTeam: c.auction.offeringTeam.Id,
		WinningTeam:  c.auction.highBidder.Id,
	}
	c.auction.highBidder.Players = append(c.auction.highBidder.Players, msg.Player)
	c.recordCompletedAuction(&msg)
	c.broadcast(SocketMessageFrom(msg))
	nextAuction := c.nextAuction(c.auction)
	if nextAuction == nil {
		c.finishDraft()
	} else {
		c.startAuction(nextAuction)
	}
}

func (c *DraftController) finishDraft() {
	c.broadcast(SocketMessageFrom(DraftComplete{}))
	c.state = DRAFT_COMPLETE
}

func (c *DraftController) startAuction(auction *AuctionInfo) {
	c.auction = auction
	log.Println("WAITING_FOR_PICK")
	c.state = WAITING_FOR_PICK
	c.broadcast(SocketMessageFrom(c.GetWaitingForPickMessage()))
}

func (c *DraftController) nextAuction(current *AuctionInfo) *AuctionInfo {
	log.Printf("nextAuction(%v)", current)
	if current == nil {
		return &AuctionInfo{
			offeringTeam: c.Teams[0],
		}
	}

	// Find out where the current team fell in the draft order.
	currTeamDraftPos := -1
	for i, team := range c.Teams {
		if team == current.offeringTeam {
			currTeamDraftPos = i
			break
		}
	}

	for i := 0; i < len(c.Teams); i++ {
		nextCandidate := c.Teams[(currTeamDraftPos+1+i)%len(c.Teams)]
		if !c.teamIsFull(nextCandidate) {
			return &AuctionInfo{
				offeringTeam: nextCandidate,
			}
		}
	}
	return nil
}

func (c *DraftController) StartBidding(pick Pick) {
	c.auction.bid = pick.Bid
	c.auction.player = pick.Player
	c.auction.highBidder = c.auction.offeringTeam
	c.auction.endTime = time.Now().Add(30 * time.Second)
	log.Println("AUCTION_IN_PROGRESS")
	c.state = AUCTION_IN_PROGRESS
	msg := SocketMessageFrom(c.GetAuctionMessage())
	c.broadcast(msg)
}

func (c *DraftController) GetJoinLeaveMessage() TeamJoinLeaveMessage {
	msg := TeamJoinLeaveMessage{
		Connected:    []TeamId{},
		Disconnected: []TeamId{},
	}
	for _, team := range c.Teams {
		if len(team.connections) == 0 {
			msg.Disconnected = append(msg.Disconnected, team.Id)
		} else {
			msg.Connected = append(msg.Connected, team.Id)
		}
	}
	return msg
}

func (c *DraftController) GetWaitingForPickMessage() WaitingForPick {
	return WaitingForPick{
		Team: c.auction.offeringTeam.Id,
	}
}

func (c *DraftController) GetPickPendingApprovalMessage() PickPendingApproval {
	return PickPendingApproval{
		Player: c.auction.player,
		Bid:    c.auction.bid,
	}
}

func (c *DraftController) GetAuctionMessage() Auction {
	return Auction{
		Player:  c.auction.player,
		Team:    c.auction.highBidder.Id,
		Bid:     c.auction.bid,
		EndTime: c.auction.endTime,
	}
}

func (c *DraftController) broadcast(msg *SocketMessage) {
	for _, team := range c.Teams {
		team.SendMessage(msg)
	}
}

func (c *DraftController) handleMessage(msg *TeamMessage) {
	team := c.owners[msg.User.Email]
	//log.Printf("Message received from connection %v team %v", msg, team.Name)
	switch {
	case msg.SocketMessage.Type == "TimeRequest":
		response := SocketMessageFrom(TimeResponse{
			Time: time.Now(),
		})
		select {
		case team.connections[msg.Connection] <- response:
		default:
			// Outbound buffer is full. Give up.
			close(team.connections[msg.Connection])
		}
	case msg.SocketMessage.Type == "Pick":
		var pick Pick
		err := json.Unmarshal(msg.SocketMessage.Data, &pick)
		if err != nil {
			log.Println("Invalid message")
			return
		}
		if c.state != WAITING_FOR_PICK {
			msg := SocketMessageFrom(PlayerRejected{
				Player: pick.Player,
				Bid:    pick.Bid,
				Reason: "Pick received when not waiting for pick",
			})
			team.SendMessage(msg)
			return
		}
		if team != c.auction.offeringTeam {
			msg := SocketMessageFrom(PlayerRejected{
				Player: pick.Player,
				Bid:    pick.Bid,
				Reason: "Not expecting pick from your team",
			})
			team.SendMessage(msg)
			return
		}
		if maxBid := c.maxTeamCanBid(team); maxBid < pick.Bid {
			msg := SocketMessageFrom(PlayerRejected{
				Player: pick.Player,
				Bid:    pick.Bid,
				Reason: fmt.Sprintf("You cannot bid more than $%.2f", float32(maxBid/100)),
			})
			team.SendMessage(msg)
			return
		}
		if !c.teamHasRoomFor(team, pick.Player) {
			msg := SocketMessageFrom(PlayerRejected{
				Player: pick.Player,
				Bid:    pick.Bid,
				Reason: "No room for player on your roster",
			})
			team.SendMessage(msg)
			return
		}
		c.StartBidding(pick)
	case msg.SocketMessage.Type == "Bid":
		var bid Bid
		err := json.Unmarshal(msg.SocketMessage.Data, &bid)
		if err != nil {
			log.Println("Invalid message")
			return
		}
		if c.state != AUCTION_IN_PROGRESS {
			msg := SocketMessageFrom(BidRejected{
				Player: bid.Player,
				Bid:    bid.Bid,
				Reason: "No auction is in progress",
			})
			team.SendMessage(msg)
			log.Println("Bid received when no auction is in progress")
			return
		}
		if bid.Player.Id != c.auction.player.Id {
			msg := SocketMessageFrom(BidRejected{
				Player: bid.Player,
				Bid:    bid.Bid,
				Reason: fmt.Sprintf("Player is not up for auction"),
			})
			team.SendMessage(msg)
			return
		}
		if maxBid := c.maxTeamCanBid(team); maxBid < bid.Bid {
			msg := SocketMessageFrom(BidRejected{
				Player: bid.Player,
				Bid:    bid.Bid,
				Reason: fmt.Sprintf("You cannot bid more than $%.2f", maxBid),
			})
			team.SendMessage(msg)
			return
		}
		if bid.Bid < c.auction.bid {
			msg := SocketMessageFrom(BidRejected{
				Player: bid.Player,
				Bid:    bid.Bid,
				Reason: "Bid is not the highest bid",
			})
			team.SendMessage(msg)
			return
		}
		if !c.teamHasRoomFor(team, bid.Player) {
			msg := SocketMessageFrom(BidRejected{
				Player: bid.Player,
				Bid:    bid.Bid,
				Reason: "No room for player on your roster",
			})
			team.SendMessage(msg)
			return
		}
		c.auction.bid = bid.Bid
		c.auction.highBidder = team
		if c.auction.endTime.Sub(time.Now()) < 20*time.Second {
			c.auction.endTime = time.Now().Add(20 * time.Second)
		}
		msg := SocketMessageFrom(c.GetAuctionMessage())
		c.broadcast(msg)
	}
}

func (c *DraftController) numConnections() int {
	connCount := 0
	for _, team := range c.Teams {
		connCount += len(team.connections)
	}
	return connCount
}

func (c *DraftController) recordCompletedAuction(auction *AuctionComplete) {
	c.CompletedAuctions = append(c.CompletedAuctions, auction)
}
