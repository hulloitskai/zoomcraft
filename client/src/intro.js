import React, { useRef, useState, useEffect } from "react";
import styled from "@emotion/styled";

import { useApolloClient, gql } from "@apollo/client";

const Container = styled.div`
  background: black;
  flex: 1;

  display: flex;
  align-items: center;
  justify-content: center;
`;

const Menu = styled.div`
  color: white;

  display: flex;
  align-items: stretch;
  flex-direction: column;

  /* prettier-ignore */
  h1, h2 { margin: 0; }

  h1 {
    font-size: 2.4rem;
    font-weight: 800;
  }

  h2 {
    margin-top: 0.1rem;
    font-size: 1.4rem;
    color: #9c9c9c;
  }

  form {
    display: flex;
    flex-direction: column;
    align-items: stretch;
  }

  input {
    &.username {
      padding: 0.7rem;
      margin-top: 1.5rem;

      border: none;
      color: #4a4a4a;
      background: #b8b8b8;
      font-size: 1.2rem;
      font-weight: 600;

      &::placeholder {
        color: #7b7b7b;
      }

      transition: background ease-in-out 200ms;
      &:hover,
      &:focus {
        background: #d8d8d8;
      }
    }

    &.connect {
      align-self: flex-end;
      margin-top: 0.7rem;
      background: white;
      border: none;
      padding: 0.7rem;
      font-weight: 700;
      cursor: pointer;

      transition: background ease-in-out 200ms;
      &:hover {
        background: #e2e2e2;
      }
    }
  }
`;

const PLAYER_USERNAMES = gql`
  query {
    players {
      username
    }
  }
`;

const Intro = ({ socket, onSubmit }) => {
  const input = useRef(null);
  const client = useApolloClient();

  const [disabled, setDisabled] = useState(!socket);
  useEffect(() => setDisabled(!socket), [socket]);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setDisabled(true);
    const { value: username } = input.current;

    // Check valid usernames.
    const { data } = await client.query({
      query: PLAYER_USERNAMES,
      fetchPolicy: "network-only",
    });
    const usernames = data.players.map((p) => p.username);
    if (!usernames.includes(username)) {
      setDisabled(false);
      return alert("No such player was found.");
    }

    // Register with username.
    socket.emit("register", { username }, ({ error }) => {
      if (error) {
        alert(`Failed to connect: ${error.toString()}`);
        setDisabled(false);
      } else if (onSubmit) {
        onSubmit({ username });
      }
    });
  };

  return (
    <Container>
      <Menu>
        <h1>COVIDCRAFT</h1>
        <h2>virtual conferencing in minecraft</h2>
        <form onSubmit={handleSubmit}>
          <input
            className="username"
            placeholder="player username"
            text="text"
            ref={input}
          />
          <input
            className="connect"
            disabled={disabled}
            value="CONNECT"
            type="submit"
          />
        </form>
      </Menu>
    </Container>
  );
};

export default Intro;
