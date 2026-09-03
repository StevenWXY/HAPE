# 外部平台开放接口文档

## 1. 接口概览

| 接口 | 方法 | 说明 |
| --- | --- | --- |
| `/openapi/works` | GET | 获取作品列表 |
| `/openapi/tpls` | GET | 获取版权模板列表 |
| `/openapi/user/bind/status` | GET | 查询用户绑定状态 |
| `/openapi/user/bind/sms` | POST | 发送绑定验证码 |
| `/openapi/user/bind` | POST | 完成用户绑定 |
| `/openapi/user/assets/count` | GET | 查询用户持有资产数量 |
| `/openapi/asset/write-off` | POST | 核销资产 |

## 2. 公共约定

### 环境地址

| 环境 | 请求基础地址 |
| --- | --- |
| 测试环境 | `https://api-test.hnccc.com/api` |
| 正式环境 | `https://api.hnccc.com/api` |

接口请求地址由请求基础地址与接口路径组成。例如，测试环境获取作品列表的完整地址为：

```text
https://api-test.hnccc.com/api/openapi/works
```

所有接口通过请求头携带平台 App ID 和独立访问密钥：

```http
x-app-id: 平台App ID
x-app-key: 平台App Key
```

POST 接口同时携带：

```http
Content-Type: application/json
```

`externalUserId` 表示外部平台用户 ID。平台 App ID 和 App Key 由我方提供，请妥善保管，不要在客户端或公开环境中泄露。

通用成功响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "msg": "success"
}
```

通用失败响应沿用现有响应结构，`code` 为业务码：

```json
{
  "code": 400,
  "message": "pageSize不能超过100",
  "data": null,
  "msg": "pageSize不能超过100"
}
```

通用响应参数：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `code` | Number | 业务状态码；`0` 表示成功，非 `0` 表示失败 |
| `message` | String | 响应结果说明 |
| `data` | Object / Array / null | 业务响应数据，具体字段见各接口的响应参数说明；失败时通常为 `null` |
| `msg` | String | 兼容响应字段，含义与 `message` 相同 |

| 业务码 | 说明 |
| --- | --- |
| `400` | 请求参数、手机号或验证码错误，验证码已过期或已使用 |
| `401` | 缺少或使用了不匹配的平台 App ID、App Key |
| `404` | 用户未绑定、手机号未注册、版权模板不存在 |
| `409` | 用户绑定关系冲突，或核销流水号参数冲突 |
| `422` | 可核销资产不足或并发竞争导致核销失败 |
| `429` | 短信发送过于频繁或验证码错误次数过多 |
| `500` | 服务器内部错误 |

## 3. 获取作品列表

```http
GET /openapi/works?pageNum=1&pageSize=20
```

### 请求参数

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `pageNum` | 否 | 页码，默认值为 `1` |
| `pageSize` | 否 | 每页数量，默认值为 `20` |

`pageNum`、`pageSize` 必须为正整数，`pageSize` 最大为 `100`。仅返回 `status=1`、`deleted=false` 的作品，并按 `_id` 倒序分页。响应字段使用白名单，不返回 `createTime` 或其他内部字段。

### 响应参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `data.list` | Array\<Object\> | 当前页作品列表 |
| `data.list[].workId` | Number | 作品 ID |
| `data.list[].worksName` | String | 作品名称 |
| `data.list[].showcase` | Array\<String\> | 作品展示图片地址列表 |
| `data.list[].authors` | Array\<Object\> | 作者列表 |
| `data.list[].authors[].id` | Number | 作者 ID |
| `data.list[].authors[].name` | String | 作者名称 |
| `data.list[].owners` | Array\<Object\> | 著作权人列表 |
| `data.list[].owners[].id` | Number | 著作权人 ID |
| `data.list[].owners[].name` | String | 著作权人名称 |
| `data.list[].worksType` | Number / null | 作品一级类型 ID |
| `data.list[].worksSubType` | Number / null | 作品二级类型 ID |
| `data.list[].worksTypeName` | String | 作品类型名称 |
| `data.list[].worksIntroduce` | String | 作品描述 |
| `data.list[].publishNum` | Number | 作品发行数量 |
| `data.total` | Number | 符合条件的作品总数 |
| `data.pageNum` | Number | 当前页码 |
| `data.pageSize` | Number | 每页数量 |

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "workId": 100001,
        "worksName": "作品名称",
        "showcase": [
          "https://example.com/image.png"
        ],
        "authors": [
          {
            "id": 100001,
            "name": "作者名称"
          }
        ],
        "owners": [
          {
            "id": 100001,
            "name": "著作权人名称"
          }
        ],
        "worksType": 1,
        "worksSubType": 10,
        "worksTypeName": "作品类型名称",
        "worksIntroduce": "作品描述",
        "publishNum": 10
      }
    ],
    "total": 100,
    "pageNum": 1,
    "pageSize": 20
  },
  "msg": "success"
}
```

作者和著作权人仅返回 `id`、`name`，不返回手机号、证件号等敏感信息。

`publishNum` 是作品发行统计，不代表剩余库存、用户余额、可核销数量或可迁移数量。

## 4. 获取版权模板列表

```http
GET /openapi/tpls?pageNum=1&pageSize=20&workId=100001
```

### 请求参数

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `pageNum` | 否 | 页码，默认值为 `1` |
| `pageSize` | 否 | 每页数量，默认值为 `20` |
| `workId` | 否 | 作品 ID；传入时只查询指定作品的版权模板 |

仅返回 `status=1`、`deleted=false` 且所属作品审核通过、未删除的模板，按 `_id` 倒序分页。作者、著作权人信息以所属作品记录为准，均只返回 `id`、`name`。

### 响应参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `data.list` | Array\<Object\> | 当前页版权模板列表 |
| `data.list[].tplId` | Number | 版权模板 ID |
| `data.list[].name` | String | 版权模板名称 |
| `data.list[].description` | String | 版权模板描述 |
| `data.list[].image` | String | 版权模板图片地址 |
| `data.list[].workId` | Number | 所属作品 ID |
| `data.list[].worksName` | String | 所属作品名称 |
| `data.list[].worksType` | Number / null | 作品一级类型 ID |
| `data.list[].worksSubType` | Number / null | 作品二级类型 ID |
| `data.list[].worksTypeName` | String | 作品类型名称 |
| `data.list[].authors` | Array\<Object\> | 作者列表 |
| `data.list[].authors[].id` | Number | 作者 ID |
| `data.list[].authors[].name` | String | 作者名称 |
| `data.list[].owners` | Array\<Object\> | 著作权人列表 |
| `data.list[].owners[].id` | Number | 著作权人 ID |
| `data.list[].owners[].name` | String | 著作权人名称 |
| `data.list[].publishCount` | Number | 版权模板发行数量 |
| `data.total` | Number | 符合条件的版权模板总数 |
| `data.pageNum` | Number | 当前页码 |
| `data.pageSize` | Number | 每页数量 |

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "tplId": 100001,
        "name": "版权名称",
        "description": "版权描述",
        "image": "https://example.com/image.png",
        "workId": 100001,
        "worksName": "作品名称",
        "worksType": 1,
        "worksSubType": 10,
        "worksTypeName": "作品类型名称",
        "authors": [
          {
            "id": 100001,
            "name": "作者名称"
          }
        ],
        "owners": [
          {
            "id": 100001,
            "name": "著作权人名称"
          }
        ],
        "publishCount": 1000
      }
    ],
    "total": 20,
    "pageNum": 1,
    "pageSize": 20
  },
  "msg": "success"
}
```

## 5. 查询用户绑定状态

```http
GET /openapi/user/bind/status?externalUserId=U123456
```

### 请求参数

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `externalUserId` | 是 | 外部平台用户 ID |

### 响应参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `data.externalUserId` | String | 外部平台用户 ID |
| `data.bound` | Boolean | 是否已绑定本平台用户 |
| `data.boundAt` | String / null | 绑定时间，使用 ISO 8601 格式；未绑定时为 `null` |

### 已绑定响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "externalUserId": "U123456",
    "bound": true,
    "boundAt": "2026-08-26T10:00:00+08:00"
  },
  "msg": "success"
}
```

### 未绑定响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "externalUserId": "U123456",
    "bound": false,
    "boundAt": null
  },
  "msg": "success"
}
```

绑定状态以海文发返回为事实源。Clipli 查询本地绑定详情时应回源调用此接口；本地缓存丢失、外部解绑或 `externalUserId` 不匹配时不得继续资产操作。`boundAt` 优先采用海文发返回值。

## 6. 发送绑定验证码

```http
POST /openapi/user/bind/sms
```

### 请求参数

```json
{
  "externalUserId": "U123456",
  "phone": "13800138000"
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `externalUserId` | 是 | 外部平台用户 ID |
| `phone` | 是 | 本平台注册手机号 |

服务端检查手机号是否为本平台注册用户，然后发送绑定验证码。验证码使用独立绑定场景，并按手机号和 IP 进行发送频率限制。响应 `data` 为空，不提供 `deliveryId`、实际过期时间或验证码状态；任何本地 challenge ID 和五分钟展示期限都只能表示 Clipli 本地流程状态。

### 响应参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `data` | Object | 发送成功时返回空对象；验证码通过短信发送，不在接口响应中返回 |

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "msg": "success"
}
```

## 7. 完成用户绑定

```http
POST /openapi/user/bind
```

### 请求参数

```json
{
  "externalUserId": "U123456",
  "phone": "13800138000",
  "smsCode": "123456"
}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `externalUserId` | 是 | 外部平台用户 ID |
| `phone` | 是 | 本平台注册手机号 |
| `smsCode` | 是 | 短信验证码 |

服务端验证手机号和短信验证码，建立外部平台用户与本平台用户的绑定关系。只有海文发返回 `bound=true` 且 `externalUserId` 与请求匹配时才保存绑定；`boundAt` 优先使用海文发返回值。验证码验证成功后立即失效，相同绑定重复提交时直接返回已有绑定结果。

同一平台下，外部用户与本平台用户严格一对一绑定。任意一方已绑定其他账号时返回 `409`，不会自动覆盖现有关系。当前不提供解绑和重新绑定接口。

### 响应参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `data.externalUserId` | String | 已完成绑定的外部平台用户 ID |
| `data.bound` | Boolean | 是否绑定成功；成功响应固定为 `true` |
| `data.boundAt` | String / null | 绑定时间，使用 ISO 8601 格式 |

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "externalUserId": "U123456",
    "bound": true,
    "boundAt": "2026-08-26T10:00:00+08:00"
  },
  "msg": "success"
}
```

## 8. 查询用户持有资产数量

```http
GET /openapi/user/assets/count?externalUserId=U123456&tplIds=100001,100002
```

### 请求参数

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `externalUserId` | 是 | 已绑定的外部平台用户 ID |
| `tplIds` | 是 | 版权模板 ID，多个 ID 使用英文逗号分隔，最多 `100` 个 |

接口按 `tplIds` 分别统计可核销资产数量。没有可核销资产的模板返回 `0`。该数量是当前最大可核销数量，并非用户历史持有总量。响应只表达 `tplId -> 当前可核销数量`，不提供独立资产 ID、钱包资产列表、历史购买数量或可自由转移数量；Clipli 不得将它包装为逐枚资产持仓。

可核销资产必须同时满足：属于绑定用户、`holderType=USER`、未删除、未履约、未在本平台或第三方挂售、未冻结、未处于禁售期且持有有效期未过期。数量查询与核销使用完全相同的查询条件。

### 响应参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `data.externalUserId` | String | 外部平台用户 ID |
| `data.list` | Array\<Object\> | 各版权模板的可核销资产数量列表 |
| `data.list[].tplId` | Number | 版权模板 ID |
| `data.list[].count` | Number | 当前可核销资产数量；没有可核销资产时为 `0` |

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "externalUserId": "U123456",
    "list": [
      {
        "tplId": 100001,
        "count": 10
      },
      {
        "tplId": 100002,
        "count": 0
      }
    ]
  },
  "msg": "success"
}
```

## 9. 核销资产

```http
POST /openapi/asset/write-off
```

### 请求参数

```json
{
  "requestNo": "WO202608260001",
  "externalUserId": "U123456",
  "tplId": 100001,
  "num": 5
}
```

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `requestNo` | String | 是 | 外部平台生成的唯一核销流水号；去除首尾空格后不能为空，长度不能超过 `128` 个字符 |
| `externalUserId` | String | 是 | 已绑定的外部平台用户 ID；去除首尾空格后不能为空，长度不能超过 `128` 个字符 |
| `tplId` | Number | 是 | 需要核销的资产模板 ID，必须为大于 `0` 的安全整数 |
| `num` | Number | 是 | 核销数量，必须为大于 `0` 且小于等于 `100` 的安全整数 |

服务端会先去除 `requestNo`、`externalUserId` 的首尾空格，并将 `tplId`、`num` 转换为数字后校验。任一参数缺失、字符串超长、不是整数、小于等于 `0`，或 `num` 大于 `100` 时，接口返回 `400` 和“参数错误”。调用方应使用 JSON Number 传递 `tplId`、`num`。这是按模板批量核销，不是指定某一枚资产的转移；接口不接收序列号、NFT ID 或单个资产 ID，也不返回真实交易哈希或资产 ID。

### 处理规则

1. 验证用户绑定关系。
2. 验证用户持有指定模板的有效资产。
3. 排除已核销或正在挂售的资产。
4. 持有数量不足时整体失败，不进行部分核销。
5. 使用数据库事务批量核销并在内部记录模板、数量和幂等请求审计；由于接口没有逐枚标识，不得声称记录了海文发实际资产 ID。
6. 相同 `requestNo` 重复请求时返回第一次的处理结果。
7. 相同 `requestNo` 但 `externalUserId`、`tplId` 或 `num` 不一致时返回 `409`。
8. 核销按资产 `_id` 升序选择，在一个事务内完成；资产不足或并发竞争时整体回滚。

### 响应参数

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `data.requestNo` | String | 外部平台核销流水号 |
| `data.externalUserId` | String | 外部平台用户 ID |
| `data.tplId` | Number | 已核销资产的版权模板 ID |
| `data.num` | Number | 本次核销数量 |
| `data.status` | String | 核销状态；成功时为 `SUCCESS` |
| `data.writeOffAt` | String | 核销完成时间，使用 ISO 8601 格式 |

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "requestNo": "WO202608260001",
    "externalUserId": "U123456",
    "tplId": 100001,
    "num": 5,
    "status": "SUCCESS",
    "writeOffAt": "2026-08-26T10:00:00+08:00"
  },
  "msg": "success"
}
```

Clipli 仅在本平台保存模板、数量和幂等请求审计；由于海文发接口没有逐枚资产标识，不得声称记录了海文发实际资产 ID。核销成功后，相应模板数量减少，也不能重复提交同一 `requestNo`。

核销是独立的“外部使用”行为：资产会标记为 `holderType=DELETED、deleted=true`，但不会创建版权履约记录，不会设置 `performed=true`，也不会增加模板的 `performNum`。当前不提供核销撤销接口。

## 10. 绑定与核销流程

1. 外部平台先调用绑定状态接口。
2. 未绑定时调用发送验证码接口。
3. 用户输入验证码后，外部平台调用绑定接口。
4. 绑定成功后，可查询用户资产数量。
5. 核销时传入唯一流水号、用户 ID、模板 ID 和核销数量。

## 11. Clipli 模板镜像与资产迁移

海文发的 `/openapi/tpls` 返回版权模板目录，可用于在 Clipli 展示和生成镜像资产，但模板目录本身不证明某个用户的持仓，也不包含 Clipli 的 Creation Credits 或 CLIP 经济参数。Clipli 因此将源模板快照与服务端版本化映射分开保存。

### 11.1 读取模板目录

```http
GET /api/v1/integrations/platform/templates?page=1&pageSize=20&workId=100002
```

该接口只读调用海文发 `GET /openapi/tpls`，返回 `template` 原始字段、可选的 `mapping` 和 `migrationReady`。模板描述按原文保存，不得直接作为 HTML 注入页面；图片仅接受安全的 `https` 地址。

### 11.2 迁移预览

```http
POST /api/v1/integrations/platform/users/100001/migrations/preview
Content-Type: application/json

{
  "tplId": 100053,
  "num": 1,
  "requestNo": "HW-MIGRATION-100053-01",
  "requestId": "migration-100053-01"
}
```

预览会读取模板目录和 `/openapi/user/assets/count`，返回持仓数量、映射规则和候选 Clipli 镜像资产；不会创建资产，也不会调用核销接口。

### 11.3 执行迁移

```http
POST /api/v1/integrations/platform/users/100001/migrations
X-Clipli-Admin-Key: <operator-key>
Content-Type: application/json

{
  "tplId": 100053,
  "num": 1,
  "requestNo": "HW-MIGRATION-100053-01",
  "requestId": "migration-100053-01",
  "accepted": true
}
```

执行顺序为：创建 `pending_external_write_off` 镜像资产 → 校验外部持仓 → 调用海文发核销 → 核销成功后激活并立即核销 Clipli 镜像资产 → 发放 Creation Credits 和 CLIP。`requestId` 和 `requestNo` 均幂等；海文发失败时镜像资产保持不可用并记录失败状态。成功迁移返回 `creditsGranted`、`clipGranted` 和本地 `redemptionIds`，同一镜像资产不可再次核销。

### 11.4 查询迁移记录

```http
GET /api/v1/integrations/platform/users/100001/migrations
```

生产环境应将 `X-Clipli-Admin-Key` 替换为正式用户鉴权、运营审批和审计权限；当前原型使用运营密钥保护不可逆迁移操作。
