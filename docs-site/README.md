# OpenBridger 文档站

文档站与 OpenBridger 主项目保存在同一仓库，但作为独立静态应用构建和部署。

```bash
cd docs-site
bun install
OPENBRIDGER_SITE_URL=https://openbridger.com bun run build
```

构建结果位于 `dist/`。生产环境部署到 `https://docs.openbridger.com`，并在主站后台把“文档链接”设置为该地址。

未设置 `OPENBRIDGER_SITE_URL` 时，本地构建默认连接 `http://localhost:3000`。
