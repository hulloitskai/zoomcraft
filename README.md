# zoomcraft

_Augmented reality voice conferencing, in Minecraft._

[![Drone][drone-img]][drone]

> Working from home? Want to spice up your team meetings a bit?
>
> Why not do it—in Minecraft!

`zoomcraft` is a single-container service that augments your Minecraft server
with real-time audio conferencing superpowers using
[WebRTC](https://webrtc.org). It works with any Minecraft server with RCON
capabilities (vanilla servers since 1.0.0).

It uses Web Audio spatialization APIs to map the audio from other players to
their in-game position, in order to create a realistic virtual presence.

[Check out a screencast of `zoomcraft`'s 3D audio
capabilties.](https://vimeo.com/417468259) _Use headphones! (or make sure your
speakers support stereo audio)._

## Usage

1. Run the Docker container on a server with access to the Minecraft game
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

### Caveats

- Not all browser fully support the Web Audio spec. This service was developed
  and tested on Chrome, so that is the recommended browser to use with
  `zoomcraft`.

- Some bluetooth headphones have limited audio channels, so when the microphone
  is active, stereo audio output is disabled (meaning that 3D audio effects will
  not work). If this happens, change your computer's audio input source to be
  something other than the bluetooth headphones (e.g. your computer's internal
  speakers).

## Architecture

`zoomcraft` consists of three components: `backend`, `gateway`, and `client`:

- [`backend`](./backend) is responsible for querying the Minecraft server for
  world and player data using [RCON](https://wiki.vg/RCON). It serves a
  [`GraphQL`](https://graphql.org/) API.
- [`gateway`](./gateway) serves both `client` and `backend`, and takes care of
  connection routing. In particular, it:
  - Routes `/api/graphql` and `/api/graphiql` to `backend`.
  - Routes `/*` to `client`.
  - Serves a [`socket.io`](https://socket.io/) server at `/api/socket` to
    relay RTC connection information between clients.j
- [`client`](./client) is a [`React`](https://reactjs.org/) frontend that
  exchanges audio with other clients using [`WebRTC`](https://webrtc.org/), and
  applies 3D effects using the
  [Web Audio API](https://developer.mozilla.org/en-US/docs/Web/API/Web_Audio_API)
  with data from `backend`.

## Testing

### Virtual Player

Couldn't manage to convince any friends to hang out with you on Minecraft?
Create a virtual player to test the platform!

When added, the `VIRTUAL` player will spawn at your current player location,
and make intermittent sounds so that you can test the platform's 3D audio
capabilities during solo testing / development.

### Backdoors

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

## TODO

1. Allow user to set a roll-off modifier.
2. "Rooms" that restrict communication to the players within an in-game
   geofence.

[drone]: https://ci.stevenxie.me/stevenxie/zoomcraft
[drone-img]: https://ci.stevenxie.me/api/badges/stevenxie/zoomcraft/status.svg
