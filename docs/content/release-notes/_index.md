---
title: "Release notes"
linkTitle: "Release notes"
description: "What changed in each twitch release, newest first."
weight: 40
---

What shipped in each release, newest first. Every tagged version builds the same
set of artifacts: archives for Linux, macOS, Windows, and FreeBSD, Linux
packages (deb, rpm, apk), a multi-arch container image on GHCR, and entries for
the package managers. Binaries are pure Go, so there is nothing to install
alongside them.

- **v0.2.0**: rebuilt on the [any-cli/kit](https://github.com/tamnd/any-cli)
  framework, so each operation is declared once and becomes a CLI command, an
  HTTP route under `serve`, an MCP tool, and a `twitch://` resource-URI
  dereference at the same time. Output gained color (`--color auto|always|never`)
  and a Markdown format. A field-coverage pass filled the stream `mature` flag
  and the channel `partner` and `affiliate` flags on search, which anonymous
  reads can reach but earlier builds left blank; a category's streams now show
  the category display name. A transient Twitch integrity check is now retried
  and, if it persists, surfaces as a rate limit (exit 5) rather than an error.

- **v0.1.0**: the first release, a standalone CLI. Read commands for streams,
  the category directory, channels (metadata, videos, clips, schedule), games
  (metadata, streams, clips), individual videos and clips, and search for both
  channels and games; the offline `ref` tools; and the shared output contract
  with `--fields`, `--template`, `--db`, and caching flags.
