import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";
// @ts-ignore
import MajtNFTModule from "./1-MajtDutchNFT";
// @ts-ignore
import NFTAuctionModule from "./3-NFTAuction";
// @ts-ignore
import PriceOracleModule from "./2-PriceOracle";
// @ts-ignore
import AuctionFactoryModule from "./4-AuctionFactory";

/**
 * 完整部署模块 - 一次性部署所有合约
 *
 * 部署顺序和依赖关系:
 * 1. NFTAuctionModule     -> 部署 NFTAuction 实现
 * 2. PriceOracleModule    -> 部署 PriceOracle
 * 3. AuctionFactoryModule -> 部署 AuctionFactory (UUPS代理 + 实现)
 * 4. MajtNFTModule        -> 部署 MajtDutchNFT
 *
 * 使用方式:
 * $ npx hardhat ignition deploy ignition/modules/0-Complete.ts --network sepolia
 *
 * 这个模块会自动处理所有依赖和初始化
 */
export default buildModule("CompleteDeploymentModule", (m) => {

  // 步骤 1: 部署测试用 NFT 合约
  const { nftContract } = m.useModule(MajtNFTModule);
  console.log("✓ MajtDutchNFT 已部署");

  // 步骤 2: 部署 PriceOracle 价格预言机
  const { priceOracle } = m.useModule(PriceOracleModule);
  console.log("✓ PriceOracle 已部署");

  // 步骤 3: 部署 NFTAuction 逻辑合约
  const { nftAuctionImpl } = m.useModule(NFTAuctionModule);
  console.log("✓ NFTAuction 已部署");

  // 步骤 3: 部署 AuctionFactory (包含代理和初始化)
  const { factoryProxy, factoryImpl } = m.useModule(AuctionFactoryModule);
  console.log("✓ AuctionFactory 已部署");



  // 返回所有重要的合约引用
  return {
    // 拍卖系统核心
    nftAuctionImpl,
    factoryImpl,
    factoryProxy,
    priceOracle,
    // 测试资产
    nftContract,
  };
});
