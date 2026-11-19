// Command nsgo-pet is a desktop pet on NSGo whose pet can be switched: each pet in resources/pets lives in a transparent window above the other windows and plays its clips by its own rules, in Big Blue Fish's format.
package main

import (
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/nsgo-dev/nsgo"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/desktop"
	"github.com/nsgo-dev/nsgo/updater"
)

func main() {
	packs, err := LoadPacks()
	if err != nil {
		log.Fatal(err)
	}
	nsgo.Run(App(packs), nsgo.WindowSize(packs[0].Width, packs[0].Height), nsgo.WindowTransparent(), nsgo.WindowNoShadow(), nsgo.SkipTaskbar(), nsgo.HideDock())
}

// petKey is the storage key of the pet the user chose last.
const petKey = "pet"

// App is the pet's window, showing the pet the user chose last, the first of packs at first.
func App(packs []*Pack) func(cx *core.Scope) core.Node {
	return func(cx *core.Scope) core.Node {
		find := func(id string) *Pack {
			for _, p := range packs {
				if p.ID == id {
					return p
				}
			}
			return packs[0]
		}
		saved, _ := cx.Storage().GetItem(petKey)
		chosen := core.NewSignal(cx, find(saved).ID)
		var current *pet
		choose := func(id string) {
			if id == chosen.GetUntracked() {
				return
			}
			if err := cx.Storage().SetItem(petKey, id); err != nil {
				log.Printf("saving the pet: %v", err)
			}
			chosen.Set(id)
		}
		menu := func() []core.MenuItem {
			items := []core.MenuItem{
				{Label: "切换桌宠", Items: petMenu(packs, chosen.GetUntracked(), choose)},
				// Sparkle's windows on macOS, an alert that installs the signed package on Linux.
				{Label: "检查更新…", OnSelect: func() { updater.Check(cx) }},
				{Separator: true},
			}
			if current != nil {
				items = append(items, current.menu()...)
			}
			return items
		}
		cx.Post(func() {
			desktop.SetWindowLevel(cx, desktop.WindowLevelFloating)
			desktop.SetWindowWorkspaces(cx, desktop.WindowWorkspacesAll)
			desktop.SetWindowPointerEvents(cx, desktop.WindowPointerPainted)
		})
		desktop.NewTray(cx, desktop.TrayOptions{Title: "🐾", Tooltip: "nsgo-pet", Menu: menu})
		return core.Dynamic(func(cx *core.Scope) core.Node {
			pack := find(chosen.Get())
			p := newPet(cx, pack, rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0x706574)))
			current = p
			cx.Post(func() {
				p.placeAtCorner()
				p.greet()
			})
			return p.view(menu)
		})
	}
}

// petMenu lists packs to choose from, those of a group in a submenu of its name after the ungrouped ones, each in the order packs has them.
func petMenu(packs []*Pack, chosen string, choose func(id string)) []core.MenuItem {
	var top []core.MenuItem
	groups := map[string]int{}
	var grouped []core.MenuItem
	for _, pk := range packs {
		item := core.MenuItem{Label: pk.Name, Radio: true, Checked: pk.ID == chosen, OnSelect: func() { choose(pk.ID) }}
		if pk.Group == "" {
			top = append(top, item)
			continue
		}
		i, ok := groups[pk.Group]
		if !ok {
			i = len(grouped)
			groups[pk.Group] = i
			grouped = append(grouped, core.MenuItem{Label: pk.Group})
		}
		grouped[i].Items = append(grouped[i].Items, item)
	}
	return append(top, grouped...)
}

// view is the pet: two players under a body that takes the pointer.
func (p *pet) view(menu func() []core.MenuItem) core.Node {
	b := p.pack.Body
	return core.View{Style: core.Style{FlexGrow: 1}, Children: []core.Node{
		// The players take no pointer: a video counts as drawn over its whole box, where the pet fills only part of it.
		core.View{PointerEvents: core.PointerEventsNone, Style: fill(), Children: []core.Node{p.slots[0].view(p, 0), p.slots[1].view(p, 1)}},
		// The pet's body takes the presses, the drags and the right clicks; the rest of the window lets the pointer through. Pointer hit testing by the clip's alpha is NSGo's #599.
		core.View{
			TestID:        "body",
			Style:         core.Style{Position: core.Absolute, Left: core.Pct(b.Left), Top: core.Pct(b.Top), Width: core.Pct(b.Width), Height: core.Pct(b.Height)},
			OnPress:       func(core.PressEvent) { p.click() },
			OnContextMenu: func(e core.ContextMenuEvent) { desktop.PopupMenu(p.cx, menu(), e.Position, nil) },
			OnPan:         p.pan,
		},
	}}
}

// fill is a style that covers the parent.
func fill() core.Style {
	return core.Style{Position: core.Absolute, Top: core.Pt(0), Right: core.Pt(0), Bottom: core.Pt(0), Left: core.Pt(0)}
}

// clipSource is where this platform's clips are: HEVC with alpha on macOS, whose player opens no WebM, VP9 with alpha in WebM elsewhere.
func clipSource(cx *core.Scope, pack *Pack, r Rule) string {
	name := strings.TrimSuffix(r.File, ".webm")
	switch cx.Env().OS() {
	case core.OSMacOS:
		return "asset:pets/" + pack.ID + "/mov/" + name + ".mov"
	}
	return "asset:pets/" + pack.ID + "/webm/" + name + ".webm"
}

// slot is one of two players: the next clip loads in the hidden one and shows once it has a picture, so the pet never blinks out between clips.
type slot struct {
	src      *core.Signal[string]
	mirrored *core.Signal[bool]
	ref      *core.Ref
}

func (s *slot) view(p *pet, i int) core.Node {
	return core.Video{
		SourceFunc: s.src.Get,
		PausedFunc: func() bool { return p.front.Get() != i && p.pending.Get() != i },
		Muted:      true,
		ResizeMode: core.Contain,
		Ref:        s.ref,
		StyleFunc: func() core.Style {
			st := fill()
			st.Opacity = core.Alpha(0)
			if p.front.Get() == i {
				st.Opacity = core.Alpha(1)
			}
			if s.mirrored.Get() {
				st.Transform = core.Transform{ScaleX: -1}
			}
			return st
		},
		OnLoad: func(core.VideoInfo) {
			if p.pending.GetUntracked() == i {
				p.front.Set(i)
				p.pending.Set(-1)
			}
		},
		OnProgress: func(position, _ time.Duration) {
			if p.front.GetUntracked() == i {
				p.progress(position)
			}
		},
		ProgressInterval: 33 * time.Millisecond,
		OnEnd: func() {
			if p.front.GetUntracked() == i && p.pending.GetUntracked() < 0 {
				p.finished()
			}
		},
		OnError: func(err error) { log.Printf("clip %s: %v", s.src.GetUntracked(), err) },
		Label:   p.pack.Name,
		TestID:  slotIDs[i],
	}
}

// slotIDs name the two players for tests.
var slotIDs = [2]string{"slot0", "slot1"}

// Phases of what the pet is doing, as Big Blue Fish names them.
const (
	phaseStartup      = "Startup"
	phaseIdle         = "Idle"
	phaseEvent        = "Event"
	phaseClick        = "Click"
	phaseDrag         = "Drag"
	phaseDaily        = "Daily"
	phaseOutbound     = "Outbound"
	phaseBetweenMoves = "BetweenMoves"
	phaseReturn       = "Return"
	phaseExit         = "Exit"
)

type pet struct {
	cx     *core.Scope
	pack   *Pack
	policy *Policy
	rnd    *rand.Rand
	slots  [2]*slot
	// front is the slot that shows, pending the one loading the next clip, -1 for none.
	front, pending *core.Signal[int]

	phase     string
	nextEvent time.Time
	lastDaily string

	// move is the walk of a clip with a MoveKind, nil without one.
	move *movement
	// dragging is whether the user holds the pet.
	dragging bool
}

// movement walks the window from start to end while a clip with a MoveKind plays, and back to origin after.
type movement struct {
	rule       Rule
	origin     core.Point
	start, end core.Points
	mirrored   bool
}

func newPet(cx *core.Scope, pack *Pack, rnd *rand.Rand) *pet {
	p := &pet{cx: cx, pack: pack, policy: pack.Policy, rnd: rnd, front: core.NewSignal(cx, 0), pending: core.NewSignal(cx, -1)}
	for i := range p.slots {
		p.slots[i] = &slot{src: core.NewSignal(cx, ""), mirrored: core.NewSignal(cx, false), ref: &core.Ref{}}
	}
	return p
}

func (p *pet) context() Context {
	c := Context{Now: time.Now()}
	if desktop.CanReadFrontmostApp(p.cx) {
		c.Coding = isCodingApp(desktop.FrontmostApp(p.cx).ID)
	}
	return c
}

// codingApps are editors and terminals by bundle ID or the start of one, with Windows' and Linux's executable and desktop names.
var codingApps = []string{
	"com.microsoft.VSCode", "com.apple.dt.Xcode", "com.jetbrains.", "dev.zed.Zed", "com.todesktop.230313mzl4w4u92",
	"com.apple.Terminal", "com.googlecode.iterm2", "dev.warp.", "com.mitchellh.ghostty", "net.kovidgoyal.kitty",
	"code", "Code.exe", "idea", "WindowsTerminal.exe", "org.gnome.Terminal",
}

func isCodingApp(id string) bool {
	for _, c := range codingApps {
		if id == c || strings.HasSuffix(c, ".") && strings.HasPrefix(id, c) || strings.HasSuffix(id, `\`+c) || strings.HasSuffix(id, "/"+c) {
			return true
		}
	}
	return false
}

// play shows rule's clip, mirrored when asked, for phase.
func (p *pet) play(r Rule, phase string, mirrored bool) {
	log.Printf("play %s (%s)", r.ID, phase)
	p.phase = phase
	src := clipSource(p.cx, p.pack, r)
	front := p.front.GetUntracked()
	back := 1 - front
	s := p.slots[back]
	s.mirrored.Set(mirrored)
	if s.src.GetUntracked() == src {
		// The hidden player has this clip already: play it again from the start and show it.
		s.ref.Seek(0)
		p.pending.Set(-1)
		p.front.Set(back)
		return
	}
	s.src.Set(src)
	p.pending.Set(back)
}

func (p *pet) choose(pool string) (Rule, bool) { return p.policy.Choose(pool, p.context(), p.rnd) }

func (p *pet) playIdle() {
	if r, ok := p.choose("Idle"); ok {
		p.play(r, phaseIdle, false)
	}
}

func (p *pet) greet() { p.play(p.policy.Find(p.pack.Clips.Greeting), phaseStartup, false) }

// eventInterval is the time between events, 40 seconds unless the pet says otherwise.
func (p *pet) eventInterval() time.Duration {
	if s := p.pack.EventSeconds; s > 0 {
		return time.Duration(s) * time.Second
	}
	return 40 * time.Second
}

// finished runs as the front clip ends, and picks what follows, as Big Blue Fish's AnimationFinished.
func (p *pet) finished() {
	switch p.phase {
	case phaseExit:
		p.cx.Quit()
		return
	case phaseDrag:
		// Hold the last pose until the pet is let go.
		return
	case phaseOutbound:
		if m := p.move; m != nil {
			p.moveTo(m.end)
			p.play(p.mustChoose("Idle"), phaseBetweenMoves, false)
			return
		}
	case phaseBetweenMoves:
		if m := p.move; m != nil {
			m.start, m.end = m.end, m.origin.X
			p.play(m.rule, phaseReturn, !m.mirrored)
			return
		}
	case phaseReturn:
		if m := p.move; m != nil {
			p.moveTo(m.origin.X)
			p.move = nil
		}
	case phaseStartup:
		p.nextEvent = time.Now().Add(p.eventInterval())
	}
	p.boundary()
}

func (p *pet) mustChoose(pool string) Rule {
	r, ok := p.choose(pool)
	if !ok {
		return p.policy.Find(p.pack.Clips.Idle)
	}
	return r
}

// boundary picks the next clip between clips, as Big Blue Fish's AnimationBoundary.
func (p *pet) boundary() {
	if p.dragging {
		p.playIdle()
		return
	}
	now := time.Now()
	if c := p.pack.Clips; c.Daily != "" {
		if day := now.Format(time.DateOnly); now.Hour() == c.DailyHour && p.lastDaily != day {
			p.lastDaily = day
			p.play(p.policy.Find(c.Daily), phaseDaily, false)
			return
		}
	}
	if !now.Before(p.nextEvent) {
		p.nextEvent = now.Add(p.eventInterval())
		if r, ok := p.choose("Event"); ok {
			p.startEvent(r)
			return
		}
	}
	p.playIdle()
}

// startEvent plays an event, walking the window for one with a MoveKind.
func (p *pet) startEvent(r Rule) {
	if r.MoveKind == "" {
		p.play(r, phaseEvent, false)
		return
	}
	f := desktop.Frame(p.cx)
	m := &movement{rule: r, origin: core.Point{X: f.X, Y: f.Y}, start: f.X}
	// Seventy points to the left, or to the right where the screen ends.
	m.end = p.clampX(f.X - 70)
	if d := m.end - m.start; d > -15 && d < 15 {
		m.end = p.clampX(f.X + 70)
	}
	m.mirrored = m.end > m.start
	p.move = m
	p.play(r, phaseOutbound, m.mirrored)
}

// progress moves the window along while a walking clip plays, as Big Blue Fish's MoveWithAnimation: the walk takes the clip's middle, between a lead and a tail.
func (p *pet) progress(position time.Duration) {
	m := p.move
	if m == nil || p.phase != phaseOutbound && p.phase != phaseReturn {
		return
	}
	lead, tail := 2.0, 2.0
	if m.rule.MoveKind == "Run" {
		lead, tail = 1.75, 4.8
	}
	fraction := min(max((position.Seconds()-lead)/(10-lead-tail), 0), 1)
	p.moveTo(m.start + (m.end-m.start)*core.Points(fraction))
}

func (p *pet) moveTo(x core.Points) {
	f := desktop.Frame(p.cx)
	desktop.SetWindowFrame(p.cx, desktop.WindowFrame{X: x, Y: f.Y, Move: true})
}

// screen is the visible frame of the display the window is on, the main display's when none holds its center.
func (p *pet) screen() core.Bounds {
	f := desktop.Frame(p.cx)
	center := core.Point{X: f.X + f.Width/2, Y: f.Y + f.Height/2}
	var main core.Bounds
	for _, d := range p.cx.Env().Displays() {
		if d.Frame.Contains(center) {
			return d.VisibleFrame
		}
		if d.Main {
			main = d.VisibleFrame
		}
	}
	return main
}

func (p *pet) clampX(x core.Points) core.Points {
	s := p.screen()
	if s.Size.Width == 0 {
		return x
	}
	return min(max(x, s.Origin.X), s.Origin.X+s.Size.Width-p.pack.Width)
}

// placeAtCorner sizes the window for the pet and puts it on the bottom right of the screen, above the Dock or the taskbar.
func (p *pet) placeAtCorner() {
	s := p.screen()
	if s.Size.Width == 0 {
		return
	}
	desktop.SetWindowFrame(p.cx, desktop.WindowFrame{
		X: s.Origin.X + s.Size.Width - p.pack.Width - 24, Y: s.Origin.Y + s.Size.Height - p.pack.Height,
		Move: true, Width: p.pack.Width, Height: p.pack.Height,
	})
}

func (p *pet) click() {
	if p.phase == phaseExit || p.phase == phaseStartup {
		return
	}
	p.move = nil
	if r, ok := p.choose("Click"); ok {
		p.play(r, phaseClick, false)
	}
}

// pan drags the pet: the pointer's place in the window is where it was pressed plus the translation, so moving the window by the translation keeps the pet under the pointer.
func (p *pet) pan(e core.PanEvent) {
	switch e.Phase {
	case core.GestureBegan:
		p.dragging = true
		p.move = nil
		p.play(p.policy.Find(p.pack.Clips.Drag), phaseDrag, false)
	case core.GestureChanged:
		f := desktop.Frame(p.cx)
		desktop.SetWindowFrame(p.cx, desktop.WindowFrame{X: f.X + e.Translation.X, Y: f.Y + e.Translation.Y, Move: true})
	case core.GestureEnded, core.GestureCancelled:
		p.dragging = false
		p.playIdle()
	}
}

func (p *pet) quit() {
	if p.phase == phaseExit {
		return
	}
	p.move = nil
	p.play(p.policy.Find(p.pack.Clips.Exit), phaseExit, false)
	// Leave even if the clip never ends.
	p.cx.AfterFunc(8*time.Second, p.cx.Quit)
}

func (p *pet) menu() []core.MenuItem {
	return []core.MenuItem{
		{Label: "来点动作", OnSelect: func() {
			if r, ok := p.choose("Event"); ok {
				p.nextEvent = time.Now().Add(p.eventInterval())
				p.startEvent(r)
			}
		}},
		{Label: "回到右下角", OnSelect: p.placeAtCorner},
		{Separator: true},
		{Label: "退出", OnSelect: p.quit},
	}
}
