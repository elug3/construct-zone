# Construct-Zone

A tiny Go web server that renders an "under construction" placeholder page using the standard-library `html/template` package. The HTML template (`templates/index.html`) is embedded into the binary via `go:embed`, so no template files need to be shipped alongside the binary.

## Layout
- `main.go` — HTTP server, page data, visitor counter, uptime/domain logic.
- `templates/index.html` — the embedded `html/template` page.

## Cursor Cloud specific instructions

- Go (1.22.x) is preinstalled. The app uses only the standard library, so there are no third-party modules to fetch.
- Run the server in development with `go run .` from the repo root. It listens on `:8080` by default; override with the `PORT` env var (e.g. `PORT=3000 go run .`).
- Endpoints: `/` serves the page (and increments the in-memory visitor counter on each non-asset request), `/healthz` returns `ok`.
- The visitor count is in-memory only and resets whenever the process restarts — this is expected, not a bug.
- The "Domain" shown on the page is taken from the request `Host` header, so it reflects whatever host you hit (e.g. `localhost:8080`).
- Lint/vet/build: `gofmt -l .` (should print nothing), `go vet ./...`, `go build ./...`.
- The ASCII gopher in `main.go` is a Go raw string literal, so it must not contain backtick characters.
