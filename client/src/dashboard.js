import React, { useEffect, useState } from "react";
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

const ICE_SERVERS = [{ urls: ["stun:stun.l.google.com:19302"] }];

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

const Dashboard = ({ username, socket }) => {
  const [conns, setConns] = useState({});
  const [streams, setStreams] = useState({});
  const [virtualPosition, setVirtualPosition] = useState(null);

  // Initialize audio, configure socket API.
  useEffect(
    () => {
      if (!socket) return;

      // Close all WebRTC connections on disconnect.
      socket.on("disconnect", () => {
        forEach(conns, (c) => c.close());
        setConns({});
      });

      // Handle registration events.
      socket.on("register", async ({ username: otherUsername, initiate }) => {
        try {
          /** @type {{conns: RTCPeerConnection[], stream: MediaStream}} */
          if (otherUsername in conns) {
            console.warn(`[socket] already connected to '${otherUsername}'`);
            return;
          }

          const conn = new RTCPeerConnection({ iceServers: ICE_SERVERS });
          setConns({ ...conns, [otherUsername]: conn });

          // Handle remote ICE candidates.
          conn.addEventListener("icecandidate", ({ candidate }) => {
            if (!candidate) return;
            console.log(`[conn(${otherUsername})] received ICE candidate`);
            socket.emit("data", {
              recipient: otherUsername,
              payload: {
                candidate: candidate.toJSON(),
              },
            });
          });

          // Handle remote streamms.
          conn.addEventListener("addstream", ({ stream }) => {
            setStreams({ ...stream, [otherUsername]: stream });
            console.log(`[conn(${otherUsername})] received stream`);
          });

          // Send local streams.
          if (username in streams) conn.addStream(streams[username]);

          // Create WebRTC offer, if initiator.
          if (initiate) {
            const desc = await conn.createOffer({ offerToReceiveAudio: true });
            await conn.setLocalDescription(desc);
            socket.emit("data", {
              recipient: otherUsername,
              payload: {
                description: desc.toJSON(),
              },
            });
          }

          console.log(`[socket] connected to '${otherUsername}'`);
        } catch (error) {
          console.error(
            `[socket] failed to connect to '${otherUsername}:`,
            error
          );
        }
      });

      socket.on("deregister", async ({ username: otherUsername }) => {
        try {
          /** @type {RTCPeerConnection} */
          if (!(otherUsername in conns)) throw new Error(`unknown username`);

          const { [otherUsername]: conn, ...otherConns } = conns;
          conn.close();
          setConns(otherConns);

          console.log(`[socket] registered '${otherUsername}'`);
        } catch (error) {
          console.error(
            `[socket]: failed to deregister '${otherUsername}':`,
            error
          );
        }
      });

      socket.on("data", async ({ sender, payload }) => {
        try {
          if (!(sender in conns)) throw new Error(`unknown username`);

          /** @type {RTCPeerConnection} */
          const senderConn = conns[sender];

          const { description, candidate } = payload;
          if (description) {
            const desc = new RTCSessionDescription(description);
            await senderConn.setRemoteDescription(desc);
            if (desc.type === "offer") {
              const answer = await senderConn.createAnswer();
              await senderConn.setLocalDescription(answer);
              socket.emit("data", {
                recipient: sender,
                payload: { description: answer.toJSON() },
              });
            }
            console.log(`[socket] set session description from '${sender}'`);
          } else if (candidate) {
            senderConn.addIceCandidate(new RTCIceCandidate(candidate));
            console.log(`[socket] set ICE candidate from '${sender}'`);
          }
        } catch (error) {
          console.error(`[socket]: handle data from '${sender}':`, error);
        }
      });

      async function setup() {
        try {
          const constraints = { audio: true, video: false };
          const stream = await navigator.mediaDevices.getUserMedia(constraints);
          setStreams({ ...streams, [username]: stream });
          forEach(conns, (c) => c.addStream(stream));
          console.info(`[audio] configured and streaming`);
        } catch (error) {
          alert("Failed to configure audio.");
          console.error(`[audio]`, error);
          return;
        }
      }
      setup();

      // Stop all tracks when tearing down.
      return () => {
        /** @type {MediaStream} */
        const stream = streams[username];
        if (stream) stream.getAudioTracks().forEach((track) => track.stop());
      };
    },
    /* eslint-disable-line react-hooks/exhaustive-deps */ [socket]
  );

  // Whenever virtual position is updated, update the virtual stream.
  const [virtualStream, setVirtualStream] = useState(null);
  useEffect(() => {
    if (!virtualPosition) return;
    const audio = new Audio(droplet);
    audio.loop = true;

    const acx = new AudioContext();
    const dst = acx.createMediaStreamDestination();
    const src = acx.createMediaElementSource(audio);
    src.connect(dst);

    setVirtualStream(dst.stream);
    audio.play();
    return () => {
      src.disconnect();
      setVirtualStream(null);
    };
  }, [virtualPosition]);

  const { data, error } = useQuery(QUERY, {
    variables: { username: username },
    pollInterval: 250,
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
        {map(streams, (stream, streamUsername) => {
          const streamPlayer = get(players, streamUsername, {});
          const { position: streamPosition } = streamPlayer;
          const own = streamUsername === username;
          return (
            <AudioCard
              key={streamUsername}
              source={own ? SourceType.OUTGOING : SourceType.INCOMING}
              stream={stream}
              username={streamUsername}
              position={position}
              relation={own ? undefined : relation(position, streamPosition)}
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

export default Dashboard;
