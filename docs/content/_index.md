---
title: "twitch"
description: "Read public Twitch streams, channels, games, videos, and clips into structured records"
heroTitle: "twitch, from the command line"
heroLead: "Read public Twitch streams, channels, games, videos, and clips into structured records. One pure-Go binary, no API key, output that pipes into the rest of your tools, and a resource-URI driver other programs can address."
heroPrimaryURL: "/getting-started/quick-start/"
heroPrimaryText: "Get started"
---

`twitch` reads public Twitch the way a logged-out browser does, shapes the
responses into clean records, and gets out of your way.

```bash
twitch streams                      # top live streams right now
twitch channel show <login>         # a channel's metadata
twitch game streams <slug>          # live streams in a category
twitch search channels <query>      # channels matching a query
twitch serve --addr :7777           # the same operations over HTTP
```

There is no API key, nothing to sign up for, and nothing to run alongside it.
Output adapts to where it goes: an aligned table on your terminal, JSONL the
moment you pipe it somewhere.

`twitch` is an independent tool and is not affiliated with Twitch.

## Two ways to use it

- **As a command** for reading Twitch by hand or in a script. Start with
  the [quick start](/getting-started/quick-start/).
- **As a resource-URI driver** so a host like
  [ant](https://github.com/tamnd/ant) can address Twitch as
  `twitch://` URIs and follow links across sites. See
  [resource URIs](/guides/resource-uris/).

Both are the same code: one operation, declared once, is a CLI command, an HTTP
route, an MCP tool, and a URI dereference.

## A note on what anonymous access reaches

`twitch` reads only what Twitch serves to a logged-out browser. Twitch's public
GraphQL endpoint is genuinely open anonymously: every command returns data from
any network, home or datacenter. There is no IP soft-wall here.

The real boundary is the account line. Anonymous access cannot reach anything
that needs a login: no following, chat, subscriptions, watch history, or per-viewer
state. Records carry only fields anonymous access can fill, so there are no
always-empty columns.

## Where to go next

- New here? Read the [introduction](/getting-started/introduction/), then the
  [quick start](/getting-started/quick-start/).
- Installing? See [installation](/getting-started/installation/).
- Doing a specific job? The [guides](/guides/) are task-first.
- Need every flag? The [CLI reference](/reference/cli/) is the full surface.
