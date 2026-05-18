package notifier

type GChatNotifier struct {
	WebhookNotifier
}

func (notifier GChatNotifier) Name() string {
	return "gchat"
}
