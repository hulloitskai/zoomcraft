package rtcsig

import (
	"encoding/json"
	"net/http"
	"sync"

	melody "gopkg.in/olahol/melody.v1"

	"github.com/cockroachdb/errors"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorcon/rcon"
	"github.com/thoas/go-funk"

	"go.stevenxie.me/covidcraft/minecraft"
	"go.stevenxie.me/covidcraft/types"
	"go.stevenxie.me/covidcraft/util/logutil"
)

// An Interchange keeps track of all active sessions, and handles relay
// commands.
type Interchange struct {
	melody *melody.Melody
	logger log.Logger

	concierge *minecraft.Concierge
	locator   *minecraft.Locator

	mux              sync.Mutex
	sessions         map[types.ID]*melody.Session
	sessionsByPlayer map[string]*melody.Session
}

const maxMessageSize = 4096

// NewInterchange creates a new Interchange.
func NewInterchange(m *melody.Melody, game *rcon.Conn, logger log.Logger) *Interchange {
	m.Config.MaxMessageSize = maxMessageSize

	game2, _ := rcon.Dial("localhost:25575", "minecraft")

	// Create subcomponents.
	var (
		concierge = minecraft.NewConcierge(
			game,
			logutil.WithComponent(logger, "concierge"),
		)
		locator = minecraft.NewLocator(
			concierge,
			game2,
			logutil.WithComponent(logger, "locator"),
		)
	)

	// Create interchange.
	in := &Interchange{
		melody:           m,
		logger:           level.NewInjector(logger, level.DebugValue()),
		concierge:        concierge,
		locator:          locator,
		sessions:         make(map[types.ID]*melody.Session),
		sessionsByPlayer: make(map[string]*melody.Session),
	}

	// Register handlers.
	m.HandleError(in.handleError)
	m.HandleMessage(in.handleMessage)
	m.HandleConnect(func(sess *melody.Session) {
		id := types.NewID()
		sess.Set(idKey, id)
	})
	m.HandleDisconnect(in.handleDisconnect)

	// Watch game disconnects.
	go func(events <-chan minecraft.AttendanceEvent) {
		for event := range events {
			if event.Action == minecraft.AttendanceLeave {
				in.handleTerminate(event.Player)
			}
		}
	}(concierge.Events())

	// Watch location updates.
	go func(events <-chan minecraft.LocationEvent) {
		for event := range events {
			for player, sess := range in.sessionsByPlayer {
				if player == event.Player {
					continue
				}
				distance := locator.Distance(player, event.Player)
				in.handleSetVolume(sess, event.Player, distance)
			}
		}
	}(locator.Events())

	return in
}

// ServeHTTP implements http.Handler.
func (ic *Interchange) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := ic.melody.HandleRequest(w, r)
	if err != nil {
		l := logutil.WithError(ic.logger, err)
		logutil.Log(l, "Failed to handle request.")
	}
}

func (ic *Interchange) handleMessage(sess *melody.Session, msg []byte) {
	l := log.With(ic.logger, "message", msg)
	id := sessionID(sess)

	if err := func() error {
		var cmd struct {
			Op      Op              `json:"op"`
			Subject types.ID        `json:"subject"`
			Payload json.RawMessage `json:"payload,omitempty"`
		}
		if err := json.Unmarshal(msg, &cmd); err != nil {
			return errors.Wrap(err, "decode message")
		}

		switch cmd.Op {
		case OpConnect:
			var payload PayloadConnect
			if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
				return errors.Wrap(err, "decode payload")
			}
			ic.handleConnect(sess, payload.Player)

		case OpSetICE, OpSetSession:
			target, ok := ic.sessions[cmd.Subject]
			if !ok {
				return errors.New("rtcsig: bad subject")
			}
			cmd := Command{
				Op:      cmd.Op,
				Subject: id,
				Payload: cmd.Payload,
			}
			msg, err := json.Marshal(cmd)
			if err != nil {
				return errors.Wrap(err, "encode command")
			}
			if err = target.Write(msg); err != nil {
				return errors.Wrap(err, "write command")
			}

		default:
			return errors.New("rtcsig: invalid operation")
		}

		logutil.Log(l, "Handled command.")
		return nil
	}(); err != nil {
		l := logutil.WithError(l, err)
		logutil.Log(l, "Failed to handle command.")
	}
}

const idKey = "id"

func (ic *Interchange) handleConnect(sess *melody.Session, player string) {
	id := sessionID(sess)
	l := log.With(ic.logger, "id", id)

	if err := func() error {
		// Ensure player is active.
		players := ic.concierge.Attendance()
		if !funk.ContainsString(players, player) {
			return sendPayload(sess, PayloadError{
				Error: "No such in-game player.",
			})
		}

		// Ensure player not claimed.
		if _, ok := ic.sessionsByPlayer[player]; ok {
			return sendPayload(sess, PayloadError{
				Error: "Player session already claimed.",
			})
		}

		// Broadcast join command to all other sessions.
		ic.mux.Lock()
		defer ic.mux.Unlock()
		for otherID, otherSess := range ic.sessions {
			// Tell sess to join otherSess.
			if err := sendPayload(sess, Command{
				Op:      OpJoin,
				Subject: otherID,
				Payload: PayloadJoin{Initiate: true},
			}); err != nil {
				return err
			}

			// Tell otherSess to join sess.
			if err := sendPayload(otherSess, Command{
				Op:      OpJoin,
				Subject: id,
				Payload: PayloadJoin{Initiate: false},
			}); err != nil {
				return err
			}
		}
		ic.sessions[id] = sess
		ic.sessionsByPlayer[player] = sess

		// Confirm connection.
		if err := sendPayload(sess, Command{
			Op:      OpConnect,
			Subject: id,
			Payload: PayloadConnect{Player: player},
		}); err != nil {
			return err
		}

		logutil.Log(level.Info(l), "A session has connected.")
		return nil
	}(); err != nil {
		l := logutil.WithError(l, err)
		logutil.Log(l, "Failed to connect session.")
	}
}

func (ic *Interchange) handleTerminate(player string) {
	l := log.With(ic.logger, "player", player)
	if err := func() error {
		sess := ic.sessionsByPlayer[player]
		if err := sess.Close(); err != nil {
			return errors.Wrap(err, "close session")
		}
		logutil.Log(l, "Terminated session due to player disconnect.")
		return nil
	}(); err != nil {
		l := logutil.WithError(l, err)
		logutil.Log(l, "Failed to terminate session.")
	}
}

func (ic *Interchange) handleSetVolume(sess *melody.Session, player string, distance float64) {
	id := sessionID(sess)
	l := log.With(ic.logger, "id", id, "player", player, "distance", distance)
	if err := func() error {
		playerSess, ok := ic.sessionsByPlayer[player]
		if !ok {
			return nil
		}
		playerID := sessionID(playerSess)

		volume := (30 - distance) / 100
		if volume < 0 {
			volume = 0
		}
		if err := sendPayload(sess, Command{
			Op:      OpSetVolume,
			Subject: playerID,
			Payload: PayloadVolume{Volume: volume},
		}); err != nil {
			return err
		}

		logutil.Log(l, "Updated volume for %s on %s.", playerID, id)
		return nil
	}(); err != nil {
		l := logutil.WithError(l, err)
		logutil.Log(l, "Failed to update volume.")
	}
}

func (ic *Interchange) handleDisconnect(sess *melody.Session) {
	id := sessionID(sess)
	l := log.With(ic.logger, "id", id)
	if err := func() error {
		// Remove session.
		ic.mux.Lock()
		delete(ic.sessions, id)
		for player, session := range ic.sessionsByPlayer {
			if session == sess {
				delete(ic.sessionsByPlayer, player)
				break
			}
		}
		defer ic.mux.Unlock()

		// Broadcast disconnect command to all other sessions.
		cmd := Command{
			Op:      OpPart,
			Subject: id,
		}
		msg, err := json.Marshal(&cmd)
		if err != nil {
			return errors.Wrap(err, "encode command")
		}
		for id, sess := range ic.sessions {
			if err := sess.Write(msg); err != nil {
				return errors.Wrapf(err, "write to %s", id.Hex())
			}
		}

		logutil.Log(level.Info(l), "A session has disconnected.")
		return nil
	}(); err != nil {
		l := logutil.WithError(ic.logger, err)
		logutil.Log(l, "Error while disconnecting a session.")
	}
}

func (ic *Interchange) handleError(sess *melody.Session, err error) {
	id := sessionID(sess)
	l := log.With(ic.logger, "id", id)
	l = logutil.WithError(l, err)
	logutil.Log(l, "An unexpected websocket error occurred.")
}

func sessionID(sess *melody.Session) types.ID {
	id, exists := sess.Get(idKey)
	if !exists {
		return types.ZeroID
	}
	return id.(types.ID)
}

func sendPayload(sess *melody.Session, payload interface{}) error {
	msg, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrap(err, "encode command")
	}
	if err = sess.Write(msg); err != nil {
		return errors.Wrap(err, "write")
	}
	return nil
}
