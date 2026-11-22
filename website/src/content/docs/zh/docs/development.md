---
title: 开发
description: 从源码运行、测试和发布 nsgo-pet。
group: 开发者
sidebar:
  order: 5
---

nsgo-pet 的存储库里有两样东西：宠物引擎，任何 NSGo 应用都能导入的 Go 模块；以及 nsgo-pet 应用，这个引擎的一个宿主，带着自己的宠物、托盘和更新器。修改代码前请先读 [ARCHITECTURE.md](https://github.com/nsgo-dev/nsgo-pet/blob/main/ARCHITECTURE.md)。

## 运行和测试

需要 Go 1.27.3 或更高版本。在存储库根目录：

```sh
go tool nsgo dev           # build and open the pet
go tool nsgo test .        # the app's headless tests, which check every pet in resources/pets
(cd pet && go test ./...)  # the engine's tests
```

两套测试都无界面运行，不加载媒体。真实播放、透明、浮于顶层、在屏幕边缘走动，以及 Linux 上 X11 和 Wayland 的差异，需要用 `go tool nsgo dev` 手动运行检查。

## 发布

把 `CHANGELOG.md` 里 `Unreleased` 下的条目移到 `## [<版本>] - <日期>` 下，把 `nsgo.yaml` 的 `version` 设为同一版本，然后推送标签 `v<版本>`。[发布工作流](https://github.com/nsgo-dev/nsgo-pet/blob/main/.github/workflows/release.yml) 会构建、签名并发布安装包，以该节作为发布说明。

变更记录见 [CHANGELOG.md](https://github.com/nsgo-dev/nsgo-pet/blob/main/CHANGELOG.md)。
