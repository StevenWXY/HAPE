# Clipli 外部合作 API

本文档面向接入 Clipli 的外部资产平台、版权平台和发行平台。Clipli 服务已部署，合作方只需使用部署方提供的 API 域名、凭据和版本配置；本文档中的路径均相对于该部署根地址。

## 1. 通用约定

- 所有路径均相对于部署方提供的 API 根地址。
- 稳定接口使用 `/api/v1` 前缀；未带版本的 `/api/*` 路径属于平台业务兼容接口。
- 生产环境使用 HTTPS。认证方式、允许来源、限流策略和环境域名由部署方在合作接入时提供。
- 外部平台凭据只能由服务端保存和使用，不得下发到浏览器或写入作品、资产响应。
- JSON 成功响应：`{ "data": ... }`
- JSON 错误响应：`{ "error": { "code": "...", "message": "..." } }`
- 写入接口的 `requestId` 必须为 8 至 100 个字符。相同 `requestId` 重试时返回首次结果，不重复核销、扣款或创建记录。
- 时间使用 UTC；完整时间优先使用 RFC 3339，兼容客户端字段可保留 `YYYY-MM-DD HH:mm`。
- 金额、额度和数量使用非负整数；费率和公式以响应中的规则版本及部署方正式公告为准。
- 外部平台应将资产、授权和事件视为事实来源；Clipli 返回的 `external` 字段用于标识来源、同步状态和快照时间。

## 2. HAPW 资产 API（新客户端）

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `GET` | `/api/v1/hapw/summary` | HAPW 总量、可核销/已核销/受限、可行权、可兑换、总估值和潜在额度汇总 |
| `GET` | `/api/v1/hapw/assets` | HAPW 资产列表、搜索、筛选、排序和分页 |
| `GET` | `/api/v1/hapw/assets/{assetId}` | 单枚 HAPW 完整详情、关联作品和可执行动作 |
| `GET` | `/api/v1/hapw/assets/{assetId}/authorization` | 权利持有方、授权范围、地域、使用方式、期限与证书来源 |
| `GET` | `/api/v1/hapw/assets/{assetId}/history` | 发行、核销、生成、行权和兑换的统一时间线 |
| `GET` | `/api/v1/hapw/assets/{assetId}/works` | 使用该 HAPW 素材生成或关联的作品 |
| `GET` | `/api/v1/hapw/assets/{assetId}/exchange-quote` | CLIP 兑换该 HAPW 的本金、5% 手续费与总额试算 |
| `GET` | `/api/v1/hapw/redemptions` | 核销记录；可用 `assetId` 筛选 |
| `GET` | `/api/v1/hapw/redemptions/{redemptionId}` | 单条核销凭证和剩余额度 |
| `POST` | `/api/v1/hapw/redemptions` | 不可逆核销 HAPW，并一次性发放创作额度和 CLIP |
| `GET` | `/api/v1/hapw/exercises` | 资产行权记录；可用 `assetId` 筛选 |
| `POST` | `/api/v1/hapw/exercises` | 向第三方平台创建 HAPW 行权签名请求 |
| `GET` | `/api/v1/hapw/exchanges` | CLIP→HAPW 储备兑换记录；可用 `assetId` 筛选 |
| `POST` | `/api/v1/hapw/exchanges` | 支付 CLIP 本金和手续费领取有限储备 HAPW |
| `GET` | `/api/v1/hapw/exchange-policy` | 每日 2 枚上限、已用额度、重置时间和储备库存 |
| `GET` | `/api/v1/platforms` | 可行权的第三方平台与官方链接 |

### 2.1 资产列表参数

`GET /api/v1/hapw/assets` 支持：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `search` | string | 匹配资产 ID、Token ID、名称或权利持有方 |
| `status` | string | `available`、`redeemed` 或 `restricted` |
| `owner` | string | 精确匹配当前持有者 |
| `rightsHolder` | string | 模糊匹配版权持有方 |
| `redeemable` | boolean | 是否可核销 |
| `exchangeAvailable` | boolean | 是否存在储备兑换份额 |
| `page` | integer | 从 1 开始，默认 1 |
| `pageSize` | integer | 默认 20，最大 100 |
| `sort` | string | `tokenId`、`value`、`-value`、`issuedAt`、`-issuedAt` |

资产详情同时保留网页所需的扁平字段，并提供四组结构化数据：

- `authorization`：版权持有方、授权范围、地域、使用类型、商用/衍生权、是否允许 AI 训练、是否可转授权和有效期。
- `provenance`：发行方、证书编号、网络、标准、合约/元数据地址、验证状态和发行时间。
- `media`：关联作品 ID、素材格式和预览地址。
- `external`：外部平台代码、平台资产 ID、官方资产地址、同步状态、同步时间和数据版本。该块是资产来源证明，不代表 Clipli 重新发行资产。

### 2.2 核销

```http
POST /api/v1/hapw/redemptions
Content-Type: application/json

{
  "requestId": "redeem-asset-2048-001",
  "assetId": "asset-2048",
  "accepted": true
}
```

核销要求资产状态为 `available`。成功后 HAPW 变为 `redeemed`，关闭行权和储备兑换，同时按 `额度 × 0.4` 向下取整发放 CLIP。

### 2.3 资产行权

```http
POST /api/v1/hapw/exercises
Content-Type: application/json

{
  "requestId": "exercise-asset-771-001",
  "assetId": "asset-771",
  "platformCode": "opensea",
  "accepted": true
}
```

资产行权保留 HAPW 未核销状态，只创建目标第三方平台的待签名请求，不发放创作额度或 CLIP。

### 2.4 CLIP 兑换 HAPW

```http
POST /api/v1/hapw/exchanges
Content-Type: application/json

{
  "requestId": "exchange-asset-528-001",
  "assetId": "asset-528",
  "accepted": true
}
```

总扣款公式为 `total = price + ceil(price × 5%)`。每个调用账户按 UTC 自然日最多成功兑换 2 枚，且每个储备份额只能领取一次；具体限额以 `GET /api/v1/hapw/exchange-policy` 返回值为准。

## 3. 平台业务 API（可选）

以下接口用于接入 Clipli 的创作、账户和资产操作流程，不属于外部资产源的最小接入要求。是否开放以及所需权限由部署方配置。

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `POST` | `/api/generations` | 使用已核销 HAPW 同时扣除创作额度和 CLIP，创建 AI 视频任务 |
| `POST` | `/api/conversions` | 为未核销 HAPW 创建区域与期限授权，扣除 18 CLIP 服务费 |
| `GET` | `/api/profile` | 当前调用方的账户、关联平台和 CLIP 余额 |
| `POST` | `/api/profile/wallet` | 关联受支持的钱包提供方（如部署方开放） |
| `POST` | `/api/profile/overseas` | 绑定第三方平台账号 |
| `DELETE` | `/api/profile/overseas` | 解绑第三方平台账号 |
| `PATCH` | `/api/profile/settings` | 更新签名确认和到期提醒开关 |

生成任务支持 `duration` 为 15、30、60 秒，`quality` 为 `standard` 或 `pro`。额度为 `ceil(duration × qualityFactor)`，CLIP 为 `ceil(额度 × 0.2)`。

## 4. 外部资产源与 CLIP 金库 API

Clipli 的 HAPW 数据以外部平台为事实来源，服务端只做规范化、校验、缓存和操作记录。平台接入方需要提供以下上游接口：

| 上游接口 | 必需字段 | 用途 |
| --- | --- | --- |
| `GET /assets` | `id`、`tokenId`、`name`、`rightsHolder`、`status`、`authorization`、`provenance` | 首次全量或游标增量同步 |
| `GET /assets/{id}` | `id`、`tokenId`、`status`、`owner`、`externalUrl`、`updatedAt` | 详情与状态校准 |
| `GET /assets/{id}/authorization` | `holder`、`scope`、`territories`、`usageTypes`、`validFrom` | 授权边界 |
| `GET /assets/{id}/media` | `linkedWorkIds`、`format` | 素材和作品关联 |
| `GET /assets/{id}/events` | `eventId`、`type`、`status`、`createdAt` | 外部事件时间线 |
| `GET /works?assetId={id}` | `id`、`title`、`linkedAssetId`、`externalUrl` | 作品关联 |
| `POST /webhooks/asset-events`（可选推送通道） | `eventId`、`assetId`、`eventType`、`occurredAt`、`signature` | 外部平台向 Clipli 配置的接收地址推送近实时失效通知；方向、签名头和重试策略在接入时确认 |

对应的 Clipli 只读 API：

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `GET` | `/api/v1/integrations/asset-requirements` | 返回上游接口、字段、超时、大小和失败降级规则 |
| `GET` | `/api/v1/integrations/asset-sources` | 所有外部源的状态、地址、同步间隔和资产数量 |
| `GET` | `/api/v1/integrations/asset-sources/{sourceCode}` | 单个平台源状态 |
| `GET` | `/api/v1/integrations/asset-sync-runs` | 同步批次的读取、校验、保存数量和错误 |

部署方为每个外部来源配置接入地址和凭据。合作方可以提供分页资产列表，也可以提供由网关聚合后的 `{ "data": [...] }` 规范化响应；具体地址、认证头和同步计划在接入登记时确认。服务端默认执行 5 秒超时、2 MiB 响应上限、字段校验和来源标记；同步失败时保留最近一次有效快照，并在同步记录中写入失败原因。

CLIP 不是公开铸币接口：`GET /api/v1/clip/treasury` 返回一次性铸造总量、平台钱包余额、DEX 流动性配置、内部账本余额和守恒结果；`GET /api/v1/clip/distribution-rules` 返回中心化规则；`GET /api/v1/clip/distributions` 返回每次核销的分发流水。成功核销才按 `floor(creditYield × 0.40)` 从金库划拨，生成/授权/储备兑换费用回到金库。

`GET /api/v1/session` 用于获取当前调用上下文与可用操作能力。账户、身份、钱包和持久化策略由部署方配置；合作方不应仅凭钱包地址推断版权、资产或平台账户权限。

## 5. 平台兼容 API（非外部资产源必需）

以下接口用于平台站点和既有客户端兼容，外部资产平台通常无需接入。它们与 `/api/v1` 使用相同的业务状态和规则：

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `GET` | `/api/health` | 服务健康状态 |
| `GET` | `/api/overview` | 首页统计、精选 HAPW 和流程 |
| `GET` | `/api/works` | 作品列表 |
| `GET` | `/api/works/{workId}` | 作品详情 |
| `GET` | `/api/assets` | 资产仪表盘聚合数据 |
| `GET` | `/api/studio` | 生成台所需的 HAPW、凭证、余额和任务 |
| `POST` | `/api/hapw/redemptions` | v1 核销接口兼容别名 |
| `POST` | `/api/exercises` | v1 行权接口兼容别名 |
| `POST` | `/api/transfers` | 旧客户端兼容别名；语义仍为资产行权 |
| `POST` | `/api/clip/hapw-exchanges` | v1 储备兑换兼容别名 |

详细字段、状态码和可执行请求以 [`api/openapi.yaml`](../api/openapi.yaml) 为准。OpenAPI 文件中的 `servers` 地址是部署占位符，实际域名以合作接入信息为准。

## 6. 常见状态码

| HTTP | 错误码示例 | 含义 |
| --- | --- | --- |
| `400` | `invalid_request_id` | 缺少有效幂等键 |
| `400` | `hapw_not_redeemable` | HAPW 已核销、受限或不存在 |
| `400` | `license_required` | 尚未核销对应 HAPW，不能使用素材生成 |
| `400` | `insufficient_generation_credits` | 对应授权凭证额度不足 |
| `400` | `insufficient_clip` | CLIP 余额不足 |
| `400` | `asset_unavailable` | HAPW 当前不能区域授权或行权 |
| `400` | `hapw_reserve_unavailable` | 储备份额不存在或已领取 |
| `503` | `clip_treasury_insufficient` | 平台 CLIP 金库无法满足中心化分发 |
| `404` | `hapw_not_found` | 未找到 HAPW |
| `429` | `hapw_exchange_daily_limit` | 已达到当日 HAPW 兑换上限 |
