package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	topic  string
}

func NewConsumer(brokers []string, topic, groupID string) (*Consumer, error) {
	consumer := &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:           brokers,
			Topic:             topic,
			GroupID:           groupID,
			MaxWait:           10 * time.Second,
			StartOffset:       kafka.FirstOffset, // ИЛИ: kafka.LastOffset
			SessionTimeout:    30 * time.Second,
			HeartbeatInterval: 3 * time.Second,
			ReadBackoffMin:    100 * time.Millisecond,
			ReadBackoffMax:    1 * time.Second,
		}),
		topic: topic,
	}

	// Проверяем подключение сразу при создании
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := consumer.CheckConnection(ctx); err != nil {
		consumer.reader.Close() // Закрываем если не удалось подключиться
		return nil, fmt.Errorf("kafka connection failed: %w", err)
	}

	return consumer, nil
}

func (c *Consumer) CheckConnection(ctx context.Context) error {
	// Берем первый брокер для проверки
	if len(c.reader.Config().Brokers) == 0 {
		return fmt.Errorf("no brokers configured")
	}

	conn, err := kafka.DialContext(ctx, "tcp", c.reader.Config().Brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to kafka broker %s: %w",
			c.reader.Config().Brokers[0], err)
	}
	defer conn.Close()

	// Проверяем, что топик существует
	partitions, err := conn.ReadPartitions(c.topic)
	if err != nil {
		return fmt.Errorf("failed to read partitions for topic %s: %w", c.topic, err)
	}

	if len(partitions) == 0 {
		return fmt.Errorf("topic %s has no partitions", c.topic)
	}

	// Дополнительно: проверяем consumer group
	if c.reader.Config().GroupID != "" {
		resp, err := conn.FindCoordinator(kafka.CoordinatorGroup, c.reader.Config().GroupID)
		if err != nil {
			return fmt.Errorf("failed to find coordinator for group %s: %w",
				c.reader.Config().GroupID, err)
		}
		fmt.Printf("Coordinator for group %s: %s:%d\n",
			c.reader.Config().GroupID, resp.Host, resp.Port)
	}

	return nil
}
func (c *Consumer) ReadEvent(ctx context.Context, v interface{}) (kafka.Message, error) {
	slog.Debug("attempting to fetch message from kafka",
		"topic", c.topic,
		"group", c.reader.Config().GroupID)

	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		slog.Error("failed to fetch message", "error", err)
		return msg, err
	}

	slog.Debug("received message",
		"key", string(msg.Key),
		"partition", msg.Partition,
		"offset", msg.Offset,
		"value_length", len(msg.Value))

	if err := json.Unmarshal(msg.Value, v); err != nil {
		return msg, err
	}

	return msg, nil
}

func (c *Consumer) CommitMessage(ctx context.Context, msg kafka.Message) error {
	return c.reader.CommitMessages(ctx, msg)
}

func (c *Consumer) ReadRawMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.FetchMessage(ctx)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func (c *Consumer) Topic() string {
	return c.topic
}
