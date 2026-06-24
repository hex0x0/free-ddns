# free-ddns

DDNS client.

| DNS Provider     | IP v4 | IP v6 | Subdomain | Multi domains | Telegram |
|------------------|:-----:|:-----:|:---------:|:-------------:|:--------:|
| Tencent          |   ✅   |   ✅   |     ✅     |       ✅       |    ✅     |
| Aliyun (WIP)     |       |       |           |               |          |
| Cloudflare (WIP) |       |       |           |               |          |

## 1. Install

```bash
sudo dpkg -i free-ddns_amd64.deb
```

This will:

1. Install binary to `/usr/local/bin/free-ddns`
2. Install config file to `/usr/local/share/free-ddns/config.yaml` if it doesn’t exist
3. Install a systemd unit: `/lib/systemd/system/free-ddns.service` if it doesn’t exist

The installer **does not start the service automatically**.

Enable service at boot (optional):

```
sudo systemctl enable free-ddns.service
```

Start service manually when ready:

```bash
sudo systemctl start free-ddns.service
```

## 2. Configuration

Config file location: `/usr/local/share/free-ddns/config.yaml`

The install script will create this file for you. After editing the config, restart the service:

```bash
sudo systemctl restart free-ddns.service
sudo systemctl status free-ddns.service
```

### 2.1 Configure domain names

Set `domainNames` to the full hostnames you want to update (one or many):

```yaml
domainNames:
  - example.com
  - home.example.com
```

Notes:

* Use a **full hostname** (FQDN). `example.com` updates the root record (`@`). `home.example.com` updates the `home`
  subdomain.
* The current domain parser assumes the **last 2 labels** are the apex domain (e.g. `a.b.example.com` -> domain
  `example.com`, subdomain `a.b`). This may not work for multi-part TLDs like `example.co.uk`.

### 2.2 Choose IP version (IPv4 / IPv6)

Set `ipAddressVersion`:

* `ipv4` -> updates **A** records (public IP from `https://ipv4.ddnspod.com`)
* `ipv6` -> updates **AAAA** records (public IP from `https://ipv6.ddnspod.com`)

```yaml
ipAddressVersion: ipv4
```

### 2.3 Choose DNS provider

Set `dnsProvider.name`:

```yaml
dnsProvider:
  name: tencent
```

Supported values in config:

* `tencent` (implemented)
* `aliyun` (WIP)
* `cloudflare` (WIP)

#### 2.3.1 Set credentials

Credentials live under `dnsProvider.credential.<provider>`. Keep this file private (it contains secrets):

```bash
sudo chmod 600 /usr/local/share/free-ddns/config.yaml
```

#### 2.3.1.1 Tencent (DNSPod)

```yaml
dnsProvider:
  name: tencent
  credential:
    tencent:
      secretId: "YOUR_SECRET_ID"
      secretKey: "YOUR_SECRET_KEY"
```

Ensure the account/key has permission to manage DNSPod records for your domain.

### 2.4 Notification (optional)

If configured, `free-ddns` will send a message when **any DNS record is updated**.

Currently supported notifier:

* `telegram`

Config example:

```yaml
notifier:
  name: telegram
  credential:
    telegram:
      chatId: "123456789"
      botToken: "123456:ABCDEF..."
```

Notes:

* `chatId` can be a user, group, or channel chat id.
* The notifier HTTP client respects `http_proxy` / `HTTP_PROXY` environment variables.

### 2.5 Full example

```yaml
# Location: /usr/local/share/free-ddns/config.yaml

domainNames:
  - example.com
  - home.example.com

ipAddressVersion: ipv4

dnsProvider:
  name: tencent
  credential:
    tencent:
      secretId: "YOUR_SECRET_ID"
      secretKey: "YOUR_SECRET_KEY"
    aliyun:
      accessKeyId: "xx"
      accessKeySecret: "xx"
    cloudflare:
      token: "xxx"

# Optional. If configured, a message will be sent when any DNS record is updated.
notifier:
  name: telegram
  credential:
    telegram:
      chatId: "123456789"
      botToken: "123456:ABCDEF..."
```
