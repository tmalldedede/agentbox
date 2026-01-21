# Repository Guidelines

## Project Structure & Module Organization
- `cmd/agentbox/` holds the main entrypoint; build output goes to `bin/agentbox`.
- `internal/` contains backend modules (API, agent adapters, container/session/profile/skill/mcp/credential/config).
- `web/` hosts the React + Vite admin UI.
- `docs/` contains architecture notes and generated OpenAPI/Swagger assets.
- `docker/` includes Dockerfiles for building agent images.
- `tests/integration/` holds Hurl-based API integration tests.

## Build, Test, and Development Commands
Backend (Go):
```
make build          # build ./cmd/agentbox into bin/
make run            # build and run backend (default :18080)
make dev            # hot reload via air
make test           # go test -v ./...
make test-coverage  # generates coverage.html
make lint           # golangci-lint run
make docker         # build Docker image
```
Frontend (from `web/`):
```
npm install
npm run dev         # Vite dev server
npm run build       # typecheck + build
npm run lint        # ESLint
npm run format      # Prettier write
```

## Coding Style & Naming Conventions
- Go code should be `gofmt`-formatted; keep packages and files aligned with existing domains (e.g., `internal/api/*_handler.go`).
- Use `*_test.go` for Go tests and table-driven patterns where appropriate.
- Frontend code is formatted by Prettier; lint with ESLint. Keep React component folders under `web/src` and follow existing naming and routing patterns.

## Testing Guidelines
- Unit tests: `go test -v ./...` or scoped (e.g., `go test -v ./internal/api/...`).
- Integration tests: `hurl --test tests/integration/api.hurl --variable host=http://localhost:18080` (backend must be running).
- There is no dedicated frontend test runner configured yet; ensure `npm run lint` and `npm run build` pass for UI changes.

## Commit & Pull Request Guidelines
- Commit messages follow a Conventional Commits style: `feat:`, `fix:`, `docs:`, `refactor:` (see recent history).
- PRs should explain the change, link related issues, and include screenshots for UI updates.
- Keep PRs focused; update or add tests when behavior changes.

## Security & Configuration Tips
- Do not commit secrets; use environment variables or local config files for credentials.
- Docker is required for running agent containers; document any new images or runtime dependencies.
