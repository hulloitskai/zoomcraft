import React, { useState, useEffect } from "react";
import styled from "@emotion/styled";

import io from "socket.io-client";

import {
  ApolloClient,
  ApolloProvider,
  HttpLink,
  InMemoryCache,
} from "@apollo/client";

import Intro from "./intro";
import Dashboard from "./dashboard";

const Container = styled.div`
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: stretch;
`;

const App = () => {
  const client = new ApolloClient({
    cache: new InMemoryCache(),
    link: new HttpLink({ uri: "./api/graphql" }),
  });

  const [socket, setSocket] = useState(null);
  const [username, setUsername] = useState(null);

  // Initialize socket API, handle builtin events.
  useEffect(
    () => {
      const socket = io({ path: "/api/socket" });
      socket.on("error", (error) => {
        console.error("[socket]", error);
      });
      socket.on("connect", () => {
        console.info("[socket] connected");
        setSocket(socket);
      });
      socket.on("disconnect", () => {
        setUsername(null);
        console.info("[socket] disconnected");
      });
      return socket.disconnect;
    },
    /* eslint-disable-line react-hooks/exhaustive-deps */ []
  );

  return (
    <ApolloProvider client={client}>
      <Container>
        {username ? (
          <Dashboard
            socket={socket}
            username={username}
            onDisconnect={() => setUsername(null)}
          />
        ) : (
          <Intro
            socket={socket}
            onSubmit={({ username }) => setUsername(username)}
          />
        )}
      </Container>
    </ApolloProvider>
  );
};

export default App;
