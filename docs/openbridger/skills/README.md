# OpenBridger 运维 Agent 技能（随仓库分发版）

单技能设计，`openbridger-ops-agent/SKILL.md` 一份包含两部分：

- **Part 1 场景路由**（决策层）：故障应急 / 发版 Release / 备份恢复 / 日常巡检 / 密钥轮换 → 一键命令
- **Part 2 命令细节手册**（执行层）：服务器布局、构建部署坑、数据库操作、定价管线、备份监控

配套：

- 总纲文档：`../ops-handbook.md`（监控/应急/Release R0-R8/治理）
- 一键 CLI：`../scripts/obctl.sh`（部署在服务器 `/srv/openbridger/obctl`，不依赖任何 AI 工具）

## 在其他 AI 工具中使用

1. **支持 SKILL.md 约定的工具**（Claude Code 等）：把 `openbridger-ops-agent/` 目录放到工具的 skills 加载路径，或直接把 SKILL.md 内容粘进 system prompt / 项目说明
2. **不支持的工具**：把 SKILL.md 全文粘进对话开头即可，内容本身就是完整指令
3. 前提条件（与工具无关）：能 `ssh ob-prod`（或在 `~/.ssh/config` 配同名 Host 别名指向 47.80.28.213）

## 移植注意

- `dangerouslyDisableSandbox: true` 是 WorkBuddy 专属参数，其他工具忽略即可
- Mac 本地路径（`~/.bun/bin/bun`、`~/backups/openbridger/`）按实际机器调整

## 同步纪律

**源头在 `.workbuddy/skills/openbridger-ops-agent/SKILL.md`（WorkBuddy 实际加载的版本）**，本目录是分发副本。改动技能后同步：

```bash
cp .workbuddy/skills/openbridger-ops-agent/SKILL.md docs/openbridger/skills/openbridger-ops-agent/SKILL.md
```

