package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	topic  string
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	fmt.Printf("Connecting to Kafka brokers: %v\n", brokers)
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 1,    // важно!
			MaxBytes: 10e6, // 10 MB, можно меньше, но не 0
			MaxWait:  1 * time.Second,
		}),
		topic: topic,
	}
}

func (c *Consumer) ReadEvent(ctx context.Context, v interface{}) (kafka.Message, error) {
	msg, err := c.reader.FetchMessage(ctx)
	if err != nil {
		return msg, err
	}

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
