package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"

	"github.com/nsgo-dev/nsgo/core"
)

// petFiles are every pet's manifest and clip rules; the clips themselves ship as resources beside them (resources/pets/<id>/mov and webm), which the player reads by asset path.
//
//go:embed resources/pets/*/pet.json resources/pets/*/animations.json
var petFiles embed.FS

// Pack is one pet: its manifest (pet.json) and its clip rules (animations.json, in Big Blue Fish's format).
type Pack struct {
	// ID is the pet's directory under resources/pets.
	ID   string `json:"-"`
	Name string `json:"name"`
	// Group is the submenu the pet is listed in, empty for the top of the list.
	Group  string      `json:"group"`
	Width  core.Points `json:"width"`
	Height core.Points `json:"height"`
	// Body is where the pet's body is in its clips, in percent of the frame: only there does the window take the pointer.
	Body struct {
		Left, Top, Width, Height float32
	} `json:"body"`
	EventSeconds int `json:"eventSeconds"`
	// Clips are the clips the pet plays by name.
	Clips struct {
		Greeting  string `json:"greeting"`
		Exit      string `json:"exit"`
		Drag      string `json:"drag"`
		Idle      string `json:"idle"`
		Daily     string `json:"daily"`
		DailyHour int    `json:"dailyHour"`
	} `json:"clips"`
	Source  string `json:"source"`
	License string `json:"license"`

	Policy *Policy `json:"-"`
}

// defaultPet is the pet shown until the user chooses another.
const defaultPet = "nsgo-chan"

// LoadPacks reads every pet in resources/pets, by ID, defaultPet first.
func LoadPacks() ([]*Pack, error) {
	dirs, err := fs.Glob(petFiles, "resources/pets/*/pet.json")
	if err != nil {
		return nil, err
	}
	sort.SliceStable(dirs, func(i, j int) bool {
		di, dj := path.Base(path.Dir(dirs[i])) == defaultPet, path.Base(path.Dir(dirs[j])) == defaultPet
		if di != dj {
			return di
		}
		return dirs[i] < dirs[j]
	})
	var packs []*Pack
	for _, manifest := range dirs {
		dir := path.Dir(manifest)
		p := &Pack{ID: path.Base(dir)}
		data, err := petFiles.ReadFile(manifest)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, p); err != nil {
			return nil, fmt.Errorf("%s: %w", manifest, err)
		}
		rules, err := petFiles.ReadFile(dir + "/animations.json")
		if err != nil {
			return nil, err
		}
		if p.Policy, err = NewPolicy(rules); err != nil {
			return nil, fmt.Errorf("%s: %w", dir, err)
		}
		if err := p.check(); err != nil {
			return nil, fmt.Errorf("%s: %w", dir, err)
		}
		packs = append(packs, p)
	}
	if len(packs) == 0 {
		return nil, fmt.Errorf("no pets in resources/pets")
	}
	return packs, nil
}

// check reports a manifest the pet cannot run with.
func (p *Pack) check() error {
	if p.Width <= 0 || p.Height <= 0 {
		return fmt.Errorf("pet.json: width and height must be positive")
	}
	for _, id := range []string{p.Clips.Greeting, p.Clips.Exit, p.Clips.Drag, p.Clips.Idle} {
		if p.Policy.Find(id).ID != id || id == "" {
			return fmt.Errorf("pet.json names clip %q, which animations.json lacks", id)
		}
	}
	if p.Clips.Daily != "" && p.Policy.Find(p.Clips.Daily).ID == "" {
		return fmt.Errorf("pet.json names daily clip %q, which animations.json lacks", p.Clips.Daily)
	}
	return nil
}
