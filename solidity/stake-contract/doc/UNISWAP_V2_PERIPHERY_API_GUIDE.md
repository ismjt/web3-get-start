# Uniswap V2 Periphery 接口文档

## 目录
1. [概述](#概述)
2. [核心合约](#核心合约)
3. [路由器接口](#路由器接口)
4. [库函数](#库函数)
5. [迁移合约](#迁移合约)
6. [使用示例](#使用示例)

---

## 概述

Uniswap V2 Periphery（外围合约）提供了一套与Uniswap V2核心合约交互的接口和工具函数。主要包括：

- **UniswapV2Router02** - 主要的交互合约（推荐使用）
- **UniswapV2Router01** - 早期版本（已废弃）
- **UniswapV2Migrator** - V1到V2的迁移工具
- **UniswapV2Library** - 数学和查询库
- **支持的接口** - ERC20、WETH、工厂等

### 关键特性
- ✅ 流动性管理（添加/移除）
- ✅ 代币交换（多种模式）
- ✅ ETH直接支持
- ✅ 链接交换路由
- ✅ 手续费处理（含自动扣费代币）
- ✅ 签名授权（Permit）
- ✅ 期限保护

---

## 核心合约

### 1. UniswapV2Router02 (主合约)

**地址**: 链上已部署
**实现**: `/contracts/UniswapV2Router02.sol`

#### 构造函数
```solidity
constructor(address _factory, address _WETH)
```
- `_factory`: Uniswap V2工厂地址
- `_WETH`: WETH代币地址

#### 关键属性
```solidity
address public immutable factory;    // 工厂合约地址
address public immutable WETH;       // WETH合约地址
```

---

### 2. UniswapV2Router01 (早期版本)

**实现**: `/contracts/UniswapV2Router01.sol`
**状态**: 已弃用，建议使用Router02

---

### 3. UniswapV2Migrator

**目标**: 从Uniswap V1迁移流动性到V2
**实现**: `/contracts/UniswapV2Migrator.sol`

#### 构造函数
```solidity
constructor(address _factoryV1, address _router)
```

#### 核心方法
```solidity
function migrate(
    address token,
    uint amountTokenMin,
    uint amountETHMin,
    address to,
    uint deadline
) external
```

**参数**:
- `token`: V1中的代币地址
- `amountTokenMin`: 接受的最小代币数量
- `amountETHMin`: 接受的最小ETH数量
- `to`: 接收V2 LP token的地址
- `deadline`: 交易期限时间戳

**流程**:
1. 从V1交换移除流动性
2. 将资产添加到V2流动性池
3. 返回未使用的资产

---

## 路由器接口

### IUniswapV2Router02 接口

继承自 `IUniswapV2Router01` 并添加新功能

#### A. 流动性管理函数

##### 1. 添加流动性 - 代币对

```solidity
function addLiquidity(
    address tokenA,
    address tokenB,
    uint amountADesired,
    uint amountBDesired,
    uint amountAMin,
    uint amountBMin,
    address to,
    uint deadline
) external returns (
    uint amountA,
    uint amountB,
    uint liquidity
)
```

**参数**:
- `tokenA`, `tokenB`: 代币地址
- `amountADesired`, `amountBDesired`: 期望的代币数量
- `amountAMin`, `amountBMin`: 滑点保护的最小数量
- `to`: LP token接收地址
- `deadline`: 交易有效期（Unix时间戳）

**返回**:
- `amountA`, `amountB`: 实际添加的数量
- `liquidity`: 获得的LP token数量

**要求**:
- 调用者需提前授权代币给路由器
- deadline >= 当前区块时间

**示例**:
```javascript
// 添加 1000 USDC 和 1 DAI 的流动性
await router.addLiquidity(
    usdcAddress,
    daiAddress,
    ethers.parseUnits("1000", 6),    // 1000 USDC
    ethers.parseEther("1"),           // 1 DAI
    ethers.parseUnits("990", 6),      // 最少990 USDC
    ethers.parseEther("0.99"),        // 最少0.99 DAI
    userAddress,
    Math.floor(Date.now() / 1000) + 60*20  // 20分钟期限
)
```

---

##### 2. 添加流动性 - ETH配对

```solidity
function addLiquidityETH(
    address token,
    uint amountTokenDesired,
    uint amountTokenMin,
    uint amountETHMin,
    address to,
    uint deadline
) external payable returns (
    uint amountToken,
    uint amountETH,
    uint liquidity
)
```

**参数**:
- `token`: 代币地址
- `amountTokenDesired`: 期望的代币数量
- `amountTokenMin`, `amountETHMin`: 滑点保护最小值
- `to`: LP token接收地址
- `deadline`: 期限时间戳
- `msg.value`: 发送的ETH数量

**返回**: 同上

**特殊说明**:
- 多余的ETH将被退款
- 需通过`msg.value`发送ETH

**示例**:
```javascript
// 添加 100 UNI 和 5 ETH 的流动性
await router.addLiquidityETH(
    uniTokenAddress,
    ethers.parseEther("100"),         // 100 UNI
    ethers.parseEther("99"),          // 最少99 UNI
    ethers.parseEther("4.95"),        // 最少4.95 ETH
    userAddress,
    Math.floor(Date.now() / 1000) + 60*20,
    { value: ethers.parseEther("5") } // 发送5 ETH
)
```

---

##### 3. 移除流动性

```solidity
function removeLiquidity(
    address tokenA,
    address tokenB,
    uint liquidity,
    uint amountAMin,
    uint amountBMin,
    address to,
    uint deadline
) public returns (uint amountA, uint amountB)
```

**参数**:
- `tokenA`, `tokenB`: 代币地址
- `liquidity`: 要移除的LP token数量
- `amountAMin`, `amountBMin`: 最小输出保护
- `to`: 输出接收地址
- `deadline`: 期限

**返回**:
- `amountA`, `amountB`: 收到的代币数量

**要求**:
- 调用者需拥有相应的LP token
- 通常需先授权LP token给路由器

**示例**:
```javascript
// 移除流动性
const lpAmount = ethers.parseEther("10");  // 10个LP tokens
await router.removeLiquidity(
    usdcAddress,
    daiAddress,
    lpAmount,
    ethers.parseUnits("990", 6),      // 最少990 USDC
    ethers.parseEther("0.99"),        // 最少0.99 DAI
    userAddress,
    Math.floor(Date.now() / 1000) + 60*20
)
```

---

##### 4. 移除流动性 - ETH配对

```solidity
function removeLiquidityETH(
    address token,
    uint liquidity,
    uint amountTokenMin,
    uint amountETHMin,
    address to,
    uint deadline
) public returns (uint amountToken, uint amountETH)
```

**参数**:
- `token`: 代币地址
- 其他同上

**返回**: 代币和ETH数量

**示例**:
```javascript
await router.removeLiquidityETH(
    uniTokenAddress,
    lpAmount,
    ethers.parseEther("99"),          // 最少99 UNI
    ethers.parseEther("4.95"),        // 最少4.95 ETH
    userAddress,
    deadline
)
```

---

##### 5. 使用签名移除流动性

```solidity
function removeLiquidityWithPermit(
    address tokenA,
    address tokenB,
    uint liquidity,
    uint amountAMin,
    uint amountBMin,
    address to,
    uint deadline,
    bool approveMax,
    uint8 v,
    bytes32 r,
    bytes32 s
) external returns (uint amountA, uint amountB)
```

**特殊参数**:
- `approveMax`: 是否授权最大值
- `v, r, s`: EIP-712签名参数

**优势**:
- 单笔交易完成授权和移除
- 无需预先授权LP token

---

##### 6. 使用签名移除流动性 - ETH

```solidity
function removeLiquidityETHWithPermit(
    address token,
    uint liquidity,
    uint amountTokenMin,
    uint amountETHMin,
    address to,
    uint deadline,
    bool approveMax,
    uint8 v,
    bytes32 r,
    bytes32 s
) external returns (uint amountToken, uint amountETH)
```

---

##### 7. 移除流动性 - 支持手续费代币

```solidity
function removeLiquidityETHSupportingFeeOnTransferTokens(
    address token,
    uint liquidity,
    uint amountTokenMin,
    uint amountETHMin,
    address to,
    uint deadline
) public returns (uint amountETH)
```

**特殊说明**:
- 支持转账时自动扣费的代币
- 检查余额而非假设转账量

---

##### 8. 带签名的费用代币支持移除

```solidity
function removeLiquidityETHWithPermitSupportingFeeOnTransferTokens(
    address token,
    uint liquidity,
    uint amountTokenMin,
    uint amountETHMin,
    address to,
    uint deadline,
    bool approveMax,
    uint8 v,
    bytes32 r,
    bytes32 s
) external returns (uint amountETH)
```

---

#### B. 交换函数

##### 1. 精确输入交换 - 代币对代币

```solidity
function swapExactTokensForTokens(
    uint amountIn,
    uint amountOutMin,
    address[] calldata path,
    address to,
    uint deadline
) external returns (uint[] memory amounts)
```

**参数**:
- `amountIn`: 输入代币的确定数量
- `amountOutMin`: 输出的最小接受数量（滑点保护）
- `path`: 交换路径（例如: [DAI, WETH, USDC]）
- `to`: 输出接收地址
- `deadline`: 期限

**返回**:
- `amounts`: 每一步的输入/输出数量数组

**工作流程**:
1. DAI -> WETH (通过DAI-WETH对)
2. WETH -> USDC (通过WETH-USDC对)

**示例**:
```javascript
// 交换100 DAI，期望至少99 USDC
const amountIn = ethers.parseEther("100");
const path = [daiAddress, wethAddress, usdcAddress];

const amounts = await router.swapExactTokensForTokens(
    amountIn,
    ethers.parseUnits("99", 6),  // 最少99 USDC
    path,
    userAddress,
    deadline
)
// amounts[0] = 100 DAI (输入)
// amounts[1] = X WETH (中间)
// amounts[2] = Y USDC (输出)
```

---

##### 2. 精确输出交换 - 代币对代币

```solidity
function swapTokensForExactTokens(
    uint amountOut,
    uint amountInMax,
    address[] calldata path,
    address to,
    uint deadline
) external returns (uint[] memory amounts)
```

**参数**:
- `amountOut`: 期望的确定输出数量
- `amountInMax`: 愿意输入的最大数量（滑点保护）
- 其他同上

**特点**:
- 输出确定，输入不确定
- 如果需要更少的输入，会返回超额的代币

**示例**:
```javascript
// 获取确定的100 USDC，最多花费102 DAI
const amounts = await router.swapTokensForExactTokens(
    ethers.parseUnits("100", 6),  // 想要100 USDC
    ethers.parseEther("102"),     // 最多花102 DAI
    [daiAddress, wethAddress, usdcAddress],
    userAddress,
    deadline
)
```

---

##### 3. 精确输入交换 - ETH到代币

```solidity
function swapExactETHForTokens(
    uint amountOutMin,
    address[] calldata path,
    address to,
    uint deadline
) external payable returns (uint[] memory amounts)
```

**要求**:
- `path[0]` 必须是 WETH 地址
- 通过 `msg.value` 发送ETH

**示例**:
```javascript
const amounts = await router.swapExactETHForTokens(
    ethers.parseEther("9.9"),      // 最少9.9 DAI
    [wethAddress, daiAddress],
    userAddress,
    deadline,
    { value: ethers.parseEther("10") }  // 发送10 ETH
)
```

---

##### 4. 精确输入交换 - 代币到ETH

```solidity
function swapExactTokensForETH(
    uint amountIn,
    uint amountOutMin,
    address[] calldata path,
    address to,
    uint deadline
) external returns (uint[] memory amounts)
```

**要求**:
- `path[path.length - 1]` 必须是 WETH 地址

**示例**:
```javascript
const amounts = await router.swapExactTokensForETH(
    ethers.parseEther("100"),      // 100 DAI
    ethers.parseEther("4.95"),     // 最少4.95 ETH
    [daiAddress, wethAddress],
    userAddress,
    deadline
)
```

---

##### 5. 精确输出交换 - ETH到代币

```solidity
function swapETHForExactTokens(
    uint amountOut,
    address[] calldata path,
    address to,
    uint deadline
) external payable returns (uint[] memory amounts)
```

**特点**:
- 代币数量确定
- ETH数量不确定（会返回多余的ETH）

---

##### 6. 精确输出交换 - 代币到ETH

```solidity
function swapTokensForExactETH(
    uint amountOut,
    uint amountInMax,
    address[] calldata path,
    address to,
    uint deadline
) external returns (uint[] memory amounts)
```

---

##### 7. 交换 - 支持手续费代币 (精确输入)

```solidity
function swapExactTokensForTokensSupportingFeeOnTransferTokens(
    uint amountIn,
    uint amountOutMin,
    address[] calldata path,
    address to,
    uint deadline
) external
```

**特点**:
- 不返回中间值，仅返回最终余额
- 支持转账自动扣费的代币

---

##### 8. 交换 - 支持手续费代币 (ETH->代币)

```solidity
function swapExactETHForTokensSupportingFeeOnTransferTokens(
    uint amountOutMin,
    address[] calldata path,
    address to,
    uint deadline
) external payable
```

---

##### 9. 交换 - 支持手续费代币 (代币->ETH)

```solidity
function swapExactTokensForETHSupportingFeeOnTransferTokens(
    uint amountIn,
    uint amountOutMin,
    address[] calldata path,
    address to,
    uint deadline
) external
```

---

#### C. 查询和工具函数

##### 1. 引用价格 (简单计算)

```solidity
function quote(
    uint amountA,
    uint reserveA,
    uint reserveB
) public pure returns (uint amountB)
```

**计算**:
```
amountB = amountA * reserveB / reserveA
```

**用途**: 基于储备计算等价的代币数量

---

##### 2. 获取输出数量

```solidity
function getAmountOut(
    uint amountIn,
    uint reserveIn,
    uint reserveOut
) public pure returns (uint amountOut)
```

**计算** (考虑0.3%手续费):
```
amountInWithFee = amountIn * 997
amountOut = (amountInWithFee * reserveOut) / (reserveIn * 1000 + amountInWithFee)
```

**用途**: 计算交换的输出数量

---

##### 3. 获取输入数量

```solidity
function getAmountIn(
    uint amountOut,
    uint reserveIn,
    uint reserveOut
) public pure returns (uint amountIn)
```

**用途**: 反向计算需要的输入数量

---

##### 4. 获取路由输出量

```solidity
function getAmountsOut(
    uint amountIn,
    address[] memory path
) public view returns (uint[] memory amounts)
```

**用途**: 计算整个交换路径的输入和输出数量

**示例**:
```javascript
const path = [daiAddress, wethAddress, usdcAddress];
const amounts = await router.getAmountsOut(
    ethers.parseEther("100"),
    path
)
// amounts[0] = 100 DAI
// amounts[1] = ? WETH
// amounts[2] = ? USDC
```

---

##### 5. 获取路由输入量

```solidity
function getAmountsIn(
    uint amountOut,
    address[] memory path
) public view returns (uint[] memory amounts)
```

**用途**: 计算达到目标输出需要的输入量

---

## 库函数

### UniswapV2Library

位置: `/contracts/libraries/UniswapV2Library.sol`

#### 1. 排序代币

```solidity
function sortTokens(address tokenA, address tokenB)
    internal pure
    returns (address token0, address token1)
```

**说明**: 返回排序后的代币地址

---

#### 2. 计算配对地址

```solidity
function pairFor(
    address factory,
    address tokenA,
    address tokenB
) internal pure returns (address pair)
```

**说明**: 使用CREATE2计算配对地址（不需要调用工厂）

---

#### 3. 获取储备

```solidity
function getReserves(
    address factory,
    address tokenA,
    address tokenB
) internal view returns (uint reserveA, uint reserveB)
```

**说明**: 获取并排序配对的储备

---

#### 4. 引用

```solidity
function quote(
    uint amountA,
    uint reserveA,
    uint reserveB
) internal pure returns (uint amountB)
```

**说明**: 基于恒定乘积公式计算等价数量

---

#### 5. 获取单对输出

```solidity
function getAmountOut(
    uint amountIn,
    uint reserveIn,
    uint reserveOut
) internal pure returns (uint amountOut)
```

**公式**:
```
fee = 0.3% (997/1000)
amountOut = (amountIn * 997 * reserveOut) / (reserveIn * 1000 + amountIn * 997)
```

---

#### 6. 获取单对输入

```solidity
function getAmountIn(
    uint amountOut,
    uint reserveIn,
    uint reserveOut
) internal pure returns (uint amountIn)
```

---

#### 7. 获取链式输出

```solidity
function getAmountsOut(
    address factory,
    uint amountIn,
    address[] memory path
) internal view returns (uint[] memory amounts)
```

---

#### 8. 获取链式输入

```solidity
function getAmountsIn(
    address factory,
    uint amountOut,
    address[] memory path
) internal view returns (uint[] memory amounts)
```

---

## 使用示例

### 例子1: 添加流动性 (代币对)

```javascript
const ethers = require('ethers');

// 连接到Ethereum网络
const provider = new ethers.JsonRpcProvider('https://...');
const wallet = new ethers.Wallet(privateKey, provider);

// 合约ABI和地址
const routerAddress = '0x7a250d5630b4cf539739df2c5dacb4c659f2488d'; // Uniswap V2 Router
const routerABI = [...]; // Router ABI

// 代币地址
const tokenA = '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48'; // USDC
const tokenB = '0x6B175474E89094C44Da98b954EedeAC495271d0F'; // DAI

// 连接路由器
const router = new ethers.Contract(routerAddress, routerABI, wallet);

// 批准代币
const erc20ABI = ['function approve(address spender, uint256 amount)'];
const tokenAContract = new ethers.Contract(tokenA, erc20ABI, wallet);
const tokenBContract = new ethers.Contract(tokenB, erc20ABI, wallet);

await tokenAContract.approve(routerAddress, ethers.parseEther('1000'));
await tokenBContract.approve(routerAddress, ethers.parseEther('1000'));

// 添加流动性
const tx = await router.addLiquidity(
    tokenA,
    tokenB,
    ethers.parseUnits('1000', 6),     // 1000 USDC
    ethers.parseEther('1000'),        // 1000 DAI
    ethers.parseUnits('900', 6),      // 最少900 USDC
    ethers.parseEther('900'),         // 最少900 DAI
    wallet.address,
    Math.floor(Date.now() / 1000) + 60 * 20  // 20分钟期限
);

await tx.wait();
console.log('流动性已添加');
```

---

### 例子2: 添加流动性 (ETH)

```javascript
// 代币地址
const tokenAddress = '0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984'; // UNI

// 添加流动性 (1 ETH + 100 UNI)
const tx = await router.addLiquidityETH(
    tokenAddress,
    ethers.parseEther('100'),        // 100 UNI
    ethers.parseEther('99'),         // 最少99 UNI
    ethers.parseEther('0.95'),       // 最少0.95 ETH
    wallet.address,
    Math.floor(Date.now() / 1000) + 60 * 20,
    {
        value: ethers.parseEther('1')  // 发送1 ETH
    }
);

await tx.wait();
console.log('ETH-UNI流动性已添加');
```

---

### 例子3: 代币交换 (确切输入)

```javascript
// 交换100 DAI 为 USDC
const daiAddress = '0x6B175474E89094C44Da98b954EedeAC495271d0F';
const wethAddress = '0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2';
const usdcAddress = '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48';

// 批准DAI
const daiContract = new ethers.Contract(daiAddress, erc20ABI, wallet);
await daiContract.approve(routerAddress, ethers.parseEther('100'));

// 执行交换
const path = [daiAddress, wethAddress, usdcAddress];
const tx = await router.swapExactTokensForTokens(
    ethers.parseEther('100'),         // 100 DAI
    ethers.parseUnits('99', 6),       // 最少99 USDC
    path,
    wallet.address,
    Math.floor(Date.now() / 1000) + 60 * 20
);

const receipt = await tx.wait();
console.log('交换完成');
```

---

### 例子4: 代币交换 (确切输出)

```javascript
// 获取确定的100 USDC，最多花102 DAI
const tx = await router.swapTokensForExactTokens(
    ethers.parseUnits('100', 6),    // 想要100 USDC
    ethers.parseEther('102'),       // 最多花102 DAI
    [daiAddress, wethAddress, usdcAddress],
    wallet.address,
    Math.floor(Date.now() / 1000) + 60 * 20
);

const receipt = await tx.wait();
console.log('已获得确定的USDC');
```

---

### 例子5: ETH交换 (到代币)

```javascript
// 用1 ETH交换DAI
const tx = await router.swapExactETHForTokens(
    ethers.parseEther('900'),  // 最少900 DAI
    [wethAddress, daiAddress],
    wallet.address,
    Math.floor(Date.now() / 1000) + 60 * 20,
    {
        value: ethers.parseEther('1')  // 发送1 ETH
    }
);

const receipt = await tx.wait();
console.log('ETH已交换为DAI');
```

---

### 例子6: 移除流动性

```javascript
// 移除LP token
const lpTokenAddress = '0x...'; // UNI-ETH LP token地址
const lpContract = new ethers.Contract(lpTokenAddress, erc20ABI, wallet);

// 批准LP token
await lpContract.approve(routerAddress, ethers.parseEther('10'));

// 移除流动性
const tx = await router.removeLiquidityETH(
    tokenAddress,
    ethers.parseEther('10'),        // 移除10个LP tokens
    ethers.parseEther('99'),        // 最少99 UNI
    ethers.parseEther('0.95'),      // 最少0.95 ETH
    wallet.address,
    Math.floor(Date.now() / 1000) + 60 * 20
);

const receipt = await tx.wait();
console.log('流动性已移除');
```

---

### 例子7: 查询操作

```javascript
// 查询交换输出量
const path = [daiAddress, wethAddress, usdcAddress];
const amounts = await router.getAmountsOut(
    ethers.parseEther('100'),
    path
);

console.log('输出数量:');
console.log('DAI输入:', ethers.formatEther(amounts[0]));
console.log('WETH中间:', ethers.formatEther(amounts[1]));
console.log('USDC输出:', ethers.formatUnits(amounts[2], 6));

// 查询交换输入量
const amountsIn = await router.getAmountsIn(
    ethers.parseUnits('100', 6),
    path
);

console.log('需要的输入:');
console.log('DAI需要:', ethers.formatEther(amountsIn[0]));
```

---

## 错误处理

### 常见错误

| 错误 | 原因 | 解决 |
|------|------|------|
| `UniswapV2Router: EXPIRED` | 交易超过期限 | 增加deadline参数 |
| `UniswapV2Router: INSUFFICIENT_OUTPUT_AMOUNT` | 实际输出少于最小值 | 降低amountOutMin或增加输入 |
| `UniswapV2Router: EXCESSIVE_INPUT_AMOUNT` | 需要的输入超过最大值 | 增加amountInMax或降低输出 |
| `ERC20: insufficient allowance` | 批准不足 | 增加授权数量 |
| `UniswapV2Library: IDENTICAL_ADDRESSES` | 代币地址相同 | 确保tokenA != tokenB |
| `UniswapV2Library: INVALID_PATH` | 路径错误 | 检查path数组 |

---

## 安全建议

### ✅ 最佳实践

1. **始终使用最小值滑点保护**
   ```javascript
   const minOut = expectedOut * 0.99;  // 1%滑点
   ```

2. **验证deadline**
   ```javascript
   const deadline = Math.floor(Date.now() / 1000) + 60 * 20;  // 20分钟
   ```

3. **检查允许的代币**
   - 验证path[0]和path[path.length-1]的正确性
   - ETH交换必须包含WETH地址

4. **使用带Permit的函数**
   - 减少交易步骤
   - 节省gas

5. **处理手续费代币**
   - 使用SupportingFeeOnTransferTokens变体
   - 检查实际收到的数量

### ❌ 避免做的事

- ❌ 设置过低的deadline
- ❌ 忽略滑点保护
- ❌ 发送不必要的多余ETH
- ❌ 在没有授权的情况下调用
- ❌ 使用过期的储备数据

---

## 主网地址

### Ethereum Mainnet
- **Router V2**: `0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D`
- **Factory**: `0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f`
- **WETH**: `0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2`

### 其他网络
- **Polygon**: 查询Uniswap文档
- **Arbitrum**: 查询Uniswap文档
- **Optimism**: 查询Uniswap文档

---

## 参考资源

- [Uniswap官方文档](https://docs.uniswap.org)
- [Github v2-periphery](https://github.com/Uniswap/v2-periphery)
- [智能合约代码](https://github.com/Uniswap/v2-core)

---

*最后更新: 2025-12-17*
