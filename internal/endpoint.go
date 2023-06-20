package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

// MakeSlackEndpoint creates endpoint for sending Slack notifications.
func MakeSlackEndpoint(config *Config, notifier SlackNotifier) func(w http.ResponseWriter, r *http.Request) {
	v := validator.New()

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		decoder := json.NewDecoder(r.Body)

		//nolint: errcheck
		defer r.Body.Close()

		var slackRequest SlackRequestBody
		if err := decoder.Decode(&slackRequest); err != nil {
			http.Error(w, `{"status": "Bad request"}`, http.StatusBadRequest)

			return
		}

		if err := v.Struct(&slackRequest); err != nil {
			http.Error(w, fmt.Sprintf(`{"status": "%s"}`, err.Error()), http.StatusBadRequest)

			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), config.Retry.Delay)
		defer cancel()

		err := Retry(notifier.NotifySlack, config.Retry.MaxRetries, config.Retry.Delay)(ctx, slackRequest.Message)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)

			return
		}

		if _, err := w.Write([]byte(`{"status": "Notification send."}`)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	}
}

// MakeSMSEndpoint creates endpoint for sending SMS notifications.
func MakeSMSEndpoint(config *Config, notifier SMSNotifier) func(w http.ResponseWriter, r *http.Request) {
	v := validator.New()

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		decoder := json.NewDecoder(r.Body)

		//nolint: errcheck
		defer r.Body.Close()

		var smsRequest SMSRequestBody
		if err := decoder.Decode(&smsRequest); err != nil {
			http.Error(w, `{"error": "Bad request"}`, http.StatusBadRequest)

			return
		}

		if err := v.Struct(&smsRequest); err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusBadRequest)

			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Retry.MaxRetries)*config.Retry.Delay+5)
		defer cancel()

		err := Retry(notifier.NotifySMS, config.Retry.MaxRetries, config.Retry.Delay)(ctx, &smsRequest)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)

			return
		}

		if _, err := w.Write([]byte(`{"status": "Notification send."}`)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	}
}
