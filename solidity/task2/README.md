# Sample Hardhat 3 Beta Project (`node:test` and `viem`)

This project showcases a Hardhat 3 Beta project using the native Node.js test runner (`node:test`) and the `viem` library for Ethereum interactions.

To learn more about the Hardhat 3 Beta, please visit the [Getting Started guide](https://hardhat.org/docs/getting-started#getting-started-with-hardhat-3). To share your feedback, join our [Hardhat 3 Beta](https://hardhat.org/hardhat3-beta-telegram-group) Telegram group or [open an issue](https://github.com/NomicFoundation/hardhat/issues/new) in our GitHub issue tracker.

## Usage

### Running Tests

To run all the tests in the project, execute the following command:

```shell
npx hardhat test
```

You can also selectively run the Solidity or `node:test` tests:

```shell
npx hardhat test solidity
npx hardhat test nodejs
```

### Make a deployment to Sepolia

This project includes an example Ignition module to deploy the contract. You can deploy this module to a locally simulated chain or to Sepolia.

To run the deployment to a local chain:

```shell
npx hardhat ignition deploy ignition/modules/MyERC20.ts
```

To run the deployment to Sepolia, you need an account with funds to send the transaction. The provided Hardhat configuration includes a Configuration Variable called `SEPOLIA_PRIVATE_KEY`, which you can use to set the private key of the account you want to use.

You can set the `SEPOLIA_PRIVATE_KEY` variable using the `hardhat-keystore` plugin or by setting it as an environment variable.

To set the `SEPOLIA_PRIVATE_KEY` config variable using `hardhat-keystore`:

```shell
npx hardhat keystore set SEPOLIA_PRIVATE_KEY
npx hardhat keystore set SEPOLIA_RPC_URL
```

After setting the variable, you can run the deployment with the Sepolia network:

```shell
npx hardhat ignition deploy --network sepolia ignition/modules/MyERC20.ts
npx hardhat ignition deploy --network sepolia ignition/modules/MajtNFT.ts
```

运行脚本
```shell
npx hardhat run scripts/send-op-tx.ts
npx hardhat run scripts/nftMint-mint.ts
```

MajtTokenzMss
# 部署情况

## 1、测试钱包地址
https://sepolia.etherscan.io/address/0xa298831ad84d1a92cf99d744c4c1c827ee06a49a

## 2、代币合约地址
https://sepolia.etherscan.io/token/0x0b52DA9a3F43C1359D088F9958f620d0BC3A99Dc

## 3、NFT合约地址
* 第一次部署：https://sepolia.etherscan.io/token/0xcae83593Aed07fBf6A258D4DA82D9d447B1f1CFD
* 第二次部署：https://sepolia.etherscan.io/token/0x8368229D9AB703C6CF86021530232d03924e4C76

### 3.1、NFT地址
* https://sepolia.etherscan.io/nft/0xcae83593Aed07fBf6A258D4DA82D9d447B1f1CFD/2
* https://sepolia.etherscan.io/nft/0x8368229D9AB703C6CF86021530232d03924e4C76/0

## 4、讨饭合约地址
https://sepolia.etherscan.io/token/0xe59E77d2659F199F9cb82964f7673C1f75C2baB8

> 测试脚本：[scripts/begging.ts](./scripts/begging.ts)，截图如下：

![BeggingContract在Sepolia测试网的部署与测试情况](./doc/sepolia-begging-test.png)