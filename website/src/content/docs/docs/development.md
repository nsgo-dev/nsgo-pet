---
title: Development
description: Run nsgo-pet from source, test it and release it.
group: Developers
sidebar:
  order: 5
---

nsgo-pet is two things in one repository: the pet engine, a Go module any NSGo app can import, and the nsgo-pet app, one host of that engine with its own pets, tray and updater. Read [ARCHITECTURE.md](https://github.com/nsgo-dev/nsgo-pet/blob/main/ARCHITECTURE.md) before changing code.

## Run and test

Needs Go 1.27.3 or later. From the repository root:

```sh
go tool nsgo dev           # build and open the pet
go tool nsgo test .        # the app's headless tests, which check every pet in resources/pets
(cd pet && go test ./...)  # the engine's tests
```

Both test suites run headless and load no media. Real playback, transparency, floating, walking at a screen's edge and Linux's X11 and Wayland differences need a manual run with `go tool nsgo dev`.

## Release

Move `CHANGELOG.md`'s `Unreleased` bullets under `## [<version>] - <date>`, set `version` in `nsgo.yaml` to the same version and push the tag `v<version>`. [The release workflow](https://github.com/nsgo-dev/nsgo-pet/blob/main/.github/workflows/release.yml) builds, signs and publishes the installers with that section as the notes.

Changes are listed in [CHANGELOG.md](https://github.com/nsgo-dev/nsgo-pet/blob/main/CHANGELOG.md).
