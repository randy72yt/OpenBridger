# Embeddings & Rerank

Knowledge-base (RAG) pipelines in Cherry Studio, Dify, Open WebUI, Zotero plugins, and LangChain rely on **Embeddings** for vector indexing and **Rerank** for second-stage relevance scoring.

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/embeddings</code></div>
<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/rerank</code></div>
<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/moderations</code></div>

## 1. Generate text embeddings

`input` accepts either a single string or an array of strings. Pick an embedding model available in [Model Square]({{CONSOLE_URL}}/pricing) (such as `text-embedding-3-small` or `text-embedding-3-large`):

```bash
curl "{{API_BASE_URL}}/embeddings" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "text-embedding-3-small",
    "input": [
      "Time complexity of self-attention mechanisms",
      "Load balancing in Mixture-of-Experts routing"
    ]
  }'
```

Each item in `data[i].embedding` contains the float vector. Note that changing embedding models changes vector dimensions and requires re-indexing your knowledge base.

## 2. Rerank candidate documents

When a rerank model is available in your group, `POST /v1/rerank` scores and orders retrieved passages against the user query:

```bash
curl "{{API_BASE_URL}}/rerank" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gte-rerank-v2",
    "query": "How do you mitigate cache stampede in distributed systems?",
    "top_n": 2,
    "documents": [
      "Cache penetration queries non-existent keys and can be filtered with Bloom filters.",
      "Cache avalanche/stampede happens when many keys expire simultaneously; add random TTL jitter and tiered caching.",
      "TCP congestion control uses slow start and congestion avoidance."
    ]
  }'
```

## 3. Content moderation

When a moderation model (e.g. `omni-moderation-latest`) is available in your group, call `POST /v1/moderations` with `{"model": "...", "input": "..."}`.

## Related

- [Cherry Studio](/guide/clients/cherry-studio/)
- [Dify, n8n & Coze](/guide/clients/dify-workflows/)
- [LangChain](/guide/clients/sdk/)
