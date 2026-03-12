# zipp-nest

```
  ,~~~~~,
 (~~~~~~~)
  `~~~~~`
```

Self-hosted backup server for [zipp](https://github.com/Inspiractus01/zipp). · **[zipp.rest](https://zipp.rest)** Run it on any machine — a VPS, Raspberry Pi, NAS, home server. Backups arrive encrypted over your private network, no cloud involved.

## Install

```bash
curl -sL https://raw.githubusercontent.com/Inspiractus01/zipp-nest/main/install.sh | bash
```

macOS and Linux · amd64 and arm64

## What it does

- Receives backups from zipp over the network
- Stores snapshots per job, auto-prunes old ones
- Built-in TUI to start/stop the server, view logs, edit settings
- Connects via Tailscale — no open ports, no port forwarding needed

## Usage

```
zipp-nest         open the TUI
zipp-nest serve   run the server directly (no TUI)
```

Open the TUI, press **Start server** — it registers as a background service (launchd or systemd) and keeps running after you close it.

## Setup

1. Install zipp-nest on your server, press **Start server**
2. Go to **Connection info** — copy the short code
3. Open zipp on your client, go to **Nest**, paste the code
4. Press `n` on any job to set it to `[nest]` or `[nest+local]`

Config: `~/.zipp-nest/config.json`
