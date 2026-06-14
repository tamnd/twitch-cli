---
title: "CLI"
description: "Every command and subcommand, with the flags that matter."
weight: 10
---

```
twitch <command> [arguments] [flags]
```

Run `twitch <command> --help` for the full flag list on any command. A reference
is whatever you have: a bare login, a numeric video id, a clip slug, a full
`twitch.tv` URL, a `clips.twitch.tv` URL, or an `@handle`.

## Stream and category commands

| Command | What it does |
|---|---|
| `streams` | Top live streams right now, sorted by viewers |
| `games` | The category directory, sorted by viewers |

## Search commands

| Command | What it does |
|---|---|
| `search channels <query>` | Channels matching a query |
| `search games <query>` | Categories matching a query |

## Channel commands

| Command | What it does |
|---|---|
| `channel show <login>` | One channel's metadata (single record) |
| `channel videos <login>` | A channel's past videos (VODs) |
| `channel clips <login>` | A channel's clips |
| `channel schedule <login>` | A channel's upcoming schedule |

## Game (category) commands

| Command | What it does |
|---|---|
| `game show <slug>` | One category's metadata (single record). Use the slug, e.g. `just-chatting`, not the display name. |
| `game streams <slug>` | Live streams in a category |
| `game clips <slug>` | Top clips in a category |

## Single-item commands

| Command | What it does |
|---|---|
| `video <id>` | One video by numeric id (single record) |
| `clip <slug>` | One clip by slug (single record) |

## Offline reference tools

| Command | What it does |
|---|---|
| `ref id <ref>` | Classify any reference into its `(kind, id)`. No network. |
| `ref url <kind> <id>` | Build the canonical URL for a `(kind, id)`. No network. |

`ref id` accepts a bare login, a numeric video id, a clip slug, a full
`twitch.tv` URL, a `clips.twitch.tv` URL, or an `@handle`. The kinds it
returns are: `channel`, `game`, `video`, `clip`.

## Framework commands

| Command | What it does |
|---|---|
| `serve [--addr]` | Serve the operations over HTTP as NDJSON |
| `mcp` | Run as an MCP server over stdio |
| `completion` | Generate a shell completion script |
| `version` | Print the version and exit |

Under `serve`, each operation is a route, for example `GET /v1/streams`,
`GET /v1/channel/show/<login>`, and `GET /v1/game/streams/<slug>`.

## Global flags

These are shared by every operation, so they work the same on every command.

| Flag | Meaning |
|---|---|
| `-o, --output` | Output format: `auto`, `table`, `markdown`, `json`, `jsonl`, `csv`, `tsv`, `url`, `raw` |
| `--fields` | Comma-separated columns to keep |
| `--template` | Go text/template applied per record |
| `--no-header` | Omit the header row in `table` and `csv` |
| `-n, --limit` | Stop after N records (0 means no limit) |
| `--color` | `auto`, `always`, or `never` |
| `--rate` | Minimum delay between requests (default `500ms`) |
| `--retries` | Retry attempts on rate limit or 5xx |
| `--timeout` | Per-request timeout |
| `--client-id` | Override the public Twitch Client-Id header |
| `--user-agent` | Override the request User-Agent |
| `--cache-ttl` | How long cached responses stay fresh (default `24h`) |
| `--no-cache` | Bypass on-disk caches |
| `--refresh` | Fetch fresh and rewrite the cache |
| `--db` | Tee every record into a store (e.g. `out.db`, `postgres://...`) |
| `--profile` | Named profile for settings and data |
| `--data-dir` | Override the data directory |
| `-q, --quiet` | Suppress progress output |
| `-v, --verbose` | Increase verbosity (repeatable) |

See [output formats](/reference/output/) for what `-o`, `--fields`, and
`--template` produce, and [configuration](/reference/configuration/) for
environment variables, profiles, and cache behaviour.
