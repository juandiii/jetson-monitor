
# Jetson Monitor

Jetson Monitor is a small, focused service to monitor HTTP endpoints on a schedule and send
notifications when endpoints become unavailable or recover. It aims to be simple to configure,
easy to run in containers, and extensible with notification providers.

Key features
- Scheduled HTTP checks using cron expressions (`cron`).
- Configurable retries and backoff to handle transient failures.
- Pluggable notification providers (Slack, Telegram) via an interface.
- Duplicate-alert suppression using a persisted state file.
- Minimal HTTP API for health and diagnostics implemented with `fiber`.

Quick start

1. Copy the example config and edit it:

```sh
cp config.sample.yml config.yml
# Edit config.yml and add your notification credentials and URLs to check
```

2. Run locally (development):

```sh
LOG_FORMAT=console LOG_LEVEL=DEBUG go run main.go
```

3. Build and run binary:

```sh
go build -o jetson-monitor ./...
./jetson-monitor
```

Configuration

The application reads configuration from `config.yml` into `config.ConfigJetson` (see `config/config.go`).
Important fields:

- `port` (int): HTTP server port (default: `38080`).
- `state_file` (string): path used to persist notification state. If empty, a file under the OS temp
	directory is used.
- `http_timeout_seconds` (int): base timeout for HTTP requests.
- `per_attempt_timeout_seconds` (int): timeout per retry attempt.
- `retry_max_tries` (int): number of retry attempts.
- `backoff_min_ms`, `backoff_max_ms`, `backoff_factor`, `backoff_jitter`: backoff configuration.
- `notification_cooldown_seconds` (int): time window used to suppress duplicate notifications
	for the same outage (default: 3600 seconds).
- `user_agent` (string): user agent used by the HTTP client.
- `notifications`: optional credentials for `telegram` and `slack` providers.
- `urls` (list): entries to check — each entry contains `url`, optional `status_code`, and `scheduler`.

Per-URL fields

- `url` (string) — target to perform the HTTP check on (required).
- `status_code` (int, optional) — if provided, the check requires the exact status code; otherwise any 2xx
	is considered healthy.
- `scheduler` (string) — cron expression. Examples: `@every 1m`, or a standard cron spec.

Example `config.yml`

```yaml
port: 38080
state_file: "" # empty -> use OS tmp dir
http_timeout_seconds: 30
per_attempt_timeout_seconds: 10
retry_max_tries: 3
backoff_min_ms: 500
backoff_max_ms: 10000
backoff_factor: 2.0
backoff_jitter: true
notification_cooldown_seconds: 3600

notifications:
	telegram:
		token: "<your-telegram-bot-token>"
		chat_id: 123456789
	slack:
		token: "<token-slack>"

urls:
	- url: https://example.com/health
	status_code: 200
	scheduler: "@every 1m"

	- url: https://another-service.local/ready
	status_code: 200
	scheduler: "0 */5 * * * *" # every 5 minutes
```

Environment variables

- `LOG_LEVEL` — logging level, e.g. `DEBUG`, `INFO`, `WARNING`, `ERROR`.
- `LOG_FORMAT` — `console` or `json`.
- Many config values can be overridden via `JETSON_*` environment variables (see `config.applyEnvOverrides`).

How it works

For each URL configured the app registers a cron job. The job performs an HTTP request (with retries and
backoff). If a check fails (network error or unexpected status), configured notification providers are
called and the checked URL is marked in the persisted state file so duplicate alerts are suppressed.
When the endpoint recovers, a recovery notification is sent and the state is cleared.

Running with Docker

Build the image:

```sh
make docker-build
```

Run an instance (example):

```sh
docker run -d --name jetson-monitor -p 38080:38080 \
	-v $(pwd)/config.yml:/app/config.yml:ro \
	-v $(pwd)/notified_state.json:/app/notified_state.json \
	-e LOG_LEVEL=debug \
	jetson-monitor:latest
```

Makefile targets

- `make build` — build local binary.
- `make docker-build` — build Docker image.
- `make test` — run tests.
- `make tidy` — run `go mod tidy`.

Extending notification providers

Add a package under `notification/` that implements the `notification.CommandProvider` interface and
wire its constructor into `scheduler.New` when you want it enabled.

Project layout highlights

- `main.go` — composition root and application bootstrap.
- `app/` — application lifecycle wrapper (cron + HTTP server orchestration).
- `config/` — configuration loading and environment overrides.
- `scheduler/` — job logic and persistence (state store).
- `notification/` — providers for Slack / Telegram.
- `sdk/` — HTTP SDK and helper functions used by the checks.

Testing

Run tests with:

```sh
make test
```

Contribution

Open issues or PRs. Keep changes focused and add tests for new behavior or providers.

License

See the `LICENSE` file in this repository.

