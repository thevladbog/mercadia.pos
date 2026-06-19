package nats

import (
	"context"
	"fmt"
	"time"

	natsgo "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type TestPublisher struct {
	conn *natsgo.Conn
	js   jetstream.JetStream
}

func NewTestPublisher(url string) (*TestPublisher, error) {
	conn, err := natsgo.Connect(url)
	if err != nil {
		return nil, err
	}
	js, err := jetstream.New(conn)
	if err != nil {
		conn.Close()
		return nil, err
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
		return nil, err
	}
	return &TestPublisher{conn: conn, js: js}, nil
}

func (p *TestPublisher) Publish(ctx context.Context, storeID string, body []byte) error {
	subject := fmt.Sprintf("%s.%s", subjectPrefix, storeID)
	_, err := p.js.Publish(ctx, subject, body)
	return err
}

func (p *TestPublisher) Close() {
	if p.conn != nil {
		p.conn.Close()
	}
}
