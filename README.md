# Internet Monitor

A small self-hosted app for monitoring your public internet connection. 

Runs periodic [Ookla Speedtests](https://www.speedtest.net/apps/cli) on a schedule, stores the results in SQLite, and graphs them over time in a simple web dashboard.

No authentication, single container, configured with environment variables — built for a homelab.

## Features

- Scheduled Ookla Speedtests (download, upload, ping, jitter, packet loss)
- Public IP and ISP IP Capture

## Tech stack

- **Backend:** Go, standard-library `net/http`, pure-Go SQLite (`modernc.org/sqlite`, no CGO). Shells out to the official Ookla `speedtest` CLI.
- **Frontend:** Vue 3 + TypeScript + Vite + Tailwind CSS, charts via Chart.js / vue-chartjs.
- **Deploy:** one Docker image (multi-arch amd64/arm64), `compose.yml`.

## Configuration

All configuration is via environment variables; all are optional.

| Variable | Default | Description |
|---|---|---|
| `SPEEDTEST_INTERVAL` | `1h` | Time between tests (Go duration, e.g. `30m`, `6h`). |
| `DATA_RETENTION` | *(unset)* | Prune results older than this (Go duration). Unset = keep forever. |

Fixed (not configurable): HTTP port `8080`, SQLite path (`/store` in the container), per-test timeout `5m`, timezone UTC.
The first test runs one minute after startup, then on the interval. 
The Ookla CLI selects the closest server automatically.

## Running with Docker

```sh
docker compose up -d --build
```

Then open <http://localhost:8080>. The SQLite database is persisted to `./store`.
