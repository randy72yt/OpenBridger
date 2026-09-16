# 模型与价格说明

实际可用模型、接口协议和计费信息以 OpenBridger 控制台的模型广场为准。客户端中的模型名称必须与控制台展示的完整模型 ID 一致。

<div class="grid"><div class="card"><h3>对话与推理</h3><p>用于内容处理、推理、编程和 Agent 工作流。</p></div><div class="card"><h3>多模态</h3><p>部分模型支持图像理解、语音、绘图或视频任务，具体取决于模型及线路。</p></div><div class="card"><h3>协议兼容</h3><p>同一个模型可能只支持 Chat Completions、Responses 或 Anthropic 中的部分协议。</p></div></div>

## 选择模型时看什么

- 先确认客户端需要的协议，不要只看模型名称。
- 对比输入、输出、缓存、工具调用以及多模态计费说明。
- 先用短请求验证，再逐步增加上下文和工具调用。
- 生产项目为模型不可用、限流和超时设计降级方案。

<div class="actions"><a class="btn primary" href="{{CONSOLE_URL}}/pricing">打开模型广场</a></div>
