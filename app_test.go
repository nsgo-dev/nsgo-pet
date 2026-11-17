package main

import (
	"math/rand/v2"
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

var macOS = nsgotest.Environment(core.Environment{OS: core.OSMacOS, Idiom: core.IdiomDesktop})

func source(app *nsgotest.App, id string) string { return app.FindMounted("id:" + id).Video.Source }

// The fish greets the user first, then goes on with an idle clip in the other player, which shows once it has a picture.
func TestGreetsThenIdles(t *testing.T) {
	app := nsgotest.Start(t, Root, macOS)
	if got := source(app, "slot1"); got != "asset:fish/mov/女仆屈膝礼仪.mov" {
		t.Fatalf("first clip = %q, want the greeting", got)
	}
	app.Play("id:slot1", core.VideoInfo{Duration: 5 * time.Second})
	app.Progress("id:slot1", 5*time.Second, 5*time.Second)
	idle := source(app, "slot0")
	if !strings.HasPrefix(idle, "asset:fish/mov/") {
		t.Fatalf("after the greeting the other player has %q", idle)
	}
	policy, _ := LoadPolicy()
	name := strings.TrimSuffix(strings.TrimPrefix(idle, "asset:fish/mov/"), ".mov")
	if r := policy.Find(name); r.Pool != "Idle" {
		t.Fatalf("after the greeting the fish plays %q from pool %q, want an idle clip", name, r.Pool)
	}
}

// Linux and Windows play the WebM clips.
func TestOtherPlatformsPlayWebM(t *testing.T) {
	app := nsgotest.Start(t, Root, nsgotest.Environment(core.Environment{OS: core.OSLinux, Idiom: core.IdiomDesktop}))
	if got := source(app, "slot1"); got != "asset:fish/webm/女仆屈膝礼仪.webm" {
		t.Fatalf("first clip = %q", got)
	}
}

// Every clip the catalog names ships in both forms.
func TestEveryClipShips(t *testing.T) {
	policy, err := LoadPolicy()
	if err != nil {
		t.Fatal(err)
	}
	if len(policy.Rules) != 106 {
		t.Fatalf("%d rules, want 106", len(policy.Rules))
	}
	for _, id := range []string{"女仆屈膝礼仪", "待机呼吸休闲", "晨间刷牙", "被鼠标拖拽悬空反馈"} {
		if policy.Find(id).ID != id {
			t.Errorf("the pet plays %q by name, which the catalog lacks", id)
		}
	}
	if _, err := os.Stat("resources/fish/mov"); err != nil {
		t.Skip("clips not fetched: run tools/fetch-assets.sh")
	}
	for _, r := range policy.Rules {
		name := strings.TrimSuffix(r.File, ".webm")
		for _, f := range []string{filepath.Join("resources/fish/webm", r.File), filepath.Join("resources/fish/mov", name+".mov")} {
			if _, err := os.Stat(f); err != nil {
				t.Errorf("%s: %v", r.ID, err)
			}
		}
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
	p, err := LoadPolicy()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name string
		id   string
		ctx  Context
		want float64
	}{
		{"sleeping only at night", "原地小憩沉眠", Context{Now: at("2026-11-17 12:00")}, 0},
		{"sleeping at night", "原地小憩沉眠", Context{Now: at("2026-11-17 23:30")}, 40},
		{"disabled", "余额-分文不剩", Context{Now: at("2026-11-17 12:00")}, 0},
		{"running on odd hours", "原地左转奔跑", Context{Now: at("2026-11-17 13:30")}, 200},
		{"no running on even hours", "原地左转奔跑", Context{Now: at("2026-11-17 14:30")}, 0},
		{"no running at the hour's edge", "原地左转奔跑", Context{Now: at("2026-11-17 13:57")}, 0},
		{"dinner near six", "吃午餐", Context{Now: at("2026-11-17 18:30")}, 1000},
		{"no dinner boost at three", "吃午餐", Context{Now: at("2026-11-17 15:00")}, 100},
		{"ice cream only in summer", "吃冰淇淋融化", Context{Now: at("2026-11-17 15:00")}, 0},
		{"coding", "写代码", Context{Now: at("2026-11-17 15:00"), Coding: true}, 1000},
		{"not coding", "写代码", Context{Now: at("2026-11-17 15:00")}, 100},
	} {
		if got := p.Weight(p.Find(c.id), c.ctx); got != c.want {
			t.Errorf("%s: weight of %s = %v, want %v", c.name, c.id, got, c.want)
		}
	}
}

// A clip of a Gregorian festival grows toward its festival weight as the day nears.
func TestFestivalWeightPeaksOnTheDay(t *testing.T) {
	p, _ := LoadPolicy()
	var rule Rule
	for _, r := range p.Rules {
		if r.Festival == "Christmas" {
			rule = r
			break
		}
	}
	if rule.ID == "" {
		t.Fatal("no Christmas clip in the catalog")
	}
	far := p.Weight(rule, Context{Now: at("2026-07-01 15:00")})
	near := p.Weight(rule, Context{Now: at("2026-12-22 15:00")})
	day := p.Weight(rule, Context{Now: at("2026-12-25 15:00")})
	if !(far < near && near < day) || day != rule.FestivalWeight {
		t.Fatalf("weights far %v, near %v, on the day %v (festival weight %v)", far, near, day, rule.FestivalWeight)
	}
}

func TestChooseKeepsToItsPool(t *testing.T) {
	p, _ := LoadPolicy()
	rnd := rand.New(rand.NewPCG(1, 2))
	ctx := Context{Now: at("2026-11-17 15:00")}
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

// A press on the fish's body plays a click reaction once the greeting is over.
func TestPressPlaysAClickReaction(t *testing.T) {
	app := nsgotest.Start(t, Root, macOS)
	app.Play("id:slot1", core.VideoInfo{Duration: 5 * time.Second})
	app.Progress("id:slot1", 5*time.Second, 5*time.Second)
	app.Play("id:slot0", core.VideoInfo{Duration: 5 * time.Second})
	app.Press("id:fish")
	got := source(app, "slot1")
	if !strings.Contains(got, "点击回应-") {
		t.Fatalf("after a press the other player has %q, want a click reaction", got)
	}
}
