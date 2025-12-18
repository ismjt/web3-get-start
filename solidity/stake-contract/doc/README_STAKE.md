# MetaNodeStake 合约学习总结

## 📚 目录
1. [合约概述](#合约概述)
2. [核心概念](#核心概念)
3. [数据结构详解](#数据结构详解)
4. [主要功能模块](#主要功能模块)
5. [使用流程](#使用流程)
6. [代码解析](#代码解析)
7. [常见问题](#常见问题)

---

## 合约概述

### 这个合约做什么？

MetaNodeStake 是一个**质押奖励合约**。用户可以：
- 质押 ETH 或其他代币到合约中
- 根据质押数量和时间获得 MetaNode 代币奖励
- 在锁定期后解质押取回原始代币

### 类比理解

想象一个银行定期存款：
- 你存入 100 元钱（质押）
- 银行每个月给你 2 元利息（奖励）
- 3 个月后你可以取出 100 元 + 6 元利息

### 合约特点

✅ **可升级合约**：使用 UUPS 模式，可升级逻辑而不改变数据
✅ **访问控制**：不同角色有不同权限
✅ **暂停机制**：紧急时可以暂停提现和领取
✅ **多资金池**：支持 ETH 和各种 ERC20 代币质押
✅ **灵活配置**：管理员可调整奖励规则

---

## 核心概念

### 1. 资金池（Pool）

什么是资金池？
- **定义**：一个质押对的配置单位。每个资金池对应一种代币
- **例子**：
  - 资金池 0：ETH 质押
  - 资金池 1：USDC 质押
  - 资金池 2：DAI 质押

### 2. 权重（Pool Weight）

权重用来分配奖励。假设总权重为 100，奖励分配规则：

```
单个资金池的奖励 = 总奖励 × (资金池权重 / 总权重)
```

**例子**：
- 每块产生 100 MetaNode
- 资金池 0 权重：60（ETH 质押）
- 资金池 1 权重：40（USDC 质押）
- 则每块 ETH 池分得 60 个，USDC 池分得 40 个

### 3. 累积奖励（accMetaNodePerST）

"acc"是累积的意思，"PerST"是每个质押代币。

**含义**：从创建资金池到现在，每质押 1 个ETH平均能获得的 MetaNode

**例子**：
- 如果 accMetaNodePerST = 5 ether
- 你质押了 10 个代币
- 你的待领取奖励大约为 10 × 5 = 50 个 MetaNode

**为什么乘以 1 ether？**
- Solidity 不支持小数，所以用乘以 1e18 来保留精度
- 计算：`奖励 = (质押数 × accMetaNodePerST) / 1e18`

### 4. 奖励计算公式

```
用户的待领取奖励 = (用户质押数 × 资金池累积奖励) - 用户已领取奖励 + 用户待领取奖励
```

**分解解释**：
1. `用户质押数 × 资金池累积奖励`：理论应得
2. 减去 `用户已领取奖励`：避免重复计算
3. 加上 `用户待领取奖励`：处理未完全领取的部分

---

## 数据结构详解

### Pool 结构体

```solidity
struct Pool {
    address stTokenAddress;      // 质押代币地址（0x0 表示 ETH）
    uint256 poolWeight;          // 这个池的权重
    uint256 lastRewardBlock;     // 上次更新奖励的区块高度
    uint256 accMetaNodePerST;    // 每个质押代币的累积奖励
    uint256 stTokenAmount;       // 当前池中质押的总代币数
    uint256 minDepositAmount;    // 最小质押数量
    uint256 unstakeLockedBlocks; // 解质押后锁定的区块数
}
```

**为什么需要这些字段？**

| 字段 | 作用 | 例子 |
|------|------|------|
| stTokenAddress | 标识是什么代币 | 0x0 表示 ETH，0x... 表示合约地址 |
| poolWeight | 控制奖励分配比例 | 权重 60 在总权重 100 中占 60% |
| lastRewardBlock | 上次更新时间戳 | 用来计算过了多少个块 |
| accMetaNodePerST | 保存累积奖励 | 新用户进来时用这个计算 |
| stTokenAmount | 池的总规模 | 用来计算单位奖励 |
| minDepositAmount | 入场门槛 | 防止太小的质押 |
| unstakeLockedBlocks | 锁定时间 | 提现需要等待这些块 |

### User 结构体

```solidity
struct User {
    uint256 stAmount;          // 用户在这个池中质押的数量
    uint256 finishedMetaNode;  // 用户已经领取的 MetaNode 数
    uint256 pendingMetaNode;   // 用户待领取的 MetaNode 数
    UnstakeRequest[] requests; // 用户的解质押请求列表
}
```

**用户的三个金额状态**：

```
总应得奖励 = finishedMetaNode + pendingMetaNode + 当前可计算的奖励
```

### UnstakeRequest 结构体

```solidity
struct UnstakeRequest {
    uint256 amount;       // 这次要解质押多少
    uint256 unlockBlocks; // 哪个区块高度后才能取出
}
```

**为什么要记录多个请求？**
用户可能多次解质押，每次可能等待时间不同。

**例子**：
```
第 100 块：用户请求解质押 10 个，锁到 150 块
第 110 块：用户请求解质押 5 个，锁到 160 块
第 155 块：用户可取出 10 个（第一次解锁）
第 160 块：用户可取出 5 个（第二次解锁）
```

---

## 主要功能模块

### 📝 1. 管理函数（Admin Only）

#### addPool() - 添加新的资金池

```solidity
function addPool(
    address _stTokenAddress,      // 质押代币地址
    uint256 _poolWeight,          // 权重
    uint256 _minDepositAmount,    // 最小质押数
    uint256 _unstakeLockedBlocks, // 锁定块数
    bool _withUpdate              // 是否更新所有池
) public onlyRole(ADMIN_ROLE)
```

**注意**：
- 第一个资金池必须是 ETH（地址为 0x0）
- 不要添加同一个代币多次，否则会计算错误
- `_withUpdate=true` 会调用 `massUpdatePools()`，可能消耗大量 gas

#### setPoolWeight() - 调整池的权重

```solidity
function setPoolWeight(
    uint256 _pid,      // 资金池 ID
    uint256 _poolWeight, // 新权重
    bool _withUpdate   // 是否先更新
) public onlyRole(ADMIN_ROLE)
```

**使用场景**：
- 提高流行代币的权重以吸引更多质押
- 降低冷门代币权重

### 💰 2. 用户操作函数

#### depositETH() - 质押 ETH

```solidity
function depositETH() public payable whenNotPaused
```

**使用方式**：
```javascript
// 在 JavaScript 中
await contract.depositETH({ value: ethers.parseEther("1.0") });
```

**流程**：
1. 验证质押数 ≥ 最小数量
2. 更新资金池累积奖励
3. 计算用户本次应得的奖励
4. 更新用户质押数

#### deposit() - 质押 ERC20 代币

```solidity
function deposit(
    uint256 _pid,    // 资金池 ID（不能是 0，因为 0 是 ETH）
    uint256 _amount  // 质押数量
) public whenNotPaused
```

**前置步骤**：
```javascript
// 1. 先授权
await tokenContract.approve(stakeContractAddress, amount);

// 2. 再质押
await stakeContract.deposit(poolId, amount);
```

**为什么要授权？**
- ERC20 代币的安全机制
- 合约需要获得转移代币的权限

#### unstake() - 申请解质押

```solidity
function unstake(
    uint256 _pid,    // 资金池 ID
    uint256 _amount  // 解质押数量
) public whenNotPaused
```

**发生的事情**：
1. 计算用户当前应得的奖励（添加到 pendingMetaNode）
2. 记录解质押请求（包括锁定期）
3. 减少用户的 stAmount

⚠️ **重要**：这不会立即返回代币，只是创建一个解质押请求！

#### withdraw() - 取出已解锁的代币

```solidity
function withdraw(uint256 _pid) public whenNotPaused
```

**流程**：
1. 检查所有解质押请求
2. 找出已解锁的请求（当前块 ≥ unlockBlocks）
3. 转移代币给用户
4. 删除已完成的请求

#### claim() - 领取 MetaNode 奖励

```solidity
function claim(uint256 _pid) public whenNotPaused
```

**流程**：
1. 更新资金池最新数据
2. 计算用户待领取的 MetaNode
3. 转移 MetaNode 给用户
4. 重置 pendingMetaNode

### 📊 3. 查询函数

#### pendingMetaNode() - 查询待领取奖励

```solidity
function pendingMetaNode(uint256 _pid, address _user)
    external view returns (uint256)
```

**返回值**：用户当前能领取的 MetaNode 数量

#### stakingBalance() - 查询质押余额

```solidity
function stakingBalance(uint256 _pid, address _user)
    external view returns (uint256)
```

**返回值**：用户当前质押的代币数量

#### withdrawAmount() - 查询可提现金额

```solidity
function withdrawAmount(uint256 _pid, address _user)
    public view returns (uint256 requestAmount, uint256 pendingWithdrawAmount)
```

**返回值**：
- `requestAmount`：所有解质押请求总数（包括锁定中的）
- `pendingWithdrawAmount`：已解锁可取出的数量

### 🔄 4. 内部奖励计算函数

#### updatePool() - 更新资金池

```solidity
function updatePool(uint256 _pid) public
```

**做什么**：
1. 计算自上次更新以来新增的 MetaNode
2. 分配给这个池的份额
3. 更新 accMetaNodePerST

**关键计算**：

```
新增块数 = 当前块 - 上次更新块
新增奖励 = 新增块数 × MetaNodePerBlock × (池权重 / 总权重)
新的累积奖励 = 旧累积奖励 + 新增奖励 / 质押总数
```

#### getMultiplier() - 计算块奖励系数

```solidity
function getMultiplier(uint256 _from, uint256 _to)
    public view returns (uint256 multiplier)
```

**返回值**：块范围内应有的总奖励（未分配权重）

**原理**：
```
multiplier = (_to - _from) × MetaNodePerBlock
```

### 🛡️ 5. 内部安全函数

#### _safeMetaNodeTransfer() - 安全的 MetaNode 转移

```solidity
function _safeMetaNodeTransfer(address _to, uint256 _amount) internal
```

**保护措施**：如果合约 MetaNode 余额不足，就转移所有可用的

**为什么需要？**
- 防止因奖励计算错误导致交易失败
- 保证尽量给用户转移奖励

#### _safeETHTransfer() - 安全的 ETH 转移

```solidity
function _safeETHTransfer(address _to, uint256 _amount) internal
```

**使用方法**：使用低级调用 `.call{value: amount}("")`

**为什么用低级调用？**
- 转移 ETH 需要用 call
- `.transfer()` 和 `.send()` 已经被认为不够灵活

---

## 使用流程

### 完整的用户操作流程

#### 场景 1：质押 ETH 并领取奖励

```
用户 --1. 调用 depositETH()
         (质押 1 ETH)
            ↓
合约 --2. 更新资金池
         计算用户奖励
         记录用户信息
            ↓
用户 --3. 等待 100 个块
            ↓
用户 --4. 调用 claim()
         领取 MetaNode 奖励
            ↓
用户 --5. 调用 unstake()
         申请解质押 1 ETH
            ↓
用户 --6. 等待 1000 个块（锁定期）
            ↓
用户 --7. 调用 withdraw()
         取回 1 ETH
```

#### 场景 2：质押 USDC 代币

```
用户 --1. 调用 approve(合约地址, 数量)
         (授权合约转移 USDC)
            ↓
用户 --2. 调用 deposit(1, 数量)
         (pid=1 是 USDC 池)
            ↓
[同上述场景步骤 3-7]
```

### 代码示例

```javascript
// 使用 ethers.js v6

// 连接到合约
const stakeContract = new ethers.Contract(
    contractAddress,
    abi,
    signer
);

// 1. 质押 ETH
const tx1 = await stakeContract.depositETH({
    value: ethers.parseEther("1.0")
});
await tx1.wait();
console.log("✓ 已质押 1 ETH");

// 2. 查询待领取奖励
const pending = await stakeContract.pendingMetaNode(0, userAddress);
console.log("待领取奖励:", ethers.formatEther(pending));

// 3. 领取奖励
const tx2 = await stakeContract.claim(0);
await tx2.wait();
console.log("✓ 已领取奖励");

// 4. 查询质押余额
const balance = await stakeContract.stakingBalance(0, userAddress);
console.log("质押余额:", ethers.formatEther(balance));

// 5. 申请解质押
const tx3 = await stakeContract.unstake(0, ethers.parseEther("0.5"));
await tx3.wait();
console.log("✓ 已申请解质押 0.5 ETH");

// 6. 等待足够块数后，查询可提现金额
const { requestAmount, pendingWithdrawAmount } =
    await stakeContract.withdrawAmount(0, userAddress);
console.log("可提现:", ethers.formatEther(pendingWithdrawAmount));

// 7. 取回代币
const tx4 = await stakeContract.withdraw(0);
await tx4.wait();
console.log("✓ 已取回代币");
```

---

## 代码解析

### 关键代码片段 1：计算奖励

**来自 pendingMetaNodeByBlockNumber()，第 459-487 行**

```solidity
function pendingMetaNodeByBlockNumber(
    uint256 _pid,
    address _user,
    uint256 _blockNumber
) public view checkPid(_pid) returns (uint256) {
    Pool storage pool_ = pool[_pid];
    User storage user_ = user[_pid][_user];
    uint256 accMetaNodePerST = pool_.accMetaNodePerST;  // 当前累积奖励
    uint256 stSupply = pool_.stTokenAmount;              // 池中总质押数

    // 如果有新的块且池不为空
    if (_blockNumber > pool_.lastRewardBlock && stSupply != 0) {
        // 1. 计算新增块数的基础奖励
        uint256 multiplier = getMultiplier(
            pool_.lastRewardBlock,
            _blockNumber
        );

        // 2. 计算这个池应该分配的奖励（考虑权重）
        uint256 MetaNodeForPool = (multiplier * pool_.poolWeight) /
            totalPoolWeight;

        // 3. 累积奖励 = 累积奖励 + 这轮新增奖励/总质押数
        accMetaNodePerST =
            accMetaNodePerST +
            (MetaNodeForPool * (1 ether)) /
            stSupply;
    }

    // 4. 计算用户应得 = 用户质押数 × 累积奖励 - 已领取 + 待领取
    return
        (user_.stAmount * accMetaNodePerST) /
        (1 ether) -
        user_.finishedMetaNode +
        user_.pendingMetaNode;
}
```

**详细说明**：

```
第一步：新增块 = 当前块 - 上次更新块
       新增奖励 = 新增块 × 每块产出

第二步：这个池的新增奖励 = 新增奖励 × (池权重 / 总权重)

第三步：平均到每个代币 = 池的新增奖励 / 质押总数
       累积奖励 += 平均值（乘以 1e18 保持精度）

第四步：用户奖励 = 用户质押 × 累积奖励 / 1e18
                - 已领取的
                + 之前待领取的
```

### 关键代码片段 2：质押逻辑

**来自 _deposit()，第 748-801 行**

```solidity
function _deposit(uint256 _pid, uint256 _amount) internal {
    Pool storage pool_ = pool[_pid];
    User storage user_ = user[_pid][msg.sender];

    // 1. 首先更新池的数据
    updatePool(_pid);

    // 2. 如果用户之前有质押，计算新增的奖励
    if (user_.stAmount > 0) {
        uint256 accST = (user_.stAmount * pool_.accMetaNodePerST) / (1 ether);
        uint256 pendingMetaNode_ = accST - user_.finishedMetaNode;

        if (pendingMetaNode_ > 0) {
            // 保存待领取奖励，后续可领取
            user_.pendingMetaNode = user_.pendingMetaNode + pendingMetaNode_;
        }
    }

    // 3. 增加用户的质押数
    if (_amount > 0) {
        user_.stAmount = user_.stAmount + _amount;
    }

    // 4. 增加池的总质押数
    pool_.stTokenAmount = pool_.stTokenAmount + _amount;

    // 5. 重新计算用户已领取的奖励（作为新的基准）
    user_.finishedMetaNode =
        (user_.stAmount * pool_.accMetaNodePerST) / (1 ether);

    emit Deposit(msg.sender, _pid, _amount);
}
```

**流程图**：
```
更新池数据
    ↓
计算用户旧奖励（如果有）→ 添加到 pendingMetaNode
    ↓
增加用户质押数
    ↓
增加池总质押数
    ↓
更新用户的基准点（finishedMetaNode）
```

**为什么要更新 finishedMetaNode？**
防止重复计算同一部分的奖励。相当于"重新设定基准点"。

### 关键代码片段 3：解质押和提现

**来自 unstake()，第 630-665 行**

```solidity
function unstake(uint256 _pid, uint256 _amount) public {
    Pool storage pool_ = pool[_pid];
    User storage user_ = user[_pid][msg.sender];

    require(user_.stAmount >= _amount, "Not enough balance");

    // 1. 更新池，计算最新奖励
    updatePool(_pid);

    // 2. 计算当前应得奖励
    uint256 pendingMetaNode_ =
        (user_.stAmount * pool_.accMetaNodePerST) / (1 ether) -
        user_.finishedMetaNode;

    // 3. 保存待领取奖励
    if (pendingMetaNode_ > 0) {
        user_.pendingMetaNode = user_.pendingMetaNode + pendingMetaNode_;
    }

    // 4. 减少质押数
    if (_amount > 0) {
        user_.stAmount = user_.stAmount - _amount;

        // 5. 创建解质押请求（包含锁定期）
        user_.requests.push(
            UnstakeRequest({
                amount: _amount,
                unlockBlocks: block.number + pool_.unstakeLockedBlocks
            })
        );
    }

    // 6. 减少池的总质押数
    pool_.stTokenAmount = pool_.stTokenAmount - _amount;

    // 7. 更新用户的基准点
    user_.finishedMetaNode =
        (user_.stAmount * pool_.accMetaNodePerST) / (1 ether);

    emit RequestUnstake(msg.sender, _pid, _amount);
}
```

**时间轴**：
```
第 100 块：用户调用 unstake(1, 100)
          → 创建 UnstakeRequest，锁定到 1100 块

第 500 块：用户调用 withdraw()
          → 检查 unlockBlocks（1100 > 500，未解锁）
          → 什么都不做

第 1100 块：用户调用 withdraw()
           → 检查 unlockBlocks（1100 ≤ 1100，已解锁）
           → 转移 100 个代币给用户
           → 删除这个请求
```

### 关键代码片段 4：提现逻辑

**来自 withdraw()，第 672-708 行**

```solidity
function withdraw(uint256 _pid) public {
    Pool storage pool_ = pool[_pid];
    User storage user_ = user[_pid][msg.sender];

    uint256 pendingWithdraw_ = 0;
    uint256 popNum_ = 0;

    // 1. 遍历所有解质押请求，找出已解锁的
    for (uint256 i = 0; i < user_.requests.length; i++) {
        if (user_.requests[i].unlockBlocks > block.number) {
            // 后续请求还未解锁，停止遍历
            break;
        }
        // 累加可取出的金额
        pendingWithdraw_ = pendingWithdraw_ + user_.requests[i].amount;
        popNum_++;
    }

    // 2. 删除已解锁的请求（前 popNum_ 个）
    for (uint256 i = 0; i < user_.requests.length - popNum_; i++) {
        user_.requests[i] = user_.requests[i + popNum_];
    }

    for (uint256 i = 0; i < popNum_; i++) {
        user_.requests.pop();
    }

    // 3. 转移代币给用户
    if (pendingWithdraw_ > 0) {
        if (pool_.stTokenAddress == address(0x0)) {
            _safeETHTransfer(msg.sender, pendingWithdraw_);
        } else {
            IERC20(pool_.stTokenAddress).safeTransfer(
                msg.sender,
                pendingWithdraw_
            );
        }
    }

    emit Withdraw(msg.sender, _pid, pendingWithdraw_, block.number);
}
```

**数组删除的巧妙方式**：
```
原数组：[100, 50, 30, 20]（都是解质押金额）
解锁了前 2 个，popNum_ = 2

第一个循环（i: 0 to 1）：
  i=0: requests[0] = requests[2] = 30
  i=1: requests[1] = requests[3] = 20

第二个循环（pop 2 次）：
  删除最后 2 个

结果：[30, 20]
```

---

## 常见问题

### Q1: 为什么 accMetaNodePerST 要乘以 1 ether (1e18)?

**A**: Solidity 只支持整数，没有浮点数。

比如计算 1000 ÷ 3 = 333.333...，如果只存 333，会丢失精度。

解决办法：**先乘以 1e18 再除法**
```
(1000 * 1e18) ÷ 3 = 333...333 * 1e18
之后再除以 1e18 恢复正常大小
```

### Q2: 为什么质押后要调用 updatePool()?

**A**: 需要最新的累积奖励数据。

如果不更新，用户的奖励计算会基于旧数据，导致少领奖励。

**流程**：
```
用户质押前 → updatePool() 更新累积奖励
         → 用户质押 → 用户的基准点 = 新累积奖励
         → 后续只计算新增部分
```

### Q3: finishedMetaNode 和 pendingMetaNode 有什么区别?

**A**:
- **finishedMetaNode**：已经确认领取过的奖励（已转移到用户）
- **pendingMetaNode**：计算出来但还没有领取的奖励

**例子**：
```
用户 A 质押 10 个，资金池累积奖励 = 5
用户 A 应得 = 10 × 5 = 50

领取一次，转移 50 个给用户：
  finishedMetaNode = 50
  pendingMetaNode = 0

后来又质押一次，新增奖励 20：
  用户 A 应得 = 10 × 5 + 10 × 2 = 70
  新增的 20 添加到 pendingMetaNode = 20

领取一次，转移 20 个：
  finishedMetaNode = 70
  pendingMetaNode = 0
```

### Q4: 为什么解质押有锁定期?

**A**: 安全机制。
- 防止用户快速进出套利
- 给项目方时间应对大量提现
- 稳定资金规模

### Q5: 调用 updatePool() 很贵吗？

**A**: 是的。
- 每调用一次就要计算一次
- 如果有很多资金池，massUpdatePools() 可能很昂贵
- **优化**：在链下计算后，只调用需要的池

### Q6: 如何查询一个用户所有的解质押请求?

**A**: 合约没有提供直接函数，需要客户端手动处理：

```javascript
// 获取用户的 User 结构
const userInfo = await stakeContract.user(poolId, userAddress);
// userInfo.requests 就是所有请求

// 逐个检查
userInfo.requests.forEach((req, idx) => {
    console.log(`请求 ${idx}: 金额=${req.amount}, 解锁块=${req.unlockBlocks}`);
});
```

### Q7: 如果合约里的 MetaNode 余额不足怎么办?

**A**: `_safeMetaNodeTransfer()` 会处理：
```solidity
// 如果要转 100，但只有 50
_safeMetaNodeTransfer(user, 100);
// → 只转 50
```

这是一个妥协方案，用户能至少领到一些奖励。

**更好的做法**：管理员应该确保合约有足够的 MetaNode。

### Q8: 初始化后能改什么参数？

**A**:

| 参数 | 初始化时 | 初始化后 | 函数名 |
|------|---------|---------|--------|
| MetaNode 代币 | ✅ 可以 | ✅ 可以 | setMetaNode() |
| startBlock | ✅ 可以 | ✅ 可以 | setStartBlock() |
| endBlock | ✅ 可以 | ✅ 可以 | setEndBlock() |
| MetaNodePerBlock | ✅ 可以 | ✅ 可以 | setMetaNodePerBlock() |
| 资金池信息 | ❌ 不能 | ✅ 可以 | addPool() |
| 池的权重 | ❌ 不能 | ✅ 可以 | setPoolWeight() |
| 池的锁定期 | ❌ 不能 | ✅ 可以 | updatePool() |

---

## 总结

MetaNodeStake 是一个典型的 **DeFi 质押合约**，核心逻辑是：

1. **质押**：用户存入代币
2. **累积**：每个块增加一些奖励
3. **领取**：用户领取 MetaNode 奖励
4. **解质押**：用户取回原始代币（有锁定期）

**学习重点**：
- 理解累积奖励的计算方式
- 掌握多资金池权重分配
- 明确用户状态转换流程
- 注意 Solidity 整数精度问题

---

**最后更新**: 2025-12-18