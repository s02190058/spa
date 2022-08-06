package app

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"

	"github.com/s02190058/spa/internal/config"
	"github.com/s02190058/spa/internal/repo"
	"github.com/s02190058/spa/internal/service"
	"github.com/s02190058/spa/internal/transport/http"
	"github.com/s02190058/spa/pkg/hasher"
	"github.com/s02190058/spa/pkg/httpserver"
	"github.com/s02190058/spa/pkg/jwt"
	"github.com/s02190058/spa/pkg/postgres"
)

func Run(cfg *config.Config) {
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.Logger.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{})
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Postgres.Username,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
		cfg.Postgres.SSLMode,
	)
	db, err := postgres.New(
		logger,
		dbURL,
		cfg.Postgres.ConnAttempts,
		cfg.Postgres.ConnTimeout,
		cfg.Postgres.MaxOpenConns,
	)
	if err != nil {
		logger.Fatalf("postgres.New: %v", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Errorf("DB.Close: %v", err)
		}
	}()

	userRepo := repo.NewUserRepo(db)
	tokenManager, err := jwt.NewTokenManager(cfg.JWT.SigningKey, cfg.JWT.TokenTTL)
	if err != nil {
		logger.Fatalf("jwt.NewTokenManager: %v", err)
	}
	passwordHasher := hasher.New(cfg.Hasher.Cost)
	userService := service.NewUserService(userRepo, tokenManager, passwordHasher)

	postRepo := repo.NewPostRepo(db)
	postService := service.NewPostService(postRepo)

	router := http.NewRouter(logger, tokenManager, userService, postService, cfg.Static)
	server := httpserver.New(logger, router, cfg.Server.Port, cfg.Server.ShutdownTimeout)

	server.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case s := <-quit:
		logger.Infof("server interrupt: %s", s.String())
	case err := <-server.Notify():
		logger.Errorf("server: %v", err)
	}

	if err := server.Shutdown(); err != nil {
		logrus.Errorf("failed to shutdown a server: %v", err)
	}
}
