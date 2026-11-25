# OpenBridger 首页与品牌视觉设计约束规范 (Design System & Implementation Spec)

> **适用范围**：面向后续接手 OpenBridger 首页（Landing Page）、子页面、营销落地页及公网展示模块的 AI Agent 与前端工程师。  
> **设计基准**：当前已上线的 Warm White + Vibrant Orange 现代大模型网关视觉体系（见 PR #20 / `feat/landing-p0` 分支）。

---

## 1. 核心品牌定位与硬性约束 (Hard Rules)

### 1.1 业务与身份定位
- **定位**：OpenBridger 是一套面向开发者与企业的**多模态大模型统一 API 中转网关**（聚合 OpenAI、Anthropic Claude、Google Gemini、DeepSeek、Qwen 等 40+ 核心厂商、200+ 生产级模型，提供专线直连、智能路由与透明计费）。
- **必须严禁的文案与借用概念（负向约束）**：
  - ❌ **严禁照搬参考站（如快跑等）的品牌资产与特定业务**：严禁出现「快跑」、「快跑影创」、「快跑画布」、「一镜到底」、「AI 短剧成片」、「抽奖/积分兑换」、「开票/返现加盟」等无关概念与文案。
  - ❌ **严禁随意引入第三方 Logo 图片或外部 CDN 链接**：厂商图标统一使用本地依赖 `@lobehub/icons`。
  - ⚠️ **必须保留开源协议声明**：页面 Footer 必须保留 AGPLv3 Section 7 规范署名：`"Frontend design and development by New API contributors."` 及 GitHub 官方链接 `https://github.com/QuantumNous/new-api`。

### 1.2 双层 Token 架构范围约束
- **L1 全局生产力底座（`web/src/styles/openbridger.css`）**：为全站默认主题（`default` preset）及所有 `document.body` Portal 弹层提供克制干净的暖白中性画布（`#FAF8F5` 底 / `#FFFFFF` 卡片 / `#14110F` 炭黑字）与品牌活力橙（`--primary: #FF6B00`），保留控制台 `ConfigDrawer` 切换其他主题预设的能力，同时严格保留 `--success` 健康绿等交通灯功能语义色。
- **L2 公开页视觉层（`web/src/styles/public-surface.css`）**：通过 `.ob-landing-theme` 为公开页（`PublicLayout`）、认证页（`AuthLayout`）、初始化向导与全局错误页叠加暖米白底 `#FFFBF7`、暖色块 `#FFF6ED`、衬线标题 `.ob-serif-heading` 及橙色光晕主按钮 `.ob-btn-primary`。
- **默认亮暗策略**：全站首次访问（无历史 Cookie）统一默认浅色模式（`DEFAULT_THEME = 'light'`），支持手动切换浅色/深色/跟随系统。

---

## 2. 色彩系统与 Token 定义 (Color Palette)

全站主视觉为**浅色暖白底（Warm Cream）+ 醒目活力橙（Vibrant Orange）**，默认浅色，保留深色模式。

### 2.1 CSS 变量规范（`public-surface.css`）

| Token 变量 | 浅色模式 (Light) | 深色模式 (Dark) | 语义说明 |
| :--- | :--- | :--- | :--- |
| `--ob-bg` | `#FFFBF7` | `#14110F` | 公开页主背景色（暖白底 / 暖黑底） |
| `--ob-surf-warm` | `#FFF6ED` | `#1C1917` | 卡片表面、次级胶囊徽章底色 |
| `--ob-ink` | `#14110F` | `#F5F5F4` | 主标题与正文字体颜色 |
| `--ob-muted` | `#57534E` | `#A8A29E` | 次要文案、描述说明文字 |
| `--ob-line` | `rgba(20, 17, 15, 0.09)` | `rgba(255, 255, 255, 0.08)` | 发丝细边框、分割线 |
| `--ob-primary` | `#FF6B00` | `#FF6B00` | 品牌强调橙色（主 CTA、激活药丸） |
| `--ob-primary-light` | `#FF8533` | `#FF8533` | 悬停态强调、渐变文字辅色 |
| `--ob-primary-soft` | `#FFF7ED` | `rgba(255, 107, 0, 0.12)` | 激活标签底色、徽章半透背景 |
| `--ob-code-bg` | `#14110F` | `#14110F` | 终端演示卡片深色背景（双模恒黑） |

### 2.2 辅助原子类
- `.ob-btn-primary`：橙色主按钮，渐变微光 + 悬停轻微位移阴影 `shadow-md shadow-orange-500/20`。
- `.ob-btn-secondary`：次级按钮，白色/卡片底，细灰边框，轻微悬停变暗。
- `.ob-chip-primary`：胶囊状态药丸，软橙底 + 橙字 + 呼吸光点。
- `.ob-hero-gradient-text`：橙红微渐变文字（`from-[#FF6B00] via-[#FF7A1A] to-[#E65100]`）。

---

## 3. 排版与字体规范 (Typography)

### 3.1 标题系统
- **主标题（H1）**：中文采用**宋体/衬线体**（Serif），营造高端杂志社论质感。
  - 类名：`.ob-serif-heading`
  - 字体族：`font-family: 'Lora Variable', 'Noto Serif SC', 'Songti SC', 'STSong', Georgia, serif;`
  - 示例文案：
    - Line 1：`一套标准接口，`
    - Line 2：`驱动全栈大模型。`
- **副标题（Subtitle）**：
  - **英文斜体副标（Italic Subtitle）**：位于 H1 正下方，采用 `italic font-medium text-muted-foreground/90 text-base sm:text-lg`。
    - 示例：*Unified Multimodal Gateway · 专线直连与毫秒级智能路由*
  - **描述段落**：`text-sm sm:text-base leading-relaxed text-muted-foreground max-w-xl`。
- **等宽数据与代码（Mono）**：
  - 数字、端点路径、Token 计费使用 `.ob-mono`（`font-mono tabular-nums`）。

---

## 4. 关键组件几何形态与结构规范 (Component Patterns)

### 4.1 悬浮胶囊导航栏（`LandingHeader`）
- **外层容器**：全圆角悬浮浮岛形态（`rounded-full h-14`），外加 `border border-black/[0.06] backdrop-blur-xl` 与柔和阴影。
- **Logo 区域**：圆形徽标底座（`rounded-full size-8`）+ 系统名 + 极小大写标签 `API Gateway`。
- **导航链接**：
  - **激活项（Active）**：胶囊药丸底色 `rounded-full px-4 py-1.5 bg-[#FFF0E6] text-[#FF6B00] font-semibold`，**绝对不要加底部下划线**。
  - **未激活项**：`rounded-full px-3.5 py-1.5 text-foreground/80 hover:text-primary transition-colors`。
- **右侧功能区**：
  - 语言、深浅主题、消息通知：统一采用圆形按钮（`rounded-full size-9`）。
  - 主操作按钮：醒目橙色胶囊按钮 `控制台 ↗`（`rounded-full ob-btn-primary px-4 py-1.5 text-xs font-semibold`），图标固定使用右上箭头 `<ArrowUpRight className="size-3.5" />`。
- **登录态用户菜单（`LandingUserMenu`）**：
  - 触发器：药丸按钮 `[ (圆头像) 用户名 <ChevronDown /> ]`，类名 `rounded-full border bg-black/[0.03] pl-1 pr-2.5 py-1`。
  - 下拉面板：`rounded-xl p-1.5 shadow-xl`，必须包含以下规范项：
    - `👤 个人设置` (`/profile`)
    - `🔑 令牌管理` (`/keys`)
    - `💳 钱包管理` (`/wallet`)
    - 分割线
    - `🚪 退出`（触发 `SignOutDialog`）

### 4.2 Hero 区域核心结构（`LandingHero`）
- **顶部 Eyebrow 标签**：
  - 胶囊药丸 `rounded-full px-4 py-1.5 ob-chip-primary text-xs font-semibold`，带呼吸圆点。
  - 文案：`OpenBridger · 多模态大模型 API 网关 · 专线直连与毫秒级智能路由`。
- **CTA 按钮组**：
  - 必须全部为圆胶囊（`rounded-full h-11`）：
    - 主按钮：`获取 API 密钥 ↗`（带 `ArrowUpRight`）
    - 次按钮：`模型广场`（白底细边框）
    - 三级按钮：`实用教程`（软橙浅底带 `BookOpen` 图标）
- **数据指标栏（Metrics）**：
  - **禁止使用独立大卡片包裹**，采用杂志风**发丝细竖线分割排版**：
    `grid grid-cols-1 sm:grid-cols-3 divide-y sm:divide-y-0 sm:divide-x divide-black/[0.08]`
  - 三列内容规范：
    - 第 1 列：`200+ 生产级模型`（说明：`GPT-5 · Claude 4.5 · DeepSeek V3.1`）
    - 第 2 列：`全模态原生兼容`（说明：`流式对话 · 8K 生图 · 极速出片`）
    - 第 3 列：`专线高可用集群`（说明：`毫秒级智能路由 · 零门槛按量计费`）
- **头部厂商走马灯 / Logo 饰条（`HeroBrandTicker`）**：
  - 位于 Hero 区域最下方，使用发丝上边框 `border-t border-black/[0.06] pt-6` 分隔。
  - 规范引入 `@lobehub/icons` 图标（`size={18}`）：
    `OpenAI`, `Claude`, `Gemini`, `DeepSeek`, `阿里巴巴 Qwen`, `xAI Grok`, `智谱 GLM`, `MiniMax 海螺`, `Kimi Moonshot`, `字节跳动 豆包`, `快手可灵 Kling`, `即梦 AI`。

### 4.3 代码交互演示卡片（`LandingTerminalDemo`）
- **双模恒定深色**：底色固定为 `#14110F`，即使在浅色模式下也是暗色极客代码窗。
- **顶栏 Tab**：圆角微标签 `Chat`, `Responses`, `Claude`, `Gemini`，激活项为橙色字。
- **请求方法 Badge**：`POST` 使用橙色圆角药丸胶囊 `bg-orange-500/20 text-orange-400 rounded-md`。

---

## 5. 国际化 (i18n) 与代码规范

1. **i18n 完整覆盖**：
   - 严禁在 JSX 中硬编码中英文字符串。
   - 所有文案必须通过 `const { t } = useTranslation()` 提取，统一放入：
     - `web/src/i18n/locales/en.json`
     - `web/src/i18n/locales/zh.json`
     - `web/src/i18n/locales/zh-TW.json`
2. **测试与质量守护**：
   - 任何改动均须保持以下命令通过：
     - `cd web && bun run typecheck`
     - `cd web && npx oxlint -c .oxlintrc.json src/features/landing`
     - `cd web && bun run test src/features/landing/__tests__/landing-page.test.tsx`
     - `cd web && bun run build`
   - 测试必须断言负向规则：确认不含 `快跑影创`、`快跑画布`、`一镜到底` 等侵权或非网关文案。

---

## 6. 核心文件映射表 (File Map)

| 文件路径 | 职责定位 |
| :--- | :--- |
| `web/src/features/landing/landing-theme.css` | 首页色彩变量、文字渐变、阴影与动画定义 |
| `web/src/features/landing/components/landing-header.tsx` | 全圆角悬浮胶囊顶栏与导航路由 |
| `web/src/features/landing/components/landing-user-menu.tsx` | 登录态胶囊头像下拉菜单 |
| `web/src/features/landing/components/sections/landing-hero.tsx` | Hero 主视觉、衬线主标题、胶囊按钮、竖线指标 |
| `web/src/features/landing/components/hero-brand-ticker.tsx` | 底部大模型厂商 Logo 走马灯饰条 |
| `web/src/features/landing/components/landing-terminal-demo.tsx` | 实时 API 交互演示与代码高亮窗 |
| `web/src/features/landing/__tests__/landing-page.test.tsx` | 首页回归与合规单元测试 |
| `web/src/i18n/locales/{en,zh,zh-TW}.json` | 多语言文案字典文件 |
