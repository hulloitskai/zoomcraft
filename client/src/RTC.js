import React, { Component, createRef, useState } from "react";
import PropTypes from "prop-types";
import styled from "@emotion/styled";

const OP_CONNECT = "connect";
const OP_JOIN = "join";
const OP_PART = "part";
const OP_SET_VOLUME = "set_volume";
const OP_SET_ICE = "set_ice";
const OP_SET_SESSION = "set_session";

const Container = styled.div`
  display: flex;
  flex-direction: column;
`;

const Title = styled.h3`
  margin: 0;
`;

const Label = styled.label`
  input {
    margin-left: 10px;
  }
`;

// const Spacer = ({ width, height, children, ...otherProps }) => (
//   <div style={{ width: width, height: height }} {...otherProps}>
//     {children}
//   </div>
// );

const Audio = styled.audio`
  display: none;
`;

const Status = styled.code`
  color: blue;
`;

const Error = styled.p`
  color: red;
`;

const SubmittableInput = ({ onSubmit, otherProps }) => {
  const [value, setValue] = useState("");
  const handleChange = ({ target }) => setValue(target.value);
  const handleKeyDown = ({ keyCode }) => keyCode === 13 && onSubmit(value);
  return (
    <input
      value={value}
      onChange={handleChange}
      onKeyDown={handleKeyDown}
      {...otherProps}
    ></input>
  );
};

class RTC extends Component {
  constructor(props) {
    super(props);
    this.audio = createRef();
    this.state = {
      error: null,
      conns: {},
      stream: null,
      connected: false,
    };
  }

  async componentDidMount() {
    this.startStream();

    const { wsUrl } = this.props;
    const ws = new WebSocket(wsUrl);

    ws.addEventListener("error", console.error);
    ws.addEventListener("open", () => console.info("WebSocket connected."));
    ws.addEventListener("close", this.handleDisconnect);
    ws.addEventListener("message", this.handleMessage);

    this.ws = ws;
  }

  componentWillUnmount() {
    this.ws.close();
  }

  startStream = async () => {
    try {
      const constraints = { audio: true, video: false };
      const stream = await navigator.mediaDevices.getUserMedia(constraints);
      this.audio.current.srcObject = stream;
      this.setState({ stream });
    } catch (error) {
      this.setState({ error });
    }
  };

  sendCommand = (command) => {
    this.ws.send(JSON.stringify(command));
  };

  handleDisconnect = (event) => {
    // Close each WebRTC connection.
    const { conns } = this.state;
    Object.values(conns).forEach((conn) => conn.close());

    // Clear conns.
    this.setState({ conns: {} });

    const { reason } = event;
    console.info(
      reason ? `WebSocket disconnected: ${reason}` : "WebSocket disconnected."
    );
  };

  handleMessage = ({ data }) => {
    const message = JSON.parse(data);
    console.log(`Received message:`, message);

    // Handle error case.
    const { error } = message;
    if (error) return this.setState({ error });

    // Handle command case.
    const { subject: id, op, payload } = message;
    switch (op) {
      case OP_CONNECT:
        this.setState({ connected: true });
        break;
      case OP_JOIN:
        const { initiate } = payload;
        return this.handleJoin({ id, initiate });
      case OP_PART:
        return this.handlePart({ id });
      case OP_SET_ICE:
        const { candidate } = payload;
        return this.handleSetICE({ id, candidate });
      case OP_SET_VOLUME:
        const { volume } = payload;
        return this.handleSetVolume({ id, volume });
      case OP_SET_SESSION:
        const { description } = payload;
        return this.handleSetSession({ id, description });
      default:
        throw new Error(`Unknown operation '${op}'.`);
    }
  };

  handleJoin = async ({ id, initiate }) => {
    try {
      /** @type {{conns: RTCPeerConnection[], stream: MediaStream}} */
      const { conns, stream } = this.state;
      if (id in conns) {
        console.warn(`Already connected to ${id}, skipping.`);
        return;
      }

      const { iceServers } = this.props;
      const conn = new RTCPeerConnection({ iceServers });
      conns[id] = conn;
      this.setState({ conns: conns });

      // Handle remote ice candidates.
      conn.addEventListener("icecandidate", ({ candidate }) => {
        console.log(`Connection ${id} produced an ICE candidate.`);
        if (!candidate) return;
        this.sendCommand({
          op: OP_SET_ICE,
          subject: id,
          payload: {
            candidate: candidate.toJSON(),
          },
        });
      });

      // Handle remote streamms.
      conn.addEventListener("addstream", ({ stream }) => {
        const audio = document.getElementById(`rtc-audio-${id}`);
        audio.srcObject = stream;
      });

      // Send local streams.
      conn.addStream(stream);

      if (initiate) {
        const desc = await conn.createOffer({ offerToReceiveAudio: true });
        await conn.setLocalDescription(desc);
        this.sendCommand({
          op: OP_SET_SESSION,
          subject: id,
          payload: {
            description: desc.toJSON(),
          },
        });
      }

      console.info(`Joined connection ${id}.`);
    } catch (error) {
      console.error(`Failed to join connection ${id}: ${error}`);
    }
  };

  handlePart = ({ id }) => {
    try {
      const { conns } = this.state;

      /** @type {RTCPeerConnection} */
      const conn = conns[id];
      if (conn === undefined) throw new Error(`Unknown connection ${id}.`);

      conn.close();
      delete conns[id];
      this.setState({ conns });
      console.info(`Parted from connection ${id}.`);
    } catch (error) {
      console.error(`Failed to part with connection ${id}: ${error}`);
    }
  };

  handleSetVolume = async ({ id, volume }) => {
    const audio = document.getElementById(`rtc-audio-${id}`);
    audio.volume = volume;
  };

  handleSetSession = async ({ id, description }) => {
    try {
      const { conns } = this.state;

      /** @type {RTCPeerConnection} */
      const conn = conns[id];
      if (conn === undefined) throw new Error(`Unknown connection ${id}.`);

      const desc = new RTCSessionDescription(description);
      await conn.setRemoteDescription(desc);

      // If the remote description was an offer, create an answer.
      if (desc.type === "offer") {
        const answer = await conn.createAnswer();
        await conn.setLocalDescription(answer);
        this.sendCommand({
          op: OP_SET_SESSION,
          subject: id,
          payload: { description: answer },
        });
      }

      console.info(`Set session for connection ${id}.`);
    } catch (error) {
      console.error(`Failed to set session for connection ${id}: ${error}`);
    }
  };

  handleSetICE = async ({ id, candidate }) => {
    try {
      const { conns } = this.state;

      /** @type {RTCPeerConnection} */
      const conn = conns[id];
      if (conn === undefined) throw new Error(`Unknown connection ${id}.`);

      conn.addIceCandidate(new RTCIceCandidate(candidate));
      console.info(`Set ICE candidate for connection ${id}.`);
    } catch (error) {
      console.error(`Failed to set ICE candidate for connection ${id}:`, error);
    }
  };

  handleConnect = async (player) => {
    this.setState({ error: null });
    this.sendCommand({ op: OP_CONNECT, payload: { player } });
    console.info(`Connected as ${player}.`);
  };

  render() {
    const { conns, error, connected } = this.state;
    return (
      <Container>
        <Title>CovidCraft</Title>
        <Label>
          Player Username:
          <SubmittableInput
            type="text"
            name="player"
            onSubmit={this.handleConnect}
          />
        </Label>
        <Audio ref={this.audio} muted autoPlay />
        {Object.keys(conns).map((id) => (
          <Audio id={`rtc-audio-${id}`} key={id} autoPlay controls />
        ))}
        {connected && <Status>CONNECTED</Status>}
        {error && <Error>{error.toString()}</Error>}
      </Container>
    );
  }
}

RTC.propTypes = {
  wsUrl: PropTypes.string.isRequired,
  iceServers: PropTypes.arrayOf(PropTypes.object).isRequired,
};

RTC.defaultProps = {
  iceServers: [{ urls: ["stun:stun.l.google.com:19302"] }],
};

export default RTC;
