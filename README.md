# Zipp Nest

```
  ,~~~~~,
 (~~~~~~~)
  `~~~~~`
```

Server-side companion for [Zipp](https://github.com/Inspiractus01/zipp). Run it on your remote machine and send backups to it over the network — no cloud, no subscription, just your own server.

## Install

```bash
curl -sL https://raw.githubusercontent.com/Inspiractus01/zipp-nest/main/install.sh | bash
```

## Usage

```bash
zipp-nest serve
```

On first run it generates an auth token and starts listening on port `9090`.

```
  )()(
 ( ●● )  zipp-nest v0.1.0
  \──/
  /||\

  token:    a3f9c2e1b4d87650...
  port:     9090
  storage:  ~/.zipp-nest/backups

  ─────────────────────────────────────
  time      job                  event
  ─────────────────────────────────────
  14:05:01  photos               ↑ 2006-01-02_14-05-01.tar.gz  (24.3 MB)
```

## API

All endpoints require `Authorization: Bearer <token>` header.

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Server status |
| `POST` | `/backups/:job` | Upload a snapshot (tar.gz body) |
| `GET` | `/backups/:job` | List snapshots for a job |
| `GET` | `/backups/` | List all jobs |

## Config

Stored at `~/.zipp-nest/config.json`:

```json
{
  "token": "a3f9c2e1b4d87650...",
  "port": 9090,
  "storagePath": "/home/user/.zipp-nest/backups"
}
```

## Connecting with Zipp

Use [Tailscale](https://tailscale.com) (free) to access your nest server from anywhere without port forwarding. Once Tailscale is set up on both machines, your server is reachable at its Tailscale hostname.

Support for connecting directly from the Zipp client is coming soon.
