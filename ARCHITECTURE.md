# nsgo-pet architecture

nsgo-pet is two things in one repository: the pet engine, a Go module any NSGo app can import, and the nsgo-pet app, one host of that engine with its own pets, tray and updater. The engine plays a pack's clips by its rules; the host decides where the pet lives and when it goes.

## Code map

| Path | Responsibility |
| --- | --- |
| `pet/` | The engine, module `github.com/nsgo-dev/nsgo-pet/pet`: loading packs, choosing clips, the view and the pet's window behavior |
| `pet/testdata/pets/` | Fixture packs of JSON alone, so the engine tests without clips |
| `app.go`, `settings.go`, `pets.go`, `stats.go` | The app: the main window, tray, settings window, saved choices, the embedded manifests and the usage statistics |
| `resources/pets/` | The app's pets, their clips and notices; never part of the engine |
| `tools/` | Converters and fetch scripts that build packs, and the release notes script |
| `website/` | The website, which reads the pets' manifests and copies their clips at build time |
| `docs/public/` | The website's docs, one folder per language, for the app's users and the engine's |
| `.github/workflows/release.yml` | Builds, signs and publishes the app's installers on a tag |

## Boundaries

- The engine is a module of its own so that importing it downloads no pets: `resources/` is about 175 MB and belongs to the app alone. The root `go.mod` replaces the engine with `./pet`, so both change together in one commit.
- The engine ships no assets and no app policy. It opens no tray, checks for no updates, writes no storage, shows no text of its own and never calls `cx.Quit()`; menus, labels and lifecycle are the host's.
- On Windows, whose NSGo player refuses video with alpha ([nsgo#862](https://github.com/nsgo-dev/nsgo/issues/862)), the engine plays the WebM clips in its own decoder over libvpx, which the host ships in `resources/windows`; elsewhere it plays them with `core.Video`. `ClipView` exports the same choice, so a host's previews of a pack, such as the app's settings grid, play where its pet does.
- Everything that moves a window is behind `Options.Window`. A pet in a host's view plays walking clips in place and dangles where it is, as it has no window to move.
- What the user is doing reaches the clips through `Context` alone: the host's `Options.Context`, or the app in front without one. Clip rules never read the host's state another way.
- The engine's exported API is public: changing or removing any of it breaks hosts, so it needs the user's agreement first.
- `Load` isolates packs, leaving out each that fails. The app is stricter and fails to start on any broken pet, since it ships only pets that load.

## Core flows

**Playing.** Two players alternate. The next clip loads in the hidden one and shows once it has a picture, so the pet never blinks out between clips. As the front clip ends, `finished` picks what follows from the phase, and `boundary` chooses between the daily clip, an event every `eventSeconds` and an idle, mirroring Big Blue Fish's `AnimationFinished` and `AnimationBoundary`.

**Walking.** In a window of its own, an event with a `MoveKind` walks the window 70 points along the screen while the clip's middle plays, idles, then walks back mirrored. A drag or a click cancels the walk.

**Leaving.** `Leave` plays the exit clip and calls the host's `done` once, as the clip ends or eight seconds later. The app's `done` quits; a host showing the pet in a view removes it instead.

## Data and state

The engine keeps no state beyond the pet it plays. The app stores the ID of the chosen pet under the storage key `pet`, whether its usage statistics are off under `stats`, and the random ID its usage statistics count an install by under `installID`. It reads packs from the manifests embedded in its binary, playing their clips from its bundled resources.

## Verification

Both test suites run headless on `nsgotest`, which loads no media. The app's tests check every shipped pet's manifest and clip files, the policy's weights and the menus; the engine's cover loading, host state, playing by ID and leaving against fixture packs, and its Windows decoder against the app's clips where libvpx loads. Real playback, transparency, floating, walking at a screen's edge and Linux's X11 and Wayland differences need a manual run with `go tool nsgo dev`.

## Subjects

| Document | When to read |
| --- | --- |
| [Pet packs](docs/architecture/packs.md) | Making a pack, changing the pack format or the clip rules |
