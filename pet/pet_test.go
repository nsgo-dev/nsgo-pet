package pet

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/nsgotest"
)

// TestMain gives the platform the main thread under nsgo test -target.
func TestMain(m *testing.M) { nsgotest.Main(m) }

var linux = nsgotest.Environment(core.Environment{OS: core.OSLinux, Idiom: core.IdiomDesktop})

func cat(t *testing.T) *Pack {
	t.Helper()
	packs, err := Load(os.DirFS("testdata/pets"), "asset:pets")
	for _, p := range packs {
		if p.ID == "cat" {
			return p
		}
	}
	t.Fatalf("no cat pet: %v", err)
	return nil
}

func source(app *nsgotest.App, id string) string { return app.FindMounted("id:" + id).Video.Source }

// A pet that fails to load is left out and named in the error; the others load.
func TestLoadLeavesOutABrokenPet(t *testing.T) {
	packs, err := Load(os.DirFS("testdata/pets"), "asset:pets")
	if len(packs) != 1 || packs[0].ID != "cat" || packs[0].Base != "asset:pets/cat" {
		t.Fatalf("packs %+v, want the cat alone", packs)
	}
	if err == nil || !strings.Contains(err.Error(), "broken: ") {
		t.Fatalf("error %v, want the broken pet named", err)
	}
}

func TestLoadFindsNoPets(t *testing.T) {
	if packs, err := Load(os.DirFS(t.TempDir()), "asset:pets"); len(packs) != 0 || err == nil {
		t.Fatalf("an empty directory loaded %v, %v", packs, err)
	}
}

// A pet loaded from a directory plays its clips from their files.
func TestLoadDirPlaysFiles(t *testing.T) {
	packs, _ := LoadDir("testdata/pets")
	abs, _ := filepath.Abs("testdata/pets/cat")
	if len(packs) != 1 || packs[0].Base != filepath.ToSlash(abs) {
		t.Fatalf("packs %+v, want the cat from %s", packs, abs)
	}
}

// show starts a pet in the host's view, beside the host's own content.
func show(t *testing.T, pack *Pack, opts Options) (*nsgotest.App, *Pet) {
	t.Helper()
	var p *Pet
	app := nsgotest.Start(t, func(cx *core.Scope) core.Node {
		p = New(cx, pack, opts)
		cx.Post(p.Start)
		return core.View{Style: core.Style{Direction: core.Row}, Children: []core.Node{core.Text{Content: "host"}, p.View()}}
	}, linux)
	return app, p
}

// The pet greets the user, then idles in the other player once the greeting ends.
func TestGreetsThenIdles(t *testing.T) {
	app, _ := show(t, cat(t), Options{})
	if got := source(app, "slot1"); got != "asset:pets/cat/webm/wave.webm" {
		t.Fatalf("first clip = %q, want the greeting", got)
	}
	app.Play("id:slot1", core.VideoInfo{Duration: time.Second})
	app.Progress("id:slot1", time.Second, time.Second)
	if got := source(app, "slot0"); got != "asset:pets/cat/webm/idle.webm" {
		t.Fatalf("after the greeting = %q, want the idle", got)
	}
}

// A host's state weighs the clips: the working clip plays only while the host says the user works.
func TestHostContext(t *testing.T) {
	working := false
	app, p := show(t, cat(t), Options{Context: func() Context { return Context{Working: working} }})
	app.Play("id:slot1", core.VideoInfo{Duration: time.Second})
	p.Act()
	app.Settle()
	if got := source(app, "slot0"); got == "asset:pets/cat/webm/work.webm" {
		t.Fatal("the working clip played while the user did not work")
	}
	working = true
	p.Act()
	app.Settle()
	if got := source(app, "slot0"); got != "asset:pets/cat/webm/work.webm" {
		t.Fatalf("while working the pet played %q", got)
	}
}

// A host plays a clip by ID, even one no pool chooses; an unknown ID plays nothing.
func TestPlayByID(t *testing.T) {
	app, p := show(t, cat(t), Options{})
	app.Play("id:slot1", core.VideoInfo{Duration: time.Second})
	played := p.Play("done")
	app.Settle()
	if !played || source(app, "slot0") != "asset:pets/cat/webm/done.webm" {
		t.Fatalf("Play(done) shows %q", source(app, "slot0"))
	}
	if p.Play("nothing") {
		t.Fatal("Play of a clip the pack lacks reported true")
	}
}

// A press on the pet's body in the host's view plays a click reaction once the greeting is over.
func TestPressPlaysAClickReaction(t *testing.T) {
	app, _ := show(t, cat(t), Options{})
	app.Play("id:slot1", core.VideoInfo{Duration: time.Second})
	app.Progress("id:slot1", time.Second, time.Second)
	app.Play("id:slot0", core.VideoInfo{Duration: time.Second})
	app.Press("id:body")
	if got := source(app, "slot1"); got != "asset:pets/cat/webm/purr.webm" {
		t.Fatalf("after a press = %q, want the click reaction", got)
	}
}

// Leave plays the exit clip and hands the pet back to the host once, as the clip ends, without quitting the app.
func TestLeaveRunsDoneOnce(t *testing.T) {
	app, p := show(t, cat(t), Options{})
	app.Play("id:slot1", core.VideoInfo{Duration: time.Second})
	app.Progress("id:slot1", time.Second, time.Second)
	app.Play("id:slot0", core.VideoInfo{Duration: time.Second})
	left := 0
	p.Leave(func() { left++ })
	app.Play("id:slot1", core.VideoInfo{Duration: time.Second})
	app.Progress("id:slot1", time.Second, time.Second)
	app.Advance(10 * time.Second)
	if left != 1 {
		t.Fatalf("done ran %d times, want once", left)
	}
	if p.Play("idle") {
		t.Fatal("a pet that left played a clip")
	}
}

// Without an end, the pet leaves eight seconds after Leave.
func TestLeaveWithoutAnEnd(t *testing.T) {
	app, p := show(t, cat(t), Options{})
	left := false
	p.Leave(func() { left = true })
	app.Advance(7 * time.Second)
	if left {
		t.Fatal("left before eight seconds")
	}
	app.Advance(time.Second)
	if !left {
		t.Fatal("still there after eight seconds")
	}
}

// A pet in its own window gets the window's options for its size.
func TestWindowOptions(t *testing.T) {
	o := WindowOptions(cat(t))
	if o.Width != 120 || o.Height != 100 || !o.Transparent || !o.NoShadow || !o.SkipTaskbar {
		t.Fatalf("options %+v", o)
	}
}

func TestCodingApps(t *testing.T) {
	for id, want := range map[string]bool{
		"com.microsoft.VSCode": true, "com.jetbrains.goland": true, "com.mitchellh.ghostty": true,
		`C:\Program Files\Microsoft VS Code\Code.exe`: true, "com.apple.Safari": false, "": false,
	} {
		if got := isCodingApp(id); got != want {
			t.Errorf("isCodingApp(%q) = %v", id, got)
		}
	}
}
