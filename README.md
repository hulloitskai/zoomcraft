# zoomcraft

_Virtual conferencing with realistic audio presence, in Minecraft._

[![Drone][drone-img]][drone]

> Working from home? Want to spice up your team meetings a bit?
>
> Why not do it—in Minecraft!

<br />

`zoomcraft` is a single-container service that augments your Minecraft server
with real-time audio conferencing superpowers using
[WebRTC](https://webrtc.org). It works with any Minecraft server with RCON
capabilities (vanilla servers since 1.0.0).

It uses Web Audio spatialization APIs to map the audio from other players to
their in-game position, in order to create a realistic virtual presence.

[Check out a screencast of `zoomcraft`'s 3D audio
capabilties.](https://vimeo.com/417468259) Use headphones! (or make sure your
speakers support stereo audio).

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

## Caveats

- Not all browser fully support the Web Audio spec. This service was developed
  and tested on Chrome, so that is the recommended browser to use with
  `zoomcraft`.

- Some bluetooth headphones have limited audio channels, so when the microphone
  is active, stereo audio output is disabled (meaning that 3D audio effects will
  not work). If this happens, change your computer's audio input source to be
  something other than the bluetooth headphones (e.g. your computer's internal
  speakers).

## Testing

Couldn't manage to convince any friends to hang out with you on Minecraft?
Create a virtual player to test the platform!

When added, the `VIRTUAL` player will spawn at your current player location,
and make intermittent sounds so that you can test the platform's 3D audio
capabilities during solo testing / development.

## TODO

1. Allow user to set a roll-off modifier.
2. "Rooms" that restrict communication to the players within an in-game
   geofence.

[drone]: https://ci.stevenxie.me/stevenxie/zoomcraft
[drone-img]: https://ci.stevenxie.me/api/badges/stevenxie/zoomcraft/status.svg
