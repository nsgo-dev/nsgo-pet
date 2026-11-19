package main

import (
	"context"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nsgo-dev/nsgo/core"
)

// NSGo refuses alpha video on Windows (internal/platform/windows/video.go): whether Media Foundation keeps the alpha of VP9 in WebM was never checked there, so its player reports VideoErrorAlphaUnsupported and asks the app to decode such clips itself. clipPlayer does that for the pet on Windows: it demuxes the WebM clip, decodes its color and alpha streams with libvpx on a goroutine of its own, and shows each frame in a core.Image through ImageSource.Pixels at the frame's time. It gives slot.view what core.Video gives it elsewhere: a load once the first picture is up, progress, the end, pausing and restarting.
// TODO(nsgo#862): once NSGo plays VP9-with-alpha WebM on Windows, go back to core.Video there and drop clipPlayer, libvpx and ebml-go https://github.com/nsgo-dev/nsgo/issues/862
type clipPlayer struct {
	cx *core.Scope
	// version counts the frames shown, the image's Version of shown.
	version *core.Signal[uint64]

	onLoad     func()
	onProgress func(position time.Duration)
	onEnd      func()
	onError    func(error)

	source string
	// load counts the loads, so a frame a stopped decoder posted late is dropped.
	load   int
	cancel context.CancelFunc
	// free are the buffers the decoder may write the next frames into, ready the frames it decoded in order and that wait for their time.
	free  chan *image.RGBA
	ready []readyFrame
	// shown is the buffer on screen, given back to the decoder once the next frame shows: the core copies a buffer in the commit that carries it, and a later frame replaces one that has not been committed yet.
	shown     *image.RGBA
	shownLoad int

	loaded, ended, decoded bool
	duration               time.Duration
	// playing is whether the clock runs: from start, the graph time of the clip's position 0, else held at position.
	paused, playing bool
	start           time.Time
	position        time.Duration
	lastProgress    time.Duration
	stopTimer       func() bool
}

// readyFrame is a decoded frame and when it shows.
type readyFrame struct {
	at  time.Duration
	buf *image.RGBA
}

// clipBuffers is how many frames the decoder may be ahead of the screen, the one shown included.
const clipBuffers = 3

// clipProgressInterval is how often a playing clip reports its position, as slot.view asks of core.Video.
const clipProgressInterval = 33 * time.Millisecond

// clipView is slot.view's player on Windows: an image, sized and styled as the video would be, that plays the clip src names and that the slot restarts through *player.
type clipView struct {
	src        func() string
	paused     func() bool
	style      func() core.Style
	onLoad     func()
	onProgress func(position time.Duration)
	onEnd      func()
	onError    func(error)
	player     **clipPlayer
	label      string
	testID     string
}

// Render implements core.Node.
func (v clipView) Render(cx *core.Scope) core.Node {
	pl := &clipPlayer{cx: cx, version: core.NewSignal(cx, uint64(0)), onLoad: v.onLoad, onProgress: v.onProgress, onEnd: v.onEnd, onError: v.onError}
	*v.player = pl
	cx.Effect(func() { pl.open(v.src()) })
	cx.Effect(func() { pl.setPaused(v.paused()) })
	cx.Cleanup(pl.stop)
	frame := func() core.ImageSource {
		v := pl.version.Get()
		if pl.shown == nil {
			return core.ImageSource{}
		}
		return core.ImageSource{Pixels: pl.shown, Version: v}
	}
	return core.Image{SourceFunc: frame, ResizeMode: core.Contain, StyleFunc: v.style, Label: v.label, TestID: v.testID}
}

// open starts playing source from its start, keeping the frame on screen until the first of source's shows.
func (pl *clipPlayer) open(source string) {
	pl.stop()
	pl.source = source
	pl.loaded, pl.ended, pl.decoded, pl.playing = false, false, false, false
	// The first tick reports the start.
	pl.position, pl.lastProgress, pl.duration = 0, -clipProgressInterval, 0
	pl.load++
	if source == "" {
		return
	}
	load := pl.load
	ctx, cancel := context.WithCancel(context.Background())
	pl.cancel = cancel
	pl.free = make(chan *image.RGBA, clipBuffers)
	go pl.decode(ctx, load, pl.free, clipPath(source))
}

// restart plays the clip again from its start, as a seek to 0 does a video's.
func (pl *clipPlayer) restart() { pl.open(pl.source) }

// stop stops the decoder and the clock; the frame on screen stays.
func (pl *clipPlayer) stop() {
	if pl.cancel != nil {
		pl.cancel()
		pl.cancel = nil
	}
	if pl.stopTimer != nil {
		pl.stopTimer()
		pl.stopTimer = nil
	}
	pl.ready = nil
}

func (pl *clipPlayer) setPaused(paused bool) {
	if paused == pl.paused {
		return
	}
	pl.paused = paused
	if paused {
		pl.hold()
		return
	}
	pl.run()
}

// run starts the clock from the position it holds, once the first frame is up and while not paused.
func (pl *clipPlayer) run() {
	if pl.playing || pl.paused || !pl.loaded || pl.ended {
		return
	}
	pl.playing = true
	pl.start = pl.cx.Now().Add(-pl.position)
	// run may be called from the effect that follows paused: the first tick, which may call the slot's handlers and arms the timer, runs under the player's own scope instead, which keeps the timer when that effect runs again.
	pl.cx.Post(pl.tick)
}

// hold stops the clock at the position it reached.
func (pl *clipPlayer) hold() {
	if !pl.playing {
		return
	}
	pl.position = pl.now()
	pl.playing = false
	if pl.stopTimer != nil {
		pl.stopTimer()
		pl.stopTimer = nil
	}
}

// now is the clip's position.
func (pl *clipPlayer) now() time.Duration {
	if !pl.playing {
		return pl.position
	}
	return pl.cx.Now().Sub(pl.start)
}

// decode runs on a goroutine of its own: it reads and demuxes the clip, then decodes its frames in order into the free buffers and posts each to the graph, until the clip ends or ctx is cancelled.
func (pl *clipPlayer) decode(ctx context.Context, load int, free chan *image.RGBA, path string) {
	fail := func(err error) {
		pl.cx.Post(func() {
			if load == pl.load {
				pl.onError(fmt.Errorf("%s: %w", filepath.Base(path), err))
			}
		})
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fail(err)
		return
	}
	c, err := demuxClip(data)
	if err != nil {
		fail(err)
		return
	}
	color, err := newVP9Decoder()
	if err != nil {
		fail(err)
		return
	}
	defer color.close()
	alpha, err := newVP9Decoder()
	if err != nil {
		fail(err)
		return
	}
	defer alpha.close()
	for range clipBuffers {
		free <- image.NewRGBA(image.Rect(0, 0, c.width, c.height))
	}
	for i, f := range c.frames {
		var buf *image.RGBA
		select {
		case buf = <-free:
		case <-ctx.Done():
			return
		}
		p, ok, err := color.decode(f.color)
		if err != nil {
			fail(fmt.Errorf("frame %d: %w", i, err))
			return
		}
		var a *yuvPlanes
		if f.alpha != nil {
			ap, aok, err := alpha.decode(f.alpha)
			if err != nil {
				fail(fmt.Errorf("frame %d's alpha: %w", i, err))
				return
			}
			if aok && ap.width >= c.width && ap.height >= c.height {
				a = &ap
			}
		}
		if !ok || p.width < c.width || p.height < c.height {
			free <- buf
			continue
		}
		p.toRGBA(buf, a)
		at := f.at
		pl.cx.Post(func() { pl.deliver(load, readyFrame{at: at, buf: buf}) })
	}
	pl.cx.Post(func() {
		if load == pl.load {
			pl.decoded, pl.duration = true, c.duration
			pl.tick()
		}
	})
}

// deliver takes a decoded frame on the graph: the first shows at once and loads the clip; the others wait for their time.
func (pl *clipPlayer) deliver(load int, f readyFrame) {
	if load != pl.load {
		return
	}
	if !pl.loaded {
		pl.show(load, f.buf)
		pl.loaded = true
		pl.position = f.at
		pl.onLoad()
		pl.run()
		return
	}
	pl.ready = append(pl.ready, f)
	// A frame that arrives while none waited may be due before the next report.
	if pl.playing && len(pl.ready) == 1 {
		pl.tick()
	}
}

// show puts buf, a frame of load, on screen, and gives the buffer it replaces back to the decoder when that buffer is of the same load: one of an earlier load is left to the collector.
func (pl *clipPlayer) show(load int, buf *image.RGBA) {
	if old := pl.shown; old != nil && pl.shownLoad == load {
		pl.free <- old
	}
	pl.shown, pl.shownLoad = buf, load
	pl.version.Set(pl.version.GetUntracked() + 1)
}

// tick shows the frames that are due, reports the position and the end, and arms the timer for the next frame or report.
func (pl *clipPlayer) tick() {
	if pl.stopTimer != nil {
		pl.stopTimer()
		pl.stopTimer = nil
	}
	if !pl.playing {
		return
	}
	pos := pl.now()
	// A decoder that falls behind skips to the latest frame that is due.
	for len(pl.ready) > 0 && pl.ready[0].at <= pos {
		f := pl.ready[0]
		pl.ready = pl.ready[1:]
		pl.show(pl.load, f.buf)
	}
	if pl.decoded && len(pl.ready) == 0 && pos >= pl.duration {
		pl.position, pl.playing, pl.ended = pl.duration, false, true
		pl.onProgress(pl.duration)
		pl.onEnd()
		return
	}
	if pos-pl.lastProgress >= clipProgressInterval {
		pl.lastProgress = pos
		pl.onProgress(pos)
	}
	next := pl.lastProgress + clipProgressInterval
	if len(pl.ready) > 0 {
		next = min(next, pl.ready[0].at)
	} else if pl.decoded {
		next = min(next, pl.duration)
	}
	pl.stopTimer = pl.cx.AfterFunc(max(next-pos, time.Millisecond), func() {
		pl.stopTimer = nil
		pl.tick()
	})
}

// resourceRoot is the folder NSGo resolves "asset:" paths against on Windows and Linux, the executable's, where its packages put the app's resources; the tests point it at resources.
var resourceRoot = func() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}()

// clipPath is where an "asset:" clip is on disk.
func clipPath(source string) string {
	rel, _ := strings.CutPrefix(source, "asset:")
	return filepath.Join(resourceRoot, filepath.FromSlash(rel))
}
