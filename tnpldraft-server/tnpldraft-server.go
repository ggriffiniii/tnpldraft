package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ggriffiniii/go/gmailid"
	"github.com/ggriffiniii/go/tnpldraft"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
)

import _ "github.com/go-sql-driver/mysql"

var clientId = flag.String("clientid", "973698106370-dmmi34cti20d3j66v6oe09l9tc0oqp3q.apps.googleusercontent.com", "The ouath client id")
var clientSecret = flag.String("clientsecret", "0amrPk5urZCSA5IVS0SYG_7q", "The oauth client secret")
var cookieName = flag.String("cookiename", "user", "The cookie name to use")
var port = flag.Int("port", 8082, "The port to listen on")
var static_dir = flag.String("static", "", "The static dir")
var db = flag.String("db", "newtnpldraft", "The database name to use")
var dbuser = flag.String("dbuser", "tnpldraft", "The database user to use")
var dbpass = flag.String("dbpass", "tnpldraft", "The database password to use")

func main() {
	flag.Parse()
	gmail := gmailid.Manager{
		ClientId:     *clientId,
		ClientSecret: *clientSecret,
		CookieName:   *cookieName,
	}
	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@/%v", *dbuser, *dbpass, *db))
	if err != nil {
		log.Fatal(err)
	}
	draftSupervisor := tnpldraft.NewSupervisor()
	r := mux.NewRouter()
	gmail.RegisterCallback(r, "/oauthcallback")
	r.Handle("/testauth", gmail.ProtectedHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		profile, _ := gmail.GetProfile(r)
		w.Write([]byte(profile.Email))
	})))
	r.Handle("/ws/{draftId}", gmail.ProtectedHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		profile, err := gmail.GetProfile(r)
		if err != nil {
			http.Error(w, "Not authenticated", http.StatusUnauthorized)
		}
		ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			log.Println(err)
			return
		}
		conn := tnpldraft.Connection{
			Ws:   ws,
			User: profile,
		}
		draftId, err := strconv.ParseInt(mux.Vars(r)["draftId"], 10, 64)
		if err != nil {
			http.Error(w, "draftid needs to be a number", 400)
			return
		}
		err = draftSupervisor.RegisterConnection(draftId, conn)
		if err != nil {
			http.Error(w, "Couldn't load draft", 500)
			return
		}
	})))
	r.Handle("/api/draft/{draftId}/playerfilter", gmail.ProtectedHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//		draftId, err := strconv.ParseInt(mux.Vars(r)["draftId"], 10, 64)
		//		if err != nil {
		//			http.Error(w, "draftid needs to be a number", 400)
		//			return
		//		}
		params := r.URL.Query()
		name := params.Get("name")
		if err := getPlayers(w, name, db); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	})))
	r.Handle("/{unused:.*}", gmail.ProtectedHandler(http.FileServer(http.Dir(*static_dir))))
	log.Println("Listening on ", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), r))
}

func getPlayers(w http.ResponseWriter, name string, db *sql.DB) error {
	rows, err := db.Query("select player.id, player.firstname, player.lastname, player.pitcher, player.catcher, player.firstbase, player.secondbase, player.thirdbase, player.shortstop, player.outfield, player.utility, mlbteam.name FROM player JOIN mlbteam ON player.mlbteam_id = mlbteam.id WHERE CONCAT(player.firstname, ' ', player.lastname) LIKE CONCAT('%', ?, '%') LIMIT 50;", name)
	if err != nil {
		return err
	}
	defer rows.Close()
	players := make([]*tnpldraft.Player, 0)
	for rows.Next() {
		var (
			id                                      int64
			firstname, lastname, team               string
			pitcher, catcher, firstbase, secondbase bool
			thirdbase, shortstop, outfield, utility bool
		)
		err := rows.Scan(&id, &firstname, &lastname, &pitcher, &catcher, &firstbase, &secondbase, &thirdbase, &shortstop, &outfield, &utility, &team)
		if err != nil {
			return err
		}
		player := tnpldraft.Player{
			Id:        id,
			Firstname: firstname,
			Lastname:  lastname,
			Mlbteam:   team,
			Positions: []string{},
		}
		if pitcher {
			player.Positions = append(player.Positions, "P")
		}
		if catcher {
			player.Positions = append(player.Positions, "C")
		}
		if secondbase {
			player.Positions = append(player.Positions, "2B")
		}
		if shortstop {
			player.Positions = append(player.Positions, "SS")
		}
		if secondbase || shortstop {
			player.Positions = append(player.Positions, "MI")
		}
		if thirdbase {
			player.Positions = append(player.Positions, "3B")
		}
		if firstbase {
			player.Positions = append(player.Positions, "1B")
		}
		if firstbase || thirdbase {
			player.Positions = append(player.Positions, "CI")
		}
		if outfield {
			player.Positions = append(player.Positions, "OF")
		}
		if !pitcher {
			player.Positions = append(player.Positions, "U")
		}

		players = append(players, &player)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	encoder := json.NewEncoder(w)
	err = encoder.Encode(players)
	return err
}
