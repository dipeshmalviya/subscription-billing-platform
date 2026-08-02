package notifier

import (
	"github.com/rs/zerolog"
)

// EmailNotifier is a mock — real implementation would call SendGrid/SES/etc.
// Logging is enough to prove the event pipeline works end-to-end.
type EmailNotifier struct {
	logger zerolog.Logger
}

func NewEmailNotifier(logger zerolog.Logger) *EmailNotifier {
	return &EmailNotifier{logger: logger}
}

func (n *EmailNotifier) Send(customerID, eventType, message string) error {
	n.logger.Info().
		Str("customer_id", customerID).
		Str("event_type", eventType).
		Str("message", message).
		Msg("notification sent (mock)")
	return nil
}