package pet

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"image"
	"slices"
	"time"

	"github.com/at-wat/ebml-go"
)

// clip is a WebM clip's VP9 frames as its file stores them: each frame's color picture and, in the block's additions, its alpha picture, a second VP9 stream.
type clip struct {
	width, height int
	frames        []clipFrame
	// duration is how long the clip plays, its last frame included.
	duration time.Duration
}

// clipFrame is one frame: when it shows, and its compressed color and alpha pictures; alpha is nil for a frame without one, which is opaque.
type clipFrame struct {
	at           time.Duration
	color, alpha []byte
}

// webmFile is what demuxClip reads of a WebM file. A video frame with alpha is a BlockGroup, as only a BlockGroup has additions; one without may be a SimpleBlock.
type webmFile struct {
	Segment struct {
		Info struct {
			TimecodeScale uint64  `ebml:"TimecodeScale"`
			Duration      float64 `ebml:"Duration"`
		} `ebml:"Info"`
		Tracks struct {
			TrackEntry []struct {
				TrackNumber     uint64 `ebml:"TrackNumber"`
				TrackType       uint64 `ebml:"TrackType"`
				CodecID         string `ebml:"CodecID"`
				DefaultDuration uint64 `ebml:"DefaultDuration"`
				Video           *struct {
					PixelWidth  uint64 `ebml:"PixelWidth"`
					PixelHeight uint64 `ebml:"PixelHeight"`
				} `ebml:"Video"`
			} `ebml:"TrackEntry"`
		} `ebml:"Tracks"`
		Cluster []struct {
			Timecode    uint64       `ebml:"Timecode"`
			SimpleBlock []ebml.Block `ebml:"SimpleBlock"`
			BlockGroup  []struct {
				Block          ebml.Block `ebml:"Block"`
				BlockAdditions struct {
					BlockMore []struct {
						BlockAddID      uint64 `ebml:"BlockAddID"`
						BlockAdditional []byte `ebml:"BlockAdditional"`
					} `ebml:"BlockMore"`
				} `ebml:"BlockAdditions"`
			} `ebml:"BlockGroup"`
		} `ebml:"Cluster"`
	} `ebml:"Segment"`
}

// trackVideo is Matroska's TrackType of a video track.
const trackVideo = 1

// alphaAddID is the BlockAddID of the additions that hold a frame's alpha picture, as Matroska's AlphaMode 1 says. It is also BlockAddID's default, which a muxer may leave out; ebml-go fills in no defaults, so an absent BlockAddID reads as 0, which no BlockMore may have.
const alphaAddID = 1

// demuxClip reads the VP9 video track of a WebM file.
func demuxClip(data []byte) (*clip, error) {
	var f webmFile
	if err := ebml.Unmarshal(bytes.NewReader(data), &f, ebml.WithIgnoreUnknown(true)); err != nil {
		return nil, err
	}
	s := &f.Segment
	scale := s.Info.TimecodeScale
	if scale == 0 {
		// Matroska's default: timecodes in milliseconds.
		scale = 1_000_000
	}
	c := &clip{}
	var track, frameDuration uint64
	for _, t := range s.Tracks.TrackEntry {
		if t.TrackType == trackVideo && t.Video != nil {
			if t.CodecID != "V_VP9" {
				return nil, fmt.Errorf("the video is %s, not VP9", t.CodecID)
			}
			track, frameDuration = t.TrackNumber, t.DefaultDuration
			c.width, c.height = int(t.Video.PixelWidth), int(t.Video.PixelHeight)
			break
		}
	}
	if track == 0 {
		return nil, errors.New("no video track")
	}
	add := func(cluster uint64, b ebml.Block, alpha []byte) error {
		if b.TrackNumber != track {
			return nil
		}
		if len(b.Data) != 1 {
			return fmt.Errorf("a video block holds %d frames, not one", len(b.Data))
		}
		ticks := int64(cluster) + int64(b.Timecode)
		c.frames = append(c.frames, clipFrame{at: time.Duration(ticks * int64(scale)), color: b.Data[0], alpha: alpha})
		return nil
	}
	for _, cl := range s.Cluster {
		for _, b := range cl.SimpleBlock {
			if err := add(cl.Timecode, b, nil); err != nil {
				return nil, err
			}
		}
		for _, g := range cl.BlockGroup {
			var alpha []byte
			for _, m := range g.BlockAdditions.BlockMore {
				if m.BlockAddID == alphaAddID || m.BlockAddID == 0 {
					alpha = m.BlockAdditional
				}
			}
			if err := add(cl.Timecode, g.Block, alpha); err != nil {
				return nil, err
			}
		}
	}
	if len(c.frames) == 0 {
		return nil, errors.New("no video frames")
	}
	// A cluster's SimpleBlocks and BlockGroups come apart above; VP9 in WebM decodes in the order frames show, as its hidden frames travel inside the shown ones.
	slices.SortStableFunc(c.frames, func(a, b clipFrame) int { return cmp.Compare(a.at, b.at) })
	last := c.frames[len(c.frames)-1].at
	switch {
	case s.Info.Duration > 0:
		c.duration = time.Duration(s.Info.Duration * float64(scale))
	case frameDuration > 0:
		c.duration = last + time.Duration(frameDuration)
	case len(c.frames) > 1:
		c.duration = last + last/time.Duration(len(c.frames)-1)
	}
	c.duration = max(c.duration, last)
	return c, nil
}

// yuvPlanes is a decoded 8-bit 4:2:0 picture: y at full size, u and v at half the width and height, rounded up.
type yuvPlanes struct {
	width, height             int
	y, u, v                   []byte
	yStride, uStride, vStride int
	// bt709 picks BT.709's matrix over BT.601's, and fullRange Y and chroma over 0 to 255 instead of 16 to 235 and 240.
	bt709, fullRange bool
}

// yuvMatrix is a YUV to RGB matrix in 8.8 fixed point: y is Y's factor, rv V's in red, gu and gv U's and V's in green, bu U's in blue.
type yuvMatrix struct{ y, rv, gu, gv, bu int32 }

var (
	bt601Limited = yuvMatrix{298, 409, 100, 208, 516}
	bt709Limited = yuvMatrix{298, 459, 55, 136, 541}
	bt601Full    = yuvMatrix{256, 359, 88, 183, 454}
	bt709Full    = yuvMatrix{256, 403, 48, 120, 475}
)

// toRGBA writes the picture into dst, premultiplied as image.RGBA is, with alpha from the luma of a, the alpha stream's picture, or opaque where a is nil. dst and a are the picture's size.
func (p *yuvPlanes) toRGBA(dst *image.RGBA, a *yuvPlanes) {
	m := bt601Limited
	switch {
	case p.bt709 && p.fullRange:
		m = bt709Full
	case p.bt709:
		m = bt709Limited
	case p.fullRange:
		m = bt601Full
	}
	var black int32 = 16
	if p.fullRange {
		black = 0
	}
	w, h := dst.Rect.Dx(), dst.Rect.Dy()
	for row := range h {
		ys := p.y[row*p.yStride:]
		us := p.u[(row/2)*p.uStride:]
		vs := p.v[(row/2)*p.vStride:]
		var as []byte
		if a != nil {
			as = a.y[row*a.yStride:]
		}
		out := dst.Pix[row*dst.Stride : row*dst.Stride+4*w]
		for col := range w {
			alpha := int32(255)
			if as != nil {
				alpha = int32(as[col])
			}
			px := out[4*col : 4*col+4 : 4*col+4]
			if alpha == 0 {
				px[0], px[1], px[2], px[3] = 0, 0, 0, 0
				continue
			}
			c := m.y * (int32(ys[col]) - black)
			d := int32(us[col/2]) - 128
			e := int32(vs[col/2]) - 128
			r := clamp8((c + m.rv*e + 128) >> 8)
			g := clamp8((c - m.gu*d - m.gv*e + 128) >> 8)
			b := clamp8((c + m.bu*d + 128) >> 8)
			if alpha < 255 {
				r, g, b = (r*alpha+127)/255, (g*alpha+127)/255, (b*alpha+127)/255
			}
			px[0], px[1], px[2], px[3] = uint8(r), uint8(g), uint8(b), uint8(alpha)
		}
	}
}

func clamp8(v int32) int32 { return min(max(v, 0), 255) }
