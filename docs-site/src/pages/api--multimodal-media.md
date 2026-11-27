# 图像、视频与语音接口

除文本与多模态对话外，OpenBridger 按标准路径提供文生图、图生图、异步视频生成以及语音合成（TTS）、语音转写（ASR）和实时语音（Realtime WebSocket）。某项能力是否可用，以 [模型广场]({{CONSOLE_URL}}/pricing) 当前在售的多模态模型与你的密钥分组为准。

## 一、图像生成与编辑

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/images/generations</code></div>
<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/images/edits</code></div>

### 文生图示例

```bash
curl "{{API_BASE_URL}}/images/generations" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-image-1",
    "prompt": "暖色调极简主义书桌插画，柔和晨光，高清细节",
    "n": 1,
    "size": "1024x1024"
  }'
```

响应的 `data[0]` 中会返回图片 `url` 或 `b64_json`（取决于模型及传入的 `response_format`）。图像编辑 `/v1/images/edits` 支持 `multipart/form-data` 上传原图 `image` 与可选遮罩 `mask`。`POST /v1/images/variations` 当前未开放（返回 `501`）。

---

## 二、异步视频生成（Sora / Veo / 可灵等）

视频生成耗时较长，采用**创建任务 -> 轮询状态 -> 获取成品**三步异步流程（需站点已开启异步任务且分组下有对应视频模型）。

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/videos</code></div>
<div class="endpoint-bar"><span class="method-badge method-get">GET</span><code>{{API_BASE_URL}}/videos/{video_id}</code></div>
<div class="endpoint-bar"><span class="method-badge method-get">GET</span><code>{{API_BASE_URL}}/videos/{video_id}/content</code></div>

### 1. 提交视频生成任务

```bash
curl "{{API_BASE_URL}}/videos" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "sora-2",
    "prompt": "航拍秋日大学校园林荫道，阳光穿过银杏叶，电影感慢镜头",
    "seconds": "5",
    "size": "1280x720"
  }'
```

接口立即返回任务 `id` 与初始状态（如 `queued` 或 `in_progress`）。

### 2. 轮询任务状态与下载

每隔 `3~5` 秒请求一次查询接口，直到状态变为完成，再从返回地址或 `/content` 端点拉取 MP4 文件，也可在控制台「任务日志」页面查看进度与结果：

```bash
# 查询状态
curl "{{API_BASE_URL}}/videos/VIDEO_TASK_ID" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY"

# 下载已完成视频内容
curl -L "{{API_BASE_URL}}/videos/VIDEO_TASK_ID/content" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -o output.mp4
```

---

## 三、语音合成（TTS）、语音识别（ASR）与实时语音

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/audio/speech</code></div>
<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/audio/transcriptions</code></div>
<div class="endpoint-bar"><span class="method-badge method-ws">WS</span><code>{{API_BASE_URL}}/realtime?model={model}</code></div>

### 1. 文本转语音（TTS）

```bash
curl "{{API_BASE_URL}}/audio/speech" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini-tts",
    "input": "欢迎使用 OpenBridger 多模态统一网关。",
    "voice": "alloy"
  }' \
  --output speech.mp3
```

### 2. 语音转文字（ASR / Whisper）

使用 `multipart/form-data` 上传本地音频文件：

```bash
curl "{{API_BASE_URL}}/audio/transcriptions" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -F file="@./lecture-clip.mp3" \
  -F model="whisper-1"
```

### 3. 实时双向语音（Realtime WebSocket）

对于支持 OpenAI Realtime 协议的模型，使用 WebSocket 连接 `wss://`（将 `{{API_BASE_URL}}` 的 `https://` 换成 `wss://`）并在请求头携带 `Authorization: Bearer $OPENBRIDGER_API_KEY` 与 `OpenAI-Beta: realtime=v1`。

## 相关

- [Dify / n8n / Coze 工作流](/guide/clients/dify-workflows/)
- [额度与计费](/guide/billing/)
- [错误码](/guide/errors/)
