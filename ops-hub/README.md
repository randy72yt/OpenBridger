# OB · NetGo 掌上总控台（Cloudflare Worker PWA + 企业微信推送）

零服务器内存占用，统一展示 **OpenBridger**（今日/7日营收、消耗、毛利、PR #17 上游模型可用率与延迟）与 **NetGo**（今日/7日订单收入、有效订阅、工单、全部节点心跳与主机体检），并通过企业微信「微信插件」将告警与每日早报直推个人微信聊天列表。

## 1. 服务端只读令牌准备（生成两个独立随机 48 位 Hex 密钥）

```bash
# 生成两把不同的只读密钥（至少 32 字符）
openssl rand -hex 24   # 用作 OB_OPS_TOKEN / OPS_SNAPSHOT_TOKEN
openssl rand -hex 24   # 用作 NETGO_OPS_TOKEN
```

1. **OpenBridger（`ob-prod`）**：
   - 在 `/srv/openbridger/.env`（及 `compose.release.yml` 透传环境）中加入：
     `OPS_SNAPSHOT_TOKEN=<第一把密钥>`
   - 部署包含本分支的新镜像后，接口地址为：`https://openbridger.com/api/ops/snapshot`
2. **NetGo（`ng-panel`）**：
   - 将 `netgo-infrastructure/observability/ops-snapshot/netgo-ops-snapshot.py` 拷至 `/usr/local/sbin/netgo-ops-snapshot.py`（`chmod 755`）；
   - 新建 `/etc/netgo-observability/ops-snapshot.env`（`chmod 600`）填入 `NETGO_OPS_TOKEN=<第二把密钥>`；
   - 安装启动 `netgo-ops-snapshot.service`，并更新 `/etc/nginx/conf.d/ops.netgo.cc.conf` reload Nginx，接口地址为：`https://ops.netgo.cc/api/netgo-ops/snapshot`

## 2. 开通企业微信「微信插件」（2 分钟，无需营业执照）

1. 打开 [企业微信官网](https://work.weixin.qq.com/) 注册一个个人团队（无需认证、无需营业执照）。
2. 进入后台 **「我的企业」** 最底部，复制 **企业 ID（`WECOM_CORP_ID`）**。
3. 进入 **「应用管理 -> 自建 -> 创建应用」**，起名 `掌上总控`，可见范围选自己，创建后记下 **`AgentId`（`WECOM_AGENT_ID`）** 和 **`Secret`（`WECOM_CORP_SECRET`）**。
   - 在该应用底部点 **「企业可信 IP」**（如提示需要可信域名/IP，或若不想配固定 IP，也可直接在企业微信建一个只有你自己的内部群 -> 添加「群机器人」复制 Webhook 地址填入 `WECOM_WEBHOOK_URL` 即免可信 IP 限制！）。
4. 进入后台 **「我的企业 -> 微信插件」**，用你平时用的**个人微信扫码关注**。关注后，所有通知都会直接弹在你个人微信聊天列表里！

## 3. 一键部署 Cloudflare Worker

```bash
cd ops-hub
bun install

# 首次登录 Cloudflare 账号并创建状态存储 KV
npx wrangler login
npx wrangler kv namespace create OPS_STATE
# 将输出的 id 填入 wrangler.toml 的 [[kv_namespaces]] id 字段

# 写入安全密钥（全程加密存储，不进代码仓库）
npx wrangler secret put DASHBOARD_PASSPHRASE      # 手机打开网页时的登录口令（至少 16 位）
npx wrangler secret put SESSION_SIGNING_SECRET    # openssl rand -hex 32
npx wrangler secret put OB_OPS_TOKEN              # 对应 ob-prod 的 OPS_SNAPSHOT_TOKEN
npx wrangler secret put NETGO_OPS_TOKEN           # 对应 ng-panel 的 NETGO_OPS_TOKEN

# 企业微信推送（二选一：自建应用三件套 或 内部群机器人 Webhook）
npx wrangler secret put WECOM_WEBHOOK_URL         # 最简单免可信IP：填企业微信群机器人 Webhook
# 或者填自建应用三件套：WECOM_CORP_ID / WECOM_AGENT_ID / WECOM_CORP_SECRET

# 发布上线
npx wrangler deploy
```

部署完成后，在手机 Safari / Chrome 打开生成的 Worker 网址（或绑定的自定义子域名），登录后点击浏览器菜单 **「添加到主屏幕」**，即可获得全屏原生 App 图标。
