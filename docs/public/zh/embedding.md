---
title: 嵌入引擎
description: 给你自己的 NSGo 应用加一只桌面宠物，放进你的布局，或放进它自己的窗口。
---

nsgo-pet 背后的宠物引擎是一个 Go 模块，任何 [NSGo](https://github.com/nsgo-dev/nsgo) 应用都能导入：它包含宠物包格式、挑选动画的规则和播放动画的视图。它不带任何宠物，由你的应用自带宠物包，各包遵循各自的许可。

## 添加模块

```sh
go get github.com/nsgo-dev/nsgo-pet/pet@main
```

## 准备宠物包

一只宠物就是一个[宠物包](../packs/)：装着动画和规则的文件夹，不需要写代码。你可以自己做，也可以分发 [nsgo-pet 的宠物](../pets/)：把它的文件夹从 [`resources/pets/`](https://github.com/nsgo-dev/nsgo-pet/tree/main/resources/pets) 复制到你应用的 `resources/pets/`，保留它的 `NOTICE.md` 和许可证。

嵌入宠物包的清单和规则，引擎会从应用资源播放它们的动画；放在目录里的宠物包则用 `pet.LoadDir(dir)` 加载。加载失败的包会被跳过，并在错误里写明：

```go
//go:embed resources/pets/*/pet.json resources/pets/*/animations.json
var petFiles embed.FS

func loadPacks() ([]*pet.Pack, error) {
	pets, _ := fs.Sub(petFiles, "resources/pets")
	return pet.Load(pets, "asset:pets")
}
```

## 展示宠物

作为一个视图放进你自己的布局：

```go
func Sidebar(cx *core.Scope) core.Node {
	p := pet.New(cx, packs[0], pet.Options{Menu: menu})
	cx.Post(p.Start)
	return p.View()
}
```

或者放进它自己的窗口浮在桌面上，走动和拖拽会移动这个窗口：

```go
desktop.OpenWindow(cx, pet.WindowOptions(pack), func(cx *core.Scope) core.Node {
	p := pet.New(cx, pack, pet.Options{Window: true, Menu: menu})
	cx.Post(p.Start)
	return p.View()
})
```

视图里的宠物没有可移动的窗口，所以会原地播放走动动画，拖拽时也在原处悬空。

## 让宠物回应你的应用

窗口、菜单和生命周期都归你的应用管：引擎不建托盘、不检查更新、不存任何东西，也不会退出应用。你可以按应用的状态驱动宠物：

| 调用 | 作用 |
| --- | --- |
| `Act` | 立即播放一个事件 |
| `Play(id)` | 按 ID 播放动画，比如对应你应用的运行中、等待或完成状态 |
| `PlaceAtCorner` | 把宠物自己的窗口移到屏幕右下角 |
| `Leave(done)` | 播放退场动画后调用 `done`，在里面关闭宠物的窗口或视图 |

想让动画跟着用户在你应用里做的事走，而不是看前台是哪个应用，就传入 `Options.Context`：

```go
pet.Options{
	Context: func() pet.Context { return pet.Context{Working: busy.Get()} },
}
```

## 许可

引擎采用 MIT 许可，宠物包则不然。分发每个宠物包时带上它的许可证；包含非商用许可宠物包（例如 NSGo 酱或大蓝鱼）的构建也只能非商用。
