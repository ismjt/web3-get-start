# Hardhat Ignition 部署指南

## 概述

本项目使用 **Hardhat Ignition** 进行智能合约部署，这是 Hardhat 团队开发的声明式部署框架。

## 部署模块结构

### 模块文件组织

```
ignition/modules/
├── 0-Complete.ts          # 完整部署（一次性部署所有）
├── 1-MajtDutchNFT.ts      # 第1步: 测试NFT合约
├── 2-PriceOracle.ts       # 第2步: 价格预言机
├── 3-NFTAuction.ts        # 第3步: 拍卖逻辑合约
└── 4-AuctionFactory.ts    # 第4步: 拍卖工厂合约 + 代理
```

### 部署依赖关系

```
1-MajtDutchNFT.ts (独立)

3-NFTAuction.ts
    ↓
    └─→ 4-AuctionFactory.ts ← 2-PriceOracle.ts
           ↑
           └─ (初始化时依赖)
```

## 各模块详解

### 模块 1: MajtDutchNFT (测试NFT)

**文件**: `1-MajtDutchNFT.ts`

```typescript
// 输出
{
  nftContract: Address  // NFT合约
}
```

**说明**:
- 部署测试用 ERC721 NFT 合约
- 基于 ERC721URIStorage，支持自定义 metadata URI
- 支持 safeMint 铸造新 NFT

**功能**:
- `safeMint(recipient, tokenUri)`: 铸造新 NFT
- 自动递增 Token ID (从 0 开始)
- 使用 IPFS 基础 URI: `ipfs://QmeSjSinHpPnmXmspMjwiXyN6zS4E9zccariGR3jxcaWtq/`

---

### 模块 2: PriceOracle (价格预言机)

**文件**: `2-PriceOracle.ts`

```typescript
// 输出
{
  priceOracle: Address  // 价格预言机合约
}
```

**说明**:
- 部署 Chainlink 价格预言机集成合约
- 构造函数自动配置 ETH 和 LINK 价格源
- 支持任意 ERC20 代币的价格查询（需先配置价格源）

**Chainlink 价格源 (Sepolia)**:

| 资产 | 地址 |
|-----|------|
| ETH/USD | 0x694AA1769357215DE4FAC081bf1f309aDC325306 |
| LINK/USD | 0xc59E3633BAAC79493d908e63626716e204A45EdF |

**配置新代币价格源**:

```typescript
// 需要后续通过工厂调用
await factory.setAuctionTokenPriceFeed(tokenAddress, chainlinkFeedAddress);
```

### 模块 3: NFTAuction (拍卖实现)

**文件**: `3-NFTAuction.ts`

```typescript
// 输出
{
  nftAuctionImpl: Address  // 拍卖逻辑实现合约
}
```

**说明**:
- 部署 NFTAuction 逻辑合约
- 这是一个实现合约，不直接交互
- 工厂合约会通过 ERC1967Proxy 为每次拍卖创建代理实例

**关键点**:
- 不需要初始化（工厂创建代理时初始化）
- 合约中使用 `_disableInitializers()` 防止直接初始化

### 模块 4: AuctionFactory (工厂合约)

**文件**: `4-AuctionFactory.ts`

```typescript
// 输出
{
  factoryImpl: Address    // 工厂实现合约
  factoryProxy: Address  // 工厂代理（实际使用地址）
}
```

**说明**:
- 部署 AuctionFactory 实现合约
- 通过 UUPSProxy 创建代理并初始化
- 初始化参数:
  - `admin`: 部署账户（可进行管理操作）
  - `auctionImplementation`: NFTAuction 实现地址
  - `priceOracle`: PriceOracle 地址

**关键点**:
- 使用 UUPS 升级模式（Upgradeable Universal Proxy Standard）
- 可升级性: 需要 UPGRADER_ROLE 权限
- 默认配置: feeRate = 250 (2.5%), feeRecipient = admin

## 部署方式

### 前置要求

```bash
# 1. 安装依赖
npm install

# 2. 配置环境变量 (.env)
SEPOLIA_RPC_URL=https://sepolia.infura.io/v3/YOUR_KEY
SEPOLIA_PRIVATE_KEY=your_private_key_here

# 3. 验证 Hardhat 配置
npx hardhat --version  # 3.x
```

### 方式 1: 完整部署（推荐）

一次性部署所有合约，自动处理依赖:

```bash
# 部署到 Sepolia
npx hardhat ignition deploy ignition/modules/0-Complete.ts --network sepolia

# 部署到本地
npx hardhat ignition deploy ignition/modules/0-Complete.ts --network localhost

# 模拟部署（不写链上）
npx hardhat ignition deploy ignition/modules/0-Complete.ts --network sepolia --dry-run
```

**输出示例**:
```
MajtNFTModule#MajtDutchNFT - 0xb3Cd2F5d2BE278b090115F2Ac291cc9ae5cE6653
AuctionFactoryModule#AuctionFactory - 0x95ab504AA0Faf75ea3f0d0A7A519c1f36225a3eD
NFTAuctionModule#NFTAuction - 0xC459b124657E668952fc0550722dB4b3767CEF75
PriceOracleModule#PriceOracle - 0x7cAa536A5FFCC2a2b12044B8f3CA3b892cBCC76b
AuctionFactoryModule#UUPSProxy - 0x782d13f3874eB51e2200e0a12B4d80A9C9D29177
```

### 方式 2: 分步部署

分别部署每个模块，便于逐步验证:

```bash

```

### 方式 3: 本地测试

```bash
# 启动本地节点
npx hardhat node

# 在另一个终端部署
npx hardhat ignition deploy ignition/modules/0-Complete.ts --network localhost
```

## 部署状态管理

Ignition 自动追踪部署状态，保存在 `ignition/deployments/` 目录:

```
ignition/deployments/
├── chain-<chainId>/
│   ├── deployed_addresses.json      # 已部署合约地址
│   └── execution_state.json         # 执行状态
```

### 恢复中断的部署

如果部署中断，Ignition 可以自动恢复:

```bash
# 继续未完成的部署
npx hardhat ignition deploy ignition/modules/0-Complete.ts --network sepolia
# 会自动检查已部署的合约，跳过已完成的步骤
```

### 查看部署信息

```bash
# 列出所有已部署合约
npx hardhat ignition status

# 获取特定链的部署地址
cat ignition/deployments/chain-11155111/deployed_addresses.json
```

## 部署后配置

### 1. 验证合约 (可选)

```bash
npx hardhat verify --network sepolia <CONTRACT_ADDRESS>
```

### 2. 配置代币价格源

```bash
# 如需支持其他 ERC20 代币，配置其 Chainlink 价格源
# 示例: 配置 USDC 价格源

ADMIN_ADDRESS=0x...  # factory 的 admin
FACTORY_ADDRESS=0x...
USDC_ADDRESS=0x...
USDC_PRICE_FEED=0x...  # Chainlink 价格源

npx hardhat run scripts/configure-token.ts --network sepolia
```

## 常见问题

### Q1: 如何修改部署参数?

编辑对应的模块文件:

```typescript
// 例如: 修改 NFT 名称
export default buildModule("MajtNFTModule", (m) => {
  const nftContract = m.contract("MajtDutchNFT", [
    "MyCustomNFT",  // 修改名称
    "MCN"           // 修改符号
  ]);
  return { nftContract };
});
```

### Q2: 如何在部署时立即铸造 NFT?

在 `MajtDutchNFT.ts` 中取消注释 `m.call()`:

```typescript
// 取消注释这部分
m.call(nftContract, "safeMint", [
  "0xYourAddress",
  "7",
]);
```

### Q3: 部署失败怎么办?

常见原因和解决方案:

| 错误 | 原因 | 解决方案 |
|-----|------|---------|
| `INSUFFICIENT_FUNDS` | 账户余额不足 | 申请测试网 Sepolia ETH |
| `NONCE_TOO_LOW` | Nonce 冲突 | 清空 ignition 状态重新部署 |
| `INVALID_ADDRESS` | 地址格式错误 | 检查 .env 中的地址格式 |
| `TIMEOUT` | RPC 连接超时 | 检查 RPC URL 或更换 RPC 节点 |

### Q4: 如何跳过某些部署步骤?

创建新的模块文件，只包含需要的步骤:

```typescript
export default buildModule("PartialDeployment", (m) => {
  // 只部署 PriceOracle 和 MajtDutchNFT
  const { priceOracle } = m.useModule("PriceOracleModule");
  const { nftContract } = m.useModule("MajtNFTModule");
  return { priceOracle, nftContract };
});
```

### Q5: 如何回滚部署?

Ignition 不支持直接回滚。解决方案:

1. 重新部署到新地址
2. 迁移状态（如果需要）
3. 更新前端配置

## 最佳实践

### 1. 部署前检查

```bash
# 验证合约编译
npx hardhat compile

# 运行测试
npx hardhat test

# 模拟部署（不实际发送交易）
npx hardhat ignition deploy ignition/modules/0-Complete.ts --dry-run
```

### 2. 环境隔离

```bash
# 创建环境特定的配置
.env.local        # 本地配置（不上传）
.env.sepolia      # Sepolia 配置
.env.mainnet      # 主网配置
```

### 3. 部署记录

```bash
# 保存部署日志
npx hardhat ignition deploy ignition/modules/0-Complete.ts --network sepolia 2>&1 | tee deployment.log

# 记录关键信息到文件
cat ignition/deployments/chain-11155111/deployed_addresses.json > DEPLOYMENT_RECORD.json
```

### 4. 版本控制

```bash
# 不提交敏感信息
echo "ignition/deployments/chain-*/execution_state.json" >> .gitignore

# 提交部署模块（源代码）
git add ignition/modules/
git commit -m "chore: update deployment modules"
```

## 参考资源

- [Hardhat Ignition 官方文档](https://hardhat.org/ignition/docs)
- [部署最佳实践](https://docs.openzeppelin.com/upgrades-plugins/1.x/hardhat-upgrades)
- [Sepolia 测试网水龙头](https://sepoliafaucet.com)
- [Etherscan Sepolia](https://sepolia.etherscan.io)

---
