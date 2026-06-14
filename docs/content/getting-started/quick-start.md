---
title: "Quick start"
description: "Fetch your first record with twitch."
weight: 30
---

Once `twitch` is on your `PATH`, look up a channel. A reference is whatever you
have: a bare login, a full `twitch.tv` URL, or an `@handle`.

```bash
twitch channel show shroud
```

By default you get an aligned table. Ask for JSON when you want to pipe it:

```bash
$ twitch channel show shroud -o json
[
  {
    "id": "...",
    "login": "shroud",
    "display_name": "shroud",
    "description": "...",
    "followers": 9000000,
    "partner": true,
    "affiliate": false,
    "live": false,
    "url": "https://www.twitch.tv/shroud"
  }
]
```

See the top live streams, or filter by category:

```bash
twitch streams -n 5
twitch game streams just-chatting -n 5
```

Search works for both channels and categories:

```bash
twitch search channels ninja -n 5 --fields login,followers
twitch search games minecraft -n 5 --fields name,slug,viewers
```

## Shape the output

The same flags work on every command:

```bash
twitch streams -n 5 --fields channel,game,viewers
twitch channel show shroud --fields login,followers,live
twitch channel videos shroud -n 5 -o jsonl | jq .title
```

`-o` takes `auto`, `table`, `json`, `jsonl`, `csv`, `tsv`, `url`, or `raw`. Left
to `auto`, it prints a table to a terminal and JSONL into a pipe, so the same
command reads well by hand and parses cleanly downstream. See
[output formats](/reference/output/) for the full contract.

## Resolve a reference offline

The `ref` commands classify and build references with no network call:

```bash
twitch ref id https://clips.twitch.tv/AbCdEf -o json   # -> {kind: clip, id: AbCdEf}
twitch ref url channel shroud                           # the canonical channel URL
```

## Serve it instead

The same operations are available over HTTP and to agents over MCP:

```bash
twitch serve --addr :7777 &
curl -s 'localhost:7777/v1/channel/show/shroud'   # NDJSON
twitch mcp                                         # MCP over stdio
```
