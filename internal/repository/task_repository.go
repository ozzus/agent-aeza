package repository

import (
	"context"
	"fmt"
	"time"

	"ozzus/agent-aeza/internal/domain"

	"ozzus/agent-aeza/internal/repository/kafka"
)

type TaskRepository interface {
	FetchTasks(ctx context.Context) ([]domain.Task, error)
}

type KafkaTaskRepository struct {
	consumer *kafka.Consumer
}

func NewKafkaTaskRepository(consumer *kafka.Consumer) TaskRepository {
	return &KafkaTaskRepository{
		consumer: consumer,
	}
}

func (r *KafkaTaskRepository) FetchTasks(ctx context.Context) ([]domain.Task, error) {
	var tasks []domain.Task

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for {
		var task domain.Task
		_, err := r.consumer.ReadEvent(timeoutCtx, &task)
		if err != nil {
			if err == context.DeadlineExceeded {
				break
			}
			return nil, fmt.Errorf("failed to read event: %w", err)
		}

		if len(tasks) >= 100 {
			break
		}
	}

	return tasks, nil
}
