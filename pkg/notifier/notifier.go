package notifier

import "os"

// Config holds the config required by the notifiers.
type Config struct {
	SlackWebhook string `yaml:"slack_webhook,omitempty"  json:"slack_webhook,omitempty"`
	WebhookURL   string `yaml:"webhook_url,omitempty"    json:"webhook_url,omitempty"`
	TeamsWebhook string `yaml:"teams_webhook,omitempty"  json:"teams_webhook,omitempty"`
	GChatWebhook string `yaml:"g_chat_webhook,omitempty" json:"g_chat_webhook,omitempty"`
}

// Notifier implements methods that sends the summary as a message across the selected notifier.
type Notifier interface {
	Name() string

	Notify(message string) error
}

// LoadConfig loads the webhook config required by the notifiers.
func LoadConfig() Config {
	return Config{
		SlackWebhook: os.Getenv(
			"SLACK_WEBHOOK_URL",
		),

		TeamsWebhook: os.Getenv(
			"TEAMS_WEBHOOK_URL",
		),

		GChatWebhook: os.Getenv(
			"GCHAT_WEBHOOK_URL",
		),

		WebhookURL: os.Getenv(
			"WEBHOOK_URL",
		),
	}
}
