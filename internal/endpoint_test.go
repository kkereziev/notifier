package internal_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/kkereziev/notifier/internal"
	"github.com/kkereziev/notifier/internal/mocks"
)

const _dotEnvFileName = ".env.test"

func TestSlackNotificationPositiveCases(t *testing.T) {
	t.Parallel()

	type test struct {
		name                        string
		requestBody                 *internal.SlackRequestBody
		expectedStatusCode          int
		expectedNotificationMessage string
	}

	tests := []test{
		{
			name:                        "Slack notifier service should notify successfully",
			requestBody:                 &internal.SlackRequestBody{Message: "Hello"},
			expectedStatusCode:          http.StatusOK,
			expectedNotificationMessage: "Hello",
		},
	}

	if err := loadEnv(); err != nil {
		t.Fatal(err)
	}

	config, err := internal.NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	var notificationMessage string

	notifierMock := &mocks.NotifierMock{
		NotifySlackFunc: func(contextMoqParam context.Context, ifaceVal any) error {
			msg := ifaceVal.(string)
			notificationMessage = msg

			return nil
		},
	}

	mux := internal.NewMux(config, notifierMock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Fatalf("failed to marshal Slack message: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/slack", bytes.NewBuffer(payload))
			res := httptest.NewRecorder()

			mux.ServeHTTP(res, req)

			if res.Result().StatusCode != tc.expectedStatusCode {
				t.Fatalf("Different status codes, expected: %v, got: %v", tc.expectedStatusCode, res.Result().StatusCode)
			}

			if tc.expectedNotificationMessage != notificationMessage {
				t.Fatalf("Different messages, expected: %s, got: %s", tc.expectedNotificationMessage, notificationMessage)
			}

			callsToSend := len(notifierMock.NotifySlackCalls())

			if callsToSend == 0 {
				t.Fatalf("Expected NotifySlack to be called at least once.")
			}
		})
	}
}

func TestSlackNotificationNegativeCases(t *testing.T) {
	t.Parallel()

	type test struct {
		name                    string
		expectedStatusCode      int
		requestBody             *internal.SlackRequestBody
		expectedResponseMessage string
	}

	tests := []test{
		{
			name:               "request should fail on validation because body is empty",
			requestBody:        &internal.SlackRequestBody{},
			expectedStatusCode: http.StatusBadRequest,
			//nolint: lll
			expectedResponseMessage: `{"error": "Key: 'SlackRequestBody.Message' Error:Field validation for 'Message' failed on the 'required' tag"}`,
		},
	}

	if err := loadEnv(); err != nil {
		t.Fatal(err)
	}

	config, err := internal.NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	notifierMock := &mocks.NotifierMock{}

	mux := internal.NewMux(config, notifierMock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Fatalf("failed to marshal Slack message: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/slack", bytes.NewBuffer(payload))
			res := httptest.NewRecorder()

			mux.ServeHTTP(res, req)

			if res.Result().StatusCode != tc.expectedStatusCode {
				t.Fatalf("Different status codes, expected: %v, got: %v", tc.expectedStatusCode, res.Result().StatusCode)
			}

			body, err := io.ReadAll(res.Result().Body)
			if err != nil {
				t.Fatalf("Error reading response body: %v", err)
			}

			if strings.TrimSpace(string(body)) != tc.expectedResponseMessage {
				t.Fatalf("Different response messages, expected: %s, got: %s", tc.expectedResponseMessage, body)
			}
		})
	}
}

func TestSMSNotificationPositiveCases(t *testing.T) {
	t.Parallel()

	type test struct {
		name                             string
		requestBody                      *internal.SMSRequestBody
		expectedStatusCode               int
		expectedNotificationMessage      string
		expectedNotificationSendToNumber string
	}

	tests := []test{
		{
			name:                             "SMS notifier service should notify successfully",
			requestBody:                      &internal.SMSRequestBody{Message: "Hello", SendToNumber: "+35988357997"},
			expectedStatusCode:               http.StatusOK,
			expectedNotificationMessage:      "Hello",
			expectedNotificationSendToNumber: "+35988357997",
		},
	}

	if err := loadEnv(); err != nil {
		t.Fatal(err)
	}

	config, err := internal.NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	var (
		notificationMessage      string
		notificationSendToNumber string
	)

	notifierMock := &mocks.NotifierMock{
		NotifySMSFunc: func(contextMoqParam context.Context, ifaceVal any) error {
			msg := ifaceVal.(*internal.SMSRequestBody)
			notificationMessage = msg.Message
			notificationSendToNumber = msg.SendToNumber

			return nil
		},
	}

	mux := internal.NewMux(config, notifierMock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Fatalf("failed to marshal SMS message: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/sms", bytes.NewBuffer(payload))
			res := httptest.NewRecorder()

			mux.ServeHTTP(res, req)

			if res.Result().StatusCode != tc.expectedStatusCode {
				t.Fatalf("Different status codes, expected: %v, got: %v", tc.expectedStatusCode, res.Result().StatusCode)
			}

			if tc.expectedNotificationMessage != notificationMessage {
				t.Fatalf("Different messages, expected: %s, got: %s", tc.expectedNotificationMessage, notificationMessage)
			}

			if tc.expectedNotificationSendToNumber != notificationSendToNumber {
				t.Fatalf(
					"Different send to numbers, expected: %s, got: %s", tc.expectedNotificationSendToNumber, notificationSendToNumber,
				)
			}

			callsToSend := len(notifierMock.NotifySMSCalls())

			if callsToSend == 0 {
				t.Fatalf("Expected NotifySlack to be called at least once.")
			}
		})
	}
}

func TestSMSNotificationNegativeCases(t *testing.T) {
	t.Parallel()

	type test struct {
		name                    string
		expectedStatusCode      int
		requestBody             *internal.SMSRequestBody
		expectedResponseMessage string
	}

	tests := []test{
		{
			name:               "request should fail on validation because message is empty",
			requestBody:        &internal.SMSRequestBody{SendToNumber: "+35988357997"},
			expectedStatusCode: http.StatusBadRequest,
			//nolint: lll
			expectedResponseMessage: `{"error": "Key: 'SMSRequestBody.Message' Error:Field validation for 'Message' failed on the 'required' tag"}`,
		},
		{
			name:               "request should fail on validation because of invalid number",
			requestBody:        &internal.SMSRequestBody{Message: "Hello", SendToNumber: "35"},
			expectedStatusCode: http.StatusBadRequest,
			//nolint: lll
			expectedResponseMessage: `{"error": "Key: 'SMSRequestBody.SendToNumber' Error:Field validation for 'SendToNumber' failed on the 'e164' tag"}`,
		},
		{
			name:               "request should fail on validation because of missing number",
			requestBody:        &internal.SMSRequestBody{Message: "Hello"},
			expectedStatusCode: http.StatusBadRequest,
			//nolint: lll
			expectedResponseMessage: `{"error": "Key: 'SMSRequestBody.SendToNumber' Error:Field validation for 'SendToNumber' failed on the 'required' tag"}`,
		},
	}

	if err := loadEnv(); err != nil {
		t.Fatal(err)
	}

	config, err := internal.NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	notifierMock := &mocks.NotifierMock{}

	mux := internal.NewMux(config, notifierMock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Fatalf("failed to marshal Slack message: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/sms", bytes.NewBuffer(payload))
			res := httptest.NewRecorder()

			mux.ServeHTTP(res, req)

			if res.Result().StatusCode != tc.expectedStatusCode {
				t.Fatalf("Different status codes, expected: %v, got: %v", tc.expectedStatusCode, res.Result().StatusCode)
			}

			body, err := io.ReadAll(res.Result().Body)
			if err != nil {
				t.Fatalf("Error reading response body: %v", err)
			}

			if strings.TrimSpace(string(body)) != tc.expectedResponseMessage {
				t.Fatalf("Different response messages, expected: %s, got: %s", tc.expectedResponseMessage, body)
			}
		})
	}
}

func TestMailNotificationPositiveCases(t *testing.T) {
	t.Parallel()

	type test struct {
		name                         string
		requestBody                  *internal.MailRequestBody
		expectedStatusCode           int
		expectedNotificationMessage  string
		expectedNotificationReceiver string
		expectedNotificationSubject  string
	}

	tests := []test{
		{
			name: "Mail notifier service should notify successfully",
			requestBody: &internal.MailRequestBody{
				Message: "Hello",
				SendTo:  "example@gmail.com",
				Subject: "Test",
			},
			expectedStatusCode:           http.StatusOK,
			expectedNotificationMessage:  "Hello",
			expectedNotificationReceiver: "example@gmail.com",
			expectedNotificationSubject:  "Test",
		},
	}

	if err := loadEnv(); err != nil {
		t.Fatal(err)
	}

	config, err := internal.NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	var (
		notificationMessage  string
		notificationReceiver string
		notificationSubject  string
	)

	notifierMock := &mocks.NotifierMock{
		NotifyMailFunc: func(contextMoqParam context.Context, ifaceVal any) error {
			msg := ifaceVal.(*internal.MailRequestBody)
			notificationMessage = msg.Message
			notificationReceiver = msg.SendTo
			notificationSubject = msg.Subject

			return nil
		},
	}

	mux := internal.NewMux(config, notifierMock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Fatalf("failed to marshal SMS message: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/mail", bytes.NewBuffer(payload))
			res := httptest.NewRecorder()

			mux.ServeHTTP(res, req)

			if res.Result().StatusCode != tc.expectedStatusCode {
				t.Fatalf("Different status codes, expected: %v, got: %v", tc.expectedStatusCode, res.Result().StatusCode)
			}

			if tc.expectedNotificationMessage != notificationMessage {
				t.Fatalf("Different messages, expected: %s, got: %s", tc.expectedNotificationMessage, notificationMessage)
			}

			if tc.expectedNotificationReceiver != notificationReceiver {
				t.Fatalf(
					"Different mail receiver, expected: %s, got: %s", tc.expectedNotificationReceiver, notificationReceiver,
				)
			}

			if tc.expectedNotificationSubject != notificationSubject {
				t.Fatalf(
					"Different mail subjects, expected: %s, got: %s", tc.expectedNotificationSubject, notificationSubject,
				)
			}

			callsToSend := len(notifierMock.NotifyMailCalls())

			if callsToSend == 0 {
				t.Fatalf("Expected NotifySlack to be called at least once.")
			}
		})
	}
}

func TestMailNotificationNegativeCases(t *testing.T) {
	t.Parallel()

	type test struct {
		name                    string
		expectedStatusCode      int
		requestBody             *internal.MailRequestBody
		expectedResponseMessage string
	}

	tests := []test{
		{
			name:               "request should fail on validation because message is empty",
			requestBody:        &internal.MailRequestBody{SendTo: "test@gmail.com", Subject: "test"},
			expectedStatusCode: http.StatusBadRequest,
			//nolint: lll
			expectedResponseMessage: `{"error": "Key: 'MailRequestBody.Message' Error:Field validation for 'Message' failed on the 'required' tag"}`,
		},
		{
			name:               "request should fail on validation because of invalid email for SendTo",
			requestBody:        &internal.MailRequestBody{Message: "Hello", SendTo: "35", Subject: "test"},
			expectedStatusCode: http.StatusBadRequest,
			//nolint: lll
			expectedResponseMessage: `{"error": "Key: 'MailRequestBody.SendTo' Error:Field validation for 'SendTo' failed on the 'email' tag"}`,
		},
		{
			name:               "request should fail on validation because of missing number",
			requestBody:        &internal.MailRequestBody{Message: "Hello", Subject: "test"},
			expectedStatusCode: http.StatusBadRequest,
			//nolint: lll
			expectedResponseMessage: `{"error": "Key: 'MailRequestBody.SendTo' Error:Field validation for 'SendTo' failed on the 'required' tag"}`,
		},
		{
			name:               "request should fail on validation because of missing subject",
			requestBody:        &internal.MailRequestBody{Message: "Hello", SendTo: "test@gmail.com"},
			expectedStatusCode: http.StatusBadRequest,
			//nolint: lll
			expectedResponseMessage: `{"error": "Key: 'MailRequestBody.Subject' Error:Field validation for 'Subject' failed on the 'required' tag"}`,
		},
	}

	if err := loadEnv(); err != nil {
		t.Fatal(err)
	}

	config, err := internal.NewConfig()
	if err != nil {
		t.Fatal(err)
	}

	notifierMock := &mocks.NotifierMock{}

	mux := internal.NewMux(config, notifierMock)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(tc.requestBody)
			if err != nil {
				t.Fatalf("failed to marshal Slack message: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/mail", bytes.NewBuffer(payload))
			res := httptest.NewRecorder()

			mux.ServeHTTP(res, req)

			if res.Result().StatusCode != tc.expectedStatusCode {
				t.Fatalf("Different status codes, expected: %v, got: %v", tc.expectedStatusCode, res.Result().StatusCode)
			}

			body, err := io.ReadAll(res.Result().Body)
			if err != nil {
				t.Fatalf("Error reading response body: %v", err)
			}

			if strings.TrimSpace(string(body)) != tc.expectedResponseMessage {
				t.Fatalf("Different response messages, expected: %s, got: %s", tc.expectedResponseMessage, body)
			}
		})
	}
}

func loadEnv() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get current working directory: %v", err)
	}

	testEnvFilePath := fmt.Sprintf("%s/../%s", dir, _dotEnvFileName)

	err = godotenv.Load(testEnvFilePath)
	if err != nil {
		return fmt.Errorf("cannot load test env file %s: %v", testEnvFilePath, err)
	}

	return nil
}
