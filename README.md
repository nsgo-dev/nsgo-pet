<p align="center">
  <img src="assets/Icon.png" alt="nsgo-pet" width="128">
</p>

<h1 align="center">nsgo-pet</h1>

<p align="center">
  <strong>English</strong> |
  <a href="README.zh-CN.md">简体中文</a>
</p>

<p align="center">
  <a href="https://github.com/oneo-me"><img alt="GitHub" src="https://img.shields.io/badge/GitHub-oneo--me-181717?logo=github&logoColor=white"></a>
  <a href="https://oneo.me"><img alt="Blog" src="https://img.shields.io/badge/Blog-oneo.me-0969da"></a>
  <a href="https://x.com/oneo666"><img alt="X" src="https://img.shields.io/badge/X-%40oneo666-000000?logo=x&logoColor=white"></a>
  <a href="https://github.com/sponsors/oneo-me"><img alt="GitHub Sponsors" src="https://img.shields.io/badge/GitHub%20Sponsors-ea4aaa?logo=githubsponsors&logoColor=white"></a>
</p>

<p align="center">
  A desktop pet for macOS and Linux built with <a href="https://github.com/nsgo-dev/nsgo">NSGo</a>, and the pet engine behind it, which other NSGo apps can embed.
</p>

<p align="center">
  <img src="docs/nsgo-chan.png" alt="NSGo 酱 and her Gopher greeting, coding, spawning goroutines, high-fiving, snacking and sleeping">
</p>

## Features

- Floats in a transparent window above other windows, on every Space and over full-screen apps, and stays off the Dock and the taskbar.
- Takes the pointer only on the pet's body, so clicks elsewhere reach the window below.
- Greets you, idles and plays events chosen by time of day, season, meals, festivals and whether an editor or a terminal is in front.
- Reacts to clicks, dangles while dragged and walks along the bottom of the screen.
- Switches between eleven pets from the tray icon or its right-click menu, and remembers your choice.
- Updates itself from this repository's releases.

## Install

Download the disk image for macOS, or the Debian package, Arch package or AppImage for Linux, from the [releases](https://github.com/nsgo-dev/nsgo-pet/releases). Until the first release, build them from the repository root with `go tool nsgo build`, which writes them to `dist/`.

Linux needs GStreamer's bad plugins (`gstreamer1.0-plugins-bad`) to play clips with alpha; the Debian and Arch packages depend on them.

## Getting started

1. Open nsgo-pet. NSGo 酱 greets you in the bottom right corner of the screen.
2. Drag her anywhere, or click her for a reaction.
3. Right-click her, or use the 🐾 tray icon, to switch pets, play an event, move her back to the corner, check for updates or quit.

## Embedding the pet engine

The engine is the Go module `github.com/nsgo-dev/nsgo-pet/pet`: the pack format, the clip rules and the view that plays them. It ships no pets and no assets; your app brings its own packs, each under its own license.

```sh
go get github.com/nsgo-dev/nsgo-pet/pet@main
```

Load packs from a directory, or from your app's resources with `pet.Load(fsys, "asset:pets")`. A pack that fails to load is left out and named in the error:

```go
packs, err := pet.LoadDir(dir)
if len(packs) == 0 {
	return err
}
```

Show a pet as a view in your own layout:

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

Your app owns the windows, menus and lifecycle. The engine opens no tray, checks for no updates, stores nothing and never quits the app.

| Call | Does |
| --- | --- |
| `Act` | Plays an event now |
| `Play(id)` | Plays a clip by ID, such as one for your app's running, waiting or done state |
| `PlaceAtCorner` | Moves the pet's own window to the bottom right of the screen |
| `Leave(done)` | Plays the exit clip, then calls `done`, where you close the pet's window or view |

`Options.Context` tells the clips' `Working` and `Coding` weights what the user is doing; without it the pet looks at the app in front. The [pack format](docs/architecture/packs.md) describes how to make packs.

## Pets

| Pet | ID | Clips | Source | License |
| --- | --- | --- | --- | --- |
| NSGo 酱 NSGo-chan | `nsgo-chan` | 15 | Drawn for nsgo-pet, with her Gopher, from the sheets in [docs/nsgo-chan](docs/nsgo-chan) | [CC BY-NC 4.0](https://creativecommons.org/licenses/by-nc/4.0/), Copyright (c) 2026 nsgo-dev: **no commercial use** ([NOTICE](resources/pets/nsgo-chan/NOTICE.md)) |
| 大蓝鱼 Big Blue Fish | `bigbluefish` | 106 | [PC2005-cloud/dsh-pet](https://github.com/PC2005-cloud/dsh-pet), via [Bai-Cute/big-blue-fish-desktop-pet](https://github.com/Bai-Cute/big-blue-fish-desktop-pet) | Character [CC BY-NC-SA 4.0](https://creativecommons.org/licenses/by-nc-sa/4.0/): **no commercial use** ([NOTICE](resources/pets/bigbluefish/NOTICE.md)) |
| Clover | `playpets-clover` | 35 | [gaerae/playpets](https://github.com/gaerae/playpets), with its summer and winter outfits | MIT, gaerae ([NOTICE](resources/pets/playpets-clover/NOTICE.md)) |
| Daisy | `playpets-daisy` | 35 | [gaerae/playpets](https://github.com/gaerae/playpets), with its summer and winter outfits | MIT, gaerae ([NOTICE](resources/pets/playpets-daisy/NOTICE.md)) |
| Pip | `playpets-pip` | 35 | [gaerae/playpets](https://github.com/gaerae/playpets), with its summer and winter outfits | MIT, gaerae ([NOTICE](resources/pets/playpets-pip/NOTICE.md)) |
| Miso | `playpets-miso` | 35 | [gaerae/playpets](https://github.com/gaerae/playpets), with its summer and winter outfits | MIT, gaerae ([NOTICE](resources/pets/playpets-miso/NOTICE.md)) |
| Pip | `pip-mascot` | 13 | [TenzinDhonyoe/pip-mascot](https://github.com/TenzinDhonyoe/pip-mascot) | MIT, Copyright (c) 2026 Tenzin Dhonyoe ([NOTICE](resources/pets/pip-mascot/NOTICE.md)) |
| Mochi | `mochi` | 20 | [miflow13/mochi-desktop](https://github.com/miflow13/mochi-desktop) | MIT, Copyright (c) 2026 Mochi contributors ([NOTICE](resources/pets/mochi/NOTICE.md)) |
| Shimeji | `shimeji` | 14 | [estenv/linux-shimeji](https://github.com/estenv/linux-shimeji), the original Shimeji by Yuki Yamada | zlib, Copyright (c) 2009-2011 Yuki Yamada; altered ([NOTICE](resources/pets/shimeji/NOTICE.md)) |
| OpenPets | `openpets` | 35 | [OpenPetsHQ/openpets](https://github.com/OpenPetsHQ/openpets), with its summer and holidays outfits | MIT, Copyright (c) 2026 OpenPets ([NOTICE](resources/pets/openpets/NOTICE.md)) |
| OpenPets (Halloween) | `openpets-halloween` | 10 | [OpenPetsHQ/openpets](https://github.com/OpenPetsHQ/openpets), its Halloween outfit | MIT, Copyright (c) 2026 OpenPets ([NOTICE](resources/pets/openpets-halloween/NOTICE.md)) |

## Platforms

| Platform | State |
| --- | --- |
| macOS | Runs, playing the HEVC-with-alpha clips. |
| Linux | Plays the WebM clips. The window floats on X11; on Wayland it cannot float or move itself. |
| Windows | Not yet: NSGo's Windows player refuses clips with alpha. |

Lunar festivals, hit testing by the clip's alpha ([nsgo#599](https://github.com/nsgo-dev/nsgo/issues/599)), speech bubbles and weather are not done yet.

## Development

Needs Go 1.27.3 or later. From the repository root:

```sh
go tool nsgo dev           # build and open the pet
go tool nsgo test .        # the app's headless tests, which check every pet in resources/pets
(cd pet && go test ./...)  # the engine's tests
```

The website is in `website/` and its docs in `docs/public/`; with [Bun](https://bun.sh), `bun install && bun run dev` there serves them and `bun run build` builds them.

To release, move `CHANGELOG.md`'s `Unreleased` bullets under `## [<version>] - <date>`, set `version` in `nsgo.yaml` to the same version and push the tag `v<version>`. [The release workflow](.github/workflows/release.yml) builds, signs and publishes the installers with that section as the notes.

Read [ARCHITECTURE.md](ARCHITECTURE.md) before changing code, and [AGENTS.md](AGENTS.md) for how Agents work here. Changes are listed in [CHANGELOG.md](CHANGELOG.md).

## License

The code, the app and the engine, is MIT, Copyright (c) 2026 nsgo-dev: see [LICENSE](LICENSE). Each pet keeps its own license, listed under [Pets](#pets) and in its `NOTICE.md`. NSGo-chan's art is CC BY-NC 4.0 and Big Blue Fish's character CC BY-NC-SA 4.0, so a build that includes either may not be used commercially.
