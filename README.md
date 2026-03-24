# Roberts Radio Control (Go Version)

A Go command-line tool for controlling Roberts radios via the UNDOK FSAPI protocol.

## Features

- Power Control (On/Off/Standby)
- Volume Control (Set, Up, Down, Mute, Toggle Mute)
- Source Mode Selection (Internet Radio, Spotify, Bluetooth, DAB, etc.)
- Playback Control (Play, Pause, Next, Previous)
- Device Information & Now Playing status

## Installation

Ensure you have [Go](https://go.dev/doc/install) installed.

```bash
go build -o radio_control radio_control.go
```

## Usage

```bash
./radio_control [flags] [command] [value]
```

### Flags

- `-i, --ip <addr>`: IP address of the radio (default: `XX.XX.XX.XX` or `RADIO_IP` environment variable)
- `-h, --help`: Show help message

### Commands

- `on`: Turn the radio ON
- `off`: Turn the radio OFF (Standby)
- `status`: Check power status
- `vol [0-32]`: Set volume or get current volume
- `volup`: Increase volume by 1
- `voldown`: Decrease volume by 1
- `mute`: Mute the radio
- `unmute`: Unmute the radio
- `togglemute`: Toggle mute state
- `mode [id]`: Set source mode or get current mode
- `next`: Next track or station
- `prev`: Previous track or station
- `play`: Start playback
- `pause`: Pause playback
- `info`: Show "Now Playing" information
- `pair`: Initiate Bluetooth pairing
- `device`: Show device information (Model, Version, IP)

## License

MIT
