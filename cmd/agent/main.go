package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"

	"ozzus/agent-aeza/internal/checks"
	"ozzus/agent-aeza/internal/domain"
)

type Checker interface {
	Type() domain.TaskType
	Check(target string, params map[string]interface{}) (*domain.CheckResult, error)
}

func main() {
	log := setupLogger()

	// ---- конфиг через env, без всяких конфиг-пакетов ----
	brokersEnv := getenvOr("KAFKA_BROKERS", "91.107.126.43:9092")
	taskTopic := getenvOr("KAFKA_TASKS_TOPIC", "agent-tasks")
	resultTopic := getenvOr("KAFKA_RESULTS_TOPIC", "check-results")
	agentID := getenvOr("AGENT_ID", "agent-1")
	location := getenvOr("AGENT_LOCATION", agentID)
	country := getenvOr("AGENT_COUNTRY", "unknown")

	brokers := strings.Split(brokersEnv, ",")

	log.Info("starting agent",
		"brokers", brokers,
		"taskTopic", taskTopic,
		"resultTopic", resultTopic,
		"agentID", agentID,
	)

	// ---- регистрируем чекеры ----
	checkers := buildCheckers(location, country)
	log.Info("checkers registered", "count", len(checkers))

	// ---- Kafka consumer (ТВОЙ РАБОЧИЙ ВАРИАНТ) ----
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    taskTopic,
		GroupID:  agentID, // каждый агент — свой consumer group
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	// ---- Kafka producer для результатов ----
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    resultTopic,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	// ---- контекст с отменой по сигналу ----
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Info("agent started, waiting for tasks...")

	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Info("context cancelled, stopping consumer loop")
				break
			}
			log.Error("failed to read message", "error", err)
			continue
		}

		var task domain.Task
		if err := json.Unmarshal(msg.Value, &task); err != nil {
			log.Error("failed to unmarshal task", "error", err, "raw", string(msg.Value))
			continue
		}

		if task.ID == "" {
			log.Warn("received task without ID, skipping")
			continue
		}

		log.Info("received task",
			"id", task.ID,
			"type", task.Type,
			"target", task.Target,
		)

		checker, ok := checkers[task.Type]
		if !ok {
			log.Error("no checker registered for task type", "type", task.Type)
			// отправим сразу фейл, чтобы бэк не висел
			res := domain.CheckResult{
				TaskID:    task.ID,
				AgentID:   agentID,
				Status:    domain.StatusFailed,
				Error:     "unsupported task type: " + string(task.Type),
				Duration:  0,
				Timestamp: time.Now().UTC(),
				Payload:   nil,
			}
			if err := sendResult(ctx, writer, log, res); err != nil {
				log.Error("failed to send error result", "error", err)
			}
			continue
		}

		start := time.Now()
		res, err := checker.Check(task.Target, task.Parameters)
		duration := time.Since(start)

		if res == nil {
			res = &domain.CheckResult{}
		}

		// если чекер вернул ошибку — считаем это фейлом проверки
		if err != nil {
			log.Error("checker failed", "error", err, "task_id", task.ID)
			res.Status = domain.StatusFailed
			if res.Error == "" {
				res.Error = err.Error()
			}
		}

		res.TaskID = task.ID
		res.AgentID = agentID
		res.Duration = duration.Milliseconds()
		if res.Timestamp.IsZero() {
			res.Timestamp = time.Now().UTC()
		}

		if err := sendResult(ctx, writer, log, *res); err != nil {
			log.Error("failed to send result", "error", err, "task_id", task.ID)
			continue
		}

		log.Info("task processed",
			"id", task.ID,
			"status", res.Status,
			"duration_ms", res.Duration,
		)
	}

	log.Info("agent stopped")
}

// ---------- helpers ----------

func buildCheckers(location, country string) map[domain.TaskType]Checker {
	checkers := []Checker{
		checks.NewHTTPChecker(10*time.Second, location, country),
		checks.NewPingChecker(5*time.Second, 4, location, country),
		checks.NewTCPChecker(5*time.Second, location, country),
		checks.NewTracerouteChecker(30, 3*time.Second, location, country),
		checks.NewDNSChecker(5*time.Second, location, country),
	}

	m := make(map[domain.TaskType]Checker, len(checkers))
	for _, c := range checkers {
		m[c.Type()] = c
	}
	return m
}

func sendResult(ctx context.Context, w *kafka.Writer, log *slog.Logger, res domain.CheckResult) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(res.TaskID),
		Value: data,
		Time:  time.Now(),
	}

	if err := w.WriteMessages(ctx, msg); err != nil {
		return err
	}
	return nil
}

func getenvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func setupLogger() *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)
}
