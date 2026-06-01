package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thangnguyen/todo_api_v1/config"
	"github.com/thangnguyen/todo_api_v1/internal/domain"
	"github.com/thangnguyen/todo_api_v1/internal/modules/auth"
	"github.com/thangnguyen/todo_api_v1/internal/modules/todo"
	"github.com/thangnguyen/todo_api_v1/internal/router"
	"github.com/thangnguyen/todo_api_v1/pkg/database"
	"github.com/thangnguyen/todo_api_v1/pkg/jwt"
	"github.com/thangnguyen/todo_api_v1/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	log := logger.New()
	defer log.Sync()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	db, err := database.New(cfg.DB)
	if err != nil {
		log.Fatal("failed to connect database", zap.Error(err))
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Todo{}, &domain.RefreshToken{}); err != nil {
		log.Fatal("failed to migrate database", zap.Error(err))
	}
	log.Info("database connected & migrated")

	jwtManager := jwt.NewManager(cfg.JWT)

	// Auth module
	userRepo := auth.NewUserRepository(db)
	refreshRepo := auth.NewRefreshTokenRepository(db)
	authUC := auth.NewUsecase(userRepo, refreshRepo, jwtManager)
	authHandler := auth.NewHandler(authUC, log)

	// Todo module
	todoRepo := todo.NewRepository(db)
	todoUC := todo.NewUsecase(todoRepo)
	todoHandler := todo.NewHandler(todoUC, log)

	r := router.New(todoHandler, authHandler, jwtManager, log)

	srv := &http.Server{
		Addr:              ":" + cfg.App.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("server started", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server crashed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", zap.Error(err))
	}
	log.Info("server stopped")
}
