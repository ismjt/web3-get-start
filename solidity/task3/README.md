# NFT拍卖场


# 快速开始

```cmd
# 1. 安装依赖
npm install

# 2. 编译合约
npx hardhat compile

# 3. 运行测试
npx hardhat test

# 4. 部署到测试网
npx hardhat run scripts/deploy.js --network sepolia
```

# 部署到测试网
## 1.获取测试网ETF
Sepolia: https://sepoliafaucet.com/
## 2.配置 RPC 节点
可以在[Metamask Developer](https://developer.metamask.io)获取Ethereum Sepolia地址：https://sepolia.infura.io/v3/4ef555db37xxxxxx51c231cfe198d7f
## 部署到 Sepolia
```cmd
# 部署合约
npx hardhat run scripts/deploy.js --network sepolia
npx hardhat ignition deploy ignition/modules/MajtDutchNFT.ts

# 验证合约（部署后）
npx hardhat verify --network sepolia CONTRACT_ADDRESS CONSTRUCTOR_ARGS
```
## 检查部署

# 运行测试

## 本地测试
```cmd
# 运行所有测试
npx hardhat test

# 运行特定测试文件
npx hardhat test test/NFTAuction.test.js

# 查看 gas 消耗
REPORT_GAS=true npx hardhat test

```


# 测试网工具

# 价格预言机
Chainlink Price Feeds: https://docs.chain.link/data-feeds/price-feeds/addresses
## Ethereum Testnet
1. ETH/USD： 0x694AA1769357215DE4FAC081bf1f309aDC325306