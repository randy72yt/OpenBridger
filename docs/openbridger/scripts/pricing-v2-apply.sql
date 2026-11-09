-- OpenBridger 费率方案 v2.3（成本对齐 + 积分制，模板保形重写）  生成时间 2026-11-05
-- 生成方式：以生产现网 billing_expr 为模板，各 tier 分支内所有系数按统一折扣因子缩放；
--   折扣池（group_ratio<0.5）因子 = 挂牌入价×0.6 / 现价 p 系数；渠道价池 = 成本入价×1.15 / 现价 p 系数。
--   完整保留 long_context / 0_200k / peak-off_peak 分档、img/ai 项与 service_tier 倍率。
-- 模型清单：现网 47 个（deepseek-v4-flash 无上游报价，维持现价不动）+ 新增 gemini-3-pro、gemini-3.1-pro（克隆 gemini-3.1-pro-preview 模板）= 49
-- 前置：备份 options / subscription_plans
--   mysqldump openbridger options subscription_plans > /srv/openbridger/backup-pricing-$(date +%F-%H%M).sql

START TRANSACTION;

UPDATE options SET value = '{"claude-fable-5":"tier(\\"base\\", p * 6.0 + c * 30.0)","claude-fable-5-1":"tier(\\"base\\", p * 6.0 + c * 30.0)","claude-haiku-4-5":"tier(\\"base\\", p * 0.6 + c * 3.0)","claude-opus-4-6":"tier(\\"base\\", p * 3.0 + c * 15.0)","claude-opus-4-7":"tier(\\"base\\", p * 3.0 + c * 15.0)","claude-opus-4-8":"tier(\\"base\\", p * 3.0 + c * 15.0)","claude-opus-5":"tier(\\"base\\", p * 3.0 + c * 15.0)","claude-opus-5-5":"tier(\\"standard\\", p * 2.4 + cr * 0.12 + cc * 3.0 + cc1h * 8 + c * 12.0)","claude-sonnet-4-6":"tier(\\"base\\", p * 1.8 + c * 9.0)","claude-sonnet-5":"tier(\\"base\\", p * 1.2 + c * 6.0)","claude-sonnet-5-5":"tier(\\"standard\\", p * 1.2 + cr * 0.12 + cc * 1.5 + cc1h * 4 + c * 6.0)","codex-auto-review":"tier(\\"base\\", p * 0.18 + c * 0.9)","deepseek-v4-flash":"weekday(\\"UTC\\") >= 1 && weekday(\\"UTC\\") <= 5 && ((hour(\\"UTC\\") >= 1 && hour(\\"UTC\\") < 4) || (hour(\\"UTC\\") >= 6 && hour(\\"UTC\\") < 10)) ? tier(\\"peak\\", p * 1.32 + cr * 0.042 + c * 3.96) : tier(\\"off_peak\\", p * 0.66 + cr * 0.021 + c * 1.98)","deepseek-v4-pro":"tier(\\"base\\", p * 0.759 + c * 2.277)","deepseek-v4.1-flash":"weekday(\\"UTC\\") >= 1 && weekday(\\"UTC\\") <= 5 && ((hour(\\"UTC\\") >= 1 && hour(\\"UTC\\") < 4) || (hour(\\"UTC\\") >= 6 && hour(\\"UTC\\") < 10)) ? tier(\\"peak\\", p * 0.345 + cr * 0.0069 + c * 1.38) : tier(\\"off_peak\\", p * 0.1725 + cr * 0.00345 + c * 0.69)","doubao-seedance-2.0":"tier(\\"base\\", p * 4.058824 + c * 13.728375)","doubao-seedance-2.0-fast":"tier(\\"base\\", p * 3.264706 + c * 8.881921)","doubao-seedance-2.0-mini":"tier(\\"base\\", p * 1.014706 + c * 0.858023)","doubao-seedance-2.5":"tier(\\"base\\", p * 6.176471 + c * 31.79066)","gemini-2.5-flash":"tier(\\"standard\\", p * 0.18 + img * 0.18 + ai * 0.6 + c * 1.5 + cr * 0.018 + cc * 0.05)","gemini-2.5-flash-lite":"tier(\\"standard\\", p * 0.06 + img * 0.06 + ai * 0.18 + c * 0.24 + cr * 0.006 + cc * 0.05)","gemini-2.5-pro":"len <= 200000 ? tier(\\"standard\\", p * 0.75 + img * 0.75 + ai * 0.75 + c * 6.0 + cr * 0.075 + cc * 0.225) : tier(\\"long_context\\", p * 1.5 + img * 1.5 + ai * 1.5 + c * 9.0 + cr * 0.15 + cc * 0.225)","gemini-3-flash":"tier(\\"standard\\", p * 0.3 + cr * 0.03 + ai * 0.6 + c * 1.8)","gemini-3.1-flash-lite":"tier(\\"base\\", p * 0.3 + c * 0.9)","gemini-3.1-pro-preview":"len <= 200000 ? tier(\\"0_200k\\", p * 1.2 + c * 7.2 + cr * 0.12) : tier(\\"200k_plus\\", p * 2.4 + c * 10.8 + cr * 0.24)","gemini-3.5-flash":"tier(\\"standard\\", p * 0.9 + img * 0.9 + ai * 0.9 + c * 5.4 + cr * 0.09 + cc * 0.05)","gemini-3.6-flash":"tier(\\"base\\", p * 0.936 + c * 4.68)","gemini-3.7-flash":"tier(\\"base\\", p * 0.45 + c * 2.25)","gemini-3.8-flash":"tier(\\"base\\", p * 0.45 + c * 2.25)","gemini-4-argon":"tier(\\"standard\\", p * 1.2 + cr * 0.06 + ai * 1.2 + c * 6.0)","glm-5.2":"tier(\\"base\\", p * 1.449 + c * 4.554)","glm-5.3":"tier(\\"base\\", p * 1.449 + c * 4.554)","gpt-5.3-codex-spark":"tier(\\"standard\\", p * 0.525 + cr * 0.0525 + c * 4.2) * (param(\\"service_tier\\") == \\"priority\\" ? 2.5 : 1) * (param(\\"service_tier\\") == \\"fast\\" ? 2.5 : 1)","gpt-5.5":"tier(\\"base\\", p * 3.0 + c * 18.0)","gpt-5.5-openai-compact":"tier(\\"base\\", p * 3.0 + c * 18.0)","gpt-5.6":"len <= 272000 ? tier(\\"standard\\", p * 2.4 + cr * 0.24 + cc * 3.0 + c * 12.0) : tier(\\"long_context\\", p * 4.8 + cr * 0.48 + cc * 6.0 + c * 18.0)","gpt-5.6-luna":"(len <= 272000 ? tier(\\"standard\\", p * 0.12 + c * 0.72 + cr * 0.012 + cc * 0.15) : tier(\\"long_context\\", p * 0.24 + c * 1.08 + cr * 0.024 + cc * 0.3)) * (param(\\"service_tier\\") == \\"fast\\" ? 2 : 1) * (param(\\"service_tier\\") == \\"priority\\" ? 2 : 1) * (param(\\"service_tier\\") == \\"flex\\" ? 0.5 : 1)","gpt-5.6-sol":"(len <= 272000 ? tier(\\"standard\\", p * 3.0 + c * 18.0 + cr * 0.3 + cc * 3.75) : tier(\\"long_context\\", p * 6.0 + c * 27.0 + cr * 0.6 + cc * 7.5)) * (param(\\"service_tier\\") == \\"fast\\" ? 2 : 1) * (param(\\"service_tier\\") == \\"priority\\" ? 2 : 1) * (param(\\"service_tier\\") == \\"flex\\" ? 0.5 : 1)","gpt-5.6-terra":"(len <= 272000 ? tier(\\"standard\\", p * 1.2 + c * 7.2 + cr * 0.12 + cc * 1.5) : tier(\\"long_context\\", p * 2.4 + c * 10.8 + cr * 0.24 + cc * 3.0)) * (param(\\"service_tier\\") == \\"fast\\" ? 2 : 1) * (param(\\"service_tier\\") == \\"priority\\" ? 2 : 1) * (param(\\"service_tier\\") == \\"flex\\" ? 0.5 : 1)","gpt-6-astra":"(len <= 272000 ? tier(\\"standard\\", p * 6.0 + c * 30.0 + cr * 0.6 + cc * 7.5) : tier(\\"long_context\\", p * 12.0 + c * 45.0 + cr * 1.2 + cc * 15.0)) * (param(\\"service_tier\\") == \\"fast\\" ? 2 : 1) * (param(\\"service_tier\\") == \\"priority\\" ? 2 : 1) * (param(\\"service_tier\\") == \\"flex\\" ? 0.5 : 1) * (param(\\"service_tier\\") == \\"ultrafast\\" ? 6 : 1)","gpt-6-luna":"tier(\\"base\\", p * 0.1265 + c * 0.6325)","gpt-6-sol":"(len <= 272000 ? tier(\\"standard\\", p * 1.2 + cr * 0.12 + cc * 1.5 + cc1h * 2.5 + c * 6.0) : tier(\\"long_context\\", p * 2.4 + cr * 0.24 + cc * 3.0 + cc1h * 5 + c * 9.0)) * (param(\\"service_tier\\") == \\"fast\\" ? 2 : 1) * (param(\\"service_tier\\") == \\"priority\\" ? 2 : 1) * (param(\\"service_tier\\") == \\"flex\\" ? 0.5 : 1)","gpt-6.1-sol":"(len <= 272000 ? tier(\\"standard\\", p * 1.2 + cr * 0.06 + cc * 1.5 + c * 6.0) : tier(\\"long_context\\", p * 2.4 + cr * 0.12 + cc * 3.0 + c * 9.0)) * (param(\\"service_tier\\") == \\"fast\\" ? 2 : 1) * (param(\\"service_tier\\") == \\"priority\\" ? 2 : 1) * (param(\\"service_tier\\") == \\"flex\\" ? 0.5 : 1)","grok-4.5":"len <= 200000 ? tier(\\"standard\\", p * 1.2 + c * 3.6 + cr * 0.18) : tier(\\"long_context\\", p * 2.4 + c * 7.2 + cr * 0.36)","grok-4.6":"len <= 200000 ? tier(\\"standard\\", p * 1.2 + c * 3.6 + cr * 0.3) : tier(\\"long_context\\", p * 2.4 + c * 7.2 + cr * 0.6)","grok-4.7":"len <= 200000 ? tier(\\"standard\\", p * 1.2 + cr * 0.3 + c * 3.6) : tier(\\"long_context\\", p * 2.4 + cr * 0.6 + c * 7.2)","kimi-k3":"tier(\\"base\\", p * 3.382353 + c * 16.911765)","gemini-3-pro":"len <= 200000 ? tier(\\"0_200k\\", p * 1.2 + c * 7.2 + cr * 0.12) : tier(\\"200k_plus\\", p * 2.4 + c * 10.8 + cr * 0.24)","gemini-3.1-pro":"len <= 200000 ? tier(\\"0_200k\\", p * 1.2 + c * 7.2 + cr * 0.12) : tier(\\"200k_plus\\", p * 2.4 + c * 10.8 + cr * 0.24)"}'
WHERE `key` = 'billing_setting.billing_expr';

UPDATE options SET value = '{"default":1,"vip":1,"svip":1}'
WHERE `key` = 'GroupRatio';

UPDATE subscription_plans SET total_amount = 1527778, subtitle = '每月 2,200 积分（面值 ¥22），约 4,000 次调用'
WHERE id = 1;
UPDATE subscription_plans SET total_amount = 1527778, subtitle = '每月 2,200 积分（面值 ¥22），6 个月内每月重置'
WHERE id = 5;

UPDATE subscription_plans SET total_amount = 3333334, subtitle = '每月 4,800 积分（面值 ¥48），约 8,700 次调用'
WHERE id = 2;
UPDATE subscription_plans SET total_amount = 3333334, subtitle = '每月 4,800 积分（面值 ¥48），6 个月内每月重置'
WHERE id = 6;

UPDATE subscription_plans SET total_amount = 6666667, subtitle = '每月 9,600 积分（面值 ¥96），约 17,400 次调用'
WHERE id = 3;
UPDATE subscription_plans SET total_amount = 6666667, subtitle = '每月 9,600 积分（面值 ¥96），6 个月内每月重置'
WHERE id = 7;

UPDATE subscription_plans SET total_amount = 16666667, subtitle = '每月 24,000 积分（面值 ¥240），约 43,600 次调用'
WHERE id = 4;
UPDATE subscription_plans SET total_amount = 16666667, subtitle = '每月 24,000 积分（面值 ¥240），6 个月内每月重置'
WHERE id = 8;

COMMIT;

-- ROLLBACK（2026-11-05 生产快照实测值；billing_expr 与 subtitle 从 mysqldump 备份恢复）
-- UPDATE options SET value='{"default":2,"svip":2,"vip":2}' WHERE `key`='GroupRatio';
-- UPDATE subscription_plans SET total_amount=2638889 WHERE id IN (1,5);   -- 旧面值 ¥38
-- UPDATE subscription_plans SET total_amount=5416667 WHERE id IN (2,6);   -- 旧面值 ¥78
-- UPDATE subscription_plans SET total_amount=9583333 WHERE id IN (3,7);   -- 旧面值 ¥138
-- UPDATE subscription_plans SET total_amount=17916667 WHERE id IN (4,8);  -- 旧面值 ¥258
-- billing_expr / subtitle 恢复：mysql openbridger < backup-pricing-*.sql（备份含全量 options）
-- 回滚后需 docker restart repo-openbridger-1
