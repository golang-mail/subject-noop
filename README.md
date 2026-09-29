# subject-noop

Minimal example go-mail filter plugin (socket-mode). It logs the subject of every email it sees and returns no actions — useful as a copy-paste template and to verify the plugin protocol end-to-end without touching mail.

## Install

```bash
go install github.com/golang-mail/subject-noop@latest
```

The binary lands in `$GOBIN` (`$GOPATH/bin` by default); make sure that directory is on your `PATH`.

## Register the plugin

Add an entry to `plugins.yml` (in go-mail's config directory, alongside `accounts.yml` / `filters.yml`). The default transport is `socket`, so no `transport:` field is needed:

```yaml
plugins:
  - name: subject-noop
    command: ["/path/to/subject-noop"]
    timeout: 5s
```

`name` is how filters reference the plugin. `command` is exec-style argv; no shell. go-mail allocates a Unix socket path under `$XDG_RUNTIME_DIR/go-mail/plugins/subject-noop.sock` and passes it to the child via `GO_MAIL_PLUGIN_SOCKET`; the plugin listens there and go-mail dials.

## Use it from a filter

In `filters.yml`, reference the plugin via the `plugin:<name>` condition field:

```yaml
filters:
  - name: log all subjects
    conditions:
      - field: plugin:subject-noop
        operator: matches
        value: ""
    actions: []
```

Each evaluated email goes through the plugin, the subject is logged via go-mail's slog (the plugin writes to stderr; go-mail forwards stderr to slog at INFO with `plugin=subject-noop`), and the plugin returns no match — so no actions run.

## Verify it's wired up

`GET /api/plugins` returns the plugin's status. Manual debug:

```bash
socat - UNIX-CONNECT:$XDG_RUNTIME_DIR/go-mail/plugins/subject-noop.sock
```

Paste an NDJSON `init` request and watch the response come back.

## Protocol cheat sheet

The plugin listens on the Unix socket whose path is in `GO_MAIL_PLUGIN_SOCKET`. Each accepted connection carries NDJSON: one JSON object per line, requests in, responses out. Stderr is free-form log output.

- `init` request → respond with manifest (`name`, `version`, `actions`, `wants`, `cacheable`).
- `evaluate` request → respond with `{}` for no-op, or `{"matched": true, "actions": [...]}` to act.
- `shutdown` request → exit cleanly.

`wants` controls which fields go-mail serializes per evaluate call — `envelope`, `headers`, `body`. This plugin only wants `envelope`, so the body is never sent over the wire.

See the [plugin protocol](https://github.com/golang-mail/go-mail/blob/main/plugin/README.md) for the full protocol.
