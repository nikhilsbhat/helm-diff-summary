package notifier

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("SLACK_WEBHOOK_URL", "https://slack.example")
	t.Setenv("TEAMS_WEBHOOK_URL", "https://teams.example")
	t.Setenv("GCHAT_WEBHOOK_URL", "https://gchat.example")
	t.Setenv("WEBHOOK_URL", "https://webhook.example")

	config := LoadConfig()

	if config.SlackWebhook == "" || config.TeamsWebhook == "" || config.GChatWebhook == "" || config.WebhookURL == "" {
		t.Fatalf("expected all webhook values, got %#v", config)
	}
}

func TestBuildNotifiers(t *testing.T) {
	config := Config{
		SlackWebhook: "https://slack.example",
		TeamsWebhook: "https://teams.example",
		GChatWebhook: "https://gchat.example",
		WebhookURL:   "https://webhook.example",
	}

	notifiers, err := buildNotifiers([]string{"slack", "teams", "gchat", "webhook"}, config)
	if err != nil {
		t.Fatalf("buildNotifiers returned error: %v", err)
	}

	names := make([]string, 0, len(notifiers))
	for _, notifier := range notifiers {
		names = append(names, notifier.Name())
	}

	if strings.Join(names, ",") != "slack,teams,gchat,webhook" {
		t.Fatalf("unexpected notifier names: %v", names)
	}
}

func TestBuildNotifiersValidation(t *testing.T) {
	tests := []struct {
		name    string
		targets []string
		config  Config
	}{
		{name: "missing slack", targets: []string{"slack"}},
		{name: "missing teams", targets: []string{"teams"}},
		{name: "missing gchat", targets: []string{"gchat"}},
		{name: "missing webhook", targets: []string{"webhook"}},
		{name: "unsupported", targets: []string{"email"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := buildNotifiers(tt.targets, tt.config); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestWebhookNotifierNotify(t *testing.T) {
	var payload webhookPayload

	restorePostWebhook(t)
	postWebhook = func(url string, body []byte) (int, string, io.ReadCloser, error) {
		if url != "https://webhook.example" {
			t.Fatalf("unexpected URL: %s", url)
		}

		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}

		return http.StatusOK, "200 OK", io.NopCloser(bytes.NewReader(nil)), nil
	}

	notifier := WebhookNotifier{URL: "https://webhook.example"}
	if err := notifier.Notify("hello"); err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}

	if payload.Text != "```hello```" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestWebhookNotifierNotifyReturnsHTTPError(t *testing.T) {
	restorePostWebhook(t)
	postWebhook = func(_ string, _ []byte) (int, string, io.ReadCloser, error) {
		return http.StatusBadGateway, "502 Bad Gateway", io.NopCloser(bytes.NewReader(nil)), nil
	}

	notifier := WebhookNotifier{URL: "https://webhook.example"}
	if err := notifier.Notify("hello"); err == nil {
		t.Fatal("expected HTTP error")
	}
}

func TestWebhookNotifierNotifyLogsBodyCloseError(t *testing.T) {
	restorePostWebhook(t)
	postWebhook = func(_ string, _ []byte) (int, string, io.ReadCloser, error) {
		return http.StatusOK, "200 OK", closeErrorReader{}, nil
	}

	notifier := WebhookNotifier{
		URL:    "https://webhook.example",
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	if err := notifier.Notify("hello"); err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}
}

func TestWebhookNotifierNotifyReturnsRequestError(t *testing.T) {
	notifier := WebhookNotifier{URL: "://bad-url"}
	if err := notifier.Notify("hello"); err == nil {
		t.Fatal("expected request error")
	}
}

func TestNotifyWithNoTargetsDoesNothing(t *testing.T) {
	if err := Notify("hello", nil); err != nil {
		t.Fatalf("Notify returned error for no targets: %v", err)
	}
}

func TestNotifyReturnsConfigError(t *testing.T) {
	for _, env := range []string{"SLACK_WEBHOOK_URL", "TEAMS_WEBHOOK_URL", "GCHAT_WEBHOOK_URL", "WEBHOOK_URL"} {
		t.Setenv(env, "")
	}

	if err := Notify("hello", []string{"slack"}); err == nil {
		t.Fatal("expected missing config error")
	}

	_ = os.Unsetenv("SLACK_WEBHOOK_URL")
}

func restorePostWebhook(t *testing.T) {
	t.Helper()

	original := postWebhook
	t.Cleanup(func() {
		postWebhook = original
	})
}

type closeErrorReader struct{}

func (closeErrorReader) Read(_ []byte) (int, error) {
	return 0, io.EOF
}

func (closeErrorReader) Close() error {
	return errClose
}

var errClose = &closeError{"close failed"}

type closeError struct {
	message string
}

func (err *closeError) Error() string {
	return err.message
}
