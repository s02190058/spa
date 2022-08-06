package http

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/s02190058/spa/internal/service"
	"log"
	"net/http"
)

type userService interface {
	SignUp(username, password string) (string, error)
	SignIn(username, password string) (string, error)
}

type userHandlers struct {
	service userService
}

func registerUserHandlers(r *mux.Router, service userService) {
	h := &userHandlers{
		service: service,
	}

	r.HandleFunc("/register", h.handleSignUp()).Methods(http.MethodPost)
	r.HandleFunc("/login", h.handleSignIn()).Methods(http.MethodPost)
}

func (h *userHandlers) handleSignUp() http.HandlerFunc {
	type inputData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		data := new(inputData)
		if err := json.NewDecoder(r.Body).Decode(data); err != nil {
			errorResponse(w, http.StatusBadRequest, ErrBadRequest)
			return
		}
		// the server closes the body anyway, but an explicit closure allows to release
		// occupied resources earlier
		if err := r.Body.Close(); err != nil {
			// TODO: change default logger
			log.Printf("userHandlers.SignUp: %v", err)
		}

		token, err := h.service.SignUp(data.Username, data.Password)
		if err != nil {
			var code int
			switch {
			case
				errors.Is(err, service.ErrInvalidUsername),
				errors.Is(err, service.ErrInvalidPassword),
				errors.Is(err, service.ErrAlreadyExists):
				code = http.StatusUnprocessableEntity
			default:
				code = http.StatusInternalServerError
			}
			errorResponse(w, code, err)
			return
		}

		response(w, http.StatusCreated, map[string]string{
			"token": token,
		})
	}
}

func (h *userHandlers) handleSignIn() http.HandlerFunc {
	type inputData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		data := new(inputData)
		if err := json.NewDecoder(r.Body).Decode(data); err != nil {
			errorResponse(w, http.StatusBadRequest, ErrBadRequest)
			return
		}
		// the server closes the body anyway, but an explicit closure allows to release
		// occupied resources earlier
		if err := r.Body.Close(); err != nil {
			// TODO: change default logger
			log.Printf("userHandlers.SignIn: %v", err)
		}

		token, err := h.service.SignIn(data.Username, data.Password)
		if err != nil {
			var code int
			switch {
			case
				errors.Is(err, service.ErrWrongPassword),
				errors.Is(err, service.ErrUserNotFound):
				code = http.StatusUnauthorized
			default:
				code = http.StatusInternalServerError
			}
			errorResponse(w, code, err)
			return
		}

		response(w, http.StatusOK, map[string]string{
			"token": token,
		})
	}
}
