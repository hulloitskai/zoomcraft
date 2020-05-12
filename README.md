# covidcraft

_Virtual conferencing with realistic audio presence, in Minecraft._

> Working from home? Want to spice up your team meetings a bit?
>
> Why not do it—in Minecraft!

<br />

`covidcraft` is a single-container service that augments your Minecraft server
with real-time audio conferencing superpowers using
[WebRTC](https://webrtc.org). It works with any Minecraft server with RCON
capabilities (vanilla servers since 1.0.0).

It uses Web Audio spatialization APIs to map the audio from other players to
their in-game position, in order to create a realistic virtual presence.

[Check out a screencast of `covidcraft`'s 3D audio
capabilties.](https://vimeo.com/417468259)

## Usage

1. Run the Docker container on a server with access to the Minecraft game
   server (vanilla / any variant with RCON access).

   ```bash
   docker run \
     -p 8080:8080 \
     -e RCON_ADDRESS=http://localhost:25575 \
     -e RCON_PASSWORD=minecraft \
     stevenxie/covidcraft
   ```

2. Join the Minecraft server.
3. Visit `http://localhost:8080` and enter your player username to begin
   conferencing.

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
