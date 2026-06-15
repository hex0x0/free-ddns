package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hex0x0/free-ddns/config"
)

// Notifier sends a notification message.
// Implementations should be best-effort and return an error on failure.
type Notifier interface {
	Notify(ctx context.Context, message string) error
}

func httpClientWithProxyFromEnv() (*http.Client, error) {
	res := &http.Client{
		Timeout: 10 * time.Second,
	}

	proxy := strings.TrimSpace(os.Getenv("http_proxy"))
	if proxy == "" {
		proxy = strings.TrimSpace(os.Getenv("HTTP_PROXY"))
	}
	if proxy == "" {
		return res, nil
	}

	// url.Parse requires a scheme. Many proxy strings are provided as host:port.
	if !strings.Contains(proxy, "://") {
		proxy = "http://" + proxy
	}
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		return nil, fmt.Errorf("parse http_proxy: %w", err)
	}

	var t *http.Transport
	switch base := res.Transport.(type) {
	case nil:
		t = http.DefaultTransport.(*http.Transport).Clone()
	case *http.Transport:
		t = base.Clone()
	default:
		return nil, fmt.Errorf("cannot set proxy on transport type %T", res.Transport)
	}

	t.Proxy = http.ProxyURL(proxyURL)
	res.Transport = t
	return res, nil
}

func InitNotifier() (Notifier, error) {
	if config.Config.Notifier == nil {
		return nil, nil
	}

	httpClient, err := httpClientWithProxyFromEnv()
	if err != nil {
		return nil, fmt.Errorf("init notifier's http client failed, err: %+v", err)
	}

	switch strings.ToLower(config.Config.Notifier.Name) {
	case "telegram":
		cred := config.Config.Notifier.Credential.Telegram
		if strings.TrimSpace(cred.ChatID) == "" || strings.TrimSpace(cred.BotToken) == "" {
			return nil, fmt.Errorf("telegram notification configured but chatId/botToken is empty")
		}
		return &TelegramNotifier{
			credential: config.Config.Notifier.Credential.Telegram,
			httpClient: httpClient,
		}, nil
	case "":
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported notification channel: %s", config.Config.Notifier.Name)
	}
}

// NotifyWithRetry sends notification with retry on failure.
func NotifyWithRetry(ctx context.Context,
	notifier Notifier,
	message string,
	maxAttempts int,
	delay time.Duration) error {

	if notifier == nil {
		return nil
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if attempt > 1 {
			t := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				t.Stop()
				return ctx.Err()
			case <-t.C:
				logrus.Warnf("notification retrying: attempt=%d/%d delay=%s", attempt, maxAttempts, delay)
			}
		}

		if err := notifier.Notify(ctx, message); err != nil {
			lastErr = err
			logrus.Warnf("send notification failed: attempt=%d/%d msg=%s err=%+v", attempt, maxAttempts, message, err)
			continue
		}
		return nil
	}
	return lastErr
}
