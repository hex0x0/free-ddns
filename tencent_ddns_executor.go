package main

import (
	"fmt"

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

func (executor *TencentDdnsExecutor) Execute() (map[string]*ExecutionResult, error) {
	res := map[string]*ExecutionResult{}

	exeFailedDomains := make([]string, 0)

	for _, dm := range executor.domainNames {
		domain, subdomain := ParseDomain(dm)
		logrus.Debugf("domain=%s subdomain=%s", domain, subdomain)

		publicIP, err := executor.ipGetter.GetPublicIP()
		if err != nil {
			logrus.Errorf("getPublicIP return err, err=%+v", err)
			exeFailedDomains = append(exeFailedDomains, dm)
			continue
		}
		logrus.Debugf("publicIP=%s", publicIP)

		currentDnsRecord, err := executor.queryDnsRecord(domain, subdomain)
		if err != nil {
			logrus.Errorf("queryDnsRecord failed, err=%+v", err)
			exeFailedDomains = append(exeFailedDomains, dm)
			continue
		}

		if currentDnsRecord == nil {
			logrus.Infof("createDnsRecord domain=%s.%s ip=%s", subdomain, domain, publicIP)
			if err := executor.createDnsRecord(domain, subdomain, publicIP); err != nil {
				logrus.Errorf("createDnsRecord failed, err=%+v", err)
				exeFailedDomains = append(exeFailedDomains, dm)
				continue
			}
			res[dm] = &ExecutionResult{
				Updated: true,
				OldIP:   "",
				NewIP:   publicIP,
			}
			continue
		}

		if *currentDnsRecord.Value != publicIP {
			logrus.Infof("updateDnsRecord domain=%s.%s old ip=%s new ip=%s", subdomain, domain, *currentDnsRecord.Value, publicIP)
			if err := executor.updateDnsRecord(currentDnsRecord.RecordId, domain, subdomain, publicIP); err != nil {
				logrus.Errorf("updateDnsRecord failed, err=%+v", err)
				exeFailedDomains = append(exeFailedDomains, dm)
				continue
			}
			res[dm] = &ExecutionResult{
				Updated: true,
				OldIP:   *currentDnsRecord.Value,
				NewIP:   publicIP,
			}
			continue
		}

		res[dm] = &ExecutionResult{
			Updated: false,
			OldIP:   *currentDnsRecord.Value,
			NewIP:   *currentDnsRecord.Value,
		}
	}

	if len(exeFailedDomains) > 0 {
		return res, errors.New(fmt.Sprintf("%v", exeFailedDomains))
	}

	return res, nil
}
