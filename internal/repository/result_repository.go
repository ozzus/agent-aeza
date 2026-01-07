package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"ozzus/agent-aeza/internal/domain"
	"ozzus/agent-aeza/internal/repository/kafka"
)

type ResultRepository interface {
	SendResult(ctx context.Context, result domain.CheckResult) error
	SendLog(ctx context.Context, logEntry domain.LogEntry) error
}

type KafkaResultRepository struct {
	resultsProducer *kafka.Producer
	logsProducer    *kafka.Producer
}

func NewKafkaResultRepository(resultsProducer, logsProducer *kafka.Producer) ResultRepository {
	return &KafkaResultRepository{
		resultsProducer: resultsProducer,
		logsProducer:    logsProducer,
	}
}

func (r *KafkaResultRepository) SendResult(ctx context.Context, result domain.CheckResult) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("WARN [ResultRepository]: failed to marshal result for logging - %v\n", err)
	}

	if err := r.resultsProducer.PublishEvent(ctx, result.TaskID, result); err != nil {
		return fmt.Errorf("failed to publish result: %w", err)
	}
	if len(resultJSON) > 0 {
		fmt.Printf("INFO [ResultRepository]: sent result for task %s to topic %s payload=%s\n", result.TaskID, r.resultsProducer.Topic(), string(resultJSON))
	} else {
		fmt.Printf("INFO [ResultRepository]: sent result for task %s to topic %s\n", result.TaskID, r.resultsProducer.Topic())
	}
	return nil
}

func (r *KafkaResultRepository) SendLog(ctx context.Context, logEntry domain.LogEntry) error {
	key := fmt.Sprintf("%s-%d", logEntry.TaskID, logEntry.Timestamp.UnixNano())
	if err := r.logsProducer.PublishEvent(ctx, key, logEntry); err != nil {
		return fmt.Errorf("failed to publish log: %w", err)
	}
	fmt.Printf("INFO [ResultRepository]: sent log for task %s to topic %s\n", logEntry.TaskID, r.logsProducer.Topic())
	return nil
}
