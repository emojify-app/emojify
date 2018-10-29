package logic

import (
	"image"
	"image/draw"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/machinebox/sdk-go/facebox"
	"github.com/nfnt/resize"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Emojify finds faces in an image and replaces them with Emoji
type Emojify interface {
	// Emojimise replaces faces in the image with random emoji
	Emojimise(image.Image, []facebox.Face) (image.Image, error)
	// GetFaces returns locations of faces in an image
	GetFaces(f io.ReadSeeker) ([]facebox.Face, error)
}

// EmojifyImpl is a concrete implementation of the Emojify interface
type EmojifyImpl struct {
	emojis     []image.Image
	faceboxURI string
}

// NewEmojify creates a new implementation of the Emojify
func NewEmojify(faceboxURI, imagePath string) Emojify {
	emojis := loadEmojis(imagePath)
	return &EmojifyImpl{
		emojis:     emojis,
		faceboxURI: faceboxURI,
	}
}

// Emojimise replaces faces in the image with random emoji
func (e *EmojifyImpl) Emojimise(src image.Image, faces []facebox.Face) (image.Image, error) {
	dstImage := image.NewRGBA(src.Bounds())
	draw.Draw(dstImage, src.Bounds(), src, image.ZP, draw.Src)

	for _, face := range faces {
		m := resize.Resize(uint(face.Rect.Height), uint(face.Rect.Width), e.randomEmoji(), resize.Lanczos3)
		sp2 := image.Point{face.Rect.Left, face.Rect.Top}
		r2 := image.Rectangle{sp2, sp2.Add(m.Bounds().Size())}

		draw.Draw(
			dstImage,
			r2,
			m,
			image.ZP,
			draw.Over)
	}
	return dstImage, nil
}

// GetFaces returns the locations of faces in an image
func (e *EmojifyImpl) GetFaces(r io.ReadSeeker) ([]facebox.Face, error) {
	_, err := r.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, err
	}

	fb := facebox.New(e.faceboxURI)
	fb.HTTPClient.Timeout = 30000 * time.Millisecond

	return fb.Check(r)
}

func (e *EmojifyImpl) randomEmoji() image.Image {
	return e.emojis[rand.Intn(len(e.emojis))]
}

func loadEmojis(path string) []image.Image {
	images := make([]image.Image, 0)
	root := path

	filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			reader, err := os.Open(root + f.Name())
			if err != nil {
				return err
			}
			defer reader.Close()

			i, _, err := image.Decode(reader)
			if err == nil {
				images = append(images, i)
			}
		}

		return nil
	})

	return images
}
