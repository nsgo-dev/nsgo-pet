---
title: Embedding the engine
description: Show pets of your own in an NSGo app, in your layout or in a window of their own.
group: Developers
sidebar:
  order: 3
---

The pet engine is the Go module `github.com/nsgo-dev/nsgo-pet/pet`: the pack format, the clip rules and the view that plays them. It ships no pets and no assets; your app brings its own packs, each under its own license.

```sh
go get github.com/nsgo-dev/nsgo-pet/pet@main
```

## Load packs

Load packs from a directory, or from your app's resources with `pet.Load(fsys, "asset:pets")`. A pack that fails to load is left out and named in the error:

```go
packs, err := pet.LoadDir(dir)
if len(packs) == 0 {
	return err
}
```

## Show a pet

As a view in your own layout:

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

Or in a window of its own, which its walks and drags move:

```go
desktop.OpenWindow(cx, pet.WindowOptions(pack), func(cx *core.Scope) core.Node {
	p := pet.New(cx, pack, pet.Options{Window: true, Menu: menu})
	cx.Post(p.Start)
	return p.View()
})
```

A pet in a view has no window to move, so it plays its walking clips in place and dangles where it is.

## What your app decides

Your app owns the windows, menus and lifecycle. The engine opens no tray, checks for no updates, stores nothing and never quits the app.

| Call | Does |
| --- | --- |
| `Act` | Plays an event now |
| `Play(id)` | Plays a clip by ID, such as one for your app's running, waiting or done state |
| `PlaceAtCorner` | Moves the pet's own window to the bottom right of the screen |
| `Leave(done)` | Plays the exit clip, then calls `done`, where you close the pet's window or view |

`Options.Context` tells the clips' `Working` and `Coding` weights what the user is doing; without it the pet looks at the app in front.

To make packs for your app, see [Pet packs](../packs/).
