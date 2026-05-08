package main

// ExecutionResult describes what happened during a DNS provider execution.
//
// Updated indicates whether the DNS record value was changed.
// OldIP represent the previous public IP addresses.
// NewIP represent the current public IP addresses.
type ExecutionResult struct {
	Updated bool
	OldIP   string `json:"old_ip"`
	NewIP   string `json:"new_ip"`
}

// DdnsExecutor represents a DNS provider implementation (e.g. Tencent DNSPod).
type DdnsExecutor interface {
	// Execute returns an ExecutionResult for each domain name.
	Execute() (map[string]*ExecutionResult, error)
}
