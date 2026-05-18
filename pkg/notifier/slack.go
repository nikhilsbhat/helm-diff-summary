package notifier

type SlackNotifier struct {
	WebhookNotifier
}

func (notifier SlackNotifier) Name() string {
	return "slack"
}
