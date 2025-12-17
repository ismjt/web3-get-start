import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

/**
 * PriceOracle 价格预言机部署模块
 * 集成 Chainlink 预言机，获取 ETH、LINK 等资产的实时价格
 *
 * 部署顺序: 2/5
 * 依赖: 无
 * 后续依赖此合约: AuctionFactory, NFTAuction (通过初始化)
 *
 * Chainlink 价格源 (Sepolia):
 * - ETH/USD: 0x694AA1769357215DE4FAC081bf1f309aDC325306
 * - LINK/USD: 0xc59E3633BAAC79493d908e63626716e204A45EdF
 */
export default buildModule("PriceOracleModule", (m) => {
  // 部署价格预言机合约
  // 构造函数中自动配置 ETH 和 LINK 的价格源
  const priceOracle = m.contract("PriceOracle");

  return {
    priceOracle,
  };
});
