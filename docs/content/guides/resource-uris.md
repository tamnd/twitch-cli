---
title: "Resource URIs"
description: "Use twitch as a database/sql-style driver so a host program can address Twitch as twitch:// URIs."
weight: 20
---

`twitch` is a command line, but the `twitch` Go package is also a small driver
that makes Twitch addressable as a resource URI. A host program registers it
the way a program registers a database driver with `database/sql`, then
dereferences `twitch://` URIs without knowing anything about how Twitch is
fetched.

The host that does this today is [ant](https://github.com/tamnd/ant), a single
binary that puts one URI namespace over a family of site tools. The examples
below use `ant`; any program that links the package gets the same behaviour.

## Mounting the driver

A host enables the driver with one blank import, exactly like `import _
"github.com/lib/pq"`:

```go
import _ "github.com/tamnd/twitch-cli/twitch"
```

The package's `init` registers a domain with the scheme `twitch` (alias `ttv`)
for the hosts `twitch.tv`, `www.twitch.tv`, `clips.twitch.tv`, and
`m.twitch.tv`. The standalone `twitch` binary does not change.

## Addressing records

A URI is `scheme://authority/id`. The `twitch` driver exposes these
authorities:

| URI                          | What it is                               |
| ---------------------------- | ---------------------------------------- |
| `twitch://channel/<login>`   | one channel, keyed by its login          |
| `twitch://video/<id>`        | one video, keyed by its numeric id       |
| `twitch://clip/<slug>`       | one clip, keyed by its slug              |
| `twitch://game/<slug>`       | one category, keyed by its slug          |

```bash
ant get twitch://channel/shroud              # the channel record
ant get twitch://video/<id>                  # the video record
ant get twitch://clip/<slug>                 # the clip record
ant url twitch://game/just-chatting          # the live https URL
ant resolve https://www.twitch.tv/shroud     # a pasted link, back to its URI
```

The same classification runs offline through the binary: `twitch ref id <ref>`
turns any reference into its `(kind, id)`, and `twitch ref url <kind> <id>`
builds the canonical URL.

## Walking the graph

`ls` lists the members of a collection, and every member is itself an
addressable URI, so a host can follow the graph and write it to disk:

```bash
ant ls     twitch://channel/shroud            # videos and clips on this channel
ant export twitch://game/just-chatting --follow 1 --to ./data
```

When records carry edges through `kit:"link"` tags, `ant export --follow` and
`ant graph` walk those edges too, across tools when a link points at another
site's scheme.

## Why this is the same code

The driver and the binary share one definition per operation. A resolver op
answers both `twitch channel show` on the command line and `ant get
twitch://channel/...` through a host, from the same handler and the same client.
There is no second implementation to keep in step.
