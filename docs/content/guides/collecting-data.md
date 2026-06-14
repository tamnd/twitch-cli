---
title: "Collect data"
description: "Turn twitch commands into datasets: pipe to jq, build a CSV, or tee straight into a database."
weight: 30
---

`twitch` writes machine-readable output by default the moment you pipe it, so
turning a command into a dataset is mostly a matter of choosing a format and a
destination. This guide collects the patterns that come up most.

It assumes the [quick start](/getting-started/quick-start/). All commands
return data from any network: there is no IP soft-wall on Twitch's public
GraphQL, so datacenter and home connections behave the same.

## Pipe one record set into jq

Left to `auto`, `twitch` prints JSONL into a pipe: one JSON object per line,
which `jq` reads without any flags.

```bash
twitch channel videos shroud | jq -r '.title'
twitch game streams just-chatting | jq 'select(.viewers > 1000) | .channel'
```

Use `-o json` when a tool wants a single array instead of a stream:

```bash
twitch channel clips shroud -n 20 -o json | jq 'length'
```

Pull a single field directly:

```bash
twitch channel show shroud -o json | jq .followers
```

## Build a CSV or TSV

Pick the columns you care about, then ask for `csv` or `tsv`:

```bash
twitch streams -n 50 --fields channel,game,viewers -o csv > streams.csv
twitch game streams just-chatting -n 50 --fields channel,viewers -o tsv > justchatting.tsv
```

`--no-header` drops the header row when a downstream tool expects bare rows.

## Tee straight into a database

`--db` writes every emitted record into a store as a side effect of reading, so
a session fills a database with no separate import step. The record's `kit:"id"`
field is the key, so re-running a command updates rows in place rather than
duplicating them.

```bash
twitch channel show shroud --db twitch.db                 # a local SQLite file
twitch channel videos shroud --db twitch.db
twitch game streams just-chatting --db twitch.db
twitch streams -n 100 --db 'postgres://localhost/twitch'
```

Because the key is stable, you can layer several commands into one store and
query across them afterwards:

```bash
twitch channel show shroud --db twitch.db
twitch channel videos shroud -n 50 --db twitch.db
sqlite3 twitch.db '.tables'
sqlite3 twitch.db 'SELECT title, views FROM video ORDER BY views DESC LIMIT 10'
```

## Cap and pace a longer pull

`--limit` stops after N records. `--rate` spaces requests out so a longer pull
stays polite, and `--cache-ttl` lets a re-run reuse what you already fetched
instead of hitting the site again.

```bash
twitch channel videos shroud --limit 200 --rate 1s -o jsonl > videos.jsonl
twitch channel clips shroud --limit 200 --refresh           # ignore the cache, fetch fresh
```

A rate limit exits 5, so a script can tell "nothing came back" apart from
"the site asked me to slow down":

```bash
if ! twitch game streams just-chatting -n 100 -o jsonl > streams.jsonl; then
  echo "command failed (exit $?)" >&2
fi
```

## Format a quick report

A template turns each record into exactly the line you want, for a changelog
entry or a chat message:

```bash
twitch channel videos shroud \
  --template '{{.Title}}: {{.Views}} views ({{.LengthSeconds}}s)'

twitch game streams just-chatting -n 10 \
  --template '{{.Channel}} playing {{.Game}} ({{.Viewers}} viewers)'
```

See [output formats](/reference/output/) for the full contract and the field
list for every record type.
