package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"ozzus/agent-aeza/internal/domain"

	kafkago "github.com/segmentio/kafka-go" // <-- alias the client lib

	repokafka "ozzus/agent-aeza/internal/repository/kafka" // <-- alias your wrapper
)

type TaskRepository interface {
	FetchTasks(ctx context.Context) ([]domain.Task, error)
	AckTask(ctx context.Context, taskID string) error
	NackTask(taskID string)
}

type KafkaTaskRepository struct {
	consumer *repokafka.Consumer

	mu       sync.Mutex
	messages map[string]kafkago.Message
}

func NewKafkaTaskRepository(consumer *repokafka.Consumer) TaskRepository {
	return &KafkaTaskRepository{
		consumer: consumer,
		messages: make(map[string]kafkago.Message),
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
	r.mu.Unlock()

	if !ok {
		return nil
	}

	const maxRetries = 3

	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		timeout := 5 * time.Second
		if deadline, ok := ctx.Deadline(); ok {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return ctx.Err()
			}
			if remaining < timeout {
				timeout = remaining
			}
		}

		commitCtx, cancel := context.WithTimeout(context.Background(), timeout)

		if err := r.consumer.CommitMessage(commitCtx, msg); err != nil {
			cancel()
			lastErr = err

			if ctx.Err() != nil {
				break
			}

			time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
			continue
		}

		cancel()

		r.mu.Lock()
		delete(r.messages, taskID)
		r.mu.Unlock()

		return nil
	}

	return fmt.Errorf("failed to commit message: %w", lastErr)
}

func (r *KafkaTaskRepository) NackTask(taskID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.messages, taskID)
}
