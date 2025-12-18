# Uniswap V2 Periphery 文档中心

欢迎查阅Uniswap V2外围合约的完整接口文档。本文档集合涵盖了所有核心接口、函数签名、使用示例和最佳实践。

## 📚 文档总览

### 1. [UNISWAP_V2_PERIPHERY_API_GUIDE.md](./UNISWAP_V2_PERIPHERY_API_GUIDE.md) - 📖 完整API指南
**类型**: 详细参考文档 | **长度**: ~2000行 | **适合**: 深入学习

包含内容：
- ✅ 核心合约详细说明
- ✅ 所有公开方法的完整文档
- ✅ 参数详解和返回值说明
- ✅ 7个完整的代码示例
- ✅ 错误处理和安全建议
- ✅ 主网地址和参考资源

**何时使用**:
- 需要理解某个方法的细节
- 编写合约集成代码
- 学习Uniswap工作原理

---

### 2. [UNISWAP_V2_INTERFACE_SPEC.md](./UNISWAP_V2_INTERFACE_SPEC.md) - 🔧 接口规范
**类型**: 技术规范文档 | **长度**: ~800行 | **适合**: 深度开发

包含内容：
- ✅ 完整的接口定义（Solidity代码）
- ✅ 接口继承关系
- ✅ 方法签名表格
- ✅ 数据结构和验证规则
- ✅ 交易流程图
- ✅ 数学公式和函数选择器
- ✅ 部署检查表

**何时使用**:
- 实现兼容接口
- 理解合约内部结构
- 智能合约审计
- 深度技术分析

---

### 3. [UNISWAP_V2_QUICK_REFERENCE.md](./UNISWAP_V2_QUICK_REFERENCE.md) - ⚡ 快速参考
**类型**: 速查手册 | **长度**: ~300行 | **适合**: 快速查阅

包含内容：
- ✅ 核心地址速查
- ✅ 常用方法一览
- ✅ 代码片段（即插即用）
- ✅ 参数速查表
- ✅ 常见错误排查
- ✅ 最佳实践检查表
- ✅ 燃料费估算

**何时使用**:
- 快速查找某个方法
- 复制代码片段
- 排查常见问题
- 开发过程中速查

---

## 🎯 按使用场景选择文档

### 🆕 我是初学者
推荐阅读顺序：
1. 先读 **快速参考** 的"核心概念"部分
2. 然后读 **API指南** 的"使用示例"
3. 最后读 **接口规范** 的"流程图"

### 👨‍💻 我要集成到项目中
推荐顺序：
1. **API指南** - 找到需要的方法
2. **快速参考** - 复制代码片段
3. **接口规范** - 理解数据结构

### 🔍 我在做代码审计
推荐顺序：
1. **接口规范** - 理解接口和流程
2. **API指南** - 理解各方法的安全考虑
3. **快速参考** - 最佳实践检查表

### 🐛 我在调试问题
推荐顺序：
1. **快速参考** - "常见错误排查"
2. **API指南** - 相关方法的详细说明
3. **接口规范** - 参数验证规则

---

## 📋 核心合约一览

| 合约 | 作用 | 主要方法数 | 推荐用途 |
|------|------|----------|---------|
| **UniswapV2Router02** | 主路由器 | 24个 | 日常交互 |
| **UniswapV2Router01** | 早期版本 | 18个 | ❌ 已弃用 |
| **UniswapV2Migrator** | V1→V2迁移 | 1个 | 流动性迁移 |
| **UniswapV2Library** | 数学库 | 8个 | 内部使用 |

---

## 🔑 核心功能分类

### 流动性管理 (6个方法)
```
addLiquidity              → 添加代币对流动性
addLiquidityETH          → 添加代币+ETH流动性
removeLiquidity          → 移除代币对流动性
removeLiquidityETH       → 移除代币+ETH流动性
removeLiquidityWithPermit        → 使用签名移除（代币对）
removeLiquidityETHWithPermit     → 使用签名移除（+ETH）
```

### 交换功能 (12个方法)
#### 基础交换
```
swapExactTokensForTokens        → 确定输入、不确定输出
swapTokensForExactTokens        → 不确定输入、确定输出
swapExactETHForTokens           → ETH→代币
swapTokensForExactETH           → 代币→ETH
swapExactTokensForETH           → 确定输入
swapETHForExactTokens           → 确定输出
```

#### 手续费代币支持（6个）
```
swapExactTokensForTokensSupportingFeeOnTransferTokens
swapExactETHForTokensSupportingFeeOnTransferTokens
swapExactTokensForETHSupportingFeeOnTransferTokens
removeLiquidityETHSupportingFeeOnTransferTokens
removeLiquidityETHWithPermitSupportingFeeOnTransferTokens
```

### 查询功能 (5个方法)
```
quote              → 计算等价数量
getAmountOut       → 计算单对输出
getAmountIn        → 计算单对输入
getAmountsOut      → 计算路由输出
getAmountsIn       → 计算路由输入
```

---

## 🚀 快速开始示例

### 示例1: 查询交换输出

```javascript
const path = [
  '0x6B175474E89094C44Da98b954EedeAC495271d0F',  // DAI
  '0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2',  // WETH
  '0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48'   // USDC
];

const amounts = await router.getAmountsOut(
  ethers.parseEther('100'),
  path
);

console.log('输出USDC:', ethers.formatUnits(amounts[2], 6));
```

### 示例2: 执行代币交换

```javascript
await tokenA.approve(routerAddr, ethers.parseEther('100'));

const tx = await router.swapExactTokensForTokens(
  ethers.parseEther('100'),     // 100 DAI
  ethers.parseUnits('99', 6),   // 最少99 USDC
  path,
  userAddr,
  Math.floor(Date.now() / 1000) + 20 * 60
);

await tx.wait();
```

### 示例3: 添加流动性

```javascript
await tokenA.approve(routerAddr, ethers.parseEther('100'));
await tokenB.approve(routerAddr, ethers.parseEther('100'));

const tx = await router.addLiquidity(
  tokenAAddr,
  tokenBAddr,
  ethers.parseEther('100'),     // 100 tokenA
  ethers.parseEther('100'),     // 100 tokenB
  ethers.parseEther('95'),      // 滑点保护
  ethers.parseEther('95'),
  userAddr,
  Math.floor(Date.now() / 1000) + 20 * 60
);

const receipt = await tx.wait();
console.log('LP token数量:', receipt);
```

---

## 📊 主要参数速查

| 参数 | 含义 | 示例 | 说明 |
|------|------|------|------|
| **path** | 交换路径 | [DAI, WETH, USDC] | 数组长度必须≥2 |
| **amountIn** | 输入数量 | 1e18 | 已缩放小数位 |
| **amountOut** | 输出数量 | 1e6 | 已缩放小数位 |
| **amountMin** | 最小保护值 | 0.95e18 | 用于滑点保护 |
| **deadline** | 期限 | block.timestamp + 1200 | Unix时间戳（秒）|
| **to** | 接收地址 | msg.sender | 接收token地址 |

---

## ⚠️ 常见陷阱与解决

| 问题 | 原因 | 解决方案 |
|------|------|---------|
| EXPIRED错误 | deadline太短 | 增加deadline: +20分钟 |
| INSUFFICIENT_OUTPUT | 滑点设置太低 | 增加滑点: 改为5% |
| 授权失败 | 批准额度不足 | 增加approve金额 |
| Path无效 | 缺少配对 | 检查所有pair是否存在 |
| 转账失败 | 代币余额不足 | 检查余额和decimals |

---

## 🔐 安全检查清单

**集成前验证**:
- [ ] Factory地址正确
- [ ] WETH地址正确
- [ ] Token decimals处理正确
- [ ] 滑点保护设置合理（1-5%）
- [ ] deadline设置足够长（10-20分钟）
- [ ] path数组有效
- [ ] 初始授权足够

**交互中验证**:
- [ ] 交易未过期
- [ ] 余额充足
- [ ] 返回值验证
- [ ] 事件正确发出
- [ ] 没有异常错误

---

## 📞 常见问题解答

### Q: Router01和Router02的区别？
A: Router02支持手续费代币，推荐使用02版本。

### Q: 如何处理自扣费代币？
A: 使用`SupportingFeeOnTransferTokens`结尾的方法。

### Q: deadline应该设置多长？
A: 推荐20分钟（1200秒），可根据网络拥堵调整。

### Q: 滑点通常设置多少？
A: 通常1-5%，高volatility时可设5-10%。

### Q: 如何计算手续费？
A: Uniswap V2手续费为恒定0.3%，已在`getAmountOut`公式中包含。

---

## 🔗 相关资源

### 官方资源
- [Uniswap官方文档](https://docs.uniswap.org)
- [V2 Github仓库](https://github.com/Uniswap/v2-periphery)
- [V2核心合约](https://github.com/Uniswap/v2-core)

### 部署地址
- **Ethereum**: 0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D
- **查看所有链**: https://docs.uniswap.org/protocol/reference/deployments

### 工具
- [Etherscan验证](https://etherscan.io)
- [UniswapLabs](https://app.uniswap.org)

---

## 📈 文档使用统计

| 文档 | 行数 | 代码示例 | 表格 | 图表 |
|------|------|---------|------|------|
| API指南 | 2000+ | 7个 | 10+ | 流程图 |
| 接口规范 | 800+ | 代码块 | 15+ | 流程图 |
| 快速参考 | 300+ | 15+ | 8+ | - |

---

## 💡 最佳实践汇总

✅ **应该做的**:
- 使用Router V2而不是V1
- 设置合理的期限和滑点
- 充足的代币授权
- 验证path有效性
- 捕获交易事件

❌ **不应该做的**:
- 忽视deadline检查
- 设置0%滑点
- 精确授权金额
- 假设path任意有效
- 不验证返回值

---

## 🎓 学习路径建议

### 初级 (1-2小时)
1. 读快速参考"核心概念"
2. 运行API指南的"例子1"
3. 理解path和deadline概念

### 中级 (3-4小时)
1. 学习所有交换方法
2. 理解流动性管理
3. 手动计算滑点

### 高级 (5-6小时)
1. 学习接口规范中的数学公式
2. 理解燃料优化
3. 实现自己的路由逻辑

---

## 📝 版本信息

| 项目 | 版本 | 日期 |
|------|------|------|
| Uniswap V2 | v2 | 2020-06 |
| Solidity | ^0.6.6 | - |
| 文档版本 | 2.0 | 2025-12-17 |

---

## 📧 反馈与更新

本文档会持续维护和更新。如有任何建议或发现错误，欢迎反馈。

---

**立即开始**: 选择上面的文档之一，根据你的需求开始学习！
