---
title: "Add a command"
description: "Model a Twitch record and expose it as a command, a route, and a tool at once."
weight: 10
---

Each read command in `twitch` is one operation, declared once, that becomes a
CLI subcommand, an HTTP route, an MCP tool, and a URI dereference. You add to
the tool in two files, and every surface updates itself.

## 1. Model the record

The records live in `twitch/types.go`. Each exported struct carries the json
tags a reader sees and the `kit` tags that decide how a host addresses it. The
`Clip` record, for example:

```go
type Clip struct {
    Slug            string  `json:"slug" kit:"id"`          // the URI id
    Title           string  `json:"title,omitempty"`
    Channel         string  `json:"channel,omitempty"`
    Curator         string  `json:"curator,omitempty"`
    Game            string  `json:"game,omitempty"`
    Views           int64   `json:"views,omitempty"`
    DurationSeconds float64 `json:"duration_seconds,omitempty"`
    CreatedAt       string  `json:"created_at,omitempty"`
    URL             string  `json:"url"`
}
```

- `kit:"id"` marks the field that becomes the URI id.
- `table:",truncate"` on a field keeps wide text from blowing up a terminal
  table.
- `kit:"link,kind=<scheme>/<type>"` on an edge field lets a host walk from one
  record to another, across tools when the link points at another site.

The client methods that fill these records live next to the queries they serve,
one file per surface; the mapping from Twitch's wire JSON to a record happens in
that same file, for example `toClip` in `clips.go`.

## 2. Declare the operation

In `twitch/domain.go`, add an input struct and a handler, then register it in
`Register`:

```go
type clipRef struct {
    Slug   string  `kit:"arg" help:"clip slug or URL"`
    Client *Client `kit:"inject"`
}

func getClip(ctx context.Context, in clipRef, emit func(*Clip) error) error {
    c, err := in.Client.GetClip(ctx, in.Slug)
    if err != nil {
        return mapErr(err)
    }
    return emit(c)
}

// inside Register(app):
kit.Handle(app, kit.OpMeta{Name: "clip", Group: "read", Single: true,
    Summary: "Show one clip by slug", URIType: "clip", Resolver: true,
    Args: []kit.Arg{{Name: "slug", Help: "clip slug or clips.twitch.tv URL"}}}, getClip)
```

That is the whole change. `kit.Handle` reflects the input for flags and the
output for the record shape, so the operation immediately becomes:

```bash
twitch clip <slug>                     # the command
curl 'localhost:7777/v1/clip/<slug>'   # the route, under serve
ant get twitch://clip/<slug>           # the URI dereference, via a host
```

## Resolver ops and list ops

Two flags shape how a host treats an operation:

- **`Single: true`** with **`Resolver: true`** marks the canonical one-record
  fetch for a `URIType`. It answers `ant get`. In `twitch` these are `channel
  show`, `video`, `clip`, and `game show`.
- **`List: true`** marks a member-lister for a parent resource. It answers
  `ant ls`. A list op emits records that are themselves addressable, so every
  member is a URI a host can follow. `channel videos`, `channel clips`,
  `game streams`, and the search feeds work this way.

## Map errors to exit codes

Return the `errs` kinds from `mapErr` so every surface reports the same outcome
with the same exit code:

```go
case errors.Is(err, ErrNotFound):
    return errs.NotFound("%s", err.Error())
case errors.Is(err, ErrRateLimited):
    return errs.RateLimited("%s", err.Error())
```

When a channel has no clips or a schedule is empty, the handler emits nothing
and the surface exits with no results (exit 3) rather than fabricating.

See [output formats](/reference/output/) for how records render, and
[resource URIs](/guides/resource-uris/) for how a host addresses them.
