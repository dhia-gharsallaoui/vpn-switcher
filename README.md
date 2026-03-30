# vpn-switcher

A terminal UI for managing multiple VPN connections with policy-based routing. Switch between OpenVPN and Tailscale connections, configure routing rules, and run network diagnostics — all from one place.

## Features

- **VPN Management** — Discover, connect, and switch between OpenVPN (systemd/NetworkManager) and Tailscale connections
- **Policy Routing** — Define per-subnet routing rules to send traffic through specific VPN interfaces
- **Network Diagnostics** — Built-in ping, traceroute, DNS lookup, and nslookup tools
- **CLI Mode** — List VPNs and switch connections without opening the TUI
- **Multi-VPN** — Optionally run multiple VPNs simultaneously with policy routing

## Install

### From source

```sh
make build
sudo make install
```

### With Go

```sh
go install github.com/dhia-gharsallaoui/vpn-switcher/cmd/vpn-switcher@latest
```

## Usage

### TUI

```sh
sudo vpn-switcher
```

Navigate between tabs with `Tab`/`Shift+Tab`. Press `?` for keybindings help.

### CLI

```sh
# List all discovered VPNs
vpn-switcher list

# Switch to a VPN by name
vpn-switcher switch "my-vpn"
```

### Flags

```
--debug     Enable debug logging
--version   Show version
--help      Show help
```

## Configuration

Copy the example config and edit it:

```sh
mkdir -p ~/.config/vpn-switcher
cp config.example.yaml ~/.config/vpn-switcher/config.yaml
```

Key settings:

```yaml
general:
  openvpn_config_dirs:
    - "/etc/openvpn/client"
    - "~/.config/openvpn"
  openvpn_method: "auto"  # "systemd", "nmcli", or "auto"
  allow_multi_vpn: false

# Override display names or add VPN profiles
vpn_profiles:
  - name: "Work VPN"
    config_path: "/etc/openvpn/client/work.conf"
    provider: "openvpn"

# Policy routing rules
routing_rules:
  - id: "corp-network"
    dest_cidr: "10.0.0.0/8"
    vpn_interface: "tun0"
    table: "100"
```

See [`config.example.yaml`](config.example.yaml) for all options.

## Requirements

- Linux
- One or more of: OpenVPN, Tailscale
- Root privileges (or polkit rules for passwordless systemd unit management)

## License

MIT
