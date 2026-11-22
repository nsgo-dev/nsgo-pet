---
title: Get started
description: Install nsgo-pet on macOS or Linux and meet your first pet.
---

nsgo-pet puts a small animated pet on your desktop. It floats above your other windows, on every Space and over full-screen apps, stays off the Dock and the taskbar, and lets your clicks through everywhere but its body.

## Install

| System | Download |
| --- | --- |
| macOS | The disk image (`.dmg`) |
| Linux | The Debian package, the Arch package or the AppImage |

All of them are on the [releases page](https://github.com/nsgo-dev/nsgo-pet/releases).

On Linux, clips play through GStreamer's bad plugins (`gstreamer1.0-plugins-bad`). The Debian and Arch packages install them for you; with the AppImage, install them from your distribution.

:::note[No release yet]
Until the first release is out, the installers have to be built from source. With [Go](https://go.dev/dl/) 1.27.3 or later:

```sh
git clone https://github.com/nsgo-dev/nsgo-pet
cd nsgo-pet
go tool nsgo build
```

The installers are written to `dist/`.
:::

## Meet your pet

1. Open nsgo-pet. NSGo-chan greets you in the bottom right corner of the screen, and her icon appears in the menu bar or the tray.
2. Click her for a reaction, or drag her anywhere on the screen.
3. Leave her be: she idles, plays something new every little while and now and then walks along the bottom of the screen.

To switch pets, check for updates or quit, see [Using nsgo-pet](using/).

## Where it runs

| System | State |
| --- | --- |
| macOS | Runs. |
| Linux | Runs. On X11 the pet floats above other windows; on Wayland it can neither float nor move its own window. |
| Windows | Not yet: NSGo's Windows player cannot play clips with a transparent background. |
