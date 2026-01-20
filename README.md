# services-health-check

目前狀態

- 支援 HTTP 健康檢測
- 支援 K8s Pod 健康檢測
- 支援 SSL 憑證到期檢測（適用 GCP Load Balancer 網域）
- 支援 Discord webhook 推播
- 設定來源：YAML 檔案 + `.env` / 環境變數覆蓋

1. 準備設定檔（建議先用範例）

  ```bash
  cp configs/example.yaml configs/local.yaml
  ```

  修改 `configs/local.yaml` 內容，至少要有：

- `checks`：HTTP / K8s / SSL 檢測目標
- `channels`：Discord webhook URL
- `routes`：告警路由（例如 `CRIT` → Discord）

1. 如果要用環境變數覆蓋

  `.env` 的格式可參考 `.env.example`。支援的 key 為。

- `CHECK_*`
- `POLICY_*`
- `CHANNEL_*`
- `ROUTE_*`
- `LOG_*`

1. 啟動

```bash
go run ./cmd/healthd -config configs/local.yaml
```

## Logger 設定

可用 YAML 或環境變數設定輸出層級、格式與檔案路徑。

YAML 範例：

```yaml
log:
  level: info
  format: text
  # file: /tmp/healthd.log
```

環境變數：

- `LOG_LEVEL`（debug/info/warn/error）
- `LOG_FORMAT`（text/json）
- `LOG_FILE`（可選）

## 快速驗證（k8s + SSL）

1. 編輯 `configs/example.yaml`，替換以下欄位：

- `kubeconfig`（本機測試可用 `~/.kube/config`）
- `context`
- `label_selector`
- `address` / `server_name`（填 GCP LB 網域或證書綁定的網域）
- `channels.discord.url`

1. 直接啟動：

```bash
go run ./cmd/healthd -config configs/example.yaml
```

## 設定優先序

1. `configs/*.yaml`
2. `.env`（可選）
3. 環境變數（可選）

## Discord webhook 格式

推播內容為純文字，格式：

```text
[STATUS] summary
message
```

## Slack / Google Chat webhook

Slack 與 Google Chat 都使用簡單的 `text` payload，設定方式與 Discord 類似：

```yaml
channels:
  - type: slack
    name: slack-alert
    url: https://hooks.slack.com/services/your/webhook
  - type: gchat
    name: gchat-alert
    url: https://chat.googleapis.com/v1/spaces/your/webhook
```

## 排程（Cron）

`schedule` 使用標準 5 欄位 cron（分鐘/小時/日期/月份/星期）。啟動後會先跑一次，再依 cron 規則執行。

範例：

```yaml
- type: k8s_pods
  name: gke-pods
  schedule: "*/5 * * * *"
```

## K8s Pod 檢測

K8s 檢測預設會嘗試 In-Cluster Config，若設定 `kubeconfig` 則會優先使用該檔案。

範例：

```yaml
- type: k8s_pods
  name: gke-pods
  namespace: default
  label_selector: app=your-app
  kubeconfig: ~/.kube/config
  context: your-context
  min_ready: 1
  schedule: "*/5 * * * *"
```

### 問題清單筆數

可用 `notify.problem_limit` 或環境變數 `PROBLEM_LIMIT` 統一調整 K8s 問題清單顯示筆數。

```yaml
notify:
  problem_limit: 5
```

## SSL 憑證到期檢測

範例：

```yaml
- type: ssl
  name: gcp-ssl
  address: your-domain.example.com:443
  server_name: your-domain.example.com
  skip_verify: false
  warn_before: 720h
  crit_before: 168h
  schedule: "0 * * * *"
```

`server_name` 可省略，會自動用 `address` 的 host 當 SNI。

## Cloudflare Token 檢測

使用 Cloudflare API `tokens/verify` 檢查 token 是否可用。

範例：

```yaml
- type: cloudflare_token
  name: cf-token
  token: ${CLOUDFLARE_TOKEN}
  schedule: "*/10 * * * *"
```

## Domain 到期檢測（WHOIS）

透過 WHOIS 讀取網域到期日。

範例：

```yaml
- type: domain_expiry
  name: domain-expiry-itrd
  domain: itrd.tw
  warn_before: 720h
  crit_before: 168h
  schedule: "0 0 * * *"
```

### 大量網域（環境變數展開）

可用 `CHECK_DOMAINS` 一次定義多個網域（逗號分隔），啟動時會自動展開成多個 `domain_expiry` 檢查。

```bash
CHECK_DOMAINS="itrd.tw,dunqian.tw,mastripms.com"
CHECK_DOMAIN_WARN_BEFORE=720h
CHECK_DOMAIN_CRIT_BEFORE=168h
CHECK_DOMAIN_SCHEDULE="0 0 * * *"
```

## 環境變數替換

YAML 內可使用 `${VAR}` 讀取環境變數，會在載入設定時自動替換。

## 注意事項

- 目前 env 覆蓋是針對單一組 check/channel/route 的簡化版（最小可用）。
- 之後可以擴充成多組配置與更細緻的路由條件。
