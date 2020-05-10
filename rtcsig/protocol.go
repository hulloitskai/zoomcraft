package rtcsig

import (
	"go.stevenxie.me/covidcraft/types"
)

// A Command is a signaling command.
type Command struct {
	Op      Op          `json:"op"`
	Subject types.ID    `json:"subject"`
	Payload interface{} `json:"payload,omitempty"`
}

// An Op represents a signaling operation.
type Op string

// The set of valid signaling operations.
const (
	OpConnect    = "connect"
	OpJoin       = "join"
	OpPart       = "part"
	OpSetICE     = "set_ice"
	OpSetVolume  = "set_volume"
	OpSetSession = "set_session"
)

type (
	// PayloadConnect is the payload sent along with an init operataion.
	PayloadConnect struct {
		Player string `json:"player"`
	}

	// PayloadJoin is the payload sent along with a join operataion.
	PayloadJoin struct {
		Initiate bool `json:"initiate"`
	}

	// PayloadVolume is the payload sent along with a volume change operation.
	PayloadVolume struct {
		Volume float64 `json:"volume"`
	}

	// PayloadRelayICE is the payload sent along with an ICE candidate relay
	// operation.
	PayloadRelayICE struct {
		Candidate []byte `json:"candidate"`
	}

	// PayloadRelaySession is the payload sent along with an session description
	// relay operation.
	PayloadRelaySession struct {
		Description []byte `json:"description"`
	}
)

// PayloadError contains an error message.
type PayloadError struct {
	Error string `json:"error"`
}
