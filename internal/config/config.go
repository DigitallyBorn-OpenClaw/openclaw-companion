package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAddr            = "/run/oc-companion/companion.sock"
	defaultGmailWebhookURL = "http://127.0.0.1:18789/hooks/gmail"
	defaultShutdownTimeout = 10 * time.Second
	defaultLogLevel        = "info"
	defaultLogFormat       = "text"
)

type Config struct {
	SocketPath               string
	GmailWebhookURL          string
	GmailWebhookToken        string
	GCPProjectID             string
	GmailPubSubTopicID       string
	GCPCredentialsFile       string
	PubSubSubscriptionPrefix string
	LogLevel                 string
	LogFormat                string
	ShutdownTimeout          time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		SocketPath:               getEnv("OC_COMPANION_SOCKET_PATH", defaultAddr),
		GmailWebhookURL:          getFirstEnvWithFallback(defaultGmailWebhookURL, "OC_OPENCLAW_GMAIL_WEBHOOK_URL", "OC_OPENCLAW_WEBHOOK_BASE_URL"),
		GmailWebhookToken:        strings.TrimSpace(os.Getenv("OC_OPENCLAW_GMAIL_WEBHOOK_TOKEN")),
		GCPProjectID:             strings.TrimSpace(os.Getenv("OC_GCP_PROJECT_ID")),
		GmailPubSubTopicID:       strings.TrimSpace(os.Getenv("OC_GCP_GMAIL_PUBSUB_TOPIC_ID")),
		GCPCredentialsFile:       strings.TrimSpace(os.Getenv("OC_GCP_CREDENTIALS_FILE")),
		PubSubSubscriptionPrefix: sanitizeSubscriptionPrefix(getEnv("OC_GCP_PUBSUB_SUBSCRIPTION_PREFIX", "oc-companion-gmail")),
		LogLevel:                 normalizeLevel(getEnv("OC_COMPANION_LOG_LEVEL", defaultLogLevel)),
		LogFormat:                normalizeFormat(getEnv("OC_COMPANION_LOG_FORMAT", defaultLogFormat)),
		ShutdownTimeout:          getEnvDuration("OC_COMPANION_SHUTDOWN_TIMEOUT", defaultShutdownTimeout),
	}

	if cfg.GmailWebhookToken == "" {
		return Config{}, errors.New("OC_OPENCLAW_GMAIL_WEBHOOK_TOKEN is required")
	}

	if cfg.GCPProjectID == "" {
		return Config{}, errors.New("OC_GCP_PROJECT_ID is required")
	}

	if cfg.GmailPubSubTopicID == "" {
		return Config{}, errors.New("OC_GCP_GMAIL_PUBSUB_TOPIC_ID is required")
	}

	if cfg.SocketPath == "" {
		return Config{}, errors.New("OC_COMPANION_SOCKET_PATH must not be empty")
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func getFirstEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}

	return ""
}

func getFirstEnvWithFallback(fallback string, keys ...string) string {
	if value := getFirstEnv(keys...); value != "" {
		return value
	}

	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return fallback
		}

		return time.Duration(seconds) * time.Second
	}

	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return fallback
	}

	return duration
}

func normalizeLevel(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug", "info", "warn", "error":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return defaultLogLevel
	}
}

func normalizeFormat(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "text", "json":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return defaultLogFormat
	}
}

func (c Config) Summary() string {
	return fmt.Sprintf(
		"socket=%s gmail_webhook=%s gcp_project=%s gmail_topic=%s subscription_prefix=%s log_level=%s log_format=%s shutdown_timeout=%s",
		c.SocketPath,
		c.GmailWebhookURL,
		c.GCPProjectID,
		c.GmailPubSubTopicID,
		c.PubSubSubscriptionPrefix,
		c.LogLevel,
		c.LogFormat,
		c.ShutdownTimeout,
	)
}

func sanitizeSubscriptionPrefix(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-':
			return r
		default:
			return '-'
		}
	}, value)
	value = strings.Trim(value, "-")
	if value == "" {
		return "oc-companion-gmail"
	}
	if value[0] < 'a' || value[0] > 'z' {
		return "oc-" + value
	}

	return value
}
