# Big Blue Fish on NSGo

A desktop pet built with [NSGo](https://github.com/nsgo-dev/nsgo), using the character and animation rules of [Big Blue Fish](https://github.com/Bai-Cute/big-blue-fish-desktop-pet).

The fish lives in a transparent, borderless window that floats above other windows on every Space, stays off the Dock and the taskbar, and lets clicks through everywhere except her body. She greets you, idles, plays an event every 40 seconds chosen by the original project's weights (time of day, season, meals, Gregorian festivals, coding in the app in front), reacts to clicks, dangles while dragged, and walks along the bottom of the screen.

## Run

```sh
tools/fetch-assets.sh   # clips into resources/fish (about 150 MB)
go tool nsgo dev
go tool nsgo test .
```

Right-click the fish or use the menu bar icon to trigger an event, return her to the corner, or quit.

## Platforms

- macOS: plays the clips as HEVC with alpha, which `tools/fetch-assets.sh` transcodes from the WebM originals.
- Linux: plays the VP9 WebM clips directly; needs GStreamer's `gstreamer1.0-plugins-bad` (codecalpha).
- Windows: not yet. NSGo's Windows player refuses clips with alpha, so the clips need the app's own decoder feeding `ImageSource.Pixels`.

## Not yet ported

Lunar festivals (they keep their plain weight), speech bubbles, the AI agent, weather and news.

## License

The clips and `animations.json` come from Big Blue Fish under the MIT license, copyright PC2005-cloud: see `resources/fish/LICENSE.txt`.
