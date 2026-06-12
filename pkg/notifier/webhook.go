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

var postWebhook = func(url string, body []byte) (int, string, io.ReadCloser, error) {
	client := resty.New()

	response, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(bytes.NewBuffer(body)).
		Post(url)
	if err != nil {
		return 0, "", nil, err
	}

	return response.StatusCode(), response.Status(), response.Body, nil
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

	statusCode, status, bodyReader, err := postWebhook(notifier.URL, body)
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		if err = Body.Close(); err != nil {
			notifier.logger.Error(err.Error())
		}
	}(bodyReader)

	if statusCode >= http.StatusBadRequest {
		return &errors.DiffSummaryError{Message: fmt.Sprintf(
			"webhook notification failed with status: %s",
			status,
		)}
	}

	return nil
}
