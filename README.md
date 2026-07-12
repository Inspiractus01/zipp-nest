# zipp-nest

```
  ,~~~~~,
 (~~~~~~~)
  `~~~~~`
```

Self-hosted backup server for [zipp](https://github.com/Inspiractus01/zipp). · **[zipp.rest](https://zipp.rest)** Run it on any machine — a VPS, Raspberry Pi, NAS, home server. Backups arrive already encrypted (age, client-side) over your private network, no cloud involved.

## Install

```bash
curl -sL https://raw.githubusercontent.com/Inspiractus01/zipp-nest/main/install.sh | bash
```

macOS and Linux · amd64 and arm64 · release binaries are checksum-verified on install

## What it does

- Receives backups from zipp over the network — streamed to disk, so even huge backups don't need RAM
- Requires a bearer token (part of the connection code); unauthenticated requests are rejected
- Stores snapshots per job as opaque encrypted blobs, auto-prunes old ones
- Built-in TUI to start/stop the server, view logs, edit settings
- Connects via Tailscale — no open ports, no port forwarding needed

## Usage

```
zipp-nest         open the TUI
zipp-nest serve   run the server directly (no TUI)
zipp-nest update  update to the latest version
```

Open the TUI, press **Start server** — it registers as a background service (launchd or systemd) and keeps running after you close it.

## Setup

1. Install zipp-nest on your server, press **Start server**
2. Go to **Connection info** — copy the connection code (address words + auth token)
3. Open zipp on your client, go to **Nest**, paste the code
4. In a job's detail view set **Backup mode** to `nest` or `both`

Config: `~/.zipp-nest/config.json` (contains the auth token — keep it private)
