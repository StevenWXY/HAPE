# Clipli

Clipli 是以 HAPW 授权证书为核心的 AI 视频创作与资产行权原型。后端已使用 Go 重构，前端保持无构建依赖。

## 技术结构

- Go 1.22+，仅使用标准库 `net/http`，无第三方运行依赖。
- `cmd/server`：进程入口、超时与优雅停机。
- `internal/domain`：HAPW、外部资产引用、授权、核销、行权、兑换、CLIP 金库与生成任务模型。
- `internal/store`：线程安全内存仓储和演示数据。
- `internal/service`：额度、CLIP、核销、生成、行权与兑换规则。
- `internal/httpapi`：REST API、兼容路由、安全响应头和静态文件服务。
- `api/openapi.yaml`：OpenAPI 3.1 接口规范。
- `docs/API.md`：面向产品与前端的中文接口总览。

当前仓储为进程内演示状态，重启后复位。原型不做真实身份验证或 KYC，钱包连接仅作为可选操作句柄。正式环境应把 `internal/store` 替换为 PostgreSQL，并把平台审计、钱包签名和外部平台凭据放在服务端边界内。

作品和 HAPW 默认使用各外部平台的演示快照。部署时可设置 `WORK_SOURCE_URL` 接入作品 API，设置 `HAPW_ASSET_SOURCE_URL` 接入规范化 HAPW 资产 API；请求超时、状态码异常、字段不完整或返回为空时保留最后一次有效快照，并在同步记录中登记原因。

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
