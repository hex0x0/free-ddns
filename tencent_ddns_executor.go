package main

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"

	"github.com/hex0x0/free-ddns/config"
)

const (
	dnsRecordLine = "默认"
)

type TencentDdnsExecutor struct {
	ipGetter    IPGetter
	client      *dnspod.Client
	domainNames []string
}

func InitTencentDdnsExecutor(cfg *config.Config) (*TencentDdnsExecutor, error) {
	credential := common.NewCredential(
		cfg.DNSProvider.Credential.Tencent.SecretID,
		cfg.DNSProvider.Credential.Tencent.SecretKey,
	)
	client, err := dnspod.NewClient(credential, "", profile.NewClientProfile())
	if err != nil {
		return nil, errors.New(fmt.Sprintf("init dnspod client failed, err: %v", err))
	}

	ipGetter := NewIpGetter(cfg)

	return &TencentDdnsExecutor{
		ipGetter:    ipGetter,
		client:      client,
		domainNames: cfg.DomainNames,
	}, nil
}

// queryDnsRecord query dns record
// doc：https://cloud.tencent.com/document/api/1427/95521
func (executor *TencentDdnsExecutor) queryDnsRecord(domain string, subdomain string) (*dnspod.RecordListItem, error) {
	req := dnspod.NewDescribeRecordFilterListRequest()
	req.Domain = common.StringPtr(domain)
	req.SubDomain = common.StringPtr(subdomain)
	req.RecordType = []*string{common.StringPtr(executor.ipGetter.GetDnsRecordType())}
	resp, err := executor.client.DescribeRecordFilterList(req)
	if err != nil {
		if resp != nil {
			logrus.Info(resp.ToJsonString())
		}
		return nil, errors.Errorf("DescribeRecordFilterList return err, domain=%s.%s, err=%+v", subdomain, domain, err)
	}
	if resp.Response == nil {
		return nil, errors.New("resp is empty")
	}

	for _, record := range resp.Response.RecordList {
		if *record.Type == executor.ipGetter.GetDnsRecordType() {
			return record, nil
		}
	}
	return nil, nil
}

// createDnsRecord create dns record
// doc：https://cloud.tencent.com/document/api/1427/56180
func (executor *TencentDdnsExecutor) createDnsRecord(domain string, subdomain string, publicIP string) error {
	req := dnspod.NewCreateRecordRequest()
	req.Domain = common.StringPtr(domain)
	req.SubDomain = common.StringPtr(subdomain)
	req.RecordType = common.StringPtr(executor.ipGetter.GetDnsRecordType())
	req.RecordLine = common.StringPtr(dnsRecordLine)
	req.Value = common.StringPtr(publicIP)
	resp, err := executor.client.CreateRecord(req)
	if err != nil {
		if resp != nil {
			logrus.Warn(resp.ToJsonString())
		}
		return errors.Errorf("client.CreateRecord failed, err=%+v", err)
	}
	return nil
}

// updateDnsRecord update dns record
// doc：https://cloud.tencent.com/document/api/1427/56157
func (executor *TencentDdnsExecutor) updateDnsRecord(recordId *uint64, domain string, subdomain string, publicIP string) error {
	req := dnspod.NewModifyRecordRequest()
	req.Domain = common.StringPtr(domain)
	req.SubDomain = common.StringPtr(subdomain)
	req.RecordType = common.StringPtr(executor.ipGetter.GetDnsRecordType())
	req.RecordLine = common.StringPtr(dnsRecordLine)
	req.Value = common.StringPtr(publicIP)
	req.RecordId = recordId
	resp, err := executor.client.ModifyRecord(req)
	if err != nil {
		if resp != nil {
			logrus.Warn(resp.ToJsonString())
		}
		return errors.Errorf("client.ModifyRecord failed, err=%+v", err)
	}
	return nil
}

func (executor *TencentDdnsExecutor) Execute() ExecutionResult {
	res := ExecutionResult{
		ExecutedAt: time.Now(),
		SuccessfulResult: SuccessfulResult{
			UpdatedDomains:    make([]DomainUpdateInfo, 0),
			UnmodifiedDomains: make([]string, 0),
		},
		FailedResult: make([]FailedInfo, 0),
	}

	publicIP, getIpErr := executor.ipGetter.GetPublicIP()

	for _, domainName := range executor.domainNames {
		if getIpErr != nil {
			logrus.Errorf("getPublicIP return err, err=%+v", getIpErr)
			res.FailedResult = append(res.FailedResult, FailedInfo{
				DomainNames: domainName,
				Reason:      "get public IP failed",
			})
			continue
		}

		logrus.Infof("publicIP=%s", publicIP)

		domain, subdomain := ParseDomain(domainName)
		logrus.Infof("domain=%s subdomain=%s", domain, subdomain)

		currentDnsRecord, err := executor.queryDnsRecord(domain, subdomain)
		if err != nil {
			logrus.Errorf("queryDnsRecord failed, err=%+v", err)
			res.FailedResult = append(res.FailedResult, FailedInfo{
				DomainNames: domainName,
				Reason:      fmt.Sprintf("queryDnsRecord failed: %v", err),
			})
			continue
		}

		if currentDnsRecord == nil {
			if err := executor.createDnsRecord(domain, subdomain, publicIP); err != nil {
				logrus.Errorf("createDnsRecord failed, err=%+v", err)
				res.FailedResult = append(res.FailedResult, FailedInfo{
					DomainNames: domainName,
					Reason:      fmt.Sprintf("createDnsRecord failed: %v", err),
				})
				continue
			}
			res.SuccessfulResult.UpdatedDomains = append(res.SuccessfulResult.UpdatedDomains, DomainUpdateInfo{
				DomainName: domainName,
				OldIp:      "",
				NewIp:      publicIP,
			})
			continue
		}

		if *currentDnsRecord.Value != publicIP {
			if err := executor.updateDnsRecord(currentDnsRecord.RecordId, domain, subdomain, publicIP); err != nil {
				logrus.Errorf("updateDnsRecord failed, err=%+v", err)
				res.FailedResult = append(res.FailedResult, FailedInfo{
					DomainNames: domainName,
					Reason:      fmt.Sprintf("updateDnsRecord failed: %v", err),
				})
				continue
			}
			res.SuccessfulResult.UpdatedDomains = append(res.SuccessfulResult.UpdatedDomains, DomainUpdateInfo{
				DomainName: domainName,
				OldIp:      *currentDnsRecord.Value,
				NewIp:      publicIP,
			})
			continue
		}

		res.SuccessfulResult.UnmodifiedDomains = append(res.SuccessfulResult.UnmodifiedDomains, domainName)
	}

	return res
}
