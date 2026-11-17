#!/bin/bash
# Fetches the frames tools/framepet/pets/pip-mascot.json uses from Pip (MIT, TenzinDhonyoe/pip-mascot) and makes resources/pets/pip-mascot with tools/framepet. Needs curl and an ffmpeg with libvpx-vp9 and hevc_videotoolbox (macOS).
set -euo pipefail

repo=TenzinDhonyoe/pip-mascot
ref=4b325c7eacb95459bc665e35053e48022e78d01d
root=$(cd "$(dirname "$0")/.." && pwd)
spec=tools/framepet/pets/pip-mascot.json
out=resources/pets/pip-mascot
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

cd "$root"
go run ./tools/framepet -spec "$spec" -list | while read -r f; do
	printf 'url = "https://raw.githubusercontent.com/%s/%s/Pip/Assets/%s"\noutput = "%s/%s"\n' "$repo" "$ref" "$f" "$tmp" "$f"
done >"$tmp/curl.cfg"
curl -fsS --max-time 300 --parallel --create-dirs -K "$tmp/curl.cfg"
go run ./tools/framepet -spec "$spec" -frames "$tmp" -out "$out"
cat >"$out/NOTICE.md" <<NOTICE
# Pip

From [Pip](https://github.com/$repo) at \`$ref\` (\`Pip/Assets\`), by Tenzin Dhonyoe. Its README says: "The artwork in \`mascot/\` and \`Pip/Assets/\` is original and shared under the same license." The clips are those frames played in sequence, cropped and scaled by \`tools/framepet\` from \`$spec\`.

Its [LICENSE](https://github.com/$repo/blob/$ref/LICENSE):

\`\`\`text
NOTICE
curl -fsS --max-time 60 "https://raw.githubusercontent.com/$repo/$ref/LICENSE" >>"$out/NOTICE.md"
printf '\140\140\140\n' >>"$out/NOTICE.md"
