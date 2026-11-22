---
title: Introduction
description: Install nsgo-pet on macOS or Linux and get to know your pet.
sidebar:
  order: 1
---

nsgo-pet is a desktop pet built with [NSGo](https://github.com/nsgo-dev/nsgo). It floats in a transparent window above your other windows, on every Space and over full-screen apps, and stays off the Dock and the taskbar.

## Install

Download the disk image for macOS, or the Debian package, Arch package or AppImage for Linux, from the [releases](https://github.com/nsgo-dev/nsgo-pet/releases).

Until the first release, build them yourself. With Go 1.27.3 or later, from a clone of the repository:

```sh
git clone https://github.com/nsgo-dev/nsgo-pet
cd nsgo-pet
go tool nsgo build
```

The installers are written to `dist/`.

Linux needs GStreamer's bad plugins (`gstreamer1.0-plugins-bad`) to play clips with alpha; the Debian and Arch packages depend on them.

## Meet your pet

1. Open nsgo-pet. NSGo-chan greets you in the bottom right corner of the screen.
2. Drag her anywhere, or click her for a reaction.
3. Right-click her, or use the 🐾 tray icon, to switch pets, play an event, move her back to the corner, check for updates or quit.

The pet takes the pointer only on its body, so clicks elsewhere reach the window below. Between your clicks it idles and plays events chosen by time of day, season, meals, festivals and whether an editor or a terminal is in front, and now and then it walks along the bottom of the screen.

Your choice of pet is remembered, and nsgo-pet updates itself from the repository's releases, checking daily and whenever you ask from its menu.

## Platforms

| Platform | State |
| --- | --- |
| macOS | Runs, playing the HEVC-with-alpha clips. |
| Linux | Plays the WebM clips. The window floats on X11; on Wayland it cannot float or move itself. |
| Windows | Not yet: NSGo's Windows player refuses clips with alpha. |

Lunar festivals, hit testing by the clip's alpha ([nsgo#599](https://github.com/nsgo-dev/nsgo/issues/599)), speech bubbles and weather are not done yet.
