package gif

import (
	"bufio"
	"errors"
	"image"
	"image/color"
	"io"
)

type preInitEntry struct {
	pm       *image.Paletted
	delay    int
	disposal uint8
}

type StreamerOptions struct {
	// LoopCount controls the number of times an animation will be
	// restarted during display.
	// A LoopCount of 0 means to loop forever.
	// A LoopCount of -1 means to show each frame only once.
	// Otherwise, the animation is looped LoopCount+1 times.
	LoopCount int

	// Config is the global color table (palette), width and height. A nil or
	// empty-color.Palette Config.ColorModel means that each frame has its own
	// color table and there is no global color table. Each frame's bounds must
	// be within the rectangle defined by the two points (0, 0) and
	// (Config.Width, Config.Height).
	//
	// For backwards compatibility, a zero-valued Config is valid to pass to
	// EncodeAll, and implies that the overall GIF's width and height equals
	// the first frame's bounds' Rectangle.Max point.
	Config image.Config

	// BackgroundIndex is the background index in the global color table, for
	// use with the DisposalBackground disposal method.
	BackgroundIndex byte
}

// Streamer allows encoding and writing images
// to w one at a time without accumulating
// an array of images. Use this to reduce memory
// consumption.
type Streamer struct {
	e *encoder
	w io.Writer

	initialized bool
	closed      bool
	index       int

	preInitEntries []*preInitEntry
}

// Creates a streamer.
func NewStreamer(w io.Writer, options *StreamerOptions) *Streamer {
	g := GIF{
		LoopCount:       options.LoopCount,
		Config:          options.Config,
		BackgroundIndex: options.BackgroundIndex,
	}
	return &Streamer{
		w: w,
		e: &encoder{g: g},
	}
}

func (s *Streamer) init() error {
	e := s.e

	if len(s.preInitEntries) != 0 {
		pm := s.preInitEntries[0].pm
		if e.g.Config == (image.Config{}) {
			p := pm.Bounds().Max
			e.g.Config.Width = p.X
			e.g.Config.Height = p.Y
		} else if e.g.Config.ColorModel != nil {
			if _, ok := e.g.Config.ColorModel.(color.Palette); !ok {
				return errors.New("gif: GIF color model must be a color.Palette")
			}
		}
	}

	if ww, ok := s.w.(writer); ok {
		e.w = ww
	} else {
		e.w = bufio.NewWriter(s.w)
	}

	if len(e.g.Image) < len(s.preInitEntries) {
		e.g.Image = make([]*image.Paletted, len(s.preInitEntries))
	}

	e.writeHeader()
	s.initialized = true

	for _, entry := range s.preInitEntries {
		s.encode(entry.pm, entry.delay, entry.disposal)
	}

	s.preInitEntries = nil
	e.g.Image = nil

	return nil
}

// Encode writes the Image m to w in GIF format.
// Note: not concurrent-safe
func (s *Streamer) Encode(pm *image.Paletted, delay int, disposal uint8) error {
	if s.closed {
		return errors.New("gif: streamer is already closed")
	}

	if s.initialized {
		s.encode(pm, delay, disposal)
		return nil
	}

	s.preInitEntries = append(s.preInitEntries, &preInitEntry{pm, delay, disposal})
	if len(s.preInitEntries) < 2 {
		return nil
	}

	return s.init()
}

func (s *Streamer) encode(pm *image.Paletted, delay int, disposal uint8) {
	e := s.e
	e.writeImageBlock(pm, delay, disposal)
	s.index++
}

// Closes the streamer. Returns an error if
// if streamer.Encode is not called at least once,
// or if config is invalid.
// Note: not concurrent-safe
func (s *Streamer) Close() error {
	if s.closed {
		return nil
	}
	if !s.initialized {
		if len(s.preInitEntries) == 0 {
			return errors.New("gif: must provide at least one image")
		}
		if err := s.init(); err != nil {
			return err
		}
	}

	e := s.e
	e.writeByte(sTrailer)
	e.flush()

	s.closed = true
	return e.err
}
