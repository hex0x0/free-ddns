package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hex0x0/free-ddns/config"
)

type schedule struct {
	Weekday time.Weekday
	Hour    int
	Minute  int
}

func nextRun(now time.Time, s schedule) time.Time {
	target := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		s.Hour,
		s.Minute,
		0,
		0,
		now.Location(),
	)

	days := (int(s.Weekday) - int(now.Weekday()) + 7) % 7

	if days == 0 && !now.Before(target) {
		days = 7
	}

	return target.AddDate(0, 0, days)
}

// StartHeartbeatScheduler starts a goroutine that sends heartbeat messages every Monday at 9:00 AM.
func StartHeartbeatScheduler(ctx context.Context, notifier Notifier) {
	if notifier == nil {
		logrus.Warn("heartbeat scheduler disabled: notifier is nil")
		return
	}

	go func() {
		logrus.Info("send heartbeat message after reboot")
		sendHeartbeat(ctx, notifier)

		s := schedule{
			Weekday: time.Monday,
			Hour:    9,
			Minute:  0,
		}

		for {
			now := time.Now()
			next := nextRun(now, s)
			timer := time.NewTimer(time.Until(next))

			logrus.Infof("next heartbeat scheduled at %s", next.Format(time.RFC3339))

			select {
			case <-ctx.Done():
				logrus.Info("heartbeat scheduler stopped")
				return
			case <-timer.C:
				sendHeartbeat(ctx, notifier)
			}
		}
	}()
}

// sendHeartbeat sends a heartbeat notification with current domain-to-IP mappings.
func sendHeartbeat(ctx context.Context, notifier Notifier) {
	logrus.Info("sending heartbeat message")

	domainIPMapping := buildDomainIPMapping()
	ipAddr := make([]map[string]string, 0)
	for k, v := range domainIPMapping {
		ipAddr = append(ipAddr, map[string]string{
			"domain name": k,
			"ip address":  v,
		})
	}
	ipAddrJsonStr := "[]"
	ipAddrJsonBytes, err := json.MarshalIndent(ipAddr, "", "    ")
	if err != nil {
		logrus.Warnf("marshal ipAddr failed, err=%+v", err)
	} else {
		ipAddrJsonStr = string(ipAddrJsonBytes)
	}

	heartbeatMsg := fmt.Sprintf("free-ddns\n\nHeartbeat Message\n\nDNS Provider: %s\n\nSend time: %s\n\nCurrent IP:\n%v",
		config.Config.DNSProvider.Name, time.Now().Format(time.RFC3339), ipAddrJsonStr)

	notifyErr := NotifyWithRetry(ctx, notifier, heartbeatMsg, 3, 2*time.Minute)
	if notifyErr != nil {
		logrus.Warnf("failed to send heartbeat message after retries: %+v", notifyErr)
	} else {
		logrus.Info("heartbeat message sent successfully")
	}
}

func buildDomainIPMapping() map[string]string {
	mapping := make(map[string]string)

	for _, n := range config.Config.DomainNames {
		ipAddr, err := NewIpGetter().GetPublicIP()
		if err != nil {
			logrus.Warnf("get public ip failed, domain name=%s err=%+v", n, err)
		} else {
			mapping[n] = ipAddr
		}
	}

	return mapping
}
