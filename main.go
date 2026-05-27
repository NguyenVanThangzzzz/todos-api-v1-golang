package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thangnguyen/todo_api_v1/internal/handler"
	"github.com/thangnguyen/todo_api_v1/internal/repository"
	"github.com/thangnguyen/todo_api_v1/internal/router"
	"github.com/thangnguyen/todo_api_v1/internal/usecase"
	"github.com/thangnguyen/todo_api_v1/pkg/logger"
	"github.com/thangnguyen/todo_api_v1/pkg/validator"
	"go.uber.org/zap"
)

func main() {
	log := logger.New()
	defer log.Sync()

	repo := repository.NewInMemoryTodoRepository()
	uc := usecase.NewTodoUsecase(repo)
	validate := validator.New()
	todoHandler := handler.NewTodoHandler(uc, validate, log)

	r := router.New(todoHandler, log)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:              ":" + port,
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
