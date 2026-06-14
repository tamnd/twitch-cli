---
title: "Output formats"
description: "The output contract every command shares: formats, fields, and templates."
weight: 30
---

Every command renders through one formatter, so the same flags work everywhere.
Pick a format with `-o`, or let `twitch` choose: a table when writing to a
terminal, JSONL when piped.

## Formats

```bash
twitch <command> -o table      # aligned columns for reading
twitch <command> -o markdown   # a Markdown table, for docs and issues
twitch <command> -o jsonl      # one JSON object per line, for piping
twitch <command> -o json       # a single JSON array
twitch <command> -o csv        # spreadsheet friendly
twitch <command> -o tsv        # tab-separated
twitch <command> -o url        # just the URL column
twitch <command> -o raw        # the underlying bytes, unformatted
```

| Format | Best for |
|---|---|
| `table` | Reading on a terminal |
| `markdown` | Pasting into docs, issues, or pull requests |
| `jsonl` | Piping into another tool, one object at a time |
| `json` | Loading a whole result as an array |
| `csv` / `tsv` | Spreadsheets and quick column math |
| `url` | Feeding URLs into other commands |
| `raw` | The unformatted bytes (response bodies) |

## Record fields

The JSON field names are the same across formats. They are the columns `--fields`
selects and the keys a template reads (capitalised). Each record type has its own
set:

| Record | Fields |
|---|---|
| Stream | `id`, `channel`, `display_name`, `title`, `game`, `game_slug`, `viewers`, `language`, `mature`, `started_at`, `url` |
| Channel | `id`, `login`, `display_name`, `description`, `followers`, `partner`, `affiliate`, `live`, `game`, `viewers`, `created`, `avatar`, `url` |
| Video | `id`, `title`, `channel`, `display_name`, `game`, `views`, `length_seconds`, `published_at`, `url` |
| Clip | `slug`, `title`, `channel`, `curator`, `game`, `views`, `duration_seconds`, `created_at`, `url` |
| Game | `id`, `name`, `slug`, `display_name`, `viewers`, `url` |
| Segment | `id`, `title`, `category`, `start_at`, `end_at`, `canceled` |
| Ref | `input`, `kind`, `id`, `url` |

## Narrowing columns

Keep only the fields you want:

```bash
twitch streams --fields channel,game,viewers
twitch channel videos shroud --fields id,title,views,length_seconds
twitch channel clips shroud --fields slug,title,views,duration_seconds
```

`--no-header` drops the header row in `table` and `csv` output, which helps when
a downstream tool expects bare rows.

## Templating rows

For full control over each line, apply a Go text/template. Fields are the JSON
keys, capitalised:

```bash
twitch game streams just-chatting --template '{{.Channel}} ({{.Viewers}} viewers)'
twitch channel videos shroud --template '{{.Title}} - {{.Views}} views'
```

## Why auto-detection helps

Because the default adapts to the destination, the same command reads well by
hand and parses cleanly in a pipe:

```bash
twitch streams -n 5                    # a table, because this is a terminal
twitch streams -n 5 | wc -l           # JSONL, because this is a pipe
```

You only reach for `-o` when you want something other than that default.
