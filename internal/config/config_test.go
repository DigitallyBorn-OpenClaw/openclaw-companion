package config

import "testing"

func TestLoad_RequiresGmailWebhookAndPubSubSettings(t *testing.T) {
	t.Setenv("OC_OPENCLAW_GMAIL_WEBHOOK_URL", "http://127.0.0.1:8080/hooks/gmail")
	t.Setenv("OC_OPENCLAW_GMAIL_WEBHOOK_TOKEN", "secret-token")
	t.Setenv("OC_GCP_PROJECT_ID", "my-project")
	t.Setenv("OC_GCP_GMAIL_PUBSUB_TOPIC_ID", "gmail-topic")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected config to load: %v", err)
	}

	if cfg.PubSubSubscriptionPrefix != "oc-companion-gmail" {
		t.Fatalf("expected default subscription prefix, got %q", cfg.PubSubSubscriptionPrefix)
	}
}

func TestLoad_AllowsLegacyWebhookEnvName(t *testing.T) {
	t.Setenv("OC_OPENCLAW_WEBHOOK_BASE_URL", "http://127.0.0.1:8080/hooks/gmail")
	t.Setenv("OC_OPENCLAW_GMAIL_WEBHOOK_TOKEN", "secret-token")
	t.Setenv("OC_GCP_PROJECT_ID", "my-project")
	t.Setenv("OC_GCP_GMAIL_PUBSUB_TOPIC_ID", "gmail-topic")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected legacy webhook env to load: %v", err)
	}

	if cfg.GmailWebhookURL != "http://127.0.0.1:8080/hooks/gmail" {
		t.Fatalf("expected legacy webhook url to populate gmail webhook, got %q", cfg.GmailWebhookURL)
	}
}

func TestSanitizeSubscriptionPrefix(t *testing.T) {
	got := sanitizeSubscriptionPrefix("123 bad.prefix")
	if got != "oc-123-bad-prefix" {
		t.Fatalf("expected sanitized prefix, got %q", got)
	}
}
