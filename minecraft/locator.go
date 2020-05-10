package minecraft

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/go-kit/kit/log"
	"github.com/gorcon/rcon"

	"go.stevenxie.me/covidcraft/util/logutil"
)

// A Locator keeps track of the players joining and leaving a server.
type Locator struct {
	conn   *rcon.Conn
	events chan LocationEvent
	ticker *time.Ticker

	mux    sync.Mutex
	coords map[string]Coords
}

const (
	locatorPollInterval = 500 * time.Millisecond
)

// NewLocator creates a Locator.
func NewLocator(c *Concierge, conn *rcon.Conn, logger log.Logger) *Locator {
	l := &Locator{
		conn:   conn,
		events: make(chan LocationEvent),
		ticker: time.NewTicker(locatorPollInterval),
		coords: make(map[string]Coords),
	}
	go func(events chan<- LocationEvent) {
		for range l.ticker.C {
			if err := func() error {
				var (
					players = c.Attendance()
					coords  = make(map[string]Coords, len(players))
				)
				for _, player := range players {
					// Execute command.
					cmd := fmt.Sprintf("data get entity %s Pos", player)
					out, err := conn.Execute(cmd)
					if err != nil {
						return errors.Wrap(err, "execute command")
					}
					if out == "No entity was found" { // player disconnected
						continue
					}

					// Parse output.
					out = out[strings.LastIndexByte(out, ':')+2:]
					out = strings.Trim(out, "[]")

					// Parse coordinate parts.
					var (
						parts = strings.Split(out, ", ")
						c     Coords
					)
					for i, part := range parts {
						f, err := strconv.ParseFloat(part[:len(part)-1], 64)
						if err != nil {
							return errors.Wrap(err, "parse coordinate part")
						}

						switch i {
						case 0:
							c.X = f
						case 1:
							c.Y = f
						case 2:
							c.Z = f
						}
					}
					coords[player] = c

					events <- LocationEvent{
						Player: player,
						Coords: c,
					}
				}

				l.mux.Lock()
				l.coords = coords
				l.mux.Unlock()
				return nil
			}(); err != nil {
				l := logutil.WithError(logger, err)
				logutil.Log(l, "Error while checking Location.")
			}
		}
	}(l.events)
	return l
}

// Coords returns the coordinates of a particular player.
func (l *Locator) Coords(player string) Coords {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.coords[player]
}

// Distance returns the distance between players a and b.
func (l *Locator) Distance(player1, player2 string) float64 {
	var c1, c2 Coords

	// Load coordinates.
	l.mux.Lock()
	c1 = l.coords[player1]
	c2 = l.coords[player2]
	defer l.mux.Unlock()

	// Quick mafs.
	var (
		dxsq = math.Pow(c2.X-c1.X, 2)
		dysq = math.Pow(c2.Y-c1.Y, 2)
		dzsq = math.Pow(c2.Z-c1.Z, 2)
	)
	return math.Sqrt(dxsq + dysq + dzsq)
}

// Events returns a channel of LocationEvents observed by the Locator.
func (l *Locator) Events() <-chan LocationEvent { return l.events }

// Stop stops the Keeper.
func (l *Locator) Stop() { l.ticker.Stop() }

// An LocationEvent is a record of a player's location.
type LocationEvent struct {
	Player string `json:"player"`
	Coords Coords `json:"coords"`
}

// Coords are coordinates for a position in a Minecraft world.
type Coords struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}
