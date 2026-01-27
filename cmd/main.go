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

	"github.com/UnendingLoop/SalesTracker/internal/repository"
	"github.com/UnendingLoop/SalesTracker/internal/service"
	"github.com/UnendingLoop/SalesTracker/internal/transport"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/dbpg"

	"github.com/wb-go/wbf/ginext"
)

func main() {
	log.Println("Starting SalesTracker application...")
	// инициализировать конфиг/ считать энвы
	appConfig := config.New()
	appConfig.EnableEnv("")
	if err := appConfig.LoadEnvFiles("./.env"); err != nil {
		log.Fatalf("Failed to load envs: %s\nExiting app...", err)
	}

	// готовим заранее слушатель прерываний - контекст для всего приложения
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// подключитсья к базе
	dbConn := repository.ConnectWithRetries(appConfig, 5, 10*time.Second)
	// накатываем миграцию
	repository.MigrateWithRetries(dbConn.Master, "./migrations", 10, 15*time.Second)

	// repo
	repo := repository.NewOperationsRepo(dbConn)
	// service
	svc := service.NewOperationService(repo)
	// handlers
	handlers := transport.NewOperationHandler(svc)
	// конфиг сервера
	mode := appConfig.GetString("GIN_MODE")
	engine := ginext.New(mode)
	operations := engine.Group("/operations")
	analytics := engine.Group("/analytics")

	engine.GET("/ping", handlers.SimplePinger)
	engine.Static("/web", "./internal/web")

	operations.POST("", handlers.CreateOperation)
	operations.GET("/:id", handlers.GetOperationByID)
	operations.GET("", handlers.GetAllOperations)
	operations.PATCH("/:id", handlers.UpdateOperationByID)
	operations.DELETE("/:id", handlers.DeleteOperationByID)
	operations.GET("/csv", handlers.ExportOperationsCSV)

	analytics.GET("", handlers.GetAnalytics)
	analytics.GET("/csv", handlers.ExportAnalyticsCSV)

	srv := &http.Server{
		Addr:    ":" + appConfig.GetString("APP_PORT"),
		Handler: engine,
	}

	// запуск сервера
	go func() {
		log.Printf("Server running on http://localhost%s\n", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil {
			switch {
			case errors.Is(err, http.ErrServerClosed):
				log.Println("Server gracefully stopping...")
			default:
				log.Printf("Server stopped: %v", err)
				stop()
			}
		}
	}()

	// слушаем контекст прерываний для запуска Graceful Shutdown
	<-ctx.Done()
	shutdown(dbConn, srv)
}

func shutdown(dbConn *dbpg.DB, srv *http.Server) {
	log.Println("Interrupt received! Starting shutdown sequence...")

	// Closing Server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Println("Failed to shutdown server correctly:", err)
	} else {
		log.Println("Server is closed.")
	}

	// Closing DB connection
	if err := dbConn.Master.Close(); err != nil {
		log.Println("Failed to close DB-conn correctly:", err)
	} else {
		log.Println("DBconn is closed.")
	}
}
