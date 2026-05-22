package main

import "time"

// DomainUpdateInfo describes a domain that had its IP updated.
type DomainUpdateInfo struct {
	DomainName string `json:"domain_name"`
	OldIp      string `json:"old_ip"`
	NewIp      string `json:"new_ip"`
}

// SuccessfulResult contains domains that were successfully processed.
type SuccessfulResult struct {
	UpdatedDomains    []DomainUpdateInfo `json:"updated_domains"`
	UnmodifiedDomains []string           `json:"unmodified_domains"`
}

// FailedInfo describes a domain that failed to update.
type FailedInfo struct {
	DomainNames string `json:"domain_names"`
	Reason      string `json:"reason"`
}

// ExecutionResult describes what happened during a DNS provider execution.
type ExecutionResult struct {
	ExecutedAt       time.Time        `json:"-"`
	SuccessfulResult SuccessfulResult `json:"successful_result"`
	FailedResult     []FailedInfo     `json:"failed_result"`
}

// DdnsExecutor represents a DNS provider implementation (e.g. Tencent DNSPod).
type DdnsExecutor interface {
	// Execute returns an ExecutionResult for all domain names.
	Execute() ExecutionResult
}
