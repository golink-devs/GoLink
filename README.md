<p align="center">
  <img src="https://raw.githubusercontent.com/golink-devs/.github/main/profile/file_0000000004c87208ab33c65ea5c8a466.png" width="200" alt="GoLink Logo">
</p>

# GoLink

GoLink is a **Lavalink-compatible audio sending node** written in Go. It acts as a drop-in replacement for Lavalink v4, allowing Discord bots to connect and stream audio with high performance and low overhead.

## Features

- **Lavalink v4 Compatible**: Support for all standard Lavalink v4 REST and WebSocket APIs.
- **DAVE E2EE**: Native support for Discord's End-to-End Encryption (DAVE/MLS).
- **High Performance**: Written in Go with FFmpeg and Opus for efficient audio processing.
- **Multi-Source Support**:
  - YouTube (via `yt-dlp`)
  - Spotify (Metadata + YouTube search)
  - Direct HTTP streams
- **Audio Filters**: Volume, Timescale (Speed/Pitch), LowPass, and more.
- **Prometheus Metrics**: Built-in monitoring at `/v1/metrics`.

## Installation

### Prerequisites

- **Go 1.22+**
- **FFmpeg** (installed on system path)
- **yt-dlp** (installed on system path)
- **libdave**: Discord's official C++ DAVE implementation (required for E2EE).

### Install libdave

Run the installation script provided in the `scripts/` directory:

```bash
./scripts/libdave_install.sh v1.1.0
```

### Build

```bash
CGO_ENABLED=1 go build -o golink ./cmd/golink
```

## Running GoLink

Place a `config.yml` file in the same directory as the binary and run:

```bash
./golink
```

### Configuration (config.yml)

```yaml
server:
  port: 2333
  host: 0.0.0.0
  password: "youshallnotpass"

sources:
  youtube: true
  spotify: true
  spotifyClientID: "YOUR_CLIENT_ID"
  spotifyClientSecret: "YOUR_CLIENT_SECRET"
  http: true

cache:
  enabled: true
  ttl: 3600

metrics:
  enabled: true

logging:
  level: INFO
```

## Docker

You can also run GoLink using Docker:

```bash
docker-compose up -d
```

## Test Bot

A basic Discord bot for testing GoLink is available in the `test_bot/` directory.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
