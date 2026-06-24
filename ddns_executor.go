package main

import (
	"github.com/sirupsen/logrus"

	"github.com/hex0x0/free-ddns/config"
)

const (
	DnsProviderTencent    = "tencent"
	DnsProviderAliyun     = "aliyun"
	DnsProviderCloudflare = "cloudflare"
)

// DdnsExecutor represents a DNS provider implementation (e.g. Tencent DNSPod).
type DdnsExecutor interface {
	// Execute returns an ExecutionResult for all domain names.
	Execute() ExecutionResult
}

func InitDdnsExecutor() DdnsExecutor {
	if config.Config.DNSProvider.Name == DnsProviderTencent {
		executor, err := InitTencentDdnsExecutor()
		if err != nil {
			logrus.Fatalf("init tencent ddns executor failed, err: %v", err)
		}
		return executor
	}

	return nil
}
