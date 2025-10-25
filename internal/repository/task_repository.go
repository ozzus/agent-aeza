package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"ozzus/agent-aeza/internal/domain"

	"ozzus/agent-aeza/internal/repository/kafka"
)

type TaskRepository interface {
	FetchTasks(ctx context.Context) ([]domain.Task, error)
	AckTask(ctx context.Context, taskID string) error
	NackTask(taskID string)
}

type KafkaTaskRepository struct {
	consumer *kafka.Consumer

	mu       sync.Mutex
	messages map[string]kafka.Message
}

func NewKafkaTaskRepository(consumer *kafka.Consumer) TaskRepository {
	return &KafkaTaskRepository{
		consumer: consumer,
		messages: make(map[string]kafka.Message),
	}
}

func (r *KafkaTaskRepository) FetchTasks(ctx context.Context) ([]domain.Task, error) {
	var tasks []domain.Task

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for len(tasks) < 100 {
		if err := timeoutCtx.Err(); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				break
			}
			return tasks, nil
		}

		var task domain.Task
		msg, err := r.consumer.ReadEvent(timeoutCtx, &task)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				break
			}
			if errors.Is(err, context.Canceled) {
				return tasks, nil
			}

			return nil, fmt.Errorf("failed to read event: %w", err)
		}

		if task.ID == "" {
			continue
		}

		r.mu.Lock()
		r.messages[task.ID] = msg
		r.mu.Unlock()

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *KafkaTaskRepository) AckTask(ctx context.Context, taskID string) error {
	r.mu.Lock()
	msg, ok := r.messages[taskID]
	if ok {
		delete(r.messages, taskID)
	}
	r.mu.Unlock()

	if !ok {
		return nil
	}

	if err := r.consumer.CommitMessage(ctx, msg); err != nil {
		return fmt.Errorf("failed to commit message: %w", err)
	}

	return nil
}

func (r *KafkaTaskRepository) NackTask(taskID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.messages, taskID)
}
