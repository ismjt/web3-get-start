import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

/**
 * NFTAuction 逻辑合约部署模块
 * 这是一个可升级合约的实现合约，通过工厂模式创建代理实例
 *
 * 部署顺序: 1/5
 * 依赖: 无
 * 后续依赖此合约: AuctionFactory
 */
export default buildModule("NFTAuctionModule", (m) => {
  // 部署 NFTAuction 实现合约
  // 这是逻辑合约，不直接交互，由工厂创建代理时使用
  const nftAuctionImpl = m.contract("NFTAuction");

  return {
    nftAuctionImpl,
  };
});
