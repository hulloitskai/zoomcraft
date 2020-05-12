import React, { Component, useState, useEffect } from "react";
import styled from "@emotion/styled";
import { gql, useQuery } from "@apollo/client";

import map from "lodash/map";
import get from "lodash/get";
import keyBy from "lodash/keyBy";
import isEmpty from "lodash/isEmpty";
import forEach from "lodash/forEach";

import AudioCard, { SourceType } from "./audiocard";
import { AddCard } from "./card";

import droplet from "./assets/droplet.wav";
import { rotate, deg2rad } from "./math";

const Container = styled.div`
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 1.5rem;

  h1 {
    margin: 0;
    font-weight: 800;
  }
`;

const Cards = styled.div`
  margin-top: 0.8rem;
  display: flex;
  flex-wrap: wrap;
`;

const QUERY = gql`
  query($username: String!) {
    players {
      username
      position
    }
    player(username: $username) {
      orientation
    }
  }
`;

const Dashboard = ({ username, streams }) => {
  const [virtualPosition, setVirtualPosition] = useState(null);
  const [virtualStream, setVirtualStream] = useState(null);

  // Whenever virtual position is updated, update the virtual stream.
  useEffect(() => {
    if (!virtualPosition) return;
    const audio = new Audio(droplet);
    audio.loop = true;
    audio.play();

    const acx = new AudioContext();
    const dst = acx.createMediaStreamDestination();
    const src = acx.createMediaElementSource(audio);
    src.connect(dst);

    setVirtualStream(dst.stream);
    return () => {
      src.disconnect();
      audio.remove();
      setVirtualStream(null);
    };
  }, [virtualPosition]);

  const { data, error } = useQuery(QUERY, {
    variables: { username: username },
    pollInterval: 100,
  });
  if (error) console.error(`[dashboard] failed to load player data`, error);
  const players = keyBy(data?.players, "username");

  // Preload position and orientation for current player.
  const position = get(players, username, {}).position;
  const orientation = data?.player?.orientation;

  // Calculates relative position.
  const relation = (position1, position2) => {
    if (!(position1 && position2 && orientation)) return undefined;
    const [x1, y1, z1] = position1;
    const [x2, y2, z2] = position2;
    const [ay, ax] = orientation;
    return rotate([x1 - x2, y2 - y1, z2 - z1], deg2rad(-ay), deg2rad(-ax));
  };

  return (
    <Container>
      <h1>{isEmpty(streams) ? "LOADING..." : "PLAYERS"} </h1>
      <Cards>
        {map(streams, (stream, targetUsername) => {
          const targetPlayer = get(players, targetUsername, {});
          const { position: targetPosition } = targetPlayer;
          const own = targetUsername === username;
          return (
            <AudioCard
              key={targetUsername}
              source={own ? SourceType.OUTGOING : SourceType.INCOMING}
              stream={stream}
              username={targetUsername}
              position={targetPosition}
              relation={own ? undefined : relation(position, targetPosition)}
              orientation={own ? orientation : undefined}
            />
          );
        })}
        {virtualStream ? (
          <AudioCard
            source={SourceType.VIRTUAL}
            stream={virtualStream}
            username="VIRTUAL"
            position={virtualPosition}
            relation={relation(position, virtualPosition)}
            onRemove={() => setVirtualPosition(null)}
          />
        ) : (
          <AddCard onClick={() => setVirtualPosition(position ?? null)} />
        )}
      </Cards>
    </Container>
  );
};

const ICE_SERVERS = [
  {
    urls: [
      "stun:stun.l.google.com:19302",
      "stun:stun1.l.google.com:19302",
      "stun:stun2.l.google.com:19302",
      "stun:stun3.l.google.com:19302",
      "stun:stun4.l.google.com:19302",
    ],
  },
];

class DashboardConnector extends Component {
  constructor(props) {
    super(props);
    this.conns = {};
    this.state = {
      streams: {},
    };
  }

  async componentDidMount() {
    const { username, socket } = this.props;
    if (!socket) return;

    socket.on("disconnect", () => {
      forEach(this.conns, (c) => c.close());
      this.conns = {};
    });

    // Handle registration events.
    socket.on("register", async ({ username: targetUsername, initiate }) => {
      try {
        if (targetUsername in this.conns) {
          console.warn(`[socket] already connected to '${targetUsername}'`);
          return;
        }

        const conn = new RTCPeerConnection({ iceServers: ICE_SERVERS });
        this.conns[targetUsername] = conn;
        this.setState(({ streams, ...otherState }) => ({
          streams: { ...streams, [targetUsername]: null },
          ...otherState,
        }));

        const negotiate = async () => {
          console.log(`[conn(${targetUsername})] negotiating connection...`);
          const desc = await conn.createOffer({
            offerToReceiveAudio: true,
            iceRestart: true,
          });
          await conn.setLocalDescription(desc);
          socket.emit("data", {
            recipient: targetUsername,
            payload: { description: desc },
          });
        };

        // Handle remote ICE candidates.
        conn.addEventListener("icecandidate", ({ candidate }) => {
          if (!candidate) return;
          console.log(`[conn(${targetUsername})] received ICE candidate`);
          socket.emit("data", {
            recipient: targetUsername,
            payload: {
              candidate: candidate.toJSON(),
            },
          });
        });

        const updateStream = () => {
          const { stream } = conn;
          this.setState(({ streams, ...otherState }) => ({
            ...otherState,
            streams: { ...streams, [targetUsername]: stream },
          }));
          console.log(`[conn(${targetUsername})] updated stream`);
        };

        // Handle remote tracks.
        conn.addEventListener("track", ({ streams }) => {
          const { connectionState } = conn;
          const [stream] = streams;
          conn.stream = stream;

          if (connectionState === "connected") updateStream();
          console.log(`[conn(${targetUsername})] received tracks`);
        });

        conn.addEventListener("connectionstatechange", () => {
          const { connectionState } = conn;
          console.log(
            `[conn(${targetUsername})] connection state is: ${connectionState}`
          );
          switch (connectionState) {
            case "connected":
              updateStream();
              break;
            case "failed":
              negotiate();
              break;
            default:
          }
        });

        // Send local tracks.
        {
          const { streams } = this.state;
          if (username in streams) {
            const stream = streams[username];
            stream.getTracks().forEach((track) => {
              conn.addTrack(track, stream);
            });
            console.log(`[conn(${targetUsername})] sent tracks`);
          }
        }

        // Create WebRTC offer, if initiator.
        if (initiate) negotiate();
        console.log(`[socket] connected to '${targetUsername}'`);
      } catch (error) {
        console.error(
          `[socket] failed to connect to '${targetUsername}:`,
          error
        );
      }
    });

    socket.on("deregister", async ({ username: targetUsername }) => {
      try {
        /** @type {RTCPeerConnection} */
        if (!(targetUsername in this.conns))
          throw new Error(`unknown username`);

        const { [targetUsername]: conn, ...otherConns } = this.conns;
        conn.close();
        this.conns = otherConns;

        const {
          [targetUsername]: stream,
          ...otherStreams
        } = this.state.streams;
        if (stream) {
          stream.getAudioTracks().forEach((track) => track.stop());
          this.setState({ streams: otherStreams });
        }

        console.log(`[socket] deregistered '${targetUsername}'`);
      } catch (error) {
        console.error(
          `[socket]: failed to deregister '${targetUsername}':`,
          error
        );
      }
    });

    socket.on("data", async ({ sender, payload }) => {
      try {
        if (!(sender in this.conns)) throw new Error(`unknown username`);
        /** @type {RTCPeerConnection} */
        const senderConn = this.conns[sender];

        const { description, candidate } = payload;
        if (description) {
          const desc = new RTCSessionDescription(description);
          await senderConn.setRemoteDescription(desc);
          if (desc.type === "offer") {
            const answer = await senderConn.createAnswer();
            await senderConn.setLocalDescription(answer);
            socket.emit("data", {
              recipient: sender,
              payload: { description: answer },
            });
          }
          console.log(`[socket] set session description from '${sender}'`);
          return;
        }

        if (candidate) {
          senderConn.addIceCandidate(new RTCIceCandidate(candidate));
          console.log(`[socket] set ICE candidate from '${sender}'`);
          return;
        }
      } catch (error) {
        console.error(`[socket]: handle data from '${sender}':`, error);
      }
    });

    try {
      const constraints = { audio: true, video: false };
      const stream = await navigator.mediaDevices.getUserMedia(constraints);
      this.setState(({ streams, ...otherState }) => ({
        streams: { [username]: stream, ...streams },
        ...otherState,
      }));
      forEach(this.conns, (conn) => {
        stream.getTracks().forEach((track) => {
          conn.addTrack(track, stream);
        });
      });
      console.info(`[audio] configured and streaming`);
    } catch (error) {
      alert("Failed to configure audio.");
      console.error(`[audio]`, error);
      return;
    }
  }

  componentWillUnmount() {
    const stream = this.state.streams[this.props.username];
    if (stream) stream.getAudioTracks().forEach((track) => track.stop());
  }

  render() {
    const { streams } = this.state;
    return <Dashboard streams={streams} {...this.props} />;
  }
}

export default DashboardConnector;
