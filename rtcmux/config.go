package rtcmux

import webrtc "github.com/pion/webrtc/v2"

var defaultConfig = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{{
		URLs: []string{"stun:stun.l.google.com:19302"},
	}},
}

const defaultCodec = webrtc.RTPCodecTypeAudio
