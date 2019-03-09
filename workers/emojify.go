package workers

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io"

	"github.com/emojify-app/cache/protos/cache"
	"github.com/emojify-app/emojify/emojify"
	"github.com/emojify-app/emojify/logging"
	"github.com/emojify-app/emojify/queue"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/machinebox/sdk-go/facebox"
)

// Emojify is a worker which processes queue items and emojifys them
type Emojify struct {
	queue     queue.Queue
	cache     cache.CacheClient
	logger    logging.Logger
	fetcher   emojify.Fetcher
	emojifier emojify.Emojify
}

// Start processing items on the queue
func (e *Emojify) Start() {
	for qi := range e.queue.Pop() {

		// check the cache
		ok, err := e.checkCache(qi.Item.ID)
		if err != nil {
			return
		}

		// if we have a cached item do not re-process
		if ok {
			return
		}

		// fetch the image
		f, img, err := e.fetchImage(qi.Item.URI)
		if err != nil {
			//done(http.StatusInternalServerError, err)
			return
		}

		// find faces in the image
		faces, err := e.findFaces(qi.Item.URI, f)
		if err != nil {
			//done(http.StatusInternalServerError, err)
			return
		}

		// process the image and replace faces with emoji
		data, err := e.processImage(qi.Item.URI, faces, img)
		if err != nil {
			//done(http.StatusInternalServerError, err)
			return
		}

		// save the cache
		err = e.saveCache(qi.Item.URI, qi.Item.ID, data)
		if err != nil {
			//done(http.StatusInternalServerError, err)
			return
		}

	}
}

// Stop gracefully stops queue processing
func (e *Emojify) Stop() {

}

func (e *Emojify) checkCache(key string) (bool, error) {
	ok, err := e.cache.Exists(context.Background(), &wrappers.StringValue{Value: key})

	return ok.GetValue(), err
}

func (e *Emojify) fetchImage(uri string) (io.ReadSeeker, image.Image, error) {
	//	fiDone := e.logger.EmojifyHandlerFetchImage(uri)
	f, err := e.fetcher.FetchImage(uri)
	if err != nil {
		//		fiDone(http.StatusInternalServerError, err)
		return nil, nil, err
	}

	//	fiDone(http.StatusOK, nil)

	// check image is valid
	img, err := e.fetcher.ReaderToImage(f)
	if err != nil {
		//		e.logger.EmojifyHandlerInvalidImage(uri, err)
		return nil, nil, err
	}

	return f, img, nil
}

func (e *Emojify) findFaces(uri string, r io.ReadSeeker) ([]facebox.Face, error) {
	//ffDone := e.logger.EmojifyHandlerFindFaces(uri)
	f, err := e.emojifier.GetFaces(r)
	if err != nil {
		//ffDone(http.StatusInternalServerError, err)
		return nil, err
	}

	//ffDone(http.StatusOK, nil)
	return f, nil
}

func (e *Emojify) processImage(uri string, faces []facebox.Face, img image.Image) ([]byte, error) {
	//	emDone := e.logger.EmojifyHandlerEmojify(uri)
	i, err := e.emojifier.Emojimise(img, faces)
	if err != nil {
		//		emDone(http.StatusInternalServerError, err)
		return nil, err
	}
	//	emDone(http.StatusOK, nil)

	// save the image
	out := new(bytes.Buffer)
	err = jpeg.Encode(out, i, &jpeg.Options{Quality: 60})
	if err != nil {
		//e.logger.EmojifyHandlerImageEncodeError(uri, err)
		return nil, err
	}

	return out.Bytes(), nil
}

func (e *Emojify) saveCache(uri, key string, data []byte) error {
	//	cpDone := e.logger.EmojifyHandlerCachePut(uri)
	ci := &cache.CacheItem{Id: key, Data: data}
	_, err := e.cache.Put(context.Background(), ci)
	if err != nil {
		//		cpDone(http.StatusInternalServerError, err)
		return err
	}

	//	cpDone(http.StatusOK, nil)
	return nil
}
