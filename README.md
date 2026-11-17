# nsgo-pet

A desktop pet built with [NSGo](https://github.com/nsgo-dev/nsgo), whose pet can be switched from the tray icon or its right-click menu.

The pet lives in a transparent, borderless window that floats above other windows on every Space, stays off the Dock and the taskbar, and lets clicks through everywhere except its body. It greets you, idles, plays an event every so often chosen by its clips' weights (time of day, season, meals, Gregorian festivals, coding in the app in front), reacts to clicks, dangles while dragged, and walks along the bottom of the screen.

## Run

```sh
tools/fetch-bigbluefish.sh   # Big Blue Fish's clips (about 150 MB)
go tool nsgo dev
go tool nsgo test .
```

## Pets

Each pet is a directory under `resources/pets/<id>/`:

| File | What it is |
| --- | --- |
| `pet.json` | Its name, window size, body rectangle in percent of the frame, event interval and the clips it plays by name (greeting, exit, drag, idle, daily) |
| `animations.json` | Its clips and when they play, in [Big Blue Fish's format](https://github.com/Bai-Cute/big-blue-fish-desktop-pet): pools `Idle`, `Event`, `Click`, `Greeting`, `Daily`, `Drag`, `Disabled`, with weights and their conditions |
| `webm/` | The clips as VP9 with alpha, for Linux |
| `mov/` | The same clips as HEVC with alpha, for macOS, whose player opens no WebM |

The clips are not in git; each pet has a script in `tools/` that fetches them.

| Pet | Source | License |
| --- | --- | --- |
| 大蓝鱼 (`bigbluefish`) | [Big Blue Fish](https://github.com/Bai-Cute/big-blue-fish-desktop-pet) | MIT, PC2005-cloud (`resources/pets/bigbluefish/LICENSE.txt`) |

## Platforms

- macOS: runs.
- Linux: plays the WebM clips; needs GStreamer's `gstreamer1.0-plugins-bad` (codecalpha).
- Windows: not yet. NSGo's Windows player refuses clips with alpha, so the clips need the app's own decoder feeding `ImageSource.Pixels`.

## Not yet

Lunar festivals (their clips keep their plain weight), speech bubbles, and taking the pointer by the clip's alpha instead of a body rectangle ([nsgo#599](https://github.com/nsgo-dev/nsgo/issues/599)).
