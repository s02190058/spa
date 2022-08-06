package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/s02190058/spa/internal/entity"
	"github.com/s02190058/spa/internal/service"
)

var (
	ErrInvalidCommentID = errors.New("invalid comment id")
	ErrInvalidPostID    = errors.New("invalid post id")
)

type postService interface {
	GetAll() ([]*entity.Post, error)
	Get(id int) (*entity.Post, error)
	GetByCategory(category string) ([]*entity.Post, error)
	GetByUsername(username string) ([]*entity.Post, error)
	Add(typ, category, title, text, url string, author *entity.User) (*entity.Post, error)
	Upvote(postID, userID int) (*entity.Post, error)
	Downvote(postID, userID int) (*entity.Post, error)
	Unvote(postID, userID int) (*entity.Post, error)
	Delete(postID, userID int) error
	AddComment(postID, userID int, body string) (*entity.Post, error)
	DeleteComment(postID, commentID, userID int) (*entity.Post, error)
}

type postHandlers struct {
	service postService
}

func registerPostHandlers(r *mux.Router, service postService, m *middleware) {
	h := &postHandlers{
		service: service,
	}

	r.HandleFunc("/posts/", h.handleGetAll()).Methods(http.MethodGet)
	r.HandleFunc("/post/{post_id}", h.handleGet()).Methods(http.MethodGet)
	r.HandleFunc("/posts/{category}", h.handleGetByCategory()).Methods(http.MethodGet)
	r.HandleFunc("/user/{username}", h.handleGetByUsername()).Methods(http.MethodGet)

	s := r.PathPrefix("/").Subrouter()
	s.Use(m.checkAuthorization)
	s.HandleFunc("/posts", h.handleCreate()).Methods(http.MethodPost)
	s.HandleFunc("/post/{post_id}/upvote", h.handleUpvote()).Methods(http.MethodGet)
	s.HandleFunc("/post/{post_id}/downvote", h.handleDownvote()).Methods(http.MethodGet)
	s.HandleFunc("/post/{post_id}/unvote", h.handleUnvote()).Methods(http.MethodGet)
	s.HandleFunc("/post/{post_id}", h.handleCreateComment()).Methods(http.MethodPost)
	s.HandleFunc("/post/{post_id}/{comment_id}", h.handleDeleteComment()).Methods(http.MethodDelete)
	s.HandleFunc("/post/{post_id}", h.handleDelete()).Methods(http.MethodDelete)
}

func (h *postHandlers) handleGetAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := h.service.GetAll()
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}

		response(w, http.StatusOK, posts)
	}
}

func (h *postHandlers) handleGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["post_id"]
		idInt, err := strconv.Atoi(id)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, ErrInvalidPostID)
			return
		}

		post, err := h.service.Get(idInt)
		if err != nil {
			var code int
			switch {
			case errors.Is(err, service.ErrPostNotFound):
				code = http.StatusNotFound
			default:
				code = http.StatusInternalServerError
			}
			errorResponse(w, code, err)
			return
		}

		response(w, http.StatusOK, post)
	}
}

func (h *postHandlers) handleGetByCategory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		category := vars["category"]

		posts, err := h.service.GetByCategory(category)
		if err != nil {
			var code int
			switch {
			case errors.Is(err, service.ErrInvalidCategory):
				code = http.StatusUnprocessableEntity
			default:
				code = http.StatusInternalServerError
			}
			errorResponse(w, code, err)
			return
		}

		response(w, http.StatusOK, posts)
	}
}

func (h *postHandlers) handleGetByUsername() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		username := vars["username"]

		posts, err := h.service.GetByUsername(username)
		if err != nil {
			var code int
			switch {
			case errors.Is(err, service.ErrUserNotFound):
				code = http.StatusNotFound
			default:
				code = http.StatusInternalServerError
			}
			errorResponse(w, code, err)
			return
		}

		response(w, http.StatusOK, posts)
	}
}

func (h *postHandlers) handleCreate() http.HandlerFunc {
	type inputData struct {
		Type     string `json:"type"`
		Category string `json:"category"`
		Title    string `json:"title"`
		Text     string `json:"text"`
		URL      string `json:"URL"`
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
			log.Printf("postHandlers.Create: %v", err)
		}

		user, err := userFromContext(r.Context())
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, ErrInternal)
			return
		}

		post, err := h.service.Add(
			data.Type,
			data.Category,
			data.Title,
			data.Text,
			data.URL,
			user,
		)
		if err != nil {
			var code int
			switch {
			case
				errors.Is(err, service.ErrInvalidType),
				errors.Is(err, service.ErrInvalidCategory),
				errors.Is(err, service.ErrInvalidText),
				errors.Is(err, service.ErrInvalidURL):
				code = http.StatusUnprocessableEntity
			default:
				code = http.StatusInternalServerError
			}
			errorResponse(w, code, err)
			return
		}

		response(w, http.StatusCreated, post)
	}
}

func (h *postHandlers) handleUpvote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["post_id"]
		idInt, err := strconv.Atoi(id)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, ErrInvalidPostID)
			return
		}

		user, err := userFromContext(r.Context())
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, ErrInternal)
			return
		}

		post, err := h.service.Upvote(idInt, user.ID)
		if err != nil {
			var code int
			switch {
			case errors.Is(err, service.ErrPostNotFound):
				code = http.StatusNotFound
			default:
				code = http.StatusInternalServerError
			}
			errorResponse(w, code, err)
			return
		}

		response(w, http.StatusOK, post)
	}
}

func (h *postHandlers) handleDownvote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["post_id"]
		idInt, err := strconv.Atoi(id)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, ErrInvalidPostID)
			return
		}

		user, err := userFromContext(r.Context())
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, ErrInternal)
			return
		}

		post, err := h.service.Downvote(idInt, user.ID)
		if err != nil {
			var code int
			switch {
			case errors.Is(err, service.ErrPostNotFound):
				code = http.StatusNotFound
			default:
				code = http.StatusInternalServerError
			}
			errorResponse(w, code, err)
			return
		}

		response(w, http.StatusOK, post)
	}
}

func (h *postHandlers) handleUnvote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["post_id"]
		idInt, err := strconv.Atoi(id)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, ErrInvalidPostID)
			return
		}

		user, err := userFromContext(r.Context())
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, ErrInternal)
			return
		}

		post, err := h.service.Unvote(idInt, user.ID)
		if err != nil {
			var code int
			switch {
			case errors.Is(err, service.ErrPostNotFound):
				code = http.StatusNotFound
			default:
				code = http.StatusInternalServerError
			}
			errorResponse(w, code, err)
			return
		}

		response(w, http.StatusOK, post)
	}
}

func (h *postHandlers) handleCreateComment() http.HandlerFunc {
	type inputData struct {
		Comment string `json:"comment"`
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

		vars := mux.Vars(r)
		id := vars["post_id"]
		idInt, err := strconv.Atoi(id)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, ErrInvalidPostID)
			return
		}

		user, err := userFromContext(r.Context())
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, ErrInternal)
			return
		}

		post, err := h.service.AddComment(idInt, user.ID, data.Comment)
		if err != nil {
			var code int
			switch {
			case errors.Is(err, service.ErrPostNotFound):
				code = http.StatusNotFound
			case errors.Is(err, service.ErrInvalidBody):
				code = http.StatusUnprocessableEntity
			default:
				code = http.StatusInternalServerError
			}
			errorResponse(w, code, err)
			return
		}

		response(w, http.StatusCreated, post)
	}
}

func (h *postHandlers) handleDeleteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		postID := vars["post_id"]
		postIDInt, err := strconv.Atoi(postID)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, ErrInvalidPostID)
			return
		}
		commentID := vars["comment_id"]
		commentIDInt, err := strconv.Atoi(commentID)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, ErrInvalidCommentID)
			return
		}

		user, err := userFromContext(r.Context())
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, ErrInternal)
			return
		}

		post, err := h.service.DeleteComment(postIDInt, commentIDInt, user.ID)
		if err != nil {
			var code int
			switch {
			case errors.Is(err, service.ErrUnauthorized):
				code = http.StatusUnauthorized
			case
				errors.Is(err, service.ErrPostNotFound),
				errors.Is(err, service.ErrCommentNotFound):
				code = http.StatusNotFound
			default:
				code = http.StatusInternalServerError
			}
			errorResponse(w, code, err)
			return
		}

		response(w, http.StatusOK, post)
	}
}

func (h *postHandlers) handleDelete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["post_id"]
		idInt, err := strconv.Atoi(id)
		if err != nil {
			errorResponse(w, http.StatusBadRequest, ErrInvalidPostID)
			return
		}

		user, err := userFromContext(r.Context())
		if err != nil {
			errorResponse(w, http.StatusInternalServerError, ErrInternal)
			return
		}

		if err := h.service.Delete(idInt, user.ID); err != nil {
			var code int
			switch {
			case errors.Is(err, service.ErrUnauthorized):
				code = http.StatusUnauthorized
			case errors.Is(err, service.ErrPostNotFound):
				code = http.StatusNotFound
			default:
				code = http.StatusInternalServerError
			}
			errorResponse(w, code, err)
			return
		}

		response(w, http.StatusOK, map[string]string{
			"message": "success",
		})
	}
}
