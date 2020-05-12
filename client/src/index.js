import React from "react";
import ReactDOM from "react-dom";
import { Global, css } from "@emotion/core";

import * as serviceWorker from "./serviceWorker";

import App from "./app";

ReactDOM.render(
  <React.StrictMode>
    <Global
      styles={css`
        body {
          margin: 0;
          font-family: Inter, -apple-system, BlinkMacSystemFont, "Segoe UI",
            "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans",
            "Helvetica Neue", sans-serif;
          -webkit-font-smoothing: antialiased;
          -moz-osx-font-smoothing: grayscale;
        }

        /* prettier-ignore */
        code, .mono {
            font-family: "IBM Plex Mono", source-code-pro, Menlo, Monaco,
              Consolas, "Courier New", monospace;
          }
      `}
    />
    <App />
  </React.StrictMode>,
  document.getElementById("root")
);

if (typeof window !== "undefined") {
  window.ZOOMCRAFT_SKIP_VALIDATION = false;
  window.ZOOMCRAFT_MAX_DISTANCE = 25;
}

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
