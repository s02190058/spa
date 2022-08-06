package http

import (
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/s02190058/spa/internal/config"
	"github.com/s02190058/spa/pkg/jwt"
)

func NewRouter(
	logger *logrus.Logger,
	tokenManager *jwt.TokenManager,
	userService userService,
	postService postService,
	static config.Static,
) *mux.Router {
	r := mux.NewRouter()
	m := &middleware{
		logger:       logger,
		tokenManager: tokenManager,
	}
	r.Use(m.setRequestID)
	r.Use(m.logRequest)

	s := r.PathPrefix("/api").Subrouter()
	registerUserHandlers(s, userService)
	registerPostHandlers(s, postService, m)
	s.PathPrefix("/").Handler(http.NotFoundHandler())

	registerStaticHandlers(r, static.Path, static.Index)

	return r
}
