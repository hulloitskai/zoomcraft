const socket = require("socket.io");

/**
 * A mapping of player usernames to sockets.
 */
const sockets = {};

module.exports = (server) => {
  const io = socket(server, { path: "/api/socket" });

  const socklog = (socket) => {
    const { id, username } = socket;
    const prefix = username ? `[socket(${username})` : `[socket/${id}]`;
    return {
      log: (...args) => console.log(prefix, ...args),
      error: (...args) => console.error(prefix, ...args),
    };
  };

  // Upon connection, register socket event handlers.
  io.on("connect", (socket) => {
    socklog(socket).log("connected");

    socket.on("error", (error) => {
      socklog(socket).error(`unexpected error: ${error}`);
    });

    // Handle socket registration using a player username.
    socket.on("register", ({ username }, reply) => {
      if (username in sockets) {
        const message = "username already registered";
        socklog(socket).error(message);
        return reply({ error: message });
      }
      reply({});

      // Register username.
      socket.username = username;
      sockets[username] = socket;

      // Tell participants about each other.
      setTimeout(() => {
        for (otherUsername in sockets) {
          if (otherUsername === username) continue;
          socket.emit("register", {
            username: otherUsername,
            initiate: true,
          });
          sockets[otherUsername].emit("register", {
            username,
            initiate: false,
          });
        }
        socklog(socket).log(`registered as '${username}'`);
      }, 1000);
    });

    // Data relay.
    socket.on("data", ({ recipient, payload }) => {
      const { username } = socket;
      sockets[recipient].emit("data", { sender: username, payload });
    });

    // Upon disconnect, emit deregister event to all participants.
    socket.on("disconnect", () => {
      const { username } = socket;
      if (!username) {
        socklog(socket).error(`disconnected unregistered socket`);
        return;
      }

      delete sockets[username];
      io.emit("deregister", { username });
      socklog(socket).log(`disconnected and deregistered`);
    });
  });
};
