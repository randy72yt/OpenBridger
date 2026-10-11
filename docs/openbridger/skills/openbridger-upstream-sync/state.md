# 上游同步状态(state)

本文件是 `scripts/upstream-sync-check.sh` 的状态源,字段格式固定(脚本只读以下三行,勿改格式):

```
baseline_release: v1.0.0-rc.34
baseline_point: 6298b0f3238461b9629dfc1c00866f8325123aa1
last_check: 2026-10-10
```

- `baseline_release`:基座已跟进的上游 release(fork 时点对应的最近发布)
- `baseline_point`:diff 区间起点(fork 点 = rc.34 后 6 个上游提交;每次整合完成后改为新 release tag 指向的 commit sha)
- `last_check`:上次运行检查脚本的日期

## 决策日志

| 日期 | release 范围 | 动作 | 备注 |
| --- | --- | --- | --- |
| 2026-09-07 | v1.0.0-rc.34+6 | fork 建仓 | fork 点 `6298b0f32` |
| 2026-10-10 | rc.35–rc.43(9 个 release) | 检查-跳过(⚪) | 判定器首跑:提交与 release notes 双扫描无安全命中(test 基建误报已过滤);热区 16/18、阵地 15/16 被上游改动——待入口②③拉动再整合 |
