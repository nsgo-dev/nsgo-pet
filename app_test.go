package main

import (
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/nsgo-dev/nsgo-pet/pet"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/nsgotest"
)

// TestMain gives the platform the main thread under nsgo test -target.
func TestMain(m *testing.M) { nsgotest.Main(m) }

var macOS = nsgotest.Environment(core.Environment{OS: core.OSMacOS, Idiom: core.IdiomDesktop})

func packs(t *testing.T) []*pet.Pack {
	t.Helper()
	ps, err := LoadPacks()
	if err != nil {
		t.Fatal(err)
	}
	return ps
}

func fish(t *testing.T) *pet.Pack {
	t.Helper()
	for _, p := range packs(t) {
		if p.ID == "bigbluefish" {
			return p
		}
	}
	t.Fatal("no bigbluefish pet")
	return nil
}

func source(app *nsgotest.App, id string) string { return app.FindMounted("id:" + id).Video.Source }

// NSGo-chan, the default pet, greets the user first, then goes on with an idle clip in the other player, which shows once it has a picture.
func TestGreetsThenIdles(t *testing.T) {
	app := nsgotest.Start(t, App(packs(t)), macOS)
	if got := source(app, "slot1"); got != "asset:pets/nsgo-chan/mov/greeting.mov" {
		t.Fatalf("first clip = %q, want the greeting", got)
	}
	app.Play("id:slot1", core.VideoInfo{Duration: 5 * time.Second})
	app.Progress("id:slot1", 5*time.Second, 5*time.Second)
	idle := source(app, "slot0")
	if !strings.HasPrefix(idle, "asset:pets/nsgo-chan/mov/") {
		t.Fatalf("after the greeting the other player has %q", idle)
	}
	name := strings.TrimSuffix(strings.TrimPrefix(idle, "asset:pets/nsgo-chan/mov/"), ".mov")
	if r := packs(t)[0].Policy.Find(name); r.Pool != "Idle" {
		t.Fatalf("after the greeting NSGo-chan plays %q from pool %q, want an idle clip", name, r.Pool)
	}
}

// Linux and Windows play the WebM clips.
func TestOtherPlatformsPlayWebM(t *testing.T) {
	app := nsgotest.Start(t, App(packs(t)), nsgotest.Environment(core.Environment{OS: core.OSLinux, Idiom: core.IdiomDesktop}))
	if got := source(app, "slot1"); got != "asset:pets/nsgo-chan/webm/greeting.webm" {
		t.Fatalf("first clip = %q", got)
	}
}

// Every pet's clips ship in both forms.
func TestEveryClipShips(t *testing.T) {
	if p := fish(t); len(p.Policy.Rules) != 106 {
		t.Fatalf("%d rules, want 106", len(p.Policy.Rules))
	}
	for _, pk := range packs(t) {
		dir := filepath.Join("resources/pets", pk.ID)
		for _, r := range pk.Policy.Rules {
			name := strings.TrimSuffix(r.File, ".webm")
			for _, f := range []string{filepath.Join(dir, "webm", r.File), filepath.Join(dir, "mov", name+".mov")} {
				if _, err := os.Stat(f); err != nil {
					t.Errorf("%s %s: %v", pk.ID, r.ID, err)
				}
			}
		}
	}
}

// A manifest naming a clip its rules lack is refused.
func TestPackNamesOnlyItsClips(t *testing.T) {
	p := *fish(t)
	p.Clips.Drag = "nothing"
	if err := p.Check(); err == nil {
		t.Fatal("a manifest naming a missing drag clip passed")
	}
}

func at(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04", s, time.Local)
	if err != nil {
		panic(err)
	}
	return t
}

func TestWeights(t *testing.T) {
	p := fish(t).Policy
	for _, c := range []struct {
		name string
		id   string
		ctx  pet.Context
		want float64
	}{
		{"sleeping only at night", "原地小憩沉眠", pet.Context{Now: at("2026-11-17 12:00")}, 0},
		{"sleeping at night", "原地小憩沉眠", pet.Context{Now: at("2026-11-17 23:30")}, 40},
		{"disabled", "余额-分文不剩", pet.Context{Now: at("2026-11-17 12:00")}, 0},
		{"running on odd hours", "原地左转奔跑", pet.Context{Now: at("2026-11-17 13:30")}, 200},
		{"no running on even hours", "原地左转奔跑", pet.Context{Now: at("2026-11-17 14:30")}, 0},
		{"no running at the hour's edge", "原地左转奔跑", pet.Context{Now: at("2026-11-17 13:57")}, 0},
		{"dinner near six", "吃午餐", pet.Context{Now: at("2026-11-17 18:30")}, 1000},
		{"no dinner boost at three", "吃午餐", pet.Context{Now: at("2026-11-17 15:00")}, 100},
		{"ice cream only in summer", "吃冰淇淋融化", pet.Context{Now: at("2026-11-17 15:00")}, 0},
		{"coding", "写代码", pet.Context{Now: at("2026-11-17 15:00"), Coding: true}, 1000},
		{"not coding", "写代码", pet.Context{Now: at("2026-11-17 15:00")}, 100},
	} {
		if got := p.Weight(p.Find(c.id), c.ctx); got != c.want {
			t.Errorf("%s: weight of %s = %v, want %v", c.name, c.id, got, c.want)
		}
	}
}

// A clip of a Gregorian festival grows toward its festival weight as the day nears.
func TestFestivalWeightPeaksOnTheDay(t *testing.T) {
	p := fish(t).Policy
	var rule pet.Rule
	for _, r := range p.Rules {
		if r.Festival == "Christmas" {
			rule = r
			break
		}
	}
	if rule.ID == "" {
		t.Fatal("no Christmas clip in the catalog")
	}
	far := p.Weight(rule, pet.Context{Now: at("2026-07-01 15:00")})
	near := p.Weight(rule, pet.Context{Now: at("2026-12-22 15:00")})
	day := p.Weight(rule, pet.Context{Now: at("2026-12-25 15:00")})
	if !(far < near && near < day) || day != rule.FestivalWeight {
		t.Fatalf("weights far %v, near %v, on the day %v (festival weight %v)", far, near, day, rule.FestivalWeight)
	}
}

func TestChooseKeepsToItsPool(t *testing.T) {
	p := fish(t).Policy
	rnd := rand.New(rand.NewPCG(1, 2))
	ctx := pet.Context{Now: at("2026-11-17 15:00")}
	for range 200 {
		r, ok := p.Choose("Click", ctx, rnd)
		if !ok || r.Pool != "Click" {
			t.Fatalf("Choose(Click) = %+v, %v", r, ok)
		}
	}
	if _, ok := p.Choose("Nothing", ctx, rnd); ok {
		t.Fatal("an empty pool chose a clip")
	}
}

// A press on the pet's body plays a click reaction once the greeting is over.
func TestPressPlaysAClickReaction(t *testing.T) {
	app := nsgotest.Start(t, App(packs(t)), macOS)
	app.Play("id:slot1", core.VideoInfo{Duration: 5 * time.Second})
	app.Progress("id:slot1", 5*time.Second, 5*time.Second)
	app.Play("id:slot0", core.VideoInfo{Duration: 5 * time.Second})
	app.Press("id:body")
	got := source(app, "slot1")
	name := strings.TrimSuffix(strings.TrimPrefix(got, "asset:pets/nsgo-chan/mov/"), ".mov")
	if r := packs(t)[0].Policy.Find(name); r.Pool != "Click" {
		t.Fatalf("after a press the other player has %q, want a click reaction", got)
	}
}

// Pets of a group are listed under its name after the ungrouped ones.
func TestPetSectionsGroup(t *testing.T) {
	ps := []*pet.Pack{{ID: "a", Name: "A"}, {ID: "b", Name: "B", Group: "G"}, {ID: "c", Name: "C"}, {ID: "d", Name: "D", Group: "G"}}
	s := petSections(ps)
	if len(s) != 2 || s[0].Group != "" || len(s[0].Packs) != 2 || s[0].Packs[1].ID != "c" || s[1].Group != "G" || len(s[1].Packs) != 2 || s[1].Packs[1].ID != "d" {
		t.Fatalf("sections = %+v", s)
	}
}

// The tray's menu opens the settings with Command-comma and quits with Command-Q.
func TestTrayMenu(t *testing.T) {
	var opened, quit bool
	m := trayMenu(func() { opened = true }, func() { quit = true })
	var labels []string
	for _, it := range m {
		if it.Separator {
			continue
		}
		labels = append(labels, it.Label+" "+it.Shortcut)
		it.OnSelect()
	}
	if !slices.Equal(labels, []string{"设置 mod+,", "退出 mod+q"}) || !opened || !quit {
		t.Fatalf("menu %v, opened %v, quit %v", labels, opened, quit)
	}
}

// A right click on the pet shows no menu.
func TestPetHasNoMenu(t *testing.T) {
	app := nsgotest.Start(t, App(packs(t)), macOS)
	app.RightClick("id:body")
	if m := app.PopupMenu; m != nil {
		t.Fatalf("a right click showed %+v", m)
	}
}

// prefs are the settings' state over a signal of the chosen pet and one of the statistics.
func prefs(t *testing.T, cx *core.Scope, chosen *core.Signal[string], stats *core.Signal[bool]) Prefs {
	return Prefs{Packs: packs(t), Chosen: chosen.Get, Choose: chosen.Set, Stats: stats.Get, SetStats: stats.Set}
}

// The settings' first page shows every pet idling, and a press on one chooses it.
func TestSettingsGridChoosesAPet(t *testing.T) {
	var chosen *core.Signal[string]
	app := nsgotest.Start(t, func(cx *core.Scope) core.Node {
		chosen = core.NewSignal(cx, "nsgo-chan")
		return Settings(cx, prefs(t, cx, chosen, core.NewSignal(cx, true)))
	}, macOS)
	for _, pk := range packs(t) {
		if got := source(app, "clip-"+pk.ID); got != idleClip(pk, core.OSMacOS) {
			t.Errorf("%s shows %q", pk.ID, got)
		}
	}
	app.Press("id:pet-bigbluefish")
	if chosen.GetUntracked() != "bigbluefish" {
		t.Fatalf("chosen %q after pressing Big Blue Fish", chosen.GetUntracked())
	}
}

// The settings check for updates now and offer the newer version found.
func TestSettingsCheckForUpdates(t *testing.T) {
	app := nsgotest.Start(t, func(cx *core.Scope) core.Node {
		return Settings(cx, prefs(t, cx, core.NewSignal(cx, "nsgo-chan"), core.NewSignal(cx, true)))
	}, macOS)
	app.Update = func(bool) core.UpdateCheck {
		return core.UpdateCheck{App: "nsgo-pet", Current: "0.1.0", Version: "0.2.0", URL: "https://github.com/nsgo-dev/nsgo-pet/releases/tag/v0.2.0"}
	}
	app.Press("id:page-通用")
	app.Press("id:update")
	if !slices.Equal(app.UpdateChecks, []bool{false}) {
		t.Fatalf("checks %v, want one the user asked for", app.UpdateChecks)
	}
	if a := app.Alerts(); len(a) != 1 || !strings.Contains(a[0].Message, "0.2.0") {
		t.Fatalf("alerts %+v, want 0.2.0 offered", a)
	}
}

// The settings' switch turns the statistics off, which the storage keeps.
func TestSettingsTurnStatsOff(t *testing.T) {
	var stats *core.Signal[bool]
	app := nsgotest.Start(t, func(cx *core.Scope) core.Node {
		stats = core.NewSignal(cx, statsOn(cx))
		p := prefs(t, cx, core.NewSignal(cx, "nsgo-chan"), stats)
		p.SetStats = func(on bool) { setStatsOn(cx, on); stats.Set(on) }
		return Settings(cx, p)
	}, macOS)
	if !stats.GetUntracked() {
		t.Fatal("statistics start off")
	}
	app.Press("id:page-通用")
	app.Press("id:stats")
	if v, _ := app.Storage().GetItem(statsKey); stats.GetUntracked() || v != "off" {
		t.Fatalf("after the switch: on %v, stored %q", stats.GetUntracked(), v)
	}
}

// A pet with outfits idles in the summer outfit in summer and in the base one in spring.
func TestOutfitsFollowTheSeason(t *testing.T) {
	var pk *pet.Pack
	for _, p := range packs(t) {
		if p.ID == "playpets-clover" {
			pk = p
		}
	}
	if pk == nil {
		t.Fatal("no playpets-clover pet")
	}
	rnd := rand.New(rand.NewPCG(3, 4))
	for _, c := range []struct {
		when, file string
	}{{"2026-07-10 15:00", "idle-summer.webm"}, {"2026-04-10 15:00", "idle.webm"}, {"2026-01-10 15:00", "idle-winter.webm"}} {
		seen := map[string]bool{}
		for range 200 {
			r, ok := pk.Policy.Choose("Idle", pet.Context{Now: at(c.when)}, rnd)
			if !ok {
				t.Fatalf("%s: no idle", c.when)
			}
			seen[r.File] = true
		}
		if !seen[c.file] {
			t.Errorf("%s: idles %v, want %s among them", c.when, seen, c.file)
		}
		for f := range seen {
			if c.file == "idle.webm" && strings.Contains(f, "-") || c.file != "idle.webm" && !strings.Contains(f, strings.TrimPrefix(strings.TrimSuffix(c.file, ".webm"), "idle")) && !strings.HasPrefix(f, "waiting") {
				t.Errorf("%s: idles in %s", c.when, f)
			}
		}
	}
}

func TestTestsReportNoStats(t *testing.T) {
	app := nsgotest.Start(t, App(packs(t)), macOS)
	if _, ok := app.Storage().GetItem(installKey); ok {
		t.Error("a test build made an install ID, so it would report its use")
	}
}

func TestStatsAgentNamesThePlatform(t *testing.T) {
	if a := statsAgent(core.OSMacOS); !strings.Contains(a, "Macintosh") {
		t.Errorf("macOS agent = %q", a)
	}
	if a := statsAgent(core.OSWeb); a != "" {
		t.Errorf("web agent = %q, want none", a)
	}
}

var windows = nsgotest.Environment(core.Environment{OS: core.OSWindows, Idiom: core.IdiomDesktop})

// On Windows a pet chosen in the settings' grid replaces the one playing, as the app's window does, and plays in the engine's own player, an image rather than a video; the engine's tests decode its frames.
func TestWindowsSettingsSwitchPets(t *testing.T) {
	ps := packs(t)
	find := func(id string) *pet.Pack {
		for _, p := range ps {
			if p.ID == id {
				return p
			}
		}
		t.Fatalf("no %s pet", id)
		return nil
	}
	app := nsgotest.Start(t, func(cx *core.Scope) core.Node {
		chosen := core.NewSignal(cx, "nsgo-chan")
		return core.View{Style: core.Style{Direction: core.Row}, Children: []core.Node{
			core.View{Style: core.Style{Width: core.Pt(620), Height: core.Pt(500)}, Children: []core.Node{Settings(cx, prefs(t, cx, chosen, core.NewSignal(cx, true)))}},
			core.Dynamic(func(cx *core.Scope) core.Node {
				p := pet.New(cx, find(chosen.Get()), pet.Options{})
				cx.Post(p.Start)
				return p.View()
			}),
		}}
	}, windows)
	player := func(name string) bool {
		v := app.FindMounted("id:slot1")
		return v.Kind.String() == "image" && v.Props.Label == name
	}
	if !player(find("nsgo-chan").Name) {
		t.Fatalf("NSGo-chan plays in a %s labeled %q, want the engine's image", app.FindMounted("id:slot1").Kind, app.FindMounted("id:slot1").Props.Label)
	}
	app.Press("id:pet-bigbluefish")
	app.WaitUntil(10*time.Second, func() bool { return player(find("bigbluefish").Name) })
}
