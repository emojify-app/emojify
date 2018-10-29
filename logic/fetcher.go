package logic

import (
	"bytes"
	"image"

	// import images for decoding
	_ "image/jpeg"
	_ "image/png"

	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// MaxFileSize is the maximum allowable image size
const MaxFileSize = 4000000 // 4MB

// Fetcher defines an interface for retrieving an image from a url
type Fetcher interface {
	FetchImage(uri string) (io.ReadSeeker, error)
	ReaderToImage(r io.ReadSeeker) (image.Image, error)
}

// FetcherImpl is the implementation of the Fetcher interface
type FetcherImpl struct {
	MaxFileSize int
}

// FetchImage downloads an image and returns a ReadSeeker
func (f *FetcherImpl) FetchImage(uri string) (io.ReadSeeker, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r := io.LimitReader(resp.Body, MaxFileSize)
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buf), nil
}

// ReaderToImage reads the contents of a ReadSeeker and returns a golang image
func (f *FetcherImpl) ReaderToImage(r io.ReadSeeker) (image.Image, error) {
	_, err := r.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, err
	}

	in, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	return in, nil
}
