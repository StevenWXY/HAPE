# HAPE-V 项目指南

## 1. 产品定位

HAPE-V 是版权素材驱动的 AI 视频创作与资产行权平台。版权持有方把素材包、改编范围和使用边界授权为 HAPW。用户持有并核销对应 HAPW 后，才可以使用素材；核销会同时发放 AI 生成额度与 HAPU，生成视频则同时消耗额度与 HAPU。

当前版本是可操作原型，使用演示资产、模拟视频任务、模拟 DEX 池和内存状态，不处理真实私钥、支付、代币合约或链上签名。作品来源通过适配器隔离，后续可接入第三方作品 API。

## 2. 专有名词

| 名称 | 定义 | 边界 |
| --- | --- | --- |
| HAPE | 平台名称 | 保存授权、核销、生成、费用与行权记录 |
| HAPW | 平台核心版权资产 | 由版权方授权，对应具体素材包和使用范围；持有并核销后才能使用素材 |
| AI 生成额度 | 与 HAPW 授权绑定的计算额度 | 核销时发放，按视频时长和质量消耗，不能脱离授权单独流转 |
| HAPU | HAPE 发行的生成服务代币 | 不设总量；核销 HAPW 时发放，也可通过 DEX 获取；生成视频时与额度共同消耗 |
| HAPW 核销 | 在 HAPE 内开启绑定素材的生成许可 | 不可逆；产生授权凭证、生成额度和 HAPU，并关闭同一资产的行权路径 |
| 资产行权 | 保留 HAPW 未核销状态并创建第三方平台签名请求 | 不产生生成额度或 HAPU，也不等同于底层版权转让或第三方平台接纳 |
| 第三方平台 | HAPE 外部的展示、发行或交易平台 | 当前入口包括海文发、OpenSea、Foundation、SuperRare、Art Blocks |

所有用户可见的“资产转移”统一称为“资产行权”。内部兼容路由可暂时保留 `/transfer`，但新增接口、页面标题、按钮、状态和文档必须使用 exercise / 行权口径。

## 3. 核心闭环

1. 版权方授权素材并发行 HAPW。
2. 用户持有 HAPW，核对版权方与使用范围。
3. 用户不可逆核销 HAPW，同时领取生成额度和 HAPU；也可以保留未核销状态，单独发起资产行权。
4. 用户配置视频时长与质量，系统同时扣除额度和 HAPU。
5. 系统保存视频任务与观察播放数据；播放量不产生 HAPU。
6. 用户可在外部 DEX 以 USDT 获取 HAPU，也可支付公开价格与 5% 手续费兑换平台储备 HAPW；每个账户每日最多兑换 2 枚。
7. 未核销 HAPW 可向选定第三方平台发起资产行权。

### 3.1 演示公式

- 生成额度：`C = ceil(S × Q)`，`S` 为秒数，标准质量 `Q=1.0`，高质量 `Q=1.5`。
- 核销发放：`G = floor(Y × 0.40)`，`Y` 为 HAPW 标注额度，`G` 为一次性发放的 HAPU。
- 生成费用：`U = ceil(C × 0.20)`，一次生成同步扣除 `C` 个额度与 `U` 个 HAPU。
- 储备兑换：`T = P + ceil(P × 5%)`，`P` 为 HAPW 储备价格，`T` 为总 HAPU 消耗。

任一余额不足时不得创建生成任务。每枚 HAPW 只能成功核销一次，每个储备份额只能被领取一次；所有余额变化接口必须支持 `requestId` 幂等。

### 3.2 DEX 演示快照

- HAPU 储备：`250,000 HAPU`
- USDT 储备：`25,000 USDT`
- 参考汇率：`1 USDT ≈ 10 HAPU`，`1 HAPU ≈ 0.10 USDT`
- 双边名义流动性：约 `50,000 USDT`

以上仅用于展示汇率和双边池信息，不是实时行情或正式成交承诺。正式交易以外部 DEX 和钱包报价为准。

## 4. 页面与功能

| 页面 | 目标 | 主要交互 |
| --- | --- | --- |
| 首页 | 解释 HAPW、生成额度与 HAPU 的闭环 | 进入作品、生成台、机制页；查看图表和边界 |
| 作品 | 浏览演示作品 | 分类、播放预览、进入详情、六语切换 |
| 作品详情 | 查看内容、版权素材和外部入口 | 查看关联 HAPW；选择海文发、OpenSea 等第三方平台 |
| 资产仪表盘 | 核对资产、额度和操作记录 | 查看 HAPW、额度、HAPU、生成记录、行权记录、DEX、HAPU→HAPW |
| AI 生成台 | 核销并生成视频 | 同时领取额度与 HAPU；预估并扣除额度与 HAPU；查看生成记录 |
| HAPU 机制 | 披露公式、流动性与边界 | 查看四步流程、公式、DEX 快照、储备兑换说明 |
| 区域授权 | 创建指定区域和期限的发行许可 | 选择 HAPW、区域和期限，以 HAPU 支付演示费用 |
| 资产行权 | 向第三方平台创建行权请求 | 选择平台、HAPW、确认资格、请求钱包签名 |
| 关于 | 说明角色、边界和常见问题 | 阅读业务说明及 8 项 Q&A |
| 钱包/账号/安全 | 演示外围账户状态 | 连接钱包、绑定海文发账号、设置签名和提醒 |

## 5. MVP 必须实现

- 响应式单页应用，支持中文、英语、西班牙语、日语、法语、韩语。
- 暗色和亮色模式，首次跟随系统，选择通过 `localStorage` 保持。
- 哈希路由：`/`、`/works`、`/work/:id`、`/assets`、`/studio`、`/hapu`、`/convert`、`/transfer`、`/bind`、`/wallet`、`/security`、`/about`。
- 作品列表、详情、预览与第三方平台选择弹窗；外部链接使用新窗口并加 `noopener noreferrer`。
- HAPW 条目显示版权方、授权范围、核销状态、额度产出、HAPU 发放量和储备价格。
- 核销后同步产生唯一授权凭证、额度余额、HAPU 流水；资产不可再次核销或行权。
- 生成支持 15/30/60 秒和标准 720p、高质量 1080p；前后端使用相同公式同步扣除额度与 HAPU。
- 播放数据只作为观察指标，页面、接口和文档不得描述为 HAPU 收益来源。
- HAPU 资产卡显示余额、参考汇率、HAPU/USDT 双边储备和名义流动性。
- HAPU→HAPW 储备兑换显示本金、5% 手续费、总消耗、每日 2 枚上限和一次性领取状态。
- 资产行权支持海文发、OpenSea、Foundation、SuperRare、Art Blocks；提交记录目标平台和签名状态。
- 关于页包含机制、权利、DEX、储备兑换、第三方平台和非托管边界 Q&A。
- 统一 `{ data }` 成功响应、`{ error: { code, message } }` 错误响应和本地化错误提示。

## 6. 暂不实现

- 真实链上交易、HAPU 合约发现、自动做市、私钥或助记词处理、KYC、真实支付。
- 真实视频模型调用、转码、内容审核、发布和失败退款。
- 复杂治理、质押、动态收益、播放挖矿、固定回购或法币兑付。
- 在正式网络、合约地址和流动性池未披露前，不向 DEX 链接自动注入代币参数。

## 7. 核心实体

```text
Work              { id, title*, category*, duration, views, creator*, summary*, description*, format*, linkedAssetId, source }
HAPWAsset          { id, tokenId, owner, rightsHolder*, authorizationScope*, creditYield, hapuPrice, exchangeAvailable, redemptionStatus }
HAPWRedemption     { id, requestId, assetId, receipt, creditsGranted, creditsRemaining, hapuGranted, status, createdAt }
GenerationAccount  { balance, lifetimeGranted, lifetimeUsed, hapuGrantPerCredit, hapuCostPerCredit, lifetimeHapuGranted, lifetimeHapuSpent }
AIVideoGeneration  { id, requestId, assetId, title*, duration, quality, creditsUsed, hapuCost, validViews, status, createdAt }
HAPUAccount        { balance, quoteAsset, dexUrl, dexPool, hapwExchangeFeeRate }
HAPUTransaction    { id, typeCode, amount, counterparty*, statusCode, txHash, createdAt }
HAPWExchange       { id, requestId, assetId, price, fee, total, feeRate, statusCode, createdAt }
HAPWExchangePolicy { date, timezone, resetsAt, dailyLimit, usedToday, remainingToday, inventoryTotal, inventoryAvailable, reached }
AssetExercise      { id, requestId, assetId, platformCode, direction*, value, statusCode, createdAt }
ExternalPlatform   { code, name, url }
```

带 `*` 的展示字段可按 `En/Es/Ja/Fr/Ko` 后缀扩展；缺少作品专名翻译时允许回退英文，产品机制与操作文案不得回退。

## 8. API 契约

### 读取

- `GET /api/overview`：首页统计与闭环概览。
- `GET /api/works`、`GET /api/works/:id`：作品列表与详情。
- `GET /api/assets`：HAPW、行权记录、第三方平台、HAPU、DEX 池、生成和储备兑换；`hapu.hapwExchangePolicy` 返回当日限额与可用库存快照。
- `GET /api/studio`：素材、凭证、额度、HAPU 和生成记录。
- `GET /api/profile`：钱包、外部账号和安全设置。

### 写入

- `POST /api/hapw/redemptions`：`{ requestId, assetId, accepted }`，核销并同步发放额度与 HAPU。
- `POST /api/generations`：`{ requestId, assetId, duration, quality, accepted }`，同步扣除额度与 HAPU。
- `POST /api/hapu/hapw-exchanges`：`{ requestId, assetId, accepted }`，支付价格和 5% 手续费领取储备 HAPW；服务端同时校验每日 2 枚上限与库存。
- `POST /api/exercises`：`{ requestId, assetId, platformCode, accepted }`，创建资产行权签名请求。
- `POST /api/transfers`：仅作为旧客户端兼容别名，输入仍按行权契约处理。
- `POST /api/conversions`：区域授权并扣除演示 HAPU 服务费。
- 钱包、账号和安全设置接口保持现有契约。

核销、生成、储备兑换、区域授权和资产行权都必须携带 8 至 100 位 `requestId`。重复提交返回第一次结果，不重复发放、扣费或创建记录。

## 9. 架构约束

- 前端无构建依赖：`public/index.html`、`public/styles.css`、`public/locales.js`、`public/app.js`。
- 服务端使用 Node.js 内置 `http`；机制公式集中在 `server/mechanism.js` 并由单元测试覆盖。
- 作品来源经 `server/work-source.js` 和 `server/adapters/works/` 归一化，前端不直接信任第三方字段或 URL。
- HAPW、授权凭证、生成额度、HAPU、生成任务、资产行权必须为独立实体，以标识互相引用。
- 外部链接只允许 `http:` 与 `https:`，服务端静态响应保留 CSP、禁止嵌入和内容类型保护头。

## 10. 视觉与交互

- 简约等距风格，深炭黑或纸白底，珊瑚橙、青绿和少量金色；禁止蓝紫主色、玻璃拟态和装饰性渐变光球。
- 页面使用实色表面、细边框、方正按钮和克制动画；尊重 `prefers-reduced-motion`。
- 顶部导航宽度稳定，页面切换不得横向位移。
- 公式、汇率、流动性、手续费与演示声明必须同时可见，不能只在成功提示中披露。
- 弹窗具备对话框语义、关闭按钮、遮罩关闭和 Escape 关闭；表单具备焦点、禁用态和本地化错误反馈。

## 11. 验收清单

1. 六种语言下首页、作品详情、资产、生成台、机制、行权、关于页无旧版播放收益或资产转移口径。
2. 核销 150 额度 HAPW 同步增加 150 额度和 60 HAPU；重复提交不重复增加。
3. 15 秒标准生成扣除 15 额度与 3 HAPU；15 秒高质量扣除 23 额度与 5 HAPU。
4. 任一余额不足、无对应凭证或未确认条款时生成失败且无部分扣款。
5. DEX 卡显示 `1 USDT ≈ 10 HAPU`、两边储备和 `50,000 USDT` 名义流动性，并标明演示快照。
6. HAPU→HAPW 显示 5% 手续费；完成后扣除总额、生成流水并禁用同一储备份额。
7. 作品详情平台选择包含五个平台和官方链接，外部打开方式安全。
8. 资产行权只接受已配置平台与可操作 HAPW，记录目标平台并等待签名。
9. 关于页 8 项 Q&A 在六种语言下完整可用。
10. 390px 移动端无横向溢出，暗色和亮色模式均具备足够对比度。
11. `npm test` 覆盖额度公式、HAPU 发放、生成费用、5% 报价、DEX 池和作品适配器。

## 12. 后续扩展

- 接入第三方作品 API，增加超时、缓存、字段校验、版权状态同步和演示数据降级。
- 接入真实 AI 视频模型，拆分排队、生成、审核、发布、失败退款状态。
- 配置正式 HAPU 网络、合约、精度、可信池和链上储备 HAPW 后，再开放真实兑换。
- 为第三方平台建立独立适配器，保存平台资产标识、上架状态、费用和下架回执。
