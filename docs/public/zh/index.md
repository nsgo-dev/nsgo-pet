---
title: 入门
description: 在 macOS 或 Linux 上安装 nsgo-pet，认识你的第一只宠物。
---

nsgo-pet 会在你的桌面上放一只会动的小宠物。它浮在其他窗口之上，出现在每个桌面空间和全屏应用上，不占程序坞和任务栏；除了它的身体，点在哪里都会落到下面的窗口。

## 安装

| 系统 | 下载 |
| --- | --- |
| macOS | 磁盘映像（`.dmg`） |
| Linux | Debian 包、Arch 包或 AppImage |

它们都在 [Releases 页面](https://github.com/nsgo-dev/nsgo-pet/releases)。

在 Linux 上，动画通过 GStreamer 的 bad 插件（`gstreamer1.0-plugins-bad`）播放。Debian 包和 Arch 包会替你装好；用 AppImage 时，请从你的发行版安装它们。

## 认识你的宠物

1. 打开 nsgo-pet。NSGo 酱会在屏幕右下角向你打招呼，菜单栏或托盘里会出现她的图标。
2. 点她看看反应，或者把她拖到屏幕上任意位置。
3. 不管她也行：她会待机，每隔一会儿做点新动作，还会不时沿着屏幕底部走动。

切换宠物、检查更新或退出，见 [使用 nsgo-pet](using/)。

## 支持的系统

| 系统 | 状态 |
| --- | --- |
| macOS | 可用。 |
| Linux | 可用。在 X11 上宠物浮于其他窗口之上；在 Wayland 上既无法浮于顶层，也无法移动自己的窗口。 |
| Windows | 暂不支持：NSGo 的 Windows 播放器还不能播放透明背景的动画。 |
