# Internet Monitor

A small self-hosted app for monitoring your public internet connection. It runs
periodic [Ookla Speedtests](https://www.speedtest.net/apps/cli) on a schedule,
stores the results in SQLite, and graphs them over time in a simple web
dashboard.

No authentication, single container, configured with environment variables —
built for a homelab.

![dashboard](https://img.shields.io/badge/stack-Go%20%2B%20Vue-blue)

## Features

- Scheduled Ookla speedtests (download, upload, ping, jitter, packet loss).
- Server, ISP, external IP and the shareable Speedtest.net result link captured
  for each run.
- Dashboard with summary cards, time-series charts (24h / 7d / 30d / all) and a
  results table. Failed runs are recorded and shown as gaps.
- On-demand "Run test now" button.
- Light/dark theme.

Planned: public IP monitoring (see [issues]/roadmap).

## Tech stack

- **Backend:** Go, standard-library `net/http`, pure-Go SQLite
  (`modernc.org/sqlite`, no CGO). Shells out to the official Ookla `speedtest` CLI.
- **Frontend:** Vue 3 + TypeScript + Vite + Tailwind CSS + DaisyUI, charts via
  Chart.js / vue-chartjs.
- **Deploy:** one Docker image (multi-arch amd64/arm64), `compose.yml`.

## Configuration

All configuration is via environment variables; all are optional.

| Variable | Default | Description |
|---|---|---|
| `SPEEDTEST_INTERVAL` | `1h` | Time between tests (Go duration, e.g. `30m`, `6h`). |
| `DATA_RETENTION` | *(unset)* | Prune results older than this (Go duration). Unset = keep forever. |

Fixed (not configurable): HTTP port `8080`, SQLite path (`/store` in the
container), per-test timeout `5m`, timezone UTC.
The first test runs one minute after startup, then on the interval. The Ookla
CLI selects the closest server automatically.

## Running with Docker

```sh
docker compose up -d --build
```

Then open <http://localhost:8080>. The SQLite database is persisted to `./store`.

## Development

Requires Go 1.25+, Node 20+/22+, and the official Ookla
[`speedtest` CLI](https://www.speedtest.net/apps/cli) on your `PATH`.

Backend:

```sh
go test ./...
go run .              # serves on :8080, first test after 1 minute
```

Frontend (with hot reload; proxies `/api` to the Go server on :8080):

```sh
cd frontend
npm install
npm run dev
```

For a production-like run, `npm run build` in `frontend/` produces
`frontend/dist`, which the Go server serves from disk.

## API

- `GET /api/results?from=<iso8601>&to=<iso8601>&limit=<n>` — results in range, newest first.
- `GET /api/results/latest` — most recent result (or `null`).
- `GET /api/stats?from=&to=` — min/max/avg per metric over the range.
- `POST /api/run` — run an on-demand test and return it.
