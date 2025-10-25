package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
			MaxWait: 10 * time.Second,
		}),
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
