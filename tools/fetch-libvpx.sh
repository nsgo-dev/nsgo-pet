#!/bin/bash
# Fetches the libvpx DLL the Windows build decodes the WebM clips with, and the winpthreads DLL it links, from MSYS2's signed mingw64 packages (pinned by their SHA-256), into resources/windows with their licenses. Needs curl and a tar that reads zstd.
set -euo pipefail

mirror=https://mirror.msys2.org/mingw/mingw64
vpx=mingw-w64-x86_64-libvpx-1.16.0-1-any.pkg.tar.zst
vpx_sum=e1c166239223d1f723971d89b932bf6fdfaaa175e1908f07f09d673188c6bab2
pthread=mingw-w64-x86_64-libwinpthread-14.0.0.r92.g818fa6510-1-any.pkg.tar.zst
pthread_sum=c426ba9ef8870b4342d5c0843efa7a9b360361994c7556f3d18f23fd30734194
# The packages carry winpthreads' license but not libvpx's, which comes from the tag the package was built from.
libvpx_tag=v1.16.0
root=$(cd "$(dirname "$0")/.." && pwd)
out=$root/resources/windows
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

fetch() {
	curl -fsSL --max-time 300 -o "$tmp/$1" "$mirror/$1"
	if [ "$(shasum -a 256 "$tmp/$1" | cut -d' ' -f1)" != "$2" ]; then
		echo "$1 does not match its SHA-256" >&2
		exit 1
	fi
	tar --zstd -xf "$tmp/$1" -C "$tmp"
}
fetch "$vpx" "$vpx_sum"
fetch "$pthread" "$pthread_sum"

mkdir -p "$out"
cp "$tmp/mingw64/bin/libvpx-1.dll" "$tmp/mingw64/bin/libwinpthread-1.dll" "$out/"
for f in LICENSE PATENTS; do
	curl -fsS --max-time 60 -o "$out/libvpx-$f" "https://raw.githubusercontent.com/webmproject/libvpx/$libvpx_tag/$f"
done
cp "$tmp/mingw64/share/licenses/libwinpthread/COPYING" "$out/libwinpthread-COPYING"
cat >"$out/NOTICE.md" <<NOTICE
# libvpx for Windows

The Windows build decodes the WebM clips with libvpx, which NSGo's Windows player cannot do with their alpha. These files come unchanged from MSYS2's mingw64 packages, fetched by \`tools/fetch-libvpx.sh\`:

| File | Package | License |
| --- | --- | --- |
| \`libvpx-1.dll\` | [$vpx]($mirror/$vpx), libvpx ${libvpx_tag#v} by the WebM Project | BSD-3-Clause, with Google's patent grant: [libvpx-LICENSE](libvpx-LICENSE), [libvpx-PATENTS](libvpx-PATENTS) |
| \`libwinpthread-1.dll\` | [$pthread]($mirror/$pthread), winpthreads by the mingw-w64 project, which libvpx-1.dll links | MIT and others: [libwinpthread-COPYING](libwinpthread-COPYING) |

The libvpx license and patent grant are those of its [$libvpx_tag](https://github.com/webmproject/libvpx/tree/$libvpx_tag) tag, which the package does not carry.
NOTICE
