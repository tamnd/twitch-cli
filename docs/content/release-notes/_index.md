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

- **initial**: the first release. Read commands for streams, the category
  directory, channels (metadata, videos, clips, schedule), games (metadata,
  streams, clips), individual videos and clips, and search for both channels
  and games; the offline `ref` tools; the shared output contract with
  `--fields`, `--template`, `--db`, and caching flags; the `serve` and `mcp`
  surfaces; and the `twitch://` resource-URI driver.
