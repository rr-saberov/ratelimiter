# ratelimiter

[![CI](https://github.com/rr-saberov/ratelimiter/actions/workflows/ci.yml/badge.svg)](https://github.com/rr-saberov/ratelimiter/actions/workflows/ci.yml)
[![CD](https://github.com/rr-saberov/ratelimiter/actions/workflows/cd.yml/badge.svg)](https://github.com/rr-saberov/ratelimiter/actions/workflows/cd.yml)
[![Go](https://img.shields.io/github/go-mod/go-version/rr-saberov/ratelimiter)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A small Go HTTP service that limits requests per client IP using a token bucket.
Built to demonstrate CI/CD with GitHub Actions. No third-party runtime dependencies.

## Quick start

With Go 1.26 or newer, run from the repository root:

```sh
go run ./cmd/ratelimiter
```

Or use Docker:

```sh
docker run --rm -p 8080:8080 ghcr.io/rr-saberov/ratelimiter:latest
```

Then, in another terminal:

```sh
for i in 1 2 3 4 5 6; do
  curl -s -o /dev/null -w '%{http_code}\n' localhost:8080/
done
```

With a fresh bucket, six requests within one second return five `200` responses
followed by `429 Too Many Requests`. Each client starts with five tokens and regains
one per second, up to the capacity. Rejected requests include a `Retry-After` header.

## Configuration

```sh
go run ./cmd/ratelimiter -addr :8080 -capacity 20 -refill 250ms
```

| Flag | Default | Description |
| --- | --- | --- |
| `-addr` | `:8080` | Listen address |
| `-capacity` | `5` | Maximum burst per client |
| `-refill` | `1s` | Time to replenish one token |
| `-version` | `false` | Print the build version and exit |

The endpoint is `GET /`. The server shuts down gracefully on `SIGINT` or `SIGTERM`.

## Development

```sh
go build ./...
go test -race -count=1 -timeout=5m ./...
```

[Task](https://taskfile.dev) shortcuts are defined in [Taskfile.yml](Taskfile.yml):

```sh
task run             # Build and start the server
task lint            # Requires golangci-lint v2.13.2
task lint:workflows   # Validate GitHub Actions workflows
task coverage        # Run tests and report coverage
task docker:smoke    # Build and test the Docker image
task --list          # Show all commands
```

Code: [token bucket library](ratelimit/), [HTTP handler](internal/httpapi/),
[server entry point](cmd/ratelimiter/).

## CI/CD

| Workflow | Trigger | Result |
| --- | --- | --- |
| [CI](.github/workflows/ci.yml) | Pushes to `main`, PRs targeting `main`, or manual runs | Build, lint, tests with race detection and coverage, Docker smoke test |
| [CD](.github/workflows/cd.yml) | GitHub Release published | Lint and test, attach binaries and SHA-256 checksums, publish Docker image to GHCR |

Releases include binaries for Linux and macOS (amd64/arm64), and Windows (amd64).
Docker images support Linux amd64/arm64; `latest` tracks stable releases in publication order.

To release, publish a GitHub Release with a new tag such as `v1.2.3`.
CD then builds and uploads the artifacts. This flow requires immutable releases to be disabled.

See the [CI/CD reference](docs/ci-cd.md) for release commands, manual checks,
branch protection, action pinning, and troubleshooting.

## Limitations

This is a demo: client buckets stay in memory without eviction and are not shared
between instances. Clients are identified by the TCP peer IP; `X-Forwarded-For` is ignored.

## License

[MIT](LICENSE)
