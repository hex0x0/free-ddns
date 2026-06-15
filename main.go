package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/hex0x0/free-ddns/config"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
		ForceColors:     true,
	})
}

func main() {
	configPath, err := config.DefaultPath()
	if err != nil {
		logrus.Fatalf("resolve default config path: %v", err)
	}

	config.MustLoad(configPath)

	err = checkConfig()
	if err != nil {
		logrus.Fatalf("check config file failed, err: %+v", err)
	}

	logrus.Infof("loaded config from %s", configPath)
	logrus.Infof("domainNames=%v ipAddressVersion=%s dnsProvider=%s\n",
		config.Config.DomainNames, config.Config.IPAddressVersion, config.Config.DNSProvider.Name)

	ddnsExecutor := initDdnsExecutor()
	if ddnsExecutor == nil {
		logrus.Fatalf("ddns executor is nil")
	}

	notifier, err := InitNotifier()
	if err != nil {
		logrus.Fatalf("init notifier failed: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Run once on startup.
	runOnce(ctx, ddnsExecutor, notifier)

	// Start heartbeat scheduler to send messages every Monday at 9:00 AM
	StartHeartbeatScheduler(ctx, notifier)

	// Then run every 15 minutes.
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	logrus.Infof("ddns executor scheduled every 15m")
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("shutting down: %v", ctx.Err())
			return
		case <-ticker.C:
			runOnce(ctx, ddnsExecutor, notifier)
		}
	}
}

func runOnce(ctx context.Context, ddnsExecutor DdnsExecutor, notifier Notifier) {
	start := time.Now()
	logrus.Infof("ddns execution started")

	res := ddnsExecutor.Execute()
	resJson, _ := json.Marshal(res)
	logrus.Infof("ddns execution finished: duration=%s res=%s", time.Since(start), resJson)

	// Notify once per execution if any domain was updated.
	if notifier != nil && (len(res.SuccessfulResult.UpdatedDomains) > 0 || len(res.FailedResult) > 0) {
		logrus.Infof("notification started")

		go func() {
			nerr := NotifyWithRetry(ctx, notifier, res.ToMessage(), 3, 2*time.Minute)
			if nerr != nil {
				logrus.Warnf("failed to send notification after retries: %+v", nerr)
			} else {
				logrus.Info("send notification successfully")
			}
		}()
	}

	select {
	case <-ctx.Done():
		return
	default:
	}
}

func checkConfig() error {
	if len(config.Config.DomainNames) == 0 {
		return errors.New("domain names are empty")
	}

	dnsProviders := map[string]string{
		DnsProviderTencent:    "",
		DnsProviderAliyun:     "",
		DnsProviderCloudflare: "",
	}
	if _, ok := dnsProviders[config.Config.DNSProvider.Name]; !ok {
		return errors.New("dns provider not supported")
	}

	return nil
}

func initDdnsExecutor() DdnsExecutor {
	if config.Config.DNSProvider.Name == DnsProviderTencent {
		executor, err := InitTencentDdnsExecutor()
		if err != nil {
			logrus.Fatalf("init tencent ddns executor failed, err: %v", err)
		}
		return executor
	}

	return nil
}
