#!/bin/bash
# Fetches the frames tools/framepet/pets/mochi.json uses from Mochi (MIT, miflow13/mochi-desktop) and makes resources/pets/mochi with tools/framepet. Needs curl and an ffmpeg with libvpx-vp9 and hevc_videotoolbox (macOS).
set -euo pipefail

repo=miflow13/mochi-desktop
ref=471342631206c40d56bbba016578926e16ca8d0b
root=$(cd "$(dirname "$0")/.." && pwd)
spec=tools/framepet/pets/mochi.json
out=resources/pets/mochi
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

cd "$root"
go run ./tools/framepet -spec "$spec" -list | while read -r f; do
	printf 'url = "https://raw.githubusercontent.com/%s/%s/assets/mochi/%s"\noutput = "%s/%s"\n' "$repo" "$ref" "$f" "$tmp" "$f"
done >"$tmp/curl.cfg"
curl -fsS --max-time 300 --parallel --create-dirs -K "$tmp/curl.cfg"
go run ./tools/framepet -spec "$spec" -frames "$tmp" -out "$out"
cat >"$out/NOTICE.md" <<NOTICE
# Mochi

From [Mochi](https://github.com/$repo) at \`$ref\` (\`assets/mochi\`), by the Mochi contributors. Its README says "Mochi is released under the [MIT License](LICENSE)", and its Artist Kit says "Mochi is released under the repository's [MIT License](../LICENSE). By submitting artwork for inclusion in Mochi, make sure you have the right to contribute it under the project's license." The clips are those frames played in sequence at the frame rates of \`assets/mochi/manifest.json\`, the 64-pixel sheets scaled to the 256-pixel canvas as Mochi does, cropped by \`tools/framepet\` from \`$spec\`.

Its [LICENSE](https://github.com/$repo/blob/$ref/LICENSE):

\`\`\`text
NOTICE
curl -fsS --max-time 60 "https://raw.githubusercontent.com/$repo/$ref/LICENSE" >>"$out/NOTICE.md"
printf '\140\140\140\n' >>"$out/NOTICE.md"
