# 事故：上游 A 全模型 503(route intent/内部 revision 错误)
日期 / 定级：2026-10-11（UTC 约 07:10 起）/ S2（核心链路：推理调用）/ 状态：进行中，待上游恢复

## 时间线
- 07:0x（UTC）v0.1.5 部署完成，验收 20/20（gpt-5.6-luna 200），13 个模型抽检通过
- 07:0x–07:1x 全目录扫测：5 个模型异常（deepseek-v4-flash 503、claude-haiku-4-5 / gemini-3.6-flash / gemini-3.7-flash 慢响应后 503、gpt-6-luna 424），按 E4 下架，38→33
- 07:4x 复测：此前通过的 glm-5.2、gpt-5.6-luna、deepseek-v4-pro、kimi-k3 也开始 503 → 判定上游大范围故障

## 影响面
试运营期（外部流量近零）。所有模型经上游 A 返回 503；本站路由、计费、部署本身健康。

## 根因
上游 A（同为 new-api 系，更新版本）内部调度故障。错误体均为其内部机制：
- `unified upstream request route intent failed: active projection authority revision 264776 differs from scheduler revision 264771 after bounded convergence`
- `active exact projection cannot admit a new request: runtime snapshot is stale`
- `The service is temporarily unavailable`
这些字符串在本仓代码中不存在（grep 验证），确认非本站缺陷。诱因与我们的模型下架无关（时间上先于 DB 变更出现）。

## 处置过程
- 模型目录扫测 + 二次复测定位；5 个稳定失败模型按 E4 下架（channels.models + abilities）
- `obctl smoke` 全过（本站边缘/容器/三域名 200）
- 不再扩大下架（属瞬时故障而非单模型下线，恢复后 33 个应自动可用；5 个已下架模型待上游稳定后重测再上架）

## 防再发
- [ ] 恢复后全目录复测一遍，恢复下架 5 模型中可用的
- [ ] RISK-104（单上游无故障转移）本次实证——第二渠道（事项 rftiXH）优先级应上调
- [ ] monitor.py 目前只测 /api/status（本站自身），对「上游全挂」无感知：加一条定时真实调用拨测（P1）
