package repo

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"github.com/s02190058/spa/internal/entity"
	"github.com/s02190058/spa/internal/service"
	"log"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) Add(user *entity.User) (*entity.User, error) {
	query := "INSERT INTO users (name, encrypted_password) " +
		"VALUES ($1, $2) " +
		"RETURNING ID"

	if err := r.db.QueryRow(
		query,
		user.Username,
		user.EncryptedPassword,
	).Scan(
		&user.ID,
	); err != nil {
		pqErr, ok := err.(*pq.Error)
		if !ok {
			// TODO: change default logger
			log.Printf("DB.QueryRow: %v", err)
			return nil, service.ErrInternal
		}

		var retErr error
		switch pqErr.Code.Name() {
		case "unique_violation":
			retErr = service.ErrAlreadyExists
		default:
			// TODO: change default logger
			log.Printf("DB.QueryRow: %v", err)
			retErr = service.ErrInternal
		}
		return nil, retErr
	}

	return user, nil
}

func (r *UserRepo) GetByUsername(username string) (*entity.User, error) {

	query := "SELECT id, name, encrypted_password " +
		"FROM users " +
		"WHERE name = $1"

	user := new(entity.User)
	if err := r.db.QueryRow(
		query,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.EncryptedPassword,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrUserNotFound
		}
		// TODO: change default logger
		log.Printf("DB.QueryRow: %v", err)
		return nil, service.ErrInternal
	}

	return user, nil
}
