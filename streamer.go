package gif

import (
	"bufio"
	"errors"
	"image"
	"image/color"
	"io"
)

type preInitImage struct {
	pm       *image.Paletted
	delay    int
	disposal uint8
}

type StreamEncoderOptions struct {
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

// StreamEncoder allows encoding and writing images
// to w one at a time without accumulating
// an array of images. Use this to reduce memory
// consumption.
type StreamEncoder struct {
	*encoder

	initialized bool
	closed      bool

	// Used to store the first two images
	// before initializing, since writeHeader
	// won't write the animation info if
	// there's only one image, and even with streaming,
	// it's possible that there's only one image encoded.
	preInitImages []*preInitImage
}

// Creates a streamer.
func NewStreamEncoder(w io.Writer, options *StreamEncoderOptions) *StreamEncoder {
	g := GIF{
		LoopCount:       options.LoopCount,
		Config:          options.Config,
		BackgroundIndex: options.BackgroundIndex,
	}
	e := &StreamEncoder{
		encoder: &encoder{
			g: g,
		},
	}
	if ww, ok := w.(writer); ok {
		e.w = ww
	} else {
		e.w = bufio.NewWriter(w)
	}
	return e
}

func (e *StreamEncoder) init() error {
	if len(e.preInitImages) != 0 {
		pm := e.preInitImages[0].pm
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

	// writeHeader() only checks the length of g.Image to
	// decide if the animation info should be written,
	// so it's assigned an slice of size 2, then disposed immediately.
	// There are surely better ways to do this, but I'd
	// rather not touch the original encoder code.
	e.g.Image = make([]*image.Paletted, len(e.preInitImages))
	e.writeHeader()
	e.g.Image = nil

	for _, img := range e.preInitImages {
		e.encode(img.pm, img.delay, img.disposal)
	}

	e.preInitImages = nil
	e.initialized = true
	return nil
}

// Encode writes the Image m to w in GIF format.
// Note: not concurrent-safe
func (e *StreamEncoder) Encode(pm *image.Paletted, delay int, disposal uint8) error {
	if e.closed {
		return errors.New("gif: streamer is already closed")
	}

	if e.initialized {
		e.encode(pm, delay, disposal)
		return nil
	}

	e.preInitImages = append(e.preInitImages, &preInitImage{pm, delay, disposal})
	if len(e.preInitImages) >= 2 {
		return e.init()
	}
	return nil
}

func (e *StreamEncoder) encode(pm *image.Paletted, delay int, disposal uint8) {
	e.writeImageBlock(pm, delay, disposal)
}

// Closes the streamer, which finalizes the write to w.
// Returns an error if streamer.Encode() is not called at least once,
// or if config is invalid.
// Note: not concurrent-safe
func (e *StreamEncoder) Close() error {
	if e.closed {
		return errors.New("gif: streamer is already closed")
	}
	if !e.initialized {
		if len(e.preInitImages) == 0 {
			return errors.New("gif: must provide at least one image")
		}
		if err := e.init(); err != nil {
			return err
		}
	}

	e.writeByte(sTrailer)
	e.flush()

	e.closed = true
	return e.err
}
