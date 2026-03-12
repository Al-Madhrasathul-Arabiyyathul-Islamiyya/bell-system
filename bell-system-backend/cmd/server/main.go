// cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"arabiyya.edu.mv/bell-system-backend/config"
	"arabiyya.edu.mv/bell-system-backend/internal/database"
	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/router"
	"arabiyya.edu.mv/bell-system-backend/internal/services"
	ws "arabiyya.edu.mv/bell-system-backend/internal/websocket"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(cfg.Server.Environment)
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}
	defer db.Close()

	// Repositories
	userRepo := database.NewUserRepository(db.DB, log)
	sessionRepo := database.NewSessionRepository(db.DB, log)
	scheduleDayRepo := database.NewScheduleDayRepository(db.DB, log)
	audioFileRepo := database.NewSystemAudioFileRepository(db.DB, log)
	scheduleItemRepo := database.NewScheduleItemRepository(db.DB, log, scheduleDayRepo, sessionRepo, audioFileRepo)

	// Services
	tokenSvc := services.NewTokenService(cfg.JWT.Secret, cfg.JWT.ExpiresIn)
	hasher := services.NewPasswordHasher()
	fileStore, err := services.NewLocalFileStorage(cfg.Storage.AudioDir)
	if err != nil {
		log.Fatal("Failed to initialize file storage", err)
	}

	// WebSocket
	hub := ws.NewHub(log)
	hubDone := make(chan struct{})
	go hub.Run(hubDone)

	notifier := ws.NewNotifier(hub)
	wsHandler := ws.NewHandler(hub, tokenSvc, log, cfg.WebSocket)

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo, tokenSvc, hasher)
	userHandler := handlers.NewUserHandler(userRepo, hasher)
	sessionHandler := handlers.NewSessionHandler(sessionRepo)
	scheduleHandler := handlers.NewScheduleHandler(scheduleItemRepo, scheduleDayRepo, notifier)
	audioHandler := handlers.NewAudioHandler(audioFileRepo, notifier)
	audioHandler.FileStorage = fileStore

	// Router
	apiRouter := router.New(authHandler, userHandler, sessionHandler, scheduleHandler, audioHandler)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           cfg.CORS.MaxAge,
	}))

	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// WebSocket endpoint — mounted before timeout middleware so long-lived connections aren't killed
	r.Handle("/ws", wsHandler)

	// API routes with timeout middleware
	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(time.Second * 30))
		r.Mount("/", apiRouter)
	})

	server := &http.Server{
		Addr:        fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:     r,
		ReadTimeout: time.Duration(cfg.Server.ReadTimeout) * time.Second,
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shut down WebSocket hub first
		close(hubDone)
		<-hub.Done()

		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("Graceful shutdown timed out.. forcing exit.", nil)
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal("Server shutdown error", err)
		}
		serverStopCtx()
	}()

	log.Info(fmt.Sprintf("Starting server on port %d", cfg.Server.Port))
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal("Server error", err)
	}

	<-serverCtx.Done()
	log.Info("Server stopped")
}
