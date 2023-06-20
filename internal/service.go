package internal

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
	"gopkg.in/gomail.v2"
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

// Twilio holds related configuration for Twilio service, responsible for sending SMS notifications.
type Twilio struct {
	client *twilio.RestClient
	number string
}

// Email holds email related configuration for sending mail notifications.
type Email struct {
	client        *gomail.Dialer
	messageSender string
}

// Service handles business logic for the application.
type Service struct {
	slack  *Slack
	twilio *Twilio
	email  *Email
}

// NewService is a constructor function for Service.
func NewService(config *Config) *Service {
	s := &Service{
		slack: &Slack{
			client:     http.DefaultClient,
			webHookURL: config.SlackWebHookURL,
		},
		twilio: &Twilio{
			client: twilio.NewRestClientWithParams(twilio.ClientParams{
				Username: config.Twilio.SID,
				Password: config.Twilio.Token,
			}),
			number: config.Twilio.Number,
		},
		email: &Email{
			client:        gomail.NewDialer(config.Mail.SMTPHost, config.Mail.SMTPPort, config.Mail.SMTPUsername, config.Mail.SMTPPassword),
			messageSender: config.Mail.EmailSender,
		},
	}

	s.email.client.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return s
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

	req, err := http.NewRequest(http.MethodPost, s.slack.webHookURL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	resp, err := s.slack.client.Do(req)
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

// NotifySMS sends SMS notification.
func (s *Service) NotifySMS(ctx context.Context, msg any) error {
	twilioMsg := msg.(*SMSRequestBody)

	params := &twilioApi.CreateMessageParams{}

	params.SetTo(twilioMsg.SendToNumber)
	params.SetFrom(s.twilio.number)
	params.SetBody(twilioMsg.Message)

	_, err := s.twilio.client.Api.CreateMessage(params)
	if err != nil {
		return err
	}

	return nil
}

// NotifyMail send mail notification.
func (s *Service) NotifyMail(_ context.Context, msg any) error {
	m := gomail.NewMessage()

	mailContent := msg.(*MailRequestBody)

	m.SetHeader("From", s.email.messageSender)
	m.SetHeader("To", mailContent.SendTo)
	m.SetHeader("Subject", mailContent.Subject)
	m.SetBody("text/plain", mailContent.Message)

	err := s.email.client.DialAndSend(m)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
