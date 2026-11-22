// Package pet plays desktop pets on NSGo: a pack's clips, chosen by its rules in Big Blue Fish's format, in a window of the pet's own or in a view of the host's. It ships no pets: the host loads its packs with Load or LoadDir, and owns the windows, menus and lifecycle around them.
package pet

import (
	"log"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/desktop"
)

// Options are how a host shows a pet.
type Options struct {
	// Window says the pet's scope is a window of its own, sized to the pet, as WindowOptions opens one: Start floats it above the other windows on every workspace and lets the pointer through outside the pet's body, walking clips and drags move it, and PlaceAtCorner puts it on the bottom right of the screen. Without it the pet is a view of its size in the host's layout, which plays every clip where it is.
	Window bool
	// Menu is what a right click on the pet shows; nil shows nothing.
	Menu func() []core.MenuItem
	// Context is what the clips' weights depend on besides the time, such as whether the user is coding or working; its Now is the clock's when zero. Nil reads whether an editor or a terminal is the app in front.
	Context func() Context
	// Rand chooses the clips; nil seeds one from the clock.
	Rand *rand.Rand
}

// WindowOptions are those of a window for pack's pet, for desktop.OpenWindow to open with Options.Window: its size, transparent, without a shadow or a taskbar entry, on every workspace and taking the pointer only where it draws.
func WindowOptions(pack *Pack) desktop.WindowOptions {
	return desktop.WindowOptions{
		Title: pack.Name, Width: pack.Width, Height: pack.Height,
		Transparent: true, NoShadow: true, SkipTaskbar: true,
		Workspaces: desktop.WindowWorkspacesAll, PointerEvents: desktop.WindowPointerPainted,
	}
}

// Pet is one pet playing in a scope.
type Pet struct {
	cx     *core.Scope
	pack   *Pack
	opts   Options
	policy *Policy
	rnd    *rand.Rand
	slots  [2]*slot
	// front is the slot that shows, pending the one loading the next clip, -1 for none.
	front, pending *core.Signal[int]

	phase     string
	nextEvent time.Time
	lastDaily string
	// left runs once the exit clip ends, nil before Leave.
	left func()

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

// New makes pack's pet in cx, which View shows and Start begins.
func New(cx *core.Scope, pack *Pack, opts Options) *Pet {
	rnd := opts.Rand
	if rnd == nil {
		rnd = rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0x706574))
	}
	p := &Pet{cx: cx, pack: pack, opts: opts, policy: pack.Policy, rnd: rnd, front: core.NewSignal(cx, 0), pending: core.NewSignal(cx, -1)}
	for i := range p.slots {
		p.slots[i] = &slot{src: core.NewSignal(cx, ""), mirrored: core.NewSignal(cx, false), ref: &core.Ref{}}
	}
	return p
}

// Pack is the pet's pack.
func (p *Pet) Pack() *Pack { return p.pack }

// Start sets up the pet's window, with Options.Window, puts it in the corner and has the pet greet the user.
func (p *Pet) Start() {
	if p.opts.Window {
		desktop.SetWindowLevel(p.cx, desktop.WindowLevelFloating)
		desktop.SetWindowWorkspaces(p.cx, desktop.WindowWorkspacesAll)
		desktop.SetWindowPointerEvents(p.cx, desktop.WindowPointerPainted)
		p.PlaceAtCorner()
	}
	p.play(p.policy.Find(p.pack.Clips.Greeting), phaseStartup, false)
}

// View is the pet: two players under a body that takes the pointer.
func (p *Pet) View() core.Node {
	b := p.pack.Body
	st := core.Style{FlexGrow: 1}
	if !p.opts.Window {
		st = core.Style{Width: core.Pt(p.pack.Width), Height: core.Pt(p.pack.Height), FlexShrink: 0}
	}
	return core.View{Style: st, Children: []core.Node{
		// The players take no pointer: a video counts as drawn over its whole box, where the pet fills only part of it.
		core.View{PointerEvents: core.PointerEventsNone, Style: fill(), Children: []core.Node{p.slots[0].view(p, 0), p.slots[1].view(p, 1)}},
		// The pet's body takes the presses, the drags and the right clicks; the rest of the window lets the pointer through. Pointer hit testing by the clip's alpha is NSGo's #599.
		core.View{
			TestID:  "body",
			Style:   core.Style{Position: core.Absolute, Left: core.Pct(b.Left), Top: core.Pct(b.Top), Width: core.Pct(b.Width), Height: core.Pct(b.Height)},
			OnPress: func(core.PressEvent) { p.click() },
			OnContextMenu: func(e core.ContextMenuEvent) {
				if p.opts.Menu != nil {
					desktop.PopupMenu(p.cx, p.opts.Menu(), e.Position, nil)
				}
			},
			OnPan: p.pan,
		},
	}}
}

// fill is a style that covers the parent.
func fill() core.Style {
	return core.Style{Position: core.Absolute, Top: core.Pt(0), Right: core.Pt(0), Bottom: core.Pt(0), Left: core.Pt(0)}
}

// clipSource is where this platform's clips are: HEVC with alpha on macOS, whose player opens no WebM, VP9 with alpha in WebM elsewhere, which Windows decodes itself (clipPlayer).
func clipSource(cx *core.Scope, pack *Pack, r Rule) string {
	name := strings.TrimSuffix(r.File, ".webm")
	switch cx.Env().OS() {
	case core.OSMacOS:
		return pack.Base + "/mov/" + name + ".mov"
	}
	return pack.Base + "/webm/" + name + ".webm"
}

// ClipView plays one of a pack's clips over and over, muted and contained in its box, as a host's list of pets previews each: in a core.Video, or on Windows in the engine's own decoder, as the pet plays its clips there. A clip the pack lacks shows nothing.
type ClipView struct {
	Pack *Pack
	// Clip is the clip's ID, as Play takes, such as Pack.Clips.Idle.
	Clip   string
	Class  string
	Style  core.Style
	TestID string
}

// Render implements core.Node.
func (v ClipView) Render(cx *core.Scope) core.Node {
	r := v.Pack.Policy.Find(v.Clip)
	if r.ID == "" {
		return core.View{Class: v.Class, Style: v.Style, TestID: v.TestID}
	}
	src := clipSource(cx, v.Pack, r)
	if cx.Env().OS() != core.OSWindows {
		return core.Video{Source: src, Repeat: true, Muted: true, ResizeMode: core.Contain, Class: v.Class, Style: v.Style, Label: v.Pack.Name, TestID: v.TestID}
	}
	// The decoder stops with the view, as clipView's cleanup stops its player; at the end the player reads the clip again and keeps its last frame up until the first shows.
	var pl *clipPlayer
	return clipView{
		src:        func() string { return src },
		paused:     func() bool { return false },
		style:      func() core.Style { return v.Style },
		onLoad:     func() {},
		onProgress: func(time.Duration) {},
		onEnd:      func() { pl.restart() },
		onError:    func(err error) { log.Printf("clip %s: %v", src, err) },
		player:     &pl,
		class:      v.Class,
		label:      v.Pack.Name,
		testID:     v.TestID,
	}
}

// slot is one of two players: the next clip loads in the hidden one and shows once it has a picture, so the pet never blinks out between clips.
type slot struct {
	src      *core.Signal[string]
	mirrored *core.Signal[bool]
	ref      *core.Ref
	// player plays the clips on Windows instead of a video, nil elsewhere.
	player *clipPlayer
}

func (s *slot) view(p *Pet, i int) core.Node {
	paused := func() bool { return p.front.Get() != i && p.pending.Get() != i }
	style := func() core.Style {
		st := fill()
		st.Opacity = core.Alpha(0)
		if p.front.Get() == i {
			st.Opacity = core.Alpha(1)
		}
		if s.mirrored.Get() {
			st.Transform = core.Transform{ScaleX: -1}
		}
		return st
	}
	loaded := func() {
		if p.pending.GetUntracked() == i {
			p.front.Set(i)
			p.pending.Set(-1)
		}
	}
	progress := func(position time.Duration) {
		if p.front.GetUntracked() == i {
			p.progress(position)
		}
	}
	ended := func() {
		if p.front.GetUntracked() == i && p.pending.GetUntracked() < 0 {
			p.finished()
		}
	}
	failed := func(err error) { log.Printf("clip %s: %v", s.src.GetUntracked(), err) }
	if p.cx.Env().OS() == core.OSWindows {
		return clipView{src: s.src.Get, paused: paused, style: style, onLoad: loaded, onProgress: progress, onEnd: ended, onError: failed, player: &s.player, label: p.pack.Name, testID: slotIDs[i]}
	}
	return core.Video{
		SourceFunc:       s.src.Get,
		PausedFunc:       paused,
		Muted:            true,
		ResizeMode:       core.Contain,
		Ref:              s.ref,
		StyleFunc:        style,
		OnLoad:           func(core.VideoInfo) { loaded() },
		OnProgress:       func(position, _ time.Duration) { progress(position) },
		ProgressInterval: 33 * time.Millisecond,
		OnEnd:            ended,
		OnError:          failed,
		Label:            p.pack.Name,
		TestID:           slotIDs[i],
	}
}

// rewind plays the slot's clip again from the start and says whether the slot may show at once. A video seeks in the player it has open; Windows' player reads and decodes the clip again, and until its first frame is up shows the frame it last showed, so the slot waits for the load as for a new clip.
func (s *slot) rewind() (now bool) {
	if s.player != nil {
		s.player.restart()
		return false
	}
	s.ref.Seek(0)
	return true
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

func (p *Pet) context() Context {
	var c Context
	if p.opts.Context != nil {
		c = p.opts.Context()
	} else if desktop.CanReadFrontmostApp(p.cx) {
		c.Coding = isCodingApp(desktop.FrontmostApp(p.cx).ID)
	}
	if c.Now.IsZero() {
		c.Now = time.Now()
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
func (p *Pet) play(r Rule, phase string, mirrored bool) {
	log.Printf("play %s (%s)", r.ID, phase)
	p.phase = phase
	src := clipSource(p.cx, p.pack, r)
	front := p.front.GetUntracked()
	back := 1 - front
	s := p.slots[back]
	s.mirrored.Set(mirrored)
	if s.src.GetUntracked() == src {
		// The hidden player has this clip already: play it again from the start and show it.
		if s.rewind() {
			p.pending.Set(-1)
			p.front.Set(back)
		} else {
			p.pending.Set(back)
		}
		return
	}
	s.src.Set(src)
	p.pending.Set(back)
}

func (p *Pet) choose(pool string) (Rule, bool) { return p.policy.Choose(pool, p.context(), p.rnd) }

func (p *Pet) playIdle() {
	if r, ok := p.choose("Idle"); ok {
		p.play(r, phaseIdle, false)
	}
}

// eventInterval is the time between events, 40 seconds unless the pet says otherwise.
func (p *Pet) eventInterval() time.Duration {
	if s := p.pack.EventSeconds; s > 0 {
		return time.Duration(s) * time.Second
	}
	return 40 * time.Second
}

// finished runs as the front clip ends, and picks what follows, as Big Blue Fish's AnimationFinished.
func (p *Pet) finished() {
	switch p.phase {
	case phaseExit:
		p.left()
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

func (p *Pet) mustChoose(pool string) Rule {
	r, ok := p.choose(pool)
	if !ok {
		return p.policy.Find(p.pack.Clips.Idle)
	}
	return r
}

// boundary picks the next clip between clips, as Big Blue Fish's AnimationBoundary.
func (p *Pet) boundary() {
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
func (p *Pet) startEvent(r Rule) {
	if r.MoveKind == "" || !p.opts.Window {
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
func (p *Pet) progress(position time.Duration) {
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

func (p *Pet) moveTo(x core.Points) {
	f := desktop.Frame(p.cx)
	desktop.SetWindowFrame(p.cx, desktop.WindowFrame{X: x, Y: f.Y, Move: true})
}

// screen is the visible frame of the display the window is on, the main display's when none holds its center.
func (p *Pet) screen() core.Bounds {
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

func (p *Pet) clampX(x core.Points) core.Points {
	s := p.screen()
	if s.Size.Width == 0 {
		return x
	}
	return min(max(x, s.Origin.X), s.Origin.X+s.Size.Width-p.pack.Width)
}

// PlaceAtCorner sizes the pet's window for the pet and puts it on the bottom right of the screen, above the Dock or the taskbar; without Options.Window it does nothing.
func (p *Pet) PlaceAtCorner() {
	if !p.opts.Window {
		return
	}
	s := p.screen()
	if s.Size.Width == 0 {
		return
	}
	desktop.SetWindowFrame(p.cx, desktop.WindowFrame{
		X: s.Origin.X + s.Size.Width - p.pack.Width - 24, Y: s.Origin.Y + s.Size.Height - p.pack.Height,
		Move: true, Width: p.pack.Width, Height: p.pack.Height,
	})
}

func (p *Pet) click() {
	if p.phase == phaseExit || p.phase == phaseStartup {
		return
	}
	p.move = nil
	if r, ok := p.choose("Click"); ok {
		p.play(r, phaseClick, false)
	}
}

// pan drags the pet: the pointer's place in the window is where it was pressed plus the translation, so moving the window by the translation keeps the pet under the pointer. A pet in the host's view dangles where it is.
func (p *Pet) pan(e core.PanEvent) {
	if p.phase == phaseExit {
		return
	}
	switch e.Phase {
	case core.GestureBegan:
		p.dragging = true
		p.move = nil
		p.play(p.policy.Find(p.pack.Clips.Drag), phaseDrag, false)
	case core.GestureChanged:
		if p.opts.Window {
			f := desktop.Frame(p.cx)
			desktop.SetWindowFrame(p.cx, desktop.WindowFrame{X: f.X + e.Translation.X, Y: f.Y + e.Translation.Y, Move: true})
		}
	case core.GestureEnded, core.GestureCancelled:
		p.dragging = false
		p.playIdle()
	}
}

// Act plays an event now, and the next one an event interval later.
func (p *Pet) Act() {
	if p.phase == phaseExit {
		return
	}
	if r, ok := p.choose("Event"); ok {
		p.nextEvent = time.Now().Add(p.eventInterval())
		p.startEvent(r)
	}
}

// Play plays the clip id names, which a host maps its own states to, such as a session running, waiting or done; the pet goes on by its rules once the clip ends. It reports false, playing nothing, for a clip the pack lacks or a pet leaving.
func (p *Pet) Play(id string) bool {
	r := p.policy.Find(id)
	if r.ID == "" || p.phase == phaseExit {
		return false
	}
	p.move = nil
	p.play(r, phaseEvent, false)
	return true
}

// Leave plays the pet's exit clip and runs done as it ends, or eight seconds later should it never end; done runs once, and a second Leave does nothing. The host closes the pet's window or view in done.
func (p *Pet) Leave(done func()) {
	if p.phase == phaseExit {
		return
	}
	if done == nil {
		done = func() {}
	}
	p.left = sync.OnceFunc(done)
	p.move = nil
	p.play(p.policy.Find(p.pack.Clips.Exit), phaseExit, false)
	// Leave even if the clip never ends.
	p.cx.AfterFunc(8*time.Second, p.left)
}
