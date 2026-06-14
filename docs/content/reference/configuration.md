---
title: "Configuration"
description: "Environment variables, defaults, profiles, the cache, and the data directory."
weight: 20
---

`twitch` needs almost no configuration: it runs anonymously against public data
out of the box. The settings below let you tune politeness, caching, and storage.

## Defaults

| Setting | Default | Flag |
|---|---|---|
| Request pacing | 500ms between requests | `--rate` |
| Retries | retried on 429/5xx | `--retries` |
| Per-request timeout | 30s | `--timeout` |
| On-disk cache | enabled, fresh for 24h | `--cache-ttl`, `--no-cache`, `--refresh` |

## Caching

Responses are cached on disk under the data directory. A cached entry is reused
until it is older than `--cache-ttl` (default `24h`). To control it:

```bash
twitch channel show shroud --cache-ttl 1h    # treat anything older than an hour as stale
twitch channel show shroud --refresh         # fetch fresh and rewrite the cache
twitch channel show shroud --no-cache        # ignore the cache entirely, do not write it
```

## The data directory

Caches and any record store live under one data directory, chosen in this order:

1. `--data-dir`
2. `TWITCH_DATA_DIR`
3. `$XDG_DATA_HOME/twitch`
4. `~/.local/share/twitch`

## Profiles

`--profile <name>` keeps a separate set of settings and data under the data
directory, so you can switch between, for example, a fast local run and a slow
polite one without re-passing flags each time.

## Environment variables

Every flag has an environment fallback, prefixed `TWITCH_` in upper case with
dashes as underscores. For example:

```bash
export TWITCH_RATE=1s            # same as --rate 1s
export TWITCH_CACHE_TTL=1h       # same as --cache-ttl 1h
export TWITCH_DATA_DIR=~/data/twitch
export TWITCH_CLIENT_ID=...      # override the public Client-Id
```

Flags win over environment variables, which win over the built-in defaults.

## Sending records to a store

`--db` tees every emitted record into a store as a side effect of reading, so a
session fills a local database without a separate import step:

```bash
twitch channel videos shroud --db out.db        # SQLite file
twitch game streams just-chatting --db 'postgres://...'
```
