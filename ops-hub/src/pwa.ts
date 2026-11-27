export const PWA_ICON_SVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 128 128">
  <defs>
    <linearGradient id="g" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0%" stop-color="#ff7a18"/>
      <stop offset="100%" stop-color="#e84500"/>
    </linearGradient>
  </defs>
  <rect width="128" height="128" rx="30" fill="#12100e"/>
  <rect x="10" y="10" width="108" height="108" rx="24" fill="url(#g)" opacity="0.14"/>
  <path d="M30 78 L52 46 L70 66 L96 34" fill="none" stroke="url(#g)" stroke-width="10" stroke-linecap="round" stroke-linejoin="round"/>
  <circle cx="96" cy="34" r="8" fill="#ff8f3f"/>
  <circle cx="34" cy="94" r="6" fill="#22c55e"/>
  <rect x="48" y="89" width="50" height="10" rx="5" fill="#faf6f2" opacity="0.85"/>
</svg>`

export const PWA_MANIFEST_JSON = JSON.stringify(
  {
    name: 'OB · NetGo 掌上总控',
    short_name: '掌上总控',
    description: 'OpenBridger 与 NetGo 实时经营纵览、模型可用率与节点监控',
    start_url: '/',
    display: 'standalone',
    background_color: '#0f0d0c',
    theme_color: '#0f0d0c',
    icons: [
      {
        src: '/icon.svg',
        sizes: 'any',
        type: 'image/svg+xml',
        purpose: 'any maskable',
      },
    ],
  },
  null,
  2
)

export function renderAppHtml(authenticated: boolean, authConfigured: boolean): string {
  return `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, viewport-fit=cover" />
  <meta name="theme-color" content="#0f0d0c" />
  <meta name="apple-mobile-web-app-capable" content="yes" />
  <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
  <meta name="apple-mobile-web-app-title" content="掌上总控" />
  <meta name="robots" content="noindex, nofollow" />
  <link rel="manifest" href="/manifest.webmanifest" />
  <link rel="icon" href="/icon.svg" type="image/svg+xml" />
  <link rel="apple-touch-icon" href="/icon.svg" />
  <title>OB · NetGo 掌上总控</title>
  <style>
    :root {
      --bg: #0f0d0c;
      --card: #191614;
      --card-2: #221e1b;
      --line: rgba(255, 255, 255, 0.08);
      --text: #faf6f2;
      --muted: #9f968e;
      --brand: #ff6b00;
      --brand-soft: rgba(255, 107, 0, 0.15);
      --ok: #22c55e;
      --ok-soft: rgba(34, 197, 94, 0.15);
      --warn: #f59e0b;
      --warn-soft: rgba(245, 158, 11, 0.16);
      --bad: #ef4444;
      --bad-soft: rgba(239, 68, 68, 0.16);
      --blue: #38bdf8;
      --blue-soft: rgba(56, 189, 248, 0.14);
    }
    * { box-sizing: border-box; -webkit-tap-highlight-color: transparent; }
    body {
      margin: 0;
      background: var(--bg);
      color: var(--text);
      font: 14px/1.5 -apple-system, BlinkMacSystemFont, "SF Pro Text", "PingFang SC", sans-serif;
      font-variant-numeric: tabular-nums;
      padding-bottom: calc(76px + env(safe-area-inset-bottom));
    }
    .topbar {
      position: sticky;
      top: 0;
      z-index: 20;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: calc(12px + env(safe-area-inset-top)) 16px 12px;
      background: rgba(15, 13, 12, 0.86);
      backdrop-filter: blur(16px);
      border-bottom: 1px solid var(--line);
    }
    .brand-title { font-size: 16px; font-weight: 700; letter-spacing: -0.02em; }
    .sub-time { font-size: 11px; color: var(--muted); margin-top: 2px; }
    .btn {
      border: 1px solid var(--line);
      background: var(--card-2);
      color: var(--text);
      border-radius: 999px;
      padding: 7px 13px;
      font-size: 12.5px;
      font-weight: 600;
      cursor: pointer;
    }
    .btn.primary {
      background: linear-gradient(135deg, #ff6b00, #e64800);
      border-color: transparent;
      color: #fff;
    }
    .btn:disabled { opacity: 0.5; }
    .wrap { max-width: 560px; margin: 0 auto; padding: 14px 14px 24px; }
    .hero-card {
      background: linear-gradient(145deg, #241912 0%, #171412 100%);
      border: 1px solid rgba(255, 107, 0, 0.25);
      border-radius: 20px;
      padding: 18px;
      margin-bottom: 14px;
    }
    .hero-label { font-size: 12px; color: var(--muted); display: flex; justify-content: space-between; align-items: center; }
    .hero-val { font-size: 32px; font-weight: 800; letter-spacing: -0.03em; margin: 6px 0 10px; color: #fff; }
    .pills { display: flex; flex-wrap: wrap; gap: 8px; }
    .pill {
      display: inline-flex;
      align-items: center;
      gap: 5px;
      padding: 3px 9px;
      border-radius: 999px;
      font-size: 11.5px;
      font-weight: 600;
    }
    .pill.ok { background: var(--ok-soft); color: #4ade80; }
    .pill.warn { background: var(--warn-soft); color: #fbbf24; }
    .pill.bad { background: var(--bad-soft); color: #f87171; }
    .pill.blue { background: var(--blue-soft); color: var(--blue); }
    .card {
      background: var(--card);
      border: 1px solid var(--line);
      border-radius: 18px;
      padding: 16px;
      margin-bottom: 12px;
    }
    .card-head {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 12px;
      font-weight: 700;
      font-size: 15px;
    }
    .kpi-grid {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 10px;
    }
    .kpi {
      background: var(--card-2);
      border-radius: 13px;
      padding: 11px 12px;
    }
    .kpi span { display: block; font-size: 11.5px; color: var(--muted); }
    .kpi strong { display: block; font-size: 18px; margin-top: 3px; }
    .kpi small { display: block; font-size: 11px; color: var(--muted); margin-top: 2px; }
    .list-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 10px;
      padding: 11px 0;
      border-top: 1px solid var(--line);
    }
    .list-row:first-child { border-top: 0; padding-top: 2px; }
    .row-title { font-weight: 600; font-size: 13.5px; word-break: break-all; }
    .row-sub { font-size: 11.5px; color: var(--muted); margin-top: 2px; }
    .bars { display: flex; gap: 2px; height: 12px; margin-top: 6px; align-items: flex-end; }
    .bar { flex: 1; height: 100%; border-radius: 2px; background: rgba(255,255,255,0.1); }
    .bar.ok { background: #22c55e; }
    .bar.warn { background: #f59e0b; }
    .bar.bad { background: #ef4444; }
    .trend-bars { display: flex; gap: 6px; height: 54px; align-items: flex-end; margin-top: 12px; padding-top: 6px; }
    .tcol { flex: 1; display: flex; flex-direction: column; align-items: center; gap: 4px; height: 100%; justify-content: flex-end; }
    .tcol-bar { width: 100%; border-radius: 5px 5px 2px 2px; background: linear-gradient(180deg, #ff852b, #ff5e00); min-height: 3px; }
    .tcol span { font-size: 10px; color: var(--muted); }
    .tabbar {
      position: fixed;
      left: 0;
      right: 0;
      bottom: 0;
      z-index: 30;
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      background: rgba(20, 17, 15, 0.94);
      backdrop-filter: blur(18px);
      border-top: 1px solid var(--line);
      padding: 6px 8px calc(6px + env(safe-area-inset-bottom));
    }
    .tab-btn {
      border: 0;
      background: none;
      color: var(--muted);
      padding: 6px 4px;
      font-size: 11px;
      font-weight: 600;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 3px;
      cursor: pointer;
    }
    .tab-btn.active { color: var(--brand); }
    .tab-icon { font-size: 17px; line-height: 1; }
    .login-box { max-width: 380px; margin: 14vh auto 0; padding: 24px 20px; }
    .input {
      width: 100%;
      padding: 12px 14px;
      border-radius: 12px;
      border: 1px solid var(--line);
      background: var(--card-2);
      color: var(--text);
      font-size: 15px;
      margin: 12px 0;
      outline: none;
    }
    .input:focus { border-color: var(--brand); }
    .err { color: #f87171; font-size: 12.5px; margin-top: 6px; }
    .action-stack { display: grid; gap: 10px; margin-bottom: 14px; }
    .toast {
      position: fixed;
      top: 72px;
      left: 50%;
      transform: translateX(-50%);
      background: #22c55e;
      color: #08120b;
      font-weight: 700;
      font-size: 12.5px;
      padding: 8px 16px;
      border-radius: 999px;
      z-index: 50;
      display: none;
    }
  </style>
</head>
<body>
  <div id="toast" class="toast"></div>
  ${
    !authConfigured
      ? `<div class="wrap"><div class="card login-box"><div class="card-head">尚未初始化安全口令</div><p class="row-sub">请通过 Wrangler 设置 <code>DASHBOARD_PASSPHRASE</code>（至少16位）与 <code>SESSION_SIGNING_SECRET</code>（至少32位）后再访问。</p></div></div>`
      : !authenticated
      ? `<div class="wrap">
          <form id="login-form" class="card login-box">
            <div class="card-head">🔐 OB · NetGo 掌上总控</div>
            <div class="row-sub">输入管理员访问口令解锁实时经营与节点状态</div>
            <input id="passphrase" type="password" class="input" placeholder="输入管理口令..." autocomplete="current-password" required />
            <button id="login-btn" type="submit" class="btn primary" style="width:100%;padding:11px">进入总控台</button>
            <div id="login-err" class="err"></div>
          </form>
        </div>
        <script>
          document.getElementById('login-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const btn = document.getElementById('login-btn');
            const errEl = document.getElementById('login-err');
            errEl.textContent = '';
            btn.disabled = true;
            try {
              const res = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ passphrase: document.getElementById('passphrase').value })
              });
              const data = await res.json();
              if (!res.ok || !data.success) {
                errEl.textContent = data.message || '口令验证失败';
                btn.disabled = false;
                return;
              }
              location.reload();
            } catch {
              errEl.textContent = '网络异常，请重试';
              btn.disabled = false;
            }
          });
        </script>`
      : `<header class="topbar">
          <div>
            <div class="brand-title">OB · NetGo 掌上总控</div>
            <div id="updated-at" class="sub-time">正在同步实时数据...</div>
          </div>
          <div style="display:flex;gap:8px">
            <button id="refresh-btn" class="btn" type="button">刷新</button>
            <button id="logout-btn" class="btn" type="button">退出</button>
          </div>
        </header>
        <main id="app-root" class="wrap"></main>
        <nav class="tabbar">
          <button class="tab-btn active" data-tab="overview" type="button"><span class="tab-icon">📊</span><span>总览</span></button>
          <button class="tab-btn" data-tab="models" type="button"><span class="tab-icon">🤖</span><span>OB 模型</span></button>
          <button class="tab-btn" data-tab="nodes" type="button"><span class="tab-icon">🌐</span><span>NetGo 节点</span></button>
          <button class="tab-btn" data-tab="alerts" type="button"><span class="tab-icon">🔔</span><span>微信推送</span></button>
        </nav>
        <script>
          let currentTab = 'overview';
          let hubData = null;

          function showToast(msg, ok = true) {
            const el = document.getElementById('toast');
            el.textContent = msg;
            el.style.background = ok ? '#22c55e' : '#ef4444';
            el.style.color = ok ? '#07130a' : '#fff';
            el.style.display = 'block';
            setTimeout(() => { el.style.display = 'none'; }, 2600);
          }

          function el(tag, attrs = {}, children = []) {
            const node = document.createElement(tag);
            for (const [k, v] of Object.entries(attrs)) {
              if (k === 'className') node.className = v;
              else if (k === 'text') node.textContent = v;
              else if (k === 'onclick') node.addEventListener('click', v);
              else node.setAttribute(k, v);
            }
            for (const c of children) {
              if (c == null) continue;
              node.appendChild(typeof c === 'string' ? document.createTextNode(c) : c);
            }
            return node;
          }

          function money(v) {
            return typeof v === 'number' && Number.isFinite(v) ? '¥' + v.toFixed(2) : '¥0.00';
          }

          function kpiBox(label, val, sub) {
            return el('div', { className: 'kpi' }, [
              el('span', { text: label }),
              el('strong', { text: val }),
              sub ? el('small', { text: sub }) : null
            ]);
          }

          function renderOverview(root, d) {
            const ob = d.openbridger.data;
            const ng = d.netgo.data;
            const totalToday = (ob ? ob.today.cash_revenue_cny : 0) + (ng ? ng.today.cash_revenue_cny : 0);
            const total7d = (ob ? ob.last_7d.cash_revenue_cny : 0) + (ng ? ng.last_7d.cash_revenue_cny : 0);

            const pills = [
              el('span', { className: 'pill ' + (d.openbridger.reachable && ob?.db_healthy ? 'ok' : 'bad'), text: d.openbridger.reachable ? 'OB 在线' : 'OB 异常' }),
              el('span', { className: 'pill ' + (d.netgo.reachable && ng?.host?.overall === 'healthy' ? 'ok' : 'bad'), text: d.netgo.reachable ? 'NetGo 正常' : 'NetGo 异常' }),
              el('span', { className: 'pill blue', text: '7日合计 ' + money(total7d) })
            ];

            root.appendChild(el('section', { className: 'hero-card' }, [
              el('div', { className: 'hero-label' }, [
                el('span', { text: '双平台今日现金总收入（北京时间）' })
              ]),
              el('div', { className: 'hero-val', text: money(totalToday) }),
              el('div', { className: 'pills' }, pills)
            ]));

            if (ob) {
              const marginTxt = ob.today.estimated_gross_margin_pct != null ? ob.today.estimated_gross_margin_pct.toFixed(1) + '% 毛利率' : '按消耗计';
              const trend = ob.daily_trend_7d || [];
              const maxRev = Math.max(1, ...trend.map(t => t.cash_revenue_cny));
              const trendEl = el('div', { className: 'trend-bars' }, trend.map(pt => {
                const h = Math.max(6, Math.round((pt.cash_revenue_cny / maxRev) * 42));
                const bar = el('div', { className: 'tcol-bar' });
                bar.style.height = h + 'px';
                return el('div', { className: 'tcol' }, [bar, el('span', { text: pt.date.slice(5) })]);
              }));

              root.appendChild(el('section', { className: 'card' }, [
                el('div', { className: 'card-head' }, [
                  el('span', { text: '🟠 OpenBridger 纵览' }),
                  el('span', { className: 'pill ok', text: 'v' + (ob.version || '-') })
                ]),
                el('div', { className: 'kpi-grid' }, [
                  kpiBox('今日现金营收', money(ob.today.cash_revenue_cny), ob.today.paid_order_count + ' 笔付款 · 新增 ' + ob.today.new_users + ' 人'),
                  kpiBox('今日消耗 / 毛利', money(ob.today.consumption_cny), '毛利 ' + money(ob.today.estimated_gross_profit_cny) + ' (' + marginTxt + ')'),
                  kpiBox('近 7 日现金营收', money(ob.last_7d.cash_revenue_cny), '7日毛利 ' + money(ob.last_7d.estimated_gross_profit_cny)),
                  kpiBox('上游渠道总余额', '$' + ob.channels.total_balance_usd.toFixed(2), '启用 ' + ob.channels.enabled + ' · 自动禁用 ' + ob.channels.auto_disabled)
                ]),
                trend.length ? trendEl : null
              ]));
            } else {
              root.appendChild(el('section', { className: 'card' }, [
                el('div', { className: 'card-head', text: '🟠 OpenBridger 暂未连接' }),
                el('div', { className: 'row-sub', text: d.openbridger.error || '请检查 OB_OPS_TOKEN' })
              ]));
            }

            if (ng) {
              root.appendChild(el('section', { className: 'card' }, [
                el('div', { className: 'card-head' }, [
                  el('span', { text: '🔵 NetGo 纵览' }),
                  el('span', { className: 'pill ' + (ng.nodes_summary.offline_visible > 0 ? 'bad' : 'ok'), text: ng.nodes_summary.online_visible + '/' + ng.nodes_summary.visible + ' 节点在线' })
                ]),
                el('div', { className: 'kpi-grid' }, [
                  kpiBox('今日订单收入', money(ng.today.cash_revenue_cny), ng.today.paid_order_count + ' 单 · 新增 ' + ng.today.new_users + ' 人'),
                  kpiBox('近 7 日订单收入', money(ng.last_7d.cash_revenue_cny), ng.last_7d.paid_order_count + ' 单成交'),
                  kpiBox('有效订阅用户', String(ng.users.active_subscribers), '总注册 ' + ng.users.total_users + ' 人'),
                  kpiBox('待处理工单', String(ng.tickets.pending_reply_tickets), '总开启工单 ' + ng.tickets.open_tickets)
                ])
              ]));
            } else {
              root.appendChild(el('section', { className: 'card' }, [
                el('div', { className: 'card-head', text: '🔵 NetGo 暂未连接' }),
                el('div', { className: 'row-sub', text: d.netgo.error || '请检查 NETGO_OPS_TOKEN' })
              ]));
            }
          }

          function renderModels(root, d) {
            const ob = d.openbridger.data;
            if (!ob) {
              root.appendChild(el('div', { className: 'card', text: 'OpenBridger 未连接：' + (d.openbridger.error || '') }));
              return;
            }
            const list = [...(ob.models?.models || [])].sort((a, b) => {
              const sa = (a.real?.requests || 0) + (a.probe?.requests || 0);
              const sb = (b.real?.requests || 0) + (b.probe?.requests || 0);
              const badA = sa > 0 && a.availability_24h < 90 ? 1 : 0;
              const badB = sb > 0 && b.availability_24h < 90 ? 1 : 0;
              if (badA !== badB) return badB - badA;
              return sb - sa;
            });

            const rows = list.length ? list.map(m => {
              const samples = (m.real?.requests || 0) + (m.probe?.requests || 0);
              const avail = samples === 0 ? null : m.availability_24h;
              const cls = avail == null ? 'blue' : avail >= 97 ? 'ok' : avail >= 85 ? 'warn' : 'bad';
              const ttft = m.real?.avg_ttft_ms || m.probe?.avg_ttft_ms || 0;
              const lat = m.real?.avg_latency_ms || m.probe?.avg_latency_ms || 0;
              const series = m.public_v2?.recent_success_series || [];
              const barsEl = series.length ? el('div', { className: 'bars' }, series.slice(-24).map(pt => {
                const c = pt.success_rate < 0 ? '' : pt.success_rate >= 97 ? 'ok' : pt.success_rate >= 80 ? 'warn' : 'bad';
                return el('span', { className: 'bar ' + c });
              })) : null;

              const left = el('div', { style: 'flex:1;min-width:0' }, [
                el('div', { className: 'row-title', text: m.model_name }),
                el('div', { className: 'row-sub', text: '真实 ' + (m.real?.requests || 0) + ' · 探测 ' + (m.probe?.requests || 0) + (ttft ? ' · 首字 ' + ttft + 'ms' : lat ? ' · 耗时 ' + lat + 'ms' : '') }),
                barsEl
              ]);
              const right = el('span', { className: 'pill ' + cls, text: avail == null ? '待采样' : avail.toFixed(1) + '%' });
              return el('div', { className: 'list-row' }, [left, right]);
            }) : [el('div', { className: 'row-sub', text: '暂无在售模型采样数据' })];

            root.appendChild(el('section', { className: 'card' }, [
              el('div', { className: 'card-head' }, [
                el('span', { text: '上游模型实时状态 (' + list.length + ')' }),
                el('span', { className: 'row-sub', text: '24h 可用率' })
              ]),
              ...rows
            ]));
          }

          function renderNodes(root, d) {
            const ng = d.netgo.data;
            if (!ng) {
              root.appendChild(el('div', { className: 'card', text: 'NetGo 未连接：' + (d.netgo.error || '') }));
              return;
            }
            const nodeRows = (ng.nodes || []).map(n => {
              const ageTxt = n.heartbeat_age_seconds != null ? n.heartbeat_age_seconds + 's 前心跳' : '无心跳记录';
              return el('div', { className: 'list-row' }, [
                el('div', {}, [
                  el('div', { className: 'row-title', text: n.name }),
                  el('div', { className: 'row-sub', text: String(n.type || '').toUpperCase() + ' · 倍率 ' + (n.rate || '1') + 'x · ' + ageTxt + (n.visible ? '' : ' · 已隐藏') })
                ]),
                el('span', { className: 'pill ' + (n.online ? 'ok' : 'bad'), text: n.online ? '在线' : '离线' })
              ]);
            });

            root.appendChild(el('section', { className: 'card' }, [
              el('div', { className: 'card-head' }, [
                el('span', { text: 'NetGo 节点列表' }),
                el('span', { className: 'pill ' + (ng.nodes_summary.offline_visible > 0 ? 'bad' : 'ok'), text: ng.nodes_summary.online_visible + '/' + ng.nodes_summary.visible + ' 在线' })
              ]),
              ...(nodeRows.length ? nodeRows : [el('div', { className: 'row-sub', text: '暂无节点记录' })])
            ]));

            const backupHours = ng.host.backup_age_seconds != null ? (ng.host.backup_age_seconds / 3600).toFixed(1) + 'h 前' : '-';
            root.appendChild(el('section', { className: 'card' }, [
              el('div', { className: 'card-head' }, [
                el('span', { text: 'ng-panel 主机与容器体检' }),
                el('span', { className: 'pill ' + (ng.host.overall === 'healthy' ? 'ok' : 'bad'), text: ng.host.overall })
              ]),
              el('div', { className: 'kpi-grid' }, [
                kpiBox('磁盘使用率', ng.host.disk_percent != null ? ng.host.disk_percent + '%' : '-', '阈值 < 80%'),
                kpiBox('可用内存', ng.host.memory_available_percent != null ? ng.host.memory_available_percent + '%' : '-', '阈值 >= 10%'),
                kpiBox('最新 MySQL 备份', backupHours, '自动异地备份'),
                kpiBox('失败检查项', String((ng.host.failures || []).length), (ng.host.failures || []).join(', ') || '全部通过')
              ])
            ]));
          }

          function renderAlerts(root, d) {
            const statusPill = el('span', {
              className: 'pill ' + (d.wecom_configured ? 'ok' : 'warn'),
              text: d.wecom_configured ? '企业微信已连接' : '待配企业微信密钥'
            });

            const testBtn = el('button', {
              className: 'btn primary',
              style: 'padding:12px',
              text: '📲 发送一条微信测试卡片',
              onclick: async () => {
                testBtn.disabled = true;
                try {
                  const r = await fetch('/api/push/test', { method: 'POST' });
                  const j = await r.json();
                  showToast(j.message || (r.ok ? '已推送到微信' : '推送失败'), r.ok && j.success);
                } catch {
                  showToast('网络请求失败', false);
                } finally {
                  testBtn.disabled = false;
                }
              }
            });

            const briefBtn = el('button', {
              className: 'btn',
              style: 'padding:12px',
              text: '📊 立即推送今日双平台经营简报到微信',
              onclick: async () => {
                briefBtn.disabled = true;
                try {
                  const r = await fetch('/api/push/briefing', { method: 'POST' });
                  const j = await r.json();
                  showToast(j.message || (r.ok ? '简报已发出' : '推送失败'), r.ok && j.success);
                  await loadHub();
                } catch {
                  showToast('网络请求失败', false);
                } finally {
                  briefBtn.disabled = false;
                }
              }
            });

            const events = d.recent_events || [];
            const eventRows = events.length ? events.map(ev => {
              const t = new Date(ev.ts * 1000).toLocaleString('zh-CN', { hour12: false });
              return el('div', { className: 'list-row' }, [
                el('div', {}, [
                  el('div', { className: 'row-title', text: ev.title }),
                  el('div', { className: 'row-sub', text: ev.summary + ' · ' + t })
                ])
              ]);
            }) : [el('div', { className: 'row-sub', text: '近期一切平稳，暂无异常告警记录' })];

            root.appendChild(el('section', { className: 'card' }, [
              el('div', { className: 'card-head' }, [el('span', { text: '微信主动推送控制' }), statusPill]),
              el('div', { className: 'action-stack' }, [testBtn, briefBtn]),
              el('div', { className: 'row-sub', text: '自动巡检：每 2 分钟检测 OB 宕机/模型掉线 & NetGo 节点离线/新工单；每天 09:00 自动推送经营早报。' })
            ]));

            root.appendChild(el('section', { className: 'card' }, [
              el('div', { className: 'card-head', text: '最近告警与推送记录' }),
              ...eventRows
            ]));
          }

          function render() {
            const root = document.getElementById('app-root');
            root.replaceChildren();
            if (!hubData) return;
            if (currentTab === 'overview') renderOverview(root, hubData);
            else if (currentTab === 'models') renderModels(root, hubData);
            else if (currentTab === 'nodes') renderNodes(root, hubData);
            else renderAlerts(root, hubData);
          }

          async function loadHub() {
            const btn = document.getElementById('refresh-btn');
            btn.disabled = true;
            try {
              const res = await fetch('/api/hub');
              if (res.status === 401) { location.reload(); return; }
              const j = await res.json();
              if (j.success && j.data) {
                hubData = j.data;
                document.getElementById('updated-at').textContent = '更新于 ' + new Date(hubData.fetched_at * 1000).toLocaleTimeString('zh-CN', { hour12: false });
                render();
              }
            } catch {
              showToast('刷新失败', false);
            } finally {
              btn.disabled = false;
            }
          }

          document.getElementById('refresh-btn').addEventListener('click', loadHub);
          document.getElementById('logout-btn').addEventListener('click', async () => {
            await fetch('/api/auth/logout', { method: 'POST' });
            location.reload();
          });
          document.querySelectorAll('.tab-btn').forEach(b => {
            b.addEventListener('click', () => {
              document.querySelectorAll('.tab-btn').forEach(x => x.classList.remove('active'));
              b.classList.add('active');
              currentTab = b.getAttribute('data-tab');
              render();
            });
          });

          loadHub();
        </script>`
  }
</body>
</html>`
}
