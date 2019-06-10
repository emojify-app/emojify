package emojify

import (
	"bytes"
	"image"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// MaxFileSize is the maximum size of a file which can be downloaded
const MaxFileSize = 10000000 // 10MB

// Fetcher defines an interface for downloading files
type Fetcher interface {
	FetchImage(uri string) (io.ReadSeeker, error)
	ReaderToImage(r io.ReadSeeker) (image.Image, error)
}

// FetcherImpl is the concrete implementation of the Fetcher
type FetcherImpl struct {
	httpClient *http.Client
}

// NewFetcher creates a new fetcher
func NewFetcher() Fetcher {
	c := &http.Client{
		Timeout: 60 * time.Second,
	}

	return &FetcherImpl{c}
}

// FetchImage does what it says on the tin
func (f *FetcherImpl) FetchImage(uri string) (io.ReadSeeker, error) {
	resp, err := f.httpClient.Get(uri)
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

// ReaderToImage convert a io Reader to an image
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
