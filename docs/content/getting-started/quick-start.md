---
title: "Quick start"
description: "Run your first twitch command."
weight: 30
---

Once `twitch` is on your `PATH`:

```bash
twitch --help       # see the command tree
twitch version      # build info
```

This is a fresh scaffold, so the command tree is just `version` for now. Add
your first real command in `cli/`, build on the `twitch` library package,
and document it here.

A good first command usually fetches one thing and prints it as JSON, so the
output pipes straight into `jq` and the rest of your tools.
