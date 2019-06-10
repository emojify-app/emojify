package workers

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"time"

	"github.com/emojify-app/cache/protos/cache"
	"github.com/emojify-app/emojify/emojify"
	"github.com/emojify-app/emojify/logging"
	"github.com/emojify-app/emojify/queue"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Emojify is a worker which processes queue items and emojifys them
type Emojify struct {
	queue       queue.Queue
	cache       cache.CacheClient
	logger      logging.Logger
	fetcher     emojify.Fetcher
	emojifier   emojify.Emojify
	errorDelay  time.Duration
	normalDelay time.Duration
	currentItem *queue.Item
}

// New returns a new Emojify worker
func New(q queue.Queue, c cache.CacheClient, l logging.Logger, f emojify.Fetcher, e emojify.Emojify, ed, nd time.Duration) *Emojify {
	return &Emojify{
		queue:       q,
		cache:       c,
		logger:      l,
		fetcher:     f,
		emojifier:   e,
		errorDelay:  ed,
		normalDelay: nd}
}

// Start processing items on the queue
func (e *Emojify) Start() {
	l := e.logger.Log().Named("worker")

	for qi := range e.queue.Pop() {

		l.Debug("Worker processing queue item", "item", qi)

		done := e.logger.WorkerProcessQueueItem(qi.Item)
		if qi.Error != nil {
			l.Error("Error returned from queue", "error", qi.Error)
			done(http.StatusInternalServerError, qi.Error)
			continue
		}

		// do we have an item to process?
		if qi.Item == nil {
			e.logger.WorkerQueueStatus(0)
			done(http.StatusNotFound, nil)
			continue
		}

		// check the cache
		ok, err := e.checkCache(qi.Item.ID)
		if err != nil {
			done(http.StatusInternalServerError, err)
			continue
		}

		// if we have a cached item do not re-process
		if ok {
			l.Debug("Found cached item", "item", qi.Item)
			done(http.StatusOK, nil)
			continue
		}

		// fetch the image
		f, img, err := e.fetchImage(qi.Item.URI)
		if err != nil {
			done(http.StatusInternalServerError, err)
			continue
		}

		// find faces in the image
		faces, err := e.findFaces(qi.Item.URI, f)
		if err != nil {
			done(http.StatusInternalServerError, err)
			continue
		}

		// process the image and replace faces with emoji
		data, err := e.processImage(qi.Item.URI, faces, img)
		if err != nil {
			done(http.StatusInternalServerError, err)
			continue
		}

		// save the cache
		err = e.saveCache(qi.Item.URI, qi.Item.ID, data)
		if err != nil {
			done(http.StatusInternalServerError, err)
			continue
		}

		done(http.StatusOK, nil)
	}
}

// Stop gracefully stops queue processing
func (e *Emojify) Stop() {

}

func (e *Emojify) checkCache(key string) (bool, error) {
	done := e.logger.CacheExists(key)

	ok, err := e.cache.Exists(context.Background(), &wrappers.StringValue{Value: key})
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			done(http.StatusNotFound, nil)
		} else {
			done(http.StatusInternalServerError, err)
		}
	}

	done(http.StatusOK, nil)
	return ok.GetValue(), err
}

func (e *Emojify) fetchImage(uri string) (io.ReadSeeker, image.Image, error) {
	done := e.logger.WorkerFetchImage(uri)

	f, err := e.fetcher.FetchImage(uri)
	if err != nil {
		done(http.StatusInternalServerError, err)
		return nil, nil, err
	}

	done(http.StatusOK, nil)

	// check image is valid
	img, err := e.fetcher.ReaderToImage(f)
	if err != nil {
		e.logger.WorkerInvalidImage(uri, err)
		return nil, nil, err
	}

	return f, img, nil
}

func (e *Emojify) findFaces(uri string, r io.ReadSeeker) ([]image.Rectangle, error) {
	done := e.logger.WorkerFindFaces(uri)

	f, err := e.emojifier.GetFaces(r)
	if err != nil {
		done(http.StatusInternalServerError, err)
		return nil, err
	}

	done(http.StatusOK, nil)
	return f, nil
}

func (e *Emojify) processImage(uri string, faces []image.Rectangle, img image.Image) ([]byte, error) {
	done := e.logger.WorkerEmojify(uri)

	i, err := e.emojifier.Emojimise(img, faces)
	if err != nil {
		done(http.StatusInternalServerError, err)
		return nil, err
	}

	done(http.StatusOK, nil)

	// save the image
	out := new(bytes.Buffer)
	err = jpeg.Encode(out, i, &jpeg.Options{Quality: 60})
	if err != nil {
		e.logger.WorkerImageEncodeError(uri, err)
		return nil, err
	}

	return out.Bytes(), nil
}

func (e *Emojify) saveCache(uri, key string, data []byte) error {
	done := e.logger.CachePut(uri)

	ci := &cache.CacheItem{Id: key, Data: data}
	_, err := e.cache.Put(context.Background(), ci)
	if err != nil {
		done(http.StatusInternalServerError, err)
		return err
	}

	done(http.StatusOK, nil)
	return nil
}
