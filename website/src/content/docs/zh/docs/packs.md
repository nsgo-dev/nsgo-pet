---
title: 宠物包
description: 构成一只宠物的动画和规则文件夹，不需要写代码。
group: 开发者
sidebar:
  order: 4
---

宠物包是一个文件夹，装着动画和挑选它们的规则，所以新增一只宠物不需要写代码。引擎读取宠物包，但自己不带任何宠物包。nsgo-pet 的宠物包在 [`resources/pets/<id>/`](https://github.com/nsgo-dev/nsgo-pet/tree/main/resources/pets)，宿主应用的宠物包则放在它的 `fs.FS` 或目录指向的位置。

## 结构

| 路径 | 内容 |
| --- | --- |
| `pet.json` | 宠物的名字、尺寸、身体范围，以及按名字指定的动画 |
| `animations.json` | 所有动画及其播放时机，采用[大蓝鱼的格式](https://github.com/Bai-Cute/big-blue-fish-desktop-pet/blob/75cd217568e713a7de8812f65dad7072798a8b70/code/BigBlueFish.Windows/assets/fish/animations.json) |
| `webm/<clip>.webm` | 每段动画的带透明通道 VP9 版本，供 Linux 和 Windows 使用 |
| `mov/<clip>.mov` | 同一段动画的带透明通道 HEVC 版本，供 macOS 使用，它的播放器打不开 WebM |
| `NOTICE.md` | 宠物的来源、作者和许可证 |

之所以要两种格式，是因为没有哪一种带透明通道的格式能在所有平台的播放器上播放。引擎按操作系统挑选格式，从不转码，所以宠物包的动画要么两种格式都有，要么都不带。

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

`body` 是按画面百分比给出的矩形。宠物只在这里接收指针，所以点在透明窗口其他地方会落到下面的窗口。`clips` 指定宠物在动画池之外播放的动画：启动时、离开时、被拖动时、作为后备待机，以及每天 `dailyHour` 点一次。清单里指定了 `animations.json` 没有的动画时，宠物包加载失败。

## animations.json

| 字段 | 含义 |
| --- | --- |
| `Pool` | `Idle` 为事件之间的待机，`Event` 每 `eventSeconds` 秒一次，以及 `Click`、`Greeting`、`Daily`、`Drag` 或 `Disabled` |
| `Weight` | 在同一动画池里被选中的相对可能性；为 0 时除非有条件提高它，否则不会播放 |
| `NightOnly`、`SeasonOnly`、`SummerWeight`、`WinterWeight`、`MorningWeight` | 时段和季节 |
| `Meal`、`MealBoost` | 在 8:00、12:00 或 18:00 前后一小时内提高权重 |
| `Festival`、`RadiusDays`、`FestivalMode`、`FestivalWeight` | 公历节日前后的权重；农历节日保持普通权重 |
| `WorkKind`、`WorkWeight` | `Coding` 或 `Working`：宿主的 `Context` 所说的用户正在做的事 |
| `MoveKind`、`HourParity` | `Run` 或 `Crab`：宠物走出 70 点再走回来，仅在它自己的窗口里 |

`Disabled` 动画不会被随机播放，但宿主仍可以按 ID 播放它们，宿主就是这样把自己的状态对应到动画上的。

## 制作动画

动画需要透明背景。从带透明通道的 PNG 帧制作：

```sh
ffmpeg -framerate 24 -i frames/%04d.png -c:v libvpx-vp9 -pix_fmt yuva420p -b:v 0 -crf 32 webm/idle.webm
ffmpeg -c:v libvpx-vp9 -i webm/idle.webm -an -c:v hevc_videotoolbox -alpha_quality 0.75 -q:v 60 -tag:v hvc1 -pix_fmt bgra mov/idle.mov
```

[`tools/`](https://github.com/nsgo-dev/nsgo-pet/tree/main/tools) 里的转换器会写出整个宠物包，包括两种格式的动画和两个 JSON 文件。各自的参数见 `-h`。

| 工具 | 输入 |
| --- | --- |
| `tools/codexpet` | 八行 192×208 的 Codex 精灵图，可带季节服装 |
| `tools/framepet` | PNG 帧，以及 `tools/framepet/pets/` 下的 JSON 描述 |
| `tools/fetch-*.sh` | 从固定提交重新获取并转换某个来源的宠物 |

## 许可

每个宠物包都在 `NOTICE.md` 和 `pet.json` 的 `license` 里带着自己的许可证。引擎的 MIT 许可不覆盖任何宠物包。宿主分发宠物包时要连同它们的许可证一起分发；包含非商用许可宠物包（例如大蓝鱼）的构建也只能非商用。
