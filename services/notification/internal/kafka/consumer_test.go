package kafka

import "testing"

func TestBuildMessage(t *testing.T) {
	cases := []struct {
		eventType string
		expected  string
	}{
		{"subscription.created", "Your subscription is now active."},
		{"subscription.canceled", "Your subscription has been canceled."},
		{"payment.succeeded", "Your payment was processed successfully."},
		{"payment.failed", "Your payment could not be processed. We'll retry shortly."},
		{"unknown.event", "You have a new notification."},
	}

	for _, c := range cases {
		msg := buildMessage(c.eventType)
		if msg != c.expected {
			t.Fatalf("buildMessage(%q) = %q, want %q", c.eventType, msg, c.expected)
		}
	}
}
