package thumbnailer

// #include "thumbnailer.h"
import "C"
import (
	"errors"
	"image"
	"io"
	"unsafe"

	"github.com/nfnt/resize"
)

var (
	// ErrCantThumbnail denotes the input file was valid but no thumbnail could
	// be generated for it (example: audio file with no cover art).
	ErrCantThumbnail = errors.New("thumbnail can't be generated")

	// ErrGetFrame denotes an unknown failure to retrieve a video frame
	ErrGetFrame = errors.New("failed to get frame")
)

// Thumbnail generates a thumbnail from a representative frame of the media.
// Images count as one frame media.
func (c *FFContext) Thumbnail(dims Dims) (thumb image.Image, err error) {
	ci, err := c.codecContext(FFVideo)
	if err != nil {
		return
	}

	var img C.struct_Buffer
	defer func() {
		if img.data != nil {
			C.free(unsafe.Pointer(img.data))
		}
	}()
	ret := C.extract_image(&img, c.avFormatCtx, ci.ctx, ci.stream)
	switch {
	case ret != 0:
		err = ffError(ret)
		return
	case img.data == nil:
		err = ErrGetFrame
		return
	}

	thumb = resize.Thumbnail(dims.Width, dims.Height,
		&image.RGBA{
			Pix:    copyCBuffer(img),
			Stride: 4 * int(img.width),
			Rect: image.Rectangle{
				Max: image.Point{
					X: int(img.width),
					Y: int(img.height),
				},
			},
		},
		resize.NearestNeighbor)
	return
}

func processMedia(rs io.ReadSeeker, src *Source, opts Options,
) (
	thumb image.Image, err error,
) {
	_, err = rs.Seek(0, 0)
	if err != nil {
		return
	}
	c, err := NewFFContext(rs)
	if err != nil {
		return
	}
	defer c.Close()

	// TODO: EXIF orientation

	src.Length = c.Length()
	src.Meta = c.Meta()
	src.HasAudio, err = c.HasStream(FFAudio)
	if err != nil {
		return
	}
	src.HasVideo, err = c.HasStream(FFVideo)
	if err != nil {
		return
	}
	if c.HasCoverArt() {
		thumb, err = processCoverArt(c.CoverArt(), opts)
	} else {
		if src.HasVideo {
			src.Dims, err = c.Dims()
			if err != nil {
				return
			}
			max := opts.MaxSourceDims
			if max.Width != 0 && src.Width > max.Width {
				err = ErrTooWide
				return
			}
			if max.Height != 0 && src.Height > max.Height {
				err = ErrTooTall
				return
			}

			thumb, err = c.Thumbnail(opts.ThumbDims)
		} else {
			err = ErrCantThumbnail
		}
	}
	return
}
