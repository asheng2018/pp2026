package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

	accHealth "github.com/ab-payment-system/internal/account/health"
	accRepo "github.com/ab-payment-system/internal/account/repository"
	accSvc "github.com/ab-payment-system/internal/account/service"
	"github.com/ab-payment-system/internal/account/state"
	adminHandler "github.com/ab-payment-system/internal/admin/handler"
	adminMiddleware "github.com/ab-payment-system/internal/admin/middleware"
	adminRepo "github.com/ab-payment-system/internal/admin/repository"
	adminSvc "github.com/ab-payment-system/internal/admin/service"
	merchantRepo "github.com/ab-payment-system/internal/merchant/repository"
	monHealth "github.com/ab-payment-system/internal/monitoring/health"
	"github.com/ab-payment-system/internal/monitoring/logging"
	orderRepo "github.com/ab-payment-system/internal/order/repository"
	"github.com/ab-payment-system/internal/scheduler/dto"
	"github.com/ab-payment-system/internal/scheduler/engine"
	"github.com/ab-payment-system/pkg/config"
	"github.com/ab-payment-system/pkg/db"
	"github.com/shopspring/decimal"
	"github.com/ab-payment-system/pkg/logger"
	pkgredis "github.com/ab-payment-system/pkg/redis"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	logger.Init(cfg.Server.Env, "info")
	log.Info().Msg("starting AB Payment Orchestrator")
	log.Info().Str("env", cfg.Server.Env).Msg("configuration loaded")

	// Database
	database, err := db.New(cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode, cfg.Database.MaxConns)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer database.Close()

	// Redis
	redisClient, err := pkgredis.New(cfg.Redis.Addrs, cfg.Redis.Password, cfg.Redis.DB, cfg.Redis.PoolSize)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
	}
	defer redisClient.Close()
	log.Info().Msg("redis connected")

	// Initialize account module
	accountRepo := accRepo.NewAccountRepo(database.DB)
	stateManager := state.NewStateManager(redisClient.Client)
	healthChecker := accHealth.New(redisClient.Client, true, cfg.Risk.MaxConsecutiveFails, cfg.Risk.DefaultThrottleTTL)
	accountService := accSvc.NewAccountService(accountRepo, stateManager, healthChecker, nil)

	// Initialize scheduler orchestrator
	orch := engine.NewOrchestrator(accountService)

	// Initialize admin module
	adminRepository := adminRepo.NewAdminRepository(database.DB)
	tokenService := adminSvc.NewTokenService(cfg.Server.JWTSecret, cfg.Server.AdminJWTExpiry)
	tokenBlacklist := adminSvc.NewTokenBlacklist(redisClient.Client)
	adminService := adminSvc.NewAdminService(adminRepository, tokenService, tokenBlacklist)
	adminH := adminHandler.NewAdminHandler(adminService)
	jwtMiddleware := adminMiddleware.Middleware(tokenService, tokenBlacklist)

	// Admin data handlers
	dashHandler := adminHandler.NewDashboardHandler(database.DB)
	orderH := adminHandler.NewOrderHandler(orderRepo.NewOrderRepo(database.DB), database.DB)
	accountH := adminHandler.NewAccountHandler(database.DB)
	merchantH := adminHandler.NewMerchantHandler(merchantRepo.NewMerchantRepo(database.DB), cfg.Server.JWTSecret)
	settlementH := adminHandler.NewSettlementListHandler(database.DB)
	proxyH := adminHandler.NewProxyHandler(database.DB)
	bSiteH := adminHandler.NewBSiteHandler(database.DB)
	riskH := adminHandler.NewRiskListHandler(database.DB)
	exchangeH := adminHandler.NewExchangeRateHandler(database.DB)
	logisticsH := adminHandler.NewLogisticsHandler(database.DB)

	// HTTP Router
	router := mux.NewRouter()

	// Health check
	healthReg := monHealth.NewHealthRegistry()
	healthReg.Register(monHealth.NewDBHealthChecker("postgres", database.PingContext))
	healthReg.Register(monHealth.NewRedisHealthChecker("redis", func(ctx context.Context) error { return redisClient.Ping(ctx).Err() }))
	router.Handle("/health", healthReg).Methods("GET")
	router.Handle("/ready", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	}))

	// Metrics
	router.Handle("/metrics", promhttp.Handler())

	// API v1
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(logging.NewAccessLogger().Middleware())

	// Admin auth routes (public)
	adminPublic := api.PathPrefix("/admin").Subrouter()
	adminPublic.HandleFunc("/login", adminH.Login).Methods("POST")

	// Admin protected routes
	adminProtected := api.PathPrefix("/admin").Subrouter()
	adminProtected.Use(jwtMiddleware)
	adminProtected.HandleFunc("/logout", adminH.Logout).Methods("POST")
	adminProtected.HandleFunc("/me", adminH.Me).Methods("GET")

	// Dashboard & data routes (JWT protected)
	adminProtected.HandleFunc("/dashboard/stats", dashHandler.GetStats).Methods("GET")
	adminProtected.HandleFunc("/dashboard/revenue", orderH.GetDailyRevenue).Methods("GET")
	adminProtected.HandleFunc("/orders", orderH.List).Methods("GET")
	adminProtected.HandleFunc("/orders/{id}", orderH.GetByID).Methods("GET")
	// Accounts CRUD
	adminProtected.HandleFunc("/accounts", accountH.List).Methods("GET")
	adminProtected.HandleFunc("/accounts", accountH.Create).Methods("POST")
	adminProtected.HandleFunc("/accounts/{id}", accountH.Update).Methods("PUT")
	adminProtected.HandleFunc("/accounts/{id}", accountH.Delete).Methods("DELETE")
	adminProtected.HandleFunc("/accounts/{id}/status", accountH.UpdateStatus).Methods("PATCH")
	// Merchants CRUD + API Keys
	adminProtected.HandleFunc("/merchants", merchantH.List).Methods("GET")
	adminProtected.HandleFunc("/merchants", merchantH.Create).Methods("POST")
	adminProtected.HandleFunc("/merchants/{id}", merchantH.GetByID).Methods("GET")
	adminProtected.HandleFunc("/merchants/{id}", merchantH.Update).Methods("PUT")
	adminProtected.HandleFunc("/merchants/{id}", merchantH.Delete).Methods("DELETE")
	adminProtected.HandleFunc("/merchants/{id}/apikeys", merchantH.GenerateAPIKey).Methods("POST")
	adminProtected.HandleFunc("/merchants/{id}/apikeys", merchantH.ListAPIKeys).Methods("GET")
	adminProtected.HandleFunc("/merchants/{id}/apikeys/{keyId}", merchantH.RevokeAPIKey).Methods("DELETE")
	// Settlements
	adminProtected.HandleFunc("/settlements", settlementH.List).Methods("GET")
	// Proxies CRUD + Batch Import
	adminProtected.HandleFunc("/proxies", proxyH.List).Methods("GET")
	adminProtected.HandleFunc("/proxies", proxyH.Create).Methods("POST")
	adminProtected.HandleFunc("/proxies/batch", proxyH.BatchImport).Methods("POST")
	adminProtected.HandleFunc("/proxies/{id}", proxyH.Update).Methods("PUT")
	adminProtected.HandleFunc("/proxies/{id}", proxyH.Delete).Methods("DELETE")
	adminProtected.HandleFunc("/proxies/{id}/status", proxyH.UpdateStatus).Methods("PATCH")
	// B-Sites CRUD
	adminProtected.HandleFunc("/b-sites", bSiteH.List).Methods("GET")
	adminProtected.HandleFunc("/b-sites", bSiteH.Create).Methods("POST")
	adminProtected.HandleFunc("/b-sites/{id}", bSiteH.GetByID).Methods("GET")
	adminProtected.HandleFunc("/b-sites/{id}", bSiteH.Update).Methods("PUT")
	adminProtected.HandleFunc("/b-sites/{id}", bSiteH.Delete).Methods("DELETE")
	adminProtected.HandleFunc("/b-sites/{id}/status", bSiteH.UpdateStatus).Methods("PATCH")
	// Risk Events
	adminProtected.HandleFunc("/risk-events", riskH.List).Methods("GET")
	// Exchange Rates
	adminProtected.HandleFunc("/exchange-rates", exchangeH.List).Methods("GET")
	// Logistics
	adminProtected.HandleFunc("/logistics", logisticsH.List).Methods("GET")

	// Payment allocation (public — called by A-site plugin via API key)
	api.HandleFunc("/allocate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" { w.WriteHeader(200); return }
		var req struct {
			OrderID    string `json:"order_id"`
			Amount     string `json:"amount"`
			Currency   string `json:"currency"`
			MerchantID string `json:"merchant_id"`
			Gateway    string `json:"gateway"`
			Strategy   string `json:"strategy,omitempty"`
			Country    string `json:"country,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
			return
		}
		result, err := orch.Allocate(r.Context(), &dto.AllocateRequest{
			OrderID: req.OrderID, Amount: decimal.RequireFromString(req.Amount), Currency: req.Currency,
			MerchantID: req.MerchantID, Gateway: req.Gateway, Strategy: req.Strategy, Country: req.Country,
		})
		if err != nil {
			log.Error().Err(err).Str("order_id", req.OrderID).Msg("allocation failed")
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusServiceUnavailable)
			return
		}
		json.NewEncoder(w).Encode(result)
	}).Methods("POST", "OPTIONS")

	api.HandleFunc("/release", func(w http.ResponseWriter, r *http.Request) {
		var req struct{ OrderID, MerchantID string }
		json.NewDecoder(r.Body).Decode(&req)
		orch.Release(r.Context(), &dto.ReleaseRequest{OrderID: req.OrderID, MerchantID: req.MerchantID})
		w.Write([]byte(`{"success":true}`))
	}).Methods("POST")

	// Proxy allocation for anti-tracking
	api.HandleFunc("/proxy-allocate", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Gateway    string `json:"gateway"`
			MerchantID string `json:"merchant_id"`
			Country    string `json:"country"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		var proxyID, proxyHost, proxyType, proxyCountry string
		var proxyPort int
		err := database.DB.QueryRowContext(r.Context(),
			`SELECT p.id, p.host, p.port, p.proxy_type, p.country
			 FROM proxies p WHERE p.status = 'online' AND p.success_rate >= 0.8
			 ORDER BY p.success_rate DESC, p.latency ASC LIMIT 1`,
		).Scan(&proxyID, &proxyHost, &proxyPort, &proxyType, &proxyCountry)
		if err != nil {
			http.Error(w, `{"error":"no available proxy"}`, http.StatusServiceUnavailable)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"proxy_id":   proxyID,
			"host":       proxyHost,
			"port":       proxyPort,
			"type":       proxyType,
			"country":    proxyCountry,
			"gateway":    req.Gateway,
			"expires_in": 600,
		})
	}).Methods("POST")

	// Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Info().Int("port", cfg.Server.Port).Msg("HTTP server started")
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
