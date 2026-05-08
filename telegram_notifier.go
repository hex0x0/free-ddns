package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/hex0x0/free-ddns/config"
)

// TelegramNotifier sends message via Telegram Bot API.
// Docs: https://core.telegram.org/bots/api#sendmessage
type TelegramNotifier struct {
	cfg        *config.Config
	httpClient *http.Client
}

func (n *TelegramNotifier) Notify(ctx context.Context, result map[string]*ExecutionResult) error {
	// NOTE: we must not URL-escape the bot token. Telegram tokens contain ':' and
	// other characters that are part of the path segment.
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.cfg.Notifier.Credential.Telegram.BotToken)

	updatedDomains := map[string]map[string]string{}

	for k, v := range result {
		if v.Updated {
			updatedDomains[k] = map[string]string{
				"old_ip": v.OldIP,
				"new_ip": v.NewIP,
			}
		}
	}

	updatedDomainsJson, _ := json.MarshalIndent(updatedDomains, "", "    ")

	messageTemplate := "free-ddns updated dns record.\n\ndns provider: %s\n\nchange of dns records: \n%s"

	message := fmt.Sprintf(
		messageTemplate,
		n.cfg.DNSProvider.Name,
		updatedDomainsJson,
	)

	logrus.Info(message)

	// Keep payload minimal.
	payload := map[string]any{
		"chat_id": n.cfg.Notifier.Credential.Telegram.ChatID,
		"text":    message,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal telegram payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("create telegram request: %w", err)
	}
	req.Header.Set("content-type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send telegram request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("telegram sendMessage failed: http %d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
