// Command codexpet turns a pet drawn as a Codex sprite sheet into an nsgo-pet folder: each row of the sheet becomes a clip in WebM (VP9 with alpha) and MOV (HEVC with alpha), and pet.json and animations.json say when each plays.
//
// A Codex sheet is eight columns of 192 by 208 pixel cells. Version 1 has nine rows and version 2 eleven; the first nine rows are the same in both: idle, running right, running left, waving, jumping, failed, waiting, working and done, using their first 6, 8, 8, 4, 5, 8, 6, 6 and 6 cells. The last two rows of version 2 look in sixteen directions and are not used.
//
//	go run ./tools/codexpet -sheet clover.png -out resources/pets/playpets-clover -name Clover -group PlayPets -source https://github.com/gaerae/playpets -license MIT
//
// -summer and -winter take the same pet's outfits for those seasons: their idle, events and clicks play then, the base sheet's the rest of the year. It needs ffmpeg with libvpx-vp9, and hevc_videotoolbox (macOS) for the MOV files.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Cell is one cell of a Codex sheet in pixels.
const (
	cellWidth  = 192
	cellHeight = 208
)

// row is one animation of a Codex sheet.
type row struct {
	name   string
	index  int
	frames int
}

// rows are the animations of both versions of the format, in sheet order.
var rows = []row{
	{"idle", 0, 6}, {"right", 1, 8}, {"left", 2, 8}, {"wave", 3, 4}, {"jump", 4, 5},
	{"error", 5, 8}, {"waiting", 6, 6}, {"working", 7, 6}, {"done", 8, 6},
}

// clip is a row played as a clip: how many times its frames repeat, and the pools it is chosen from.
type clip struct {
	row   string
	loops int
	rules []rule
}

// rule is an animations.json entry for a clip.
type rule struct {
	pool   string
	weight float64
}

// clips say which rows the pet plays and when: it idles and waits, works, finishes and jumps between, waves and jumps when clicked, and dangles in the failed pose while dragged.
var clips = []clip{
	{"idle", 4, []rule{{"Idle", 100}}},
	{"waiting", 3, []rule{{"Idle", 30}, {"Event", 60}}},
	{"working", 3, []rule{{"Event", 100}}},
	{"done", 3, []rule{{"Event", 60}}},
	{"jump", 2, []rule{{"Event", 40}, {"Click", 60}}},
	{"wave", 3, []rule{{"Click", 100}, {"Greeting", 100}}},
	{"error", 2, []rule{{"Drag", 100}}},
}

// framesPerSecond is how fast a Codex pet's frames play.
const framesPerSecond = 8

// entry is one rule of animations.json, in Big Blue Fish's format.
type entry struct {
	ID         string  `json:"Id"`
	File       string  `json:"File"`
	Pool       string  `json:"Pool"`
	Weight     float64 `json:"Weight"`
	SeasonOnly string  `json:"SeasonOnly,omitempty"`
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
	Left   float32 `json:"left"`
	Top    float32 `json:"top"`
	Width  float32 `json:"width"`
	Height float32 `json:"height"`
}

type clipNames struct {
	Greeting string `json:"greeting"`
	Exit     string `json:"exit"`
	Drag     string `json:"drag"`
	Idle     string `json:"idle"`
}

func main() {
	sheet := flag.String("sheet", "", "the pet's Codex sprite sheet, PNG or WebP")
	summer := flag.String("summer", "", "the sheet of its summer outfit")
	winter := flag.String("winter", "", "the sheet of its winter outfit")
	out := flag.String("out", "", "the pet's folder, resources/pets/<id>")
	name := flag.String("name", "", "the pet's name in the menu")
	group := flag.String("group", "", "the submenu the pet is listed in")
	source := flag.String("source", "", "where the pet comes from")
	license := flag.String("license", "", "the pet's art license")
	flag.Parse()
	if *sheet == "" || *out == "" || *name == "" || *license == "" {
		flag.Usage()
		os.Exit(2)
	}
	for _, dir := range []string{"webm", "mov"} {
		if err := os.MkdirAll(filepath.Join(*out, dir), 0o755); err != nil {
			log.Fatal(err)
		}
	}
	tmp, err := os.MkdirTemp("", "codexpet")
	if err != nil {
		log.Fatal(err)
	}
	entries, err := convert(tmp, *sheet, *summer, *winter, *out)
	os.RemoveAll(tmp)
	if err != nil {
		log.Fatal(err)
	}
	m := manifest{
		Name: *name, Group: *group, Width: 144, Height: 156,
		Body:         body{Left: 22, Top: 12, Width: 56, Height: 84},
		EventSeconds: 30,
		Clips:        clipNames{Greeting: "wave-greeting", Exit: "wave-greeting", Drag: "error-drag", Idle: "idle-idle"},
		Source:       *source, License: *license,
	}
	write(filepath.Join(*out, "pet.json"), m)
	write(filepath.Join(*out, "animations.json"), entries)
	fmt.Printf("%s: %d clips\n", *out, len(entries))
}

// convert makes the clips of every outfit in out, working in tmp, and returns their rules.
func convert(tmp, sheet, summer, winter, out string) ([]entry, error) {
	var entries []entry
	outfits := []struct{ sheet, suffix string }{{sheet, ""}, {summer, "-summer"}, {winter, "-winter"}}
	for _, o := range outfits {
		if o.sheet == "" {
			continue
		}
		for _, c := range clips {
			file := c.row + o.suffix
			if err := makeClip(o.sheet, c, file, out, tmp); err != nil {
				return nil, fmt.Errorf("%s %s: %w", o.sheet, c.row, err)
			}
			entries = append(entries, rulesFor(c, file, o.suffix, summer != "", winter != "")...)
		}
	}
	return entries, nil
}

// rulesFor are the rules of clip c in file: the base outfit's play in the seasons without an outfit of their own, each outfit's in its season. Greeting and Drag are the base outfit's all year.
func rulesFor(c clip, file, suffix string, hasSummer, hasWinter bool) []entry {
	var out []entry
	for _, r := range c.rules {
		id := c.row + suffix + "-" + strings.ToLower(r.pool)
		if r.pool == "Greeting" || r.pool == "Drag" {
			if suffix == "" {
				out = append(out, entry{ID: id, File: file + ".webm", Pool: r.pool, Weight: r.weight})
			}
			continue
		}
		switch suffix {
		case "-summer":
			out = append(out, entry{ID: id, File: file + ".webm", Pool: r.pool, Weight: r.weight, SeasonOnly: "Summer"})
		case "-winter":
			out = append(out, entry{ID: id, File: file + ".webm", Pool: r.pool, Weight: r.weight, SeasonOnly: "Winter"})
		default:
			if !hasSummer && !hasWinter {
				out = append(out, entry{ID: id, File: file + ".webm", Pool: r.pool, Weight: r.weight})
				continue
			}
			for _, season := range []string{"Spring", "Summer", "Autumn", "Winter"} {
				if season == "Summer" && hasSummer || season == "Winter" && hasWinter {
					continue
				}
				out = append(out, entry{ID: id + "-" + strings.ToLower(season), File: file + ".webm", Pool: r.pool, Weight: r.weight, SeasonOnly: season})
			}
			// The fallback idle the pet plays by name, never chosen by itself.
			if r.pool == "Idle" && c.row == "idle" {
				out = append(out, entry{ID: id, File: file + ".webm", Pool: "Disabled", Weight: 0})
			}
		}
	}
	return out
}

// makeClip cuts row c.row's frames out of sheet, plays them c.loops times at framesPerSecond, and writes file.webm and file.mov under out.
func makeClip(sheet string, c clip, file, out, tmp string) error {
	var r row
	for _, x := range rows {
		if x.name == c.row {
			r = x
		}
	}
	dir := filepath.Join(tmp, file)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	var list strings.Builder
	for i := range r.frames {
		png := filepath.Join(dir, fmt.Sprintf("%d.png", i))
		crop := fmt.Sprintf("crop=%d:%d:%d:%d", cellWidth, cellHeight, i*cellWidth, r.index*cellHeight)
		if err := ffmpeg("-i", sheet, "-vf", crop, "-frames:v", "1", "-pix_fmt", "rgba", png); err != nil {
			return err
		}
	}
	for range c.loops {
		for i := range r.frames {
			fmt.Fprintf(&list, "file '%d.png'\nduration %g\n", i, 1.0/framesPerSecond)
		}
	}
	// The concat demuxer holds the last frame only when it is listed once more.
	fmt.Fprintf(&list, "file '%d.png'\n", r.frames-1)
	concat := filepath.Join(dir, "frames.txt")
	if err := os.WriteFile(concat, []byte(list.String()), 0o644); err != nil {
		return err
	}
	webm := filepath.Join(out, "webm", file+".webm")
	if err := ffmpeg("-f", "concat", "-safe", "0", "-i", concat, "-r", fmt.Sprint(framesPerSecond), "-c:v", "libvpx-vp9", "-pix_fmt", "yuva420p", "-b:v", "0", "-crf", "30", "-auto-alt-ref", "0", webm); err != nil {
		return err
	}
	mov := filepath.Join(out, "mov", file+".mov")
	return ffmpeg("-c:v", "libvpx-vp9", "-i", webm, "-an", "-c:v", "hevc_videotoolbox", "-alpha_quality", "0.75", "-q:v", "65", "-tag:v", "hvc1", "-pix_fmt", "bgra", mov)
}

func ffmpeg(args ...string) error {
	cmd := exec.Command("ffmpeg", append([]string{"-v", "error", "-y", "-nostdin"}, args...)...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func write(path string, v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		log.Fatal(err)
	}
}
