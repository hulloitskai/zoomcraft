package rtcmux

import (
	"io"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/cockroachdb/errors"
	webrtc "github.com/pion/webrtc/v2"

	"go.stevenxie.me/covidcraft/types"
	"go.stevenxie.me/covidcraft/util/logutil"
)

// A Source represents a raw incoming WebRTC stream, backed by a single
// webrtc.PeerConnection.
//
// TODO: Handle in-flight errors.
type Source struct {
	id     types.ID
	err    error
	logger log.Logger

	api         *webrtc.API
	conn        *webrtc.PeerConnection
	transceiver *webrtc.RTPTransceiver

	mux   sync.Mutex
	track *webrtc.Track
}

// NewSource creates a new Source.
func NewSource(api *webrtc.API, logger log.Logger) (*Source, error) {
	id := types.NewID()
	logger = log.With(logger, "id", id)
	logger = logutil.WithComponent(logger, "source")
	logger = level.NewInjector(logger, level.DebugValue())

	conn, err := api.NewPeerConnection(defaultConfig)
	if err != nil {
		return nil, errors.Wrap(err, "create webrtc.PeerConnection")
	}
	transceiver, err := conn.AddTransceiverFromKind(defaultCodec)
	if err != nil {
		return nil, errors.Wrap(err, "add transceiver")
	}

	return &Source{
		id:     id,
		logger: logger,

		api:         api,
		conn:        conn,
		transceiver: transceiver,
	}, nil
}

func (src *Source) streamTrack(remote *webrtc.Track, _ *webrtc.RTPReceiver) {
	src.mux.Lock()

	// Create local track from remote.
	var err error
	if src.track, err = src.conn.NewTrack(
		remote.PayloadType(),
		remote.SSRC(),
		src.id.Hex(),
		"source",
	); err != nil {
		src.handleError(err, "Failed to create local track.")
		src.mux.Unlock()
		return
	}
	src.mux.Unlock()

	// Continuously transfer data chunks from remote to local.
	chunk := make([]byte, 1400)
	for {
		i, err := remote.Read(chunk)
		if err != nil {
			src.handleError(err, "Failed to read from remote track.")
			break
		}

		// ErrClosedPipe means we don't have any subscribers, this is ok if no
		// peers have connected yet.
		_, err = src.track.Write(chunk[:i])
		if (err != nil) && (err != io.ErrClosedPipe) {
			src.handleError(err, "Failed to write to local track.")
			break
		}
	}
}

func (src *Source) handleError(err error, msg string) {
	logutil.Log(logutil.WithError(src.logger, err), msg)
	src.err = err
}
