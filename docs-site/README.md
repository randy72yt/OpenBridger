# OpenBridger 文档站

公开教程在这个目录。`docs/openbridger/` 是内部运维文档，不会被本目录的构建收进去。

迭代流程、核对清单和部署方式见 [DOCS_WORKFLOW.md](./DOCS_WORKFLOW.md)。

```bash
cd docs-site
bun install
OPENBRIDGER_SITE_URL=https://openbridger.com bun run test
OPENBRIDGER_SITE_URL=https://openbridger.com bun run build
```

构建结果在 `dist/`。生产站点是 `https://docs.openbridger.com`。合并代码不会自动发布；发布是把 `dist/` 的内容同步到服务器上的静态目录，见流程文档。不要在生产主机上构建。

未设置 `OPENBRIDGER_SITE_URL` 时，构建使用 `https://openbridger.com`。本地控制台预览可改成 `http://localhost:3000`。
