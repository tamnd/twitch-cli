---
title: "Introduction"
description: "What twitch is and how it is put together."
weight: 10
---

`twitch` reads public Twitch streams, channels, games, videos, and clips into
structured records.

It is a single binary. It reads Twitch the way a logged-out browser does,
shapes the responses into clean records, and gets out of your way. There is no
API key, nothing to sign up for, and nothing to run alongside it. `twitch` is an
independent tool and is not affiliated with Twitch.

## How it is built

- A **library package** (`twitch`) holds the GraphQL client and the typed data
  models. It paces requests, sets an honest User-Agent, retries the transient
  failures any public site throws under load, and paginates cursor connections.
  Every query goes to Twitch's public GraphQL endpoint using the web client's
  public Client-Id, the same id a logged-out browser sends.
- A **domain** (`twitch/domain.go`) declares each operation once on the
  [any-cli/kit](https://github.com/tamnd/any-cli) framework. That single
  declaration becomes a CLI command, an HTTP route, an MCP tool, and a
  resource-URI dereference.
- A thin **`cmd/twitch`** hands the assembled app to `kit.Run`, which builds the
  command tree and the serve and mcp surfaces.

## One operation, four surfaces

Because an operation is surface-neutral, the same `channel show` you run on the
command line is also a route and a tool:

```bash
twitch channel show <login>                  # the command
twitch serve --addr :7777                    # GET /v1/channel/show/<login>
twitch mcp                                   # the channel show tool, over stdio
ant get twitch://channel/<login>             # the URI dereference (via a host)
```

## What it reads

The read commands cover streams, channels, games, videos, clips, schedules,
and search:

```bash
twitch streams                  twitch game show <slug>
twitch games                    twitch game streams <slug>
twitch search channels <query>  twitch game clips <slug>
twitch search games <query>     twitch video <id>
twitch channel show <login>     twitch clip <slug>
twitch channel videos <login>
twitch channel clips <login>
twitch channel schedule <login>
```

Two offline tools, `twitch ref id` and `twitch ref url`, classify and build
references without touching the network.

## Scope

`twitch` is a read-only client over data Twitch already serves publicly. That
narrow scope keeps it a single small binary with no database, no daemon, and no
setup.

Next: [install it](/getting-started/installation/), then take the
[quick start](/getting-started/quick-start/).
