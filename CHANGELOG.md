# Changelog

## [Unreleased]

## [0.1.0] - 2026-11-22

nsgo-pet is a desktop pet for macOS and Linux built with NSGo, and the pet engine behind it, which other NSGo apps can embed.

- Add a desktop pet that floats above other windows and plays its clips by time of day, season, meals, festivals and the app in front.
- Add eleven pets, among them NSGo-chan, Big Blue Fish, PlayPets, Pip, Mochi, Shimeji and OpenPets, switched in the settings window that the tray icon opens. (#2, #4, #8, #9, #10, #11 by @houko)
- Add signed installers for macOS and Linux that update themselves, checked daily and from the settings. (#12 by @houko)
- Add the website, with an interactive NSGo-chan, a preview of every pet and the docs in English and Chinese.
- Add anonymous usage statistics, which the settings turn off.
- Add the pet engine as the Go module `github.com/nsgo-dev/nsgo-pet/pet`, for NSGo apps to show their own pets in a window or a view. (#14 by @ayuayue, #15 by @oneo-me)
- License NSGo-chan's art under CC BY-NC 4.0, so a build that includes her may not be used commercially.
