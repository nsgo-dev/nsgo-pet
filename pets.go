package main

import (
	"embed"
	"io/fs"
	"slices"

	"github.com/nsgo-dev/nsgo-pet/pet"
)

// petFiles are every pet's manifest and clip rules; the clips themselves ship as resources beside them (resources/pets/<id>/mov and webm), which the player reads by asset path.
//
//go:embed resources/pets/*/pet.json resources/pets/*/animations.json
var petFiles embed.FS

// defaultPet is the pet shown until the user chooses another.
const defaultPet = "nsgo-chan"

// LoadPacks reads every pet in resources/pets, by ID, defaultPet first; a pet that fails to load fails them all, as this app ships only pets that load.
func LoadPacks() ([]*pet.Pack, error) {
	pets, err := fs.Sub(petFiles, "resources/pets")
	if err != nil {
		return nil, err
	}
	packs, err := pet.Load(pets, "asset:pets")
	if err != nil {
		return nil, err
	}
	slices.SortStableFunc(packs, func(a, b *pet.Pack) int {
		switch {
		case a.ID == defaultPet && b.ID != defaultPet:
			return -1
		case b.ID == defaultPet && a.ID != defaultPet:
			return 1
		}
		return 0
	})
	return packs, nil
}
