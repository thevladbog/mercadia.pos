package nats

import (
	"context"
	"errors"
	"fmt"
	"time"

	natsc "github.com/nats-io/nats.go"
)

type Connection struct {
	conn *natsc.Conn
}

func Connect(url string) (*Connection, error) {
	conn, err := natsc.Connect(url, natsc.Timeout(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}
	return &Connection{conn: conn}, nil
}

func (c *Connection) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Connection) Ping(ctx context.Context) error {
	if c.conn == nil {
		return errors.New("nats connection is nil")
	}
	if status := c.conn.Status(); status != natsc.CONNECTED {
		return fmt.Errorf("nats status: %s", status)
	}
	if err := c.conn.FlushTimeout(5 * time.Second); err != nil {
		return fmt.Errorf("nats flush: %w", err)
	}
	return nil
}
