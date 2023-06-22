package notifying

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kkereziev/notifier/internal/config"
	"github.com/kkereziev/notifier/internal/entity"
	"github.com/kkereziev/notifier/internal/storage"
	"github.com/twilio/twilio-go"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

// SlackRequestBody is an object containing data for Slack notification endpoint.
type SlackRequestBody struct {
	Message string `validate:"required" json:"message"`
}

// SMSRequestBody is an object containing data for SMS notification endpoint.
type SMSRequestBody struct {
	Message      string `validate:"required" json:"message"`
	SendToNumber string `validate:"required,e164" json:"send_to_number"`
}

// MailRequestBody is an object containing data for mail notification endpoint.
type MailRequestBody struct {
	Message string `validate:"required" json:"message"`
	SendTo  string `validate:"required,email" json:"send_to"`
	Subject string `validate:"required" json:"subject"`
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
	logger           *zap.SugaredLogger
	slack            *Slack
	twilio           *Twilio
	email            *Email
	tx               storage.Transactioner
	notificationRepo NotificationStorage
}

// NewService is a constructor function for Service.
func NewService(
	config *config.Config,
	logger *zap.SugaredLogger,
	repo NotificationStorage,
	tx storage.Transactioner,
) *Service {
	s := &Service{
		logger: logger,
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
			client: gomail.NewDialer(
				config.Mail.SMTPHost,
				config.Mail.SMTPPort,
				config.Mail.SMTPUsername,
				config.Mail.SMTPPassword,
			),
			messageSender: config.Mail.EmailSender,
		},
		notificationRepo: repo,
		tx:               tx,
	}

	s.email.client.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return s
}

// NotifySlack sends Slack notification.
func (s *Service) NotifySlack(ctx context.Context) error {
	return s.tx.Tx(ctx, func(ctx context.Context) error {
		notifications, err := s.notificationRepo.GetBatchByType(ctx, entity.SlackNotificationType, 100)
		if err != nil {
			return err
		}

		for _, notification := range notifications {
			err := s.callSlackClient(ctx, notification)
			if err != nil {
				s.logger.Errorf("Failed to send notification for %v: %v", notification, err)

				notification.Status = entity.FailedNotificationStatus
			} else {
				notification.Status = entity.SentNotificationStatus
			}

			notification.UpdatedAt = time.Now().UTC()

			if err := s.notificationRepo.Update(ctx, notification); err != nil {
				return fmt.Errorf("Failed to update notification %v: %v", notification, err)
			}
		}

		return nil
	})
}

// NotifyMail send mail notification.
func (s *Service) NotifyMail(ctx context.Context) error {
	return s.tx.Tx(ctx, func(ctx context.Context) error {
		notifications, err := s.notificationRepo.GetBatchByType(ctx, entity.EmailNotificationType, 100)
		if err != nil {
			return err
		}

		for _, notification := range notifications {
			err := s.callMailClient(notification)
			if err != nil {
				s.logger.Errorf("Failed to send notification for %v: %v", notification, err)

				notification.Status = entity.FailedNotificationStatus
			} else {
				notification.Status = entity.SentNotificationStatus
			}

			notification.UpdatedAt = time.Now().UTC()

			if err := s.notificationRepo.Update(ctx, notification); err != nil {
				return fmt.Errorf("Failed to update notification %v: %v", notification, err)
			}
		}

		return nil
	})
}

// NotifySMS sends SMS notification.
func (s *Service) NotifySMS(ctx context.Context) error {
	return s.tx.Tx(ctx, func(ctx context.Context) error {
		notifications, err := s.notificationRepo.GetBatchByType(ctx, entity.SMSNotificationType, 100)
		if err != nil {
			return err
		}

		for _, notification := range notifications {
			err := s.callSMSClient(notification)
			if err != nil {
				s.logger.Errorf("Failed to send notification for %v: %v", notification, err)

				notification.Status = entity.FailedNotificationStatus
			} else {
				notification.Status = entity.SentNotificationStatus
			}

			notification.UpdatedAt = time.Now().UTC()

			if err := s.notificationRepo.Update(ctx, notification); err != nil {
				return fmt.Errorf("Failed to update notification %v: %v", notification, err)
			}
		}

		return nil
	})
}

func (s *Service) callSMSClient(notification *entity.Notification) error {
	smsNotification, err := notification.SMSNotification()
	if err != nil {
		return fmt.Errorf("failed to unmarshal notification: %v", err)
	}

	params := &twilioApi.CreateMessageParams{}

	params.SetTo(smsNotification.SendTo)
	params.SetFrom(s.twilio.number)
	params.SetBody(smsNotification.Text)

	if _, err := s.twilio.client.Api.CreateMessage(params); err != nil {
		return err
	}

	return nil
}

func (s *Service) callMailClient(notification *entity.Notification) error {
	mailNotification, err := notification.MailNotification()
	if err != nil {
		return fmt.Errorf("failed to unmarshal notification: %v", err)
	}

	m := gomail.NewMessage()

	m.SetHeader("From", s.email.messageSender)
	m.SetHeader("To", mailNotification.SendTo)
	m.SetHeader("Subject", mailNotification.Subject)
	m.SetBody("text/plain", mailNotification.Text)

	if err := s.email.client.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func (s *Service) callSlackClient(ctx context.Context, notification *entity.Notification) error {
	slackMessage, err := notification.SlackNotification()
	if err != nil {
		return err
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

	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	return nil
}
