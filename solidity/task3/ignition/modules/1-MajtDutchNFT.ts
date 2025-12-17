import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

/**
 * MajtDutchNFT 测试NFT部署模块
 * ERC721 标准NFT合约，用于拍卖系统的测试和演示
 *
 * 部署顺序: 1/5
 * 依赖: 无
 * 用途: 为拍卖系统提供测试NFT
 *
 * 功能:
 * - 基于 ERC721URIStorage，每个Token可自定义metadata URI
 * - 支持 mint 铸造新NFT
 */
export default buildModule("MajtNFTModule", (m) => {
  // 部署 MajtDutchNFT 合约
  // 参数: name="MajtDutchNFT", symbol="MDNFT"
  const nftContract = m.contract("MajtDutchNFT", ["MajtDutchNFT", "MDNFT"]);

  // 可选: 为测试账户铸造NFT
  // 注释掉以避免不必要的测试NFT生成

  // 测试钱包一 - 铸造1个NFT
  const nftOwner = m.getAccount(0);
  m.call(nftContract, "safeMint", [
    nftOwner,
    "1",
  ], { id: "safeMintToken1" });

  /*
   m.call(nftContract, "safeMint", [
     "0xeCBB02Aac0ea98E218b8E3A891F04a87a2663970",
     "ipfs://QmeSjSinHpPnmXmspMjwiXyN6zS4E9zccariGR3jxcaWtq/1",
   ]);

   // 测试钱包二 - 铸造3个NFT
   m.call(nftContract, "safeMint", [
     "0xA298831AD84d1A92CF99d744c4C1C827EE06A49A",
     "ipfs://QmeSjSinHpPnmXmspMjwiXyN6zS4E9zccariGR3jxcaWtq/2",
   ]);

   m.call(nftContract, "safeMint", [
     "0xA298831AD84d1A92CF99d744c4C1C827EE06A49A",
     "ipfs://QmeSjSinHpPnmXmspMjwiXyN6zS4E9zccariGR3jxcaWtq/3",
   ]);

   m.call(nftContract, "safeMint", [
     "0xA298831AD84d1A92CF99d744c4C1C827EE06A49A",
     "ipfs://QmeSjSinHpPnmXmspMjwiXyN6zS4E9zccariGR3jxcaWtq/4",
   ]);
   */

  return {
    nftContract,
  };
});
