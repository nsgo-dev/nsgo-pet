#!/bin/bash
# Prints the section of CHANGELOG.md for version, such as 0.2.0, without its heading: the release's notes. Fails when the section is missing or empty, so a release never goes out without notes.
set -euo pipefail

version=${1:?usage: tools/release-notes.sh <version>}
root=$(cd "$(dirname "$0")/.." && pwd)
notes=$(awk -v heading="## [$version]" '
	index($0, heading) == 1 { found = 1; next }
	found && /^## / { exit }
	found { print }
' "$root/CHANGELOG.md")
if ! grep -q '[^[:space:]]' <<<"$notes"; then
	echo "CHANGELOG.md has no notes for $version under \"## [$version] - <date>\"" >&2
	exit 1
fi
printf '%s\n' "$notes"
