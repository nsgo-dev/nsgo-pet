---
title: 嵌入引擎
description: 在 NSGo 应用里展示你自己的宠物，放进你的布局，或放进它自己的窗口。
group: 开发者
sidebar:
  order: 3
---

宠物引擎是 Go 模块 `github.com/nsgo-dev/nsgo-pet/pet`：包含宠物包格式、动画规则和播放它们的视图。它不带任何宠物和素材，由你的应用自带宠物包，各包遵循各自的许可。

```sh
go get github.com/nsgo-dev/nsgo-pet/pet@main
```

## 加载宠物包

从目录加载宠物包，或用 `pet.Load(fsys, "asset:pets")` 从应用资源加载。加载失败的包会被跳过，并在错误里写明：

```go
packs, err := pet.LoadDir(dir)
if len(packs) == 0 {
	return err
}
```

## 展示宠物

作为一个视图放进你自己的布局：

```go
func Sidebar(cx *core.Scope) core.Node {
	p := pet.New(cx, packs[0], pet.Options{
		Menu:    menu,
		Context: func() pet.Context { return pet.Context{Working: busy.Get()} },
	})
	cx.Post(p.Start)
	return p.View()
}
```

或放进它自己的窗口，走动和拖拽会移动这个窗口：

```go
desktop.OpenWindow(cx, pet.WindowOptions(pack), func(cx *core.Scope) core.Node {
	p := pet.New(cx, pack, pet.Options{Window: true, Menu: menu})
	cx.Post(p.Start)
	return p.View()
})
```

视图里的宠物没有可移动的窗口，所以会原地播放走动动画，拖拽时也在原处悬空。

## 由你的应用决定的事

窗口、菜单和生命周期都归你的应用管。引擎不建托盘、不检查更新、不存任何东西，也不会退出应用。

| 调用 | 作用 |
| --- | --- |
| `Act` | 立即播放一个事件 |
| `Play(id)` | 按 ID 播放动画，比如对应你应用的运行中、等待或完成状态 |
| `PlaceAtCorner` | 把宠物自己的窗口移到屏幕右下角 |
| `Leave(done)` | 播放退场动画后调用 `done`，在里面关闭宠物的窗口或视图 |

`Options.Context` 告诉动画的 `Working` 和 `Coding` 权重用户在做什么；不提供时宠物会看前台应用。

为你的应用制作宠物包，见 [宠物包](../packs/)。
