---
title: "Troubleshooting"
description: "The handful of things that trip people up, and how to fix each one."
weight: 40
---

Most of these come down to how Twitch serves its data or a mismatched input, not
a bug in `twitch`.

## A command returns no results (exit 3)

Because Twitch's public GraphQL is genuinely open anonymously, exit 3 means the
resource is actually empty. There is no IP soft-wall here, so the network you
run from does not affect which commands return data. If `channel clips shroud`
exits 3, that channel really has no clips at this moment. If `channel schedule
shroud` exits 3, no segments are scheduled yet.

## `game show` returns "Gql: service error"

`game show` takes a slug, not a display name. Use the slug, for example
`just-chatting`, not "Just Chatting". The slug is the last segment of the
Twitch URL for that category, and `twitch ref id` can classify it:

```bash
twitch game show just-chatting          # correct
twitch ref id https://www.twitch.tv/directory/game/Just%20Chatting -o json
```

## Requests start failing or returning 429

`twitch` already paces requests (500ms between them by default) and retries the
transient failures, but a hard rate limit still means backing off. Raise the
delay with `--rate` (for example `--rate 1s`) and retry later. A rate limit
surfaces as exit 5.

```bash
twitch channel videos shroud --rate 1s -n 200
```

## Nothing is found for something you expected

Check that the reference is spelled the way Twitch uses it: a bare login, a
numeric video id, a clip slug, a full `twitch.tv` URL, or an `@handle`.
`twitch ref id <ref>` shows how `twitch` classifies what you gave it, with no
network call. A genuinely missing channel, video, or clip surfaces as exit 6
(not found).

```bash
twitch ref id https://www.twitch.tv/shroud -o json
twitch ref id https://clips.twitch.tv/AbCdEf -o json
```

## Anonymous access limits

Anonymous access cannot reach anything that needs a login: following, chat,
subscriptions, watch history, or per-viewer state. If a command exits 4
(need-auth), that surface requires a logged-in session and is outside the scope
of `twitch`.

## The binary is not on your PATH

`go install` puts the binary in `$(go env GOPATH)/bin` (usually `~/go/bin`), and
a release archive leaves it wherever you unpacked it. If your shell cannot find
`twitch`, add that directory to your `PATH`. See
[installation](/getting-started/installation/).

## Seeing what twitch actually did

When something behaves unexpectedly, `-v` adds per-request detail so you can see
the URLs it hit and the responses it got. That is usually enough to tell a rate
limit apart from a genuinely empty result. `--refresh` rules out a stale cache
entry.

## Exit codes

Every surface reports the same outcome with the same code:

| Code | Meaning |
|---|---|
| 0 | OK |
| 2 | Usage error |
| 3 | No results (the resource is genuinely empty) |
| 4 | Need authentication (surface requires a login) |
| 5 | Rate limited |
| 6 | Not found (unknown login, deleted video or clip, bad slug) |
| 8 | Network error |
