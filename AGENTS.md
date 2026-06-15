# Construct-Zone

A tiny Go web server that renders an "under construction" placeholder page using the standard-library `html/template` package. The HTML template (`templates/index.html`) is embedded into the binary via `go:embed`, so no template files need to be shipped alongside the binary.

## Layout
- `main.go` — HTTP server, page data, session-cookie visitor counter, uptime/address logic.
- `templates/index.html` — the embedded `html/template` page.

## Cursor Cloud specific instructions

- Go (1.22.x) is preinstalled. The app uses only the standard library, so there are no third-party modules to fetch.
- Run the server in development with `go run .` from the repo root. It listens on `:8080` by default; override with the `PORT` env var (e.g. `PORT=3000 go run .`).
- Endpoints: `/` serves the page, `/healthz` returns `ok`.
- Visitor counting is unique-per-browser via the `cz_session` cookie: the in-memory counter only increments for requests with no session cookie (new visitors), so reloads by the same browser do NOT inflate the count. Testing this in a browser requires cookies to persist across reloads — use a normal/persistent profile, not a per-request fresh profile (a headless Chrome run with a stable `--user-data-dir` reproduces the "count stays constant on reload" behavior).
- The count is in-memory only and resets whenever the process restarts — this is expected, not a bug.
- The page is served with `Cache-Control: no-store` because it is dynamic (live counters, per-request stats); browser hard-reloads aren't needed to see template changes after a server restart.
- The "Address" shown on the page is taken from the request `Host` header, so it reflects whatever host you hit (e.g. `localhost:8080`).
- View mode (Desktop/Mobile badge + layout) is automatic and driven entirely by the CSS `@media (max-width: 600px)` query — there is no JS toggle. The "Uptime" and "Server Time (UTC)" values tick every second via a small client-side script.
- "Uptime" is the OS/system uptime read from `/proc/uptime` (Linux), not the process runtime; it falls back to the process start time if `/proc/uptime` is unavailable. The client ticks it using the derived boot time, so on a long-lived host the value will already be large at startup.
- Lint/vet/build: `gofmt -l .` (should print nothing), `go vet ./...`, `go build ./...`.
- The ASCII gopher in `main.go` is a Go raw string literal, so it must not contain backtick characters.
