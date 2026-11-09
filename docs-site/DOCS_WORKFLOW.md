# OpenBridger 文档站迭代流程

公开教程的源码在 `docs-site/`。构建结果是静态 HTML，生产站点是 `https://docs.openbridger.com`。

`docs/openbridger/` 是内部运维文档（手册、台账、事故、恢复步骤）。不要把那些文件放进文档站，也不要在教程里复制服务器路径、数据库密码、上游供应商真名或兑换码批次明细。

本文件只描述怎么改教程。它不部署生产。

## 1. 改哪、看哪

| 要改的内容 | 位置 |
| --- | --- |
| 中文教程 | `docs-site/src/pages/*.md` |
| 英文教程 | `docs-site/src/pages-en/*.md`，与中文同一文件名 |
| 侧栏顺序和标题 | `docs-site/src/site.mjs` 的 `site` 与 `siteEn`，两边的 slug 列表必须一致 |
| 页头、搜索、主题 | `docs-site/public/app.js`、`docs-site/public/styles.css`、`docs-site/scripts/build.mjs` |
| 控制台和 API 地址 | 构建时替换。`{{CONSOLE_URL}}`、`{{API_BASE_URL}}`。生产构建设置 `OPENBRIDGER_SITE_URL=https://openbridger.com`，API Base 为该地址加 `/v1` |

路由规则：`guide--clients--cursor.md` 变成 `/zh/guide/clients/cursor/` 和 `/en/guide/clients/cursor/`。首页 slug 是 `introduction`，路径是 `/zh/` 和 `/en/`，不要链到 `/introduction/`。

站内链接写成根路径，例如 `[快速开始](/guide/quickstart/)` 或 `<a href="/guide/quickstart/">`。构建时会加上语言前缀。不要手写 `/zh/` 或 `/en/`。

控制台深链用完整占位符：`{{CONSOLE_URL}}/keys`、`/wallet`、`/pricing`、`/playground`、`/usage-logs`、`/subscriptions`、`/user-agreement`、`/privacy-policy`。

## 2. 一页教程要有什么

按这个顺序写，缺了再补，不要为了凑格式重复：

1. 这一页解决什么问题，用一两句说完。
2. 开始之前：账号、密钥、要先打开的控制台页面。
3. 步骤。菜单名以当前客户端为准；对不上时写明「找不到这项就不能按本页接入」，不要编一个不存在的按钮。
4. 能复制的配置：curl、以及需要时的 Python / Node。密钥用 `sk-...` 或环境变量，不要写真实密钥。
5. 怎么确认成功：使用日志里能对上密钥、模型 ID、请求 ID。
6. 常见错误：链到错误码页（`guide--errors.md`）里对应的状态，不要只写「检查网络」。
7. 相关页，用站内链接。

术语保持一致：API 密钥、Base URL、模型 ID、模型广场、钱包、使用日志、AI 体验台、额度、预扣、结算、分组、兑换码。英文页用 API key、Base URL、model ID、Model Square、wallet、Usage Logs、Playground。

截图只在步骤靠文字会看错时留占位，并写明要拍哪一个屏幕。不要用假图，也不要每页都放。

## 3. 什么不能写

- 不写假评价、假用户数、假「全网最低」或「无限」一类说法。
- 不写文档里维护的模型单价表。单价以模型广场为准。
- 不写翻墙、绕过网络限制、科学上网。这是开发者 API 文档。
- 不把用户协议改写成相反的意思。协议与教程冲突时，先改协议再改教程，或在教程里写明以主站协议为准。
- 代码里查不到的生产数字（当前赠送额度、当前汇率、当前在售模型清单）不要写成确定事实。写「以钱包 / 确认页 / 模型广场为准」，并把问题留在本文末尾的待确认列表。

## 4. 对照真实行为，而不是对照记忆

改调用方式、计费、充值、密钥、错误码之前，先在仓库里对上实现，再改两边语言。

| 教程在讲 | 先读 |
| --- | --- |
| 路径和协议 | `router/relay-router.go`、`router/task-plugin-protocol-router.go` |
| 密钥和请求头 | `middleware/auth.go`、`controller/token.go`、`web/src/features/keys/` |
| 错误正文 | `middleware/utils.go` 的 `abortWithOpenAiMessage`、`i18n/locales/zh-CN.yaml` |
| 预扣和结算 | `service/billing_session.go`、`service/billing.go` |
| 额度单位 | `common/constants.go` 的 `QuotaPerUnit`（默认 500000 = 1 美元） |
| 充值 | `controller/topup.go`、`setting/operation_setting/payment_setting_old.go` |
| 订阅扣费顺序 | `service/billing_session.go` 的 `NewBillingSession` |
| 控制台路径和中文菜单 | `web/src/routeTree.gen.ts`、`web/src/i18n/locales/zh.json` |
| 协议里已经写死的范围、退款、套餐 | `legal-drafts/published/`。主站页面优先于这份草稿 |

最小验证，不需要把生产密钥写进仓库：

```bash
# 密钥和模型换成你自己的。模型 ID 从模型广场复制。
curl -i "$OPENBRIDGER_BASE_URL/models" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY"

curl -i "$OPENBRIDGER_BASE_URL/chat/completions" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"MODEL_ID","messages":[{"role":"user","content":"pong"}]}'
```

然后在使用日志里找到 `X-Oneapi-Request-Id`。curl 成功但客户端失败时，只改客户端那一页。

本地没有可用密钥时，不要编一段「调用成功」的响应。把未核对的句子改成控制台核对步骤，或记进待确认列表。

## 5. 本地预览

在仓库根目录之外的 `docs-site/` 里构建。文档站不依赖 Go 模块。

```bash
cd docs-site
bun install
OPENBRIDGER_SITE_URL=https://openbridger.com bun run test
OPENBRIDGER_SITE_URL=https://openbridger.com bun run build
```

`bun run test` 会构建 `dist/` 并检查路由、站内链接和禁用用语。`bun run dev` 会监听 `src/` 并反复构建。

未设置 `OPENBRIDGER_SITE_URL` 时，构建默认使用 `https://openbridger.com`（见 `scripts/build.mjs`）。要用本机控制台预览时再改成 `http://localhost:3000`。

看页面：进入 `docs-site/dist/`，用任意静态文件服务器打开 `index.html`。`index.html` 会跳到 `/zh/`。不要只看 Markdown 源码，链接前缀是构建时加上的。

## 6. 提交前核对

- [ ] 改了中文页，就改了同名英文页。只改一边会让语言切换打开过期内容。
- [ ] `site.mjs` 里中英 slug 顺序一致，每个 slug 都有两个文件。
- [ ] 站内链接能在侧栏找到，没有 `/introduction/` 这种不会生成的路径。
- [ ] 示例里的 Base URL 用占位符。OpenAI 兼容含 `/v1`，Claude Code / Anthropic 用站点根。
- [ ] 没有真实密钥、兑换码、支付密钥、上游供应商域名。
- [ ] 没有模型单价、没有和协议相反的服务范围或退款承诺。
- [ ] `cd docs-site && OPENBRIDGER_SITE_URL=https://openbridger.com bun run test` 通过。
- [ ] 这次没有执行生产 rsync。

## 7. 部署（单独的运维动作）

文档站不是 Cloudflare Pages，仓库里也没有 Wrangler 配置。主站 CI 不会构建 `docs-site/`。合并 PR 不会更新 `https://docs.openbridger.com`。

生产是 Nginx 上的静态目录 `/srv/openbridger/docs-site/`，域名 `docs.openbridger.com` 经 Cloudflare 代理。备份不包含这个目录。

更新方式：在内存足够的机器上构建，把 `docs-site/dist/` 里面的文件同步到服务器的 `/srv/openbridger/docs-site/`。不要在生产主机上 `bun install` 或 `go build`，2C4G 上构建会把机器打满。不要把 `src/` 或整个 `docs-site/` 源码目录盖到 Web 根目录。

同步之后在本机确认这些地址返回 200：`/`、`/zh/`、`/en/`、`/zh/guide/quickstart/`、`/app.js`、`/styles.css`。再抽一页看 Base URL 已是 `https://openbridger.com/v1`，而不是占位符。

主站「文档链接」的配置项应指向 `https://docs.openbridger.com`。

## 8. 产品改了以后何时改文档

不设固定周更。下面任何一条落地，就在同一个 PR 里改教程，或紧跟着补一个只改 `docs-site/` 的 PR：

- 新增或关闭一条对外路径（路由文件）。
- 鉴权头、错误码、预扣规则、钱包扣费顺序变了。
- 充值方式、最低金额的含义、支付回跳地址变了。
- 控制台改名或改了 URL（密钥、钱包、模型广场、日志）。
- 用户协议或隐私政策升版本，尤其是服务范围、退款、套餐是否开放。

只改内部运维手册时，不要顺手改公开教程，除非用户能在控制台看到那项变化。

## 9. 待站长确认

下面几项这次没有写成确定事实。确认之后删掉对应的「以页面为准」，改成一句具体说明。

1. 生产环境的 `quota_display_type`。代码默认是 USD。运维记录写的是按人民币展示。公开教程要求用户看确认页，没有写死汇率。
2. `Price` 与 `USDExchangeRate`。代码默认都是 7.3。有一份运维示例用 7.2。两边不一致，所以教程不公布汇率。
3. `QuotaForNewUser`。代码默认 0。运维记录写的是 500000（约 1 美元）。教程只让用户看注册后的钱包。
4. 用户协议 v1.1 仍写「仅面向非中国大陆地区用户」，隐私政策同样。若下一阶段要面向中国大陆开发者，先更新主站协议，再改教程里的服务范围段。在那之前教程保持与协议一致。
5. 协议 v1.1 写明套餐购买和自动续费暂未开放。若已经实际上架套餐，先改协议，再改订阅页。
6. 状态页 `https://stats.uptimerobot.com/aHV66DcfpB` 是否仍是要公开的地址。
7. 「试运行、非 7×24、一般问题 1 个工作日内首次回复」是否仍是对外承诺。
8. 规范的 API 主机名是否只保留 `openbridger.com`。教程现在写明 `api.openbridger.com` 与主站是同一套网关，示例统一用主站。
9. 快速开始、API 密钥、钱包三页的「截图待补」需要当前控制台截图替换。
10. 若需要「当前在售模型」专页，应从模型广场或定价接口生成，不要手写进 Markdown。
