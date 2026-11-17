#!/bin/bash
# Fetches OpenPets' default pet (OpenPetsHQ/openpets, MIT) with its outfits and converts it with tools/codexpet: the summer outfit plays in summer and the holidays outfit in winter, and the Halloween, pajamas and weekend outfits are pets of their own.
set -euo pipefail

repo=OpenPetsHQ/openpets
ref=163c80281c5c2739776773e4993516471bfd71cf
raw=https://raw.githubusercontent.com/$repo/$ref/apps/desktop/assets
root=$(cd "$(dirname "$0")/.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

curl -fsSL -o "$tmp/default.webp" "$raw/default-pet-spritesheet.webp"
for outfit in summer holidays halloween pajamas weekend; do
	curl -fsSL -o "$tmp/$outfit.webp" "$raw/default-pet-outfits/$outfit.webp"
done

pet() {
	local id=$1 name=$2 note=$3
	shift 3
	(cd "$root" && go run ./tools/codexpet "$@" -out "resources/pets/$id" -name "$name" -group OpenPets \
		-source "https://github.com/$repo" -license "MIT, Copyright (c) 2026 OpenPets")
	cat >"$root/resources/pets/$id/NOTICE.md" <<NOTICE
# $name

From [OpenPets](https://github.com/$repo) at \`$ref\`: $note. The repository's MIT license covers it (the repository makes no separate statement about its art). The clips are the sheets' rows, converted by \`tools/codexpet\`.

MIT License, Copyright (c) 2026 OpenPets: see <https://github.com/$repo/blob/$ref/LICENSE>.
NOTICE
}

pet openpets "OpenPets" "\`apps/desktop/assets/default-pet-spritesheet.webp\`, with \`default-pet-outfits/summer.webp\` in summer and \`holidays.webp\` in winter" \
	-sheet "$tmp/default.webp" -summer "$tmp/summer.webp" -winter "$tmp/holidays.webp"
pet openpets-halloween "OpenPets (Halloween)" "\`apps/desktop/assets/default-pet-outfits/halloween.webp\`" -sheet "$tmp/halloween.webp"
pet openpets-pajamas "OpenPets (Pajamas)" "\`apps/desktop/assets/default-pet-outfits/pajamas.webp\`" -sheet "$tmp/pajamas.webp"
pet openpets-weekend "OpenPets (Weekend)" "\`apps/desktop/assets/default-pet-outfits/weekend.webp\`" -sheet "$tmp/weekend.webp"
