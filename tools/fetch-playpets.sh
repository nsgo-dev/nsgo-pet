#!/bin/bash
# Fetches PlayPets' four starter pets (MIT, gaerae/playpets) with their summer and winter outfits, and converts each with tools/codexpet into resources/pets/playpets-<name>.
set -euo pipefail

repo=gaerae/playpets
ref=852a06225c1bac93123dc07a56f6d7e500b55be1
root=$(cd "$(dirname "$0")/.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

for pet in clover daisy pip miso; do
	base=https://raw.githubusercontent.com/$repo/$ref/assets/pets/$pet
	name=$(curl -fsSL "$base/pet.json" | sed -n 's/.*"displayName": *"\([^"]*\)".*/\1/p')
	args=()
	curl -fsSL -o "$tmp/$pet.png" "$base/spritesheet.png"
	for season in summer winter; do
		if curl -fsSL -o "$tmp/$pet-$season.png" "$base/seasonal/$season.png"; then
			args+=("-$season" "$tmp/$pet-$season.png")
		fi
	done
	(cd "$root" && go run ./tools/codexpet -sheet "$tmp/$pet.png" "${args[@]}" -out "resources/pets/playpets-$pet" \
		-name "$name" -group PlayPets -source "https://github.com/$repo" -license "MIT, Copyright (c) gaerae")
	cat >"$root/resources/pets/playpets-$pet/NOTICE.md" <<NOTICE
# $name

From [PlayPets](https://github.com/$repo) at \`$ref\` (\`assets/pets/$pet\`), by gaerae. Its README says the MIT license covers "the pet mascot art in \`assets/\`". The clips are the sheet's rows, converted by \`tools/codexpet\`.

MIT License, Copyright (c) gaerae: see <https://github.com/$repo/blob/$ref/LICENSE>.
NOTICE
done
