package pet

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/nsgo-dev/nsgo/core"
)

// Pack is one pet: its manifest (pet.json) and its clip rules (animations.json, in Big Blue Fish's format).
type Pack struct {
	// ID is the pet's directory.
	ID   string `json:"-"`
	Name string `json:"name"`
	// Group is the submenu the pet is listed in, empty for the top of the list.
	Group  string      `json:"group"`
	Width  core.Points `json:"width"`
	Height core.Points `json:"height"`
	// Body is where the pet's body is in its clips, in percent of the frame: only there does the pet take the pointer.
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

	// Base is where the pet's mov and webm directories play from: an "asset:" path in the app's resources or a file path.
	Base   string  `json:"-"`
	Policy *Policy `json:"-"`
}

// Load reads every pet in fsys, each a directory at its top holding pet.json and animations.json, in the order of their IDs; base is where fsys's directories play from, such as "asset:pets" for resources/pets. A pet that fails to load is left out, and the error names each one; the error is nil when every pet loads, and says so when none does.
func Load(fsys fs.FS, base string) ([]*Pack, error) {
	manifests, err := fs.Glob(fsys, "*/pet.json")
	if err != nil {
		return nil, err
	}
	var packs []*Pack
	var errs []error
	for _, manifest := range manifests {
		dir := path.Dir(manifest)
		p, err := loadPack(fsys, dir)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", dir, err))
			continue
		}
		p.Base = base + "/" + dir
		packs = append(packs, p)
	}
	if len(packs) == 0 {
		errs = append(errs, errors.New("no pets"))
	}
	return packs, errors.Join(errs...)
}

// LoadDir reads every pet in the directory dir, as Load does, playing their clips from its files.
func LoadDir(dir string) ([]*Pack, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	return Load(os.DirFS(abs), filepath.ToSlash(abs))
}

func loadPack(fsys fs.FS, dir string) (*Pack, error) {
	p := &Pack{ID: dir}
	data, err := fs.ReadFile(fsys, dir+"/pet.json")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, p); err != nil {
		return nil, fmt.Errorf("pet.json: %w", err)
	}
	rules, err := fs.ReadFile(fsys, dir+"/animations.json")
	if err != nil {
		return nil, err
	}
	if p.Policy, err = NewPolicy(rules); err != nil {
		return nil, err
	}
	if err := p.Check(); err != nil {
		return nil, err
	}
	return p, nil
}

// Check reports a manifest the pet cannot run with.
func (p *Pack) Check() error {
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
