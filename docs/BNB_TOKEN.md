# CLIP BNB 合约与经济模型

本目录包含面向 BNB Smart Chain 的 Solidity 合约。Hardhat 本地网络使用 Chain ID `97`，对应 BNB Smart Chain Testnet；代码没有 ETH 主网地址、RPC 或生产私钥。

## 供应模型

- 最大供应硬顶：`1,000,000,000 CLIP`。
- 创世供应：`900,000,000 CLIP`，部署时一次性进入金库，不存在公开铸币 API。
- 自适应发行储备：最多 `100,000,000 CLIP`，由 `AdaptiveMinter` 受限发行。
- 每个 epoch 使用活跃用户增长 50%、已验证持有人增长 30%、结算量增长 20% 的加权分数；增长不足不发行。
- 每 epoch 最多按创世供应的 25 bps 发行，每年最多 5%；提案排队时即锁定年度和总量配额。
- 发行提案必须等待延迟窗口，Guardian 可取消或暂停；建议生产环境将 DEFAULT_ADMIN_ROLE 和 Guardian 置于 OpenZeppelin Timelock + 多签之后。

## 合约

- `ClipToken.sol`：OpenZeppelin ERC20、Permit、Burnable、Pausable、AccessControl；只有 `MINTER_ROLE` 可铸币，且受硬顶限制。
- `ClipliTimelock.sol`：OpenZeppelin TimelockController 包装，用于延迟治理和角色移交。
- `AdaptiveMinter.sol`：指标快照、增长计算、延迟提案、年度/硬顶配额、Guardian 暂停。
- `ClipAirdrop.sol`：按 `assetRef + Merkle root` 的资产持有快照空投；位图防重复领取，活动资金隔离，过期只回收本活动余额。
- `ClipliDexPair.sol`：用于本地测试的最小恒积 AMM，带 0.30% 手续费、截止时间、滑点、重入保护和恒积检查。生产环境应优先使用经过审计的 PancakeSwap/Uniswap 部署，不要直接把该 pair 当作主网 DEX。

## 部署与权限

```bash
npx hardhat run scripts/deploy-bnb.js --network hardhat
```

配置真实 BNB 测试网时，使用密钥管理系统注入 `BNB_TESTNET_RPC_URL` 和 `BNB_DEPLOYER_PRIVATE_KEY`，再运行 `--network bscTestnet`。仓库中的 `.env.bnb.example` 只有占位符；不要把真实私钥写入文件、前端或 CI 日志。

部署脚本在 `bscTestnet`/`bsc` 网络会拒绝缺少治理、金库或供应参数的运行，避免把临时部署者误设为生产权限；Hardhat 本地网络才允许使用临时默认值。

生产部署必须显式设置 `CLIP_TREASURY`、`CLIP_INITIAL_SUPPLY`、`CLIP_HARD_CAP`、`CLIP_TIMELOCK_PROPOSER`、`CLIP_TIMELOCK_EXECUTOR`、`CLIP_GUARDIAN`、`CLIP_METRICS_ORACLE` 和 `CLIP_CAMPAIGN_OPERATOR`，并由硬件钱包或多签执行。部署脚本默认只使用 Hardhat 临时账户；不从仓库读取私钥，也不会生成或保存生产私钥。部署完成后，部署者只保留临时接线权限，Token、发行器和空投活动的最终管理员都移交到 Timelock；Timelock 同时持有发行器 Guardian 和 Token Pauser 角色，独立 Guardian 负责即时响应。

## 安全边界

本版本未连接 BNB RPC、未部署正式合约、未配置真实 DEX 池、预言机或外部执行器。空投 Merkle root 必须由可审计的资产快照生成，不能仅凭用户提交的钱包地址；真实 CLIP 空投必须由独立执行器或多签钱包广播，并将链上交易哈希回写 CLIPLI。合约上线前应完成第三方审计、形式化不变量检查、权限演练、灾备和法律/合规评估。
