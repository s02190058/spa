package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/s02190058/spa/internal/entity"
	"github.com/s02190058/spa/internal/service"
)

type PostRepo struct {
	db *sql.DB
}

func NewPostRepo(db *sql.DB) *PostRepo {
	return &PostRepo{
		db: db,
	}
}

func getVotesByPostID(tx *sql.Tx, id int) ([]*entity.Vote, error) {
	query := "SELECT user_id, vote " +
		"FROM votes " +
		"WHERE post_id = $1"

	rows, err := tx.Query(query, id)
	if err != nil {
		// TODO: change default logger
		log.Printf("Tx.Query: %v", err)
		return nil, service.ErrInternal
	}

	votes := make([]*entity.Vote, 0)
	for rows.Next() {
		vote := new(entity.Vote)
		if err := rows.Scan(
			&vote.UserID,
			&vote.Vote,
		); err != nil {
			// TODO: change default logger
			log.Printf("Rows.Scan: %v", err)
			return nil, service.ErrInternal
		}

		votes = append(votes, vote)
	}
	if err := rows.Err(); err != nil {
		// TODO: change default logger
		log.Printf("Rows.Err: %v", err)
		return nil, service.ErrInternal
	}

	return votes, nil
}

func getCommentsByPostID(tx *sql.Tx, id int) ([]*entity.Comment, error) {
	query := "SELECT c.id, u.id, u.name, c.body, c.created " +
		"FROM comments c " +
		"JOIN users u " +
		"ON c.user_id = u.id " +
		"WHERE c.post_id = $1"

	rows, err := tx.Query(query, id)
	if err != nil {
		// TODO: change default logger
		log.Printf("Tx.Query: %v", err)
		return nil, service.ErrInternal
	}

	comments := make([]*entity.Comment, 0)
	for rows.Next() {
		comment := new(entity.Comment)
		comment.Author = new(entity.User)
		if err := rows.Scan(
			&comment.ID,
			&comment.Author.ID,
			&comment.Author.Username,
			&comment.Body,
			&comment.Created,
		); err != nil {
			// TODO: change default logger
			log.Printf("Rows.Scan: %v", err)
			return nil, service.ErrInternal
		}

		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		// TODO: change default logger
		log.Printf("Rows.Err: %v", err)
		return nil, service.ErrInternal
	}

	return comments, nil
}

func getWithConditions(tx *sql.Tx, conditions ...string) ([]*entity.Post, error) {
	query := "SELECT p.id, t.name, c.name, p.title, p.text, p.url, u.id, u.name, p.views, p.created " +
		"FROM posts p " +
		"JOIN types t " +
		"ON p.type_id = t.id " +
		"JOIN categories c " +
		"ON p.category_id = c.id " +
		"JOIN users u " +
		"ON p.user_id = u.id"

	if len(conditions) > 0 {
		query += " WHERE "
		for i, condition := range conditions {
			if i > 0 {
				query += " AND "
			}
			query += condition
		}
	}

	rows, err := tx.Query(query)
	if err != nil {
		// TODO: change default logger
		log.Printf("Tx.Query: %v", err)
		return nil, service.ErrInternal
	}

	posts := make([]*entity.Post, 0)
	for rows.Next() {
		post := new(entity.Post)
		post.Author = new(entity.User)
		if err := rows.Scan(
			&post.ID,
			&post.Type,
			&post.Category,
			&post.Title,
			&post.Text,
			&post.URL,
			&post.Author.ID,
			&post.Author.Username,
			&post.Views,
			&post.Created,
		); err != nil {
			// TODO: change default logger
			log.Printf("Rows.Scan: %v", err)
			return nil, service.ErrInternal
		}

		posts = append(posts, post)
	}

	for _, post := range posts {
		votes, err := getVotesByPostID(tx, post.ID)
		if err != nil {
			return nil, err
		}
		post.Votes = votes
		post.CalcAndSetScore()
		post.CalcAndSetUpvotePercentage()

		comments, err := getCommentsByPostID(tx, post.ID)
		if err != nil {
			return nil, err
		}
		post.Comments = comments
	}

	return posts, nil
}

func (r *PostRepo) GetAll() ([]*entity.Post, error) {
	tx, err := r.db.Begin()
	if err != nil {
		// TODO: change default logger
		log.Printf("DB.Begin: %v", err)
		return nil, service.ErrInternal
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			// TODO: change default logger
			log.Printf("Tx.Rollback: %v", err)
		}
	}()

	posts, err := getWithConditions(tx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		// TODO: change default logger
		log.Printf("Tx.Commit: %v", err)
		return nil, service.ErrInternal
	}

	return posts, nil
}

func get(tx *sql.Tx, id int) (*entity.Post, error) {
	query := "SELECT p.id, t.name, c.name, p.title, p.text, p.url, u.id, u.name, p.views, p.created " +
		"FROM posts p " +
		"JOIN types t " +
		"ON p.type_id = t.id " +
		"JOIN categories c " +
		"ON p.category_id = c.id " +
		"JOIN users u " +
		"ON p.user_id = u.id " +
		"WHERE p.id = $1"

	post := new(entity.Post)
	post.Author = new(entity.User)
	if err := tx.QueryRow(
		query,
		id,
	).Scan(
		&post.ID,
		&post.Type,
		&post.Category,
		&post.Title,
		&post.Text,
		&post.URL,
		&post.Author.ID,
		&post.Author.Username,
		&post.Views,
		&post.Created,
	); err != nil {
		// TODO: change default logger
		log.Printf("Tx.QueryRow: %v", err)
		return nil, service.ErrInternal
	}

	votes, err := getVotesByPostID(tx, id)
	if err != nil {
		return nil, err
	}
	post.Votes = votes
	post.CalcAndSetScore()
	post.CalcAndSetUpvotePercentage()

	comments, err := getCommentsByPostID(tx, id)
	if err != nil {
		return nil, err
	}
	post.Comments = comments

	return post, nil
}

func (r *PostRepo) Get(id int) (*entity.Post, error) {
	tx, err := r.db.Begin()
	if err != nil {
		// TODO: change default logger
		log.Printf("DB.Begin: %v", err)
		return nil, service.ErrInternal
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			// TODO: change default logger
			log.Printf("Tx.Rollback: %v", err)
		}
	}()

	query := "UPDATE posts " +
		"SET views = views + 1" +
		"WHERE id = $1"

	res, err := tx.Exec(query, id)
	if err != nil {
		// TODO: change default logger
		log.Printf("Tx.Exec: %v", err)
		return nil, service.ErrInternal
	}
	n, err := res.RowsAffected()
	if err != nil {
		// TODO: change default logger
		log.Printf("Result.RowsAffected: %v", err)
		return nil, service.ErrInternal
	}
	if n == 0 {
		return nil, service.ErrPostNotFound
	}

	post, err := get(tx, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		// TODO: change default logger
		log.Printf("Tx.Commit: %v", err)
		return nil, service.ErrInternal
	}

	return post, nil
}

func (r *PostRepo) GetByCategory(category string) ([]*entity.Post, error) {
	tx, err := r.db.Begin()
	if err != nil {
		// TODO: change default logger
		log.Printf("DB.Begin: %v", err)
		return nil, service.ErrInternal
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			// TODO: change default logger
			log.Printf("Tx.Rollback: %v", err)
		}
	}()

	query := "SELECT id " +
		"FROM categories " +
		"WHERE name = $1"

	var categoryID int
	if err := tx.QueryRow(
		query,
		category,
	).Scan(
		&categoryID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrInvalidCategory
		}
		// TODO: change default logger
		log.Printf("Tx.QueryRow: %v", err)
		return nil, service.ErrInternal
	}

	posts, err := getWithConditions(tx, fmt.Sprintf("c.id = %d", categoryID))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		// TODO: change default logger
		log.Printf("Tx.Commit: %v", err)
		return nil, service.ErrInternal
	}

	return posts, nil
}

func (r *PostRepo) GetByUsername(username string) ([]*entity.Post, error) {
	tx, err := r.db.Begin()
	if err != nil {
		// TODO: change default logger
		log.Printf("DB.Begin: %v", err)
		return nil, service.ErrInternal
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			// TODO: change default logger
			log.Printf("Tx.Rollback: %v", err)
		}
	}()

	query := "SELECT id " +
		"FROM users " +
		"WHERE name = $1"

	var userID int
	if err := tx.QueryRow(
		query,
		username,
	).Scan(
		&userID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrUserNotFound
		}
		// TODO: change default logger
		log.Printf("Tx.QueryRow: %v", err)
		return nil, service.ErrInternal
	}

	posts, err := getWithConditions(tx, fmt.Sprintf("u.id = %d", userID))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		// TODO: change default logger
		log.Printf("Tx.Commit: %v", err)
		return nil, service.ErrInternal
	}

	return posts, nil
}

func (r *PostRepo) Add(post *entity.Post) (*entity.Post, error) {
	tx, err := r.db.Begin()
	if err != nil {
		// TODO: change default logger
		log.Printf("DB.Begin: %v", err)
		return nil, service.ErrInternal
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			// TODO: change default logger
			log.Printf("Tx.Rollback: %v", err)
		}
	}()

	query := "SELECT id " +
		"FROM types " +
		"WHERE name = $1"

	var typeID int
	if err := tx.QueryRow(
		query,
		post.Type,
	).Scan(
		&typeID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrInvalidType
		}
		// TODO: change default logger
		log.Printf("Tx.QueryRow: %v", err)
		return nil, service.ErrInternal
	}

	query = "SELECT id " +
		"FROM categories " +
		"WHERE name = $1"

	var categoryID int
	if err := tx.QueryRow(
		query,
		post.Category,
	).Scan(
		&categoryID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, service.ErrInvalidCategory
		}
		// TODO: change default logger
		log.Printf("Tx.QueryRow: %v", err)
		return nil, service.ErrInternal
	}

	query = "INSERT INTO posts (type_id, category_id, title, text, url, user_id) " +
		"VALUES ($1, $2, $3, $4, $5, $6) " +
		"RETURNING id, created"

	if err := tx.QueryRow(
		query,
		typeID,
		categoryID,
		post.Title,
		post.Text,
		post.URL,
		post.Author.ID,
	).Scan(
		&post.ID,
		&post.Created,
	); err != nil {
		// TODO: change default logger
		log.Printf("Tx.QueryRow: %v", err)
		return nil, service.ErrInternal
	}

	query = "INSERT INTO votes (post_id, user_id, vote) " +
		"VALUES ($1, $2, $3)"

	if _, err := tx.Exec(
		query,
		post.ID,
		post.Votes[0].UserID,
		post.Votes[0].Vote,
	); err != nil {
		// TODO: change default logger
		log.Printf("Tx.Exec: %v", err)
		return nil, service.ErrInternal
	}

	if err := tx.Commit(); err != nil {
		// TODO: change default logger
		log.Printf("Tx.Commit: %v", err)
		return nil, service.ErrInternal
	}

	return post, nil
}

func checkPost(tx *sql.Tx, id int) error {

	query := "SELECT " +
		"FROM posts " +
		"WHERE id = $1"

	if err := tx.QueryRow(
		query,
		id,
	).Scan(); err != nil {
		var retErr error
		switch {
		case errors.Is(err, sql.ErrNoRows):
			retErr = service.ErrPostNotFound
		default:
			// TODO: change default logger
			log.Printf("Tx.QueryRow: %v", err)
			retErr = service.ErrInternal
		}
		return retErr
	}

	return nil
}

func (r *PostRepo) AddVote(postID, userID, vote int) (*entity.Post, error) {
	tx, err := r.db.Begin()
	if err != nil {
		// TODO: change default logger
		log.Printf("DB.Begin: %v", err)
		return nil, service.ErrInternal
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			// TODO: change default logger
			log.Printf("Tx.Rollback: %v", err)
		}
	}()

	if err := checkPost(tx, postID); err != nil {
		return nil, err
	}

	query := "UPDATE votes " +
		"SET vote = $1 " +
		"WHERE post_id = $2 AND user_id = $3"

	res, err := tx.Exec(query, vote, postID, userID)
	if err != nil {
		// TODO: change default logger
		log.Printf("tx.Exec: %v", err)
		return nil, service.ErrInternal
	}

	n, err := res.RowsAffected()
	if err != nil {
		// TODO: change default logger
		log.Printf("Result.RowsAffected: %v", err)
		return nil, service.ErrInternal
	}
	if n == 0 {

		query = "INSERT INTO votes (post_id, user_id, vote) " +
			"VALUES ($1, $2, $3)"

		if _, err := tx.Exec(
			query,
			postID,
			userID,
			vote,
		); err != nil {
			// TODO: change default logger
			log.Printf("tx.Exec: %v", err)
			return nil, service.ErrInternal
		}
	}

	post, err := get(tx, postID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		// TODO: change default logger
		log.Printf("Tx.Commit: %v", err)
		return nil, service.ErrInternal
	}

	return post, nil
}

func (r *PostRepo) DeleteVote(postID, userID int) (*entity.Post, error) {
	tx, err := r.db.Begin()
	if err != nil {
		// TODO: change default logger
		log.Printf("DB.Begin: %v", err)
		return nil, service.ErrInternal
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			// TODO: change default logger
			log.Printf("Tx.Rollback: %v", err)
		}
	}()

	if err := checkPost(tx, postID); err != nil {
		return nil, err
	}

	query := "DELETE FROM votes " +
		"WHERE post_id = $1 AND user_id = $2"

	if _, err := tx.Exec(
		query,
		postID,
		userID,
	); err != nil {
		// TODO: change default logger
		log.Printf("Tx.Exec: %v", err)
		return nil, service.ErrInternal
	}

	post, err := get(tx, postID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		// TODO: change default logger
		log.Printf("Tx.Commit: %v", err)
		return nil, service.ErrInternal
	}

	return post, nil
}

func (r *PostRepo) AddComment(postID, userID int, body string) (*entity.Post, error) {
	tx, err := r.db.Begin()
	if err != nil {
		// TODO: change default logger
		log.Printf("DB.Begin: %v", err)
		return nil, service.ErrInternal
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			// TODO: change default logger
			log.Printf("Tx.Rollback: %v", err)
		}
	}()

	if err := checkPost(tx, postID); err != nil {
		return nil, err
	}

	query := "INSERT INTO comments (post_id, user_id, body) " +
		"VALUES ($1, $2, $3)"

	if _, err := tx.Exec(
		query,
		postID,
		userID,
		body,
	); err != nil {
		// TODO: change default logger
		log.Printf("Tx.Exec: %v", err)
		return nil, service.ErrInternal
	}

	post, err := get(tx, postID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		// TODO: change default logger
		log.Printf("Tx.Commit: %v", err)
		return nil, service.ErrInternal
	}

	return post, nil
}

func checkComment(tx *sql.Tx, id int) error {

	query := "SELECT " +
		"FROM comments " +
		"WHERE id = $1"

	if err := tx.QueryRow(
		query,
		id,
	).Scan(); err != nil {
		var retErr error
		switch {
		case errors.Is(err, sql.ErrNoRows):
			retErr = service.ErrCommentNotFound
		default:
			// TODO: change default logger
			log.Printf("Tx.QueryRow: %v", err)
			retErr = service.ErrInternal
		}
		return retErr
	}

	return nil
}

func (r *PostRepo) DeleteComment(postID, commentID, userID int) (*entity.Post, error) {
	tx, err := r.db.Begin()
	if err != nil {
		// TODO: change default logger
		log.Printf("DB.Begin: %v", err)
		return nil, service.ErrInternal
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			// TODO: change default logger
			log.Printf("Tx.Rollback: %v", err)
		}
	}()

	if err := checkPost(tx, postID); err != nil {
		return nil, err
	}

	if err := checkComment(tx, commentID); err != nil {
		return nil, err
	}

	query := "DELETE FROM comments " +
		"WHERE id = $1 AND post_id = $2 AND user_id = $3"

	res, err := tx.Exec(query, commentID, postID, userID)
	if err != nil {
		// TODO: change default logger
		log.Printf("Tx.Exec: %v", err)
		return nil, service.ErrInternal
	}

	n, err := res.RowsAffected()
	if err != nil {
		// TODO: change default logger
		log.Printf("Result.RowsAffected: %v", err)
		return nil, service.ErrInternal
	}
	if n == 0 {
		return nil, service.ErrUnauthorized
	}

	post, err := get(tx, postID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		// TODO: change default logger
		log.Printf("Tx.Commit: %v", err)
		return nil, service.ErrInternal
	}

	return post, err
}

func (r *PostRepo) Delete(postID, userID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		// TODO: change default logger
		log.Printf("DB.Begin: %v", err)
		return service.ErrInternal
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			// TODO: change default logger
			log.Printf("Tx.Rollback: %v", err)
		}
	}()

	if err := checkPost(tx, postID); err != nil {
		return err
	}

	query := "DELETE FROM posts " +
		"WHERE id = $1 AND user_id = $2"

	res, err := tx.Exec(query, postID, userID)
	if err != nil {
		// TODO: change default logger
		log.Printf("Tx.Exec: %v", err)
		return service.ErrInternal
	}

	n, err := res.RowsAffected()
	if err != nil {
		// TODO: change default logger
		log.Printf("Result.RowsAffected: %v", err)
		return service.ErrInternal
	}
	if n == 0 {
		return service.ErrUnauthorized
	}

	if err := tx.Commit(); err != nil {
		// TODO: change default logger
		log.Printf("Tx.Commit: %v", err)
		return service.ErrInternal
	}

	return nil
}
