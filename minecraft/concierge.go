package minecraft

import (
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gorcon/rcon"
)

// A Concierge keeps track of the players joining and leaving a server.
type Concierge struct {
	conn   *rcon.Conn
	events chan AttendanceEvent
	ticker *time.Ticker

	mux     sync.Mutex
	players []string
}

const conciergePollInterval = time.Second

// NewConcierge creates a Concierge.
func NewConcierge(conn *rcon.Conn, logger log.Logger) *Concierge {
	c := &Concierge{
		conn:    conn,
		events:  make(chan AttendanceEvent),
		ticker:  time.NewTicker(conciergePollInterval),
		players: []string{"stevenxie", "david"},
	}
	// go func(events chan<- AttendanceEvent) {
	// 	players := make(map[string]bool)
	// 	for range c.ticker.C {
	// 		if err := func() error {
	// 			// Execute command and parse output.
	// 			out, err := conn.Execute("list")
	// 			if err != nil {
	// 				return errors.Wrap(err, "execute command")
	// 			}
	// 			out = out[strings.LastIndexByte(out, ':')+2:]
	// 			currentPlayers := strings.Split(out, ", ")

	// 			for _, player := range currentPlayers {
	// 				// If there is a player that we didn't have before, send an event.
	// 				if _, ok := players[player]; !ok {
	// 					events <- AttendanceEvent{
	// 						Player: player,
	// 						Action: AttendanceJoin,
	// 					}
	// 				}
	// 				players[player] = true
	// 			}

	// 			{
	// 				nextPlayers := make(map[string]bool)
	// 				for player, active := range players {
	// 					// If there was a previously recorded player that is not active,
	// 					// send an event.
	// 					if !active {
	// 						events <- AttendanceEvent{
	// 							Player: player,
	// 							Action: AttendanceLeave,
	// 						}
	// 					} else {
	// 						nextPlayers[player] = false
	// 					}
	// 				}
	// 				players = nextPlayers
	// 			}

	// 			// Update k.players.
	// 			c.mux.Lock()
	// 			c.players = make([]string, 0, len(players))
	// 			for player := range players {
	// 				c.players = append(c.players, player)
	// 			}
	// 			fmt.Println(c.players)
	// 			c.mux.Unlock()

	// 			return nil
	// 		}(); err != nil {
	// 			l := logutil.WithError(logger, err)
	// 			logutil.Log(l, "Error while checking attendance.")
	// 		}
	// }
	// }(c.events)
	return c
}

// Attendance returns the current list of the players currently on a server.
func (c *Concierge) Attendance() []string {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.players
}

// Events returns a channel of AttendanceEvents observed by the Concierge.
func (c *Concierge) Events() <-chan AttendanceEvent { return c.events }

// Stop stops the Keeper.
func (c *Concierge) Stop() { c.ticker.Stop() }

// An AttendanceEvent is produced when a player joins or leaves a server.
type AttendanceEvent struct {
	Player string           `json:"player"`
	Action AttendanceAction `json:"action"`
}

// An AttendanceAction is one of: AttendanceJoin, AttendanceLeave.
type AttendanceAction uint8

// The set of valid AttendanceActions.
const (
	AttendanceJoin AttendanceAction = iota
	AttendanceLeave
)
