# twitch

A fast, friendly command line for [Twitch](https://www.twitch.tv). One binary
that lists the top live streams, walks the category directory, opens a channel
with its videos, clips, and schedule, resolves a single video or clip, searches
channels and categories, and optionally tees everything into a local SQLite
store. No API key, no login, nothing to run alongside it.

```
twitch streams -n 6 --fields channel,game,viewers
```

```
╭────────────────┬────────────────┬─────────╮
│ CHANNEL        │ GAME           │ VIEWERS │
├────────────────┼────────────────┼─────────┤
│ eslcs          │ Counter-Strike │ 56023   │
│ strogo         │ Counter-Strike │ 73622   │
│ ohnepixel      │ Counter-Strike │ 53686   │
│ gofns          │ VALORANT       │ 41289   │
│ ow_esports     │ Overwatch      │ 40945   │
│ cs2_paragon_ru │ Counter-Strike │ 38960   │
╰────────────────┴────────────────┴─────────╯
```

On a terminal the table header and JSON values are colorized; piped to a file or
another program the output drops to plain text so it parses cleanly. Use
`--color always` to keep color through a pipe, or `--color never` to drop it.

Full documentation: [twitch-cli.tamnd.com](https://twitch-cli.tamnd.com).

## Why

Reading Twitch programmatically usually means registering an application, holding
an OAuth token, and learning the Helix API, all to see things a logged-out
browser shows for free. `twitch` talks to the same public GraphQL endpoint the
website itself uses, with the web client's public id, so there is no key to
register and no token to refresh. It puts the public surface (streams, the
category directory, channels, videos, clips, schedules, and search) behind one
tool with real output formats and pipelines that compose.

## Install

```sh
go install github.com/tamnd/twitch-cli/cmd/twitch@latest
```

Or grab a prebuilt binary from the [releases page](https://github.com/tamnd/twitch-cli/releases).
The binary is pure Go with no runtime dependencies. You can also run the
container image:

```sh
docker run --rm ghcr.io/tamnd/twitch:latest --help
```

Build from source:

```sh
git clone https://github.com/tamnd/twitch-cli
cd twitch-cli
make build      # produces ./bin/twitch
```

## Quick start

```sh
twitch streams                        # top live streams right now
twitch games                          # the category directory, by viewers
twitch channel show shroud            # a channel's metadata
twitch channel videos shroud          # a channel's past videos
twitch game streams just-chatting     # live streams in a category
twitch search channels ninja          # channels matching a query
twitch clip <slug>                    # one clip by slug
twitch video <id>                     # one video by id
```

Most commands accept a bare login, a numeric id, a clip slug, an `@handle`, or a
full Twitch URL wherever they take a reference. The `ref` commands resolve those
offline, with no network call:

```sh
twitch ref id https://clips.twitch.tv/AwesomeClip -o json
```

```json
[
  {
    "input": "https://clips.twitch.tv/AwesomeClip",
    "kind": "clip",
    "id": "AwesomeClip",
    "url": "https://clips.twitch.tv/AwesomeClip"
  }
]
```

## How it works

Twitch renders its site from one backend, a public GraphQL endpoint at
`gql.twitch.tv/gql`. It answers a logged-out reader using the web client's
public Client-Id, a well-known id sent as a header, not an account, paired with
a browser user-agent. `twitch` sends full GraphQL query strings, walks the Relay
cursor connections to honor `--limit`, paces and caches requests, and maps each
reply onto a clean record. No API key, no token.

## What anonymous access reaches

`twitch` reads only what Twitch serves to a logged-out browser, and Twitch
serves a lot. Every command above returns data with nothing but the web
client's public id and a browser user-agent, from a home network or a datacenter
alike. There is no IP soft-wall to work around. Occasionally the endpoint
answers a request with an integrity check; running the command again clears it.

That access has a clear edge: it does not reach anything that needs an account.
No following, no chat, no subscriptions, no watch history, no per-viewer state.
Records carry only fields anonymous access can fill, so there is no
always-empty column. Twitch removed total channel view counts from the public
API, so a channel record does not carry one.

When something is genuinely missing the exit code says which, so a script can
tell the cases apart:

| Exit | Meaning |
| --- | --- |
| 0 | ok |
| 2 | usage error |
| 3 | no results (the resource is genuinely empty) |
| 4 | need auth, or a withheld surface |
| 5 | rate limited (raise `--rate`) |
| 6 | not found (unknown login, deleted video or clip, bad slug) |
| 8 | network error |

## Commands

| Command | What it does |
| --- | --- |
| `streams` | Top live streams right now |
| `games` | The category directory, by viewers |
| `search channels <query>` | Channels matching a query |
| `search games <query>` | Categories matching a query |
| `channel show <login>` | A channel's metadata |
| `channel videos <login>` | A channel's past videos |
| `channel clips <login>` | A channel's clips |
| `channel schedule <login>` | A channel's upcoming schedule |
| `game show <slug>` | A category's metadata |
| `game streams <slug>` | Live streams in a category |
| `game clips <slug>` | Top clips in a category |
| `video <id>` | One video by id |
| `clip <slug>` | One clip by slug |
| `ref id <ref>` | Classify a reference into its (kind, id), offline |
| `ref url <kind> <id>` | Build the canonical URL for a (kind, id), offline |
| `serve` | Serve the same operations over HTTP as NDJSON |
| `mcp` | Serve the same operations to an agent over MCP |
| `version` | Print version, commit, and build date |

A category is addressed by its slug, like `just-chatting`, not its display name.
Run `twitch <command> --help` for the full flag list on any command.

## Output

Every command shares one output contract. The default adapts to where output
goes, a table on a terminal and JSONL in a pipe, so the same command reads well
by hand and parses cleanly downstream.

```sh
twitch games -n 6 --fields name,slug,viewers -o table
```

```
╭────────────────────┬────────────────────┬─────────╮
│ NAME               │ SLUG               │ VIEWERS │
├────────────────────┼────────────────────┼─────────┤
│ Just Chatting      │ just-chatting      │ 433775  │
│ Counter-Strike     │ counter-strike     │ 433112  │
│ VALORANT           │ valorant           │ 197936  │
│ League of Legends  │ league-of-legends  │ 113310  │
│ Overwatch          │ overwatch-2        │ 102851  │
│ Grand Theft Auto V │ grand-theft-auto-v │ 94081   │
╰────────────────────┴────────────────────┴─────────╯
```

Pick the format with `-o table|json|jsonl|csv|tsv|url|raw`, choose columns with
`--fields a,b,c`, render a custom line with `--template`, drop the header with
`--no-header`, and cap results with `-n/--limit`. The `url` format prints just
the canonical URL of each record, which is handy for piping into another tool.

## Recipes

A channel as JSON, piped to jq:

```sh
twitch channel show ninja -o json | jq '{login, followers, partner}'
```

```json
{
  "login": "ninja",
  "followers": 19259019,
  "partner": true
}
```

Every live stream in a category as JSONL, for a downstream job:

```sh
twitch game streams just-chatting -n 100 -o jsonl > just-chatting.jsonl
```

A channel's recent videos with just the columns you want:

```sh
twitch channel videos shroud -n 4 --fields title,game,views
```

The canonical URLs of the top streams, one per line:

```sh
twitch streams -n 20 -o url
```

Tee a search into a local SQLite store, keyed by each record's id, then query it:

```sh
twitch search channels speedrun -n 200 --db twitch.db
```

## Serve it

The same operations are available over HTTP and as an MCP tool set for agents,
with no extra code:

```sh
twitch serve --addr :7777    # GET /v1/... returns NDJSON
twitch mcp                   # speak MCP over stdio
```

## Use it as a resource-URI driver

`twitch` registers a `twitch` domain the way a program registers a database
driver with `database/sql`. A host enables it with one blank import:

```go
import _ "github.com/tamnd/twitch-cli/twitch"
```

Then [ant](https://github.com/tamnd/ant) (or any program that links the package)
dereferences `twitch://` URIs without knowing anything about Twitch:

```sh
ant get twitch://channel/<login>      # fetch a channel
ant get twitch://video/<id>           # fetch a video
ant get twitch://clip/<slug>          # fetch a clip
ant url twitch://game/<slug>          # the live https URL
```

## Development

```
cmd/twitch/  thin main: hands cli.NewApp to kit.Run
cli/         assembles the kit App from the twitch domain
twitch/      the library: GraphQL client, queries, data models, and
             domain.go (the driver)
docs/        tago documentation site
```

```sh
make build      # ./bin/twitch
make test       # go test ./...
make vet        # go vet ./...
```

Every read command is declared once as a kit operation in `twitch/domain.go`.
That single declaration becomes the CLI subcommand, the HTTP route, and the MCP
tool, so the three surfaces never drift.

## Releasing

Push a version tag and GitHub Actions runs GoReleaser, which builds the archives,
Linux packages, the multi-arch GHCR image, checksums, SBOMs, and a cosign
signature:

```sh
git tag v0.1.0
git push --tags
```

The Homebrew and Scoop steps self-disable until their tokens exist, so the first
release works with no extra secrets.

## License

`twitch` is an independent tool and is not affiliated with Twitch. Apache-2.0,
see [LICENSE](LICENSE).
