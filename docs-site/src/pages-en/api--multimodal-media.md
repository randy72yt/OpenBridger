# Image, Video & Audio APIs

Beyond text and vision chat, OpenBridger routes OpenAI-compatible image generation/editing, asynchronous video generation, text-to-speech (TTS), speech-to-text (ASR), and Realtime WebSocket audio. Check [Model Square]({{CONSOLE_URL}}/pricing) to see which multimodal models are active for your key's group.

## 1. Image generation & editing

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/images/generations</code></div>
<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/images/edits</code></div>

```bash
curl "{{API_BASE_URL}}/images/generations" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-image-1",
    "prompt": "Warm minimalist study desk illustration in morning sunlight, crisp detail",
    "n": 1,
    "size": "1024x1024"
  }'
```

The response returns `data[0].url` or `data[0].b64_json` depending on the model and `response_format`. `/v1/images/edits` accepts `multipart/form-data` with `image` and optional `mask`. `POST /v1/images/variations` is not implemented (`501`).

---

## 2. Asynchronous video generation

Video generation follows a three-step async workflow (**create -> poll status -> fetch MP4**) when async video tasks and a video-capable model are enabled on your group.

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/videos</code></div>
<div class="endpoint-bar"><span class="method-badge method-get">GET</span><code>{{API_BASE_URL}}/videos/{video_id}</code></div>
<div class="endpoint-bar"><span class="method-badge method-get">GET</span><code>{{API_BASE_URL}}/videos/{video_id}/content</code></div>

### Submit & poll a video task

```bash
# 1. Create video task
curl "{{API_BASE_URL}}/videos" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "sora-2",
    "prompt": "Aerial slow-motion shot over an autumn university campus boulevard",
    "seconds": "5",
    "size": "1280x720"
  }'

# 2. Poll status every 3-5 seconds
curl "{{API_BASE_URL}}/videos/VIDEO_TASK_ID" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY"

# 3. Download completed MP4
curl -L "{{API_BASE_URL}}/videos/VIDEO_TASK_ID/content" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -o output.mp4
```

---

## 3. Speech (TTS, ASR & Realtime WebSocket)

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/audio/speech</code></div>
<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/audio/transcriptions</code></div>
<div class="endpoint-bar"><span class="method-badge method-ws">WS</span><code>{{API_BASE_URL}}/realtime?model={model}</code></div>

### Text-to-Speech (TTS)

```bash
curl "{{API_BASE_URL}}/audio/speech" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-mini-tts",
    "input": "Welcome to the OpenBridger unified multimodal gateway.",
    "voice": "alloy"
  }' \
  --output speech.mp3
```

### Speech-to-Text (ASR)

```bash
curl "{{API_BASE_URL}}/audio/transcriptions" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -F file="@./lecture-clip.mp3" \
  -F model="whisper-1"
```

### Realtime Audio (WebSocket)

Connect over `wss://` (replacing `https://` in `{{API_BASE_URL}}`) to `/v1/realtime?model=MODEL_ID` with headers `Authorization: Bearer $OPENBRIDGER_API_KEY` and `OpenAI-Beta: realtime=v1`.

## Related

- [Dify, n8n & Coze](/guide/clients/dify-workflows/)
- [Credits and billing](/guide/billing/)
- [Error codes](/guide/errors/)
