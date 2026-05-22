package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hex0x0/free-ddns/config"
)

// TelegramNotifier sends message via Telegram Bot API.
// Docs: https://core.telegram.org/bots/api#sendmessage
type TelegramNotifier struct {
	cfg        *config.Config
	httpClient *http.Client
}

func (n *TelegramNotifier) Notify(ctx context.Context, result ExecutionResult) error {
	// NOTE: we must not URL-escape the bot token. Telegram tokens contain ':' and
	// other characters that are part of the path segment.
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.cfg.Notifier.Credential.Telegram.BotToken)

	updatedDomainsJson, _ := json.MarshalIndent(result.SuccessfulResult.UpdatedDomains, "", "    ")
	failedDomainsJson, _ := json.MarshalIndent(result.FailedResult, "", "    ")

	messageTemplate := "free-ddns updated dns record.\n\ndns provider: %s\n\nexecuted_at: %v\n\nsuccessful_result:\n%s\n\nfailed_result:\n%s"

	message := fmt.Sprintf(
		messageTemplate,
		n.cfg.DNSProvider.Name,
		result.ExecutedAt.Format(time.RFC3339),
		string(updatedDomainsJson),
		string(failedDomainsJson),
	)

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
