package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/ab-payment-system/internal/gateway/bsite/paypage"
	"github.com/ab-payment-system/internal/gateway/bsite/redirect"
	"github.com/ab-payment-system/internal/gateway/middleware"
	"github.com/ab-payment-system/internal/gateway/security"
	"github.com/ab-payment-system/internal/gateway/token"
	"github.com/ab-payment-system/internal/monitoring/health"
	"github.com/ab-payment-system/internal/monitoring/logging"
	"github.com/ab-payment-system/pkg/config"
	"github.com/ab-payment-system/pkg/logger"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	logger.Init(cfg.Server.Env, "info")
	log.Info().Msg("starting AB Payment Gateway Server")
	log.Info().Str("env", cfg.Server.Env).Msg("configuration loaded")

	// Initialize token generator
	tokenGen := token.NewGenerator(cfg.Server.JWTSecret)

	// Initialize security
	_ = security.NewOriginChecker() // originChecker reserved for future use

	// Bot detector
	botDetector := middleware.NewBotDetector()

	// Router
	router := mux.NewRouter()
	router.Use(logging.NewAccessLogger().Middleware())
	router.Use(botDetector.Middleware())

	// Health
	healthReg := health.NewHealthRegistry()
	router.Handle("/health", healthReg).Methods("GET")
	router.Handle("/ready", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	})).Methods("GET")

	// CORS
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Payment page handler
	payHandler := paypage.NewHandler(tokenGen)
	router.Handle("/pay/{token}", payHandler).Methods("GET")

	// Payment status redirect
	router.HandleFunc("/pay/success", func(w http.ResponseWriter, r *http.Request) {
		redirect.HandleSuccess(w, r, r.URL.Query().Get("order_id"),
			r.URL.Query().Get("amount"), r.URL.Query().Get("redirect_url"))
	}).Methods("GET")

	router.HandleFunc("/pay/failed", func(w http.ResponseWriter, r *http.Request) {
		redirect.HandleFailure(w, r, r.URL.Query().Get("order_id"),
			r.URL.Query().Get("reason"), r.URL.Query().Get("redirect_url"))
	}).Methods("GET")

	// A-site SDK script
	router.HandleFunc("/sdk/ab-bridge.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		// Serve the A-site jump script
		http.ServeFile(w, r, "internal/gateway/asite/jump_script.js")
	}).Methods("GET")

	// Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Info().Int("port", cfg.Server.Port).Msg("gateway server started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Info().Msg("shutdown complete")
}
