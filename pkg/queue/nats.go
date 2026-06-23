package queue

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type NATSClient struct {
	Conn *nats.Conn
	JS   nats.JetStreamContext
}

type MessageHandler func(subject string, data []byte) error

func New(urls []string, token string) (*NATSClient, error) {
	opts := []nats.Option{nats.Name("ab-payment")}
	if token != "" {
		opts = append(opts, nats.Token(token))
	}
	nc, err := nats.Connect(urls[0], opts...)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}
	log.Info().Msg("nats connected")
	return &NATSClient{Conn: nc, JS: js}, nil
}

func (n *NATSClient) Publish(subject string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = n.JS.Publish(subject, b)
	return err
}

func (n *NATSClient) Subscribe(queue, subject string, handler MessageHandler) (*nats.Subscription, error) {
	return n.JS.QueueSubscribe(subject, queue, func(m *nats.Msg) {
		handler(m.Subject, m.Data)
	})
}

func (n *NATSClient) Close() {
	n.Conn.Close()
}
