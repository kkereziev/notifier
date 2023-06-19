package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SlackMessage represents body for Slack message.
type SlackMessage struct {
	Text string `json:"text"`
}

// Slack holds Slack related configuration for sending notifications.
type Slack struct {
	webHookURL string
	client     *http.Client
}

// Service handles business logic for the application.
type Service struct {
	Slack *Slack
}

// NewService is a constructor function for Service.
func NewService(config *Config) *Service {
	return &Service{
		Slack: &Slack{
			client:     http.DefaultClient,
			webHookURL: config.SlackWebHookURL,
		},
	}
}

var _ Notifier = (*Service)(nil)

// NotifySlack sends Slack notification.
func (s *Service) NotifySlack(ctx context.Context, msg any) error {
	slackMsg := msg.(string)

	slackMessage := SlackMessage{
		Text: slackMsg,
	}

	payload, err := json.Marshal(slackMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.Slack.webHookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	resp, err := s.Slack.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack notification: %v", err)
	}

	//nolint: errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	return nil
}
