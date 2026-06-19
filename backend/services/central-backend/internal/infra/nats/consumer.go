package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"mercadia.dev/pos/services/central-backend/internal/app"
)

const (
	streamName      = "MERCADIA_STORE_EDGE"
	subjectPrefix   = "mercadia.store-edge.sync"
	consumerDurable = "central-backend-sync"
)

type SyncEventAccepter interface {
	AcceptEvents(ctx context.Context, command app.AcceptSyncEventsCommand) (app.SyncEventsResult, error)
}

type Consumer struct {
	conn      *natsgo.Conn
	js        jetstream.JetStream
	consumer  jetstream.Consumer
	sync      SyncEventAccepter
	consume   jetstream.ConsumeContext
	connected atomic.Bool
}

type ConsumerOption func(*Consumer)

func NewConsumer(url string, sync SyncEventAccepter, options ...ConsumerOption) (*Consumer, error) {
	if url == "" {
		return nil, errors.New("nats url is required")
	}
	if sync == nil {
		return nil, errors.New("sync service is required")
	}

	conn, err := natsgo.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("create jetstream: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      streamName,
		Subjects:  []string{subjectPrefix + ".>"},
		Retention: jetstream.LimitsPolicy,
	})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure jetstream stream: %w", err)
	}

	consumer, err := js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Durable:       consumerDurable,
		FilterSubject: subjectPrefix + ".>",
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       30 * time.Second,
		MaxDeliver:    10,
	})
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure jetstream consumer: %w", err)
	}

	c := &Consumer{
		conn:     conn,
		js:       js,
		consumer: consumer,
		sync:     sync,
	}
	c.connected.Store(conn.IsConnected())

	conn.SetReconnectHandler(func(_ *natsgo.Conn) {
		c.connected.Store(true)
	})
	conn.SetDisconnectErrHandler(func(_ *natsgo.Conn, err error) {
		c.connected.Store(false)
		if err != nil {
			slog.Warn("central nats disconnected", "error", err)
		}
	})
	conn.SetClosedHandler(func(_ *natsgo.Conn) {
		c.connected.Store(false)
	})

	for _, option := range options {
		option(c)
	}

	return c, nil
}

func DefaultURL() string {
	return "nats://127.0.0.1:4222"
}

func (c *Consumer) Connected() bool {
	return c.connected.Load() && c.conn.IsConnected()
}

func (c *Consumer) HealthCheck(_ context.Context) error {
	if !c.Connected() {
		return errors.New("nats broker is not connected")
	}
	return nil
}

func (c *Consumer) Close() {
	if c.consume != nil {
		c.consume.Stop()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	c.connected.Store(false)
}

func (c *Consumer) Run(ctx context.Context) error {
	consumeCtx, err := c.consumer.Consume(func(msg jetstream.Msg) {
		if err := c.handleMessage(ctx, msg); err != nil {
			slog.Warn("sync event consume failed",
				"subject", msg.Subject(),
				"error", err,
			)
			if nakErr := msg.Nak(); nakErr != nil {
				slog.Warn("sync event nack failed", "error", nakErr)
			}
			return
		}
		if ackErr := msg.Ack(); ackErr != nil {
			slog.Warn("sync event ack failed", "error", ackErr)
		}
	}, jetstream.ConsumeErrHandler(func(_ jetstream.ConsumeContext, err error) {
		slog.Warn("sync consumer error", "error", err)
	}))
	if err != nil {
		return fmt.Errorf("start sync consumer: %w", err)
	}
	c.consume = consumeCtx

	slog.Info("central sync consumer started",
		"stream", streamName,
		"durable", consumerDurable,
		"filter", subjectPrefix+".>",
	)

	<-ctx.Done()
	consumeCtx.Stop()
	return ctx.Err()
}

func (c *Consumer) handleMessage(ctx context.Context, msg jetstream.Msg) error {
	return ProcessSyncMessage(ctx, c.sync, msg.Subject(), msg.Data())
}

func ProcessSyncMessage(ctx context.Context, sync SyncEventAccepter, subject string, data []byte) error {
	storeID, err := StoreIDFromSubject(subject)
	if err != nil {
		return err
	}

	syncMessage, err := DecodeSyncMessage(data)
	if err != nil {
		return err
	}

	result, err := sync.AcceptEvents(ctx, app.AcceptSyncEventsCommand{
		StoreID:        storeID,
		IdempotencyKey: IdempotencyKey(storeID, syncMessage.EventID),
		Events: []app.SyncEventInput{{
			EventID:    syncMessage.EventID,
			EventType:  syncMessage.EventType,
			OccurredAt: syncMessage.OccurredAt,
			Payload:    syncMessage.Payload,
		}},
	})
	if err != nil {
		return err
	}

	slog.Info("sync event accepted from nats",
		"store_id", storeID,
		"event_id", syncMessage.EventID,
		"event_type", syncMessage.EventType,
		"accepted", result.Accepted,
	)
	return nil
}

type SyncMessage struct {
	EventID    string          `json:"eventId"`
	EventType  string          `json:"eventType"`
	Payload    json.RawMessage `json:"payload"`
	OccurredAt time.Time       `json:"occurredAt"`
}

func DecodeSyncMessage(data []byte) (SyncMessage, error) {
	var message SyncMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return SyncMessage{}, fmt.Errorf("decode sync message: %w", err)
	}
	if message.EventID == "" || message.EventType == "" {
		return SyncMessage{}, errors.New("sync message missing eventId or eventType")
	}
	if len(message.Payload) == 0 {
		message.Payload = json.RawMessage(`{}`)
	}
	return message, nil
}

func StoreIDFromSubject(subject string) (string, error) {
	prefix := subjectPrefix + "."
	if !strings.HasPrefix(subject, prefix) {
		return "", fmt.Errorf("unexpected subject %q", subject)
	}
	storeID := strings.TrimPrefix(subject, prefix)
	if storeID == "" {
		return "", errors.New("subject missing store id")
	}
	return storeID, nil
}

func IdempotencyKey(storeID, eventID string) string {
	return fmt.Sprintf("nats:%s:%s", storeID, eventID)
}
