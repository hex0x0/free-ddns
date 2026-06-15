package main

import (
	"encoding/json"
	"fmt"
	"time"
)

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
	DnsProvider      string           `json:"dns_provider"`
	SuccessfulResult SuccessfulResult `json:"successful_result"`
	FailedResult     []FailedInfo     `json:"failed_result"`
}

func (e ExecutionResult) ToMessage() string {
	updatedDomainsJson, _ := json.MarshalIndent(e.SuccessfulResult.UpdatedDomains, "", "    ")
	failedDomainsJson, _ := json.MarshalIndent(e.FailedResult, "", "    ")

	messageTemplate := "free-ddns\n\nDNS records updated\n\nDNS provider: %s\n\nExecute Time: %v\n\nSuccessful Result:\n%s\n\nFailed Result:\n%s"

	return fmt.Sprintf(
		messageTemplate,
		e.DnsProvider,
		e.ExecutedAt.Format(time.RFC3339),
		string(updatedDomainsJson),
		string(failedDomainsJson),
	)
}
