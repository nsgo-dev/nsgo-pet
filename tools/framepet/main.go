// Command framepet turns a pet drawn as PNG frames into an nsgo-pet folder: a JSON spec lists each clip's frames in order, how fast they play and when the clip plays, and framepet writes every clip as WebM (VP9 with alpha) and MOV (HEVC with alpha) with pet.json and animations.json beside them.
//
//	go run ./tools/framepet -spec tools/framepet/pets/mochi.json -frames <dir> -out resources/pets/mochi
//	go run ./tools/framepet -spec tools/framepet/pets/mochi.json -list
//
// -list prints the frame files the spec uses, relative to -frames, one per line, so a fetch script downloads only those.
//
// A frame is a PNG path relative to -frames. "sheet.png#3" is the fourth cell of a sheet drawn as one row of square cells, and a suffix "*4" holds the frame for 4 ticks of 1/fps instead of 1. Frames of another size than canvas are scaled to it first, nearest neighbor, as pixel art drawn on a smaller grid is. The clips are cropped to the union of every frame's opaque pixels plus margin, so the window holds only the art, and scaled by pixelScale.
//
// It needs ffmpeg with libvpx-vp9, and hevc_videotoolbox (macOS) for the MOV files.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// spec is the JSON a pet is made from.
type spec struct {
	Name    string `json:"name"`
	Group   string `json:"group"`
	Source  string `json:"source"`
	License string `json:"license"`
	// Canvas is the frame size in pixels; frames of another size are scaled to it.
	Canvas [2]int `json:"canvas"`
	// Margin is how many canvas pixels are kept around the art when cropping.
	Margin int `json:"margin"`
	// PixelScale is the clips' size against the cropped canvas, 1 when left out.
	PixelScale float64 `json:"pixelScale"`
	// PixelArt scales with nearest neighbor instead of Lanczos.
	PixelArt bool `json:"pixelArt"`
	// Height is the window's height in points; its width follows the clips' aspect.
	Height       int `json:"height"`
	EventSeconds int `json:"eventSeconds"`
	// Body is the pet's body in percent of the frame; left out, it is the idle clip's art.
	Body  *body     `json:"body"`
	Play  clipNames `json:"play"`
	Clips []clip    `json:"clips"`
}

// clip is one clip of the pet: its frames, played loops times at fps, and the rules it is chosen by.
type clip struct {
	Name   string   `json:"name"`
	FPS    float64  `json:"fps"`
	Loops  int      `json:"loops"`
	Frames []string `json:"frames"`
	Rules  []entry  `json:"rules"`
}

// entry is one rule of animations.json, in Big Blue Fish's format. In the spec Id defaults to <clip>-<pool> and File is the clip's.
type entry struct {
	ID             string   `json:"Id"`
	File           string   `json:"File"`
	Pool           string   `json:"Pool"`
	Weight         float64  `json:"Weight"`
	NightOnly      bool     `json:"NightOnly,omitempty"`
	SeasonOnly     string   `json:"SeasonOnly,omitempty"`
	SummerWeight   *float64 `json:"SummerWeight,omitempty"`
	WinterWeight   *float64 `json:"WinterWeight,omitempty"`
	MorningWeight  *float64 `json:"MorningWeight,omitempty"`
	Meal           string   `json:"Meal,omitempty"`
	MealBoost      float64  `json:"MealBoost,omitempty"`
	Festival       string   `json:"Festival,omitempty"`
	RadiusDays     float64  `json:"RadiusDays,omitempty"`
	FestivalMode   string   `json:"FestivalMode,omitempty"`
	FestivalWeight float64  `json:"FestivalWeight,omitempty"`
	WorkKind       string   `json:"WorkKind,omitempty"`
	WorkWeight     float64  `json:"WorkWeight,omitempty"`
	MoveKind       string   `json:"MoveKind,omitempty"`
	HourParity     int      `json:"HourParity,omitempty"`
}

// manifest is pet.json.
type manifest struct {
	Name         string    `json:"name"`
	Group        string    `json:"group,omitempty"`
	Width        int       `json:"width"`
	Height       int       `json:"height"`
	Body         body      `json:"body"`
	EventSeconds int       `json:"eventSeconds"`
	Clips        clipNames `json:"clips"`
	Source       string    `json:"source"`
	License      string    `json:"license"`
}

type body struct {
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// clipNames are the rule IDs the pet plays by name.
type clipNames struct {
	Greeting  string `json:"greeting"`
	Exit      string `json:"exit"`
	Drag      string `json:"drag"`
	Idle      string `json:"idle"`
	Daily     string `json:"daily,omitempty"`
	DailyHour int    `json:"dailyHour,omitempty"`
}

// frame is one entry of a clip's frame list.
type frame struct {
	file  string // the PNG, relative to -frames
	cell  int    // the cell of a one-row sheet, -1 for a whole image
	ticks int    // how many 1/fps ticks it shows
}

func main() {
	specPath := flag.String("spec", "", "the pet's spec, JSON")
	frames := flag.String("frames", "", "the directory the spec's frame paths are relative to")
	out := flag.String("out", "", "the pet's folder, resources/pets/<id>")
	list := flag.Bool("list", false, "print the frame files the spec uses and exit")
	flag.Parse()
	if *specPath == "" || !*list && (*frames == "" || *out == "") {
		flag.Usage()
		os.Exit(2)
	}
	s, err := readSpec(*specPath)
	if err != nil {
		log.Fatal(err)
	}
	if *list {
		seen := map[string]bool{}
		for _, c := range s.Clips {
			for _, ref := range c.Frames {
				f, _ := parseFrame(ref)
				if !seen[f.file] {
					seen[f.file] = true
					fmt.Println(f.file)
				}
			}
		}
		return
	}
	tmp, err := os.MkdirTemp("", "framepet")
	if err != nil {
		log.Fatal(err)
	}
	err = build(s, *frames, *out, tmp)
	os.RemoveAll(tmp)
	if err != nil {
		log.Fatal(err)
	}
}

// readSpec reads and checks a spec, filling in its defaults.
func readSpec(path string) (*spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	s := &spec{}
	if err := dec.Decode(s); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if s.Name == "" || s.License == "" || s.Height <= 0 || s.Canvas[0] <= 0 || s.Canvas[1] <= 0 {
		return nil, fmt.Errorf("%s: name, license, height and canvas are required", path)
	}
	if s.PixelScale == 0 {
		s.PixelScale = 1
	}
	if s.EventSeconds == 0 {
		s.EventSeconds = 40
	}
	ids := map[string]bool{}
	for i := range s.Clips {
		c := &s.Clips[i]
		if c.Name == "" || c.FPS <= 0 || len(c.Frames) == 0 || len(c.Rules) == 0 {
			return nil, fmt.Errorf("%s: clip %q needs a name, fps, frames and rules", path, c.Name)
		}
		if c.Loops == 0 {
			c.Loops = 1
		}
		for _, ref := range c.Frames {
			if _, err := parseFrame(ref); err != nil {
				return nil, fmt.Errorf("%s: clip %s: %w", path, c.Name, err)
			}
		}
		for j := range c.Rules {
			r := &c.Rules[j]
			if r.Pool == "" {
				return nil, fmt.Errorf("%s: clip %s: a rule without a pool", path, c.Name)
			}
			if r.ID == "" {
				r.ID = c.Name + "-" + strings.ToLower(r.Pool)
			}
			r.File = c.Name + ".webm"
			if ids[r.ID] {
				return nil, fmt.Errorf("%s: rule %q twice", path, r.ID)
			}
			ids[r.ID] = true
		}
	}
	for _, id := range []string{s.Play.Greeting, s.Play.Exit, s.Play.Drag, s.Play.Idle} {
		if !ids[id] {
			return nil, fmt.Errorf("%s: play names rule %q, which no clip has", path, id)
		}
	}
	if s.Play.Daily != "" && !ids[s.Play.Daily] {
		return nil, fmt.Errorf("%s: play names daily rule %q, which no clip has", path, s.Play.Daily)
	}
	return s, nil
}

// parseFrame reads "file.png", "sheet.png#3" and either with a "*4" hold.
func parseFrame(ref string) (frame, error) {
	f := frame{file: ref, cell: -1, ticks: 1}
	if i := strings.LastIndexByte(f.file, '*'); i >= 0 {
		n, err := strconv.Atoi(f.file[i+1:])
		if err != nil || n < 1 {
			return f, fmt.Errorf("frame %q: bad hold", ref)
		}
		f.file, f.ticks = f.file[:i], n
	}
	if i := strings.LastIndexByte(f.file, '#'); i >= 0 {
		n, err := strconv.Atoi(f.file[i+1:])
		if err != nil || n < 0 {
			return f, fmt.Errorf("frame %q: bad cell", ref)
		}
		f.file, f.cell = f.file[:i], n
	}
	if !strings.HasSuffix(f.file, ".png") {
		return f, fmt.Errorf("frame %q: not a PNG", ref)
	}
	return f, nil
}

// build makes every clip of s from the frames in dir, writes them and the pet's JSON under out, and works in tmp.
func build(s *spec, dir, out, tmp string) error {
	images := map[string]*image.NRGBA{}
	var union image.Rectangle
	for _, c := range s.Clips {
		for _, ref := range c.Frames {
			f, _ := parseFrame(ref)
			key := cellKey(f)
			if images[key] != nil {
				continue
			}
			img, err := loadCell(filepath.Join(dir, f.file), f.cell, s.Canvas)
			if err != nil {
				return err
			}
			images[key] = img
			union = union.Union(opaqueBounds(img))
		}
	}
	if union.Empty() {
		return fmt.Errorf("every frame is transparent")
	}
	crop := union.Inset(-s.Margin).Intersect(image.Rect(0, 0, s.Canvas[0], s.Canvas[1]))
	width, height := even(float64(crop.Dx())*s.PixelScale), even(float64(crop.Dy())*s.PixelScale)
	fmt.Printf("art %v of the %dx%d canvas, clips %dx%d pixels\n", crop, s.Canvas[0], s.Canvas[1], width, height)

	// Every distinct frame, cropped, once.
	files := map[string]string{}
	for key, img := range images {
		path := filepath.Join(tmp, fmt.Sprintf("%d.png", len(files)))
		if err := writePNG(path, img.SubImage(crop)); err != nil {
			return err
		}
		files[key] = path
	}
	for _, d := range []string{"webm", "mov"} {
		if err := os.MkdirAll(filepath.Join(out, d), 0o755); err != nil {
			return err
		}
	}
	var entries []entry
	for _, c := range s.Clips {
		if err := makeClip(s, c, files, width, height, out, tmp); err != nil {
			return fmt.Errorf("clip %s: %w", c.Name, err)
		}
		entries = append(entries, c.Rules...)
	}

	b := body{}
	if s.Body != nil {
		b = *s.Body
	} else {
		var art image.Rectangle
		for _, c := range s.Clips {
			if !hasRule(c, s.Play.Idle) {
				continue
			}
			for _, ref := range c.Frames {
				f, _ := parseFrame(ref)
				art = art.Union(opaqueBounds(images[cellKey(f)]))
			}
		}
		art = art.Intersect(crop).Sub(crop.Min)
		pct := func(v, of int) float64 { return math.Round(float64(v)*1000/float64(of)) / 10 }
		b = body{Left: pct(art.Min.X, crop.Dx()), Top: pct(art.Min.Y, crop.Dy()), Width: pct(art.Dx(), crop.Dx()), Height: pct(art.Dy(), crop.Dy())}
	}
	m := manifest{
		Name: s.Name, Group: s.Group,
		Width: int(math.Round(float64(s.Height) * float64(crop.Dx()) / float64(crop.Dy()))), Height: s.Height,
		Body: b, EventSeconds: s.EventSeconds, Clips: s.Play, Source: s.Source, License: s.License,
	}
	if err := writeJSON(filepath.Join(out, "pet.json"), m); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(out, "animations.json"), entries); err != nil {
		return err
	}
	fmt.Printf("%s: %d clips, %d rules, window %dx%d points\n", out, len(s.Clips), len(entries), m.Width, m.Height)
	return nil
}

func hasRule(c clip, id string) bool {
	for _, r := range c.Rules {
		if r.ID == id {
			return true
		}
	}
	return false
}

func cellKey(f frame) string { return f.file + "#" + strconv.Itoa(f.cell) }

// loadCell reads a PNG, takes cell of it when cell is not -1, and scales it to canvas, nearest neighbor, when its size differs.
func loadCell(path string, cell int, canvas [2]int) (*image.NRGBA, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	src, err := png.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	bounds := src.Bounds()
	if cell >= 0 {
		side := bounds.Dy()
		if (cell+1)*side > bounds.Dx() {
			return nil, fmt.Errorf("%s: no cell %d in a %dx%d sheet of square cells", path, cell, bounds.Dx(), bounds.Dy())
		}
		bounds = image.Rect(bounds.Min.X+cell*side, bounds.Min.Y, bounds.Min.X+(cell+1)*side, bounds.Max.Y)
	}
	img := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(img, img.Bounds(), src, bounds.Min, draw.Src)
	if img.Bounds().Dx() == canvas[0] && img.Bounds().Dy() == canvas[1] {
		return img, nil
	}
	scaled := image.NewNRGBA(image.Rect(0, 0, canvas[0], canvas[1]))
	for y := range canvas[1] {
		for x := range canvas[0] {
			scaled.SetNRGBA(x, y, img.NRGBAAt(x*img.Bounds().Dx()/canvas[0], y*img.Bounds().Dy()/canvas[1]))
		}
	}
	return scaled, nil
}

// opaqueBounds is the smallest rectangle holding every pixel of img that is not fully transparent.
func opaqueBounds(img *image.NRGBA) image.Rectangle {
	var r image.Rectangle
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.Pix[img.PixOffset(x, y)+3] != 0 {
				r = r.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	return r
}

// even rounds v to an even number of pixels, as 4:2:0 video needs.
func even(v float64) int { return max(2, int(math.Round(v/2))*2) }

// makeClip plays c's frames c.Loops times at c.FPS, scaled to width by height, and writes <name>.webm and <name>.mov under out.
func makeClip(s *spec, c clip, files map[string]string, width, height int, out, tmp string) error {
	var list strings.Builder
	var last string
	for range c.Loops {
		for _, ref := range c.Frames {
			f, _ := parseFrame(ref)
			last = files[cellKey(f)]
			fmt.Fprintf(&list, "file '%s'\nduration %g\n", last, float64(f.ticks)/c.FPS)
		}
	}
	// The concat demuxer holds the last frame only when it is listed once more.
	fmt.Fprintf(&list, "file '%s'\n", last)
	concat := filepath.Join(tmp, c.Name+".txt")
	if err := os.WriteFile(concat, []byte(list.String()), 0o644); err != nil {
		return err
	}
	flags := "lanczos"
	if s.PixelArt {
		flags = "neighbor"
	}
	scale := fmt.Sprintf("scale=%d:%d:flags=%s", width, height, flags)
	fps := strconv.FormatFloat(c.FPS, 'f', -1, 64)
	webm := filepath.Join(out, "webm", c.Name+".webm")
	if err := ffmpeg("-f", "concat", "-safe", "0", "-i", concat, "-vf", scale, "-r", fps, "-c:v", "libvpx-vp9", "-pix_fmt", "yuva420p", "-b:v", "0", "-crf", "30", "-auto-alt-ref", "0", webm); err != nil {
		return err
	}
	mov := filepath.Join(out, "mov", c.Name+".mov")
	return ffmpeg("-c:v", "libvpx-vp9", "-i", webm, "-an", "-c:v", "hevc_videotoolbox", "-alpha_quality", "0.75", "-q:v", "65", "-tag:v", "hvc1", "-pix_fmt", "bgra", mov)
}

func ffmpeg(args ...string) error {
	cmd := exec.Command("ffmpeg", append([]string{"-v", "error", "-y", "-nostdin"}, args...)...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func writePNG(path string, img image.Image) error {
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(w, img); err != nil {
		w.Close()
		return err
	}
	return w.Close()
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
