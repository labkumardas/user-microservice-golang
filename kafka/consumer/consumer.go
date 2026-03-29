package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"user-microservice-golang/kafka"
)

// Handler is the function signature each topic handler must implement
type Handler func(ctx context.Context, event kafka.EventEnvelope) error

// Consumer wraps a sarama ConsumerGroup
type Consumer struct {
	group   sarama.ConsumerGroup
	topics  []string
	handler *groupHandler
	logger  *zap.Logger
}

// NewConsumer creates a Kafka consumer group
func NewConsumer(brokers []string, groupID string, topics []string, logger *zap.Logger) (*Consumer, error) {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_8_0_0

	// Offset strategy: start from oldest on first run
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Offsets.AutoCommit.Enable = false // manual commit for at-least-once
	cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}
	cfg.Consumer.Return.Errors = true
	cfg.Consumer.MaxWaitTime = 500 * time.Millisecond

	group, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, fmt.Errorf("kafka: failed to create consumer group: %w", err)
	}

	logger.Info("kafka consumer group created",
		zap.String("group", groupID),
		zap.Strings("topics", topics),
	)

	return &Consumer{
		group:   group,
		topics:  topics,
		handler: newGroupHandler(logger),
		logger:  logger,
	}, nil
}

// RegisterHandler maps a topic to its handler function
func (c *Consumer) RegisterHandler(topic string, fn Handler) {
	c.handler.register(topic, fn)
	c.logger.Info("kafka handler registered", zap.String("topic", topic))
}

// Start begins consuming in a blocking loop; call in a goroutine.
// It respects ctx cancellation for graceful shutdown.
func (c *Consumer) Start(ctx context.Context) {
	c.logger.Info("kafka consumer started", zap.Strings("topics", c.topics))

	for {
		if err := c.group.Consume(ctx, c.topics, c.handler); err != nil {
			c.logger.Error("kafka consume error", zap.Error(err))
		}

		if ctx.Err() != nil {
			c.logger.Info("kafka consumer context cancelled — shutting down")
			return
		}

		// Brief pause before reconnect attempt
		time.Sleep(2 * time.Second)
	}
}

// Close gracefully shuts down the consumer group
func (c *Consumer) Close() error {
	if err := c.group.Close(); err != nil {
		c.logger.Error("kafka consumer close failed", zap.Error(err))
		return err
	}
	c.logger.Info("kafka consumer closed")
	return nil
}

// ─── sarama ConsumerGroupHandler implementation ───────────────────────────────

type groupHandler struct {
	handlers map[string]Handler
	logger   *zap.Logger
}

func newGroupHandler(logger *zap.Logger) *groupHandler {
	return &groupHandler{
		handlers: make(map[string]Handler),
		logger:   logger,
	}
}

func (h *groupHandler) register(topic string, fn Handler) {
	h.handlers[topic] = fn
}

// Setup is called at the beginning of a new session (sarama interface)
func (h *groupHandler) Setup(sarama.ConsumerGroupSession) error { return nil }

// Cleanup is called at the end of a session (sarama interface)
func (h *groupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim processes messages from a single partition claim
func (h *groupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.processMessage(session, msg)
	}
	return nil
}

func (h *groupHandler) processMessage(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	var event kafka.EventEnvelope
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.Error("kafka: failed to unmarshal event",
			zap.String("topic", msg.Topic),
			zap.Error(err),
		)
		// Mark as processed even on bad payload to avoid infinite retry
		session.MarkMessage(msg, "")
		return
	}

	h.logger.Info("kafka event received",
		zap.String("topic", msg.Topic),
		zap.String("event_id", event.EventID),
		zap.String("event_type", event.EventType),
		zap.Int32("partition", msg.Partition),
		zap.Int64("offset", msg.Offset),
	)

	fn, ok := h.handlers[msg.Topic]
	if !ok {
		h.logger.Warn("kafka: no handler registered for topic", zap.String("topic", msg.Topic))
		session.MarkMessage(msg, "")
		return
	}

	ctx := context.Background()
	if err := fn(ctx, event); err != nil {
		h.logger.Error("kafka: handler error",
			zap.String("topic", msg.Topic),
			zap.String("event_id", event.EventID),
			zap.Error(err),
		)
		// Do NOT mark — message will be redelivered (at-least-once guarantee)
		return
	}

	// Commit offset only after successful processing
	session.MarkMessage(msg, "")
}
