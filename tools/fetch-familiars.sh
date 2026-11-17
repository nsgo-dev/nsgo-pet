#!/bin/bash
# Fetches the original pets of Familiars (jalcantarab/agent-familiars, MIT) and converts each with tools/codexpet into resources/pets/familiars-<id>. Imported pets, which keep their upstream licenses, and Zentri, drawn from the Zentrik logo, whose brand rights the repository does not grant, are left out.
set -euo pipefail

repo=jalcantarab/agent-familiars
ref=34e78c3a34bb750bcf436c1c7858d87f095fdb81
raw=https://raw.githubusercontent.com/$repo/$ref
root=$(cd "$(dirname "$0")/.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

curl -fsSL -o "$tmp/catalog.json" "$raw/catalog/pets.json"
jq -r '.[] | select(.origin == "original" and .license == "MIT" and .id != "zentri") | [.id, .displayName, .folder, .author] | @tsv' "$tmp/catalog.json" >"$tmp/pets.tsv"

convert() {
	local id=$1 name=$2 folder=$3 author=$4 sheet
	sheet=$(curl -fsSL "$raw/$folder/pet.json" | jq -r '.spritesheetPath // "spritesheet.webp"')
	curl -fsSL -o "$tmp/$id.webp" "$raw/$folder/$sheet"
	go run ./tools/codexpet -sheet "$tmp/$id.webp" -out "resources/pets/familiars-$id" -name "$name" -group Familiars \
		-source "https://github.com/$repo/tree/$ref/$folder" -license "MIT, Copyright (c) Jorge Alcantara"
	cat >"resources/pets/familiars-$id/NOTICE.md" <<NOTICE
# $name

From [Familiars](https://github.com/$repo) at \`$ref\` (\`$folder\`), by $author. Its \`catalog/pets.json\` lists it as an original pet under the MIT license, and its NOTICE.md says the root MIT license applies to "original Familiars pet bundles". The clips are the sheet's rows, converted by \`tools/codexpet\`.

MIT License: see <https://github.com/$repo/blob/$ref/LICENSE>.
NOTICE
	rm -f "$tmp/$id.webp"
}
export -f convert
export raw repo ref tmp
cd "$root"
# go run builds once before the parallel runs reuse it.
go build -o /dev/null ./tools/codexpet
tr '\t' '\037' <"$tmp/pets.tsv" | xargs -P 4 -I{} bash -c 'IFS=$'"'"'\037'"'"' read -r id name folder author <<<"$1"; convert "$id" "$name" "$folder" "$author"' _ {}
