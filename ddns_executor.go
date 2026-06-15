package main

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
