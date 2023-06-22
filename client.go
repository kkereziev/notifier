package notifier

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/kkereziev/notifier/internal"
	"github.com/kkereziev/notifier/internal/adding"
	pb "github.com/kkereziev/notifier/internal/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Config holds configuration for the notifier client.
type Config struct {
	ServerPort              int           `validate:"required"`
	ServerHost              string        `validate:"required"`
	ServerConnectionTimeout int           `validate:"required"`
	Delay                   time.Duration `validate:"required"`
	Retries                 int           `validate:"required"`
}

// Addr returns the address of the server.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
}

// Client is a client for the notifier service.
// It is used to send notifications to the server.
type Client struct {
	conn    *grpc.ClientConn
	client  pb.NotificationServiceClient
	retries int
	delay   time.Duration
}

// NewClient returns a new client for the notifier service.
func NewClient(config *Config, options ...grpc.DialOption) (*Client, error) {
	if err := validator.New().Struct(config); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(config.ServerConnectionTimeout))
	defer cancel()

	conn, err := grpc.DialContext(ctx, config.Addr(), options...)
	if err != nil {
		return nil, errors.Wrap(err, "connecting to server failed:")
	}

	client := pb.NewNotificationServiceClient(conn)

	return &Client{conn: conn, client: client, delay: config.Delay, retries: config.Retries}, nil
}

// SendSlackNotification sends Slack notification for the server to process.
func (c *Client) SendSlackNotification(msg string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.delay*time.Duration(c.retries+5))
	defer cancel()

	ctxWithHeader := appendHeaders(ctx)

	err := internal.RetryGRPC(
		func(ctx context.Context, a any) error {
			_, err := c.client.SendSlackNotification(ctx, &pb.SendSlackNotificationRequest{Message: msg})
			if err != nil {
				return err
			}

			return nil
		},
		c.retries,
		c.delay,
	)(ctxWithHeader, nil)
	if err != nil {
		return err
	}

	return nil
}

// SendSMSNotification sends SMS notification for the server to process.
func (c *Client) SendSMSNotification(msg, sendTo string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.delay*time.Duration(c.retries)+5)
	defer cancel()

	ctxWithHeader := appendHeaders(ctx)

	err := internal.RetryGRPC(
		func(ctx context.Context, a any) error {
			_, err := c.client.SendSMSNotification(ctx, &pb.SendSMSNotificationRequest{Message: msg, SendTo: sendTo})
			if err != nil {
				return err
			}

			return nil
		},
		c.retries,
		c.delay,
	)(ctxWithHeader, nil)

	if err != nil {
		return err
	}

	return nil
}

// SendMailNotification sends mail notification for the server to process.
func (c *Client) SendMailNotification(msg, sendTo, subject string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.delay*time.Duration(c.retries+5))
	defer cancel()

	ctxWithHeader := appendHeaders(ctx)

	return internal.RetryGRPC(
		func(ctx context.Context, a any) error {
			_, err := c.client.SendMailNotification(
				ctx,
				&pb.SendMailNotificationRequest{Message: msg, SendTo: sendTo, Subject: subject},
			)
			if err != nil {
				return err
			}

			return nil
		},
		c.retries,
		c.delay,
	)(ctxWithHeader, nil)
}

// DefaultOptions returns default gRPC dial options.
func DefaultOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}
}

func appendHeaders(ctx context.Context) context.Context {
	md := metadata.New(map[string]string{adding.IdempotencyHeader: uuid.New().String()})

	return metadata.NewOutgoingContext(ctx, md)
}
