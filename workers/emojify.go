package workers

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"io"
	"net/http"

	"github.com/emojify-app/cache/protos/cache"
	"github.com/emojify-app/emojify/emojify"
	"github.com/emojify-app/emojify/logging"
	"github.com/emojify-app/emojify/queue"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/machinebox/sdk-go/facebox"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
		done := e.logger.WorkerProcessQueueItem(qi.Item)

		if qi.Error != nil {
			done(http.StatusInternalServerError, qi.Error)
			break
		}

		// check the cache
		ok, err := e.checkCache(qi.Item.ID)
		if err != nil {
			done(http.StatusInternalServerError, err)
			break
		}

		// if we have a cached item do not re-process
		if ok {
			done(http.StatusOK, nil)
			break
		}

		// fetch the image
		f, img, err := e.fetchImage(qi.Item.URI)
		if err != nil {
			done(http.StatusInternalServerError, err)
			break
		}

		// find faces in the image
		faces, err := e.findFaces(qi.Item.URI, f)
		if err != nil {
			done(http.StatusInternalServerError, err)
			break
		}

		// process the image and replace faces with emoji
		data, err := e.processImage(qi.Item.URI, faces, img)
		if err != nil {
			done(http.StatusInternalServerError, err)
			break
		}

		// save the cache
		err = e.saveCache(qi.Item.URI, qi.Item.ID, data)
		if err != nil {
			done(http.StatusInternalServerError, err)
			break
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

func (e *Emojify) findFaces(uri string, r io.ReadSeeker) ([]facebox.Face, error) {
	done := e.logger.WorkerFindFaces(uri)

	f, err := e.emojifier.GetFaces(r)
	if err != nil {
		done(http.StatusInternalServerError, err)
		return nil, err
	}

	done(http.StatusOK, nil)
	return f, nil
}

func (e *Emojify) processImage(uri string, faces []facebox.Face, img image.Image) ([]byte, error) {
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