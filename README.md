# Clipli

Clipli 是以 HAPW 授权证书为核心的 AI 视频创作与资产行权原型。后端已使用 Go 重构，前端保持无构建依赖。

## 技术结构

- Go 1.22+，仅使用标准库 `net/http`，无第三方运行依赖。
- `cmd/server`：进程入口、超时与优雅停机。
- `internal/domain`：HAPW、外部资产引用、授权、核销、行权、兑换、CLIP 金库与生成任务模型。
- `internal/store`：单进程持久化账本、事务回滚及文件排他锁；首次启动资产和余额为空，重启恢复已有真实记录。
- `internal/service`：额度、CLIP、核销、生成、行权、兑换和外部平台绑定规则；外部平台能力通过可替换适配器接入。
- `internal/httpapi`：REST API、兼容路由、安全响应头和静态文件服务。
- `api/openapi.yaml`：OpenAPI 3.1 接口规范。
- `docs/API.md`：面向产品与前端的中文接口总览。

账本默认保存到 `.data/clipli-state.json`，可通过 `CLIPLI_STATE_FILE` 指定持久化路径。原型仍是单账户本地会话，尚未实现多用户登录鉴权和钱包签名证明，不应直接开放到公网处理真实客户资产。规模化部署需要数据库事务、多用户账本和鉴权。真实空投由 `scripts/execute-airdrop.js` 独立签名执行，API 服务只读取 BNB RPC 并核验回执，不接收私钥。

海文发七项接口、兑换审批和空投实现的逐项核对、验证范围及正式配置见 [海文发集成核对报告](docs/HAIWEN_INTEGRATION_AUDIT.md)。

作品和资产默认均为空。`WORK_SOURCE_URL` 可接入真实作品；`HAPW_ASSET_SOURCE_URL` 仅接收可信平台确认的划转资产，必须包含来源平台、外部资产 ID、`external.transferId`、`external.syncStatus: migrated` 与归属账号。普通持仓目录不等于已划转资产，不能直接入账。有效空数组会清空本次快照；错误响应保留最近有效数据，不回退到演示资产。

外部平台手机号绑定和资产核销已提供联调框架：`POST /api/v1/integrations/platform/verification-codes` 发送验证码、`POST /api/v1/integrations/platform/bindings` 提交绑定、`GET /api/v1/integrations/platform/users/{userId}/assets?tplIds=100001,100002` 查询模板当前可核销数量、`POST /api/v1/integrations/platform/users/{userId}/migrations` 按模板数量核销并创建 Clipli 镜像。未配置 `CLIPLI_EXTERNAL_PLATFORM_URL` 时接口明确返回未配置错误；配置后由服务端按海文发协议调用作品、模板、绑定、计数和批量核销接口，认证头为 `x-app-id/x-app-key`。

## 运行

需要 Go 1.22 或更高版本。

```bash
npm start
```

也可以直接运行：

```bash
PORT=4173 HOST=127.0.0.1 go run ./cmd/server
```

打开 `http://127.0.0.1:4173`。健康检查位于 `GET /api/health`。

## 测试

```bash
npm test
```

测试覆盖 HAPW 核销幂等、中心化 CLIP 金库守恒、额度与 CLIP 计算、兑换报价、每日限额、外部资产引用、资产查询和 HTTP 响应契约。

## API 选择

新客户端使用 `/api/v1/hapw/*`。现有网页继续使用 `/api/*` 兼容接口，两者调用同一服务层，不会产生两套状态或规则。

- 完整接口清单：[docs/API.md](docs/API.md)
- OpenAPI 规范：[api/openapi.yaml](api/openapi.yaml)

### 钱包与空投配置

```bash
CLIPLI_ADMIN_API_KEY=<operator-key> \
CLIPLI_AIRDROP_EXECUTOR_KEY=<executor-callback-key> \
PORT=4173 go run ./cmd/server
```

### 海文发联调配置

凭据只在服务端环境变量中配置，不要放入前端、日志或 Git：

```bash
CLIPLI_EXTERNAL_PLATFORM_URL=https://api-test.hnccc.com/api \
CLIPLI_EXTERNAL_PLATFORM_APP_ID=100001 \
CLIPLI_EXTERNAL_PLATFORM_APP_KEY=<海文发 AppKey> \
PORT=4173 go run ./cmd/server
```

如需启用海文发模板迁移，额外配置服务端维护的版本化映射（示例只映射测试模板 `100053`）：

```bash
CLIPLI_EXTERNAL_ASSET_MAPPINGS='[{"tplId":100053,"version":"haiwen-2026-09-v1","creditYield":150,"clipPrice":520,"currency":"CNY","active":true}]'
```

`GET /api/v1/integrations/platform/works` 只读获取海文发 `/openapi/works` 作品目录（`publishNum` 仅为发行统计）；`GET /api/v1/integrations/platform/templates` 只读获取海文发 `/openapi/tpls` 模板并返回映射状态；`POST /api/v1/integrations/platform/users/{userId}/migrations/preview` 只读校验模板、当前可核销数量和候选镜像资产；真正迁移接口需要 `X-Clipli-Admin-Key`，会按 `requestNo` 幂等地创建待核销的 Clipli 镜像资产并核销海文发资产。海文发核销成功后，Clipli 会在同一笔迁移中激活并核销镜像资产，自动结算 Creation Credits 和 CLIP；镜像资产不能再次核销。

海文发适配器不支持未带 `tplIds` 的全量资产列表；查询用户数据时必须在 URL 传入 `tplIds`，返回值仅为当前可核销数量。生产环境将 URL 替换为 `https://api.hnccc.com/api`，并使用独立生产凭据。

不配置上述变量时，公开钱包资产、资格和空投查询仍可用，运营创建与执行器回调接口返回配置缺失。外部执行器应轮询 `GET /api/v1/admin/airdrops`，使用平台托管钱包完成 CLIP 转账，再回调 `POST /api/v1/internal/airdrops/{id}/result`。详细请求体、规则公式和安全边界见 [docs/API.md](docs/API.md) 的“钱包资产与 CLIP 空投 API”。

当前 CLIP 空投只接受 BNB Smart Chain 测试网 `0x61` 与主网 `0x38`。正式 BEP-20 合约、BNB RPC 和签名执行器尚未配置，运营控制台可调用 `POST /api/v1/admin/airdrops/{airdropId}/simulate` 验证地址、队列和中心化金库流程。返回的 `sim-bnb-*` 是平台模拟回执，不是链上交易；合约上线后再由外部执行器广播真实交易并回写 `0x` 哈希。

### BNB CLIP 合约

Solidity 合约、部署脚本和参数说明见 [docs/BNB_TOKEN.md](docs/BNB_TOKEN.md)。本地可运行：

```bash
npm run contracts:compile
npm run contracts:test
npx hardhat run scripts/deploy-bnb.js --network hardhat
```

`ClipToken` 固定初始发行 9 亿、硬上限 10 亿；`AdaptiveMinter` 按经审计的指标、延迟和年度上限释放预留额度；`ClipAirdrop` 使用 Merkle 证明并按活动隔离托管余额。`ClipliDexPair` 仅用于本地 BNB/EVM 测试，生产兑换应接入经过审计的 PancakeSwap 路由，不应把测试 Pair 部署到主网。

## 前端资产与平台绑定

导航右上角同时提供钱包与外部平台入口。海文发通过现有短信验证和绑定接口关联账号，绑定后可查询平台持仓；OpenSea 等尚未接入的平台明确展示状态。资产仪表盘仅显示已确认划转的资产及来源凭据，没有资产时提供绑定引导。旧免验证绑定接口已经退役。

此项目仍使用进程内账户与账本状态，不是正式多用户部署。真实账户认证、持久化、金库资金与结算执行器需要正式部署接入；默认不注入任何演示资产或金库余额。海文发迁移写入保留运营审核权限，浏览器不持有运营密钥。

浏览器回归脚本为 `scripts/test-real-assets-ui.js`，需要可用的 Playwright 和本机 Chrome。启动本地服务后运行 `CLIPLI_TEST_URL=http://127.0.0.1:4175 node scripts/test-real-assets-ui.js`。验证码与持仓响应仅在测试浏览器中模拟，不会发送真实短信或操作外部资产。

现有海文发本地配置使用测试接口。页面会标注测试环境，允许绑定与持仓联调；服务端禁止测试平台划转计入正式资产。真实划转需配置生产接口及正式结算条件。
