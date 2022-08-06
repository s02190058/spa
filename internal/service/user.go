package service

import (
	"errors"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/s02190058/spa/internal/entity"
	"github.com/s02190058/spa/pkg/hasher"
	"github.com/s02190058/spa/pkg/jwt"
)

var (
	ErrInvalidUsername = errors.New("invalid username")
	ErrInvalidPassword = errors.New("invalid password")
	ErrAlreadyExists   = errors.New("already exists")
	ErrUserNotFound    = errors.New("user not found")
	ErrWrongPassword   = errors.New("wrong password")
)

type userRepo interface {
	Add(user *entity.User) (*entity.User, error)
	GetByUsername(username string) (*entity.User, error)
}

type UserService struct {
	repo           userRepo
	tokenManager   *jwt.TokenManager
	passwordHasher *hasher.Hasher
}

func NewUserService(repo userRepo, tokenManager *jwt.TokenManager, hasher *hasher.Hasher) *UserService {
	return &UserService{
		repo:           repo,
		tokenManager:   tokenManager,
		passwordHasher: hasher,
	}
}

func (s *UserService) SignUp(username, password string) (string, error) {
	if validation.Validate(username, validation.Length(1, 32), is.PrintableASCII) != nil {
		return "", ErrInvalidUsername
	}
	if validation.Validate(password, validation.Length(8, 72)) != nil {
		return "", ErrInvalidPassword
	}

	encryptedPassword, err := s.passwordHasher.Encrypt(password)
	if err != nil {
		return "", ErrInternal
	}

	user, err := s.repo.Add(&entity.User{
		Username:          username,
		EncryptedPassword: encryptedPassword,
	})
	if err != nil {
		return "", err
	}

	tokenString, err := s.tokenManager.Create(user)
	if err != nil {
		return "", ErrInternal
	}

	return tokenString, nil
}

func (s *UserService) SignIn(username, password string) (string, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return "", err
	}

	if !s.passwordHasher.Compare(user.EncryptedPassword, password) {
		return "", ErrWrongPassword
	}

	tokenString, err := s.tokenManager.Create(user)
	if err != nil {
		return "", ErrInternal
	}

	return tokenString, nil
}
