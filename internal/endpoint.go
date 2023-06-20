package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// SlackRequestBody is an object used to validate request parameters for Slack notification endpoint.
type SlackRequestBody struct {
	Message string `validate:"required"`
}

// MakeSlackEndpoint creates Slack endpoint for handling request.
func MakeSlackEndpoint(config *Config, notifier SlackNotifier) func(w http.ResponseWriter, r *http.Request) {
	v := validator.New()

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		decoder := json.NewDecoder(r.Body)

		//nolint: errcheck
		defer r.Body.Close()

		var slackRequestBody SlackRequestBody
		if err := decoder.Decode(&slackRequestBody); err != nil {
			http.Error(w, `{"status": "Bad request"}`, http.StatusBadRequest)

			return
		}

		if err := v.Struct(&slackRequestBody); err != nil {
			http.Error(w, fmt.Sprintf(`{"status": "%s"}`, err.Error()), http.StatusBadRequest)

			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), config.Retry.Delay)
		defer cancel()

		err := Retry(notifier.NotifySlack, config.Retry.MaxRetries, config.Retry.Delay)(ctx, slackRequestBody.Message)
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
