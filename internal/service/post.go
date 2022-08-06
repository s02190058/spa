package service

import (
	"errors"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/s02190058/spa/internal/entity"
	"sort"
)

const (
	upvote   = 1
	downvote = -1
)

var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidType     = errors.New("invalid post type")
	ErrInvalidCategory = errors.New("invalid post category")
	ErrInvalidTitle    = errors.New("invalid post title")
	ErrInvalidText     = errors.New("invalid post text")
	ErrInvalidURL      = errors.New("invalid post url")
	ErrPostNotFound    = errors.New("post not found")
	ErrInvalidBody     = errors.New("invalid comment body")
	ErrCommentNotFound = errors.New("comment not found")
)

type postRepo interface {
	GetAll() ([]*entity.Post, error)
	Get(id int) (*entity.Post, error)
	GetByCategory(category string) ([]*entity.Post, error)
	GetByUsername(username string) ([]*entity.Post, error)
	Add(post *entity.Post) (*entity.Post, error)
	AddVote(postID, userID, vote int) (*entity.Post, error)
	DeleteVote(postID, userID int) (*entity.Post, error)
	Delete(postID, userID int) error
	AddComment(postID, userID int, body string) (*entity.Post, error)
	DeleteComment(postID, commentID, userID int) (*entity.Post, error)
}

type PostService struct {
	repo postRepo
}

func NewPostService(repo postRepo) *PostService {
	return &PostService{
		repo: repo,
	}
}

func (s *PostService) GetAll() ([]*entity.Post, error) {
	posts, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Score > posts[j].Score
	})

	return posts, nil
}

func (s *PostService) Get(id int) (*entity.Post, error) {
	return s.repo.Get(id)
}

func (s *PostService) GetByCategory(category string) ([]*entity.Post, error) {
	posts, err := s.repo.GetByCategory(category)
	if err != nil {
		return nil, err
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Score > posts[j].Score
	})

	return posts, nil
}
func (s *PostService) GetByUsername(username string) ([]*entity.Post, error) {
	posts, err := s.repo.GetByUsername(username)
	if err != nil {
		return nil, err
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Created.After(posts[j].Created)
	})

	return posts, nil
}

func (s *PostService) Add(
	typ, category, title, text, url string,
	author *entity.User,
) (*entity.Post, error) {
	if validation.Validate(title, validation.Length(1, 1<<10)) != nil {
		return nil, ErrInvalidTitle
	}
	if url != "" && validation.Validate(url, is.URL) != nil {
		return nil, ErrInvalidURL
	} else if validation.Validate(text, validation.Length(4, 1<<20)) != nil {
		return nil, ErrInvalidText
	}

	post := &entity.Post{
		Type:     typ,
		Category: category,
		Title:    title,
		Text:     text,
		URL:      url,
		Author:   author,
		Votes: []*entity.Vote{
			{
				UserID: author.ID,
				Vote:   upvote,
			},
		},
		Comments: []*entity.Comment{},
	}

	post, err := s.repo.Add(post)
	if err != nil {
		return nil, err
	}

	post.CalcAndSetScore()
	post.CalcAndSetUpvotePercentage()

	return post, nil
}

func (s *PostService) Upvote(postID, userID int) (*entity.Post, error) {
	return s.repo.AddVote(postID, userID, upvote)
}

func (s *PostService) Downvote(postID, userID int) (*entity.Post, error) {
	return s.repo.AddVote(postID, userID, downvote)
}

func (s *PostService) Unvote(postID, userID int) (*entity.Post, error) {
	return s.repo.DeleteVote(postID, userID)
}

func (s *PostService) AddComment(postID, userID int, body string) (*entity.Post, error) {
	if validation.Validate(body, validation.Length(1, 1<<20)) != nil {
		return nil, ErrInvalidBody
	}
	return s.repo.AddComment(postID, userID, body)
}

func (s *PostService) DeleteComment(postId, commentID, userID int) (*entity.Post, error) {
	return s.repo.DeleteComment(postId, commentID, userID)
}

func (s *PostService) Delete(postID, userID int) error {
	return s.repo.Delete(postID, userID)
}
