# btsw

A tiny Linux CLI that quickly connects and disconnects paired bluetooth devices.

- List Bluetooth adapters
- List paired devices on any bluetooth adapter
- Connect/disconnect paired devices by id 

btsw talks to BlueZ over the system D‑Bus using [github.com/godbus/dbus](github.com/godbus/dbus).

## Requirements

- Linux with BlueZ running (bluetoothd on the system bus)

Note On NixOS. Enable Bluetooth in your system configuration:
> hardware.bluetooth.enable = true

## Install

### Nix flakes

Run with flakes:
- `nix run https://flakes.adriano.fyi/btsw -- adapters`
- `nix run https://flakes.adriano.fyi/btsw -- devices`
- `nix run https://flakes.adriano.fyi/btsw -- connect 0`
- `nix run https://flakes.adriano.fyi/btsw -- disconnect 4`

### Flake input

```
inputs.btsw = {
  url = "https://flakes.adriano.fyi/btsw";
  inputs.nixpkgs.follows = "nixpkgs";
};
```

### Build from source (Go)

- `git clone http://github.com/acaloiaro/btsw`
- Build:
  - `go build -o btsw .`
- Optionally install to your PATH:
  - `install -m 0755 btsw /usr/local/bin/`

## Usage

```
Usage:
  btsw [--adapter=hciX] <command> [args]

Commands:
  adapters                List all available Bluetooth adapters.
  devices                 List paired devices on the selected adapter (default: hci0).
  connect <id>            Connect to a paired device by its numeric ID (from 'devices').
  disconnect <id>         Disconnect from a paired device by its numeric ID (from 'devices').

Options:
  --adapter=hciX          Specify which Bluetooth adapter to use (default: hci0).
  --version               Print the version. 

Examples:
  btsw adapters               # List available bluetooth adapaters 
  btsw devices                # List available devices on the chosen bluetooth adapter 
  btsw --adapter=hci1 devices # List devics on hci1
  btsw connect 0              # Connect to device 0 (from 'devices' list)
  btsw disconnect 1           # Disconnect from device 1 (from 'devics' list)
```

Notes:
- The device index is stable per run, sorted by MAC address.
- Only paired devices are shown and targetable.

## Troubleshooting

- No adapters found:
  - Ensure bluetoothd is running (systemctl status bluetooth)
  - On NixOS: hardware.bluetooth.enable = true
- No paired devices:
  - Pair first using bluetoothctl (scan on, pair, trust)
- Permission denied / not authorized:
  - Some systems require Polkit authorization for connect/disconnect
  - Try from a logged-in user session; avoid sudo unless necessary

## Acknowledgments

- Uses [github.com/godbus/dbus](github.com/godbus/dbus) to speak to BlueZ on the system D‑Bus.
