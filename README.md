# zoomcraft

_Realistic audio presence, in Minecraft._

[![git gag][tag-img]][tag]
[![drone ci][drone-img]][drone]

> Are you working from home? Spending all day on Zoom calls?
>
> Missing that presence that you feel when everybody's working in the same
> office?
>
> Well, why not recapture that feeling by bringing everybody together—in
> Minecraft!

`zoomcraft` is a single-container service that augments your Minecraft server
with real-time 3D audio presence over [WebRTC](https://webrtc.org). It works
with any Minecraft server with RCON capabilities (vanilla servers since 1.0.0).

It uses [web audio spatialization APIs](https://developer.mozilla.org/en-US/docs/Web/API/Web_Audio_API/Web_audio_spatialization_basics)
to map the audio from other players to their in-game position, in order to
create a realistic virtual presence.

[Check out a screencast of `zoomcraft`'s 3D audio
capabilties.](https://vimeo.com/417864067) _Use headphones! (or make sure your
speakers support stereo audio)._

## Usage

1. Run the Docker image on a server with access to the Minecraft game
   server (vanilla / any variant with RCON access).

   ```bash
   docker run \
     -p 8080:8080 \
     -e RCON_ADDRESS=http://localhost:25575 \
     -e RCON_PASSWORD=minecraft \
     stevenxie/zoomcraft
   ```

2. Join the Minecraft server.
3. Visit `http://localhost:8080` and enter your player username to begin
   conferencing.
4. Expose port `8080` on a public address to allow other players to
   connect to `zoomcraft`. Have fun!

> If you experience any connection issues or other bugs, please
> [submit an issue](https://github.com/stevenxie/zoomcraft/issues/new/choose) so
> that I can look into it!

### Caveats

- Not all browser fully support the web audio spec. This service was developed
  and tested on Chrome, so that is the recommended browser to use with
  `zoomcraft`.

- Some bluetooth headphones have limited audio channels, so when the microphone
  is active, stereo audio output is disabled (meaning that 3D audio effects will
  not work). If this happens, change your computer's audio input source to be
  something other than the bluetooth headphones (e.g. your computer's internal
  speakers).

- Some Chrome browser extensions may interfere with the WebRTC connection
  negotiation process. Run `zoomcraft` in an incognito window if you
  encounter connection issues.

- ~~The first connection attempt sometimes takes much longer than future
  attempts. Be prepared for other players to be `connecting` for up to 30
  seconds.~~ This should be a lot better starting from `v1.1.1` due to the
  implementation of WebRTC negotiation timeouts. Still happening?
  [File an issue!](https://github.com/stevenxie/zoomcraft/issues/new/choose)

## Architecture

`zoomcraft` consists of three components: `backend`, `gateway`, and `client`:

- [`backend`](./backend) is responsible for querying the Minecraft server for
  world and player data using [RCON](https://wiki.vg/RCON). It exposes this
  information over a [`GraphQL`](https://graphql.org/) API.

  > Interested in forking `zoomcraft` to support another game? This is the
  > code that you should probably change!
  >
  > In particular, check out
  > [`backend/minecraft/player_service.go`](./backend/minecraft/player_service.go)
  > for an implementation of getting game data through `RCON`, and
  > [`backend/graphql/minecraft.resolvers.go`](./backend/graphql/minecraft.resolvers.go)
  > to see how that service is called from the `GraphQL` layer.

- [`gateway`](./gateway) serves both `client` and `backend`, and takes care of
  connection routing. In particular, it:
  - Routes `/api/graphql` and `/api/graphiql` to `backend`.
  - Routes `/*` to `client`.
  - Serves a [`socket.io`](https://socket.io/) server at `/api/socket` to
    relay WebRTC connection information between clients.
- [`client`](./client) is a [React](https://reactjs.org/) frontend that
  exchanges audio with other clients using [WebRTC](https://webrtc.org/), and
  applies 3D effects with the
  [web audio API](https://developer.mozilla.org/en-US/docs/Web/API/Web_Audio_API)
  using data from `backend`.

## Advanced Usage

### Virtual Player

Couldn't manage to convince any friends to hang out with you on Minecraft?
Create a virtual player to test the platform!

When added, the `VIRTUAL` player will spawn at your current player location,
and make intermittent sounds so that you can test the platform's 3D audio
capabilities during solo testing / development.

### Client Overrides

The following global variables can be used to alter the behavior on `client`,
by typing them into the browser console.

_You must apply them before connecting in order for them to take effect._

- To skip player-validation in order to test audio conferencing capabilities
  without Minecraft:

  ```js
  ZOOMCRAFT_SKIP_VALIDATION = true;
  ```

- To change the maximum audible distance (after which other players are no
  longer audible):

  ```js
  ZOOMCRAFT_MAX_DISTANCE = /* distance in blocks */
  ```

- To change the rate at which player position data is updated:

  ```js
  ZOOMCRAFT_POLL_INTERVAL = /* duration in milliseconds */
  ```

- To use custom ICE servers for WebRTC:

  ```js
  ZOOMCRAFT_ICE_SERVERS = [
    {
      urls: [
        /* ... */
      ],
    },
  ];
  ```

- To change the WebRTC negotiation timeout:

  ```js
  ZOOMCRAFT_NEGOTIATION_TIMEOUT = /* duration in milliseconds */
  ```

## TODO

1. Create a UI for changing the maximum audible distance.
2. Implement "rooms" that restrict communication to the players within an
   in-game geofence.

[drone]: https://ci.stevenxie.me/stevenxie/zoomcraft
[drone-img]: https://ci.stevenxie.me/api/badges/stevenxie/zoomcraft/status.svg
[tag]: https://github.com/stevenxie/zoomcraft/tags
[tag-img]: https://img.shields.io/github/v/tag/stevenxie/zoomcraft?label=latest&color=black&sort=semver
