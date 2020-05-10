import React from "react";
import { Global, css } from "@emotion/core";

import RTC from "./RTC";

function App() {
  const apiUrl = new URL(process.env.REACT_APP_API_URL);
  const { protocol: apiProtocol, host, pathname } = apiUrl;

  const wsProtocol = apiProtocol.startsWith("https") ? "wss" : "ws";
  const wsPath = pathname.endsWith("/") ? pathname : `${pathname}/`;
  const wsUrl = `${wsProtocol}://${host}${wsPath}ws`;

  return (
    <>
      <Global
        styles={css`
          body {
            margin: 0;
            font-family: Inter, -apple-system, BlinkMacSystemFont, "Segoe UI",
              "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans",
              "Droid Sans", "Helvetica Neue", sans-serif;
            -webkit-font-smoothing: antialiased;
            -moz-osx-font-smoothing: grayscale;
          }

          @supports (font-variation-settings: normal) {
            body {
              font-family: "Inter var", -apple-system, BlinkMacSystemFont,
                "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell",
                "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
            }
          }

          /* prettier-ignore */
          code, .mono {
            font-family: "IBM Plex Mono", source-code-pro, Menlo, Monaco,
              Consolas, "Courier New", monospace;
          }
        `}
      />
      <div className="app">
        <RTC wsUrl={wsUrl} />
      </div>
    </>
  );
}

export default App;
