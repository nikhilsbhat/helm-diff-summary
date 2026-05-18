package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/nikhilsbhat/helm-diff-summary/pkg/errors"
	"resty.dev/v3"
)

// WebhookNotifier holds the information required by the notifiers.
type WebhookNotifier struct {
	URL    string
	logger *slog.Logger
}

type webhookPayload struct {
	Text string `json:"text"`
}

// Name returns the name of the selected webhook.
func (notifier WebhookNotifier) Name() string {
	return "webhook"
}

// Notify sends the message across appropriate notifier opted.
func (notifier WebhookNotifier) Notify(message string) error {
	message = fmt.Sprintf("```%s```", message)

	payload := webhookPayload{
		Text: message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := resty.New()

	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(bytes.NewBuffer(body)).
		Post(notifier.URL)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			notifier.logger.Error(err.Error())
		}
	}(response.Body)

	if response.StatusCode() >= http.StatusBadRequest {
		return &errors.DiffSummaryError{Message: fmt.Sprintf(
			"webhook notification failed with status: %s",
			response.Status(),
		)}
	}

	return nil
}
