# twitch

Read public Twitch streams, channels, clips, and categories into structured records.

`twitch` is a single pure-Go binary. It reads public Twitch the way a logged-out
browser does: the top live streams, the category directory, a channel with its
videos, clips, and schedule, a single video or clip, and search over channels
and categories. It talks to Twitch's public GraphQL API with the web client's
public id, so there is no API key, no login, and nothing to run alongside it.

The same package is also a [resource-URI driver](#use-it-as-a-resource-uri-driver),
so a host program like [ant](https://github.com/tamnd/ant) can address Twitch as
`twitch://` URIs.

`twitch` is an independent tool and is not affiliated with Twitch.

## Install

```bash
go install github.com/tamnd/twitch-cli/cmd/twitch@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/twitch-cli/releases), or run
the container image:

```bash
docker run --rm ghcr.io/tamnd/twitch:latest --help
```

## Usage

```bash
twitch streams                      # top live streams right now
twitch games                        # the category directory, by viewers
twitch search channels <query>      # channels matching a query
twitch search games <query>         # categories matching a query

twitch channel show <login>         # a channel's metadata
twitch channel videos <login>       # a channel's past videos
twitch channel clips <login>        # a channel's clips
twitch channel schedule <login>     # a channel's upcoming schedule

twitch game show <slug>             # a category's metadata
twitch game streams <slug>          # live streams in a category
twitch game clips <slug>            # top clips in a category

twitch video <id>                   # one video by id
twitch clip <slug>                  # one clip by slug

twitch ref id <ref>                 # classify any reference into its (kind, id)
twitch ref url <kind> <id>          # build the canonical URL for a (kind, id)

twitch --help                       # the whole command tree
```

A reference is whatever you have: a bare login, a full `twitch.tv` URL, a
`clips.twitch.tv` link, an `@handle`, or a numeric video id. The `ref` commands
resolve these offline, with no network call.

Every command shares one output contract: `-o table|json|jsonl|csv|tsv|url|raw`,
`--fields` to pick columns, `--template` for a custom line, and `--limit` to cap
results. The default adapts to where output goes (a table on a terminal, JSONL
in a pipe), so the same command reads well by hand and parses cleanly
downstream.

```bash
twitch channel show shroud -o json | jq .followers
twitch game streams just-chatting --limit 5 --fields channel,viewers
```

## What anonymous access reaches

`twitch` reads only what Twitch serves to a logged-out browser, and Twitch
serves a lot. Every command above returns data with nothing but the web
client's public id and a browser user-agent, from a home network or a datacenter
alike. There is no IP soft-wall to work around.

That access has a clear edge: it does not reach anything that needs an account.
No following, no chat, no subscriptions, no watch history, no per-viewer state.
Records carry only fields anonymous access can fill, so there is no
always-empty column. Twitch removed total channel view counts from the public
API, so a channel record does not carry one.

When something is genuinely missing the exit code says which: a command that
finds nothing exits 3 (no results), a rate limit exits 5, a withheld surface
exits 4 (need auth), and an unknown login or deleted video exits 6 (not found).
A script can tell those apart.

## Serve it

The same operations are available over HTTP and as an MCP tool set for agents,
with no extra code:

```bash
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

```bash
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

```bash
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

```bash
git tag v0.1.0
git push --tags
```

The Homebrew and Scoop steps self-disable until their tokens exist, so the first
release works with no extra secrets.

## License

Apache-2.0. See [LICENSE](LICENSE).
