package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"user-microservice-golang/kafka"
)

// Producer wraps a sarama SyncProducer with structured logging
type Producer struct {
	client sarama.SyncProducer
	logger *zap.Logger
	source string // service name stamped on every event
}

// NewProducer creates and returns a ready-to-use Kafka SyncProducer
func NewProducer(brokers []string, logger *zap.Logger) (*Producer, error) {
	cfg := sarama.NewConfig()

	// Reliability settings
	cfg.Producer.RequiredAcks = sarama.WaitForAll // ack from all ISR replicas
	cfg.Producer.Retry.Max = 5
	cfg.Producer.Retry.Backoff = 250 * time.Millisecond
	cfg.Producer.Return.Successes = true
	cfg.Producer.Return.Errors = true

	// Message settings
	cfg.Producer.Compression = sarama.CompressionSnappy
	cfg.Producer.MaxMessageBytes = 1024 * 1024 // 1 MB

	// Idempotent producer (exactly-once delivery within a session)
	cfg.Producer.Idempotent = true
	cfg.Net.MaxOpenRequests = 1

	cfg.Version = sarama.V2_8_0_0

	client, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("kafka: failed to create producer: %w", err)
	}

	logger.Info("kafka producer connected", zap.Strings("brokers", brokers))

	return &Producer{
		client: client,
		logger: logger,
		source: "user-service",
	}, nil
}

// Publish sends an EventEnvelope to the given topic.
// The userID is used as the partition key so all events for a user
// land on the same partition (ordering guarantee).
func (p *Producer) Publish(ctx context.Context, topic, userID string, payload interface{}) error {
	event := kafka.NewEvent(topic, p.source, payload)

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("kafka: marshal event failed: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Key:       sarama.StringEncoder(userID), // partition by user
		Value:     sarama.ByteEncoder(data),
		Timestamp: event.Timestamp,
	}

	partition, offset, err := p.client.SendMessage(msg)
	if err != nil {
		p.logger.Error("kafka publish failed",
			zap.String("topic", topic),
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return fmt.Errorf("kafka: publish to %s failed: %w", topic, err)
	}

	p.logger.Debug("kafka event published",
		zap.String("topic", topic),
		zap.String("event_id", event.EventID),
		zap.String("event_type", event.EventType),
		zap.String("user_id", userID),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset),
	)

	return nil
}

// Close gracefully shuts down the producer
func (p *Producer) Close() error {
	if err := p.client.Close(); err != nil {
		p.logger.Error("kafka producer close failed", zap.Error(err))
		return err
	}
	p.logger.Info("kafka producer closed")
	return nil
}
