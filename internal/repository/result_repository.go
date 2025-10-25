package repository

import (
	"context"
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
	if err := r.resultsProducer.PublishEvent(ctx, result.TaskID, result); err != nil {
		return fmt.Errorf("failed to publish result: %w", err)
	}
	return nil
}

func (r *KafkaResultRepository) SendLog(ctx context.Context, logEntry domain.LogEntry) error {
	key := fmt.Sprintf("%s-%d", logEntry.TaskID, logEntry.Timestamp.UnixNano())
	if err := r.logsProducer.PublishEvent(ctx, key, logEntry); err != nil {
		return fmt.Errorf("failed to publish log: %w", err)
	}
	return nil
}
