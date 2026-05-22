# free-ddns

DDNS client.

| DNS Provider     | IP v4 | IP v6 | Subdomain | Multi domains | Telegram |
|------------------|:-----:|:-----:|:---------:|:-------------:|:--------:|
| Tencent          |   ✅   |   ✅   |     ✅     |       ✅       |    ✅     |
| Aliyun (WIP)     |       |       |           |               |          |
| Cloudflare (WIP) |       |       |           |               |          |

## 1. Install

```bash
wget -qO- https://raw.githubusercontent.com/hex0x0/free-ddns/refs/heads/main/install_free-ddns.sh | bash
```

This will:

1. Install the binary via `go install github.com/hex0x0/free-ddns@latest`
2. Copy it to `/usr/local/bin/free-ddns`
3. Create `$HOME/.config/free-ddns/config.yaml` if it doesn’t exist
4. Install a systemd unit: `/etc/systemd/system/free-ddns.service` if it doesn’t exist

Enable and start service:

```
sudo systemctl enable free-ddns
sudo systemctl start free-ddns
```

## 2. Configuration

Config file location: `$HOME/.config/free-ddns/config.yaml`

The install script will create this file for you (if it doesn’t exist). After editing the config, restart the service:

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
chmod 600 "$HOME/.config/free-ddns/config.yaml"
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
# Location: ~/.config/free-ddns/config.yaml

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
