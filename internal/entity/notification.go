package entity

import (
	"encoding/json"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// NotificationType is a type that holds eligible notification types.
type NotificationType string

const (
	// SMSNotificationType is a type that indicates that the notification is of type SMS.
	SMSNotificationType NotificationType = "SMS"

	// SlackNotificationType is a type that indicates that the notification is of type Slack.
	SlackNotificationType NotificationType = "SLACK"

	// EmailNotificationType is a type that indicates that the notification is of type Email.
	EmailNotificationType NotificationType = "EMAIL"
)

// NotificationStatus is a type that holds eligible notification statuses.
type NotificationStatus string

const (
	// FailedNotificationStatus is a status that indicates that the notification sending has failed.
	FailedNotificationStatus NotificationStatus = "FAILED"

	// SentNotificationStatus is a status that indicates that the notification has been sent.
	SentNotificationStatus NotificationStatus = "SENT"

	// PendingNotificationStatus is a status that indicates that the notification is yet to be sent.
	PendingNotificationStatus NotificationStatus = "PENDING"
)

// Notification is a struct that holds the notification data.
type Notification struct {
	ID        uuid.UUID          `db:"id"`
	Type      NotificationType   `db:"type"`
	Status    NotificationStatus `db:"status"`
	Data      json.RawMessage    `db:"data"`
	CreatedAt time.Time          `db:"created_at"`
	UpdatedAt time.Time          `db:"updated_at"`
}

// SlackNotification returns the notification data as SlackNotification.
func (n *Notification) SlackNotification() (*SlackNotification, error) {
	var slackNotification SlackNotification
	if err := json.Unmarshal(n.Data, &slackNotification); err != nil {
		return nil, err
	}

	return &slackNotification, nil
}

// SMSNotification returns the notification data as SMSNotification.
func (n *Notification) SMSNotification() (*SMSNotification, error) {
	var smsNotification SMSNotification
	if err := json.Unmarshal(n.Data, &smsNotification); err != nil {
		return nil, err
	}

	return &smsNotification, nil
}

// MailNotification returns the notification data as MailNotification.
func (n *Notification) MailNotification() (*MailNotification, error) {
	var mailNotification MailNotification
	if err := json.Unmarshal(n.Data, &mailNotification); err != nil {
		return nil, err
	}

	return &mailNotification, nil
}

// SlackNotification is a struct that holds the Slack notification data.
type SlackNotification struct {
	Text string `validate:"required" json:"text"`
}

// NewSlackNotification returns a new SlackNotification.
func NewSlackNotification(msg string) (*SlackNotification, error) {
	notification := &SlackNotification{Text: msg}

	if err := validator.New().Struct(notification); err != nil {
		return nil, err
	}

	return notification, nil
}

// SMSNotification is a struct that holds the SMS notification data.
type SMSNotification struct {
	Text   string `validate:"required" json:"text"`
	SendTo string `validate:"required,e164" json:"send_to_number"`
}

// NewSMSNotification returns a new SMSNotification.
func NewSMSNotification(msg string, sendTo string) (*SMSNotification, error) {
	notification := &SMSNotification{Text: msg, SendTo: sendTo}

	if err := validator.New().Struct(notification); err != nil {
		return nil, err
	}

	return notification, nil
}

// MailNotification is a struct that holds the email notification data.
type MailNotification struct {
	Text    string `validate:"required" json:"text"`
	SendTo  string `validate:"required,email" json:"send_to"`
	Subject string `validate:"required" json:"subject"`
}

// NewMailNotification returns a new MailNotification.
func NewMailNotification(msg string, sendTo string, subject string) (*MailNotification, error) {
	notification := &MailNotification{Text: msg, SendTo: sendTo, Subject: subject}

	if err := validator.New().Struct(notification); err != nil {
		return nil, err
	}

	return notification, nil
}
