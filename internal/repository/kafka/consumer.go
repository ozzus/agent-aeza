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

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			StartOffset: kafka.FirstOffset,
			Topic:       topic,
			GroupID:     groupID,
			MaxWait:     10 * time.Second,
		}),
		topic: topic,
	}
}
func (c *Consumer) CheckConnection(ctx context.Context) error {
	conn, err := kafka.DialContext(ctx, "tcp", c.reader.Config().Brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to kafka: %w", err)
	}
	defer conn.Close()

	// Проверить топик
	partitions, err := conn.ReadPartitions(c.topic)
	if err != nil {
		return fmt.Errorf("failed to read partitions: %w", err)
	}

	log.Info("kafka connection ok", "topic", c.topic, "partitions", len(partitions))
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
