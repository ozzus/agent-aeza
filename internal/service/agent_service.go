package service

import (
	"context"
	"fmt"
	"time"

	"ozzus/agent-aeza/internal/domain"
	"ozzus/agent-aeza/internal/repository"
)

// Checker интерфейс (оставляем для будущего)
type Checker interface {
	Check(target string, parameters map[string]interface{}) (*domain.CheckResult, error)
	Type() domain.TaskType
}

// AgentService основной сервис агента
type AgentService struct {
	taskRepo     repository.TaskRepository
	resultRepo   repository.ResultRepository
	checkers     map[domain.TaskType]Checker // ПУСТОЙ - без заглушек!
	agentID      string
	pollInterval time.Duration
	isRunning    bool
}

// Config конфигурация сервиса
type Config struct {
	AgentID      string
	PollInterval time.Duration
}

func NewAgentService(
	taskRepo repository.TaskRepository,
	resultRepo repository.ResultRepository,
	config Config,
) *AgentService {
	if config.PollInterval == 0 {
		config.PollInterval = 30 * time.Second
	}

	return &AgentService{
		taskRepo:     taskRepo,
		resultRepo:   resultRepo,
		checkers:     make(map[domain.TaskType]Checker), // ПУСТОЙ массив!
		agentID:      config.AgentID,
		pollInterval: config.PollInterval,
		isRunning:    false,
	}
}

// RegisterChecker регистрирует checker для определенного типа задач
func (s *AgentService) RegisterChecker(taskType domain.TaskType, checker Checker) {
	s.checkers[taskType] = checker
	s.logInfo("Checker registered", map[string]interface{}{
		"task_type":      taskType,
		"total_checkers": len(s.checkers),
	})
}

// Start запускает основной цикл обработки задач
func (s *AgentService) Start(ctx context.Context) error {
	s.isRunning = true
	s.logInfo("Agent service started", map[string]interface{}{
		"agent_id":      s.agentID,
		"poll_interval": s.pollInterval,
		"checkers":      len(s.checkers), // Покажет 0 - это ЧЕСТНО
	})

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.processTasks(ctx); err != nil {
				s.logError("Failed to process tasks", map[string]interface{}{
					"error": err.Error(),
				})
			}
		case <-ctx.Done():
			s.isRunning = false
			s.logInfo("Agent service stopped", nil)
			return nil
		}
	}
}

// processTasks обрабатывает все доступные задачи
func (s *AgentService) processTasks(ctx context.Context) error {
	tasks, err := s.taskRepo.FetchTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch tasks: %w", err)
	}

	if len(tasks) == 0 {
		return nil // Нет задач - нормальная ситуация
	}

	s.logInfo("Found tasks to process", map[string]interface{}{
		"task_count": len(tasks),
	})

	var processedCount, skippedCount int

	for _, task := range tasks {
		// Пытаемся обработать задачу
		processed, err := s.tryProcessTask(ctx, task)
		if err != nil {
			s.logError("Task processing failed", map[string]interface{}{
				"task_id": task.ID,
				"error":   err.Error(),
			})
		}

		if processed {
			processedCount++
		} else {
			skippedCount++
		}
	}

	s.logInfo("Tasks processing summary", map[string]interface{}{
		"total":     len(tasks),
		"processed": processedCount,
		"skipped":   skippedCount,
	})

	return nil
}

// tryProcessTask пытается обработать задачу, возвращает true если обработана
func (s *AgentService) tryProcessTask(ctx context.Context, task domain.Task) (bool, error) {
	// Логируем получение задачи
	s.sendLog(ctx, domain.LogEntry{
		TaskID:    task.ID,
		AgentID:   s.agentID,
		Level:     domain.LogLevelInfo,
		Message:   fmt.Sprintf("Received %s task for %s", task.Type, task.Target),
		Timestamp: time.Now(),
	})

	// Проверяем, есть ли подходящий checker
	checker, exists := s.checkers[task.Type]
	if !exists {
		// НЕТ checker'а - логируем и пропускаем задачу
		s.sendLog(ctx, domain.LogEntry{
			TaskID:    task.ID,
			AgentID:   s.agentID,
			Level:     domain.LogLevelError,
			Message:   fmt.Sprintf("No checker available for task type: %s", task.Type),
			Timestamp: time.Now(),
		})
		return false, nil // Задача не обработана, но это не ошибка сервиса
	}

	// Есть checker - выполняем проверку
	startTime := time.Now()
	result, err := checker.Check(task.Target, task.Parameters)
	duration := time.Since(startTime)

	if err != nil {
		s.sendLog(ctx, domain.LogEntry{
			TaskID:    task.ID,
			AgentID:   s.agentID,
			Level:     domain.LogLevelError,
			Message:   fmt.Sprintf("Checker execution failed: %v", err),
			Timestamp: time.Now(),
		})
		return false, err
	}

	// Заполняем результат
	result.TaskID = task.ID
	result.AgentID = s.agentID
	result.Duration = duration.Milliseconds()
	result.Timestamp = time.Now()

	// Отправляем результат
	if err := s.resultRepo.SendResult(ctx, *result); err != nil {
		return false, fmt.Errorf("failed to send result: %w", err)
	}

	// Логируем успешное завершение
	s.sendLog(ctx, domain.LogEntry{
		TaskID:    task.ID,
		AgentID:   s.agentID,
		Level:     domain.LogLevelInfo,
		Message:   fmt.Sprintf("Check completed with status: %s", result.Status),
		Timestamp: time.Now(),
	})

	return true, nil
}

// Остальные методы без изменений...
func (s *AgentService) sendLog(ctx context.Context, logEntry domain.LogEntry) {
	if err := s.resultRepo.SendLog(ctx, logEntry); err != nil {
		fmt.Printf("Failed to send log: %v\n", err)
	}
}

func (s *AgentService) logInfo(message string, fields map[string]interface{}) {
	fmt.Printf("INFO [%s]: %s", s.agentID, message)
	if fields != nil {
		fmt.Printf(" - %v", fields)
	}
	fmt.Println()
}

func (s *AgentService) logError(message string, fields map[string]interface{}) {
	fmt.Printf("ERROR [%s]: %s", s.agentID, message)
	if fields != nil {
		fmt.Printf(" - %v", fields)
	}
	fmt.Println()
}

// HealthCheck возвращает статус здоровья сервиса
func (s *AgentService) HealthCheck(ctx context.Context) error {
	if !s.isRunning {
		return fmt.Errorf("service is not running")
	}
	// Сервис здоров, даже если нет checker'ов - это нормально на данном этапе
	return nil
}

// GetStatus возвращает статус сервиса
func (s *AgentService) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"agent_id":      s.agentID,
		"is_running":    s.isRunning,
		"poll_interval": s.pollInterval.String(),
		"checkers":      len(s.checkers),       // Покажет 0 - это ЧЕСТНО
		"status":        "RUNNING_NO_CHECKERS", // Явно указываем статус
	}
}
