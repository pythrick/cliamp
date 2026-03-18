# Telemetry

cliamp sends a single anonymous ping once per calendar month so we can count monthly active users.

## What is collected

- A randomly generated UUID (created on first launch, stored locally)
- The app version string

That's it. No IP logging, no usage data, no personal information.

## How it works

1. On first launch, a random UUID is generated and saved to `~/.config/cliamp/.telemetry_id`
2. A one-time startup notice explains telemetry and how to disable it
3. Each launch checks if a ping has already been sent this month
4. If not, a single background `POST` request is sent to `https://telemetry.cliamp.stream/ping`
5. The request is fire-and-forget — it never blocks the app or surfaces errors

The JSON payload is:

```json
{"uuid":"<random-id>","version":"<cliamp-version>"}
```

## Disable telemetry

- Persistent: set `telemetry = false` in `~/.config/cliamp/config.toml`
- One-off session: run `cliamp --no-telemetry`

## Storage

The telemetry state file is located at:

```
~/.config/cliamp/.telemetry_id
```

It contains the UUID and the last ping month in JSON format.
