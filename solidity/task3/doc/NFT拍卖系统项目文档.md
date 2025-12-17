# NFT拍卖系统项目文档

## 目录
1. [项目概述](#项目概述)
2. [系统架构](#系统架构)
3. [智能合约详解](#智能合约详解)
4. [部署步骤](#部署步骤)
5. [使用流程](#使用流程)
6. [API参考](#api参考)

---

## 项目概述

### 功能说明

NFT拍卖系统是一个基于Solidity的去中心化NFT拍卖平台，具有以下核心功能：

#### 核心功能
- **多链NFT拍卖**：支持在主链（Sepolia）进行NFT拍卖
- **多币种竞拍**：支持使用ETH和ERC20代币进行竞拍
- **价格预言机集成**：使用Chainlink预言机获取实时价格，所有出价均以USD计价
- **跨链竞拍**：支持来自其他链（如Polygon）的跨链竞拍功能（待验证）
- **UUPS可升级代理**：使用UUPS代理模式，支持智能合约升级
- **工厂模式**：采用Uniswap V2风格的工厂模式，每场拍卖部署独立实例

#### 关键特性
- **防重入保护**：使用ReentrancyGuard防止重入攻击
- **访问控制**：基于角色的访问控制（AccessControl）
- **安全转账**：使用安全的`safeTransferFrom`和`call`进行资金转移
- **手续费机制**：平台可收取可配置的手续费

### 项目亮点

| 特性 | 说明 |
|-----|------|
| 链上价格预言机 | 集成Chainlink获取ETH、LINK等资产实时价格 |
| 跨链互操作 | 使用CCIP实现不同区块链间的消息传递和竞拍 |
| 代理模式 | UUPS升级代理，可无缝升级逻辑而保持状态 |
| 工厂模式 | 每场拍卖独立部署，降低风险，提高隔离性 |
| USD定价 | 所有出价统一转换为USD，跨币种比较更直观 |

---

## 系统架构

### 合约结构图

```
┌─────────────────────────────────────────────┐
│         AuctionFactory (工厂合约)             │
│    (可升级, UUPS代理模式)                    │
│  - 管理所有拍卖实例                          │
│  - 创建新拍卖                                │
│  - 管理手续费                                │
└─────────────┬──────────────────────────────┘
              │ 创建多个独立实例
              ▼
    ┌────────────────────────┐
    │   NFTAuction (拍卖)    │  (多个实例，ERC1967Proxy)
    │  - 管理出价             │
    │  - 执行结算             │
    │  - 结束拍卖             │
    └────────────────────────┘

┌─────────────────────────────────────────────┐
│      PriceOracle（价格预言机)                 │
│  - 获取ETH/USD价格                           │
│  - 获取ERC20/USD价格                         │
│  - 支持配置多种代币价格源                       │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│      CcipAdapter (跨链适配器)                 │
│  - 处理跨链消息                               │
│  - 管理跨链竞拍                               │
│  - 支持跨链NFT转移                            │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│      MajtDutchNFT (测试NFT)                  │
│  - 基于ERC721URIStorage                      │
│  - 支持自定义metadata URI                     │
└─────────────────────────────────────────────┘
```

### 数据流向

```
1. NFT拍卖创建流程
   NFT Owner → Approve NFT to Factory → CreateAuction
   ↓
   Factory transfers NFT → Auction Contract (locked)

2. 竞拍流程
   Bidder → bidEth() or bidERC20()
   ↓
   PriceOracle converts to USD
   ↓
   Validate bid amount
   ↓
   Refund previous bidder
   ↓
   Update highest bidder

3. 拍卖结束流程
   endAuction() called after duration expires
   ↓
   Transfer NFT to winner or return to seller
   ↓
   Transfer funds to seller (minus platform fee)
   ↓
   emit AuctionEnded event
```

---

## 智能合约详解

### 1. NFTAuction (拍卖合约)

**文件位置**: `contracts/NFTAuction.sol`

#### 核心结构体

```solidity
struct AuctionData {
    address seller;        // NFT卖家地址
    address nftContract;   // NFT合约地址
    uint256 tokenId;       // NFT Token ID
    uint256 startPrice;    // 起拍价（USD, 18位小数）
    uint256 startTime;     // 拍卖开始时间
    uint256 duration;      // 拍卖时长（秒）
    uint256 feeRate;       // 手续费率（基点）
}

struct NftBidInfo {
    address highestBidder; // 最高出价者
    uint256 usdPrice;      // 最高出价（USD）
    address paymentToken;  // 支付代币(0为ETH)
    uint256 tokenAmount;   // 代币数量
}
```

#### 主要函数

| 函数名 | 可见性 | 说明 |
|--------|--------|------|
| `initialize()` | external | 初始化拍卖合约（代理模式） |
| `bidEth()` | external | 使用ETH出价（payable） |
| `bidERC20()` | external | 使用ERC20代币出价 |
| `endAuction()` | external | 结束拍卖（检查时间条件） |
| `claim()` | public | 领取NFT和资金 |
| `cancelAuction()` | external | 卖家取消拍卖（无出价时） |
| `getAuctionInfo()` | external | 获取完整拍卖信息 |
| `timeRemaining()` | external | 获取剩余时间 |
| `canEnd()` | external | 检查拍卖是否可结束 |

#### 关键逻辑

**出价验证**:
```
1. 检查拍卖是否在进行中 (startTime ≤ now ≤ startTime + duration)
2. 检查拍卖未结束 (!auctionEnded)
3. 检查出价金额转换为USD后：
   - 不低于起拍价 (usdPrice ≥ startPrice)
   - 高于当前最高出价 (usdPrice > nftBidInfo.usdPrice)
4. 退还上一个出价者的资金
5. 更新最高出价者信息
```

**结算流程**:
```
1. 拍卖时间结束时调用 endAuction()
2. 若有出价者：
   - 计算平台手续费: fee = tokenAmount × feeRate / 10000
   - 卖家收款: sellerAmount = tokenAmount - fee
   - NFT转移给出价最高者
   - 资金转移给卖家
3. 若无出价者：
   - NFT返回给卖家
   - 资金返回给相关方
```

### 2. AuctionFactory (拍卖工厂合约)

**文件位置**: `contracts/AuctionFactory.sol`

#### 核心功能

工厂合约是系统的入口点，负责：
- 创建新的拍卖合约实例（ERC1967Proxy）
- 管理所有拍卖列表
- 配置平台参数

#### 主要函数

| 函数名 | 可见性 | 说明 |
|--------|--------|------|
| `initialize()` | public | 初始化工厂合约 |
| `createAuction()` | external | 创建新拍卖（需NFT授权） |
| `upgradeAuctionImplementation()` | external | 升级拍卖实现 |
| `setFeeRate()` | external | 设置手续费率（最高10%） |
| `setAuctionTokenPriceFeed()` | external | 配置代币价格源 |
| `getAuction()` | external | 按索引获取拍卖地址 |
| `getSellerAuctions()` | external | 获取卖家所有拍卖 |
| `getAuctionsByNFT()` | external | 获取特定NFT的拍卖列表 |
| `endAuction()` | external | 结束指定拍卖 |
| `withdrawFees()` | external | 提取ETH手续费 |
| `withdrawERC20Fees()` | external | 提取ERC20手续费 |

#### 工厂模式设计

```
创建拍卖步骤:
1. createAuction() called
2. 验证NFT所有权和授权
3. 从卖家转移NFT到Factory
4. 部署ERC1967Proxy:
   - Implementation: auctionImplementation
   - InitData: initialize() 调用数据
5. 将NFT从Factory转移到Proxy合约（锁定）
6. 记录拍卖信息
7. 返回拍卖地址

优点:
- 隔离: 每场拍卖独立合约实例
- 升级: 可随时升级拍卖逻辑
- 灵活: 支持不同的拍卖配置
```

### 3. PriceOracle (价格预言机)

**文件位置**: `contracts/lib/PriceOracle.sol`

#### 功能说明

集成Chainlink V3预言机获取资产价格：

| 资产 | Sepolia预言机地址 |
|-----|-------------------|
| ETH/USD | 0x694AA1769357215DE4FAC081bf1f309aDC325306 |
| LINK/USD | 0xc59E3633BAAC79493d908e63626716e204A45EdF |

#### 主要函数

| 函数名 | 说明 |
|--------|------|
| `getEthPrice()` | 获取ETH/USD价格（18位小数） |
| `getTokenPrice(address token)` | 获取任意ERC20/USD价格 |
| `setTokenPriceFeed()` | 配置新代币的价格源 |
| `calculateUsdValue()` | 将代币数量转换为USD价值 |

#### 价格转换逻辑

```solidity
// ETH价格转换示例
usdValue = ethAmount * getEthPrice() / 10^18

// ERC20价格转换示例
usdValue = tokenAmount * getTokenPrice(token) / 10^decimals

// Chainlink数据精度标准化
- Chainlink返回8位精度价格
- 内部统一为18位精度
- 调用时需考虑代币精度
```

### 4. MajtDutchNFT (测试NFT合约)

**文件位置**: `contracts/MajtDutchNFT.sol`

#### 基本信息

- **基类**: ERC721URIStorage（支持每个Token自定义URI）
- **功能**: 用于测试，支持铸造NFT

#### 主要函数

| 函数名 | 说明 |
|--------|------|
| `safeMint()` | 铸造新NFT（仅owner） |

#### 示例

```solidity
// 铸造NFT
tokenId = nft.safeMint(
    recipient,
    "ipfs://Qm..."  // metadata URI
);

// 授权拍卖工厂
nft.approve(factory, tokenId);
// 或
nft.setApprovalForAll(factory, true);
```

---

## 部署步骤

### 前置要求

1. **环境配置**
   - Node.js >= 18.0
   - npm >= 9.0
   - Hardhat >= 3.0

2. **依赖安装**
   ```bash
   npm install
   ```

3. **环境变量** (`.env` 文件)
   ```env
   SEPOLIA_RPC_URL=https://sepolia.infura.io/v3/YOUR_INFURA_KEY
   SEPOLIA_PRIVATE_KEY=replace_private_key_here
   ```

### 部署流程

#### 1. 编译合约

```bash
npx hardhat compile
```

编译输出在 `artifacts/` 目录

#### 2. 本地测试（可选）

```bash
# 启动本地节点
npx hardhat node

# 在另一个终端运行测试
npx hardhat test --network localhost
```

#### 3. 部署到Sepolia测试网

**方案A: 使用Hardhat Ignition（推荐）**

```bash
# 部署所有模块
npx hardhat ignition deploy ignition/modules/1-MajtDutchNFT.ts \
  --network sepolia
  
  npx hardhat ignition deploy ignition/modules/2-PriceOracle.ts \
  --network sepolia
  
npx hardhat ignition deploy ignition/modules/3-NFTAuction.ts \
  --network sepolia

npx hardhat ignition deploy ignition/modules/4-AuctionFactory.ts \
  --network sepolia
```

**方案B: 使用自定义部署脚本**

```bash
npx hardhat run scripts/deploy-sepolia.ts --network sepolia
```

#### 4. 部署后验证

```bash
# 验证智能合约
npx hardhat verify --network sepolia <CONTRACT_ADDRESS> <CONSTRUCTOR_ARGS>
```

### 部署清单

| 合约 | 地址 | 说明 |
|-----|------|------|
| PriceOracle | - | 价格预言机（非升级） |
| NFTAuction (Impl) | - | 拍卖逻辑实现 |
| AuctionFactory (Proxy) | - | 工厂代理（升级） |
| MajtDutchNFT | - | 测试NFT |

### 初始化步骤

部署后需执行以下初始化：

```javascript
// 1. 初始化工厂（如使用代理）
factory.initialize(adminAddress, auctionImplementation, priceOracleAddress)

// 2. 配置代币价格源（可选）
priceOracle.setTokenPriceFeed(tokenAddress, chainlinkFeed)

// 3. 配置CCIP适配器（跨链功能）
// ccipAdapter.setAuctionContract(auctionAddress)
// ccipAdapter.allowlistSourceChain(chainSelector, true)
// ccipAdapter.allowlistDestinationChain(chainSelector, true)
```

---

## 使用流程

### 场景1: 卖家创建拍卖

#### 步骤

```javascript
// 1. 铸造或准备NFT
const tokenId = await nft.safeMint(sellerAddress, "ipfs://metadata");

// 2. 授权工厂转移NFT
await nft.approve(factoryAddress, tokenId);
// 或
await nft.setApprovalForAll(factoryAddress, true);

// 3. 创建拍卖
const tx = await factory.createAuction(
    nftAddress,
    tokenId,
    "1000000000000000000", // 1000 USD (18位小数)
    Math.floor(Date.now() / 1000) + 60, // 60秒后开始
    3600  // 1小时拍卖时长
);

const receipt = await tx.wait();
// 从事件日志中获取拍卖合约地址
const auctionAddress = receipt.events.find(e => e.event === 'AuctionCreated').args.auction;
```

### 场景2: 竞拍者ETH竞拍

```javascript
const auction = await ethers.getContractAt('NFTAuction', auctionAddress);

// 获取起拍价信息
const { auctionData } = await auction.getAuctionInfo();

// 计算需要出价的ETH数量
// 假设ETH = $2000，起拍价 = $1000，需要0.5 ETH
const bidAmount = ethers.parseEther("0.5");

// 出价
const tx = await auction.bidEth({
    value: bidAmount
});

await tx.wait();

// 验证出价
const { nftBidInfo } = await auction.getAuctionInfo();
console.log("Current highest bidder:", nftBidInfo.highestBidder);
console.log("Current highest bid (USD):", nftBidInfo.usdPrice.toString());
```

### 场景3: 竞拍者ERC20竞拍

```javascript
const token = await ethers.getContractAt('IERC20', tokenAddress);
const auction = await ethers.getContractAt('NFTAuction', auctionAddress);

// 1. 授权拍卖合约转移代币
const bidAmount = ethers.parseUnits("1000", 18); // 假设1000 token
await token.approve(auctionAddress, bidAmount);

// 2. 出价
const tx = await auction.bidERC20(bidAmount, tokenAddress);
await tx.wait();
```

### 场景4: 结束拍卖并领取

```javascript
const auction = await ethers.getContractAt('NFTAuction', auctionAddress);

// 检查拍卖是否可结束
const canEnd = await auction.canEnd();
require(canEnd, "Auction cannot end yet");

// 1. 结束拍卖
const tx = await auction.endAuction();
await tx.wait();

// 2. 自动执行了claim()，无需额外操作
// 或者手动调用
await auction.claim();

// 验证结果
const { auctionEnded, auctionClaimed, nftBidInfo } = await auction.getAuctionInfo();
console.log("Auction ended:", auctionEnded);
console.log("Auction claimed:", auctionClaimed);
console.log("Winner:", nftBidInfo.highestBidder);
```

## API参考

### NFTAuction

#### bidEth()
```solidity
function bidEth() external payable nonReentrant
```

**参数**:
- `msg.value`: ETH出价金额

**要求**:
- 拍卖进行中
- ETH金额 > 0
- USD价值 >= 起拍价
- USD价值 > 当前最高出价

**事件**: `BidPlaced(bidder, amount, usdValue)`

---

#### bidERC20()
```solidity
function bidERC20(uint256 amount, address paymentToken) external nonReentrant
```

**参数**:
- `amount`: ERC20代币数量
- `paymentToken`: ERC20合约地址

**要求**:
- 拍卖进行中
- 代币数量 > 0
- 已授权拍卖合约转移
- USD价值验证同上

**事件**: `BidPlaced(bidder, amount, usdValue)`

---

#### endAuction()
```solidity
function endAuction() external
```

**要求**:
- 拍卖时间已过期
- 拍卖未结束
- 自动调用`claim()`

**事件**: `AuctionEnded(winner, amount)`

---

#### getAuctionInfo()
```solidity
function getAuctionInfo() external view returns (
    AuctionData memory,
    NftBidInfo memory,
    bool auctionEnded,
    bool auctionClaimed
)
```

**返回值**:
- 拍卖数据结构
- 出价信息结构
- 拍卖结束标志
- 领取标志

---

### AuctionFactory

#### createAuction()
```solidity
function createAuction(
    address nftContract,
    uint256 tokenId,
    uint256 startPriceUSD,
    uint256 startTime,
    uint256 duration
) external returns (address)
```

**参数**:
- `nftContract`: NFT合约地址
- `tokenId`: NFT Token ID
- `startPriceUSD`: 起拍价（USD, 18位小数）
- `startTime`: 拍卖开始时间戳
- `duration`: 拍卖时长（秒）

**要求**:
- 调用者是NFT所有者
- NFT已授权给工厂

**返回**: 拍卖合约代理地址

**事件**: `AuctionCreated(auction, seller, nft, tokenId, price)`

---

#### setFeeRate()
```solidity
function setFeeRate(uint256 _feeRate) external onlyRole(DEFAULT_ADMIN_ROLE)
```

**参数**:
- `_feeRate`: 手续费率（基点，最高1000 = 10%）

---

#### withdrawFees()
```solidity
function withdrawFees(address payable to) external onlyRole(DEFAULT_ADMIN_ROLE) nonReentrant
```

提取累积的ETH手续费

---

### PriceOracle

#### getEthPrice()
```solidity
function getEthPrice() public view returns (uint256)
```

**返回**: ETH/USD价格（18位小数）

**示例**:
- 返回 2000e18 表示 $2000/ETH

---

#### calculateUsdValue()
```solidity
function calculateUsdValue(uint256 amount, address paymentToken) public view returns (uint256)
```

**参数**:
- `amount`: 代币数量
- `paymentToken`: 代币地址（address(0)表示ETH）

**返回**: USD价值（18位小数）

---

#### setTokenPriceFeed()
```solidity
function setTokenPriceFeed(address token, address priceFeed) public onlyOwner
```

配置新代币的Chainlink价格源

---

## 常见问题

### Q1: 如何改变拍卖时间？
需要卖家取消当前拍卖并重新创建新的拍卖。

### Q2: 支持哪些ERC20代币？
任何有Chainlink价格源的ERC20代币都支持，可通过`setTokenPriceFeed()`扩展。

### Q3: 平台手续费如何计算？
```
手续费 = 出价金额 × 手续费率 / 10000
卖家收益 = 出价金额 - 手续费
```

### Q4: 跨链拍卖的延迟是多少？
取决于Chainlink CCIP的消息确认速度，一般需要数分钟到数小时。

### Q5: 如何升级拍卖逻辑？
工厂持有人可调用`upgradeAuctionImplementation()`升级逻辑，后续新拍卖会使用新逻辑。

---

## 安全考虑

### 已实现的安全措施

1. **重入保护**: 使用`ReentrancyGuard`修饰器
2. **访问控制**: 基于角色的权限管理
3. **价格保护**: 通过预言机验证价格，防止价格操纵
4. **时间检查**: 严格验证拍卖时间窗口
5. **代理安全**: UUPS升级模式，带有授权检查
6. **转账安全**: 使用安全转账方法，处理失败回退


## 参考资源

- [Solidity文档](https://docs.soliditylang.org/)
- [OpenZeppelin合约](https://docs.openzeppelin.com/contracts/)
- [Chainlink预言机](https://docs.chain.link/)
- [Hardhat文档](https://hardhat.org/docs)
- [UUPS升级模式](https://docs.openzeppelin.com/contracts/4.x/api/proxy#UUPSUpgradeable)

---

**文档版本**: 1.0

**最后更新**: 2025年12月

**维护者**: majt 
