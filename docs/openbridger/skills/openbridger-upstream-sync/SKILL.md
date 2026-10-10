---
name: openbridger-upstream-sync
description: OpenBridger 上游(new-api)同步值守——按 release 粒度检查基座更新,无变化跳过,有变化按取舍规则整合(安全修复必摘/适配器按需/重构跳过),过验证闸门并 PR 留痕。当任务涉及同步上游、吸收 new-api 更新、检查上游 release、cherry-pick 上游修复时使用。策略依据 docs/openbridger/ops-handbook.md §5.6;状态文件 state.md(本目录)。
agent_created: true
---

# OpenBridger 上游同步 Agent

你是 OpenBridger 的上游同步值守。**第一原则:只按 release 粒度同步、只摘需要的文件、每次整合过同一套验证闸门、全程 PR 留痕;绝不整分支 merge/rebase 上游。**

配套:`scripts/upstream-sync-check.sh`(检查器,exit 0=无变化 / exit 2=有报告)、`state.md`(基线与决策日志)、策略总纲 `ops-handbook.md` §5.6。

## 流程

### Step 0 检查(永远从这开始)

```bash
docs/openbridger/scripts/upstream-sync-check.sh
```

- **exit 0**:上游 release 未变 → 回报「无变化」即可结束,顺手把 `state.md` 的 `last_check` 更新为今天。
- **exit 2**:输出整合报告 → 继续 Step 1。

### Step 1 读报告 + release notes

对报告列出的**每个** release 读官方说明:

```bash
gh api repos/QuantumNous/new-api/releases/tags/<tag> --jq .body
```

结合报告里的「热区命中」「安全相关提交」「relay/channel 规模」三块,把区间内所有上游变更分成四桶(取舍规则,ops-handbook §5.6):

| 桶 | 判定 | 动作 |
| --- | --- | --- |
| ① 安全修复(auth/relay/webhook) | release notes 或提交明确是修复漏洞 | **无条件摘**,冲突手动解 |
| ② 新渠道、新模型适配器 | `relay/channel/<新目录>`、dto、注册行 | **问用户是否商业需要**,需要才摘;纯附加零冲突 |
| ③ 上游重构、目录搬家、接口改形 | 移动/重命名/抽象层变化 | **跳过**,除非它阻塞了①;跳过要在 PR 描述写明 |
| ④ 与定价控制面重叠的计费新功能 | billing/pricing 相关新能力 | **跳过**(自研替代在生产运行) |

产出一份「摘取清单」(摘什么文件/为什么)+「跳过清单」(跳过什么/为什么),先给用户过目再动手;纯①类且量小时可直接执行。

### Step 2 执行整合

```bash
git switch -c codex/upstream-sync-<YYYYMMDD>
git checkout v<release-tag> -- <path1> <path2> ...   # 注意:用 release tag,不用 upstream/main
```

- 热区文件若命中(报告会列出):不能整文件 checkout,改为 `git diff <range> -- <file>` 逐块把上游改动合入我们的版本,保住我们的就地修改。
- **计费阵地(GUARD,报告单独标出)**:`pkg/billingexpr/`、`relay/helper/price.go`、`service/quota.go`/`text_quota.go` 等我们计费路径依赖的上游文件——**billingexpr 引擎是上游自带的,我们的定价控制面是叠加层,上游动这些文件就是在动我们的地基**。命中时:禁止整文件 checkout;逐块 diff 审查上游改动是否移动我们的接线或改变计费语义;PR 描述必须附每个阵地文件的审查结论;闸门必跑 `go test ./pkg/billingexpr/... ./service/... ./relay/...` 与定价控制面用例。
- 上游的计费/定价「新功能」提交(expression drafts 等)即使看着诱人也一律④桶跳过——防止与自研控制面双体系并行。
- 同一 PR 只做一个主题(比如「安全修复」与「新适配器」分两个 PR)。

### Step 3 验证闸门(不过不上)

```bash
go build ./... && go test ./...
cd relaykit && GOWORK=off go build ./... && cd ..        # relaykit 独立构建
cd web && bun install --frozen-lockfile && bun run build && bun run vitest run && cd ..
```

- 涉及 model/数据库行为:跑三库矩阵(SQLite/MySQL/PG,见 AGENTS.md 数据库兼容规则),不可只测单库。
- 全绿后 PR(标题:`chore: upstream sync <tag> 摘取 X/跳过 Y`),PR 描述附摘取/跳过清单与验证结果。

### Step 4 收尾

1. merge 后更新 `state.md`:`baseline_release` → 新 tag;`baseline_point` → 该 tag 指向的 commit sha(`git rev-parse <tag>^{commit}`);决策日志追加一行。
2. **本技能到此为止**。生产部署是另一个技能(openbridger-ops-agent)的 R0–R8 流程,不要混在同步 PR 里。
3. 若 Step 0 报告有①类命中,提醒用户:部署优先级高于常规节奏。

## 红线

- 不 merge、不 rebase 上游分支;只 `git checkout <release-tag> -- <路径>` 或逐块 diff 合并。
- 不追 upstream/main 的中间提交;只认 release tag。
- 不给上游提 PR 反哺定价控制面(自有体系,上游未必兼容)。
- 热区清单(脚本内 `HOT_FILES`)与计费阵地清单(`GUARD_FILES`)同 ops-handbook §5.6 登记必须一致;我们新增就地修改或新增计费接缝时,三处同步更新。
- AGPL-3.0 不变:代码继续公开,配置/密钥不入库。

## 给非 WorkBuddy 工具的移植说明

本文件自包含,全文粘进任何 AI 工具的对话/系统提示即可执行。前提:仓库内能跑 `git`、`gh`(或脚本的 ls-remote 回退)、`go`、`bun`。不需要服务器权限(部署环节除外)。
