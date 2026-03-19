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
	"arabiyya.edu.mv/bell-system-backend/internal/scheduler"
	"arabiyya.edu.mv/bell-system-backend/internal/services"
	ws "arabiyya.edu.mv/bell-system-backend/internal/websocket"
	"arabiyya.edu.mv/bell-system-backend/pkg/clock"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	log, err := logger.New(cfg.Server.Environment)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer log.Close()

	clk, err := clock.New(cfg.Scheduler.Timezone)
	if err != nil {
		return fmt.Errorf("failed to load timezone: %w", err)
	}

	db, err := database.New(cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Repositories
	userRepo := database.NewUserRepository(db.DB, log)
	sessionRepo := database.NewSessionRepository(db.DB, log)
	scheduleDayRepo := database.NewScheduleDayRepository(db.DB, log)
	audioFileRepo := database.NewSystemAudioFileRepository(db.DB, log)
	scheduleItemRepo := database.NewScheduleItemRepository(db.DB, log, scheduleDayRepo, sessionRepo, audioFileRepo)
	stateRepo := database.NewSystemStateRepository(db.DB, log)

	// Services
	tokenSvc := services.NewTokenService(cfg.JWT.Secret, cfg.JWT.ExpiresIn)
	hasher := services.NewPasswordHasher()
	fileStore, err := services.NewLocalFileStorage(cfg.Storage.AudioDir)
	if err != nil {
		return fmt.Errorf("failed to initialize file storage: %w", err)
	}

	// WebSocket
	hub := ws.NewHub(log)
	hubDone := make(chan struct{})
	go hub.Run(hubDone)

	notifier := ws.NewNotifier(hub)
	wsHandler := ws.NewHandler(hub, tokenSvc, log, cfg.WebSocket)

	// Scheduler
	sessionRepo.NowFunc = clk.Now
	sched := scheduler.New(sessionRepo, scheduleItemRepo, stateRepo, notifier, log, cfg.Scheduler.CheckInterval)
	sched.SetNowFunc(clk.Now)
	schedDone := make(chan struct{})
	if cfg.Scheduler.Enabled {
		go sched.Run(schedDone)
	}

	reloadNotifier := &scheduler.ReloadNotifier{Inner: notifier, Scheduler: sched}

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo, tokenSvc, hasher)
	userHandler := handlers.NewUserHandler(userRepo, hasher)
	sessionHandler := handlers.NewSessionHandler(sessionRepo)
	scheduleHandler := handlers.NewScheduleHandler(scheduleItemRepo, scheduleDayRepo, sessionRepo, reloadNotifier)
	scheduleHandler.NowFunc = clk.Now
	audioHandler := handlers.NewAudioHandler(audioFileRepo, reloadNotifier)
	audioHandler.FileStorage = fileStore
	systemHandler := handlers.NewSystemHandler(stateRepo, sched, reloadNotifier)

	// Router
	apiRouter := router.New(tokenSvc, authHandler, userHandler, sessionHandler, scheduleHandler, audioHandler, systemHandler)

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

	serverCtx, serverStopCtx := context.WithCancel(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		select {
		case <-sig:
		case <-ctx.Done():
		}

		// Shut down scheduler first, then WebSocket hub
		if cfg.Scheduler.Enabled {
			close(schedDone)
			<-sched.Done()
		}
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
		return fmt.Errorf("server error: %w", err)
	}

	<-serverCtx.Done()
	log.Info("Server stopped")
	return nil
}
