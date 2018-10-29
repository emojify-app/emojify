package emojify

import (
	"context"
	"net/http"

	"github.com/emojify-app/emojify/handlers"
	"github.com/emojify-app/emojify/logic"
	"github.com/gorilla/mux"
	hclog "github.com/hashicorp/go-hclog"
)

// Server is the main server for the application
type Server struct {
	logger hclog.Logger
	server *http.Server

	bindAddress string
	mux         *mux.Router
}

// NewServer creates a new server
func NewServer(bindAddress, cacheAddress, consulAddress, faceboxAddress, serviceName, imagePath string) *Server {
	l, e, f, c := setupDependencies(cacheAddress, consulAddress, faceboxAddress, serviceName, imagePath)
	mux := setupRouter(l, c, e, f)

	return &Server{
		logger:      l,
		bindAddress: bindAddress,
		mux:         mux,
	}
}

// Start the server
func (s *Server) Start() {
	s.server = &http.Server{
		Addr:    s.bindAddress,
		Handler: s.mux,
	}

	err := s.server.ListenAndServe()
	if err != nil {
		s.logger.Error("Error starting server", "error", err)
	}
}

// Stop the server
func (s *Server) Stop() {
	err := s.server.Shutdown(context.Background())
	if err != nil {
		s.logger.Error("Error stopping server", "error", err)
	}
}

func setupDependencies(cacheAddress, consulAddress, faceboxAddress, serviceName, imagePath string) (hclog.Logger, logic.Emojify, logic.Fetcher, logic.Cache) {
	l := hclog.Default()

	c, err := logic.NewCache(cacheAddress, consulAddress, serviceName)
	if err != nil {
		l.Error("Error creating cache client", "error", err)
	}

	f := &logic.FetcherImpl{}
	e := logic.NewEmojify(faceboxAddress, imagePath)

	return l, e, f, c
}

func setupRouter(l hclog.Logger, c logic.Cache, e logic.Emojify, f logic.Fetcher) *mux.Router {
	r := mux.NewRouter()
	ch := handlers.NewCacheHandler(l.Named("cacheHandler"), c)
	r.HandleFunc("/{image}", ch.ServeHTTP).Methods("GET")

	hh := handlers.HealthHandler{}
	r.HandleFunc("/health", hh.ServeHTTP).Methods("GET")

	eh := handlers.NewEmojifyHandler(e, f, c, l.Named("emojiHandler"))
	r.HandleFunc("/", eh.ServeHTTP).Methods("POST")

	return r
}
