package http

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/s02190058/spa/internal/entity"
	"github.com/s02190058/spa/pkg/jwt"
)

type ctxRequestIDKey int

var requestIDKey ctxRequestIDKey

var (
	ErrUnauthorized = errors.New("unauthorized")
)

type middleware struct {
	logger       *logrus.Logger
	tokenManager *jwt.TokenManager
}

func (m *middleware) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
	})
}

func (m *middleware) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := m.logger.WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(requestIDKey),
		})

		logger.Infof("started %s %s", r.Method, r.RequestURI)

		start := time.Now()
		rw := &responseWriter{
			ResponseWriter: w,
			code:           http.StatusOK,
		}
		next.ServeHTTP(rw, r)

		var level logrus.Level
		switch {
		case rw.code >= 500:
			level = logrus.ErrorLevel
		case rw.code >= 400:
			level = logrus.WarnLevel
		default:
			level = logrus.InfoLevel
		}
		logger.Logf(
			level,
			"completed with %d %s in %v",
			rw.code,
			http.StatusText(rw.code),
			time.Since(start),
		)
	})
}

func (m *middleware) checkAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("authorization")
		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			errorResponse(w, http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		token := headerParts[1]
		res, err := m.tokenManager.Check(token)
		if err != nil {
			errorResponse(w, http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		user := &entity.User{}
		if err := mapstructure.Decode(res, user); err != nil {
			errorResponse(w, http.StatusInternalServerError, ErrInternal)
			return
		}

		ctx := contextWithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
