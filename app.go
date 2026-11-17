// Command bluefish is Big Blue Fish, a desktop pet, on NSGo: the fish lives in a transparent window above the other windows and plays its clips by Big Blue Fish's own rules.
package main

import (
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/nsgo-dev/nsgo"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/desktop"
)

// The window is half the clips' 640 by 360 pixels.
const (
	petWidth  core.Points = 320
	petHeight core.Points = 180
)

func main() {
	nsgo.Run(Root, nsgo.WindowSize(petWidth, petHeight), nsgo.WindowTransparent(), nsgo.WindowNoShadow(), nsgo.SkipTaskbar(), nsgo.HideDock())
}

// Root is the pet's window.
func Root(cx *core.Scope) core.Node {
	policy, err := LoadPolicy()
	if err != nil {
		log.Fatal(err)
	}
	p := newPet(cx, policy, rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0x6669736821)))
	// Once the window is up: Post runs after this first render.
	cx.Post(func() {
		desktop.SetWindowLevel(cx, desktop.WindowLevelFloating)
		desktop.SetWindowWorkspaces(cx, desktop.WindowWorkspacesAll)
		desktop.SetWindowPointerEvents(cx, desktop.WindowPointerPainted)
		p.placeAtCorner()
		p.greet()
	})
	desktop.NewTray(cx, desktop.TrayOptions{Title: "🐟", Tooltip: "Big Blue Fish", Menu: p.menu})
	return core.View{Style: core.Style{FlexGrow: 1}, Children: []core.Node{
		// The players take no pointer: a video counts as drawn over its whole box, where the fish fills only its middle.
		core.View{PointerEvents: core.PointerEventsNone, Style: fill(), Children: []core.Node{p.slots[0].view(p, 0), p.slots[1].view(p, 1)}},
		// The fish's body takes the presses, the drags and the right clicks; the rest of the window lets the pointer through.
		core.View{
			TestID:        "fish",
			Style:         core.Style{Position: core.Absolute, Left: core.Pct(bodyLeft), Top: core.Pct(bodyTop), Width: core.Pct(bodyWidth), Height: core.Pct(bodyHeight)},
			OnPress:       func(core.PressEvent) { p.click() },
			OnContextMenu: func(e core.ContextMenuEvent) { desktop.PopupMenu(cx, p.menu(), e.Position, nil) },
			OnPan:         p.pan,
		},
	}}
}

// The fish's body in the clips, in percent of the frame: about the middle third, from her ears to her feet.
const (
	bodyLeft   = 33
	bodyTop    = 15
	bodyWidth  = 34
	bodyHeight = 80
)

// fill is a style that covers the parent.
func fill() core.Style {
	return core.Style{Position: core.Absolute, Top: core.Pt(0), Right: core.Pt(0), Bottom: core.Pt(0), Left: core.Pt(0)}
}

// clipSource is where this platform's clips are: HEVC with alpha on Apple platforms, whose player opens no WebM, VP9 with alpha in WebM elsewhere.
func clipSource(cx *core.Scope, r Rule) string {
	name := strings.TrimSuffix(r.File, ".webm")
	switch cx.Env().OS() {
	case core.OSMacOS, core.OSIOS:
		return "asset:fish/mov/" + name + ".mov"
	}
	return "asset:fish/webm/" + name + ".webm"
}

// slot is one of two players: the next clip loads in the hidden one and shows once it has a picture, so the fish never blinks out between clips.
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
		Label:   "Big Blue Fish",
		TestID:  slotIDs[i],
	}
}

// slotIDs name the two players for tests.
var slotIDs = [2]string{"slot0", "slot1"}

// Phases of what the fish is doing, as Big Blue Fish names them.
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

// eventInterval is the time between events, Big Blue Fish's default.
const eventInterval = 40 * time.Second

type pet struct {
	cx     *core.Scope
	policy *Policy
	rnd    *rand.Rand
	slots  [2]*slot
	// front is the slot that shows, pending the one loading the next clip, -1 for none.
	front, pending *core.Signal[int]

	phase     string
	nextEvent time.Time
	lastBrush string

	// move is the walk of a clip with a MoveKind, nil without one.
	move *movement
	// dragging is whether the user holds the fish.
	dragging bool
}

// movement walks the window from start to end while a clip with a MoveKind plays, and back to origin after.
type movement struct {
	rule       Rule
	origin     core.Point
	start, end core.Points
	mirrored   bool
}

func newPet(cx *core.Scope, policy *Policy, rnd *rand.Rand) *pet {
	p := &pet{cx: cx, policy: policy, rnd: rnd, front: core.NewSignal(cx, 0), pending: core.NewSignal(cx, -1)}
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
	src := clipSource(p.cx, r)
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

func (p *pet) greet() { p.play(p.policy.Find("女仆屈膝礼仪"), phaseStartup, false) }

// finished runs as the front clip ends, and picks what follows, as Big Blue Fish's AnimationFinished.
func (p *pet) finished() {
	switch p.phase {
	case phaseExit:
		p.cx.Quit()
		return
	case phaseDrag:
		// Hold the last pose until the fish is let go.
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
		p.nextEvent = time.Now().Add(eventInterval)
	}
	p.boundary()
}

func (p *pet) mustChoose(pool string) Rule {
	r, ok := p.choose(pool)
	if !ok {
		return p.policy.Find("待机呼吸休闲")
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
	if day := now.Format(time.DateOnly); now.Hour() == 23 && p.lastBrush != day {
		p.lastBrush = day
		p.play(p.policy.Find("晨间刷牙"), phaseDaily, false)
		return
	}
	if !now.Before(p.nextEvent) {
		p.nextEvent = now.Add(eventInterval)
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
	return min(max(x, s.Origin.X), s.Origin.X+s.Size.Width-petWidth)
}

// placeAtCorner puts the fish on the bottom right of the screen, above the Dock or the taskbar.
func (p *pet) placeAtCorner() {
	s := p.screen()
	if s.Size.Width == 0 {
		return
	}
	desktop.SetWindowFrame(p.cx, desktop.WindowFrame{
		X: s.Origin.X + s.Size.Width - petWidth - 24, Y: s.Origin.Y + s.Size.Height - petHeight,
		Move: true, Width: petWidth, Height: petHeight,
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

// pan drags the fish: the pointer's place in the window is where it was pressed plus the translation, so moving the window by the translation keeps the fish under the pointer.
func (p *pet) pan(e core.PanEvent) {
	switch e.Phase {
	case core.GestureBegan:
		p.dragging = true
		p.move = nil
		p.play(p.policy.Find("被鼠标拖拽悬空反馈"), phaseDrag, false)
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
	p.play(p.policy.Find("女仆屈膝礼仪"), phaseExit, false)
	// Leave even if the clip never ends.
	p.cx.AfterFunc(8*time.Second, p.cx.Quit)
}

func (p *pet) menu() []core.MenuItem {
	return []core.MenuItem{
		{Label: "来点动作", OnSelect: func() {
			if r, ok := p.choose("Event"); ok {
				p.nextEvent = time.Now().Add(eventInterval)
				p.startEvent(r)
			}
		}},
		{Label: "回到右下角", OnSelect: p.placeAtCorner},
		{Separator: true},
		{Label: "退出", OnSelect: p.quit},
	}
}
