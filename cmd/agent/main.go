package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	nethttp "net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	apihttp "ozzus/agent-aeza/internal/api/http"
	"ozzus/agent-aeza/internal/backend"
	"ozzus/agent-aeza/internal/checks"
	"ozzus/agent-aeza/internal/config"
	"ozzus/agent-aeza/internal/domain"
	"ozzus/agent-aeza/internal/lib/logger/slogpretty"
	"ozzus/agent-aeza/internal/repository"
	"ozzus/agent-aeza/internal/repository/kafka"
	"ozzus/agent-aeza/internal/service"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Настраиваем логгер
	log := setupLogger(cfg.Env)

	log.Info("starting application",
		"env", cfg.Env,
		"agent", cfg.Agent.Name,
	)

	backendClient, err := backend.NewClient(cfg.Backend.URL, cfg.Agent.Name, cfg.Agent.Token)
	if err != nil {
		log.Error("failed to initialize backend client", "error", err)
		os.Exit(1)
	}

	log.Info("initializing Kafka components")

	taskConsumer := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topics.Tasks, cfg.Agent.Name)
	defer taskConsumer.Close()

	resultsProducer := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topics.Results)
	defer resultsProducer.Close()

	logsProducer := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topics.Logs)
	defer logsProducer.Close()

	taskRepo := repository.NewKafkaTaskRepository(taskConsumer)
	resultRepo := repository.NewKafkaResultRepository(resultsProducer, logsProducer)

	agentService := service.NewAgentService(
		taskRepo,
		resultRepo,
		service.Config{
			AgentID:      cfg.Agent.Name,
			PollInterval: 30 * time.Second,
		},
	)

	log.Debug("initializing checkers")
	location := cfg.Agent.Name
	country := cfg.Agent.Country

	agentService.RegisterChecker(domain.TaskTypeHTTP, checks.NewHTTPChecker(cfg.GetHTTPTimeout(), location, country))
	agentService.RegisterChecker(domain.TaskTypePing, checks.NewPingChecker(cfg.GetPingTimeout(), 4, location, country))
	agentService.RegisterChecker(domain.TaskTypeTCP, checks.NewTCPChecker(cfg.GetTCPTimeout(), location, country))
	agentService.RegisterChecker(domain.TaskTypeTraceroute, checks.NewTracerouteChecker(30, cfg.GetPingTimeout(), location, country))
	agentService.RegisterChecker(domain.TaskTypeDNS, checks.NewDNSChecker(cfg.GetDNSTimeout(), location, country))

	healthController := apihttp.NewHealthController(agentService, cfg.Agent.Name)

	router := apihttp.NewRouter(healthController)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	startHeartbeat(ctx, &wg, backendClient, log, heartbeatInterval)

	// Запускаем агент
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("starting agent service",
			"backend", cfg.Backend.URL,
			"kafka_brokers", cfg.Kafka.Brokers,
		)
		if err := agentService.Start(ctx); err != nil {
			log.Error("agent service failed", "error", err)
			os.Exit(1)
		}
	}()

	// Запускаем HTTP сервер
	httpServer := &nethttp.Server{
		Addr:    ":" + cfg.Server.HealthPort,
		Handler: router,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("starting health server", "port", cfg.Server.HealthPort)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
			log.Error("HTTP server failed", "error", err)
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1) // ← ОБЪЯВЛЯЕМ ПЕРЕМЕННУЮ quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	log.Info("application started and ready",
		"health_port", cfg.Server.HealthPort,
		"agent_id", cfg.Agent.Name,
	)

	<-quit
	log.Info("shutting down agent...")
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server shutdown failed", "error", err)
	}

	wg.Wait()
	log.Info("agent stopped gracefully")
}

const (
	envLocal          = "local"
	envDev            = "dev"
	envProd           = "prod"
	heartbeatInterval = 30 * time.Second
)

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = setupPrettySlog()
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}

func startHeartbeat(ctx context.Context, wg *sync.WaitGroup, client *backend.Client, log *slog.Logger, interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Second
	}

	send := func() {
		hbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := client.Heartbeat(hbCtx); err != nil {
			log.Error("heartbeat failed", "error", err.Error())
			return
		}

		log.Debug("heartbeat sent")
	}

	if wg != nil {
		wg.Add(1)
	}

	go func() {
		if wg != nil {
			defer wg.Done()
		}

		send()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				send()
			case <-ctx.Done():
				log.Debug("heartbeat loop stopped")
				return
			}
		}
	}()
}
