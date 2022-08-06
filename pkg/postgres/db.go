package postgres

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"time"
)

func New(
	logger *logrus.Logger,
	url string,
	connAttempts int,
	connTimeout time.Duration,
	maxOpenConns int,
) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	for connAttempts > 0 {
		if err = db.Ping(); err == nil {
			break
		}

		connAttempts--

		logger.Infof("trying to connect to the postgres server, attempts left: %d", connAttempts)

		time.Sleep(connTimeout)
	}

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)

	return db, nil
}
