package httpserver

import (
	"context"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type Server struct {
	server          *http.Server
	logger          *logrus.Logger
	notify          chan error
	shutdownTimeout time.Duration
}

func New(logger *logrus.Logger, handler http.Handler, port string, shutdownTimeout time.Duration) *Server {
	return &Server{
		server: &http.Server{
			Addr:    ":" + port,
			Handler: handler,
		},
		logger:          logger,
		notify:          make(chan error, 1),
		shutdownTimeout: shutdownTimeout,
	}
}

func (s *Server) Start() {
	s.logger.Infof("starting server on %s", s.server.Addr)
	go func() {
		s.notify <- s.server.ListenAndServe()
	}()
}

func (s *Server) Notify() <-chan error {
	return s.notify
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}
