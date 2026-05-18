package notifier

type TeamsNotifier struct {
	WebhookNotifier
}

func (notifier TeamsNotifier) Name() string {
	return "teams"
}
