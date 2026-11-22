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

## The menu and the settings

Click NSGo-chan's icon in the menu bar or the tray. Its menu has 设置 (Command-comma), which opens the settings, and 退出 (Command-Q), which says goodbye and quits.

The settings have two pages, in Chinese:

| Page | What it does |
| --- | --- |
| 宠物 | Shows every pet; click one to switch to it, and your choice is remembered next time |
| 通用 | Checks for a new version now, and turns the anonymous usage statistics on or off |

Every pet and the license it comes under is listed in [Pet licenses](../pets/).

## Updates

nsgo-pet checks the [releases](https://github.com/nsgo-dev/nsgo-pet/releases) once a day and whenever you press 检查 on the settings' 通用 page. On macOS an update window shows the new version; on Linux an alert offers to install it.

## Something's wrong?

| Problem | What to do |
| --- | --- |
| The pet sits behind other windows or stays put on Linux | Wayland doesn't let it float or move; log in to an X11 session. |
| The pet shows no animation on Linux | Install GStreamer's bad plugins, `gstreamer1.0-plugins-bad`. |
| The pet is out of sight | Choose a pet in 设置, or restart nsgo-pet: a pet starts in the bottom right corner. |

Anything else, [open an issue](https://github.com/nsgo-dev/nsgo-pet/issues).
