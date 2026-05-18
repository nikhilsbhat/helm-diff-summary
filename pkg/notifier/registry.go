package notifier

import (
	"fmt"

	"github.com/nikhilsbhat/helm-diff-summary/pkg/errors"
)

// Notify loads the notifier configs and sends the messages to the targeted notifiers.
func Notify(message string, targets []string) error {
	config := LoadConfig()

	notifiers, err := buildNotifiers(targets, config)
	if err != nil {
		return err
	}

	for _, notifier := range notifiers {
		if err = notifier.Notify(message); err != nil {
			return err
		}
	}

	return nil
}

func buildNotifiers(targets []string, config Config) ([]Notifier, error) {
	notifiers := make([]Notifier, 0)

	for _, target := range targets {
		switch target {
		case "slack":
			if config.SlackWebhook == "" {
				return nil, &errors.DiffSummaryError{Message: "SLACK_WEBHOOK_URL not set"}
			}

			notifiers = append(notifiers, SlackNotifier{WebhookNotifier{URL: config.SlackWebhook}})
		case "teams":
			if config.TeamsWebhook == "" {
				return nil, &errors.DiffSummaryError{Message: "TEAMS_WEBHOOK_URL not set"}
			}

			notifiers = append(notifiers, TeamsNotifier{WebhookNotifier{URL: config.TeamsWebhook}})
		case "gchat":
			if config.GChatWebhook == "" {
				return nil, &errors.DiffSummaryError{Message: "GCHAT_WEBHOOK_URL not set"}
			}

			notifiers = append(notifiers, GChatNotifier{WebhookNotifier{URL: config.GChatWebhook}})
		case "webhook":
			if config.WebhookURL == "" {
				return nil, &errors.DiffSummaryError{Message: "WEBHOOK_URL not set"}
			}

			notifiers = append(notifiers, WebhookNotifier{URL: config.WebhookURL})
		default:
			return nil, &errors.DiffSummaryError{Message: fmt.Sprintf("unsupported notifier: %s", target)}
		}
	}

	return notifiers, nil
}
