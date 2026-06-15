package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/hex0x0/free-ddns/config"
)

// IPGetter provides functions for retrieving the current public IP address
// and corresponding DNS record type.
//
// NOTE: The method names are intentionally unexported. That means only types in
// the same package can implement this interface.
type IPGetter interface {
	// GetDnsRecordType returns the DNS record type, e.g. "A" for IPv4 and "AAAA" for IPv6.
	GetDnsRecordType() string

	// GetPublicIP queries an external service to fetch the current public IP.
	GetPublicIP() (string, error)
}

func NewIpGetter() IPGetter {
	if config.Config.IPAddressVersion == "ipv4" {
		return IPV4Getter{}
	}

	return IPV6Getter{}
}

// IPV4Getter implements IPGetter for IPv4.
type IPV4Getter struct{}

func (IPV4Getter) GetDnsRecordType() string {
	return "A"
}

// GetPublicIP queries an IPv4 public-IP service.
//
// Reference: https://github.com/ihmily/ip-info-api?tab=readme-ov-file
func (IPV4Getter) GetPublicIP() (string, error) {
	const url = "https://ipv4.ddnspod.com"

	return query(url)
}

// IPV6Getter implements IPGetter for IPv6.
type IPV6Getter struct{}

func (IPV6Getter) GetDnsRecordType() string {
	return "AAAA"
}

// GetPublicIP queries an IPv6 public-IP service.
//
// Reference: https://www.ddnspod.com/
func (IPV6Getter) GetPublicIP() (string, error) {
	const url = "https://ipv6.ddnspod.com"

	return query(url)
}

func query(url string) (string, error) {
	httpClient := http.Client{Timeout: 10 * time.Second}
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", errors.Errorf("getPublicIP err=%+v", err)
	}
	httpReq.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")

	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", errors.Errorf("getPublicIP err=%+v", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("getPublicIP http status code=%d", httpResp.StatusCode))
	}

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", errors.Errorf("getPublicIP err=%+v", err)
	}

	return strings.TrimSpace(string(bodyBytes)), nil
}
