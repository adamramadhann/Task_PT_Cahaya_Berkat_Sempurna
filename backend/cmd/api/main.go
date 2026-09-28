package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"task-management-backend/internal/config"
	"task-management-backend/internal/database"
	"task-management-backend/internal/middleware"
	"task-management-backend/internal/task"
)

func main() {
	cfg := config.Load()

	db, err := database.ConnectMySQL(cfg.DBDSN)
	if err != nil {
		log.Fatalf("mysql: %v", err)
	}
	defer db.Close()
	log.Printf("connected to mysql")

	rdb, err := database.ConnectRedis(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()
	log.Printf("connected to redis")

	repo := task.NewRepository(db)
	cache := task.NewRedisCache(rdb, cfg.CacheTTL)
	svc := task.NewService(repo, cache)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.ErrorHandler())
	task.Register(r.Group("/api"), svc)

	srv := &http.Server{Addr: ":" + cfg.AppPort, Handler: r}

	go func() {
		log.Printf("listening on :%s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	log.Printf("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
