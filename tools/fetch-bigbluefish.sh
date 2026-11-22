#!/bin/bash
# Fetches Big Blue Fish's 640x360 clips again (VP9 with alpha in WebM) into resources/fish/webm, and transcodes each to HEVC with alpha in resources/fish/mov for Apple platforms, whose player opens no WebM. Needs gh, curl and an ffmpeg with libvpx-vp9 and hevc_videotoolbox (macOS).
set -euo pipefail

repo=Bai-Cute/big-blue-fish-desktop-pet
ref=75cd217568e713a7de8812f65dad7072798a8b70
dir=code/BigBlueFish.Windows/assets/fish
root=$(cd "$(dirname "$0")/.." && pwd)
out=$root/resources/pets/bigbluefish
mkdir -p "$out/webm" "$out/mov"

for f in animations.json LICENSE.txt; do
	gh api -H 'Accept: application/vnd.github.raw' "repos/$repo/contents/$dir/$f?ref=$ref" >"$out/$f"
done

gh api "repos/$repo/contents/$dir/compat?ref=$ref" --jq '.[] | "\(.name)\t\(.download_url)"' |
	while IFS=$'\t' read -r name url; do
		[ -s "$out/webm/$name" ] || curl -fsSL --max-time 120 -o "$out/webm/$name" "$url"
	done

transcode() {
	local in=$1 mov
	mov="$2/$(basename "${in%.webm}").mov"
	[ -s "$mov" ] && return 0
	# libvpx decodes the alpha the WebM carries; VideoToolbox encodes it as HEVC's alpha layer.
	timeout 300 ffmpeg -v error -y -c:v libvpx-vp9 -i "$in" -an -vf format=rgba,premultiply=inplace=1 -c:v hevc_videotoolbox -alpha_quality 0.75 -q:v 60 -tag:v hvc1 -pix_fmt bgra "$mov"
}
export -f transcode
find "$out/webm" -name '*.webm' -print0 | xargs -0 -P 4 -I{} bash -c 'transcode "$1" "$2"' _ {} "$out/mov"

echo "$(find "$out/webm" -name '*.webm' | wc -l) WebM, $(find "$out/mov" -name '*.mov' | wc -l) MOV in $out"
