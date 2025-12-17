import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";
// @ts-ignore
import NFTAuctionModule from "./3-NFTAuction";
// @ts-ignore
import PriceOracleModule from "./2-PriceOracle";

/**
 * AuctionFactory 工厂合约部署模块
 * 管理所有拍卖实例的创建和部署，采用 UUPS 升级代理模式
 *
 * 部署顺序: 3/5 (最后)
 * 依赖: NFTAuctionModule (nftAuctionImpl), PriceOracleModule (priceOracle)
 *
 * 工作流程:
 * 1. 部署 AuctionFactory 实现合约
 * 2. 通过 UUPSProxy 代理部署并初始化工厂
 * 3. 初始化参数:
 *    - admin: 部署账户地址（可管理合约）
 *    - auctionImplementation: NFTAuction 实现地址
 *    - priceOracle: PriceOracle 地址
 * 4. 初始化后的工厂可创建新的拍卖实例
 */
export default buildModule("AuctionFactoryModule", (m) => {
  // 引入依赖的模块
  const { nftAuctionImpl } = m.useModule(NFTAuctionModule);
  const { priceOracle } = m.useModule(PriceOracleModule);

  // 获取部署者账户地址 (作为 admin)
  const admin = m.getAccount(0);

  // 部署 AuctionFactory 逻辑合约
  const factoryImpl = m.contract("AuctionFactory");

  // 部署 UUPSProxy 代理合约
  // 初始化数据会在代理构造时调用 initialize 函数
  const factoryProxy = m.contract("UUPSProxy", [
      factoryImpl,
      m.encodeFunctionCall(
        factoryImpl,
          "initialize",
          [admin, nftAuctionImpl, priceOracle]
      ),
  ]);

  return {
    factoryImpl,
    factoryProxy,
  };
});
