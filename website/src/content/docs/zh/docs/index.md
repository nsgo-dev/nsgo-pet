---
title: 入门
description: 在 macOS 或 Linux 上安装 nsgo-pet，认识你的宠物。
sidebar:
  order: 1
---

nsgo-pet 是用 [NSGo](https://github.com/nsgo-dev/nsgo) 构建的桌面宠物。它在透明窗口里浮于其他窗口之上，出现在每个桌面空间和全屏应用上，不占程序坞和任务栏。

## 安装

从 [Releases](https://github.com/nsgo-dev/nsgo-pet/releases) 下载 macOS 的磁盘映像，或 Linux 的 Debian 包、Arch 包或 AppImage。

首个版本发布之前，请自行构建。需要 Go 1.27.3 或更高版本，在克隆的存储库里运行：

```sh
git clone https://github.com/nsgo-dev/nsgo-pet
cd nsgo-pet
go tool nsgo build
```

安装包会写到 `dist/`。

Linux 需要 GStreamer 的 bad 插件（`gstreamer1.0-plugins-bad`）才能播放带透明通道的动画；Debian 包和 Arch 包已依赖它们。

## 认识你的宠物

1. 打开 nsgo-pet。NSGo 酱会在屏幕右下角向你打招呼。
2. 把她拖到任意位置，或点她看看反应。
3. 右键点她，或使用 🐾 托盘图标，可以切换宠物、播放事件、把她移回角落、检查更新或退出。

宠物只有身体部分接收指针，点在别处会落到下面的窗口。你不点它的时候，它会待机，并按时段、季节、饭点、节日，以及前台是不是编辑器或终端来播放事件，还会不时沿着屏幕底部走动。

nsgo-pet 会记住你选的宠物，并从存储库的发行版自动更新：每天检查一次，也可以随时从菜单里检查。

## 平台

| 平台 | 状态 |
| --- | --- |
| macOS | 可用，播放带透明通道的 HEVC 动画。 |
| Linux | 播放 WebM 动画。窗口在 X11 上可以浮于顶层；在 Wayland 上无法浮于顶层，也无法自己移动。 |
| Windows | 暂不支持：NSGo 的 Windows 播放器不接受带透明通道的视频。 |

农历节日、按动画透明通道的命中测试（[nsgo#599](https://github.com/nsgo-dev/nsgo/issues/599)）、对话气泡和天气还没有实现。
