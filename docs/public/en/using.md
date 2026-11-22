---
title: Using nsgo-pet
description: What your pet does on its own, and what its menu lets you do.
---

## Living with your pet

Your pet takes the pointer only on its body, so it never blocks the window behind it. Between your clicks it idles and plays an event now and then, chosen by:

- the time of day and the season, napping at night;
- mealtimes around 8:00, 12:00 and 18:00;
- festivals;
- whether an editor or a terminal is the app in front.

Click it for a reaction, drag it and it dangles from the pointer, and drop it wherever you like.

## The menu

Right-click the pet, or click the 🐾 icon in the menu bar or the tray. The menu's items are in Chinese:

| Item | What it does |
| --- | --- |
| 切换桌宠 | Switches to another pet; your choice is remembered next time |
| 检查更新… | Checks for a new version now |
| 来点动作 | Plays an event right away |
| 回到右下角 | Moves the pet back to the bottom right corner of the screen |
| 退出 | Says goodbye and quits |

Every pet and the license it comes under is listed in [Pet licenses](../pets/).

## Updates

nsgo-pet checks the [releases](https://github.com/nsgo-dev/nsgo-pet/releases) once a day and whenever you choose 检查更新…. On macOS an update window shows the new version; on Linux an alert offers to install it.

## Something's wrong?

| Problem | What to do |
| --- | --- |
| The pet sits behind other windows or stays put on Linux | Wayland doesn't let it float or move; log in to an X11 session. |
| The pet shows no animation on Linux | Install GStreamer's bad plugins, `gstreamer1.0-plugins-bad`. |
| The pet is out of sight | Choose 回到右下角 from the 🐾 icon. |

Anything else, [open an issue](https://github.com/nsgo-dev/nsgo-pet/issues).
