package pet

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/at-wat/ebml-go"
	"github.com/nsgo-dev/nsgo/core"
	"github.com/nsgo-dev/nsgo/nsgotest"
)

// The tests run in the engine's folder, which ships no clips: they play nsgo-pet's, and on Windows its libvpx, from the app's resources beside it rather than beside the test's executable.
func init() { resourceRoot = filepath.Join("..", "resources") }

// appPets is where nsgo-pet's pets are.
var appPets = filepath.Join("..", "resources", "pets")

func readClip(t *testing.T, path string) *clip {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	c, err := demuxClip(data)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return c
}

// Big Blue Fish's clips are 640 by 360 at 24 frames a second, each frame with its alpha picture, and play for their Duration.
func TestDemuxBigBlueFish(t *testing.T) {
	c := readClip(t, filepath.Join(appPets, "bigbluefish", "webm", "待机呼吸休闲.webm"))
	if c.width != 640 || c.height != 360 {
		t.Fatalf("size %dx%d, want 640x360", c.width, c.height)
	}
	if len(c.frames) < 24 {
		t.Fatalf("%d frames", len(c.frames))
	}
	for i, f := range c.frames {
		if len(f.color) == 0 || len(f.alpha) == 0 {
			t.Fatalf("frame %d: %d bytes of color, %d of alpha", i, len(f.color), len(f.alpha))
		}
		if i > 0 && f.at <= c.frames[i-1].at {
			t.Fatalf("frame %d at %v after frame %d at %v", i, f.at, i-1, c.frames[i-1].at)
		}
	}
	if c.frames[0].at != 0 {
		t.Fatalf("the first frame shows at %v", c.frames[0].at)
	}
	step := c.frames[1].at - c.frames[0].at
	if step < 41*time.Millisecond || step > 42*time.Millisecond {
		t.Fatalf("frames %v apart, want 1/24 s", step)
	}
	last := c.frames[len(c.frames)-1].at
	if c.duration < last || c.duration > last+2*step {
		t.Fatalf("duration %v with the last frame at %v", c.duration, last)
	}
}

// Every one of nsgo-pet's WebM clips demuxes, so the Windows player can open all of them.
func TestDemuxEveryClip(t *testing.T) {
	packs, err := LoadDir(appPets)
	if err != nil {
		t.Fatal(err)
	}
	for _, pk := range packs {
		for _, r := range pk.Policy.Rules {
			c := readClip(t, clipPath(pk.Base+"/webm/"+r.File))
			if c.width <= 0 || c.height <= 0 || c.duration <= 0 {
				t.Errorf("%s %s: %dx%d for %v", pk.ID, r.ID, c.width, c.height, c.duration)
			}
			for i, f := range c.frames {
				if f.alpha == nil {
					t.Errorf("%s %s: frame %d has no alpha", pk.ID, r.ID, i)
					break
				}
			}
		}
	}
}

// A BlockMore without a BlockAddID holds the alpha picture, as BlockAddID's default is 1: muxers that leave out elements at their default write it so.
func TestDemuxAlphaWithoutBlockAddID(t *testing.T) {
	type blockMore struct {
		BlockAdditional []byte `ebml:"BlockAdditional"`
	}
	type blockGroup struct {
		Block          ebml.Block `ebml:"Block"`
		BlockAdditions struct {
			BlockMore []blockMore `ebml:"BlockMore"`
		} `ebml:"BlockAdditions"`
	}
	type video struct {
		PixelWidth  uint64 `ebml:"PixelWidth"`
		PixelHeight uint64 `ebml:"PixelHeight"`
	}
	type trackEntry struct {
		TrackNumber uint64 `ebml:"TrackNumber"`
		TrackType   uint64 `ebml:"TrackType"`
		CodecID     string `ebml:"CodecID"`
		Video       video  `ebml:"Video"`
	}
	type cluster struct {
		Timecode   uint64       `ebml:"Timecode"`
		BlockGroup []blockGroup `ebml:"BlockGroup"`
	}
	var f struct {
		Segment struct {
			Tracks struct {
				TrackEntry []trackEntry `ebml:"TrackEntry"`
			} `ebml:"Tracks"`
			Cluster []cluster `ebml:"Cluster"`
		} `ebml:"Segment"`
	}
	f.Segment.Tracks.TrackEntry = []trackEntry{{TrackNumber: 1, TrackType: trackVideo, CodecID: "V_VP9", Video: video{PixelWidth: 2, PixelHeight: 2}}}
	g := blockGroup{Block: ebml.Block{TrackNumber: 1, Keyframe: true, Data: [][]byte{{0xc0}}}}
	g.BlockAdditions.BlockMore = []blockMore{{BlockAdditional: []byte{0xa0}}}
	f.Segment.Cluster = []cluster{{BlockGroup: []blockGroup{g}}}
	var buf bytes.Buffer
	if err := ebml.Marshal(&f, &buf); err != nil {
		t.Fatal(err)
	}
	c, err := demuxClip(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(c.frames) != 1 || !bytes.Equal(c.frames[0].alpha, []byte{0xa0}) {
		t.Fatalf("frames %+v, want one with its alpha", c.frames)
	}
}

// planes makes a 2 by 2 picture of one Y, U and V.
func planes(y, u, v byte, fullRange bool) *yuvPlanes {
	return &yuvPlanes{width: 2, height: 2, y: []byte{y, y, y, y}, u: []byte{u}, v: []byte{v}, yStride: 2, uStride: 1, vStride: 1, fullRange: fullRange}
}

func alphaPlanes(a ...byte) *yuvPlanes {
	return &yuvPlanes{width: 2, height: 2, y: a, yStride: 2}
}

func TestToRGBA(t *testing.T) {
	for _, c := range []struct {
		name  string
		p, a  *yuvPlanes
		want0 [4]uint8
		want3 [4]uint8
	}{
		{"limited white, opaque", planes(235, 128, 128, false), nil, [4]uint8{255, 255, 255, 255}, [4]uint8{255, 255, 255, 255}},
		{"limited black", planes(16, 128, 128, false), alphaPlanes(255, 255, 255, 255), [4]uint8{0, 0, 0, 255}, [4]uint8{0, 0, 0, 255}},
		{"full white", planes(255, 128, 128, true), nil, [4]uint8{255, 255, 255, 255}, [4]uint8{255, 255, 255, 255}},
		// BT.601 red is Y 81, U 90, V 240 in limited range.
		{"red, then transparent", planes(81, 90, 240, false), alphaPlanes(255, 255, 255, 0), [4]uint8{255, 0, 0, 255}, [4]uint8{0, 0, 0, 0}},
		// Half-transparent white is premultiplied.
		{"half white", planes(235, 128, 128, false), alphaPlanes(128, 128, 128, 128), [4]uint8{128, 128, 128, 128}, [4]uint8{128, 128, 128, 128}},
	} {
		dst := image.NewRGBA(image.Rect(0, 0, 2, 2))
		c.p.toRGBA(dst, c.a)
		got0, got3 := [4]uint8(dst.Pix[0:4]), [4]uint8(dst.Pix[12:16])
		if far(got0, c.want0) || far(got3, c.want3) {
			t.Errorf("%s: %v and %v, want %v and %v", c.name, got0, got3, c.want0, c.want3)
		}
	}
}

// far reports whether a and b differ by more than rounding.
func far(a, b [4]uint8) bool {
	for i := range a {
		if d := int(a[i]) - int(b[i]); d < -2 || d > 2 {
			return true
		}
	}
	return false
}

// A real clip decodes with libvpx into a picture whose corners are transparent and whose middle, where the pet is, is not. It needs libvpx: on Windows nsgo-pet's DLL in its resources/windows, elsewhere the system's.
func TestDecodeClip(t *testing.T) {
	if err := loadVPX(); err != nil {
		t.Skip(err)
	}
	c := readClip(t, filepath.Join(appPets, "bigbluefish", "webm", "待机呼吸休闲.webm"))
	color, err := newVP9Decoder()
	if err != nil {
		t.Fatal(err)
	}
	defer color.close()
	alpha, err := newVP9Decoder()
	if err != nil {
		t.Fatal(err)
	}
	defer alpha.close()
	dst := image.NewRGBA(image.Rect(0, 0, c.width, c.height))
	for i, f := range c.frames {
		p, ok, err := color.decode(f.color)
		if err != nil || !ok {
			t.Fatalf("frame %d: %v, %v", i, ok, err)
		}
		a, ok, err := alpha.decode(f.alpha)
		if err != nil || !ok {
			t.Fatalf("frame %d's alpha: %v, %v", i, ok, err)
		}
		if p.width != c.width || p.height != c.height || a.width != c.width || a.height != c.height {
			t.Fatalf("frame %d is %dx%d with alpha %dx%d", i, p.width, p.height, a.width, a.height)
		}
		p.toRGBA(dst, &a)
	}
	// The alpha stream is lossy, so transparent parts may keep an alpha of a few levels.
	if corner := dst.RGBAAt(0, 0); corner.A > 8 {
		t.Errorf("the corner is %v, want transparent", corner)
	}
	opaque := 0
	for y := c.height / 4; y < 3*c.height/4; y++ {
		for x := c.width / 3; x < 2*c.width/3; x++ {
			if dst.RGBAAt(x, y).A == 255 {
				opaque++
			}
		}
	}
	if opaque < c.width*c.height/30 {
		t.Errorf("%d opaque pixels in the middle, want the fish", opaque)
	}
}

// BenchmarkDecodeFrame measures one of Big Blue Fish's 640 by 360 frames: decoding its color and alpha pictures and writing them as RGBA, what the Windows player does 24 times a second.
func BenchmarkDecodeFrame(b *testing.B) {
	if err := loadVPX(); err != nil {
		b.Skip(err)
	}
	data, err := os.ReadFile(filepath.Join(appPets, "bigbluefish", "webm", "待机呼吸休闲.webm"))
	if err != nil {
		b.Fatal(err)
	}
	c, err := demuxClip(data)
	if err != nil {
		b.Fatal(err)
	}
	dst := image.NewRGBA(image.Rect(0, 0, c.width, c.height))
	var color, alpha *vp9Decoder
	for i := 0; b.Loop(); i++ {
		n := i % len(c.frames)
		if n == 0 {
			color.close()
			alpha.close()
			if color, err = newVP9Decoder(); err != nil {
				b.Fatal(err)
			}
			if alpha, err = newVP9Decoder(); err != nil {
				b.Fatal(err)
			}
		}
		p, _, err := color.decode(c.frames[n].color)
		if err != nil {
			b.Fatal(err)
		}
		a, _, err := alpha.decode(c.frames[n].alpha)
		if err != nil {
			b.Fatal(err)
		}
		p.toRGBA(dst, &a)
	}
	color.close()
	alpha.close()
}

var windows = nsgotest.Environment(core.Environment{OS: core.OSWindows, Idiom: core.IdiomDesktop})

// appPack is one of nsgo-pet's pets, loaded from its directory, so the Windows player reads its clips from their files.
func appPack(t *testing.T, id string) *Pack {
	t.Helper()
	packs, err := LoadDir(appPets)
	for _, p := range packs {
		if p.ID == id {
			return p
		}
	}
	t.Fatalf("no %s pet: %v", id, err)
	return nil
}

// showing is whether the slot id shows a decoded picture.
func showing(app *nsgotest.App, id string) bool {
	v := app.FindMounted("id:" + id)
	return v.Image.Source.Pixels != nil && v.Style.Opacity == 1
}

// On Windows the clips play in the engine's own decoder: the greeting's first picture shows in an image, and once the greeting ends an idle clip takes over in the other slot. It needs libvpx, as TestDecodeClip does.
func TestWindowsPlaysClipsItself(t *testing.T) {
	if err := loadVPX(); err != nil {
		t.Skip(err)
	}
	pack := appPack(t, "nsgo-chan")
	app := nsgotest.Start(t, func(cx *core.Scope) core.Node {
		p := New(cx, pack, Options{})
		cx.Post(p.Start)
		return p.View()
	}, windows)
	app.WaitUntil(10*time.Second, func() bool { return showing(app, "slot1") })
	if showing(app, "slot0") {
		t.Fatal("both slots show")
	}
	// The fake clock moves on while the decoder keeps up in real time.
	deadline := time.Now().Add(30 * time.Second)
	for !showing(app, "slot0") {
		if time.Now().After(deadline) {
			t.Fatal("no idle clip after the greeting")
		}
		app.Advance(40 * time.Millisecond)
	}
	if showing(app, "slot1") {
		t.Fatal("the greeting still shows over the idle clip")
	}
}

// A host switching pets on Windows, as nsgo-pet's settings do, starts the new pet's own players, which show its greeting.
func TestWindowsSwitchesPets(t *testing.T) {
	if err := loadVPX(); err != nil {
		t.Skip(err)
	}
	nsgoChan, fish := appPack(t, "nsgo-chan"), appPack(t, "bigbluefish")
	var chosen *core.Signal[*Pack]
	app := nsgotest.Start(t, func(cx *core.Scope) core.Node {
		chosen = core.NewSignal(cx, nsgoChan)
		return core.Dynamic(func(cx *core.Scope) core.Node {
			p := New(cx, chosen.Get(), Options{})
			cx.Post(p.Start)
			return p.View()
		})
	}, windows)
	app.WaitUntil(10*time.Second, func() bool { return showing(app, "slot1") })
	app.Post(func() { chosen.Set(fish) })
	app.WaitUntil(10*time.Second, func() bool {
		v := app.FindMounted("id:slot1")
		return v.Props.Label == fish.Name && showing(app, "slot1") && v.Image.Source.Pixels.Rect.Dx() == 640
	})
}

// A clip chosen again while the hidden slot still holds it shows only once Windows' player has decoded it anew, not with the frame the slot last showed.
func TestWindowsReplaysAClipFromItsStart(t *testing.T) {
	if err := loadVPX(); err != nil {
		t.Skip(err)
	}
	pack := appPack(t, "nsgo-chan")
	var p *Pet
	app := nsgotest.Start(t, func(cx *core.Scope) core.Node {
		p = New(cx, pack, Options{})
		cx.Post(p.Start)
		return p.View()
	}, windows)
	app.WaitUntil(10*time.Second, func() bool { return showing(app, "slot1") })
	app.Post(func() { p.Play("idle") })
	app.WaitUntil(10*time.Second, func() bool { return showing(app, "slot0") && !showing(app, "slot1") })
	stale := app.FindMounted("id:slot1").Image.Source.Version
	app.Post(func() { p.Play(pack.Clips.Greeting) })
	app.Settle()
	if v := app.FindMounted("id:slot1"); v.Style.Opacity == 1 && v.Image.Source.Version == stale {
		t.Fatal("the slot shows the greeting's old frame")
	}
	app.WaitUntil(10*time.Second, func() bool {
		v := app.FindMounted("id:slot1")
		return showing(app, "slot1") && v.Image.Source.Version > stale
	})
	if showing(app, "slot0") {
		t.Fatal("both slots show")
	}
}
