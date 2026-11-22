---
title: Pet packs
description: The folder of clips and rules that makes a pet, with no code.
group: Developers
sidebar:
  order: 4
---

A pack is a folder of clips and the rules that choose them, so a new pet needs no code. The engine reads packs; it ships none. nsgo-pet's packs are in [`resources/pets/<id>/`](https://github.com/nsgo-dev/nsgo-pet/tree/main/resources/pets), and a host's live wherever its `fs.FS` or directory points.

## Layout

| Path | What it is |
| --- | --- |
| `pet.json` | The pet's name, size, body and the clips it plays by name |
| `animations.json` | Every clip and when it plays, in [Big Blue Fish's format](https://github.com/Bai-Cute/big-blue-fish-desktop-pet/blob/75cd217568e713a7de8812f65dad7072798a8b70/code/BigBlueFish.Windows/assets/fish/animations.json) |
| `webm/<clip>.webm` | Each clip as VP9 with alpha, for Linux and Windows |
| `mov/<clip>.mov` | The same clip as HEVC with alpha, for macOS, whose player opens no WebM |
| `NOTICE.md` | Where the pet comes from, its authors and its license |

Both clip forms exist because no single format with alpha plays on every platform's player. The engine picks the form by OS and never transcodes, so a pack's clips ship in both forms or not at all.

## pet.json

```json
{
  "name": "大蓝鱼",
  "width": 320,
  "height": 180,
  "body": { "left": 33, "top": 15, "width": 34, "height": 80 },
  "eventSeconds": 40,
  "clips": {
    "greeting": "女仆屈膝礼仪",
    "exit": "女仆屈膝礼仪",
    "drag": "被鼠标拖拽悬空反馈",
    "idle": "待机呼吸休闲",
    "daily": "晨间刷牙",
    "dailyHour": 23
  },
  "source": "https://github.com/PC2005-cloud/dsh-pet",
  "license": "Character CC BY-NC-SA 4.0 (non-commercial); see NOTICE.md"
}
```

`body` is a rectangle in percent of the frame. Only there does the pet take the pointer, so clicks elsewhere in a transparent window reach the window below. `clips` names the clips the pet plays outside the pools: on start, on leaving, while dragged, as the fallback idle, and once a day at `dailyHour`. A manifest naming a clip that `animations.json` lacks fails to load.

## animations.json

| Field | Meaning |
| --- | --- |
| `Pool` | `Idle` between events, `Event` every `eventSeconds`, `Click`, `Greeting`, `Daily`, `Drag`, or `Disabled` |
| `Weight` | How likely the clip is against the others of its pool; 0 never plays unless a condition raises it |
| `NightOnly`, `SeasonOnly`, `SummerWeight`, `WinterWeight`, `MorningWeight` | Time of day and season |
| `Meal`, `MealBoost` | A boost within an hour of 8:00, 12:00 or 18:00 |
| `Festival`, `RadiusDays`, `FestivalMode`, `FestivalWeight` | Weight near a Gregorian festival; lunar festivals keep their plain weight |
| `WorkKind`, `WorkWeight` | `Coding` or `Working`: what the host's `Context` says the user is doing |
| `MoveKind`, `HourParity` | `Run` or `Crab`: the pet walks 70 points and back, in a window of its own only |

`Disabled` clips never play by chance, but a host can still play them by ID, which is how a host maps its own states onto clips.

## Making clips

Clips need a transparent background. From PNG frames with alpha:

```sh
ffmpeg -framerate 24 -i frames/%04d.png -c:v libvpx-vp9 -pix_fmt yuva420p -b:v 0 -crf 32 webm/idle.webm
ffmpeg -c:v libvpx-vp9 -i webm/idle.webm -an -c:v hevc_videotoolbox -alpha_quality 0.75 -q:v 60 -tag:v hvc1 -pix_fmt bgra mov/idle.mov
```

The converters in [`tools/`](https://github.com/nsgo-dev/nsgo-pet/tree/main/tools) write a whole pack, both clip forms and both JSON files. Their flags are in each one's `-h`.

| Tool | Input |
| --- | --- |
| `tools/codexpet` | A Codex sprite sheet of eight 192×208 rows, with optional seasonal outfits |
| `tools/framepet` | PNG frames and a JSON spec under `tools/framepet/pets/` |
| `tools/fetch-*.sh` | One source's pets, fetched again from a pinned commit and converted |

## Licensing

Each pack carries its own license in `NOTICE.md` and `pet.json`'s `license`. The engine's MIT license covers no pack. A host that ships packs ships their licenses with them, and a pack under a non-commercial license, such as Big Blue Fish's, makes the build that includes it non-commercial.
