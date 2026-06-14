# twitch

Browse live streams, top categories, and channels on [Twitch](https://www.twitch.tv/) from the command line.

`twitch` is a single pure-Go binary. No API key required.

## Install

```bash
go install github.com/tamnd/twitch-cli/cmd/twitch@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/twitch-cli/releases), or run the container image:

```bash
docker run --rm ghcr.io/tamnd/twitch:latest --help
```

## Usage

```bash
# Show top 20 live streams
twitch streams

# Show streams for a specific game
twitch streams --game "Minecraft"

# Show top categories/games
twitch categories

# Search for channels
twitch search ninja

# JSON output
twitch streams -o json -n 10

# Table output
twitch categories -o table
```

## Commands

| Command | Description |
|---------|-------------|
| `streams` | List top live streams |
| `categories` | List top categories/games by viewer count |
| `search <query>` | Search for channels by name |
| `version` | Show version information |

## Streams flags

```
-g, --game string   filter by game/category name
```

## Global flags

```
-o, --output string    output format: table|json|jsonl|csv|tsv|url|raw (default "auto")
-n, --limit int        limit number of records (default 20)
    --fields strings   comma-separated columns to include
    --no-header        omit header row
    --template string  Go text/template per record
    --timeout duration per-request timeout (default 30s)
    --delay duration   minimum spacing between requests
    --retries int      retry attempts on 429/5xx (default 3)
```

## License

Apache-2.0. See [LICENSE](LICENSE).
