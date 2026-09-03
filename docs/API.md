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

## 3. 钱包资产与 CLIP 空投 API

钱包接口采用浏览器钱包的 EIP-1193 标准。连接接口只登记用户选择的钱包地址、网络和提供方；Clipli 不接收私钥，也不在浏览器内保存平台空投密钥。钱包资产是外部资产平台同步后的只读快照，实际生产接入必须由来源平台提供持有人地址或按地址查询能力。

### 3.1 连接钱包

```http
POST /api/v1/wallet/connect
Content-Type: application/json

{
  "provider": "MetaMask",
  "address": "0x1111111111111111111111111111111111111111",
  "chainId": "0x61"
}
```

`provider` 支持 `MetaMask`、`Coinbase Wallet`、`WalletConnect` 和 `Venly`。`address` 必须是 EVM 地址，`chainId` 只接受 BNB Smart Chain 测试网 `0x61`（97）或主网 `0x38`（56）。旧客户端仍可使用 `POST /api/profile/wallet`；断开连接使用 `DELETE /api/v1/wallet/connect`（兼容路径为 `DELETE /api/profile/wallet`）。生产环境若需要证明签名归属，应在业务层增加 challenge / `personal_sign` 或 EIP-4361 校验；仅凭地址不能证明控制权。

### 3.2 查询钱包内关联资产

```http
GET /api/v1/wallet/assets?address=0x1111111111111111111111111111111111111111
```

返回 `walletAddress`、`items` 和 `totalItems`。每个 `WalletAsset` 包含 `assetId`、`tokenId`、余额、HAPW 标准、来源平台代码、来源资产 ID、同步状态、`canRedeem` 和 `canExercise`。响应中的 `sourceOfTruth` 固定为 `external-platform`。接口不会把未在外部平台快照中确认的资产当作持仓。

外部资产源应满足以下任一方式：

- `GET /assets?owner={walletAddress}` 返回该地址的资产；或
- 全量 `GET /assets` 的每条资产返回规范化 `owner` 字段，Clipli 再按地址过滤。

### 3.3 空投规则与资格

```http
GET /api/v1/airdrop-rules
GET /api/v1/wallet/airdrop-eligibility?ruleCode=hapw_redemption&address=0x1111111111111111111111111111111111111111&assetId=asset-2048
```

内置规则如下：

| 规则 | 触发条件 | 数量 | 资金来源 |
| --- | --- | --- | --- |
| `hapw_redemption` | 指定 HAPW 核销成功，且钱包已连接 | `floor(creditYield × 0.40)` CLIP | 该次核销已登记的中心化分发额度，不重复扣金库 |
| `admin_approved` | 管理员批准指定钱包与数量 | 请求中的 `amount` CLIP | 从 CLIP 平台金库预留，失败时返还 |

资格接口会返回 `eligible`、`reason`、预计 `amount` 以及已创建的 `existingId`。核销时若钱包已连接，平台自动创建空投任务；若核销时未连接，管理员可在用户连接并完成资格确认后使用 `admin_approved` 或按内部流程补建任务。

### 3.4 空投创建、执行与查询

#### BNB 网络测试边界

CLIP 的目标发行网络为 BNB Smart Chain。当前尚未部署正式 BEP-20 合约，也未接入 BNB RPC 或平台钱包签名服务，因此所有空投先限定在两个 BNB 网络配置中：测试网 `0x61`（97）和主网 `0x38`（56）。两者都返回 `simulationOnly: true`，主网选项也不会产生真实主网交易。

运营控制台可以用模拟执行验证以下闭环：校验 EVM 地址 → 创建空投任务 → 预留 CLIP 金库额度 → 确认任务 → 返回可追踪的模拟回执。模拟回执使用 `sim-bnb-...` 前缀，故意不采用 `0x` 交易哈希格式；它只代表平台账本状态已确认，不代表链上交易已经打包。

读取网络配置：

```http
GET /api/v1/admin/bnb-networks
X-Clipli-Admin-Key: <operator-key>
```

创建任务时使用 BNB Chain ID：

用户和合作方可读取：

```http
GET /api/v1/airdrops?address=0x1111111111111111111111111111111111111111&status=queued
```

平台运营方使用 `X-Clipli-Admin-Key` 调用：

```http
POST /api/v1/admin/airdrops
X-Clipli-Admin-Key: <operator-key>
Content-Type: application/json

{
  "requestId": "airdrop-wallet-2048-001",
  "ruleCode": "admin_approved",
  "walletAddress": "0x2222222222222222222222222222222222222222",
  "chainId": "0x61",
  "amount": 25,
  "token": "CLIP",
  "assetId": "asset-2048"
}
```

尚未部署合约时，可由运营控制台调用模拟执行：

```http
POST /api/v1/admin/airdrops/{airdropId}/simulate
X-Clipli-Admin-Key: <operator-key>
Content-Type: application/json

{ "chainId": "0x61" }
```

成功响应中的 `status` 为 `confirmed`、`executionMode` 为 `simulated-bnb`、`simulated` 为 `true`，并带有 `network` 与 `txHash: "sim-bnb-..."`。该接口不会消耗 BNB、不会调用合约，也不会写入私钥。合约上线后，保留同一空投创建接口，由外部执行器通过 BNB RPC 广播真实交易，再使用下面的执行器回写接口提交真实 `0x` 交易哈希。

同一 `requestId` 幂等。`admin_approved` 会先从中心化金库预留 CLIP；任务失败时自动释放预留。空投记录状态为 `queued → submitted → confirmed`，失败状态为 `failed`。外部执行器负责使用平台钱包或托管签名服务完成链上转账，然后回调：

```http
POST /api/v1/internal/airdrops/{airdropId}/result
X-Clipli-Executor-Key: <executor-key>
Content-Type: application/json

{ "status": "confirmed", "txHash": "0x...", "executorRef": "batch-2026-08-22-01" }
```

执行器密钥只在服务端环境变量 `CLIPLI_AIRDROP_EXECUTOR_KEY` 中配置，运营密钥只在 `CLIPLI_ADMIN_API_KEY` 中配置。Clipli 前端永远不接触这两类密钥，也不直接签名或广播交易。建议执行器按 `walletAddress + chainId + token + amount + airdropId` 做二次幂等校验，并将链上交易哈希回写。

## 4. 平台业务 API（可选）

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

## 5. 外部资产源与 CLIP 金库 API

Clipli 的 HAPW 数据以外部平台为事实来源，服务端只做规范化、校验、缓存和操作记录。平台接入方需要提供以下上游接口：

| 上游接口 | 必需字段 | 用途 |
| --- | --- | --- |
| `GET /assets` | `id`、`tokenId`、`name`、`owner`、`rightsHolder`、`status`、`authorization`、`provenance` | 首次全量或游标增量同步；`owner` 用于钱包资产过滤 |
| `GET /assets?owner={walletAddress}`（推荐） | 同上 | 直接返回指定钱包的 HAPW 关联资产 |
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

CLIP 的公开 API 不提供铸币权限：`GET /api/v1/clip/treasury` 返回 BNB 合约的创世供应、治理预留/硬顶对应的本地金库快照、DEX 流动性配置、内部账本余额和守恒结果；`GET /api/v1/clip/distribution-rules` 返回中心化规则；`GET /api/v1/clip/distributions` 返回每次核销的分发流水。成功核销才按 `floor(creditYield × 0.40)` 从金库划拨，生成/授权/储备兑换费用回到金库；链上增发只能由延迟治理发行器执行。

`GET /api/v1/session` 用于获取当前调用上下文与可用操作能力。账户、身份、钱包和持久化策略由部署方配置；合作方不应仅凭钱包地址推断版权、资产或平台账户权限。

### 5.1 海文发模板镜像迁移

- `GET /api/v1/integrations/platform/templates`：只读调用海文发 `/openapi/tpls`，返回模板原始字段、Clipli 版本化映射和 `migrationReady`。模板描述按原文保存，客户端必须转义后展示。
- `POST /api/v1/integrations/platform/users/{userId}/migrations/preview`：只读校验模板、外部用户持仓数量和候选 Clipli 镜像资产，不会创建资产或核销海文发。
- `POST /api/v1/integrations/platform/users/{userId}/migrations`：受 `X-Clipli-Admin-Key` 保护的不可逆迁移。服务端先创建 `pending_external_write_off` 镜像资产，再核验持仓并调用海文发核销；海文发成功后激活并立即完成 Clipli 核销，发放 Creation Credits 和 CLIP。`requestId` 和 `requestNo` 均幂等。
- `GET /api/v1/integrations/platform/users/{userId}/migrations`：查询迁移审计记录。

迁移完成“海文发原资产核销 → Clipli 一致模板镜像资产创建/激活/核销 → Creation Credits 和 CLIP 发放”。成功响应中的 `creditsGranted` 和 `clipGranted` 记录本次结算；用户不能再次核销同一镜像资产。

## 6. 外部平台用户绑定与资产核销（联调框架）

以下接口用于 Clipli 与外部平台的手机号绑定和资产操作。外部平台是验证码校验、用户资产持仓和核销结果的事实来源；Clipli 只保存绑定关系、查询快照元数据和核销审计记录。当前未配置 `CLIPLI_EXTERNAL_PLATFORM_URL` 时，服务使用内存演示适配器（验证码固定为 `123456`，仅用于联调，不得用于生产）。

### 6.1 请求发送验证码

```http
POST /api/v1/integrations/platform/verification-codes
Content-Type: application/json

{"userId":"clip-user-1001","externalUserId":"100001","phone":"13800138000","requestId":"verify-1001-001"}
```

响应为 `202`，只返回 `verificationId`、脱敏手机号、投递状态和过期时间，不返回验证码本身。

### 6.2 提交绑定

```http
POST /api/v1/integrations/platform/bindings
Content-Type: application/json

{"userId":"clip-user-1001","externalUserId":"100001","code":"123456","verificationId":"verification-...","requestId":"bind-1001-001"}
```

Clipli 将验证码交给外部平台验证；`phone` 可选，省略时沿用第一步挑战中的手机号。外部平台成功返回 `externalUserId` 后，Clipli 才保存 `userId ↔ externalUserId` 绑定。相同 `requestId` 或同一用户的重试返回已有绑定，不重复建立关系。

可通过 `GET /api/v1/integrations/platform/bindings?userId=clip-user-1001` 读取当前绑定状态。

### 6.3 查询用户持仓

```http
GET /api/v1/integrations/platform/users/clip-user-1001/assets
```

响应中的 `items`、`assetCount` 和 `totalQuantity` 来自外部平台，`sourceOfTruth` 固定为 `external-platform`。未完成绑定的用户返回 `409 user_not_bound`。

### 6.4 按唯一流水号核销资产

```http
POST /api/v1/integrations/platform/users/clip-user-1001/assets/asset-2048/redemptions
Content-Type: application/json

{"serialNumber":"serial-20260826-0001","requestId":"redeem-1001-0001"}
```

也可使用 `POST /api/v1/integrations/platform/redemptions`，在请求体中同时传入 `userId` 和 `assetId`；`serialNo` 作为 `serialNumber` 的兼容别名。Clipli 以 `userId + assetId + serialNumber` 去重，并将流水号原样传给外部平台，重复请求直接返回首次核销记录，不重复调用或核销。

生产或测试联调时设置以下服务端变量。凭据永不下发前端，也不写入日志：

```bash
CLIPLI_EXTERNAL_PLATFORM_URL=https://api-test.hnccc.com/api
CLIPLI_EXTERNAL_PLATFORM_APP_ID=100001
CLIPLI_EXTERNAL_PLATFORM_APP_KEY=<海文发 AppKey>
CLIPLI_EXTERNAL_PLATFORM_TPL_IDS=100001,100002
```

适配器按海文发协议调用 `POST /openapi/user/bind/sms`、`POST /openapi/user/bind`、`GET /openapi/user/bind/status`、`GET /openapi/user/assets/count` 和 `POST /openapi/asset/write-off`，请求头为 `x-app-id`、`x-app-key`。`CLIPLI_EXTERNAL_PLATFORM_TPL_IDS` 仅用于兼容的全量持仓查询；精确数量查询应传 `tplIds`。业务码 `400/401/404/409/422/429` 会映射为对应 HTTP 状态，重复 `requestNo` 返回首次核销结果。

## 7. 平台兼容 API（非外部资产源必需）

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

## 8. 常见状态码

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
| `400` | `verification_code_invalid` | 外部平台拒绝手机号或验证码 |
| `409` | `user_not_bound` | 用户尚未完成外部平台绑定 |
| `409` | `serial_number_reused` | 唯一流水号已用于其他资产核销 |
| `502` | `external_platform_unavailable` | 外部平台暂时不可用或未配置 |
