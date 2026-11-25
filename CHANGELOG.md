# 变更日志

## 2026-11-25（v1.15.14-panel）

### 面板跨平台统一：Qoder 规则提示 / 全平台积分构成与用量 / CodeArts 登录并签到

**Qoder（官方规则变更，必须一眼可见）**：官方文档确认自 2026-11-19 10:00 起，
每日 100 通用 Credits 仅面向 Pro 试用期及 Lite/Pro/Pro+/Ultra 个人付费订阅，
体验版（Free）不再开放。签到池新增红色规则条 + 可展开的完整活动规则
（时间、资格、领取窗口、有效期、扣减顺序、免费用户 FAQ）；Free 账号的状态
直接显示「官方已停发（Free）」，不再只写含糊的「未查到可领活动」。

**MiniMax（规则澄清）**：新增规则条，明确官方现行签到**不限 Pro**（每日 400，
第 4/7 天 1000，一周约 4000，30 天有效），Pro 的差异是活动期额度重置卡；
避免把 Qoder 的 Pro 限制套到 MiniMax 上。

**积分总览（可读性）**：平台标题从 13px 淡灰改为 15px 正文墨色 + 平台色条 +
平台徽标；五个平台的余额单元格统一为「剩余 / 总量 + 已用 · 计划/来源」，
WorkBuddy / Qoder / Trae / MiniMax / CodeArts 不再是五种写法。

**积分构成（全平台）**：标题改为「积分构成 · 全平台」，五个平台进同一网格；
卡片字段完全统一（剩余大字、总量/已用/包数、来源彩条、签到可得/到期）。
新增 MiniMax 官方 `/credit/details` 逐包明细（授予/剩余/已消耗/到期/类型），
CodeArts 只有总量就如实只画总量；Qoder 只有「套餐内 / 附加」两桶就画两桶。
WorkBuddy 的到期分布与逐包明细保留在下方独立区块。

**用量页（不再只有 WorkBuddy）**：新增「各平台额度用量」区块，五个平台
逐账号列出「已用 / 总量 / 剩余 / 使用率 / 数据口径」。WorkBuddy 有上游逐请求
用量；其余平台官方未开放逐请求，只展示额度快照，并明确标注「拿不到 ≠ 0」。

**CodeArts（华为）签到修复**：
- `today_gain` 只认**今天**（UTC+8）的领取记录；此前 11-04 领的 100 分会一直
  显示「今日 +100」，与 `checked_in_today=false` 自相矛盾；
- 账号行常驻「签到」按钮；会话失效时变成「登录并签到」——打开登录窗口，
  登录成功后自动保存、自动领取当天积分（前端再补一次幂等签到并回显结果）；
- 修正单号签到 toast 写死「+1000 积分」（实际 100）；
- 同人组跳过日志区分「今天已领到」与「今天打过但失败（需重登）」，
  不再让用户等一个不会发生的自动重试。

**任务中心（一键完成）**：
- 每个可自动任务行新增「一键完成」/「领取」按钮，账号组新增「一键完成本账号」；
- 需客户端/真实操作的任务行新增「指引」按钮，原样显示原因，不伪装成可完成；
- 全局按钮改名「一键完成可自动任务」，并明确：没有自动动作的任务不会被算作完成。

## 2026-11-22（v1.15.13-panel）

### 修复：Qoder「QQ 显示已签到」是**假标记**——陈旧 claim_cid 被 qoderPersonVerdict 当成「今日已领」

**现象（11-19 起每天必现）**：后台日志每 30 分钟一轮永远一模一样——
「本次签到 0 · 今日已领 1 · 未下发 1 · 待签 2」；QQ 号在面板显示「已签到」，
可积分余额自 11-19 起**零增长**、账本里也查不到它当天的任何条目。

**⚠️ 更正**：v1.15.12 条目②里「QQ 号当天真领到（本地可信记录 17:59）」的表述
是**错的**——那条「可信记录」正是本条修复的假标记产物，被它带偏出了「同设备配额」
的错误结论。实测（7 个环境码 × real/derived × 双 base 全量探测）：上游自 11-19 起
就没给这两个号下发每日活动，**一个号都没签上过**，不存在「只签到一个」。

**根因链（证据全部对上）**：

1. gmail 的号池快照存着上次领取的 `claim_cid`（11-18 那天的活动），QQ 没有；
2. 每轮 pool：QQ pending → 拿 gmail 的**陈旧 cid** 去只读探测 → QQ 当时真领过
   那条活动 → 服务端如实回 `status=CLAIMED + claimedAt=2026-11-18T02:03:41Z`；
3. `qoderPersonVerdict` 见 CLAIMED 即判 **Owner** → QQ 被标成「今日已领」：
   - `MarkClaimedAt` 写进**陈旧**领取时刻（Claims 停在 11-18，与实测值严丝合缝）；
   - `MarkCheckedIn` 每轮刷新 Entries（17:33:42 → 17:59:29，假「已签到」）；
   - 账本记录用陈旧 claimedAt 算日期 → **合并进 11-18 的历史条目**（claimed 粘性），
     当天审计里这个号凭空消失——这就是「账本没有 QQ 条目却已签到」的矛盾根源；
4. 假 Entries 记录又让 `CheckinRecordTrusted` 通过 → 次轮 default 分支继续判
   already → **自我续命的死循环**（11-19 09:57 起每轮如此）。

**改动（`internal/ideauth/qoder_pool_checkin.go`）**：

- 新增 `qoderClaimedAtStale` / `qoderVerdictIfClaimToday`：探测结论（Owner/
  SamePerson）携带的 claimedAt **确凿陈旧**（能解析且不是今天 UTC+8）⇒ 降为
  Unknown——宁可不判，绝不拿旧账当今日已签。claimedAt 缺失不推翻（上游奖励
  记录都带领取时刻，实测），已接入 pool 的 candidates/snapCands 两处探测与
  扫描侧 path B；
- 扫描侧 path A 与 pool 的 already 归属探测、`qoderApplyCheckinSideEffects`：
  `MarkClaimedAt` 一律只认**今天**的领取时刻（防 Claims 回拨）；
- `qoderClaimRecordOf`：账本 Date/At 必须落在**今天**（记的是「今天这次尝试」），
  陈旧 claimedAt 不得改写账本日期、不得把当条尝试并进历史日期；
- 「额度归属是同人的另一个号」的 already：保持 already 相位（与扫描侧口径一致）
  但 `checked_in_today=false`，且副作用层**跳过 MarkCheckedIn**（`qoderItemFromResult`
  尊重 `res.CheckedInToday`、`qoderFillAlreadyQuotaOwner` 补 false）——本号没签上
  就绝不能在面板上显示假「已签到」；
- 数据修复：清除 QQ 的假 Entries 记录（Claims 里 11-18 的真实领取保留）。

**测试**：新增契约回归 `TestQoderStaleSnapshotCidNotAlready`（陈旧 cid + CLAIMED +
昨天 claimedAt ⇒ 必须保持 pending、不写任何签到记录、账本条目落在今天）；
全套离线测试通过；live 实测（真实上游）：两号回到待签名单、Already=0、
无任何记录写入、账本当天条目齐全。

---

## 2026-11-22（v1.15.12-panel）

### 修复三连：豆包不能签到 / Qoder 只签到一个 / 一键签到非常慢

**① 豆包不能签到 → 根因是出口地区（实测定案）**

两个号都签不上时抓包看到：活动页显示「该页面暂时不可用」、
`POST /alice/ug/campaign/v3/detail/` 返回 `campaign not found`、`landing/get` 返回
710010702 system error。用 `myip.ipip.net` 预检发现本机出口是
`107.149.161.250（美国 加利福尼亚州 圣何塞）` —— 豆包签到活动**仅对中国大陆开放**，
海外出口下页面必死，跟 Cookie / UA 都无关（UA 仍随客户端升到 2.35.12）。

改动（`internal/panel/ide_doubao.go`）：

- 新增 `doubaoProbeEgress()`（2 分钟缓存）：出口非中国大陆时直接返回新相
  `region_blocked`，说明当前出口 IP/地区与改法（切回中国大陆出口），**不再为每个号
  白启动一次无头浏览器干等 60 秒**；
- 页面扫描循环里检测「该页面暂时不可用」→ 立即退出并附出口信息，不等满 60 秒
  （预检失败时的兜底路径）；
- 多账号签到/状态核实从逐号串行改为**并发 2**（每号独立 goroutine + 独立浏览器
  + 独立调试端口）；去掉全局 `doubaoCDPMu` 锁，新增 `doubaoPortMu` 只串行化
  「选端口 → 启动 Chrome → 端口绑定」这一小段（freePort 是 TOCTOU，两个并发会话
  可能挑中同一端口）。

**② Qoder「只能签到一个」→ 同一台物理设备每天只发一份额度，如实点名占用者**

实测（11-19 起）：QQ 号当天真领到（本地可信记录 17:59），gmail 号从 11-19 起连续
pending——上游对**后来账号隐藏当天活动**（不报错、不 401、不回 SAME_PERSON），
面板却只显示「待核实」，后台还每 30 分钟空转重试。环境码 3/0、real/derived 全试过
都绕不过——这不是身份问题，是设备配额问题。

改动（`internal/ideauth/qoder_pool_checkin.go`）：

- 新增 `QoderDeviceQuotaHolder(excludeEmail)`：双方都不走独立设备/出口、且同设备
  另一个号今天有可信签到记录 → 返回占用者；
- `QoderPoolScan` 的 pending 分支、`QoderCheckinPool` 的 default 与 `QoderVerdictOther`
  分支都接入：quota_taken_by 点名 + 文案明确「同设备」而不是「同一个人」；
- `QoderPendingEmails` 收紧：`checked_in_today=true` 或 `quota_taken_by` 非空的号
  不再进入重试名单（重试永远不会成功，只是在烧上游请求）；
- 前端文案从「同一人已领」统一改为「额度已被占用」，strategy 文案改为
  「额度已被占用（同人/同设备）」——覆盖两种占用来源，不再对设备占用误称
  「服务端判定同人」。

**③ 一键签到慢 → 上面两条合起来就是答案**

豆包侧：地区预检秒回（0 个浏览器实例）+ 多号并发 2 + 页面不可用快速退出，
一键签到不再被豆包拖满 60 秒 × 账号数。Qoder 侧：重试名单清空后，30 分钟一轮的
上游探测也停了。

**测试**：`TestQoderDifferentPersonNotBlocked` / `TestQoderPersonLinkClearedWhenProbeSaysOther`
按新契约更新——探测明确「不是同一个人」时仍不写同人 blocked，但同设备占用必须
点名（文案必须出现「同设备」、不得出现「同一个人」）。

---

## 2026-11-17（v1.15.11-panel）

### 新增：国服 `skill_share`（分享一个技能给好友）从「需人工」变为**可一键完成**

**旧结论（原文保留）**：面板把 `skill_share` 归到 manual 桶（无自动动作），README 记的是
「18 个成长任务里 17 个可自动完成，剩余展示操作指引」。

**⚠️ 2026-11-17 实测更正**：`skill_share`（+50c +5e）**可以纯 API 完成**，判据链四步：

1. **accept 报名**：不 accept 时上游 `target=0`，行为事件不计分；accept 后 `target` 变 1
   （面板此前把它当 manual，连报名都没做，目标数一直是 0，看起来像「不可做」）。
2. **取真实市场技能**：`POST /portal/operation-platform/market/skill/list`
   （GET 该路径返回 404 page not found，**必须 POST**）；编造的 `skillId` 不计数。
3. **服务端创建分享链接**：`POST /api/v2/open-platform/asset-shares/c2c/channels`，
   返回 `channel_id` / `share_code`。
   ⚠️ 请求体里的 `asset_ref` 必须是 **JSON 字符串**且键名为 `asset_id`：
   写成对象 `{"assetId":…}` 或键名用 `assetId`，上游一律拒为「请求参数不合法」（code 100001）。
4. **上报判据事件**：`skill_share_channel_action`，且只认 **`action=wechat` +
   `panelAction=success`**（`panelAction=click` 实测不计分），`id` 填第 3 步的 `channel_id`。

**实测结果**：两个国服账号（畈子喜欢吃 / 寻）`skill_share` 均从 `0/1` → `1/1 completed`，
面板自动领奖 **各 +50 分 +5 能**（新二进制实测：`ok=true, credit=50, energy=5, claimed=true`）。

**改动**：

- `internal/upstream/market_skill.go`（新）：`MarketSkillList` / `CreateSkillShareChannel` /
  `DesktopSkillShareChannelAction`，文件头记录上述实测口径与踩坑。
- `internal/upstream/device.go`（新）：`RealDeviceIDs()` 读本机客户端真实 `machineId` / `qimei36`，
  判据事件额外用真实设备指纹再报一份（任务 target=1，重复上报不重复计分）。
- `internal/upstream/market_skill_test.go`（新）：锁死 `asset_ref` 形态与 `wechat/success` 口径。
- `internal/panel/skillshare.go`（新）+ `autotask.go` 注册 `skill_share` 自动动作。
- 扫描待办：国服 `skill_share` 从 manual 桶迁到可自动执行桶（实测扫描 manual_count 9 → 7）。

**剩余人工项**：国际版 5 项 checklist（需在 WorkBuddy 客户端完成，见 v1.15.10-panel 条目）
+ 国服 `wb_wechat_oa_subscribe_task`（需在微信关注公众号并满 24 小时，服务端按真实关注状态判定）。

---

## 2026-11-17（v1.15.10-panel）

### 修复：国际版成长路径「上报了但没反应」——事件通道曾被上游静默丢弃

**旧结论（2026-10-15，原文保留）**：「WorkBuddy 国际版：4 项动作均上报成功（`status=done`），
但任务进度未变化，仍未点亮。国际版 checklist 与国内口径判据不同，网关代做不了。」

**⚠️ 2026-11-17 实测更正（推翻上面结论里的「上报成功」前提）**：

- 旧结论把「HTTP 200」当成了「事件被上游接收」，但**国际版 report 端点对事件有强校验**：
  缺 `reportDelay` 或 `timestamp` 直接 **400 丢弃**：
  `{"code":10001,"msg":"10001:event missing both reportDelay and timestamp fields"}`。
  面板的桌面对话事件链（5.5.6 schema）恰好两个字段都没有 —— 所谓「上报成功」其实是
  **从未被接收**。国服（codebuddy.cn）对此宽容，所以旧版本只在国服实测过，国际版全被丢弃。
- 已修：`upstream.desktopFingerprint` 补齐 `reportDelay`（语义＝上报时刻 − 事件时刻，
  当场立即上报即为 0）与 `timestamp`；新增回归测试锁死这两个字段。
- 事件口径改为**客户端原样**（从国际版客户端 `app.asar` 实测逆向得到）：
  - checklist 的**任务码本身就是客户端 eventCode**：`skill_installed` / `expert_summoned` /
    `template_used` / `wechat_linked`；
  - 必须带业务字段，否则不计分：`skill_installed→{name,skillId,skillVersion}`、
    `expert_summoned→{id,name,expertTitle,type,version}`、
    `template_used→{template_id,task_mode}`（注意是**下划线**）、
    `chat_response_complete→{conversationId,requestId,artifactType}`；
  - **双通道**：客户端同时打 `/v2/report`（标准事件，msg=OK）与 `/llm/data/report`
    （模型事件，msg=success）。面板新增 `ReportDesktopEventAll`，global 账号走双通道，
    国服保持单通道（国服没有该端点，硬打会 404 噪音）。
- **实测现状（诚实结论）**：按上述客户端口径上报后，两条通道**均返回成功**，但
  checklist 的 `status` 仍为 `available`，`/v2/activity/growth/profile` 仍
  `completed=0 / total=5 / first_visit=true` —— 服务端对国际版 checklist 的计分**另有判据**，
  纯 API 点不亮。国际版**也没有** accept / claim / complete 端点（全部 404 或
  `400 invalid request`），即完成判定纯在服务端。
- 故面板**不再对国际版 checklist 承诺「一键完成」**：新增 `client` 分类（与 `auto` 区分），
  扫描结果里标「需在 WorkBuddy 客户端完成」并**逐项给出操作指引**；用户在客户端做完，
  面板回读自动核销。执行结果一律 `unverified`（**绝不谎报 done**），并写明原因
  （「已上报未计分」/「未取到真实业务 ID」）。

### 修复：国际版对话按 realm 分流 + 桌面对话事件头统一

- `upstream.DesktopChatWithExpert` 按 `a.IsGlobal()` 分流：国际版用 `default-model` +
  英文提示词；出站头改走 `ChatHeaders`（统一会话头族），`X-IDE-Version` 取实际客户端版本，
  不再硬编码 `5.5.6`。

### 其他

- 签到池操作列 / 表头对齐修复，弹窗与设置页冗余文案精简（`internal/panel/app.js`、
  `internal/panel/index.html`）。

## 2026-11-17（v1.15.9-panel）

### 新增：Google Gemini 官方 API Key 通道

- 新增 `internal/gemini`：把 **Google 官方 API Key** 接入网关，模型名带 `gemini:` 前缀
  （例如 `gemini:gemini-3-flash-preview`），同一套 `/v1/chat/completions` 端点直接调用。
- 上游 `generativelanguage.googleapis.com/v1beta/openai` **本身就是 OpenAI 兼容**，故只做
  「改写 model + 带 key 转发 + 流式透传」，不做协议翻译。
- 密钥文件 `data/gemini_api_keys.json`（不入库）：按 `groups` 分组，每组一个模型前缀，
  便于把「订阅」与「按量」彻底隔离；组内多 key **轮换**，401/402/429 自动换下一把。
- 分流位置在 `chatCompletions` 的 **realm 判断与选号之前**，按原始 model 字符串判定，
  不会误伤 `cn:` / `global:` / `trae:` / `minimax:` 等既有前缀。

### 出站代理：Gemini 必须走代理（实测更正）

- ⚠️ **2026-11-17 实测更正**：本通道最初实现为「显式直连（`Proxy: nil`）」，理由是
  「本机系统代理常拒连」。**该结论方向反了**，实测数据：
  - 直连 `generativelanguage.googleapis.com` → **25s 超时**；
  - 走 `http://127.0.0.1:10808` → **683ms 返回 HTTP 200**。
  ⇒ Gemini 通道**必须走代理**，"直连"才是会挂掉的那条路。
- 代理地址解析优先级：**配置项 `proxy_url` → 环境变量 `HTTPS_PROXY`/`HTTP_PROXY` →
  Windows 注册表系统代理**（`ProxyEnable=1` + `ProxyServer`）。
- 刻意**不用** Go 的 `http.ProxyFromEnvironment`：它只读环境变量、**不读注册表**，
  而服务由计划任务拉起时环境变量常被清空，结果就是静默直连卡死（旧版 502 的真正成因）。
- ⚠️ 旧实现里「默认客户端必须直连」的单元测试已随之作废，原文保留在
  `internal/gemini/gemini_test.go` 并就地标注更正。

### 实测结论（2026-11-17，真打上游）

- `/v1/models` 出现 **64 个** `gemini:` 条目；对话实测 HTTP 200：
  `gemini-3-flash-preview` / `gemini-3.5-flash` / `gemini-3.1-flash-lite` 均返回真实内容；
- 流式 `stream:true` → `Content-Type: text/event-stream`，SSE 帧与 `[DONE]` 正常；
- 免费层配额边界：Pro 系（`gemini-3.1-pro-preview` 等）返回 **429**（额度为 0，约 5h 重置）；
  `gemini-2.5-flash` / `gemini-2.5-pro` 已被 Google 对新用户下线（**404**）。

### 文档

- README 新增「🤖 Google Gemini 官方 API Key 通道」专节（配置 / 出站代理 / 调用 / 已知边界），
  并同步平台覆盖表、核心能力表、架构图与「关键断言 ↔ 代码出处」。

## 2026-11-15（v1.15.9-panel，同版本续记）

### Qoder 设备身份：自动分配，不用手动逐号配

- 新增 `device_mode=auto`（**默认值**）：号池里每个账号自动带一条出口配置，**用户零操作**；
  浏览器 profile 目录也自动生成（按邮箱隔离的短哈希），不再手填。
- **auto 展开为「真实设备」** —— 2026-11-15 实测更正：auto 原先是「号池首个号 real、其余
  按邮箱派生」，但派生身份会被上游按设备维度过滤掉每日额度（实测：某账号生效身份为派生时
  面板 `pending` 领不到，同一账号改成真实设备后**立刻 +100 Credits**）。
  ⇒ **能领到 > 身份不同**；派生模式降级为手动高级选项，面板上标注了风险。
- 存量迁移：老版本把每个号都写成 `real`；判定「用户从未手动改过」（没配代理 + profile 目录
  就是自动生成的那个）的条目自动迁移到 `auto`，配了代理/改过目录名的保持不动。
- 新增 `POST /panel/api/ide/qoder/egress/reset-auto`（一键恢复默认设备身份，代理保留）
  + 面板「全部恢复自动」按钮 + 账号卡片上的「真实设备 / 派生设备 / 自动」标签。

### Qoder pending 判定：三层证据

1. 客户端身份能看到今天的额度 ⇒ 「客户端能看到、面板查不到 ⇒ 设备身份不一致」；
2. **真实设备身份**能看到、当前身份看不到 ⇒ 「上游按当前这套设备身份过滤掉了本号」；
3. 两套环境码 + 真实身份都看不到 ⇒ 「上游确实没给本号下发」。

第 2 层是新增的：派生/手动配置出问题时会立刻说清原因，而不是让用户去猜。

### 其它

- 领养任务：前置上报失败不再当日封印（一次网络抖动不该把 +300 Credits 推迟一整天）。
- 面板 API 缓存：handler panic 不再让缓存条目永久挂死（后续请求全部卡住、只能重启进程）。
- Qoder 探测不再触发 refresh（避免轮换并烧掉号池里的 refresh_token）。
- **文档脱敏**：README / CHANGELOG / 分发说明 / 代码注释里的真实邮箱一律改为不可逆短码
  （`账号 45da34` / `账号 569875`）。

## 2026-11-11

### Qoder 每日额度

- 修复「已下发被误判未下发」：面板不再把 runtime-info 环境码固定为国际版 3，改为按账号自动尝试 3 / 0；哪个环境能看到当天 `CLAIM_BENEFIT`，就用哪个环境领取。
- 每个账号确认可用的环境码写入 `qoder_accounts.json` 的 `runtime_env`，后续扫描与签到直接复用，不再重复探测。
- `runtime-info.exe` 结果增加 10 分钟缓存，减少每次扫描重复启动进程（单次约 3-4 秒）。
- 签到成功日志改为 `<邮箱> 邮箱实测成功：+N Credits（环境码、活动、领取时刻）`，直接写明是哪一个邮箱，不用含糊指代。

### 积分构成

- Trae / Qoder 积分构成改为与 WorkBuddy 一致的条形卡片：大号剩余额度、构成条、图例。
- Trae 同类积分包合并为一段（如「签到奖励 x29」），避免几十个日包把图例撑爆。
- 签到奖励、最近到期 / 不过期信息以文字展示在卡片底部；Qoder 显示 `+100 Credits/日` 与上游到期时间（9999-12-31 视为不过期）。
- Qoder 不再只显示当前登录号，改为号池逐号查询并展示。

### 核账记录

- 两个 Qoder 账号 2026-11-11 均从上游拿到独立 grant：
  - 账号 45da34：`01a1c862-5fec-750c-9f58-6a3802f5d057`，+100 Credits，10:14:12。
  - 账号 569875：`01a1caf0-f1fe-76b5-9ea0-38b3dacb2cf0`，+100 Credits，22:09:10。
- 附加包额度：账号 45da34 2100 / 已用 0 / 剩余 2100；账号 569875 1500 / 已用 558 / 剩余 942。
