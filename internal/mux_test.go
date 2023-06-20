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

const _dotEnvFileName = ".env.dist"

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
			expectedResponseMessage: `{"status": "Key: 'SlackRequestBody.Message' Error:Field validation for 'Message' failed on the 'required' tag"}`,
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
