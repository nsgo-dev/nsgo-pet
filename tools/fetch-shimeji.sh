#!/bin/bash
# Fetches the images tools/framepet/pets/shimeji.json uses from Shimeji for Linux (zlib, estenv/linux-shimeji, the original Shimeji by Yuki Yamada) and makes resources/pets/shimeji with tools/framepet. The spec's clips follow conf/Actions.xml: its poses' images and lengths, in ticks of 20 ms (50 per second). Needs curl and an ffmpeg with libvpx-vp9 and hevc_videotoolbox (macOS).
set -euo pipefail

repo=estenv/linux-shimeji
ref=b90492d326cb24b9ceba914481a78a5e2470c0dc
root=$(cd "$(dirname "$0")/.." && pwd)
spec=tools/framepet/pets/shimeji.json
out=resources/pets/shimeji
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

cd "$root"
go run ./tools/framepet -spec "$spec" -list | while read -r f; do
	printf 'url = "https://raw.githubusercontent.com/%s/%s/img/%s"\noutput = "%s/%s"\n' "$repo" "$ref" "$f" "$tmp" "$f"
done >"$tmp/curl.cfg"
curl -fsS --max-time 300 --parallel --create-dirs -K "$tmp/curl.cfg"
go run ./tools/framepet -spec "$spec" -frames "$tmp" -out "$out"
cat >"$out/NOTICE.md" <<NOTICE
# Shimeji

From [Shimeji for Linux](https://github.com/$repo) at \`$ref\` (\`img/shime*.png\`), a port of the original [Shimeji](http://www.group-finity.com/Shimeji) by Yuki Yamada. Its README says: "This project inherits the ZLIB/LIBPNG license of the original program."

**This is an altered version.** The images are cropped, scaled and encoded as video clips by \`tools/framepet\` from \`$spec\`. The clips follow the poses and lengths of \`conf/Actions.xml\`: 落下する (drop-in), ジャンプ (jump), 立つ (stand), 座る (sit), 座って見上げる (look-up), 座って首が回る (head-spin), 寝そべる (sprawl), ずりずり played in place (creep), 転ぶ without its last pose (trip), 跳ねる (bounce), 分裂1 (split), and ドラッグされる (pinched). They play in nsgo-pet, not in Shimeji, and are not the original software.

Its [LICENSE](https://github.com/$repo/blob/$ref/LICENSE), unaltered:

\`\`\`text
NOTICE
curl -fsS --max-time 60 "https://raw.githubusercontent.com/$repo/$ref/LICENSE" | tr -d '\r' >>"$out/NOTICE.md"
printf '\140\140\140\n' >>"$out/NOTICE.md"
