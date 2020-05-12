/** @jsx jsx */
import { jsx, css } from "@emotion/core";
import { useRef, useState, useEffect } from "react";
import styled from "@emotion/styled";

import { Mic, MicOff, Volume1, VolumeX, Trash } from "react-feather";

import Visualizer from "./visualizer";
import Card from "./card";

const StyledCard = styled(Card)`
  color: #ababab;
  background: ${({ darker }) => (darker ? "black" : "#2F2F2F")};
  box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.5);

  code {
    font-size: 0.9rem;
    font-weight: 700;

    &.dimmed {
      margin-right: 02rem;
      color: #4e4e4e;
    }
  }
`;

const Expanded = styled.div`
  flex: 1;
`;

const Row = styled.div`
  width: 100%;
  display: flex;
`;

const Column = styled.div`
  display: flex;
  flex-direction: column;
`;

const StyledVisualizer = styled(Visualizer)`
  height: 3rem;
  width: 60%;
  margin-bottom: 0.3rem;
`;

const Clickable = styled.span`
  cursor: pointer;
`;

export const SourceType = {
  VIRTUAL: "virtual",
  OUTGOING: "outgoing",
  INCOMING: "incoming",
};

const Remove = styled(Trash)`
  margin-right: 0.1rem;

  --size: 1.15rem;
  color: #ff3455;
  width: var(--size);
  height: var(--size);

  cursor: pointer;
  transition: color ease-in-out 200ms;
  &:hover {
    color: #ff637c;
  }
`;

const SoundSwitch = ({ source, disabled, ...otherProps }) => {
  const props = {
    css: css`
      color: #ff3455;
      height: ${source === SourceType.OUTGOING ? 1.25 : 1.4}rem;
      transition: color 200ms ease-in-out;
      &:hover {
        color: #ff637c;
      }
    `,
    ...otherProps,
  };
  return (
    <Clickable>
      {source === SourceType.OUTGOING ? (
        disabled ? (
          <MicOff {...props} />
        ) : (
          <Mic {...props} />
        )
      ) : disabled ? (
        <VolumeX {...props} />
      ) : (
        <Volume1 {...props} />
      )}
    </Clickable>
  );
};

const AudioCard = ({
  source,
  stream,
  username,
  position,
  relation,
  orientation,
  onRemove,
}) => {
  const audio = useRef(null);
  const [output, setOutput] = useState(stream);
  const [panner, setPanner] = useState(null);

  useEffect(() => {
    if (!stream) return;
    if (source === SourceType.OUTGOING) return;

    // HACK: Allow stream modification from peer.
    {
      let audio = new Audio();
      audio.muted = true;
      audio.srcObject = stream;
      audio.addEventListener("canplaythrough", () => {
        audio = null;
      });
      audio.srcObject = stream;
    }

    const AudioContext = window.AudioContext ?? window.webkitAudioContext;
    /** @type {AudioContext} */
    const acx = new AudioContext();

    const panner = acx.createPanner();
    panner.distanceModel = "linear";
    panner.panningModel = "HRTF";
    panner.maxDistance = 3000; // 30 blocks
    setPanner(panner);

    const dst = acx.createMediaStreamDestination();
    const src = acx.createMediaStreamSource(stream);
    src.connect(panner).connect(dst);
    setOutput(dst);
    audio.current.srcObject = dst.stream;

    // When cleaning up, disconnect everything.
    return () => {
      src.disconnect();
      panner.disconnect();
      setPanner(null);
    };
  }, [stream, source]);

  // Panner updates.
  useEffect(() => {
    if (!(panner && relation)) return;
    const [x, y, z] = relation.map((x) => parseInt(x * 100));
    panner.setPosition(x, y, z);
  }, [panner, relation]);

  const [track] = stream.getAudioTracks();
  const [disabled, setDisabled] = useState(false);
  useEffect(() => {
    track.enabled = !disabled;
  }, [track, disabled]);

  const formatInfo = (info, unit = "") => {
    if (!info) return "unknown";
    return info.map((x) => `${parseInt(x)}${unit}`).join(" ");
  };
  return (
    <StyledCard darker={source === SourceType.OUTGOING}>
      <audio ref={audio} autoPlay={source !== SourceType.OUTGOING} />
      <Row>
        <Column>
          <code>{formatInfo(position)}</code>
          {relation && <code className="dimmed">{formatInfo(relation)}</code>}
          {orientation && (
            <code className="dimmed">{formatInfo(orientation, "°")}</code>
          )}
        </Column>
        <Expanded />
        <SoundSwitch
          source={source}
          disabled={disabled}
          onClick={() => setDisabled(!disabled)}
        />
      </Row>
      <Expanded />
      <StyledVisualizer stream={output} color="#ababab" />
      <Row>
        <h1>{username}</h1>
        <Expanded />
        {source === SourceType.VIRTUAL && <Remove onClick={onRemove} />}
      </Row>
    </StyledCard>
  );
};

export default AudioCard;
