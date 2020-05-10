package rtcsig

import (
	"sync"

	"github.com/cockroachdb/errors"
	"go.stevenxie.me/covidcraft/types"
	melody "gopkg.in/olahol/melody.v1"
)

// A Channel is a communications channel for a melody.Session.
type Channel struct {
	mux   sync.Mutex
	owner *melody.Session
	peers map[types.ID]*melody.Session
}

// NewChannel creates a channel for sess.
func NewChannel(sess *melody.Session) *Channel {
	return &Channel{
		owner: sess,
		peers: make(map[types.ID]*melody.Session),
	}
}

// AddPeer adds a peer session to ch.
func (ch *Channel) AddPeer(sess *melody.Session) {
	ch.mux.Lock()
	defer ch.mux.Unlock()

	id, ok := sess.Get("id")
	if !ok {
		panic(errors.New("rtcsig: session has no ID"))
	}
	ch.peers[id.(types.ID)] = sess
}

func (ch *Channel) RemovePeer(sess *melody.Session) {
	// ch.owner.Write()
}
