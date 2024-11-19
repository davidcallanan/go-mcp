# Development

## Setup

 - Install `git`.
 - Install Go version `^1.23.2`.
 - Clone this repository using `git clone https://github.com/davidcallanan/go-mcp`.

## Run Test Server

 - `cd testserver`
 - `go run main.go`
 - Launch Minecraft 1.14.
 - Click on "Multiplayer -> Add Server", and enter `localhost` as the server address.
 - Connect to the server.

## Project Structure

The folder `testserver` implements a simple server using our custom protocol. It implements the most high-level functionality such as player movement and the server status message.

The package `src/javaserver` implements the low-level state management and packet orchestration of a Java Edition server, and exposes a high-level API for use by a server implementation.

The package `src/javaio` implements the low-level packet encoding and decoding. It has a particular file structure.

 - The `pkt_VVVV_<packet_name>.go` files implement the raw encoding and decoding of particular packets, where `VVVV` is our custom hexidecimal version number from when this packet type was first introduced. See `docs/protocol_versions.md`.
 - The `type_<type_name>.go` files implement the encoding and decoding of particular data types, and these data types are reused across packets.
 - Some packets are organized by the state the server must be in at the moment that these packets make sense. They are organized into `packets_<state>.go`, `parse_<state>.go` and `emit_<state>.go` files.