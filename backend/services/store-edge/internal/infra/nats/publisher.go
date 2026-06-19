package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"mercadia.dev/pos/services/store-edge/internal/app"
	"mercadia.dev/pos/services/store-edge/internal/domain"
)

const (
	streamName    = "MERCADIA_STORE_EDGE"
	subjectPrefix = "mercadia.store-edge.sync"
)

type Publisher struct {
	conn     *natsgo.Conn
	js       jetstream.JetStream
	outbox   app.OutboxRepository
	interval time.Duration
	batch    int
	now      func() time.Time

	connected atomic.Bool
}

type PublisherOption func(*Publisher)

func WithPublisherInterval(interval time.Duration) PublisherOption {
	return func(publisher *Publisher) {
		publisher.interval = interval
	}
}

func WithPublisherBatchSize(batch int) PublisherOption {
	return func(publisher *Publisher) {
		publisher.batch = batch
	}
}

func NewPublisher(url string, outbox app.OutboxRepository, options ...PublisherOption) (*Publisher, error) {
	if url == "" {
		return nil, errors.New("nats url is required")
	}
	if outbox == nil {
		return nil, errors.New("outbox repository is required")
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

	publisher := &Publisher{
		conn:     conn,
		js:       js,
		outbox:   outbox,
		interval: time.Second,
		batch:    100,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
	publisher.connected.Store(conn.IsConnected())

	conn.SetReconnectHandler(func(_ *natsgo.Conn) {
		publisher.connected.Store(true)
	})
	conn.SetDisconnectErrHandler(func(_ *natsgo.Conn, err error) {
		publisher.connected.Store(false)
		if err != nil {
			slog.Warn("nats disconnected", "error", err)
		}
	})
	conn.SetClosedHandler(func(_ *natsgo.Conn) {
		publisher.connected.Store(false)
	})

	for _, option := range options {
		option(publisher)
	}

	return publisher, nil
}

func (p *Publisher) Connected() bool {
	return p.connected.Load() && p.conn.IsConnected()
}

func (p *Publisher) HealthCheck(_ context.Context) error {
	if !p.Connected() {
		return errors.New("nats broker is not connected")
	}
	return nil
}

func (p *Publisher) Close() {
	if p.conn != nil {
		p.conn.Close()
	}
	p.connected.Store(false)
}

func (p *Publisher) Run(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.publishPending(ctx); err != nil {
				slog.Warn("outbox publish cycle failed", "error", err)
			}
		}
	}
}

func (p *Publisher) publishPending(ctx context.Context) error {
	events, err := p.outbox.ListPendingOutboxEvents(ctx, p.batch)
	if err != nil {
		return err
	}
	for _, event := range events {
		if err := p.publishEvent(ctx, event); err != nil {
			slog.Warn("publish outbox event failed", "eventId", event.ID, "error", err)
		}
	}
	return nil
}

type syncMessage struct {
	EventID    string          `json:"eventId"`
	EventType  string          `json:"eventType"`
	Payload    json.RawMessage `json:"payload"`
	OccurredAt time.Time       `json:"occurredAt"`
}

func (p *Publisher) publishEvent(ctx context.Context, event domain.OutboxEvent) error {
	storeID, err := storeIDFromOutboxEvent(event)
	if err != nil {
		return err
	}

	body, err := json.Marshal(syncMessage{
		EventID:    event.ID,
		EventType:  event.EventType,
		Payload:    event.Payload,
		OccurredAt: event.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("marshal sync message: %w", err)
	}

	subject := fmt.Sprintf("%s.%s", subjectPrefix, storeID)
	if _, err := p.js.Publish(ctx, subject, body); err != nil {
		return fmt.Errorf("publish jetstream message: %w", err)
	}

	updated, err := p.outbox.MarkOutboxEventPublished(ctx, event.ID, p.now())
	if err != nil {
		return fmt.Errorf("mark outbox event published: %w", err)
	}
	if !updated {
		slog.Debug("outbox event already published", "eventId", event.ID)
	}
	return nil
}

func storeIDFromOutboxEvent(event domain.OutboxEvent) (string, error) {
	var payload struct {
		StoreID string `json:"storeId"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return "", fmt.Errorf("decode outbox payload storeId: %w", err)
	}
	if payload.StoreID == "" {
		return "", errors.New("outbox payload missing storeId")
	}
	return payload.StoreID, nil
}
