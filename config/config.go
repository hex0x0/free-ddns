package config

// Config models the user configuration file (by default: $HOME/.config/free-ddns/config.yaml).
//
// The struct is intended to be used with a YAML unmarshaller (e.g. gopkg.in/yaml.v3)
// so the `yaml` tags must match the YAML keys.
type Config struct {
	DomainNames      []string    `yaml:"domainNames" json:"domainNames"`
	IPAddressVersion string      `yaml:"ipAddressVersion" json:"ipAddressVersion"`
	DNSProvider      DNSProvider `yaml:"dnsProvider" json:"dnsProvider"`

	// Notifier is optional. When configured, free-ddns will send a message
	// to the chosen channel if any DNS record is updated.
	Notifier *Notifier `yaml:"notifier" json:"notifier"`
}

type DNSProvider struct {
	Name       string     `yaml:"name" json:"name"`
	Credential Credential `yaml:"credential" json:"credential"`
}

type Credential struct {
	Tencent    TencentCredential    `yaml:"tencent" json:"tencent"`
	Aliyun     AliyunCredential     `yaml:"aliyun" json:"aliyun"`
	Cloudflare CloudflareCredential `yaml:"cloudflare" json:"cloudflare"`
}

type TencentCredential struct {
	SecretID  string `yaml:"secretId" json:"secretId"`
	SecretKey string `yaml:"secretKey" json:"secretKey"`
}

type AliyunCredential struct {
	AccessKeyID     string `yaml:"accessKeyId" json:"accessKeyId"`
	AccessKeySecret string `yaml:"accessKeySecret" json:"accessKeySecret"`
}

type CloudflareCredential struct {
	Token string `yaml:"token" json:"token"`
}

// Notifier describes where to send update notifications.
//
// Example:
// notifier:
//
//	name: telegram
//	credential:
//	  telegram:
//	    chatId: "123"
//	    botToken: "xxx"
type Notifier struct {
	Name       string             `yaml:"name" json:"name"`
	Credential NotifierCredential `yaml:"credential" json:"credential"`
}

type NotifierCredential struct {
	Telegram TelegramCredential `yaml:"telegram" json:"telegram"`
}

type TelegramCredential struct {
	ChatID   string `yaml:"chatId" json:"chatId"`
	BotToken string `yaml:"botToken" json:"botToken"`
}
