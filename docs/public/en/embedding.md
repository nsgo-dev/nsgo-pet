---
title: Embedding the engine
description: Add a desktop pet to your own NSGo app, in your layout or in a window of its own.
---

The pet engine behind nsgo-pet is a Go module any [NSGo](https://github.com/nsgo-dev/nsgo) app can import: the pack format, the rules that choose clips and the view that plays them. It ships no pets; your app brings its own packs, each under its own license.

## Add the module

```sh
go get github.com/nsgo-dev/nsgo-pet/pet@main
```

## Bring a pack

A pet is a [pack](../packs/): a folder of clips and rules, with no code. Make your own, or ship one of [nsgo-pet's pets](../pets/) by copying its folder from [`resources/pets/`](https://github.com/nsgo-dev/nsgo-pet/tree/main/resources/pets) into your app's `resources/pets/`, keeping its `NOTICE.md` and its license.

Embed the packs' manifests and rules, and the engine plays their clips from your app's resources. Packs in a directory load with `pet.LoadDir(dir)` instead. A pack that fails to load is left out and named in the error:

```go
//go:embed resources/pets/*/pet.json resources/pets/*/animations.json
var petFiles embed.FS

func loadPacks() ([]*pet.Pack, error) {
	pets, _ := fs.Sub(petFiles, "resources/pets")
	return pet.Load(pets, "asset:pets")
}
```

## Show a pet

As a view in your own layout:

```go
func Sidebar(cx *core.Scope) core.Node {
	p := pet.New(cx, packs[0], pet.Options{Menu: menu})
	cx.Post(p.Start)
	return p.View()
}
```

Or floating on the desktop in a window of its own, which its walks and drags move:

```go
desktop.OpenWindow(cx, pet.WindowOptions(pack), func(cx *core.Scope) core.Node {
	p := pet.New(cx, pack, pet.Options{Window: true, Menu: menu})
	cx.Post(p.Start)
	return p.View()
})
```

A pet in a view has no window to move, so it plays its walking clips in place and dangles where it is.

## Make it react to your app

Your app owns the windows, menus and lifecycle: the engine opens no tray, checks for no updates, stores nothing and never quits. Drive the pet from your app's state instead:

| Call | Does |
| --- | --- |
| `Act` | Plays an event now |
| `Play(id)` | Plays a clip by ID, such as one for your app's running, waiting or done state |
| `PlaceAtCorner` | Moves the pet's own window to the bottom right of the screen |
| `Leave(done)` | Plays the exit clip, then calls `done`, where you close the pet's window or view |

To have the pet's clips follow what the user is doing in your app, rather than which app is in front, pass `Options.Context`:

```go
pet.Options{
	Context: func() pet.Context { return pet.Context{Working: busy.Get()} },
}
```

## Licenses

The engine is MIT; no pack is. Ship each pack's license with it, and keep in mind that a pack under a non-commercial license, such as NSGo-chan's or Big Blue Fish's, makes the build that includes it non-commercial.
