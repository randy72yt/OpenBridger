# Embeddings 向量与 Rerank

构建本地知识库（RAG）、论文语义检索（如 Cherry Studio 知识库、Dify、Open WebUI、LangChain）时，通常需要调用 **向量化（Embeddings）** 与 **重排序（Rerank）** 接口。

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/embeddings</code></div>
<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/rerank</code></div>
<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/moderations</code></div>

## 1. 创建文本向量（Embeddings）

`input` 可以是单条字符串或字符串数组，请从 [模型广场]({{CONSOLE_URL}}/pricing) 选择向量模型（如 `text-embedding-3-small`、`text-embedding-3-large` 等）：

```bash
curl "{{API_BASE_URL}}/embeddings" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "text-embedding-3-small",
    "input": [
      "注意力机制的时间复杂度分析",
      "混合专家模型（MoE）的路由负载均衡"
    ]
  }'
```

响应体 `data[i].embedding` 为浮点向量数组。切换不同向量模型后，知识库原有的索引维度不同，需重新构建索引。

## 2. 检索结果重排序（Rerank）

在向量召回 Top-K 候选段落后，可用 Rerank 模型按查询相关性精细重排（需当前分组包含支持 Rerank 端点的模型）：

```bash
curl "{{API_BASE_URL}}/rerank" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gte-rerank-v2",
    "query": "如何防止分布式缓存雪崩？",
    "top_n": 2,
    "documents": [
      "缓存穿透是指查询不存在的数据，可通过布隆过滤器拦截。",
      "缓存雪崩是指大量键在同一时刻过期，可通过给 TTL 增加随机抖动与多级缓存缓解。",
      "TCP 拥塞控制包含慢启动与拥塞避免阶段。"
    ]
  }'
```

## 3. 内容审核（Moderations）

若当前分组配置了审核模型（如 `omni-moderation-latest`），可通过 `POST /v1/moderations` 检测输入文本合规性。

## 相关

- [Cherry Studio](/guide/clients/cherry-studio/)
- [Dify / n8n / Coze 工作流](/guide/clients/dify-workflows/)
- [LangChain 接入](/guide/clients/sdk/)
