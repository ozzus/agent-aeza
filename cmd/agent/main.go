package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ozzus/agent-aeza/config"
	"ozzus/agent-aeza/internal/api/http"
	"ozzus/agent-aeza/internal/checks"
	"ozzus/agent-aeza/internal/domain"
	"ozzus/agent-aeza/internal/repository"
	"ozzus/agent-aeza/internal/repository/kafka"
	"ozzus/agent-aeza/internal/service"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализируем Kafka компоненты
	taskConsumer := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topics.Tasks, cfg.Agent.Name)
	defer taskConsumer.Close()

	resultsProducer := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topics.Results)
	defer resultsProducer.Close()

	logsProducer := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topics.Logs)
	defer logsProducer.Close()

	// Создаем репозитории
	taskRepo := repository.NewKafkaTaskRepository(taskConsumer)
	resultRepo := repository.NewKafkaResultRepository(resultsProducer, logsProducer)

	// Создаем сервис агента
	agentService := service.NewAgentService(
		taskRepo,
		resultRepo,
		service.Config{
			AgentID:      cfg.Agent.Name,
			PollInterval: 30 * time.Second,
		},
	)

	// Регистрируем чекеры (временные заглушки)
	agentService.RegisterChecker(domain.TaskTypeHTTP, &checks.HTTPChecker{})
	agentService.RegisterChecker(domain.TaskTypePing, &checks.PingChecker{})
	agentService.RegisterChecker(domain.TaskTypeTCP, &checks.TCPChecker{})
	agentService.RegisterChecker(domain.TaskTypeTraceroute, &checks.TracerouteChecker{})
	agentService.RegisterChecker(domain.TaskTypeDNS, &checks.DNSChecker{})

	// Создаем health controller
	healthController := http.NewHealthController(agentService, cfg.Agent.Name)

	// Создаем и запускаем HTTP сервер
	router := http.NewRouter(healthController)

	// Создаем контекст с обработкой сигналов завершения
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем агент в горутине
	go func() {
		log.Printf("Starting agent service with ID: %s", cfg.Agent.Name)
		if err := agentService.Start(ctx); err != nil {
			log.Fatalf("Agent service failed: %v", err)
		}
	}()

	// Запускаем HTTP сервер в горутине
	go func() {
		log.Printf("Starting health server on :%s", cfg.Server.HealthPort)
		if err := router.Run(":" + cfg.Server.HealthPort); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Ожидаем сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down agent...")
	cancel()
	time.Sleep(1 * time.Second) // Даем время для graceful shutdown
	log.Println("Agent stopped")
}
