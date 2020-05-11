const dotenv = require("dotenv");
dotenv.config();

const express = require("express");
const { createProxyMiddleware } = require("http-proxy-middleware");

const { GATEWAY_PORT, BACKEND_PORT, CLIENT_PORT, CLIENT_PATH } = process.env;

// Create app.
const app = express();

// Proxy requests to /api to external backend.
app.use(
  "/api",
  createProxyMiddleware({
    target: `http://localhost:${BACKEND_PORT}`,
    pathRewrite: { "^/api": "" },
  })
);

// Serve static client files.
if (CLIENT_PORT) {
  app.use(
    createProxyMiddleware({
      target: `http://localhost:${CLIENT_PORT}`,
      ws: true,
    })
  );
} else if (CLIENT_PATH) {
  app.use(express.static(CLIENT_PATH));
}

// Create gateway server.
const server = require("http").createServer(app);

// Register socket server.
const socket = require("./socket");
socket(server);

// Start gateway server.
const port = GATEWAY_PORT || 8080;
server.listen(port, () => console.log(`[gateway]`, `listening on :${port}`));
